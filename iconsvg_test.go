package lotusui

import (
	"strings"
	"testing"
)

// Every icon the manifest promises must be embedded AND rasterize to
// VISIBLE pixels — a bad download (404 page) or an unsupported SVG
// feature (gradient fills that once rendered fully transparent) must
// fail the build, not silently render an empty square.
func TestEmbeddedIconsRasterize(t *testing.T) {
	entries, err := iconFS.ReadDir("assets/icons")
	if err != nil {
		t.Fatal(err)
	}
	found := 0
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".svg") {
			continue
		}
		found++
		name := strings.TrimSuffix(e.Name(), ".svg")
		img := renderSVGIcon(name, 32, rgb(0x33, 0x33, 0x33))
		if img == nil {
			t.Errorf("icon %q failed to parse/rasterize", name)
			continue
		}
		visible := 0
		for i := 3; i < len(img.Pix); i += 4 {
			if img.Pix[i] > 0 {
				visible++
			}
		}
		// A 32x32 icon should paint a meaningful share of its square;
		// near-zero means the shapes silently failed to fill.
		if visible < 32 {
			t.Errorf("icon %q rendered only %d visible pixels of 1024 — its fills aren't drawing", name, visible)
		}
	}
	if found == 0 {
		t.Fatal("no SVG icons embedded — run `make icons`")
	}
}
