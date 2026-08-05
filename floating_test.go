package lotusui

import (
	"image"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
)

// TestMeasurePassClaimsNoSite pins the mechanism: floating consumers
// paint only at site 0, so a discarded pass must leave the counter
// untouched. Without this, the live pass looks like a duplicate.
func TestMeasurePassClaimsNoSite(t *testing.T) {
	var s layoutSites
	now := time.Unix(1, 0)

	beginMeasurePass()
	if got := s.next(now); got != -1 {
		t.Errorf("measure pass site = %d, want -1 (never paints)", got)
	}
	endMeasurePass()

	if got := s.next(now); got != 0 {
		t.Errorf("live pass site = %d, want 0 — the panel would not paint", got)
	}
	if got := s.next(now); got != 1 {
		t.Errorf("second live site = %d, want 1 (one shared widget, two sites)", got)
	}
}

// TestFloatingSurvivesCardMeasurePass is the regression that blocked
// vaultalia at v0.1.0: Card lays its content out TWICE per frame (a
// discarded pass to size its chrome, then the live one) with the same
// gtx.Now. The discarded pass used to claim site 0, so every Select,
// Popover or menu inside a Card was suppressed and its panel never
// appeared — while the same widget outside a Card opened fine.
//
// It asserts through the real Card code path, and deliberately not on
// returned Dimensions: the bug is invisible to size assertions.
func TestFloatingSurvivesCardMeasurePass(t *testing.T) {
	th := NewTheme()
	var ops op.Ops
	var r input.Router
	gtx := testCtx(&ops, &r, layout.Constraints{Max: image.Pt(600, 400)})

	var sites layoutSites
	var got []int
	content := func(gtx layout.Context) layout.Dimensions {
		got = append(got, sites.next(gtx.Now))
		return layout.Dimensions{Size: image.Pt(120, 32)}
	}
	Card(th, CardProps{}, content)(gtx)

	if len(got) < 2 {
		t.Fatalf("Card laid its content out %d time(s); this test assumes the measure+live pair", len(got))
	}
	live := 0
	for _, idx := range got {
		if idx == 0 {
			live++
		}
	}
	if live != 1 {
		t.Fatalf("sites seen inside Card = %v — exactly one pass must be site 0", got)
	}
	if got[len(got)-1] != 0 {
		t.Errorf("the LAST (painting) pass got site %d, want 0 — floating content would be suppressed", got[len(got)-1])
	}
}

// TestSelectInCardLaysOutOpen is the shape of the reported repro: an
// open Select inside a Card must lay out without panicking, at the
// hostile constraints the smoke test uses elsewhere.
func TestSelectInCardLaysOutOpen(t *testing.T) {
	th := NewTheme()
	sel := Select{Options: SelectOpts("Alpha", "Beta", "Gamma")}
	sel.open = true
	for _, cs := range []layout.Constraints{
		{Max: image.Pt(600, 400)},
		{Max: image.Pt(3, 2000)},
		{Min: image.Pt(320, 200), Max: image.Pt(320, 200)},
	} {
		var ops op.Ops
		var r input.Router
		gtx := testCtx(&ops, &r, cs)
		Card(th, CardProps{}, func(gtx layout.Context) layout.Dimensions {
			return sel.Layout(th, gtx, "Env")
		})(gtx)
	}
}
