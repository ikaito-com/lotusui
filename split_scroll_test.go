package lotusui

import (
	"image"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
)

func TestSplitBoxScrollContentMaxYAccountsForPad(t *testing.T) {
	th := NewTheme()
	var list widget.List
	var ops op.Ops
	outer := 200
	gtx := layout.Context{
		Ops:         &ops,
		Constraints: layout.Constraints{Max: image.Pt(300, outer)},
	}
	padPx := gtx.Dp(CardProps{}.Pad())
	want := outer - 2*padPx
	var got int
	tall := func(gtx layout.Context) layout.Dimensions {
		got = gtx.Constraints.Max.Y
		return layout.Dimensions{Size: image.Pt(gtx.Constraints.Max.X, 2000)}
	}
	SplitBoxScroll(th, &list, outer, tall)(gtx)
	if got != want {
		t.Fatalf("content Max.Y = %d, want outer-2*Pad = %d (pad=%d)", got, want, padPx)
	}
}

func TestSplitBoxFillScrollReassertsMinY(t *testing.T) {
	th := NewTheme()
	var list widget.List
	var ops op.Ops
	outer := 200
	gtx := layout.Context{
		Ops:         &ops,
		Constraints: layout.Constraints{Max: image.Pt(300, outer)},
	}
	padPx := gtx.Dp(CardProps{}.Pad())
	want := outer - 2*padPx
	var gotMin, gotMax int
	SplitBoxFillScroll(th, &list, outer, func(gtx layout.Context) layout.Dimensions {
		gotMin, gotMax = gtx.Constraints.Min.Y, gtx.Constraints.Max.Y
		return layout.Dimensions{Size: image.Pt(gtx.Constraints.Max.X, gotMax)}
	})(gtx)
	if gotMin != want || gotMax != want {
		t.Fatalf("content Min/Max.Y = %d/%d, want %d/%d", gotMin, gotMax, want, want)
	}
}
