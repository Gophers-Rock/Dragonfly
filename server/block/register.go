package block

import (
	"github.com/df-mc/dragonfly/server/block/grass"
	_ "github.com/df-mc/dragonfly/server/internal/block_internal"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	_ "unsafe" // Imported for compiler directives.
)

//go:generate go run ../../cmd/blockhash -o hash.go .

// init registers all blocks implemented by Dragonfly.
func init() {
	world.RegisterBlock(Air{})
	world.RegisterBlock(Stone{})
	world.RegisterBlock(Stone{Smooth: true})
	world.RegisterBlock(Granite{})
	world.RegisterBlock(Granite{Polished: true})
	world.RegisterBlock(Diorite{})
	world.RegisterBlock(Diorite{Polished: true})
	world.RegisterBlock(Andesite{})
	world.RegisterBlock(Andesite{Polished: true})
	world.RegisterBlock(Grass{})
	world.RegisterBlock(DirtPath{})
	world.RegisterBlock(Dirt{})
	world.RegisterBlock(Dirt{Coarse: true})
	world.RegisterBlock(Cobblestone{})
	world.RegisterBlock(Cobblestone{Mossy: true})
	world.RegisterBlock(Bedrock{})
	world.RegisterBlock(Bedrock{InfiniteBurning: true})
	world.RegisterBlock(Obsidian{})
	world.RegisterBlock(Obsidian{Crying: true})
	world.RegisterBlock(DiamondBlock{})
	world.RegisterBlock(Glass{})
	world.RegisterBlock(Glowstone{})
	world.RegisterBlock(EmeraldBlock{})
	world.RegisterBlock(EndBricks{})
	world.RegisterBlock(GoldBlock{})
	world.RegisterBlock(NetheriteBlock{})
	world.RegisterBlock(IronBlock{})
	world.RegisterBlock(CoalBlock{})
	world.RegisterBlock(Beacon{})
	world.RegisterBlock(Sponge{})
	world.RegisterBlock(Sponge{Wet: true})
	world.RegisterBlock(LapisBlock{})
	world.RegisterBlock(Terracotta{})
	world.RegisterBlock(GlassPane{})
	world.RegisterBlock(IronBars{})
	world.RegisterBlock(NetherBrickFence{})
	world.RegisterBlock(EndStone{})
	world.RegisterBlock(Netherrack{})
	world.RegisterBlock(QuartzBricks{})
	world.RegisterBlock(Clay{})
	world.RegisterBlock(AncientDebris{})
	world.RegisterBlock(EmeraldOre{})
	world.RegisterBlock(DiamondOre{})
	world.RegisterBlock(LapisOre{})
	world.RegisterBlock(NetherGoldOre{})
	world.RegisterBlock(GoldOre{})
	world.RegisterBlock(IronOre{})
	world.RegisterBlock(CoalOre{})
	world.RegisterBlock(NetherQuartzOre{})
	world.RegisterBlock(Melon{})
	world.RegisterBlock(Sand{})
	world.RegisterBlock(Sand{Red: true})
	world.RegisterBlock(Gravel{})
	world.RegisterBlock(Bricks{})
	world.RegisterBlock(SoulSand{})
	world.RegisterBlock(Barrier{})
	world.RegisterBlock(SeaLantern{})
	world.RegisterBlock(SoulSoil{})
	world.RegisterBlock(BlueIce{})
	world.RegisterBlock(GildedBlackstone{})
	world.RegisterBlock(Shroomlight{})
	world.RegisterBlock(InvisibleBedrock{})
	world.RegisterBlock(DragonEgg{})
	world.RegisterBlock(NoteBlock{})

	registerAll(allSigns())
	registerAll(allBasalt())
	registerAll(allBeetroot())
	registerAll(allBoneBlock())
	registerAll(allCake())
	registerAll(allCarpet())
	registerAll(allCarrots())
	registerAll(allChests())
	registerAll(allConcrete())
	registerAll(allConcretePowder())
	registerAll(allCocoaBeans())
	registerAll(allCoral())
	registerAll(allCoralBlocks())
	registerAll(allEndBrickStairs())
	registerAll(allWool())
	registerAll(allStainedTerracotta())
	registerAll(allGlazedTerracotta())
	registerAll(allStainedGlass())
	registerAll(allStainedGlassPane())
	registerAll(allLanterns())
	registerAll(allFire())
	registerAll(allPlanks())
	registerAll(allFence())
	registerAll(allFenceGates())
	registerAll(allWoodStairs())
	registerAll(allDoors())
	registerAll(allTrapdoors())
	registerAll(allWoodSlabs())
	registerAll(allLogs())
	registerAll(allLeaves())
	registerAll(allTorches())
	registerAll(allPumpkinStems())
	registerAll(allPumpkins())
	registerAll(allLitPumpkins())
	registerAll(allMelonStems())
	registerAll(allFarmland())
	registerAll(allLava())
	registerAll(allWater())
	registerAll(allKelp())
	registerAll(allPotato())
	registerAll(allWheat())
	registerAll(allQuartz())
	registerAll(allNetherWart())
	registerAll(allGrassPlants())
	registerAll(allSandstones())
}

