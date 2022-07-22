package world

import (
	"image/color"
	"sync"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type MapData struct {
	Pixels        [][]color.RGBA
	TrackEntities map[Entity]struct{}
	TrackBlocks   map[cube.Pos]struct{}
	// Scale should be 0 to 4. TODO: verify.
	Scale byte
	// Locked map should not be affected by world content (block) changes.
	Locked bool
}

type MapDataViewer interface {
	ViewMapDataChange(*ViewableMapData)
}

type ViewableMapData struct {
	mapID int64
	world *World

	viewersMu sync.RWMutex
	viewers   map[MapDataViewer]struct{}

	pixelsMu, trackEntitiesMu, trackBlocksMu sync.RWMutex

	data MapData
}

// ChangePixels broadcast *packet.ClientBoundMapItemData to viewers with packet.MapUpdateFlagTexture.
func (d *ViewableMapData) ChangePixels(pixels [][]color.RGBA) {
	d.pixelsMu.Lock()
	defer d.pixelsMu.Unlock()

	d.data.Pixels = pixels
	d.change(packet.MapUpdateFlagTexture)
}

// AddTrackEntity broadcast *packet.ClientBoundMapItemData to viewers with packet.MapUpdateFlagDecoration.
func (d *ViewableMapData) AddTrackEntity(e Entity) {
	d.trackEntitiesMu.Lock()
	defer d.trackEntitiesMu.Unlock()

	s := struct{}{}
	if d.data.TrackEntities == nil {
		d.data.TrackEntities = map[Entity]struct{}{e: s}
	} else {
		d.data.TrackEntities[e] = s
	}
	d.change(packet.MapUpdateFlagDecoration)
}

// RemoveTrackEntity broadcast *packet.ClientBoundMapItemData to viewers with packet.MapUpdateFlagDecoration.
func (d *ViewableMapData) RemoveTrackEntity(e Entity) {
	d.trackEntitiesMu.Lock()
	defer d.trackEntitiesMu.Unlock()

	if d.data.TrackEntities != nil {
		delete(d.data.TrackEntities, e)
		d.change(packet.MapUpdateFlagDecoration)
	}
}

// AddTrackBlock broadcast *packet.ClientBoundMapItemData to viewers with packet.MapUpdateFlagDecoration.
func (d *ViewableMapData) AddTrackBlock(pos cube.Pos) {
	d.trackBlocksMu.Lock()
	defer d.trackBlocksMu.Unlock()

	s := struct{}{}
	if d.data.TrackBlocks == nil {
		d.data.TrackBlocks = map[cube.Pos]struct{}{pos: s}
	} else {
		d.data.TrackBlocks[pos] = s
	}
	d.change(packet.MapUpdateFlagDecoration)
}

// RemoveTrackBlock broadcast *packet.ClientBoundMapItemData to viewers with packet.MapUpdateFlagDecoration.
func (d *ViewableMapData) RemoveTrackBlock(pos cube.Pos) {
	d.trackBlocksMu.Lock()
	defer d.trackBlocksMu.Unlock()

	if d.data.TrackBlocks != nil {
		delete(d.data.TrackBlocks, pos)
		d.change(packet.MapUpdateFlagDecoration)
	}
}

// GetMapData ...
func (d *ViewableMapData) GetMapData() MapData {
	d.pixelsMu.RLock()
	d.trackEntitiesMu.RLock()
	d.trackBlocksMu.RLock()
	defer d.pixelsMu.RUnlock()
	defer d.trackEntitiesMu.RUnlock()
	defer d.trackBlocksMu.RUnlock()

	return d.data
}

func (d *ViewableMapData) change(updateFlag byte) {
	d.broadcast(updateFlag)
	d.save()
}

func (d *ViewableMapData) broadcast(updateFlag byte) {
	d.viewersMu.RLock()
	defer d.viewersMu.RUnlock()

	for viewer := range d.viewers {
		viewer.ViewMapDataChange(d)
	}
}

func (d *ViewableMapData) save() {
	// TODO: save()
}

// AddViewer ...
func (d *ViewableMapData) AddViewer(v MapDataViewer) {
	m.viewersMu.Lock()
	defer d.viewersMu.Unlock()

	s := struct{}{}
	if d.viewers == nil {
		m.viewers = map[MapDataViewer]struct{}{v: s}
	} else {
		m.viewers[v] = s
	}
}

// RemoveViewer ...
func (d *ViewableMapData) RemoveViewer(v MapDataViewer) {
	m.viewersMu.Lock()
	defer d.viewersMu.Unlock()

	if d.viewers != nil {
		delete(m.viewers, v)
	}
}

// EncodeNBT provides value of field map ID, scale and is scaling for item.BaseMap.EncodeNBT().
// Returns empty map if nil.
func (d *ViewableMapData) EncodeItemNBT() map[string]any {
	if d == nil {
		return map[string]any{}
	}

	data := d.GetData()
	return map[string]any{
		"map_uuid":       d.mapID,
		"map_scale":      data.Scale,
		"map_is_scaling": data.Scale > 0,
	}
}

// GetDimension returns the dimension of map's belonging world.
func (d *ViewableMapData) GetDimension() world.Dimension {
	return d.world.Dimension()
}