package world

import (
	"fmt"
	"github.com/df-mc/atomic"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/physics"
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/internal"
	"github.com/df-mc/dragonfly/server/internal/sliceutil"
	"github.com/df-mc/dragonfly/server/world/chunk"
	"github.com/go-gl/mathgl/mgl64"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"math/rand"
	"sync"
	"time"
)

// World implements a Minecraft world. It manages all aspects of what players can see, such as blocks,
// entities and particles.
// World generally provides a synchronised state: All entities, blocks and players usually operate in this
// world, so World ensures that all its methods will always be safe for simultaneous calls.
// A nil *World is safe to use but not functional.
type World struct {
	log internal.Logger
	d   Dimension
	// advance is a bool that specifies if this World should advance the current tick, time and weather saved in the
	// Settings struct held by the World.
	advance bool

	o sync.Once

	npd, epd atomic.Value[*World]

	set     *Settings
	prov    atomic.Value[Provider]
	handler atomic.Value[Handler]
	gen     atomic.Value[Generator]

	weather
	ticker

	rdonly atomic.Bool

	lastPos   ChunkPos
	lastChunk *chunkData

	closing chan struct{}
	running sync.WaitGroup

	chunkMu sync.Mutex
	// chunks holds a cache of chunks currently loaded. These chunks are cleared from this map after some time
	// of not being used.
	chunks map[ChunkPos]*chunkData

	entityMu sync.RWMutex
	// entities holds a map of entities currently loaded and the last ChunkPos that the Entity was in.
	// These are tracked so that a call to RemoveEntity can find the correct entity.
	entities map[Entity]ChunkPos

	r               *rand.Rand
	randomTickSpeed atomic.Uint32

	updateMu sync.Mutex
	// scheduledUpdates is a map of tick time values indexed by the block position at which an update is
	// scheduled. If the current tick exceeds the tick value passed, the block update will be performed
	// and the entry will be removed from the map.
	scheduledUpdates map[cube.Pos]int64
	neighbourUpdates []neighbourUpdate

	viewersMu sync.Mutex
	viewers   map[*Loader]Viewer
}

// New creates a new initialised world. The world may be used right away, but it will not be saved or loaded
// from files until it has been given a different provider than the default. (NopProvider)
// By default, the name of the world will be 'World'.
// The Settings passed specify the initial settings of the World created. These Settings are changed as soon as
// Provider is called, at which point they will be replaced with the Settings as created by the Provider passed. If nil
// is passed as Settings, default settings are used.
func New(log internal.Logger, d Dimension, s *Settings) *World {
	if s == nil {
		s = defaultSettings()
	}
	w := &World{
		advance:          s.ref.Inc() == 1,
		r:                rand.New(rand.NewSource(time.Now().Unix())),
		scheduledUpdates: map[cube.Pos]int64{},
		entities:         map[Entity]ChunkPos{},
		viewers:          map[*Loader]Viewer{},
		prov:             *atomic.NewValue[Provider](NopProvider{}),
		gen:              *atomic.NewValue[Generator](NopGenerator{}),
		handler:          *atomic.NewValue[Handler](NopHandler{}),
		randomTickSpeed:  *atomic.NewUint32(3),
		log:              log,
		set:              s,
		closing:          make(chan struct{}),
		d:                d,
	}
	w.weather, w.ticker = weather{w: w}, ticker{w: w}

	w.initChunkCache()
	go w.tickLoop()
	go w.chunkCacheJanitor()
	return w
}

// Name returns the display name of the world. Generally, this name is displayed at the top of the player list
// in the pause screen in-game.
// If a provider is set, the name will be updated according to the name that it provides.
func (w *World) Name() string {
	w.set.Lock()
	defer w.set.Unlock()
	return w.set.Name
}

// Dimension returns the Dimension assigned to the World in world.New. The sky colour and behaviour of a variety of
// world features differ based on the Dimension assigned to a World.
func (w *World) Dimension() Dimension {
	if w == nil {
		return nopDim{}
	}
	return w.d
}

// Range returns the range in blocks of the World (min and max). It is equivalent to calling World.Dimension().Range().
func (w *World) Range() cube.Range {
	if w == nil {
		return cube.Range{}
	}
	return w.d.Range()
}

// Block reads a block from the position passed. If a chunk is not yet loaded at that position, the chunk is
// loaded, or generated if it could not be found in the world save, and the block returned. Chunks will be
// loaded synchronously.
func (w *World) Block(pos cube.Pos) Block {
	if w == nil || pos.OutOfBounds(w.Range()) {
		// Fast way out.
		return air()
	}
	c := w.chunk(chunkPosFromBlockPos(pos))
	defer c.Unlock()

	rid := c.Block(uint8(pos[0]), int16(pos[1]), uint8(pos[2]), 0)
	if nbtBlocks[rid] {
		// The block was also a block entity, so we look it up in the block entity map.
		if nbtB, ok := c.e[pos]; ok {
			return nbtB
		}
	}
	b, _ := BlockByRuntimeID(rid)
	return b
}

// Biome reads the biome at the position passed. If a chunk is not yet loaded at that position, the chunk is
// loaded, or generated if it could not be found in the world save, and the biome returned. Chunks will be
// loaded synchronously.
func (w *World) Biome(pos cube.Pos) Biome {
	if w == nil || pos.OutOfBounds(w.Range()) {
		// Fast way out.
		return ocean()
	}
	c := w.chunk(chunkPosFromBlockPos(pos))
	defer c.Unlock()

	id := int(c.Biome(uint8(pos[0]), int16(pos[1]), uint8(pos[2])))
	b, ok := BiomeByID(id)
	if !ok {
		w.log.Errorf("could not find biome by ID %v", id)
	}
	return b
}

// blockInChunk reads a block from the world at the position passed. The block is assumed to be in the chunk
// passed, which is also assumed to be locked already or otherwise not yet accessible.
func (w *World) blockInChunk(c *chunkData, pos cube.Pos) Block {
	if pos.OutOfBounds(w.Range()) {
		// Fast way out.
		return air()
	}
	rid := c.Block(uint8(pos[0]), int16(pos[1]), uint8(pos[2]), 0)
	if nbtBlocks[rid] {
		// The block was also a block entity, so we look it up in the block entity map.
		if b, ok := c.e[pos]; ok {
			return b
		}
	}
	b, _ := BlockByRuntimeID(rid)
	return b
}

