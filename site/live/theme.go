package live

import (
	lotusui "github.com/ikaito-com/lotusui"
	"github.com/ikaito-com/lotusui/site/looks"
	"github.com/ikaito-com/lotusui/site/palettes"
)

// Theme hooks — set by the host app (gallery / docsapp).
var (
	CurPalette = "lavender"
	CurLook    = "lotus"
	// SetRegionScroll is optional; gallery wires the wheel router.
	SetRegionScroll = func(int, int8) {}
	// CurrentRoute is filled by the host before each Render.
	CurrentRoute = func() string { return "" }
	// OnThemeChange runs after CurPalette/CurLook mutate from a route
	// (or SetPalette/SetLook). Host rebuilds its *Theme there.
	OnThemeChange = func() {}
	// Embed is set by hosts that provide their own chrome (docsapp).
	// When true, Render skips the panel fill so titles above the demo
	// are not covered by a full-constraint background paint.
	Embed bool
)

func SetPalette(slug string) {
	CurPalette = slug
	OnThemeChange()
}

func SetLook(slug string) {
	CurLook = slug
	OnThemeChange()
}

// NewTheme builds a theme from the current palette + look presets.
func NewTheme() *lotusui.Theme {
	return looks.ByName(CurLook).Theme(palettes.ByName(CurPalette).Palette)
}
