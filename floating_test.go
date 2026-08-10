package lotusui

import (
	"image"
	"testing"
	"time"

	"gioui.org/widget"

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

// TestFloatingWidgetsInsideCard lays every floating consumer inside a
// Card — the shape that runs a THROWAWAY pass before the live one.
// HoverCard, Tooltip and the menu triggers index per-site state by the
// site number, so a measure pass that reports "no site" must not reach
// that indexing: doing so panicked the whole Gio program, and in wasm
// that reads as an endless "Go program has already exited" once the
// dead callbacks keep firing on hover.
func TestFloatingWidgetsInsideCard(t *testing.T) {
	th := NewTheme()
	var (
		hc      HoverCard
		tip     Tooltip
		sub     DropdownMenuSub
		trigger DropdownMenuTrigger
		cm      ContextMenu
		btn     widget.Clickable
		sel     = Select{Options: SelectOpts("A", "B")}
		pop     Popover
	)
	body := func(gtx layout.Context) layout.Dimensions { return LabelBody(th, "body").Layout(gtx) }

	cases := map[string]layout.Widget{
		"HoverCard": func(gtx layout.Context) layout.Dimensions {
			return hc.Layout(th, gtx, body, LabelBody(th, "trigger").Layout)
		},
		"Tooltip": func(gtx layout.Context) layout.Dimensions {
			return tip.Layout(th, gtx, "tip", LabelBody(th, "trigger").Layout)
		},
		"DropdownMenuTrigger": func(gtx layout.Context) layout.Dimensions {
			return trigger.Layout(th, gtx, "Open", DropdownMenuItem(th, &btn, "One", false))
		},
		"DropdownMenuSub": func(gtx layout.Context) layout.Dimensions {
			return sub.Item(th, "More", DropdownMenuItem(th, &btn, "One", false))(gtx)
		},
		"ContextMenu": func(gtx layout.Context) layout.Dimensions {
			return cm.Layout(th, gtx, LabelBody(th, "area").Layout,
				ContextMenuItem(th, &btn, "One", false))
		},
		"Select": func(gtx layout.Context) layout.Dimensions {
			return sel.Layout(th, gtx, "Env")
		},
		"Popover": func(gtx layout.Context) layout.Dimensions {
			d := LabelBody(th, "anchor").Layout(gtx)
			pop.Layout(th, gtx, d.Size, body)
			return d
		},
	}
	for name, w := range cases {
		t.Run(name, func(t *testing.T) {
			var ops op.Ops
			var r input.Router
			gtx := testCtx(&ops, &r, layout.Constraints{Max: image.Pt(600, 400)})
			// Card measures, then paints — two passes, one frame.
			Card(th, CardProps{}, w)(gtx)
			// And again inside a Grid, whose scratch pass measures too.
			Grid{Columns: 2, Gap: Space.SM}.Layout(th, gtx, Cell(w), Cell(w))
		})
	}
}
