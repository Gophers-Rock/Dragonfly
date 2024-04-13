package item

// NetheriteIngot is a rare mineral crafted with 4 pieces of netherite scrap and 4 gold ingots.
type NetheriteIngot struct{}

// EncodeItem ...
func (NetheriteIngot) EncodeItem() (name string, meta int16) {
	return "minecraft:netherite_ingot", 0
}

// TrimMaterial ...
func (NetheriteIngot) TrimMaterial() string {
	return "netherite"
}

// MaterialColor ...
func (NetheriteIngot) MaterialColor() string {
	return "j"
}

// PayableForBeacon ...
func (NetheriteIngot) PayableForBeacon() bool {
	return true
}