// HighestLightBlocker gets the Y value of the highest fully light blocking block at the x and z values
// passed in the world.
func (w *World) HighestLightBlocker(x, z int) int {
	if w == nil {
		return w.Range()[0]
	}
	c := w.chunk(ChunkPos{int32(x >> 4), int32(z >> 4)})
	defer c.Unlock()
	return int(c.HighestLightBlocker(uint8(x), uint8(z)))
}

// HighestBlock looks up the highest non-air block in the world at a specific x and z in the world. The y
// value of the highest block is returned, or 0 if no blocks were present in the column.
func (w *World) HighestBlock(x, z int) int {
	if w == nil {
		return w.Range()[0]
	}
	c := w.chunk(ChunkPos{int32(x >> 4), int32(z >> 4)})
	defer c.Unlock()
	return int(c.HighestBlock(uint8(x), uint8(z)))
}

// highestObstructingBlock returns the highest block in the world at a given x and z that has at least a solid top or
// bottom face.
func (w *World) highestObstructingBlock(x, z int) int {
	if w == nil {
		return 0
	}
	yHigh := w.HighestBlock(x, z)
	for y := yHigh; y >= w.Range()[0]; y-- {
		pos := cube.Pos{x, y, z}
		m := w.Block(pos).Model()
		if m.FaceSolid(pos, cube.FaceUp, w) || m.FaceSolid(pos, cube.FaceDown, w) {
			return y
		}
	}
	return w.Range()[0]
}

// SetOpts holds several parameters that may be set to disable updates in the World of different kinds as a result of
// a call to SetBlock.
type SetOpts struct {
	// DisableBlockUpdates makes SetBlock not update any neighbouring blocks as a result of the SetBlock call.
	DisableBlockUpdates bool
	// DisableLiquidDisplacement disables the displacement of liquid blocks to the second layer (or back to the first
	// layer, if it already was on the second layer). Disabling this is not strongly recommended unless performance is
	// very important or where it is known no liquid can be present anyway.
	DisableLiquidDisplacement bool
}

// SetBlock writes a block to the position passed. If a chunk is not yet loaded at that position, the chunk is
// first loaded or generated if it could not be found in the world save.
// SetBlock panics if the block passed has not yet been registered using RegisterBlock().
// Nil may be passed as the block to set the block to air.
//
// A SetOpts struct may be passed to additionally modify behaviour of SetBlock, specifically to improve performance
// under specific circumstances. Nil should be passed where performance is not essential, to make sure the world is
// updated adequately.
//
// SetBlock should be avoided in situations where performance is critical when needing to set a lot of blocks
// to the world. BuildStructure may be used instead.
func (w *World) SetBlock(pos cube.Pos, b Block, opts *SetOpts) {
	if w == nil || pos.OutOfBounds(w.Range()) {
		// Fast way out.
		return
	}
	if opts == nil {
		opts = &SetOpts{}
	}

	x, y, z := uint8(pos[0]), int16(pos[1]), uint8(pos[2])
	c := w.chunk(chunkPosFromBlockPos(pos))

	rid := BlockRuntimeID(b)

	var before uint32
	if rid != airRID && !opts.DisableLiquidDisplacement {
		before = c.Block(x, y, z, 0)
	}

	c.SetBlock(x, y, z, 0, rid)
	if nbtBlocks[rid] {
		c.e[pos] = b
	} else {
		delete(c.e, pos)
	}

	if !opts.DisableLiquidDisplacement {
		if rid == airRID {
			if li := c.Block(x, y, z, 1); li != airRID {
				c.SetBlock(x, y, z, 0, li)
			}
		} else if liquidDisplacingBlocks[rid] && liquidBlocks[before] {
			l, _ := BlockByRuntimeID(before)
			if liq := l.(Liquid); b.(LiquidDisplacer).CanDisplace(liq) && liq.LiquidDepth() == 8 {
				c.SetBlock(x, y, z, 1, before)
			}
		}
	}

	viewers := slices.Clone(c.v)
	c.Unlock()

	for _, viewer := range viewers {
		viewer.ViewBlockUpdate(pos, b, 0)
	}

	if !opts.DisableBlockUpdates {
		w.doBlockUpdatesAround(pos)
	}
}

// SetBiome sets the biome at the position passed. If a chunk is not yet loaded at that position, the chunk is
// first loaded or generated if it could not be found in the world save.
func (w *World) SetBiome(pos cube.Pos, b Biome) {
	if w == nil || pos.OutOfBounds(w.Range()) {
		// Fast way out.
		return
	}
	c := w.chunk(chunkPosFromBlockPos(pos))
	defer c.Unlock()
	c.SetBiome(uint8(pos[0]), int16(pos[1]), uint8(pos[2]), uint32(b.EncodeBiome()))
}

