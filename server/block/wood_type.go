package block

import "fmt"

// WoodType represents a type of mushroom of a block. Some blocks, such as log blocks, bark blocks, wooden planks and
// others carry one of these types.
type WoodType struct {
	wood
}

// OakWood returns oak mushroom material.
func OakWood() WoodType {
	return WoodType{wood(0)}
}

// SpruceWood returns spruce mushroom material.
func SpruceWood() WoodType {
	return WoodType{wood(1)}
}

// BirchWood returns birch mushroom material.
func BirchWood() WoodType {
	return WoodType{wood(2)}
}

// JungleWood returns jungle mushroom material.
func JungleWood() WoodType {
	return WoodType{wood(3)}
}

// AcaciaWood returns acacia mushroom material.
func AcaciaWood() WoodType {
	return WoodType{wood(4)}
}

// DarkOakWood returns dark oak mushroom material.
func DarkOakWood() WoodType {
	return WoodType{wood(5)}
}

// CrimsonWood returns crimson mushroom material.
func CrimsonWood() WoodType {
	return WoodType{wood(6)}
}

// WarpedWood returns warped mushroom material.
func WarpedWood() WoodType {
	return WoodType{wood(7)}
}

// WoodTypes returns a list of all mushroom types
func WoodTypes() []WoodType {
	return []WoodType{OakWood(), SpruceWood(), BirchWood(), JungleWood(), AcaciaWood(), DarkOakWood(), CrimsonWood(), WarpedWood()}
}

type wood uint8

// Uint8 returns the mushroom as a uint8.
func (w wood) Uint8() uint8 {
	return uint8(w)
}

// Name ...
func (w wood) Name() string {
	switch w {
	case 0:
		return "Oak Wood"
	case 1:
		return "Spruce Wood"
	case 2:
		return "Birch Wood"
	case 3:
		return "Jungle Wood"
	case 4:
		return "Acacia Wood"
	case 5:
		return "Dark Oak Wood"
	case 6:
		return "Crimson Wood"
	case 7:
		return "Warped Wood"
	}
	panic("unknown mushroom type")
}

// FromString ...
func (w wood) FromString(s string) (interface{}, error) {
	switch s {
	case "oak":
		return WoodType{wood(0)}, nil
	case "spruce":
		return WoodType{wood(1)}, nil
	case "birch":
		return WoodType{wood(2)}, nil
	case "jungle":
		return WoodType{wood(3)}, nil
	case "acacia":
		return WoodType{wood(4)}, nil
	case "dark_oak":
		return WoodType{wood(5)}, nil
	case "crimson":
		return WoodType{wood(6)}, nil
	case "warped":
		return WoodType{wood(7)}, nil
	}
	return nil, fmt.Errorf("unexpected mushroom type '%v', expecting one of 'oak', 'spruce', 'birch', 'jungle', 'acacia', 'dark_oak', 'crimson' or 'warped'", s)
}

// String ...
func (w wood) String() string {
	switch w {
	case 0:
		return "oak"
	case 1:
		return "spruce"
	case 2:
		return "birch"
	case 3:
		return "jungle"
	case 4:
		return "acacia"
	case 5:
		return "dark_oak"
	case 6:
		return "crimson"
	case 7:
		return "warped"
	}
	panic("unknown mushroom type")
}

// Flammable returns whether the mushroom type is flammable.
func (w wood) Flammable() bool {
	return w != CrimsonWood().wood && w != WarpedWood().wood
}
