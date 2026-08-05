package lotusui

import (
	"image/color"
	"os"
	"strings"
	"testing"
)

// TestIconsAreNotSolidBlocks catches SVG paths our rasterizer fills
// solid instead of drawing: a glyph inks a fraction of its box, a
// degenerate path inks nearly all of it.
func TestIconsAreNotSolidBlocks(t *testing.T) {
	ents, err := os.ReadDir("assets/icons")
	if err != nil {
		t.Skip("no icons dir")
	}
	for _, e := range ents {
		if !strings.HasSuffix(e.Name(), ".svg") {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".svg")
		img := renderSVGIcon(name, 24, color.NRGBA{R: 0, G: 0, B: 0, A: 255})
		if img == nil {
			t.Errorf("%s: did not rasterize", name)
			continue
		}
		inked := 0
		b := img.Bounds()
		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				if _, _, _, a := img.At(x, y).RGBA(); a > 0x4000 {
					inked++
				}
			}
		}
		frac := float64(inked) / float64(b.Dx()*b.Dy())
		if frac > 0.80 {
			t.Errorf("%s: %.0f%% of the box is inked — path renders as a solid block", name, frac*100)
		}
	}
}
