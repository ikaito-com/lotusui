// Package looks defines the docs site's look-and-feel presets — the
// second, orthogonal axis of theming beside color palettes. A look is
// everything about a theme that ISN'T color: typeface, spacing
// rhythm, corner radii, base text size. Look × palette compose into
// one NewTheme call — no combinatorial builds, no core changes:
// lotusui's options were designed to be composed, and composition
// resolves once at construction.
package looks

import (
	"embed"
	"sync"

	"gioui.org/font"
	"gioui.org/font/opentype"
	"gioui.org/unit"

	lotusui "github.com/ikaito-com/lotusui"
)

// Fonts are fetched at development time (like icons) and committed —
// builds never need the network. Each is a variable-weight family.
//
//go:embed fonts/*.ttf
var fontFS embed.FS

// Preset is one look: the non-color half of a theme, plus the CSS
// metadata the static pages need to wear the same look.
type Preset struct {
	Slug, Name string
	// Famous-app shorthand shown in the picker.
	Hint string

	Radius   lotusui.RadiusScale
	Space    lotusui.SpaceScale
	TextSize unit.Sp

	// FontFile names an embedded TTF (empty = lotusui's DM Sans);
	// CSSFamily is the family name pages use via @font-face.
	FontFile  string
	CSSFamily string
}

// Presets, default first.
var Presets = []Preset{
	{
		Slug: "lotus", Name: "Lotus", Hint: "the lotusui default",
		Radius:    lotusui.RadiusScale{XS: 4, SM: 6, MD: 10, LG: 16},
		Space:     lotusui.SpaceScale{XS: 4, SM: 8, MD: 16, LG: 24, XL: 32},
		TextSize:  16,
		CSSFamily: "DM Sans",
	},
	{
		Slug: "editorial", Name: "Editorial", Hint: "serif, sharp, print-like",
		Radius:   lotusui.RadiusScale{XS: 1, SM: 2, MD: 4, LG: 8},
		Space:    lotusui.SpaceScale{XS: 4, SM: 10, MD: 18, LG: 28, XL: 40},
		TextSize: 17,
		FontFile: "lora.ttf", CSSFamily: "Lora",
	},
	{
		Slug: "coast", Name: "Coast", Hint: "friendly, rounded, roomy",
		Radius:   lotusui.RadiusScale{XS: 6, SM: 10, MD: 16, LG: 24},
		Space:    lotusui.SpaceScale{XS: 6, SM: 10, MD: 18, LG: 28, XL: 40},
		TextSize: 16,
		FontFile: "nunito.ttf", CSSFamily: "Nunito",
	},
	{
		Slug: "console", Name: "Console", Hint: "dense, crisp, tool-like",
		Radius:   lotusui.RadiusScale{XS: 2, SM: 4, MD: 6, LG: 10},
		Space:    lotusui.SpaceScale{XS: 2, SM: 6, MD: 12, LG: 18, XL: 24},
		TextSize: 14,
		FontFile: "inter.ttf", CSSFamily: "Inter",
	},
	{
		Slug: "playful", Name: "Playful", Hint: "chunky, extra-round, airy",
		Radius:   lotusui.RadiusScale{XS: 8, SM: 12, MD: 18, LG: 28},
		Space:    lotusui.SpaceScale{XS: 6, SM: 12, MD: 20, LG: 30, XL: 44},
		TextSize: 16,
		FontFile: "baloo2.ttf", CSSFamily: "Baloo 2",
	},
}

// ByName resolves a look slug; the default when unknown.
func ByName(slug string) Preset {
	for _, p := range Presets {
		if p.Slug == slug {
			return p
		}
	}
	return Presets[0]
}

// FontBytes returns the embedded TTF for a preset (nil for the
// default look — lotusui embeds DM Sans itself).
func (p Preset) FontBytes() []byte {
	if p.FontFile == "" {
		return nil
	}
	b, _ := fontFS.ReadFile("fonts/" + p.FontFile)
	return b
}

var (
	facesMu    sync.Mutex
	facesCache = map[string][]font.FontFace{}
)

// Faces parses a preset's font collection, cached — parsing runs once
// per look ever, not per theme rebuild.
func (p Preset) Faces() []font.FontFace {
	if p.FontFile == "" {
		return nil
	}
	facesMu.Lock()
	defer facesMu.Unlock()
	if f, ok := facesCache[p.Slug]; ok {
		return f
	}
	faces, err := opentype.ParseCollection(p.FontBytes())
	if err != nil {
		return nil
	}
	facesCache[p.Slug] = faces
	return faces
}

// Theme composes a look with a palette — the whole point: two
// orthogonal preset axes, one ordinary NewTheme call.
func (p Preset) Theme(pal lotusui.Palette) *lotusui.Theme {
	opts := []lotusui.ThemeOption{
		lotusui.WithPalette(pal),
		lotusui.WithRadius(p.Radius),
		lotusui.WithSpace(p.Space),
		lotusui.WithTextSize(p.TextSize),
	}
	if f := p.Faces(); f != nil {
		opts = append(opts, lotusui.WithFaces(f))
	}
	return lotusui.NewTheme(opts...)
}
