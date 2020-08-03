package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/df-mc/dragonfly/dragonfly/world/particle"
	"github.com/go-gl/mathgl/mgl64"
)

// Sponge is a block that can be used to remove water around itself when placed, turning into a wet sponge in the
// process.
type Sponge struct {
	noNBT
	solid

	// Wet specifies whether the dry or the wet variant of the block is used.
	Wet bool
}

// BreakInfo ...
func (s Sponge) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    0.6,
		Drops:       simpleDrops(item.NewStack(s, 1)),
		Effective:   nothingEffective,
		Harvestable: alwaysHarvestable,
	}
}

// EncodeItem ...
func (s Sponge) EncodeItem() (id int32, meta int16) {
	if s.Wet {
		meta = 1
	}

	return 19, meta
}

// EncodeBlock ...
func (s Sponge) EncodeBlock() (name string, properties map[string]interface{}) {
	if s.Wet {
		return "minecraft:sponge", map[string]interface{}{"sponge_type": "wet"}
	}
	return "minecraft:sponge", map[string]interface{}{"sponge_type": "dry"}
}

// Hash ...
func (s Sponge) Hash() uint64 {
	return hashSponge | (uint64(boolByte(s.Wet)) << 32)
}

// UseOnBlock places the sponge, absorbs nearby water if it's still dry and flags it as wet if any water has been
// absorbed.
func (s Sponge) UseOnBlock(pos world.BlockPos, face world.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(w, pos, face, s)
	if !used {
		return
	}

	place(w, pos, s, user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick checks for nearby water flow. If water could be found and the sponge is dry, it will absorb the
// water and be flagged as wet.
func (s Sponge) NeighbourUpdateTick(pos, _ world.BlockPos, w *world.World) {
	// The sponge is dry, so it can absorb nearby water.
	if !s.Wet {
		if s.absorbWater(pos, w) > 0 {
			// Water has been absorbed, so we flag the sponge as wet.
			s.setWet(pos, w)
		}
	}
}

// setWet flags a sponge as wet. It replaces the block at pos by a wet sponge block and displays a block break
// particle at the sponge's position with an offset of 0.5 on each axis.
func (s Sponge) setWet(pos world.BlockPos, w *world.World) {
	s.Wet = true
	w.SetBlock(pos, s)
	w.AddParticle(pos.Vec3().Add(mgl64.Vec3{0.5, 0.5, 0.5}), particle.BlockBreak{Block: Water{Depth: 1}})
}

// absorbWater replaces water blocks near the sponge by air out to a taxicab geometry of 7 in all directions.
// The maximum for absorbed blocks is 65.
// The returned int specifies the amount of replaced water blocks.
func (s Sponge) absorbWater(pos world.BlockPos, w *world.World) int {
	// distanceToSponge binds a world.BlockPos to its distance from the sponge's position.
	type distanceToSponge struct {
		block    world.BlockPos
		distance int32
	}

	queue := make([]distanceToSponge, 0)
	queue = append(queue, distanceToSponge{pos, 0})

	// A sponge can only absorb up to 65 water blocks.
	replaced := 0
	for replaced < 65 {
		if len(queue) == 0 {
			break
		}

		// Pop the next distanceToSponge entry from the queue.
		next := queue[0]
		queue = queue[1:]

		next.block.Neighbours(func(neighbour world.BlockPos) {
			liquid, found := w.Liquid(neighbour)
			if found {
				if _, isWater := liquid.(Water); isWater {
					w.SetLiquid(neighbour, nil)
					replaced++
					if next.distance < 7 {
						queue = append(queue, distanceToSponge{neighbour, next.distance + 1})
					}
				}
			}
		})
	}

	return replaced
}
