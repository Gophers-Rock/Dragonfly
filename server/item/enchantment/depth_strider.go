package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// DepthStrider is a boot enchantment that increases underwater movement speed.
type DepthStrider struct{}

// Name ...
func (DepthStrider) Name() string {
	return "Depth Strider"
}

// MaxLevel ...
func (DepthStrider) MaxLevel() int {
	return 3
}

// MinCost ...
func (DepthStrider) MinCost(level int) int {
	return 10 * level
}

// MaxCost ...
func (d DepthStrider) MaxCost(level int) int {
	return d.MinCost(level) + 15
}

// Rarity ...
func (DepthStrider) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityRare
}

// CompatibleWithEnchantment ...
func (DepthStrider) CompatibleWithEnchantment(item.EnchantmentType) bool {
	// TODO: Frost Walker
	return true
}

// CompatibleWithItem ...
func (DepthStrider) CompatibleWithItem(i world.Item) bool {
	b, ok := i.(item.BootsType)
	return ok && b.Boots()
}
