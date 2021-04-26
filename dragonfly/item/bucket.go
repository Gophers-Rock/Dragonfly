package item

import (
	"github.com/df-mc/dragonfly/dragonfly/block/cube"
	"github.com/df-mc/dragonfly/dragonfly/internal/item_internal"
	"github.com/df-mc/dragonfly/dragonfly/item/bucket"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/df-mc/dragonfly/dragonfly/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// Bucket is a tool used to carry water, lava, milk and fish.
type Bucket struct {
	// Content is the content that the bucket has. By default, this value resolves to an empty bucket.
	Content bucket.Content
}

// MaxCount returns 16.
func (b Bucket) MaxCount() int {
	if b.Empty() {
		return 16
	}
	return 1
}

// Empty returns true if the bucket is empty.
func (b Bucket) Empty() bool {
	return b.Content == bucket.Content{}
}

// UseOnBlock handles the bucket filling and emptying logic.
func (b Bucket) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, _ User, ctx *UseContext) bool {
	if b.Empty() {
		return b.fillFrom(pos, w, ctx)
	}

	var liq world.Liquid
	if b.Content == bucket.Water() {
		liq = item_internal.Water
	} else if b.Content == bucket.Lava() {
		liq = item_internal.Lava
	} else {
		return false
	}

	if bl := w.Block(pos); canDisplace(bl, liq) || replaceableWith(bl, liq) {
		w.SetLiquid(pos, liq)
	} else if bl := w.Block(pos.Side(face)); canDisplace(bl, liq) || replaceableWith(bl, liq) {
		w.SetLiquid(pos.Side(face), liq)
	} else {
		return false
	}

	w.PlaySound(pos.Vec3Centre(), sound.BucketEmpty{Liquid: liq})
	ctx.NewItem = NewStack(Bucket{}, 1)
	ctx.NewItemSurvivalOnly = true
	ctx.SubtractFromCount(1)
	return true
}

// fillFrom fills a bucket from the liquid at the position passed in the world. If there is no liquid or if
// the liquid is no source, fillFrom returns false.
func (b Bucket) fillFrom(pos cube.Pos, w *world.World, ctx *UseContext) bool {
	liquid, ok := w.Liquid(pos)
	if !ok {
		return false
	}
	if liquid.LiquidDepth() != 8 || liquid.LiquidFalling() {
		// Only allow picking up liquid source blocks.
		return false
	}
	w.SetLiquid(pos, nil)
	w.PlaySound(pos.Vec3Centre(), sound.BucketFill{Liquid: liquid})

	if item_internal.IsWater(liquid) {
		ctx.NewItem = NewStack(Bucket{Content: bucket.Water()}, 1)
	} else {
		ctx.NewItem = NewStack(Bucket{Content: bucket.Lava()}, 1)
	}
	ctx.NewItemSurvivalOnly = true
	ctx.SubtractFromCount(1)
	return true
}

// EncodeItem ...
func (b Bucket) EncodeItem() (id int32, meta int16) {
	switch b.Content {
	case bucket.Water():
		return 325, 8
	case bucket.Lava():
		return 325, 10
	}
	return 325, 0
}

type replaceable interface {
	ReplaceableBy(b world.Block) bool
}

func replaceableWith(b world.Block, with world.Block) bool {
	if r, ok := b.(replaceable); ok {
		return r.ReplaceableBy(with)
	}
	return false
}

func canDisplace(b world.Block, liq world.Liquid) bool {
	if d, ok := b.(world.LiquidDisplacer); ok {
		return d.CanDisplace(liq)
	}
	return false
}
