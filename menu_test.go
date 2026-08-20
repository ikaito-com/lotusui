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

func menuTestItems(th *Theme) []layout.Widget {
	var b1, b2, b3 widget.Clickable
	return []layout.Widget{
		DropdownMenuLabel(th, "Navigation"),
		DropdownMenuItem(th, &b1, "Back", false),
		DropdownMenuItem(th, &b2, "Forward", false),
		DropdownMenuSeparator(th),
		DropdownMenuShortcutItem(th, &b3, "Reload", "⌘R", false),
	}
}

// TestMenuPanelHugsUnboundedMax pins the rule the floating menus rely
// on: rows fill a REAL panel width, but against Gio's 2^14 "infinite"
// max they report their intrinsic width instead. Filling it painted a
// 16384px panel — the context menu ran off the window with its corners
// and border out of sight.
func TestMenuPanelHugsUnboundedMax(t *testing.T) {
	th := NewTheme()
	var ops op.Ops
	var r input.Router
	gtx := testCtx(&ops, &r, layout.Constraints{Max: image.Pt(900, 700)})

	gtx.Constraints = layout.Constraints{Max: image.Pt(1<<14, 1<<14)}
	hug := DropdownMenu(th, menuTestItems(th)...)(gtx)
	if hug.Size.X >= unbounded {
		t.Fatalf("panel filled the unbounded max: width %d", hug.Size.X)
	}
	if hug.Size.X < 40 {
		t.Fatalf("panel hugged to nothing: width %d", hug.Size.X)
	}

	// Against a real width the rows still span the panel.
	gtx.Constraints = layout.Constraints{Min: image.Pt(300, 0), Max: image.Pt(300, 1<<14)}
	filled := DropdownMenu(th, menuTestItems(th)...)(gtx)
	if filled.Size.X != 300 {
		t.Errorf("bounded panel = %d wide, want 300", filled.Size.X)
	}
}

// TestMenuPanelWidthClamps checks the resolved panel width: at least
// the shadcn min-w, at most the cap (a Width prop, or the available
// window), and always hugging the content in between.
func TestMenuPanelWidthClamps(t *testing.T) {
	th := NewTheme()
	var ops op.Ops
	var r input.Router
	gtx := testCtx(&ops, &r, layout.Constraints{Max: image.Pt(900, 700)})
	items := menuTestItems(th)
	natural := menuPanelWidth(th, gtx, 0, 1<<14, items...)

	cases := []struct {
		name       string
		minW, maxW int
		want       int
	}{
		{"hugs between the bounds", 0, 1 << 14, natural},
		{"min-w floors a narrow menu", natural + 120, 1 << 14, natural + 120},
		{"cap clamps a wide menu", 0, natural - 40, natural - 40},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := menuPanelWidth(th, gtx, c.minW, c.maxW, items...); got != c.want {
				t.Errorf("menuPanelWidth(min %d, max %d) = %d, want %d", c.minW, c.maxW, got, c.want)
			}
		})
	}
}

// TestContextMenuPanelFitsWindow lays an OPEN ContextMenu out at a
// window-sized constraint and requires the panel it would paint to fit
// — the docs page showed one spilling far past the right edge.
func TestContextMenuPanelFitsWindow(t *testing.T) {
	th := NewTheme()
	avail := image.Pt(900, 700)
	var ops op.Ops
	var r input.Router
	gtx := testCtx(&ops, &r, layout.Constraints{Max: avail})
	items := menuTestItems(th)

	minW := gtx.Dp(224)
	if w := menuPanelWidth(th, gtx, minW, avail.X, items...); w > avail.X {
		t.Errorf("panel %d wide inside a %d window", w, avail.X)
	}

	cm := ContextMenu{}
	cm.open = true
	cm.at = image.Pt(40, 30)
	var btn widget.Clickable
	cm.Layout(th, gtx, Button(th, &btn, "Right-click", ButtonProps{}), items...)
}

// TestContextMenuPanelPaintsNarrow is the end-to-end proof, measured
// the way a user meets it: the menu rows own the pointer cursor, so a
// pointer far to the right of a hugged panel must NOT be over a row.
// With rows filling the 2^14 max they reached x=16384 and every point
// on the line was "inside the menu".
func TestContextMenuPanelPaintsNarrow(t *testing.T) {
	th := NewTheme()
	var cm ContextMenu
	var btn widget.Clickable
	items := menuTestItems(th)

	var ops op.Ops
	var r input.Router
	now := time.Unix(1, 0)
	frame := func() {
		ops.Reset()
		now = now.Add(16 * time.Millisecond)
		gtx := testCtx(&ops, &r, layout.Constraints{Max: image.Pt(900, 700)})
		gtx.Now = now
		cm.Layout(th, gtx, Button(th, &btn, "Right-click me", ButtonProps{}), items...)
		r.Frame(&ops)
	}

	frame()
	// The platform context gesture, at the top-left of the content.
	r.Queue(pointer.Event{
		Kind: pointer.Press, Source: pointer.Mouse,
		Buttons: pointer.ButtonSecondary, Position: f32.Pt(20, 10),
	})
	frame() // opens
	if !cm.open {
		t.Fatal("the context gesture did not open the menu")
	}
	frame() // paints the panel with the pointer parked at the press

	// x=600 is far outside a hugged panel (min-w is 224dp) and well
	// inside a 16384px one. Land on the first row's line.
	r.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: f32.Pt(600, 40)})
	frame()
	if got := r.Cursor(); got == pointer.CursorPointer {
		t.Errorf("a menu row still covers x=600: the panel did not hug (cursor %v)", got)
	}
}
