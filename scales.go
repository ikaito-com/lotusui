package lotusui

import (
	"image/color"
	"math"
)

// A ColorScale is one hue graded from near-white to near-black in ten
// steps — the 50…900 convention: 50 is the faintest tint (subtle
// backgrounds), 500 the anchor (solid fills), 900 the darkest ink.
// Scales are the raw material palettes and schemes are built from;
// they are plain values resolved at startup, never computed per frame.
type ColorScale struct {
	C50, C100, C200, C300, C400, C500, C600, C700, C800, C900 color.NRGBA
}

// ScaleFrom grades a full 50…900 scale from one anchor color: the
// anchor keeps its hue and saturation and becomes C500; lighter steps
// interpolate from the anchor toward near-white, darker steps toward
// near-black — RELATIVE to the anchor's own lightness, so the ladder
// stays strictly ordered whether the brand color is a pastel or a deep
// ink. One brand color in, a complete usable scale out.
func ScaleFrom(anchor color.NRGBA) ColorScale {
	h, s, l := rgbToHSL(anchor)
	lighter := func(t float64) color.NRGBA { return hslToRGB(h, s, l+(0.97-l)*t) }
	darker := func(t float64) color.NRGBA { return hslToRGB(h, s, l-(l-0.14)*t) }
	return ColorScale{
		C50:  lighter(0.92),
		C100: lighter(0.80),
		C200: lighter(0.60),
		C300: lighter(0.40),
		C400: lighter(0.20),
		C500: anchor,
		C600: darker(0.25),
		C700: darker(0.45),
		C800: darker(0.65),
		C900: darker(0.85),
	}
}

// The stock scales (Gray, Red, … Pink) live in scales_gen.go as
// LITERALS — graded at build time by the go:generate below, so the
// package carries plain values: zero init-time math, every hex
// visible in the source and on pkg.go.dev.
//
//go:generate go run ./cmd/lotusui gen-scales -o scales_gen.go

// Scheme maps the scale onto the button variants the SATURATED way:
// C500 solid fill with white text, faint tint for subtle — and the
// interaction ladder exactly as the web convention has it: base .500,
// hover .600, active .700 (and .100/.200/.300 for the subtle family).
func (s ColorScale) Scheme() Scheme {
	return Scheme{
		Solid: s.C500, SolidHover: s.C600, SolidActive: s.C700,
		OnSolid: rgb(0xFF, 0xFF, 0xFF),
		Subtle:  s.C100, SubtleHover: s.C200, SubtleActive: s.C300,
		OnSubtle: s.C700,
		Outline:  s.C500,
	}
}

// SoftScheme maps the scale onto the button variants the PASTEL way
// lotusui's own defaults use: tinted fill with deep same-hue ink —
// quieter than Scheme, right for interfaces that lead with content.
// Interaction steps walk the light end of the ladder.
func (s ColorScale) SoftScheme() Scheme {
	return Scheme{
		Solid: s.C100, SolidHover: s.C200, SolidActive: s.C300,
		OnSolid: s.C800,
		Subtle:  s.C50, SubtleHover: s.C100, SubtleActive: s.C200,
		OnSubtle: s.C600,
		Outline:  s.C600,
	}
}

// ── HSL plumbing ─────────────────────────────────────────────────────────

func rgbToHSL(c color.NRGBA) (h, s, l float64) {
	r := float64(c.R) / 255
	g := float64(c.G) / 255
	b := float64(c.B) / 255
	max := math.Max(r, math.Max(g, b))
	min := math.Min(r, math.Min(g, b))
	l = (max + min) / 2
	if max == min {
		return 0, 0, l
	}
	d := max - min
	if l > 0.5 {
		s = d / (2 - max - min)
	} else {
		s = d / (max + min)
	}
	switch max {
	case r:
		h = (g - b) / d
		if g < b {
			h += 6
		}
	case g:
		h = (b-r)/d + 2
	default:
		h = (r-g)/d + 4
	}
	return h / 6, s, l
}

func hslToRGB(h, s, l float64) color.NRGBA {
	if s == 0 {
		v := uint8(math.Round(l * 255))
		return color.NRGBA{R: v, G: v, B: v, A: 0xFF}
	}
	var q float64
	if l < 0.5 {
		q = l * (1 + s)
	} else {
		q = l + s - l*s
	}
	p := 2*l - q
	conv := func(t float64) uint8 {
		if t < 0 {
			t++
		}
		if t > 1 {
			t--
		}
		var v float64
		switch {
		case t < 1.0/6:
			v = p + (q-p)*6*t
		case t < 1.0/2:
			v = q
		case t < 2.0/3:
			v = p + (q-p)*(2.0/3-t)*6
		default:
			v = p
		}
		return uint8(math.Round(v * 255))
	}
	return color.NRGBA{R: conv(h + 1.0/3), G: conv(h), B: conv(h - 1.0/3), A: 0xFF}
}
