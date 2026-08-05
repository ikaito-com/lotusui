package lotusui

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// Split scroll helpers — lotusui EXTENSION for panes inside Split.
// Scrollable insets shadowRoom for page-level card shadows; these
// helpers budget CardProps{}.Pad() instead and never nest Scrollable
// (that would double-inset). material.List measures items with
// effectively unbounded Max.Y, so Flexed "fill remaining" inside a
// list item cannot work — use SplitBoxFillScroll for fill + pinned
// footer, not an outer list.

// paneListStyle is the shared scroller for Split panes: track flush
// with content (no Gio default 2dp MajorPadding gap under titles/tabs).
func paneListStyle(th *Theme, list *widget.List) material.ListStyle {
	ls := material.List(th.Material, list)
	ls.Track.MajorPadding = 0
	return ls
}

// SplitColumnScroll is a fixed-height column viewport. Stack
// natural-height SplitBoxes inside; the COLUMN scrolls, not each card.
// maxH <= 0 uses Constraints.Max.Y. Size is always (Max.X, maxH).
// Inside a list item Max.Y is unbounded — pass an explicit maxH.
func SplitColumnScroll(th *Theme, list *widget.List, maxH int, content layout.Widget) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		h := maxH
		if h <= 0 {
			h = gtx.Constraints.Max.Y
		}
		if h > gtx.Constraints.Max.Y {
			h = gtx.Constraints.Max.Y
		}
		gtx.Constraints.Min = image.Pt(gtx.Constraints.Max.X, h)
		gtx.Constraints.Max = gtx.Constraints.Min
		list.Axis = layout.Vertical
		d := paneListStyle(th, list).Layout(gtx, 1, func(gtx layout.Context, _ int) layout.Dimensions {
			return content(gtx)
		})
		d.Size = gtx.Constraints.Constrain(d.Size)
		return d
	}
}

// SplitBoxScroll hugs while the body is shorter than maxH (0 = Max.Y);
// past that the body scrolls inside the card. Content Max.Y is limited
// to maxH − 2×CardProps{}.Pad() so children do not paint under the pad.
func SplitBoxScroll(th *Theme, list *widget.List, maxH int, content layout.Widget) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		limit := maxH
		if limit <= 0 {
			limit = gtx.Constraints.Max.Y
		}
		if limit > gtx.Constraints.Max.Y {
			limit = gtx.Constraints.Max.Y
		}
		pad := gtx.Dp(CardProps{}.Pad())
		innerW := gtx.Constraints.Max.X - 2*pad
		if innerW < 0 {
			innerW = 0
		}
		// Measure natural body height without writing to the frame.
		var mop op.Ops
		mgtx := gtx
		mgtx.Ops = &mop
		mgtx.Constraints.Min = image.Pt(innerW, 0)
		mgtx.Constraints.Max = image.Pt(innerW, 1<<30)
		nat := content(mgtx)
		if nat.Size.Y+2*pad <= limit {
			return SplitBox(th, content)(gtx)
		}
		bodyMax := limit - 2*pad
		if bodyMax < 0 {
			bodyMax = 0
		}
		cgtx := gtx
		cgtx.Constraints.Min = image.Pt(gtx.Constraints.Max.X, limit)
		cgtx.Constraints.Max = cgtx.Constraints.Min
		list.Axis = layout.Vertical
		ls := paneListStyle(th, list)
		return SurfaceCard(th, cgtx, func(gtx layout.Context) layout.Dimensions {
			igtx := gtx
			igtx.Constraints.Max.Y = bodyMax
			igtx.Constraints.Min.Y = 0
			return ls.Layout(igtx, 1, func(gtx layout.Context, _ int) layout.Dimensions {
				// material.List measures items with unbounded Max.Y —
				// re-assert the pane budget so children see Pad() math.
				gtx.Constraints.Max.Y = bodyMax
				if gtx.Constraints.Min.Y > bodyMax {
					gtx.Constraints.Min.Y = bodyMax
				}
				return content(gtx)
			})
		})
	}
}

// SplitBoxFillScroll paints card chrome at least maxH tall (Min.Y =
// Max.Y), flush in a column. Short content stays top-aligned. Content
// is NOT wrapped in a list — after Card zeros content Min.Y this
// re-asserts Min.Y = Max.Y (= maxH − 2×Pad) so callers can
// Flexed-expand a body and pin a Rigid footer. Overflow is the
// content's job. list is reserved for API symmetry with the other
// scroll helpers; it is unused here.
func SplitBoxFillScroll(th *Theme, list *widget.List, maxH int, content layout.Widget) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		_ = list
		limit := maxH
		if limit <= 0 {
			limit = gtx.Constraints.Max.Y
		}
		if limit > gtx.Constraints.Max.Y {
			limit = gtx.Constraints.Max.Y
		}
		pad := gtx.Dp(CardProps{}.Pad())
		bodyMax := limit - 2*pad
		if bodyMax < 0 {
			bodyMax = 0
		}
		cgtx := gtx
		cgtx.Constraints.Min = image.Pt(gtx.Constraints.Max.X, limit)
		cgtx.Constraints.Max = cgtx.Constraints.Min
		return SurfaceCard(th, cgtx, func(gtx layout.Context) layout.Dimensions {
			igtx := gtx
			igtx.Constraints.Max.Y = bodyMax
			igtx.Constraints.Min.Y = bodyMax
			return content(igtx)
		})
	}
}
