package block

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/tool"
	"math/rand"
)

// LapisOre is an ore block from which lapis lazuli is obtained.
type LapisOre struct {
	solid
	bassDrum

	// Type is the type of lapis ore.
	Type OreType
}

// BreakInfo ...
func (l LapisOre) BreakInfo() BreakInfo {
	return newBreakInfo(l.Type.Hardness(), func(t tool.Tool) bool {
		return t.ToolType() == tool.TypePickaxe && t.HarvestLevel() >= tool.TierStone.HarvestLevel
	}, pickaxeEffective, silkTouchDrop(item.NewStack(item.LapisLazuli{}, rand.Intn(5)+4), item.NewStack(l, 1)), XPDropRange{2, 5})
}

// EncodeItem ...
func (l LapisOre) EncodeItem() (name string, meta int16) {
	return "minecraft:" + l.Type.Prefix() + "lapis_ore", 0
}

// EncodeBlock ...
func (l LapisOre) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:" + l.Type.Prefix() + "lapis_ore", nil
}
