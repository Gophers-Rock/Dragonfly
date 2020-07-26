package block

import "github.com/df-mc/dragonfly/dragonfly/item"

// Netherrack is a block found in The Nether.
type Netherrack struct {
	noNBT
	solid
}

// BreakInfo ...
func (e Netherrack) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    0.4,
		Harvestable: pickaxeHarvestable,
		Effective:   pickaxeEffective,
		Drops:       simpleDrops(item.NewStack(e, 1)),
	}
}

// EncodeItem ...
func (e Netherrack) EncodeItem() (id int32, meta int16) {
	return 87, 0
}

// EncodeBlock ...
func (e Netherrack) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:netherrack", nil
}

// Hash ...
func (e Netherrack) Hash() uint64 {
	return hashNetherrack
}
