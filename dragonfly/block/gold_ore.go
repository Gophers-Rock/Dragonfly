package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/item/tool"
)

// GoldOre is a rare mineral block found underground.
type GoldOre struct {
	noNBT
	solid
}

// BreakInfo ...
func (g GoldOre) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness: 3,
		Harvestable: func(t tool.Tool) bool {
			return t.ToolType() == tool.TypePickaxe && t.HarvestLevel() >= tool.TierIron.HarvestLevel
		},
		Effective: pickaxeEffective,
		Drops:     simpleDrops(item.NewStack(g, 1)),
	}
}

// EncodeItem ...
func (g GoldOre) EncodeItem() (id int32, meta int16) {
	return 14, 0
}

// EncodeBlock ...
func (g GoldOre) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:gold_ore", nil
}

// Hash ...
func (g GoldOre) Hash() uint64 {
	return hashGoldOre
}
