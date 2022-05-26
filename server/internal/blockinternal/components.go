package blockinternal

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/world"
)

// Components returns all the components of the given custom block group. If the group; has no components, a nil map
// and false are returned.
func Components(identifier string, group []world.CustomBlock) (map[string]any, error) {
	if len(group) == 0 {
		// We don't have any blocks in the group so return false.
		return nil, nil
	}

	base := group[0]
	builder := NewComponentBuilder(identifier, group)
	if r, ok := base.(block.Rotatable); ok {
		for condition, value := range r.Rotation() {
			encoded := map[string]any{
				"x": float32(value.X()),
				"y": float32(value.Y()),
				"z": float32(value.Z()),
			}
			if !condition.Exists() {
				builder.AddComponent("minecraft:rotation", encoded)
				continue
			}
			builder.AddPermutation(condition.String(), map[string]any{"minecraft:rotation": encoded})
		}
	}
	if l, ok := base.(block.LightEmitter); ok {
		builder.AddComponent("minecraft:block_light_emission", map[string]any{
			"emission": float32(l.LightEmissionLevel() / 15),
		})
	}
	if d, ok := base.(block.LightDiffuser); ok {
		builder.AddComponent("minecraft:block_light_filter", map[string]any{
			"lightLevel": int32(d.LightDiffusionLevel()),
		})
	}
	if i, ok := base.(block.Breakable); ok {
		info := i.BreakInfo()
		builder.AddComponent("minecraft:destroy_time", map[string]any{
			"value": float32(info.Hardness),
		})
		// TODO: Explosion resistance.
	}
	if f, ok := base.(block.Frictional); ok {
		builder.AddComponent("minecraft:friction", map[string]any{
			"value": float32(f.Friction()),
		})
	}
	if f, ok := base.(block.Flammable); ok {
		info := f.FlammabilityInfo()
		builder.AddComponent("minecraft:flammable", map[string]any{
			"flame_odds": int32(info.Encouragement),
			"burn_odds":  int32(info.Flammability),
		})
	}
	if c, ok := base.(world.CustomItem); ok {
		category := c.Category()
		builder.AddComponent("minecraft:creative_category", map[string]any{
			"category": category.Name(),
			"group":    category.String(),
		})
	}
	return builder.Construct(), nil
}