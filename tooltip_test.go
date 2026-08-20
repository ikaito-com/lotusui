package lotusui

import (
	"image"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
)

// TestFloatingTagsAreDistinct pins the zero-size trap: an event tag
// type with NO fields makes every new(T) return runtime.zerobase — one
// address shared by every instance in the program — so every Tooltip
// (and HoverCard) registered THE SAME tag with Gio, the first one laid
// out drained the hover event, and the label popped on the wrong
// widget. Any future tag type must carry a byte.
func TestFloatingTagsAreDistinct(t *testing.T) {
	if a, b := new(tipTrig), new(tipTrig); a == b {
		t.Errorf("tipTrig is zero-sized: two tags share address %p", a)
	}
	if a, b := new(hoverTrig), new(hoverTrig); a == b {
		t.Errorf("hoverTrig is zero-sized: two tags share address %p", a)
	}
}

// TestTooltipHoverStaysOnItsOwnSite is the end-to-end proof: two
// Tooltips laid out one above the other, the pointer resting on the
// SECOND. Only the second may consider itself hovered — the docs page
// showed every tooltip in the first example box.
func TestTooltipHoverStaysOnItsOwnSite(t *testing.T) {
	th := NewTheme()
	var first, second Tooltip
	var b1, b2 widget.Clickable

	var ops op.Ops
	var r input.Router
	now := time.Unix(1, 0)
	frame := func() {
		ops.Reset()
		// Each frame gets its own clock: layoutSites numbers the call
		// sites within ONE frame and resets when gtx.Now moves.
		now = now.Add(16 * time.Millisecond)
		gtx := testCtx(&ops, &r, layout.Constraints{Max: image.Pt(400, 400)})
		gtx.Now = now
		layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return first.Layout(th, gtx, "first", Button(th, &b1, "One", ButtonProps{}))
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return second.Layout(th, gtx, "second", Button(th, &b2, "Two", ButtonProps{}))
			}),
		)
		r.Frame(&ops)
	}

	frame() // register the areas
	// Rest the pointer deep inside the SECOND trigger. The first button
	// is ~40dp tall at 1:1, so y=60 is past it.
	r.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: f32.Pt(30, 60)})
	frame() // deliver Enter

	if first.anyOver() {
		t.Error("the first tooltip reports hover for a pointer over the second")
	}
	if !second.anyOver() {
		t.Error("the hovered tooltip never saw its own Enter")
	}
}
