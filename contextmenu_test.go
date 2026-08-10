package lotusui

import (
	"image"
	"testing"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
)

// TestContextMenuPlace pins the placement rules: down-right of the
// pointer, flip up/left at a known edge (the native menu move), clamp
// when neither side fits, and no flipping at all against the 2^14
// "infinite" sentinel — inside a scroller the window edge is unknowable.
func TestContextMenuPlace(t *testing.T) {
	panel := image.Pt(200, 150)
	avail := image.Pt(600, 400)
	cases := []struct {
		name         string
		press, avail image.Point
		want         image.Point
	}{
		{"fits down-right", image.Pt(10, 20), avail, image.Pt(10, 20)},
		{"right edge flips left", image.Pt(500, 20), avail, image.Pt(300, 20)},
		{"bottom edge flips up", image.Pt(10, 380), avail, image.Pt(10, 230)},
		{"corner flips both", image.Pt(500, 380), avail, image.Pt(300, 230)},
		{"no room either side clamps", image.Pt(50, 20), image.Pt(220, 400), image.Pt(20, 20)},
		{"panel wider than avail clamps to 0", image.Pt(10, 20), image.Pt(120, 400), image.Pt(0, 20)},
		{"unbounded never flips", image.Pt(16000, 16000), image.Pt(1<<14, 1<<14), image.Pt(16000, 16000)},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := contextMenuPlace(c.press, panel, c.avail); got != c.want {
				t.Errorf("place(%v, %v, %v) = %v, want %v", c.press, panel, c.avail, got, c.want)
			}
		})
	}
}

// TestContextMenuOpenLaysOut lays an OPEN ContextMenu out at the
// hostile constraint shapes, inside a Card — the measure+live double
// pass that broke floating widgets twice before this family existed.
func TestContextMenuOpenLaysOut(t *testing.T) {
	th := NewTheme()
	var btn widget.Clickable
	cm := ContextMenu{}
	cm.open = true
	cm.at = image.Pt(40, 30)
	for _, cs := range []layout.Constraints{
		{Max: image.Pt(600, 400)},
		{Max: image.Pt(3, 2000)},
		{Min: image.Pt(320, 200), Max: image.Pt(320, 200)},
	} {
		var ops op.Ops
		var r input.Router
		gtx := testCtx(&ops, &r, cs)
		Card(th, CardProps{}, func(gtx layout.Context) layout.Dimensions {
			return cm.Layout(th, gtx, LabelBody(th, "area").Layout,
				ContextMenuItem(th, &btn, "Back", false),
				ContextMenuSeparator(th),
				ContextMenuItem(th, &btn, "Delete", true),
			)
		})(gtx)
	}
}