// BuildStructure builds a Structure passed at a specific position in the world. Unlike SetBlock, it takes a
// Structure implementation, which provides blocks to be placed at a specific location.
// BuildStructure is specifically tinkered to be able to process a large batch of chunks simultaneously and
// will do so within much less time than separate SetBlock calls would.
// The method operates on a per-chunk basis, setting all blocks within a single chunk part of the structure
// before moving on to the next chunk.
func (w *World) BuildStructure(pos cube.Pos, s Structure) {
	if w == nil {
		return
	}
	dim := s.Dimensions()
	width, height, length := dim[0], dim[1], dim[2]
	maxX, maxY, maxZ := pos[0]+width, pos[1]+height, pos[2]+length

	for chunkX := pos[0] >> 4; chunkX <= maxX>>4; chunkX++ {
		for chunkZ := pos[2] >> 4; chunkZ <= maxZ>>4; chunkZ++ {
			// We approach this on a per-chunk basis, so that we can keep only one chunk in memory at a time
			// while not needing to acquire a new chunk lock for every block. This also allows us not to send
			// block updates, but instead send a single chunk update once.
			chunkPos := ChunkPos{int32(chunkX), int32(chunkZ)}
			c := w.chunk(chunkPos)
			f := func(x, y, z int) Block {
				actual := cube.Pos{pos[0] + x, pos[1] + y, pos[2] + z}
				if actual[0]>>4 == chunkX && actual[2]>>4 == chunkZ {
					return w.blockInChunk(c, actual)
				}
				return w.Block(actual)
			}
			baseX, baseZ := chunkX<<4, chunkZ<<4
			subs := c.Sub()
			for i, sub := range subs {
				baseY := (i + (w.Range()[0] >> 4)) << 4
				if baseY>>4 < pos[1]>>4 {
					continue
				} else if baseY >= maxY {
					break
				}

				for localY := 0; localY < 16; localY++ {
					yOffset := baseY + localY
					if yOffset > w.Range()[1] || yOffset >= maxY {
						// We've hit the height limit for blocks.
						break
					} else if yOffset < w.Range()[0] || yOffset < pos[1] {
						// We've got a block below the minimum, but other blocks might still reach above
						// it, so don't break but continue.
						continue
					}
					for localX := 0; localX < 16; localX++ {
						xOffset := baseX + localX
						if xOffset < pos[0] || xOffset >= maxX {
							continue
						}
						for localZ := 0; localZ < 16; localZ++ {
							zOffset := baseZ + localZ
							if zOffset < pos[2] || zOffset >= maxZ {
								continue
							}
							b, liq := s.At(xOffset-pos[0], yOffset-pos[1], zOffset-pos[2], f)
							if b != nil {
								rid := BlockRuntimeID(b)
								sub.SetBlock(uint8(xOffset), uint8(yOffset), uint8(zOffset), 0, rid)

								if nbtBlocks[rid] {
									c.e[pos] = b
								} else {
									delete(c.e, pos)
								}
							}
							if liq != nil {
								sub.SetBlock(uint8(xOffset), uint8(yOffset), uint8(zOffset), 1, BlockRuntimeID(liq))
							} else if len(sub.Layers()) > 1 {
								sub.SetBlock(uint8(xOffset), uint8(yOffset), uint8(zOffset), 1, airRID)
							}
						}
					}
				}
			}
			// After setting all blocks of the structure within a single chunk, we show the new chunk to all
			// viewers once, and unlock it.
			for _, viewer := range c.v {
				viewer.ViewChunk(chunkPos, c.Chunk, c.e)
			}
			c.Unlock()
		}
	}
}

// Liquid attempts to return any liquid block at the position passed. This liquid may be in the foreground or
// in any other layer.
// If found, the liquid is returned. If not, the bool returned is false and the liquid is nil.
func (w *World) Liquid(pos cube.Pos) (Liquid, bool) {
	if w == nil || pos.OutOfBounds(w.Range()) {
		// Fast way out.
		return nil, false
	}
	c := w.chunk(chunkPosFromBlockPos(pos))
	defer c.Unlock()
	x, y, z := uint8(pos[0]), int16(pos[1]), uint8(pos[2])

	id := c.Block(x, y, z, 0)
	b, ok := BlockByRuntimeID(id)
	if !ok {
		w.log.Errorf("failed getting liquid: cannot get block by runtime ID %v", id)
		return nil, false
	}
	if liq, ok := b.(Liquid); ok {
		return liq, true
	}
	id = c.Block(x, y, z, 1)

	b, ok = BlockByRuntimeID(id)
	if !ok {
		w.log.Errorf("failed getting liquid: cannot get block by runtime ID %v", id)
		return nil, false
	}
	liq, ok := b.(Liquid)
	return liq, ok
}

// SetLiquid sets the liquid at a specific position in the world. Unlike SetBlock, SetLiquid will not
// overwrite any existing blocks. It will instead be in the same position as a block currently there, unless
// there already is a liquid at that position, in which case it will be overwritten.
// If nil is passed for the liquid, any liquid currently present will be removed.
func (w *World) SetLiquid(pos cube.Pos, b Liquid) {
	if w == nil || pos.OutOfBounds(w.Range()) {
		// Fast way out.
		return
	}
	chunkPos := chunkPosFromBlockPos(pos)
	c := w.chunk(chunkPos)
	if b == nil {
		w.removeLiquids(c, pos)
		c.Unlock()
		w.doBlockUpdatesAround(pos)
		return
	}
	x, y, z := uint8(pos[0]), int16(pos[1]), uint8(pos[2])
	if !replaceable(w, c, pos, b) {
		if displacer, ok := w.blockInChunk(c, pos).(LiquidDisplacer); !ok || !displacer.CanDisplace(b) {
			c.Unlock()
			return
		}
	}
	rid := BlockRuntimeID(b)
	if w.removeLiquids(c, pos) {
		c.SetBlock(x, y, z, 0, rid)
		for _, v := range c.v {
			v.ViewBlockUpdate(pos, b, 0)
		}
	} else {
		c.SetBlock(x, y, z, 1, rid)
		for _, v := range c.v {
			v.ViewBlockUpdate(pos, b, 1)
		}
	}
	c.Unlock()

	w.doBlockUpdatesAround(pos)
}

// removeLiquids removes any liquid blocks that may be present at a specific block position in the chunk
// passed.
// The bool returned specifies if no blocks were left on the foreground layer.
func (w *World) removeLiquids(c *chunkData, pos cube.Pos) bool {
	x, y, z := uint8(pos[0]), int16(pos[1]), uint8(pos[2])

	noneLeft := false
	if noLeft, changed := w.removeLiquidOnLayer(c.Chunk, x, y, z, 0); noLeft {
		if changed {
			for _, v := range c.v {
				v.ViewBlockUpdate(pos, air(), 0)
			}
		}
		noneLeft = true
	}
	if _, changed := w.removeLiquidOnLayer(c.Chunk, x, y, z, 1); changed {
		for _, v := range c.v {
			v.ViewBlockUpdate(pos, air(), 1)
		}
	}
	return noneLeft
}

