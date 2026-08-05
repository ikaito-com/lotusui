package lotusui

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// Kbd renders a keyboard-key hint — a small bordered cap in muted
// ink: the "⌘K" beside a search field. Display only.
func Kbd(th *Theme, key string) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		lgtx := gtx
		lgtx.Constraints.Min = image.Point{}
		m := op.Record(gtx.Ops)
		dims := layout.Inset{Top: 1, Bottom: 1, Left: 5, Right: 5}.Layout(lgtx, func(gtx layout.Context) layout.Dimensions {
			l := material.Label(th.Material, Sp(th, 11.0/16.0), key)
			l.Color = th.Palette.FgSubtle
			l.MaxLines = 1
			return l.Layout(gtx)
		})
		call := m.Stop()
		r := ClampCorner(gtx.Dp(th.Radius.SM), dims.Size)
		defer clip.UniformRRect(image.Rectangle{Max: dims.Size}, r).Push(gtx.Ops).Pop()
		paint.Fill(gtx.Ops, th.Palette.BgSubtle)
		widget.Border{Color: th.Palette.Border, Width: unit.Dp(1), CornerRadius: th.Radius.SM}.Layout(gtx,
			func(gtx layout.Context) layout.Dimensions { return layout.Dimensions{Size: dims.Size} })
		call.Add(gtx.Ops)
		return dims
	}
}
