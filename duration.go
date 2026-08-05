package lotusui

import (
	"image/color"
	"time"

	"gioui.org/layout"
	"gioui.org/widget"
)

// DurationScale is a theme's motion timing ladder — Chakra-style
// duration tokens, resolved once at NewTheme. Components pick a step
// by role: Fast for hover/color, Normal for small motion (switch,
// accordion), Slow for overlay entrances (dialog, slides).
//
// A zero step snaps (no interpolation) — useful for tests or a
// reduced-motion preference.
type DurationScale struct {
	Fast   time.Duration // hover / color — default 150ms
	Normal time.Duration // small motion — default 200ms
	Slow   time.Duration // overlay enter — default 300ms
}

// Duration holds the default motion scale. Treat it as read-only:
// per-app customization goes through NewTheme(WithDuration(...)),
// which components read as th.Duration. Split (no Theme on Layout)
// reads this package value for its strip clock.
var Duration = DurationScale{
	Fast:   150 * time.Millisecond,
	Normal: 200 * time.Millisecond,
	Slow:   300 * time.Millisecond,
}

// WithDuration replaces the motion timing scale. It also updates the
// package-level Duration used by Split (whose Layout has no Theme).
func WithDuration(d DurationScale) ThemeOption {
	return func(t *Theme) {
		t.Duration = d
		Duration = d
	}
}

// hoverToward eases a Clickable-keyed 0…1 clock toward target using
// Duration.Fast. State lives on the Theme for the life of each
// Clickable pointer (same retention model as the widgets themselves).
func (t *Theme) hoverToward(gtx layout.Context, btn *widget.Clickable, target float32) float32 {
	if btn == nil {
		return target
	}
	d := t.Duration.Fast
	if d <= 0 {
		return target
	}
	if t.hovers == nil {
		t.hovers = make(map[*widget.Clickable]*slideAnim)
	}
	a := t.hovers[btn]
	if a == nil {
		a = &slideAnim{}
		t.hovers[btn] = a
	}
	return a.advance(gtx, target, d)
}

// lerpNRGBA blends a→b by t in [0,1], including alpha.
func lerpNRGBA(a, b color.NRGBA, t float32) color.NRGBA {
	if t <= 0 {
		return a
	}
	if t >= 1 {
		return b
	}
	return color.NRGBA{
		R: uint8(float32(a.R) + (float32(b.R)-float32(a.R))*t + 0.5),
		G: uint8(float32(a.G) + (float32(b.G)-float32(a.G))*t + 0.5),
		B: uint8(float32(a.B) + (float32(b.B)-float32(a.B))*t + 0.5),
		A: uint8(float32(a.A) + (float32(b.A)-float32(a.A))*t + 0.5),
	}
}
