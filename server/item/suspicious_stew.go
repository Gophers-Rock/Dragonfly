package item

import (
	"github.com/df-mc/dragonfly/server/world"
)

// SuspiciousStew is a food item that can give the player a status effect that depends on the flower used to craft it.
type SuspiciousStew struct {
	defaultFood
	Effect StewEffect
}

// MaxCount ...
func (SuspiciousStew) MaxCount() int {
	return 1
}

// AlwaysConsumable ...
func (SuspiciousStew) AlwaysConsumable() bool {
	return true
}

// EncodeItem ...
func (s SuspiciousStew) EncodeItem() (name string, meta int16) {
	return "suspicious_stew", int16(s.Effect.Uint8())
}

// Consume ...
func (s SuspiciousStew) Consume(_ *world.World, c Consumer) Stack {
	c.AddEffect(s.Effect.Type())
	c.Saturate(6, 7.2)

	return NewStack(Bowl{}, 1)
}
