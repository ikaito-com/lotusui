// Package palettes defines the docs site's selectable themes — eleven
// palettes PRE-BUILT at compile time, shared by the static-site
// generator (which emits them as CSS variables) and the wasm gallery
// (which swaps them at runtime). This is exactly the multi-theme
// pattern an app built on lotusui uses: a theme is ~200 bytes of
// data, so shipping eleven costs nothing, and switching is a pointer
// swap.
package palettes

import (
	"fmt"
	"image/color"
	"math"

	lotusui "github.com/ikaito-com/lotusui"
)

// Preset is one selectable look, adapted from a four-color source
// palette by ROLE, not by wash: the canvas and borders stay NEUTRAL
// (the default cool gray and white panels — never hue-tinted), the
// hero accent appears AS ITSELF on solid fills the way the source
// palette intends, and the palette's dark color — a different hue,
// deliberately — carries the ink. Color lives where color belongs;
// everything else stays quiet.
type Preset struct {
	Slug    string
	Name    string
	Palette lotusui.Palette
}

// vivid adapts one source palette: anchor is the hero accent used
// VERBATIM as BrandSolid; ink (optional, zero = keep neutral) is the
// palette's dark, used for the foreground ladder. Backgrounds and
// borders are left at the neutral defaults on purpose.
func vivid(slug, name string, anchor, ink color.NRGBA) Preset {
	s := lotusui.ScaleFrom(anchor)
	p := lotusui.DefaultPalette
	p.BrandSolid = anchor      // the palette's actual accent, not a pastel of it
	p.BrandEmphasized = s.C600 // pressed: same hue, darker
	p.BrandSubtle = s.C50      // faint fills stay faint
	p.BrandFg = s.C700         // readable same-hue ink on white
	// Text on the solid fill: white on dark accents, deep ink on light
	// ones — chosen by relative luminance so it always reads.
	if relLuminance(anchor) < 0.3 {
		p.BrandContrast = color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}
	} else if ink != (color.NRGBA{}) {
		p.BrandContrast = ink
	} else {
		p.BrandContrast = s.C800
	}
	if ink != (color.NRGBA{}) {
		p.Fg = ink
		p.FgMuted = mix(ink, 0.22)
		p.FgSubtle = mix(ink, 0.45)
		p.BgInverted = ink
	}
	return Preset{Slug: slug, Name: name, Palette: p}
}

// status harmonizes the semantic colors with the palette — each
// non-zero anchor re-grades that status from one of the palette's own
// colors (ink = C700 for AA contrast, background = C50), so every
// preset is a COMPLETE theme, not an accent swap. Zero anchors keep
// the neutral defaults.
func (pr Preset) status(success, warning, info, danger color.NRGBA) Preset {
	grade := func(anchor color.NRGBA, ink, bg *color.NRGBA) {
		if anchor == (color.NRGBA{}) {
			return
		}
		s := lotusui.ScaleFrom(anchor)
		*ink, *bg = s.C700, s.C50
	}
	grade(success, &pr.Palette.Success, &pr.Palette.SuccessBg)
	grade(warning, &pr.Palette.Warning, &pr.Palette.WarningBg)
	grade(info, &pr.Palette.Info, &pr.Palette.InfoBg)
	if danger != (color.NRGBA{}) {
		s := lotusui.ScaleFrom(danger)
		pr.Palette.Danger, pr.Palette.DangerBg = s.C600, s.C50
		pr.Palette.DangerContrast = color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}
	}
	return pr
}

// mix blends c toward white by t.
func mix(c color.NRGBA, t float64) color.NRGBA {
	m := func(v uint8) uint8 { return uint8(float64(v)*(1-t) + 255*t + 0.5) }
	return color.NRGBA{R: m(c.R), G: m(c.G), B: m(c.B), A: 0xFF}
}

// relLuminance is WCAG relative luminance — the same measure the CLI's
// contrast checker uses.
func relLuminance(c color.NRGBA) float64 {
	lin := func(v uint8) float64 {
		f := float64(v) / 255
		if f <= 0.04045 {
			return f / 12.92
		}
		return math.Pow((f+0.055)/1.055, 2.4)
	}
	return 0.2126*lin(c.R) + 0.7152*lin(c.G) + 0.0722*lin(c.B)
}

func rgb(r, g, b uint8) color.NRGBA { return color.NRGBA{R: r, G: g, B: b, A: 0xFF} }

