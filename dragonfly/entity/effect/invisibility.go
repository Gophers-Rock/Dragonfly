package effect

import (
	"github.com/df-mc/dragonfly/dragonfly/entity"
	"image/color"
	"time"
)

// Invisibility is a lasting effect that causes the affected entity to turn invisible. While invisible, the
// entity's armour is still visible and effect particles will still be displayed.
type Invisibility struct {
	lastingEffect
}

// Start ...
func (Invisibility) Start(e entity.Living) {
	if i, ok := e.(interface {
		SetInvisible()
		SetVisible()
	}); ok {
		i.SetInvisible()
	}
}

// End ...
func (Invisibility) End(e entity.Living) {
	if i, ok := e.(interface {
		SetInvisible()
		SetVisible()
	}); ok {
		i.SetVisible()
	}
}

// WithDuration ...
func (i Invisibility) WithDuration(d time.Duration) entity.Effect {
	return Invisibility{i.withDuration(d)}
}

// RGBA ...
func (Invisibility) RGBA() color.RGBA {
	return color.RGBA{R: 0x7f, G: 0x83, B: 0x92, A: 0xff}
}
