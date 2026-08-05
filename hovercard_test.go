package lotusui

import (
	"image"
	"testing"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

// Sharing one HoverCard across N call sites (e.g. every "GB" in a cost
// table) must not stack N floating panels. Only the active site paints.
func TestHoverCardMultiSiteOnlyActivePaints(t *testing.T) {
	th := NewTheme()
	var hc HoverCard
	hc.OpenDelay = time.Millisecond
	hc.open = true
	hc.active = 1

	var ops op.Ops
	gtx := layout.Context{
		Ops:         &ops,
		Constraints: layout.Exact(image.Pt(400, 200)),
		Now:         time.Now(),
	}
	body := LabelBody(th, "Gigabytes of files stored for one month").Layout
	trig := func(label string) layout.Widget {
		return LabelBody(th, label).Layout
	}

	// Three Layout calls = three sites; only active==1 may paint.
	hc.Layout(th, gtx, body, trig("GB"))
	if hc.sites.seq != 1 || hc.active != 1 {
		t.Fatalf("site0: seq=%d active=%d", hc.sites.seq, hc.active)
	}
	hc.Layout(th, gtx, body, trig("GB"))
	if hc.sites.seq != 2 {
		t.Fatalf("site1: seq=%d", hc.sites.seq)
	}
	hc.Layout(th, gtx, body, trig("GB"))
	if hc.sites.seq != 3 {
		t.Fatalf("site2: seq=%d", hc.sites.seq)
	}
	if len(hc.trigs) < 3 {
		t.Fatalf("expected 3 stable trigger tags, got %d", len(hc.trigs))
	}
	// Tags must stay stable across frames (event identity).
	t0, t1, t2 := hc.trigs[0], hc.trigs[1], hc.trigs[2]
	gtx.Now = gtx.Now.Add(time.Second)
	hc.Layout(th, gtx, body, trig("GB"))
	if hc.trigs[0] != t0 || hc.trigs[1] != t1 || hc.trigs[2] != t2 {
		t.Fatal("trigger tags must be heap-stable across frames")
	}
	if hc.sites.seq != 1 {
		t.Fatalf("new frame should reset seq, got %d", hc.sites.seq)
	}
}

// Width is a max: short content must not be stretched to 320dp, or a
// centered tip on "GB" still looks far from the word (empty chrome).
func TestHoverCardHugsContentWidth(t *testing.T) {
	th := NewTheme()
	var h HoverCard
	h.Width = 320

	var gotMin, gotMax int
	content := func(gtx layout.Context) layout.Dimensions {
		gotMin = gtx.Constraints.Min.X
		gotMax = gtx.Constraints.Max.X
		return layout.Dimensions{Size: image.Pt(48, 16)}
	}

	var ops op.Ops
	gtx := layout.Context{
		Ops:         &ops,
		Constraints: layout.Exact(image.Pt(400, 300)),
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
	}

	h.layoutCard(th, gtx, image.Pt(20, 14), content)
	if gotMin != 0 {
		t.Fatalf("content Min.X = %d, want 0 (hug; do not stretch to Width)", gotMin)
	}
	// UniformInset(MD) shrinks Max before content — still a cap, not a floor.
	wantMax := 320 - 2*gtx.Dp(th.Space.MD)
	if gotMax != wantMax {
		t.Fatalf("content Max.X = %d, want %d (Width minus inset)", gotMax, wantMax)
	}
}
