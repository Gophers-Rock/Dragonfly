package model

import (
	"github.com/df-mc/dragonfly/dragonfly/entity/physics"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
)

type Lantern struct{}

func (l Lantern) AABB(pos world.BlockPos, w *world.World) []physics.AABB {
	return []physics.AABB{physics.NewAABB(mgl64.Vec3{}, mgl64.Vec3{0.375, 0.5, 0.375})}
}

func (l Lantern) FaceSolid(pos world.BlockPos, face world.Face, w *world.World) bool {
	return false
}