// removeLiquidOnLayer removes a liquid block from a specific layer in the chunk passed, returning true if
// successful.
func (w *World) removeLiquidOnLayer(c *chunk.Chunk, x uint8, y int16, z, layer uint8) (bool, bool) {
	id := c.Block(x, y, z, layer)

	b, ok := BlockByRuntimeID(id)
	if !ok {
		w.log.Errorf("failed removing liquids: cannot get block by runtime ID %v", id)
		return false, false
	}
	if _, ok := b.(Liquid); ok {
		c.SetBlock(x, y, z, layer, airRID)
		return true, true
	}
	return id == airRID, false
}

// additionalLiquid checks if the block at a position has additional liquid on another layer and returns the
// liquid if so.
func (w *World) additionalLiquid(pos cube.Pos) (Liquid, bool) {
	if pos.OutOfBounds(w.Range()) {
		// Fast way out.
		return nil, false
	}
	c := w.chunk(chunkPosFromBlockPos(pos))
	id := c.Block(uint8(pos[0]), int16(pos[1]), uint8(pos[2]), 1)
	c.Unlock()
	b, ok := BlockByRuntimeID(id)
	if !ok {
		w.log.Errorf("failed getting liquid: cannot get block by runtime ID %v", id)
		return nil, false
	}
	liq, ok := b.(Liquid)
	return liq, ok
}

// Light returns the light level at the position passed. This is the highest of the sky and block light.
// The light value returned is a value in the range 0-15, where 0 means there is no light present, whereas
// 15 means the block is fully lit.
func (w *World) Light(pos cube.Pos) uint8 {
	if w == nil || pos[1] < w.Range()[0] {
		// Fast way out.
		return 0
	}
	if pos[1] > w.Range()[1] {
		// Above the rest of the world, so full skylight.
		return 15
	}
	c := w.chunk(chunkPosFromBlockPos(pos))
	defer c.Unlock()
	return c.Light(uint8(pos[0]), int16(pos[1]), uint8(pos[2]))
}

// SkyLight returns the skylight level at the position passed. This light level is not influenced by blocks
// that emit light, such as torches or glowstone. The light value, similarly to Light, is a value in the
// range 0-15, where 0 means no light is present.
func (w *World) SkyLight(pos cube.Pos) uint8 {
	if w == nil || pos[1] < w.Range()[0] {
		// Fast way out.
		return 0
	}
	if pos[1] > w.Range()[1] {
		// Above the rest of the world, so full skylight.
		return 15
	}
	c := w.chunk(chunkPosFromBlockPos(pos))
	defer c.Unlock()
	return c.SkyLight(uint8(pos[0]), int16(pos[1]), uint8(pos[2]))
}

// Time returns the current time of the world. The time is incremented every 1/20th of a second, unless
// World.StopTime() is called.
func (w *World) Time() int {
	if w == nil {
		return 0
	}
	w.set.Lock()
	defer w.set.Unlock()
	return int(w.set.Time)
}

// SetTime sets the new time of the world. SetTime will always work, regardless of whether the time is stopped
// or not.
func (w *World) SetTime(new int) {
	if w == nil {
		return
	}
	w.set.Lock()
	w.set.Time = int64(new)
	w.set.Unlock()

	viewers, _ := w.allViewers()
	for _, viewer := range viewers {
		viewer.ViewTime(new)
	}
}

// StopTime stops the time in the world. When called, the time will no longer cycle and the world will remain
// at the time when StopTime is called. The time may be restarted by calling World.StartTime().
// StopTime will not do anything if the time is already stopped.
func (w *World) StopTime() {
	w.enableTimeCycle(false)
}

// StartTime restarts the time in the world. When called, the time will start cycling again and the day/night
// cycle will continue. The time may be stopped again by calling World.StopTime().
// StartTime will not do anything if the time is already started.
func (w *World) StartTime() {
	w.enableTimeCycle(true)
}

// enableTimeCycle enables or disables the time cycling of the World.
func (w *World) enableTimeCycle(v bool) {
	if w == nil {
		return
	}
	w.set.Lock()
	defer w.set.Unlock()
	w.set.TimeCycle = v
}

// Temperature returns the temperature in the World at a specific position. Higher altitudes and different biomes
// influence the temperature returned.
func (w *World) Temperature(pos cube.Pos) float64 {
	const (
		tempDrop = 1.0 / 600
		seaLevel = 64
	)
	diff := pos[1] - seaLevel
	if diff < 0 {
		diff = 0
	}
	return w.Biome(pos).Temperature() - float64(diff)*tempDrop
}

// AddParticle spawns a particle at a given position in the world. Viewers that are viewing the chunk will be
// shown the particle.
func (w *World) AddParticle(pos mgl64.Vec3, p Particle) {
	if w == nil {
		return
	}
	p.Spawn(w, pos)
	for _, viewer := range w.Viewers(pos) {
		viewer.ViewParticle(pos, p)
	}
}

// PlaySound plays a sound at a specific position in the world. Viewers of that position will be able to hear
// the sound if they're close enough.
func (w *World) PlaySound(pos mgl64.Vec3, s Sound) {
	ctx := event.C()
	w.Handler().HandleSound(ctx, s, pos)
	ctx.Continue(func() {
		for _, viewer := range w.Viewers(pos) {
			viewer.ViewSound(pos, s)
		}
	})
}

var (
	worldsMu sync.RWMutex
	// entityWorlds holds a list of all entities added to a world. It may be used to look up the world that an
	// entity is currently in.
	entityWorlds = map[Entity]*World{}
)

// AddEntity adds an entity to the world at the position that the entity has. The entity will be visible to
// all viewers of the world that have the chunk of the entity loaded.
// If the chunk that the entity is in is not yet loaded, it will first be loaded.
// If the entity passed to AddEntity is currently in a world, it is first removed from that world.
func (w *World) AddEntity(e Entity) {
	if w == nil {
		return
	}
	// Remove the Entity from any previous World it might be in.
	e.World().RemoveEntity(e)

	worldsMu.Lock()
	entityWorlds[e] = w
	worldsMu.Unlock()

	chunkPos := chunkPosFromVec3(e.Position())
	w.entityMu.Lock()
	w.entities[e] = chunkPos
	w.entityMu.Unlock()

	c := w.chunk(chunkPos)
	c.entities = append(c.entities, e)
	viewers := slices.Clone(c.v)
	c.Unlock()

	for _, v := range viewers {
		// We show the entity to all viewers currently in the chunk that the entity is spawned in.
		showEntity(e, v)
	}

	w.Handler().HandleEntitySpawn(e)
}

