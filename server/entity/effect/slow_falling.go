package effect

import (
	"image/color"
)

// SlowFalling is a lasting effect that causes the affected entity to fall very
// slowly.
var SlowFalling slowFalling

type slowFalling struct {
	nopLasting
}

// RGBA ...
func (slowFalling) RGBA() color.RGBA {
	return color.RGBA{R: 0xf7, G: 0xf8, B: 0xe0, A: 0xff}
}
