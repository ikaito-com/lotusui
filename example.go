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

// Example is Preview|Code chrome: one bordered rounded card whose tab
// strip is the top of the card (surface2), not a free-floating Tabs well.
// A lotusui extension for docs and in-app recipe surfaces — pair with
// CodeBlock{Nested: true} for the Code panel.
//
// State must outlive the frame (which tab is showing).
type Example struct {
	preview, code widget.Clickable
	showCode      bool
}

// ExampleProps are one Layout's slots. Preview is required; Code nil
// means Preview-only (no Code tab).
type ExampleProps struct {
	Preview layout.Widget
	Code    layout.Widget
}

// Layout paints the Example card.
func (e *Example) Layout(th *Theme, gtx layout.Context, o ExampleProps) layout.Dimensions {
	maxX := gtx.Constraints.Max.X
	if maxX < 1 {
		return layout.Dimensions{}
	}
	hasCode := o.Code != nil
	if e.preview.Clicked(gtx) {
		e.showCode = false
	}
	if hasCode && e.code.Clicked(gtx) {
		e.showCode = true
	}
	if !hasCode {
		e.showCode = false
	}

	r := gtx.Dp(th.Radius.MD)
	bw := gtx.Dp(unit.Dp(1))
	tabH := gtx.Dp(unit.Dp(40))

	body := o.Preview
	if e.showCode && o.Code != nil {
		body = o.Code
	}
	if body == nil {
		body = func(gtx layout.Context) layout.Dimensions { return layout.Dimensions{} }
	}

	innerW := maxX - 2*bw
	if innerW < 1 {
		innerW = 1
	}
	bodyH := measureExampleBody(gtx, innerW, body)
	if bodyH < 1 {
		bodyH = 1
	}
	totalH := tabH + bodyH
	sz := image.Pt(maxX, totalH)

	cl := clip.UniformRRect(image.Rectangle{Max: sz}, r).Push(gtx.Ops)

	paint.Fill(gtx.Ops, th.Palette.BgPanel)

	strip := clip.Rect{Max: image.Pt(maxX, tabH)}.Push(gtx.Ops)
	paint.Fill(gtx.Ops, th.Palette.BgSubtle)
	strip.Pop()
	hair := clip.Rect{Min: image.Pt(0, tabH-bw), Max: image.Pt(maxX, tabH)}.Push(gtx.Ops)
	paint.Fill(gtx.Ops, th.Palette.BorderSubtle)
	hair.Pop()

	layout.Inset{
		Top: unit.Dp(6), Bottom: unit.Dp(6),
		Left: unit.Dp(8), Right: unit.Dp(8),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(e.tab(th, &e.preview, "Preview", !e.showCode, hasCode)),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if !hasCode {
					return layout.Dimensions{}
				}
				return layout.Spacer{Width: unit.Dp(4)}.Layout(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if !hasCode {
					return layout.Dimensions{}
				}
				return e.tab(th, &e.code, "Code", e.showCode, true)(gtx)
			}),
		)
	})

	trans := op.Offset(image.Pt(0, tabH)).Push(gtx.Ops)
	cgtx := gtx
	cgtx.Constraints.Min = image.Point{}
	cgtx.Constraints.Max.X = maxX
	cgtx.Constraints.Max.Y = bodyH
	body(cgtx)
	trans.Pop()

	cl.Pop()

	// Border last so the strip never paints over the card outline.
	widget.Border{
		Color: th.Palette.Border, Width: unit.Dp(1), CornerRadius: th.Radius.MD,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: sz}
	})

	return layout.Dimensions{Size: sz}
}

func measureExampleBody(gtx layout.Context, width int, w layout.Widget) int {
	mgtx := gtx
	mgtx.Constraints.Min = image.Point{}
	mgtx.Constraints.Max.X = width
	mgtx.Constraints.Max.Y = 1 << 20
	return measureContent(mgtx, w).Size.Y
}

func (e *Example) tab(th *Theme, btn *widget.Clickable, label string, active, clickable bool) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		inner := func(gtx layout.Context) layout.Dimensions {
			if clickable {
				pointer.CursorPointer.Add(gtx.Ops)
			}
			l := material.Label(th.Material, unit.Sp(12.5), label)
			l.Font.Weight = 600
			if active {
				l.Color = th.Palette.Fg
			} else if clickable && btn.Hovered() {
				l.Color = th.Palette.Fg
			} else {
				l.Color = th.Palette.FgSubtle
			}
			m := op.Record(gtx.Ops)
			d := l.Layout(gtx)
			call := m.Stop()
			padX, padY := gtx.Dp(10), gtx.Dp(7)
			sz := image.Pt(d.Size.X+2*padX, d.Size.Y+2*padY)
			if active {
				rr := gtx.Dp(unit.Dp(7))
				cl := clip.UniformRRect(image.Rectangle{Max: sz}, rr).Push(gtx.Ops)
				paint.Fill(gtx.Ops, th.Palette.BgPanel)
				cl.Pop()
				widget.Border{
					Color: th.Palette.Border, Width: unit.Dp(1), CornerRadius: unit.Dp(7),
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Dimensions{Size: sz}
				})
			} else if clickable && btn.Hovered() {
				rr := gtx.Dp(unit.Dp(7))
				cl := clip.UniformRRect(image.Rectangle{Max: sz}, rr).Push(gtx.Ops)
				paint.Fill(gtx.Ops, th.Palette.BgPanel)
				cl.Pop()
			}
			op.Offset(image.Pt(padX, padY)).Add(gtx.Ops)
			call.Add(gtx.Ops)
			return layout.Dimensions{Size: sz}
		}
		if !clickable {
			return inner(gtx)
		}
		return btn.Layout(gtx, inner)
	}
}