// RemoveEntity removes an entity from the world that is currently present in it. Any viewers of the entity
// will no longer be able to see it.
// RemoveEntity operates assuming the position of the entity is the same as where it is currently in the
// world. If it can not find it there, it will loop through all entities and try to find it.
// RemoveEntity assumes the entity is currently loaded and in a loaded chunk. If not, the function will not do
// anything.
func (w *World) RemoveEntity(e Entity) {
	if w == nil {
		return
	}
	w.entityMu.Lock()
	chunkPos, found := w.entities[e]
	w.entityMu.Unlock()
	if !found {
		// The entity currently isn't in this world.
		return
	}

	w.Handler().HandleEntityDespawn(e)

	worldsMu.Lock()
	delete(entityWorlds, e)
	worldsMu.Unlock()

	c, ok := w.chunkFromCache(chunkPos)
	if !ok {
		// The chunk wasn't loaded, so we can't remove any entity from the chunk.
		return
	}
	c.Lock()
	c.entities = sliceutil.DeleteVal(c.entities, e)
	viewers := slices.Clone(c.v)
	c.Unlock()

	w.entityMu.Lock()
	delete(w.entities, e)
	w.entityMu.Unlock()

	for _, v := range viewers {
		v.HideEntity(e)
	}
}

// EntitiesWithin does a lookup through the entities in the chunks touched by the AABB passed, returning all
// those which are contained within the AABB when it comes to their position.
func (w *World) EntitiesWithin(aabb physics.AABB, ignored func(Entity) bool) []Entity {
	if w == nil {
		return nil
	}
	// Make an estimate of 16 entities on average.
	m := make([]Entity, 0, 16)

	minPos, maxPos := chunkPosFromVec3(aabb.Min()), chunkPosFromVec3(aabb.Max())

	for x := minPos[0]; x <= maxPos[0]; x++ {
		for z := minPos[1]; z <= maxPos[1]; z++ {
			c, ok := w.chunkFromCache(ChunkPos{x, z})
			if !ok {
				// The chunk wasn't loaded, so there are no entities here.
				continue
			}
			c.Lock()
			entities := slices.Clone(c.entities)
			c.Unlock()

			for _, entity := range entities {
				if ignored != nil && ignored(entity) {
					continue
				}
				if aabb.Vec3Within(entity.Position()) {
					// The entity position was within the AABB, so we add it to the slice to return.
					m = append(m, entity)
				}
			}
		}
	}
	return m
}

// Entities returns a list of all entities currently added to the World.
func (w *World) Entities() []Entity {
	if w == nil {
		return nil
	}
	w.entityMu.RLock()
	defer w.entityMu.RUnlock()
	m := make([]Entity, 0, len(w.entities))
	for e := range w.entities {
		m = append(m, e)
	}
	return m
}

// OfEntity attempts to return a world that an entity is currently in. If the entity was not currently added
// to a world, the world returned is nil and the bool returned is false.
func OfEntity(e Entity) (*World, bool) {
	worldsMu.RLock()
	w, ok := entityWorlds[e]
	worldsMu.RUnlock()
	return w, ok
}

// Spawn returns the spawn of the world. Every new player will by default spawn on this position in the world
// when joining.
func (w *World) Spawn() cube.Pos {
	if w == nil {
		return cube.Pos{}
	}
	w.set.Lock()
	s := w.set.Spawn
	w.set.Unlock()
	if s[1] > w.Range()[1] {
		s[1] = w.highestObstructingBlock(s[0], s[2]) + 1
	}
	return s
}

// SetSpawn sets the spawn of the world to a different position. The player will be spawned in the center of
// this position when newly joining.
func (w *World) SetSpawn(pos cube.Pos) {
	if w == nil {
		return
	}
	w.set.Lock()
	w.set.Spawn = pos
	w.set.Unlock()

	viewers, _ := w.allViewers()
	for _, viewer := range viewers {
		viewer.ViewWorldSpawn(pos)
	}
}

// DefaultGameMode returns the default game mode of the world. When players join, they are given this game
// mode.
// The default game mode may be changed using SetDefaultGameMode().
func (w *World) DefaultGameMode() GameMode {
	if w == nil {
		return GameModeSurvival
	}
	w.set.Lock()
	defer w.set.Unlock()
	return w.set.DefaultGameMode
}

// SetTickRange sets the range in chunks around each Viewer that will have the chunks (their blocks and entities)
// ticked when the World is ticked.
func (w *World) SetTickRange(v int) {
	if w == nil {
		return
	}
	w.set.Lock()
	defer w.set.Unlock()
	w.set.TickRange = int32(v)
}

// tickRange returns the tick range around each Viewer.
func (w *World) tickRange() int {
	w.set.Lock()
	defer w.set.Unlock()
	return int(w.set.TickRange)
}

// SetDefaultGameMode changes the default game mode of the world. When players join, they are then given that
// game mode.
func (w *World) SetDefaultGameMode(mode GameMode) {
	if w == nil {
		return
	}
	w.set.Lock()
	defer w.set.Unlock()
	w.set.DefaultGameMode = mode
}

// Difficulty returns the difficulty of the world. Properties of mobs in the world and the player's hunger
// will depend on this difficulty.
func (w *World) Difficulty() Difficulty {
	if w == nil {
		return DifficultyNormal
	}
	w.set.Lock()
	defer w.set.Unlock()
	return w.set.Difficulty
}

// SetDifficulty changes the difficulty of a world.
func (w *World) SetDifficulty(d Difficulty) {
	if w == nil {
		return
	}
	w.set.Lock()
	defer w.set.Unlock()
	w.set.Difficulty = d
}