func init() {
	world.RegisterItem(Air{})
	world.RegisterItem(Stone{})
	world.RegisterItem(Stone{Smooth: true})
	world.RegisterItem(Granite{})
	world.RegisterItem(Granite{Polished: true})
	world.RegisterItem(Diorite{})
	world.RegisterItem(Diorite{Polished: true})
	world.RegisterItem(Andesite{})
	world.RegisterItem(Andesite{Polished: true})
	world.RegisterItem(Grass{})
	world.RegisterItem(DirtPath{})
	world.RegisterItem(Dirt{})
	world.RegisterItem(Dirt{Coarse: true})
	world.RegisterItem(Cobblestone{})
	world.RegisterItem(Bedrock{})
	world.RegisterItem(Kelp{})
	world.RegisterItem(Chest{})
	world.RegisterItem(Cobblestone{Mossy: true})
	world.RegisterItem(Leaves{Wood: OakWood(), Persistent: true})
	world.RegisterItem(Leaves{Wood: SpruceWood(), Persistent: true})
	world.RegisterItem(Leaves{Wood: BirchWood(), Persistent: true})
	world.RegisterItem(Leaves{Wood: JungleWood(), Persistent: true})
	world.RegisterItem(Leaves{Wood: AcaciaWood(), Persistent: true})
	world.RegisterItem(Leaves{Wood: DarkOakWood(), Persistent: true})
	for _, c := range Colours() {
		world.RegisterItem(Concrete{Colour: c})
		world.RegisterItem(ConcretePowder{Colour: c})
		world.RegisterItem(StainedTerracotta{Colour: c})
		world.RegisterItem(Carpet{Colour: c})
		world.RegisterItem(Wool{Colour: c})
		world.RegisterItem(StainedGlass{Colour: c})
		world.RegisterItem(StainedGlassPane{Colour: c})
		world.RegisterItem(GlazedTerracotta{Colour: c})
	}
	for _, b := range allLight() {
		world.RegisterItem(b.(world.Item))
	}
	for _, w := range WoodTypes() {
		world.RegisterItem(Planks{Wood: w})
		world.RegisterItem(WoodSign{Wood: w})
		world.RegisterItem(WoodDoor{Wood: w})
		world.RegisterItem(WoodTrapdoor{Wood: w})
		world.RegisterItem(WoodFenceGate{Wood: w})
		world.RegisterItem(WoodFence{Wood: w})
		world.RegisterItem(WoodSlab{Wood: w})
		world.RegisterItem(WoodSlab{Wood: w, Double: true})
		world.RegisterItem(WoodStairs{Wood: w})
		world.RegisterItem(Log{Wood: w})
		world.RegisterItem(Log{Wood: w, Stripped: true})
	}
	world.RegisterItem(Obsidian{})
	world.RegisterItem(Obsidian{Crying: true})
	world.RegisterItem(DiamondBlock{})
	world.RegisterItem(Glass{})
	world.RegisterItem(Glowstone{})
	world.RegisterItem(EmeraldBlock{})
	world.RegisterItem(EndBricks{})
	world.RegisterItem(EndBrickStairs{})
	world.RegisterItem(NetheriteBlock{})
	world.RegisterItem(GoldBlock{})
	world.RegisterItem(IronBlock{})
	world.RegisterItem(CoalBlock{})
	world.RegisterItem(Beacon{})
	world.RegisterItem(Sponge{})
	world.RegisterItem(Sponge{Wet: true})
	world.RegisterItem(LapisBlock{})
	world.RegisterItem(Terracotta{})
	world.RegisterItem(Quartz{})
	world.RegisterItem(Quartz{Smooth: true})
	world.RegisterItem(ChiseledQuartz{})
	world.RegisterItem(QuartzPillar{})
	world.RegisterItem(QuartzBricks{})
	world.RegisterItem(GlassPane{})
	world.RegisterItem(IronBars{})
	world.RegisterItem(NetherBrickFence{})
	for _, c := range allCoral() {
		world.RegisterItem(c.(world.Item))
	}
	for _, c := range allCoralBlocks() {
		world.RegisterItem(c.(world.Item))
	}
	for _, s := range allSandstones() {
		world.RegisterItem(s.(world.Item))
	}
	world.RegisterItem(Pumpkin{})
	world.RegisterItem(LitPumpkin{})
	world.RegisterItem(Pumpkin{Carved: true})
	world.RegisterItem(EndStone{})
	world.RegisterItem(Netherrack{})
	world.RegisterItem(Clay{})
	world.RegisterItem(BoneBlock{})
	world.RegisterItem(Lantern{Type: NormalFire()})
	world.RegisterItem(Lantern{Type: SoulFire()})
	world.RegisterItem(AncientDebris{})
	world.RegisterItem(EmeraldOre{})
	world.RegisterItem(DiamondOre{})
	world.RegisterItem(LapisOre{})
	world.RegisterItem(NetherGoldOre{})
	world.RegisterItem(GoldOre{})
	world.RegisterItem(IronOre{})
	world.RegisterItem(CoalOre{})
	world.RegisterItem(NetherQuartzOre{})
	world.RegisterItem(CocoaBean{})
	world.RegisterItem(WheatSeeds{})
	world.RegisterItem(BeetrootSeeds{})
	world.RegisterItem(Potato{})
	world.RegisterItem(Carrot{})
	world.RegisterItem(PumpkinSeeds{})
	world.RegisterItem(MelonSeeds{})
	world.RegisterItem(Melon{})
	world.RegisterItem(Sand{})
	world.RegisterItem(Sand{Red: true})
	world.RegisterItem(Gravel{})
	world.RegisterItem(Bricks{})
	world.RegisterItem(SoulSand{})
	world.RegisterItem(Barrier{})
	world.RegisterItem(Basalt{})
	world.RegisterItem(Basalt{Polished: true})
	world.RegisterItem(SeaLantern{})
	world.RegisterItem(SoulSoil{})
	world.RegisterItem(BlueIce{})
	world.RegisterItem(GildedBlackstone{})
	world.RegisterItem(Shroomlight{})
	world.RegisterItem(Torch{Type: NormalFire()})
	world.RegisterItem(Torch{Type: SoulFire()})
	world.RegisterItem(Cake{})
	world.RegisterItem(NetherWart{})
	world.RegisterItem(InvisibleBedrock{})
	world.RegisterItem(NoteBlock{Pitch: 24})
	world.RegisterItem(DragonEgg{})
	world.RegisterItem(GrassPlant{})
	world.RegisterItem(GrassPlant{Type: grass.NetherSprouts()})
	world.RegisterItem(GrassPlant{Type: grass.Fern()})
	world.RegisterItem(GrassPlant{Type: grass.TallGrass()})
	world.RegisterItem(GrassPlant{Type: grass.LargeFern()})
	world.RegisterItem(Farmland{})

	world.RegisterItem(item.Bucket{Content: Water{}})
	world.RegisterItem(item.Bucket{Content: Lava{}})
}

