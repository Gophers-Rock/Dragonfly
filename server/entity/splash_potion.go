package entity

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// SplashPotion is an item that grants effects when thrown.
type SplashPotion struct {
	splashable
	transform

	age   int
	close bool

	owner world.Entity

	c *ProjectileComputer
}

// NewSplashPotion ...
func NewSplashPotion(pos mgl64.Vec3, owner world.Entity, t potion.Potion) *SplashPotion {
	s := &SplashPotion{
		owner: owner,

		splashable: splashable{t: t, m: 0.75},
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

// BBox ...
func (s *SplashPotion) BBox() cube.BBox {
	return cube.Box(-0.125, 0, -0.125, 0.125, 0.25, 0.125)
}

// Tick ...
func (s *SplashPotion) Tick(w *world.World, current int64) {
	if s.close {
		_ = s.Close()
		return
	}
	s.mu.Lock()
	m, result := s.c.TickMovement(s, s.pos, s.vel, 0, 0, s.ignores)
	s.pos, s.vel = m.pos, m.vel
	s.mu.Unlock()

	s.age++
	m.Send()

	if m.pos[1] < float64(w.Range()[0]) && current%10 == 0 {
		s.close = true
		return
	}

	if result != nil {
		s.splash(s, w, m.pos, result, s.BBox())
		s.close = true
	}
}

// ignores returns whether the SplashPotion should ignore collision with the entity passed.
func (s *SplashPotion) ignores(entity world.Entity) bool {
	_, ok := entity.(Living)
	return !ok || entity == s || (s.age < 5 && entity == s.owner)
}

// New creates a SplashPotion with the position and velocity provided. It doesn't spawn the SplashPotion,
// only returns it.
func (s *SplashPotion) New(pos, vel mgl64.Vec3, t potion.Potion, owner world.Entity) world.Entity {
	splash := NewSplashPotion(pos, owner, t)
	splash.vel = vel
	splash.owner = owner
	return splash
}

// Explode ...
func (s *SplashPotion) Explode(explosionPos mgl64.Vec3, impact float64, _ block.ExplosionConfig) {
	s.mu.Lock()
	s.vel = s.vel.Add(s.pos.Sub(explosionPos).Normalize().Mul(impact))
	s.mu.Unlock()
}

// Owner ...
func (s *SplashPotion) Owner() world.Entity {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.owner
}

// DecodeNBT decodes the properties in a map to a SplashPotion and returns a new SplashPotion entity.
func (s *SplashPotion) DecodeNBT(data map[string]any) any {
	return s.New(
		nbtconv.MapVec3(data, "Pos"),
		nbtconv.MapVec3(data, "Motion"),
		potion.From(nbtconv.Map[int32](data, "PotionId")),
		nil,
	)
}

// EncodeNBT encodes the SplashPotion entity's properties as a map and returns it.
func (s *SplashPotion) EncodeNBT() map[string]any {
	return map[string]any{
		"Pos":      nbtconv.Vec3ToFloat32Slice(s.Position()),
		"Motion":   nbtconv.Vec3ToFloat32Slice(s.Velocity()),
		"PotionId": int32(s.t.Uint8()),
	}
}