// SetRandomTickSpeed sets the random tick speed of blocks. By default, each sub chunk has 3 blocks randomly
// ticked per sub chunk, so the default value is 3. Setting this value to 0 will stop random ticking
// altogether, while setting it higher results in faster ticking.
func (w *World) SetRandomTickSpeed(v int) {
	if w == nil {
		return
	}
	w.randomTickSpeed.Store(uint32(v))
}

// ScheduleBlockUpdate schedules a block update at the position passed after a specific delay. If the block at
// that position does not handle block updates, nothing will happen.
func (w *World) ScheduleBlockUpdate(pos cube.Pos, delay time.Duration) {
	if w == nil || pos.OutOfBounds(w.Range()) {
		return
	}
	w.updateMu.Lock()
	defer w.updateMu.Unlock()
	if _, exists := w.scheduledUpdates[pos]; exists {
		return
	}
	w.set.Lock()
	t := w.set.CurrentTick
	w.set.Unlock()

	w.scheduledUpdates[pos] = t + delay.Nanoseconds()/int64(time.Second/20)
}

// doBlockUpdatesAround schedules block updates directly around and on the position passed.
func (w *World) doBlockUpdatesAround(pos cube.Pos) {
	if w == nil || pos.OutOfBounds(w.Range()) {
		return
	}

	changed := pos

	w.updateMu.Lock()
	w.updateNeighbour(pos, changed)
	pos.Neighbours(func(pos cube.Pos) {
		w.updateNeighbour(pos, changed)
	}, w.Range())
	w.updateMu.Unlock()
}

// neighbourUpdate represents a position that needs to be updated because of a neighbour that changed.
type neighbourUpdate struct {
	pos, neighbour cube.Pos
}

// updateNeighbour ticks the position passed as a result of the neighbour passed being updated.
func (w *World) updateNeighbour(pos, changedNeighbour cube.Pos) {
	w.neighbourUpdates = append(w.neighbourUpdates, neighbourUpdate{pos: pos, neighbour: changedNeighbour})
}

// Provider changes the provider of the world to the provider passed. If nil is passed, the NopProvider
// will be set, which does not read or write any data.
func (w *World) Provider(p Provider) {
	if w == nil {
		return
	}
	if p == nil {
		p = NopProvider{}
	}

	w.set.Lock()
	defer w.set.Unlock()

	p.Settings(w.set)
	w.prov.Store(p)

	w.initChunkCache()
}

// ReadOnly makes the world read only. Chunks will no longer be saved to disk, just like entities and data
// in the level.dat.
func (w *World) ReadOnly() {
	if w == nil {
		return
	}
	w.rdonly.Store(true)
}

// Generator changes the generator of the world to the one passed. If nil is passed, the generator is set to
// the default, NopGenerator.
func (w *World) Generator(g Generator) {
	if w == nil {
		return
	}
	if g == nil {
		g = NopGenerator{}
	}
	w.gen.Store(g)
}

// Handle changes the current Handler of the world. As a result, events called by the world will call
// handlers of the Handler passed.
// Handle sets the world's Handler to NopHandler if nil is passed.
func (w *World) Handle(h Handler) {
	if w == nil {
		return
	}
	if h == nil {
		h = NopHandler{}
	}
	w.handler.Store(h)
}

// Viewers returns a list of all viewers viewing the position passed. A viewer will be assumed to be watching
// if the position is within one of the chunks that the viewer is watching.
func (w *World) Viewers(pos mgl64.Vec3) (viewers []Viewer) {
	if w == nil {
		return nil
	}
	c, ok := w.chunkFromCache(chunkPosFromVec3(pos))
	if !ok {
		return nil
	}
	c.Lock()
	defer c.Unlock()
	return slices.Clone(c.v)
}

// SetPortalDestinations sets the destination worlds for any nether and end portals in the World respectively. In order
// for either portals to work, SetPortalDestinations must first be called.
// Nil may be passed as destination to prevent the respective portal from transporting entities in it.
func (w *World) SetPortalDestinations(nether, end *World) {
	w.npd.Store(nether)
	w.epd.Store(end)
}

// PortalDestinations returns the destination worlds for nether and end portals respectively. Upon entering portals in
// this World, entities are moved to the respective destination worlds.
func (w *World) PortalDestinations() (nether, end *World) {
	return w.npd.Load(), w.epd.Load()
}

// Close closes the world and saves all chunks currently loaded.
func (w *World) Close() error {
	if w == nil {
		return nil
	}
	w.o.Do(w.close)
	return nil
}

// close stops the World from ticking, saves all chunks to the Provider and updates the world's settings.
func (w *World) close() {
	// Let user code run anything that needs to be finished before the World is closed.
	w.Handler().HandleClose()
	w.Handle(NopHandler{})

	close(w.closing)
	w.running.Wait()

	w.log.Debugf("Saving chunks in memory to disk...")

	w.chunkMu.Lock()
	w.lastChunk = nil
	toSave := maps.Clone(w.chunks)
	maps.Clear(w.chunks)
	w.chunkMu.Unlock()

	for pos, c := range toSave {
		w.saveChunk(pos, c)
	}

	if !w.rdonly.Load() {
		w.log.Debugf("Updating level.dat values...")

		w.set.ref.Dec()
		w.provider().SaveSettings(w.set)
	}

	w.log.Debugf("Closing provider...")
	if err := w.provider().Close(); err != nil {
		w.log.Errorf("error closing world provider: %v", err)
	}
}

// allViewers returns a list of all loaders of the world, regardless of where in the world they are viewing.
func (w *World) allViewers() ([]Viewer, []*Loader) {
	w.viewersMu.Lock()
	defer w.viewersMu.Unlock()
	return maps.Values(w.viewers), maps.Keys(w.viewers)
}

// addWorldViewer adds a viewer to the world. Should only be used while the viewer isn't viewing any chunks.
func (w *World) addWorldViewer(l *Loader) {
	w.viewersMu.Lock()
	w.viewers[l] = l.viewer
	w.viewersMu.Unlock()
	l.viewer.ViewTime(w.Time())
	w.set.Lock()
	raining, thundering := w.set.Raining, w.set.Raining && w.set.Thundering
	w.set.Unlock()
	l.viewer.ViewWeather(raining, thundering)
	l.viewer.ViewWorldSpawn(w.Spawn())
}