// Presets, default first: lotusui's hand-tuned Lavender, then ten
// looks adapted from a much-loved community palette collection.
var Presets = []Preset{
	{Slug: "lavender", Name: "Lavender", Palette: lotusui.DefaultPalette},

	// The built-in dark palette — proof that dark mode is just a
	// palette swap, and the docs site's own dark theme.
	{Slug: "midnight", Name: "Midnight", Palette: lotusui.DefaultDarkPalette},

	// #222831 #393E46 #00ADB5 #EEEEEE — slate depths, one cyan beacon.
	vivid("deep-harbor", "Deep Harbor", rgb(0x00, 0xAD, 0xB5), rgb(0x22, 0x28, 0x31)).
		status(color.NRGBA{}, color.NRGBA{}, rgb(0x00, 0xAD, 0xB5), color.NRGBA{}),

	// #F9ED69 #F08A5D #B83B5E #6A2C70 — dusk ramp from gold to plum.
	vivid("summer-sunset", "Summer Sunset", rgb(0xB8, 0x3B, 0x5E), rgb(0x6A, 0x2C, 0x70)).
		status(color.NRGBA{}, rgb(0xF9, 0xED, 0x69), color.NRGBA{}, rgb(0xF0, 0x8A, 0x5D)),

	// #08D9D6 #252A34 #FF2E63 #EAEAEA — hot pink and cyan after dark.
	vivid("neon-arcade", "Neon Arcade", rgb(0xFF, 0x2E, 0x63), rgb(0x25, 0x2A, 0x34)).
		status(color.NRGBA{}, color.NRGBA{}, rgb(0x08, 0xD9, 0xD6), rgb(0xFF, 0x2E, 0x63)),

	// #A8D8EA #AA96DA #FCBAD3 #FFFFD2 — lilac, pink and cream pastels.
	vivid("cotton-candy", "Cotton Candy", rgb(0xAA, 0x96, 0xDA), color.NRGBA{}).
		status(color.NRGBA{}, color.NRGBA{}, rgb(0xA8, 0xD8, 0xEA), rgb(0xFC, 0xBA, 0xD3)),

	// #E4F9F5 #30E3CA #11999E #40514E — mint water over green slate.
	vivid("mint-lagoon", "Mint Lagoon", rgb(0x11, 0x99, 0x9E), rgb(0x40, 0x51, 0x4E)).
		status(rgb(0x30, 0xE3, 0xCA), color.NRGBA{}, color.NRGBA{}, color.NRGBA{}),

	// #E3FDFD #CBF1F5 #A6E3E9 #71C9CE — pale ice, all light.
	vivid("arctic-frost", "Arctic Frost", rgb(0x71, 0xC9, 0xCE), color.NRGBA{}).
		status(color.NRGBA{}, color.NRGBA{}, rgb(0xA6, 0xE3, 0xE9), color.NRGBA{}),

	// #00B8A9 #F8F3D4 #F6416C #FFDE7D — teal and watermelon on cream.
	vivid("tropical-punch", "Tropical Punch", rgb(0x00, 0xB8, 0xA9), color.NRGBA{}).
		status(rgb(0x00, 0xB8, 0xA9), rgb(0xFF, 0xDE, 0x7D), color.NRGBA{}, rgb(0xF6, 0x41, 0x6C)),

	// #48466D #3D84A8 #46CDCF #ABEDD8 — indigo fading to sea mint.
	vivid("ocean-drift", "Ocean Drift", rgb(0x3D, 0x84, 0xA8), rgb(0x48, 0x46, 0x6D)).
		status(rgb(0x46, 0xCD, 0xCF), color.NRGBA{}, rgb(0x3D, 0x84, 0xA8), color.NRGBA{}),

	// #2B2E4A #E84545 #903749 #53354A — bright red on midnight navy.
	vivid("cherry-noir", "Cherry Noir", rgb(0xE8, 0x45, 0x45), rgb(0x2B, 0x2E, 0x4A)).
		status(color.NRGBA{}, color.NRGBA{}, color.NRGBA{}, rgb(0xE8, 0x45, 0x45)),

	// #E23E57 #88304E #522546 #311D3F — raspberry deepening to plum.
	vivid("berry-wine", "Berry Wine", rgb(0xE2, 0x3E, 0x57), rgb(0x31, 0x1D, 0x3F)).
		status(color.NRGBA{}, color.NRGBA{}, color.NRGBA{}, rgb(0xE2, 0x3E, 0x57)),
}

// ByName resolves a preset slug; the default preset when unknown.
func ByName(slug string) Preset {
	for _, p := range Presets {
		if p.Slug == slug {
			return p
		}
	}
	return Presets[0]
}

// Hex renders a palette color for CSS emission.
func Hex(c color.NRGBA) string {
	return fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B)
}
