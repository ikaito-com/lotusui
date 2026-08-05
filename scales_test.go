package lotusui

import (
	"image/color"
	"testing"
)

// A scale must grade strictly light→dark whatever the anchor's own
// lightness — a pastel brand and a deep ink brand both have to produce
// ten usable, ordered steps.
func TestScaleFromMonotonic(t *testing.T) {
	anchors := map[string]color.NRGBA{
		"mid":    rgb(0x31, 0x82, 0xCE),
		"dark":   rgb(0x31, 0x97, 0x95), // teal: darker than the ladder midpoint
		"pastel": rgb(0xE0, 0xD9, 0xFF), // lavender: far lighter than midpoint
		"ink":    rgb(0x2F, 0x25, 0x66),
	}
	for name, a := range anchors {
		s := ScaleFrom(a)
		steps := []color.NRGBA{s.C50, s.C100, s.C200, s.C300, s.C400, s.C500, s.C600, s.C700, s.C800, s.C900}
		prev := 2.0
		for i, c := range steps {
			_, _, l := rgbToHSL(c)
			if l >= prev {
				t.Errorf("%s: step %d lightness %.3f not darker than previous %.3f", name, i, l, prev)
			}
			prev = l
		}
	}
}