// removeWorldViewer removes a viewer from the world. Should only be used while the viewer isn't viewing any chunks.
func (w *World) removeWorldViewer(l *Loader) {
	w.viewersMu.Lock()
	delete(w.viewers, l)
	w.viewersMu.Unlock()
}

// addViewer adds a viewer to the world at a given position. Any events that happen in the chunk at that
// position, such as block changes, entity changes etc., will be sent to the viewer.
func (w *World) addViewer(c *chunkData, loader *Loader) {
	if w == nil {
		return
	}
	c.v = append(c.v, loader.viewer)
	c.l = append(c.l, loader)

	entities := slices.Clone(c.entities)
	c.Unlock()

	for _, entity := range entities {
		showEntity(entity, loader.viewer)
	}
}

// removeViewer removes a viewer from the world at a given position. All entities will be hidden from the
// viewer and no more calls will be made when events in the chunk happen.
func (w *World) removeViewer(pos ChunkPos, loader *Loader) {
	if w == nil {
		return
	}
	c, ok := w.chunkFromCache(pos)
	if !ok {
		return
	}
	c.Lock()
	i := slices.Index(c.l, loader)
	c.v = slices.Delete(c.v, i, i+1)
	c.l = slices.Delete(c.l, i, i+1)
	e := slices.Clone(c.entities)
	c.Unlock()

	// After removing the loader from the chunk, we also need to hide all entities from the viewer.
	for _, entity := range e {
		loader.viewer.HideEntity(entity)
	}
}

// provider returns the provider of the world. It should always be used, rather than direct field access, in
// order to provide synchronisation safety.
func (w *World) provider() Provider {
	return w.prov.Load()
}

// Handler returns the Handler of the world. It should always be used, rather than direct field access, in
// order to provide synchronisation safety.
func (w *World) Handler() Handler {
	if w == nil {
		return NopHandler{}
	}
	return w.handler.Load()
}

// chunkFromCache attempts to fetch a chunk at the chunk position passed from the cache. If not found, the
// chunk returned is nil and false is returned.
func (w *World) chunkFromCache(pos ChunkPos) (*chunkData, bool) {
	w.chunkMu.Lock()
	c, ok := w.chunks[pos]
	w.chunkMu.Unlock()
	return c, ok
}

// showEntity shows an entity to a viewer of the world. It makes sure everything of the entity, including the
// items held, is shown.
func showEntity(e Entity, viewer Viewer) {
	viewer.ViewEntity(e)
	viewer.ViewEntityState(e)
	viewer.ViewEntityItems(e)
	viewer.ViewEntityArmour(e)
}

// chunk reads a chunk from the position passed. If a chunk at that position is not yet loaded, the chunk is
// loaded from the provider, or generated if it did not yet exist. Both of these actions are done
// synchronously.
// An error is returned if the chunk could not be loaded successfully.
// chunk locks the chunk returned, meaning that any call to chunk made at the same time has to wait until the
// user calls Chunk.Unlock() on the chunk returned.
func (w *World) chunk(pos ChunkPos) *chunkData {
	w.chunkMu.Lock()
	if pos == w.lastPos && w.lastChunk != nil {
		c := w.lastChunk
		w.chunkMu.Unlock()
		c.Lock()
		return c
	}
	c, ok := w.chunks[pos]
	if !ok {
		var err error
		c, err = w.loadChunk(pos)
		if err != nil {
			w.chunkMu.Unlock()
			w.log.Errorf("load chunk: failed loading %v: %v\n", pos, err)
			return c
		}
		chunk.LightArea([]*chunk.Chunk{c.Chunk}, int(pos[0]), int(pos[1])).Fill()
		c.Unlock()
		w.chunkMu.Lock()

		w.calculateLight(pos)
	}
	w.lastChunk, w.lastPos = c, pos
	w.chunkMu.Unlock()

	c.Lock()
	return c
}

// setChunk sets the chunk.Chunk passed at a specific ChunkPos without replacing any entities at that
// position.
//lint:ignore U1000 This method is explicitly present to be used using compiler directives.
func (w *World) setChunk(pos ChunkPos, c *chunk.Chunk, e map[cube.Pos]Block) {
	if w == nil {
		return
	}
	if e == nil {
		e = map[cube.Pos]Block{}
	}
	w.chunkMu.Lock()
	defer w.chunkMu.Unlock()

	data, ok := w.chunks[pos]
	if ok {
		data.Lock()
		defer data.Unlock()
	} else {
		data = newChunkData(c)
		w.chunks[pos] = data
	}
	data.e = e
}

// loadChunk attempts to load a chunk from the provider, or generates a chunk if one doesn't currently exist.
func (w *World) loadChunk(pos ChunkPos) (*chunkData, error) {
	c, found, err := w.provider().LoadChunk(pos)
	if err != nil {
		return newChunkData(chunk.New(airRID, w.d.Range())), err
	}

	if !found {
		// The provider doesn't have a chunk saved at this position, so we generate a new one.
		data := newChunkData(chunk.New(airRID, w.d.Range()))
		w.chunks[pos] = data
		data.Lock()
		w.chunkMu.Unlock()

		w.gen.Load().GenerateChunk(pos, data.Chunk)
		return data, nil
	}
	data := newChunkData(c)
	w.chunks[pos] = data

	ent, err := w.provider().LoadEntities(pos)
	if err != nil {
		return nil, fmt.Errorf("error loading entities of chunk %v: %w", pos, err)
	}
	data.entities = make([]Entity, 0, len(ent))

	// Iterate through the entities twice and make sure they're added to all relevant maps. Note that this iteration
	// happens twice to avoid having to lock both worldsMu and entityMu. This is intentional, to avoid deadlocks.
	worldsMu.Lock()
	for _, e := range ent {
		data.entities = append(data.entities, e)
		entityWorlds[e] = w
	}
	worldsMu.Unlock()

	w.entityMu.Lock()
	for _, e := range ent {
		w.entities[e] = pos
	}
	w.entityMu.Unlock()

	blockEntities, err := w.provider().LoadBlockNBT(pos)
	if err != nil {
		return nil, fmt.Errorf("error loading block entities of chunk %v: %w", pos, err)
	}
	w.loadIntoBlocks(data, blockEntities)

	data.Lock()
	w.chunkMu.Unlock()
	return data, nil
}

