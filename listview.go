package lotusui

import (
	"image"

	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// The list family — a lotusui extension: Scrollable (whole-content
// scrolling for a screen's mixed content), ListView (its VIRTUALIZED
// sibling for long, uniform collections) and HoverRow (the one way a
// list row reads as interactive). State lives in the caller's
// widget.List / widget.Clickable, immediate-mode style.

// Scrollable makes arbitrarily-tall content reachable inside a bounded
// viewport by wrapping it in a single-item material.List — a scrollbar
// and trackpad/wheel scrolling instead of silent clipping.
//
// Prefer ScrollArea when content hosts Floating widgets (Select, Menu,
// Popover, …): layout.List records children and traps op.Defer portals.
// Scrollable stays useful when you want material.List's scrollbar and
// the content has no floating layer.
//
// content must not itself contain a layout.Flexed expecting to fill
// "remaining" vertical space — layout.List measures a list item against
// an effectively unbounded height. Give anything that needs a bounded
// pane inside scrollable content a fixed height instead (or use
// SplitBoxFillScroll for fill + pinned footer without an outer list).
//
// The list's viewport CLIPS its content (CSS overflow:hidden) — a card
// flush against the viewport edge would get its drop shadow flat-cut
// there. shadowRoom insets the content just enough that the shadow's
// widest ring (3dp grow + 2dp drop, see cardShadow) always has room to
// paint. Split pane helpers budget CardProps{}.Pad() instead — never nest
// Scrollable inside them by default (double inset).
const shadowRoom = unit.Dp(6)

func Scrollable(th *Theme, list *widget.List, gtx layout.Context, content layout.Widget) layout.Dimensions {
	list.Axis = layout.Vertical
	d := material.List(th.Material, list).Layout(gtx, 1, func(gtx layout.Context, _ int) layout.Dimensions {
		return layout.UniformInset(shadowRoom).Layout(gtx, content)
	})
	d.Size = gtx.Constraints.Constrain(d.Size)
	return d
}

// ListView is Scrollable's VIRTUALIZED sibling for long, uniform
// collections: only the rows intersecting the viewport are laid out,
// so a 10,000-row list costs a screenful per frame, not 10,000 rows.
// Scrollable lays its whole content out every frame — right for a
// screen's mixed content, wrong past a few dozen rows; reach for
// ListView the moment a collection can grow with data.
//
// list holds the scroll position and must outlive the frame; row lays
// out item i at the viewport's width.
func ListView(th *Theme, list *widget.List, gtx layout.Context, count int, row func(gtx layout.Context, i int) layout.Dimensions) layout.Dimensions {
	list.Axis = layout.Vertical
	d := material.List(th.Material, list).Layout(gtx, count, row)
	d.Size = gtx.Constraints.Constrain(d.Size)
	return d
}

// HoverRow makes one list item interactive the app's one way: the full
// row is clickable, shows the pointer (hand) cursor, and fills a
// rounded pill in a tint one step off the wrapper card's white
// (Surface2) while hovered — and keeps it while active (the selected
// row), the same treatment as the sidebar's active pill, so every
// interactive row in the app reads identically. content lays out bare;
// this wraps it in the row's padding.
func HoverRow(th *Theme, btn *widget.Clickable, active bool, content layout.Widget) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return btn.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
			gtx.Constraints.Min.Y = 0
			m := op.Record(gtx.Ops)
			dims := layout.UniformInset(th.Space.SM).Layout(gtx, content)
			call := m.Stop()
			defer clip.UniformRRect(image.Rectangle{Max: dims.Size}, gtx.Dp(th.Radius.SM)).Push(gtx.Ops).Pop()
			pointer.CursorPointer.Add(gtx.Ops)
			target := float32(0)
			if active || btn.Hovered() {
				target = 1
			}
			if active {
				// Selected rows stay filled; keep the clock warm.
				th.hoverToward(gtx, btn, 1)
				paint.Fill(gtx.Ops, th.Palette.BgSubtle)
			} else if m := th.hoverToward(gtx, btn, target); m > 0.01 {
				paint.Fill(gtx.Ops, fadeNRGBA(th.Palette.BgSubtle, m))
			}
			call.Add(gtx.Ops)
			return dims
		})
	}
}
