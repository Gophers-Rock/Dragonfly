package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/entity/physics"
	"github.com/df-mc/dragonfly/server/entity/physics/trace"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"image/color"
	"time"
)

// SplashPotion is an item that grants effects when thrown.
type SplashPotion struct {
	transform
	yaw, pitch float64

	ticksLived int

	closeNextTick bool

	owner world.Entity

	t potion.Potion
	c *ProjectileComputer
}

// NewSplashPotion ...
func NewSplashPotion(pos mgl64.Vec3, yaw, pitch float64, owner world.Entity, t potion.Potion) *SplashPotion {
	s := &SplashPotion{
		yaw:   yaw,
		pitch: pitch,
		owner: owner,

		t: t,
		c: &ProjectileComputer{&MovementComputer{
			Gravity:           0.05,
			Drag:              0.01,
			DragBeforeGravity: true,
		}},
	}
	s.transform = newTransform(s, pos)

	return s
}

// Name ...
func (s *SplashPotion) Name() string {
	return "Splash Potion"
}

// EncodeEntity ...
func (s *SplashPotion) EncodeEntity() string {
	return "minecraft:splash_potion"
}

// AABB ...
func (s *SplashPotion) AABB() physics.AABB {
	return physics.NewAABB(mgl64.Vec3{-0.125, 0, -0.125}, mgl64.Vec3{0.125, 0.25, 0.125})
}

// Rotation ...
func (s *SplashPotion) Rotation() (float64, float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.yaw, s.pitch
}

// Type returns the type of potion the splash potion will grant effects for when thrown.
func (s *SplashPotion) Type() potion.Potion {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.t
}

// Tick ...
func (s *SplashPotion) Tick(current int64) {
	if s.closeNextTick {
		_ = s.Close()
		return
	}
	s.mu.Lock()
	m, result := s.c.TickMovement(s, s.pos, s.vel, s.yaw, s.pitch, s.ignores)
	s.pos, s.vel, s.yaw, s.pitch = m.pos, m.vel, m.yaw, m.pitch
	s.mu.Unlock()

	s.ticksLived++
	m.Send()

	if m.pos[1] < float64(s.World().Range()[0]) && current%10 == 0 {
		s.closeNextTick = true
		return
	}

	if result != nil {
		w := s.World()
		aabb := s.AABB().Translate(m.pos)

		effects := s.t.Effects()
		colour := color.RGBA{R: 0x38, G: 0x5d, B: 0xc6, A: 0xff}

		if len(effects) > 0 {
			colour, _ = effect.ResultingColour(effects)

			ignore := func(entity world.Entity) bool {
				_, canSplash := entity.(Living)
				return !canSplash || entity == s
			}

			for _, otherEntity := range w.EntitiesWithin(aabb.GrowVec3(mgl64.Vec3{8.25, 4.25, 8.25}), ignore) {
				splashEntity := otherEntity.(Living)
				if !splashEntity.AABB().Translate(splashEntity.Position()).IntersectsWith(aabb.GrowVec3(mgl64.Vec3{4.125, 2.125, 4.125})) {
					continue
				}

				dist := world.Distance(splashEntity.Position(), m.pos)
				if dist > 4 {
					continue
				}

				distFactor := 1 - dist/4
				if entityResult, ok := result.(trace.EntityResult); ok && entityResult.Entity() == otherEntity {
					distFactor = 1
				}

				for _, eff := range effects {
					if potentEff, ok := eff.Type().(effect.PotentType); ok {
						splashEntity.AddEffect(effect.NewInstant(potentEff.WithPotency(distFactor), eff.Level()))
						continue
					}

					dur := time.Duration(float64(eff.Duration()) * 0.75 * distFactor)
					if dur < time.Second {
						continue
					}
					splashEntity.AddEffect(effect.New(eff.Type().(effect.LastingType), eff.Level(), dur))
				}
			}
		} else if s.t.Equals(potion.Water()) {
			switch blockResult := result.(type) {
			case trace.BlockResult:
				blockPos := blockResult.BlockPosition().Side(blockResult.Face())
				if w.Block(blockPos) == fire() {
					w.SetBlock(blockPos, air())
				}

				for _, f := range cube.HorizontalFaces() {
					horizontalPos := blockPos.Side(f)
					if w.Block(horizontalPos) == fire() {
						w.SetBlock(horizontalPos, air())
					}
				}
			case trace.EntityResult:
				// TODO: Damage endermen, blazes, striders and snow golems when implemented and rehydrate axolotls.
			}
		}

		w.AddParticle(m.pos, particle.Splash{Colour: colour})
		w.PlaySound(m.pos, sound.GlassBreak{})

		s.closeNextTick = true
	}
}

// ignores returns whether the SplashPotion should ignore collision with the entity passed.
func (s *SplashPotion) ignores(entity world.Entity) bool {
	_, ok := entity.(Living)
	return !ok || entity == s || (s.ticksLived < 5 && entity == s.owner)
}

// New creates a SplashPotion with the position, velocity, yaw, and pitch provided. It doesn't spawn the SplashPotion,
// only returns it.
func (s *SplashPotion) New(pos, vel mgl64.Vec3, yaw, pitch float64, t potion.Potion) world.Entity {
	splash := NewSplashPotion(pos, yaw, pitch, nil, t)
	splash.vel = vel
	return splash
}

// Owner ...
func (s *SplashPotion) Owner() world.Entity {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.owner
}

// Own ...
func (s *SplashPotion) Own(owner world.Entity) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.owner = owner
}

// DecodeNBT decodes the properties in a map to a SplashPotion and returns a new SplashPotion entity.
func (s *SplashPotion) DecodeNBT(data map[string]interface{}) interface{} {
	return s.New(
		nbtconv.MapVec3(data, "Pos"),
		nbtconv.MapVec3(data, "Motion"),
		float64(nbtconv.MapFloat32(data, "Yaw")),
		float64(nbtconv.MapFloat32(data, "Pitch")),
		potion.From(nbtconv.MapInt32(data, "PotionId")),
	)
}

// EncodeNBT encodes the SplashPotion entity's properties as a map and returns it.
func (s *SplashPotion) EncodeNBT() map[string]interface{} {
	yaw, pitch := s.Rotation()
	return map[string]interface{}{
		"Pos":      nbtconv.Vec3ToFloat32Slice(s.Position()),
		"Yaw":      yaw,
		"Pitch":    pitch,
		"Motion":   nbtconv.Vec3ToFloat32Slice(s.Velocity()),
		"Damage":   0.0,
		"PotionId": s.t.Uint8(),
	}
}

// air returns an air block.
func air() world.Block {
	f, ok := world.BlockByName("minecraft:air", map[string]interface{}{})
	if !ok {
		panic("could not find air block")
	}
	return f
}
