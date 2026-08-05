package lotusui

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
)

// Popover is the anchored floating panel for arbitrary content — the
// general form of Select's dropdown, built on the same portal
// primitive. The caller owns Open (toggle it from its trigger's
// click); pressing anywhere else or Escape closes it.
//
//	if trigger.Clicked(gtx) { pop.Open = !pop.Open }
//	dims := lotusui.Button(th, &trigger, "Open", lotusui.ButtonProps{})(gtx)
//	pop.Layout(th, gtx, dims.Size, panelContent)
//
// PopoverAlign positions the panel's horizontal edge relative to the
// anchor. The zero value is Center — matching shadcn Popover /
// HoverCardContent (`align="center"`). Start shares the leading edge;
// End shares the trailing edge.
type PopoverAlign int

const (
	PopoverCenter PopoverAlign = iota
	PopoverStart
	PopoverEnd
)

type Popover struct {
	Open bool
	// Width is the maximum panel width. Zero matches the anchor width
	// (dropdown-under-button). Non-zero: the panel hugs content up to
	// Width — same doctrine as HoverCard.
	Width unit.Dp
	// Align positions the panel against the anchor edge.
	Align PopoverAlign

	sites   layoutSites
	dismiss dismisser
}

// Layout renders the floating panel 4dp below an anchor of the given
// size, when Open. Call it immediately after laying out the anchor,
// at the anchor's position. If Layout runs more than once per frame
// while Open, only the first call paints — one Popover, one panel.
func (p *Popover) Layout(th *Theme, gtx layout.Context, anchor image.Point, content layout.Widget) {
	idx := p.sites.next(gtx.Now)
	if !p.Open {
		return
	}
	if idx == 0 && p.dismiss.Dismissed(gtx) {
		p.Open = false
		return
	}
	if idx != 0 {
		return
	}
	Floating(gtx, func(gtx layout.Context) layout.Dimensions {
		p.dismiss.Add(gtx)

		maxW := anchor.X
		minW := anchor.X // match trigger when Width is zero
		if p.Width != 0 {
			maxW = gtx.Dp(p.Width)
			minW = 0 // hug up to Width
		}
		gtx.Constraints = layout.Constraints{
			Min: image.Pt(minW, 0),
			Max: image.Pt(maxW, gtx.Dp(480)),
		}
		m := op.Record(gtx.Ops)
		dims := layout.UniformInset(th.Space.MD).Layout(gtx, content)
		call := m.Stop()

		size := dims.Size
		if p.Width == 0 {
			size.X = maxW
		}
		x := (anchor.X - size.X) / 2 // PopoverCenter (zero)
		switch p.Align {
		case PopoverStart:
			x = 0
		case PopoverEnd:
			x = anchor.X - size.X
		}
		defer op.Offset(image.Pt(x, anchor.Y+gtx.Dp(4))).Push(gtx.Ops).Pop()

		r := gtx.Dp(th.Radius.MD)
		cardShadow(gtx, size, r)
		defer clip.UniformRRect(image.Rectangle{Max: size}, r).Push(gtx.Ops).Pop()
		paint.Fill(gtx.Ops, th.Palette.BgPanel)
		widget.Border{Color: th.Palette.Border, Width: unit.Dp(1), CornerRadius: th.Radius.MD}.Layout(gtx,
			func(gtx layout.Context) layout.Dimensions { return layout.Dimensions{Size: size} })
		call.Add(gtx.Ops)
		return layout.Dimensions{Size: size}
	})
}