// readSlice reads an interface slice from a map at the key passed.
//noinspection GoCommentLeadingSpace
func readSlice(m map[string]interface{}, key string) []interface{} {
	//lint:ignore S1005 Double assignment is done explicitly to prevent panics.
	v, _ := m[key]
	b, _ := v.([]interface{})
	return b
}

// readString reads a string from a map at the key passed.
//noinspection GoCommentLeadingSpace
func readString(m map[string]interface{}, key string) string {
	//lint:ignore S1005 Double assignment is done explicitly to prevent panics.
	v, _ := m[key]
	b, _ := v.(string)
	return b
}

// readInt32 reads an int32 from a map at the key passed.
//noinspection GoCommentLeadingSpace
func readInt32(m map[string]interface{}, key string) int32 {
	//lint:ignore S1005 Double assignment is done explicitly to prevent panics.
	v, _ := m[key]
	b, _ := v.(int32)
	return b
}

// readByte reads a byte from a map at the key passed.
//noinspection GoCommentLeadingSpace
func readByte(m map[string]interface{}, key string) byte {
	//lint:ignore S1005 Double assignment is done explicitly to prevent panics.
	v, _ := m[key]
	b, _ := v.(byte)
	return b
}

func registerAll(blocks []world.Block) {
	for _, b := range blocks {
		world.RegisterBlock(b)
	}
}
