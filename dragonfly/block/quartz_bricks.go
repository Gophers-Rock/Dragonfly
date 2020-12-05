package block

import "github.com/df-mc/dragonfly/dragonfly/item"

// QuartzBricks is a mineral block used only for decoration.
type QuartzBricks struct {
	noNBT
	solid
	bassDrum
}

// BreakInfo ...
func (q QuartzBricks) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    0.8,
		Harvestable: pickaxeHarvestable,
		Effective:   pickaxeEffective,
		Drops:       simpleDrops(item.NewStack(q, 1)),
	}
}

// EncodeItem ...
func (QuartzBricks) EncodeItem() (id int32, meta int16) {
	return -304, 0
}
