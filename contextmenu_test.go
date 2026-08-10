package lotusui

import (
	"image"
	"testing"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"gioui.org/widget/material"
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

// TestEditorContextMenuInsideCard is the vaultalia lesson applied
// up-front: a widget that appears only with real data needs a
// headless layout test inside a real Card — the measure+live double
// pass. It also pins the selection contract: laying the open menu out
// must leave the editor's selection intact, because Copy acts on it.
func TestEditorContextMenuInsideCard(t *testing.T) {
	th := NewTheme()
	var ecm EditorContextMenu
	var ed widget.Editor
	ed.ReadOnly = true
	ed.SetText("alpha\nbeta\ngamma")
	ed.SetCaret(0, 5)
	ecm.Menu.open = true
	ecm.Menu.at = image.Pt(12, 10)
	for i := 0; i < 2; i++ {
		var ops op.Ops
		var r input.Router
		gtx := testCtx(&ops, &r, layout.Constraints{Max: image.Pt(600, 400)})
		Card(th, CardProps{}, func(gtx layout.Context) layout.Dimensions {
			return ecm.Layout(th, gtx, &ed, func(gtx layout.Context) layout.Dimensions {
				return material.Editor(th.Material, &ed, "").Layout(gtx)
			})
		})(gtx)
		r.Frame(&ops)
	}
	if got := ed.SelectedText(); got != "alpha" {
		t.Errorf("selection after layout = %q, want %q — Copy would copy the wrong text", got, "alpha")
	}
}
