package item

import "fmt"

// Colour represents the colour of a block. Typically, Minecraft blocks have a total of 16 different colours.
type Colour struct {
	colour
}

// ColourWhite returns the white colour.
func ColourWhite() Colour {
	return Colour{colour(0)}
}

// ColourOrange returns the orange colour.
func ColourOrange() Colour {
	return Colour{colour(1)}
}

// ColourMagenta returns the magenta colour.
func ColourMagenta() Colour {
	return Colour{colour(2)}
}

// ColourLightBlue returns the light blue colour.
func ColourLightBlue() Colour {
	return Colour{colour(3)}
}

// ColourYellow returns the yellow colour.
func ColourYellow() Colour {
	return Colour{colour(4)}
}

// ColourLime returns the lime colour.
func ColourLime() Colour {
	return Colour{colour(5)}
}

// ColourPink returns the pink colour.
func ColourPink() Colour {
	return Colour{colour(6)}
}

// ColourGrey returns the grey colour.
func ColourGrey() Colour {
	return Colour{colour(7)}
}

// ColourLightGrey returns the light grey colour.
func ColourLightGrey() Colour {
	return Colour{colour(8)}
}

// ColourCyan returns the cyan colour.
func ColourCyan() Colour {
	return Colour{colour(9)}
}

// ColourPurple returns the purple colour.
func ColourPurple() Colour {
	return Colour{colour(10)}
}

// ColourBlue returns the blue colour.
func ColourBlue() Colour {
	return Colour{colour(11)}
}

// ColourBrown returns the brown colour.
func ColourBrown() Colour {
	return Colour{colour(12)}
}

// ColourGreen returns the green colour.
func ColourGreen() Colour {
	return Colour{colour(13)}
}

// ColourRed returns the red colour.
func ColourRed() Colour {
	return Colour{colour(14)}
}

// ColourBlack returns the black colour.
func ColourBlack() Colour {
	return Colour{colour(15)}
}

// Colours returns a list of all existing colours.
func Colours() []Colour {
	return []Colour{
		ColourWhite(), ColourOrange(), ColourMagenta(), ColourLightBlue(), ColourYellow(), ColourLime(), ColourPink(), ColourGrey(),
		ColourLightGrey(), ColourCyan(), ColourPurple(), ColourBlue(), ColourBrown(), ColourGreen(), ColourRed(), ColourBlack(),
	}
}

type colour uint8

// String ...
func (c colour) String() string {
	switch c {
	default:
		return "white"
	case 1:
		return "orange"
	case 2:
		return "magenta"
	case 3:
		return "light_blue"
	case 4:
		return "yellow"
	case 5:
		return "lime"
	case 6:
		return "pink"
	case 7:
		return "gray"
	case 8:
		return "silver"
	case 9:
		return "cyan"
	case 10:
		return "purple"
	case 11:
		return "blue"
	case 12:
		return "brown"
	case 13:
		return "green"
	case 14:
		return "red"
	case 15:
		return "black"
	}
}

// FromString ...
func (c colour) FromString(s string) (interface{}, error) {
	switch s {
	case "white":
		return Colour{colour(0)}, nil
	case "orange":
		return Colour{colour(1)}, nil
	case "magenta":
		return Colour{colour(2)}, nil
	case "light_blue":
		return Colour{colour(3)}, nil
	case "yellow":
		return Colour{colour(4)}, nil
	case "lime", "light_green":
		return Colour{colour(5)}, nil
	case "pink":
		return Colour{colour(6)}, nil
	case "grey", "gray":
		return Colour{colour(7)}, nil
	case "light_grey", "light_gray", "silver":
		return Colour{colour(8)}, nil
	case "cyan":
		return Colour{colour(9)}, nil
	case "purple":
		return Colour{colour(10)}, nil
	case "blue":
		return Colour{colour(11)}, nil
	case "brown":
		return Colour{colour(12)}, nil
	case "green":
		return Colour{colour(13)}, nil
	case "red":
		return Colour{colour(14)}, nil
	case "black":
		return Colour{colour(15)}, nil
	}
	return nil, fmt.Errorf("unexpected colour '%v'", s)
}

// Uint8 ...
func (c colour) Uint8() uint8 {
	return uint8(c)
}