// calculateLight calculates the light in the chunk passed and spreads the light of any of the surrounding
// neighbours if they have all chunks loaded around it as a result of the one passed.
func (w *World) calculateLight(centre ChunkPos) {
	for x := int32(-1); x <= 1; x++ {
		for z := int32(-1); z <= 1; z++ {
			// For all the neighbours of this chunk, if they exist, check if all neighbours of that chunk
			// now exist because of this one.
			pos := ChunkPos{centre[0] + x, centre[1] + z}
			if _, ok := w.chunks[pos]; ok {
				// Attempt to spread the light of all neighbours into the ones surrounding them.
				w.spreadLight(pos)
			}
		}
	}
}

// spreadLight spreads the light from the chunk passed at the position passed to all neighbours if each of
// them is loaded.
func (w *World) spreadLight(pos ChunkPos) {
	chunks := make([]*chunk.Chunk, 0, 9)
	for z := int32(-1); z <= 1; z++ {
		for x := int32(-1); x <= 1; x++ {
			neighbour, ok := w.chunks[ChunkPos{pos[0] + x, pos[1] + z}]
			if !ok {
				// Not all surrounding chunks existed: Stop spreading light as we can't do it completely yet.
				return
			}
			chunks = append(chunks, neighbour.Chunk)
		}
	}
	for _, neighbour := range chunks {
		neighbour.Lock()
	}
	// All chunks of the current one are present, so we can spread the light from this chunk
	// to all chunks.
	chunk.LightArea(chunks, int(pos[0])-1, int(pos[1])-1).Spread()
	for _, neighbour := range chunks {
		neighbour.Unlock()
	}
}

// loadIntoBlocks loads the block entity data passed into blocks located in a specific chunk. The blocks that
// have NBT will then be stored into memory.
func (w *World) loadIntoBlocks(c *chunkData, blockEntityData []map[string]any) {
	c.e = make(map[cube.Pos]Block, len(blockEntityData))
	for _, data := range blockEntityData {
		pos := blockPosFromNBT(data)

		id := c.Block(uint8(pos[0]), int16(pos[1]), uint8(pos[2]), 0)
		b, ok := BlockByRuntimeID(id)
		if !ok {
			w.log.Errorf("error loading block entity data: could not find block state by runtime ID %v", id)
			continue
		}
		if nbt, ok := b.(NBTer); ok {
			b = nbt.DecodeNBT(data).(Block)
		}
		c.e[pos] = b
	}
}

// saveChunk is called when a chunk is removed from the cache. We first compact the chunk, then we write it to
// the provider.
func (w *World) saveChunk(pos ChunkPos, c *chunkData) {
	c.Lock()
	// We allocate a new map for all block entities.
	m := make([]map[string]any, 0, len(c.e))
	for pos, b := range c.e {
		if n, ok := b.(NBTer); ok {
			// Encode the block entities and add the 'x', 'y' and 'z' tags to it.
			data := n.EncodeNBT()
			data["x"], data["y"], data["z"] = int32(pos[0]), int32(pos[1]), int32(pos[2])
			m = append(m, data)
		}
	}
	if !w.rdonly.Load() {
		c.Compact()
		if err := w.provider().SaveChunk(pos, c.Chunk); err != nil {
			w.log.Errorf("error saving chunk %v to provider: %v", pos, err)
		}
		s := make([]SaveableEntity, 0, len(c.entities))
		for _, e := range c.entities {
			if saveable, ok := e.(SaveableEntity); ok {
				s = append(s, saveable)
			}
		}
		if err := w.provider().SaveEntities(pos, s); err != nil {
			w.log.Errorf("error saving entities in chunk %v to provider: %v", pos, err)
		}
		if err := w.provider().SaveBlockNBT(pos, m); err != nil {
			w.log.Errorf("error saving block NBT in chunk %v to provider: %v", pos, err)
		}
	}
	ent := c.entities
	c.entities = nil
	c.Unlock()

	for _, e := range ent {
		_ = e.Close()
	}
}

// initChunkCache initialises the chunk cache of the world to its default values.
func (w *World) initChunkCache() {
	w.chunkMu.Lock()
	w.chunks = make(map[ChunkPos]*chunkData)
	w.chunkMu.Unlock()
}

// chunkCacheJanitor runs until the world is running, cleaning chunks that are no longer in use from the cache.
func (w *World) chunkCacheJanitor() {
	t := time.NewTicker(time.Minute * 5)
	defer t.Stop()

	w.running.Add(1)
	chunksToRemove := map[ChunkPos]*chunkData{}
	for {
		select {
		case <-t.C:
			w.chunkMu.Lock()
			for pos, c := range w.chunks {
				c.Lock()
				v := len(c.v)
				c.Unlock()
				if v == 0 {
					chunksToRemove[pos] = c
					delete(w.chunks, pos)
					if w.lastPos == pos {
						w.lastChunk = nil
					}
				}
			}
			w.chunkMu.Unlock()

			for pos, c := range chunksToRemove {
				w.saveChunk(pos, c)
				delete(chunksToRemove, pos)
			}
		case <-w.closing:
			w.running.Done()
			return
		}
	}
}

// chunkData represents the data of a chunk including the block entities and loaders. This data is protected
// by the mutex present in the chunk.Chunk held.
type chunkData struct {
	*chunk.Chunk
	e        map[cube.Pos]Block
	v        []Viewer
	l        []*Loader
	entities []Entity
}

// newChunkData returns a new chunkData wrapper around the chunk.Chunk passed.
func newChunkData(c *chunk.Chunk) *chunkData {
	return &chunkData{Chunk: c, e: map[cube.Pos]Block{}}
}
