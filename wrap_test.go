package lotusui

import (
	"image"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

func TestWrapTwoWideChildrenTwoLines(t *testing.T) {
	var ops op.Ops
	gtx := layout.Context{
		Ops: &ops,
		Constraints: layout.Constraints{
			Max: image.Pt(100, 1<<20),
		},
		Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1},
	}
	wide := func(gtx layout.Context) layout.Dimensions {
		// Intrinsic ~80 wide; two with gap won't fit in 100.
		return layout.Dimensions{Size: image.Pt(80, 20)}
	}
	dims := Wrap(unit.Dp(8), layout.Middle, wide, wide)(gtx)
	// One line would be 80+8+80 = 168; wrapped height ≥ 20+8+20 = 48.
	if dims.Size.Y < 48 {
		t.Fatalf("Wrap height = %d, want ≥ 48 (two lines)", dims.Size.Y)
	}
	if dims.Size.X > 100 {
		t.Fatalf("Wrap width = %d, exceeds Max.X 100", dims.Size.X)
	}
}

func TestWrapDoesNotSqueezeToOneColumn(t *testing.T) {
	var ops op.Ops
	gtx := layout.Context{
		Ops: &ops,
		Constraints: layout.Constraints{
			Max: image.Pt(50, 1<<20),
		},
		Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1},
	}
	var gotMaxX int
	child := func(gtx layout.Context) layout.Dimensions {
		gotMaxX = gtx.Constraints.Max.X
		// Report natural 40×16 if allowed, else fill Max.
		w := 40
		if w > gtx.Constraints.Max.X {
			w = gtx.Constraints.Max.X
		}
		return layout.Dimensions{Size: image.Pt(w, 16)}
	}
	Wrap(unit.Dp(4), layout.Start, child, child, child)(gtx)
	// Must not measure with Max.X squeezed to ~1 (letter-stack).
	if gotMaxX < 40 && gotMaxX != 50 {
		// Last child laid out: either intrinsic 40 on a line alone, or
		// capped at Max.X 50 when alone-wider. Never a 1px squeeze.
		t.Fatalf("child Max.X = %d; Wrap must measure intrinsic, not squeeze", gotMaxX)
	}
}
