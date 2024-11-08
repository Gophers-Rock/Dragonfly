package entity

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/world"
	"golang.org/x/exp/maps"
	"reflect"
)

// EffectManager manages the effects of an entity. The effect manager will only store effects that last for
// a specific duration. Instant effects are applied instantly and not stored.
type EffectManager struct {
	effects map[reflect.Type]effect.Effect
}

// NewEffectManager creates and returns a new initialised EffectManager.
func NewEffectManager() *EffectManager {
	return &EffectManager{effects: map[reflect.Type]effect.Effect{}}
}

// Add adds an effect to the manager. If the effect is instant, it is applied to the Living entity passed
// immediately. If not, the effect is added to the EffectManager and is applied to the entity every time the
// Tick method is called.
// Effect levels of 0 or below will not do anything.
// Effect returns the final effect it added to the entity. That might be the effect passed or an effect with
// a higher level/duration than the one passed. Add panics if the effect has a negative duration or level.
func (m *EffectManager) Add(e effect.Effect, entity Living) effect.Effect {
	lvl, dur := e.Level(), e.Duration()
	if lvl <= 0 {
		panic(fmt.Sprintf("(*EffectManager).Add: effect cannot have level of 0 or below: %v", lvl))
	}
	if dur < 0 {
		panic(fmt.Sprintf("(*EffectManager).Add: effect cannot have negative duration: %v", dur))
	}
	t, ok := e.Type().(effect.LastingType)
	if !ok {
		e.Type().Apply(entity, lvl, 0)
		return e
	}
	typ := reflect.TypeOf(e.Type())

	existing, ok := m.effects[typ]
	if !ok {
		m.effects[typ] = e

		t.Start(entity, lvl)
		return e
	}
	if existing.Level() > lvl || (existing.Level() == lvl && existing.Duration() > dur) {
		return existing
	}
	m.effects[typ] = e

	existing.Type().(effect.LastingType).End(entity, existing.Level())
	t.Start(entity, lvl)
	return e
}

// Remove removes any Effect present in the EffectManager with the type of the effect passed.
func (m *EffectManager) Remove(e effect.Type, entity Living) {
	t := reflect.TypeOf(e)
	if existing, ok := m.effects[t]; ok {
		existing.Type().(effect.LastingType).End(entity, existing.Level())
	}
	delete(m.effects, t)
}

// Effect returns the effect instance and true if the entity has the effect. If not found, it will return an empty
// effect instance and false.
func (m *EffectManager) Effect(e effect.Type) (effect.Effect, bool) {
	existing, ok := m.effects[reflect.TypeOf(e)]
	return existing, ok
}

// Effects returns a list of all effects currently present in the effect manager. This will never include
// effects that have expired.
func (m *EffectManager) Effects() []effect.Effect {
	return maps.Values(m.effects)
}

// Tick ticks the EffectManager, applying all of its effects to the Living entity passed when applicable and
// removing expired effects.
func (m *EffectManager) Tick(entity Living, tx *world.Tx) {
	update := false

	for i, eff := range m.effects {
		if m.expired(eff) {
			delete(m.effects, i)
			eff.Type().(effect.LastingType).End(entity, eff.Level())
			update = true
			continue
		}
		eff = eff.TickDuration()
		eff.Type().Apply(entity, eff.Level(), eff.Duration())
		m.effects[i] = eff
	}

	if update {
		for _, v := range tx.Viewers(entity.Position()) {
			v.ViewEntityState(entity)
		}
	}
}

// expired checks if an Effect has expired.
func (m *EffectManager) expired(e effect.Effect) bool {
	return e.Duration() <= 0
}
