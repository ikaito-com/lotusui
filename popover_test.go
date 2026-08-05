package lotusui

import (
	"image"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

// Popover.Width non-zero must not force Min.X — same hug doctrine as HoverCard.
func TestPopoverWidthHugsContent(t *testing.T) {
	th := NewTheme()
	var p Popover
	p.Open = true
	p.Width = 320

	var gotMin int
	content := func(gtx layout.Context) layout.Dimensions {
		gotMin = gtx.Constraints.Min.X
		return layout.Dimensions{Size: image.Pt(40, 20)}
	}

	var ops op.Ops
	gtx := layout.Context{
		Ops:         &ops,
		Constraints: layout.Exact(image.Pt(400, 300)),
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
	}
	p.Layout(th, gtx, image.Pt(80, 24), content)
	// Content is inside UniformInset(MD); Min should still be 0 at the
	// outer constraint — inset only shrinks Max.
	if gotMin != 0 {
		t.Fatalf("content Min.X = %d, want 0 (Width is max, not fixed)", gotMin)
	}
}

func TestPopoverZeroWidthMatchesAnchor(t *testing.T) {
	th := NewTheme()
	var p Popover
	p.Open = true

	var gotMin, gotMax int
	content := func(gtx layout.Context) layout.Dimensions {
		gotMin = gtx.Constraints.Min.X
		gotMax = gtx.Constraints.Max.X
		return layout.Dimensions{Size: image.Pt(10, 10)}
	}

	var ops op.Ops
	gtx := layout.Context{
		Ops:         &ops,
		Constraints: layout.Exact(image.Pt(400, 300)),
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
	}
	anchor := image.Pt(120, 28)
	p.Layout(th, gtx, anchor, content)
	// Inside inset: Min/Max reduced by 2*MD from outer Min=Max=120.
	pad := 2 * gtx.Dp(th.Space.MD)
	if gotMin != 120-pad || gotMax != 120-pad {
		t.Fatalf("zero Width: content constraints Min=%d Max=%d, want both %d (match anchor)", gotMin, gotMax, 120-pad)
	}
}
