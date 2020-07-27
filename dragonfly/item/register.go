package item

import (
	"github.com/df-mc/dragonfly/dragonfly/item/armour"
	"github.com/df-mc/dragonfly/dragonfly/item/bucket"
	"github.com/df-mc/dragonfly/dragonfly/item/tool"
	"github.com/df-mc/dragonfly/dragonfly/world"
)

//noinspection SpellCheckingInspection
func init() {
	world.RegisterItem("minecraft:wooden_pickaxe", Pickaxe{Tier: tool.TierWood})
	world.RegisterItem("minecraft:golden_pickaxe", Pickaxe{Tier: tool.TierGold})
	world.RegisterItem("minecraft:stone_pickaxe", Pickaxe{Tier: tool.TierStone})
	world.RegisterItem("minecraft:iron_pickaxe", Pickaxe{Tier: tool.TierIron})
	world.RegisterItem("minecraft:diamond_pickaxe", Pickaxe{Tier: tool.TierDiamond})
	world.RegisterItem("minecraft:netherite_pickaxe", Pickaxe{Tier: tool.TierNetherite})

	world.RegisterItem("minecraft:wooden_axe", Axe{Tier: tool.TierWood})
	world.RegisterItem("minecraft:golden_axe", Axe{Tier: tool.TierGold})
	world.RegisterItem("minecraft:stone_axe", Axe{Tier: tool.TierStone})
	world.RegisterItem("minecraft:iron_axe", Axe{Tier: tool.TierIron})
	world.RegisterItem("minecraft:diamond_axe", Axe{Tier: tool.TierDiamond})
	world.RegisterItem("minecraft:netherite_axe", Axe{Tier: tool.TierNetherite})

	world.RegisterItem("minecraft:wooden_shovel", Shovel{Tier: tool.TierWood})
	world.RegisterItem("minecraft:golden_shovel", Shovel{Tier: tool.TierGold})
	world.RegisterItem("minecraft:stone_shovel", Shovel{Tier: tool.TierStone})
	world.RegisterItem("minecraft:iron_shovel", Shovel{Tier: tool.TierIron})
	world.RegisterItem("minecraft:diamond_shovel", Shovel{Tier: tool.TierDiamond})
	world.RegisterItem("minecraft:netherite_shovel", Shovel{Tier: tool.TierNetherite})

	world.RegisterItem("minecraft:wooden_sword", Sword{Tier: tool.TierWood})
	world.RegisterItem("minecraft:golden_sword", Sword{Tier: tool.TierGold})
	world.RegisterItem("minecraft:stone_sword", Sword{Tier: tool.TierStone})
	world.RegisterItem("minecraft:iron_sword", Sword{Tier: tool.TierIron})
	world.RegisterItem("minecraft:diamond_sword", Sword{Tier: tool.TierDiamond})
	world.RegisterItem("minecraft:netherite_sword", Sword{Tier: tool.TierNetherite})

	world.RegisterItem("minecraft:leather_helmet", Helmet{Tier: armour.TierLeather})
	world.RegisterItem("minecraft:golden_helmet", Helmet{Tier: armour.TierGold})
	world.RegisterItem("minecraft:chainmail_helmet", Helmet{Tier: armour.TierChain})
	world.RegisterItem("minecraft:iron_helmet", Helmet{Tier: armour.TierIron})
	world.RegisterItem("minecraft:diamond_helmet", Helmet{Tier: armour.TierDiamond})
	world.RegisterItem("minecraft:netherite_helmet", Helmet{Tier: armour.TierNetherite})

	world.RegisterItem("minecraft:leather_chestplate", Chestplate{Tier: armour.TierLeather})
	world.RegisterItem("minecraft:golden_chestplate", Chestplate{Tier: armour.TierGold})
	world.RegisterItem("minecraft:chainmail_chestplate", Chestplate{Tier: armour.TierChain})
	world.RegisterItem("minecraft:iron_chestplate", Chestplate{Tier: armour.TierIron})
	world.RegisterItem("minecraft:diamond_chestplate", Chestplate{Tier: armour.TierDiamond})
	world.RegisterItem("minecraft:netherite_chestplate", Chestplate{Tier: armour.TierNetherite})

	world.RegisterItem("minecraft:leather_leggings", Leggings{Tier: armour.TierLeather})
	world.RegisterItem("minecraft:golden_leggings", Leggings{Tier: armour.TierGold})
	world.RegisterItem("minecraft:chainmail_leggings", Leggings{Tier: armour.TierChain})
	world.RegisterItem("minecraft:iron_leggings", Leggings{Tier: armour.TierIron})
	world.RegisterItem("minecraft:diamond_leggings", Leggings{Tier: armour.TierDiamond})
	world.RegisterItem("minecraft:netherite_leggings", Leggings{Tier: armour.TierNetherite})

	world.RegisterItem("minecraft:leather_boots", Boots{Tier: armour.TierLeather})
	world.RegisterItem("minecraft:golden_boots", Boots{Tier: armour.TierGold})
	world.RegisterItem("minecraft:chainmail_boots", Boots{Tier: armour.TierChain})
	world.RegisterItem("minecraft:iron_boots", Boots{Tier: armour.TierIron})
	world.RegisterItem("minecraft:diamond_boots", Boots{Tier: armour.TierDiamond})
	world.RegisterItem("minecraft:netherite_boots", Boots{Tier: armour.TierNetherite})

	world.RegisterItem("minecraft:bucket", Bucket{})
	world.RegisterItem("minecraft:bucket", Bucket{Content: bucket.Water()})
	world.RegisterItem("minecraft:bucket", Bucket{Content: bucket.Lava()})

	world.RegisterItem("minecraft:shears", Shears{})

	world.RegisterItem("minecraft:diamond", Diamond{})
	world.RegisterItem("minecraft:glowstone_dust", GlowstoneDust{})
	world.RegisterItem("minecraft:dye", LapisLazuli{})
	world.RegisterItem("minecraft:emerald", Emerald{})
	world.RegisterItem("minecraft:gold_ingot", GoldIngot{})
	world.RegisterItem("minecraft:gold_nugget", GoldNugget{})
	world.RegisterItem("minecraft:iron_ingot", IronIngot{})
	world.RegisterItem("minecraft:coal", Coal{})
	world.RegisterItem("minecraft:netherite_ingot", NetheriteIngot{})
	world.RegisterItem("minecraft:clay_ball", ClayBall{})

	world.RegisterItem("minecraft:stick", Stick{})
}
