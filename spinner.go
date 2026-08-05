package lotusui

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

// Spinner is the standalone loading indicator: the same three-quarter
// arc Button's loading state draws, rotating on the frame clock, in
// the theme's muted ink (pass a Scheme'd color via SpinnerTint for
// anything else). It self-invalidates — mount it only while loading.
func Spinner(th *Theme, size unit.Dp) layout.Widget {
	return SpinnerTint(th, size, th.Palette.FgSubtle)
}

// SpinnerTint is Spinner in an explicit color.
func SpinnerTint(th *Theme, size unit.Dp, col color.NRGBA) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		gtx.Execute(op.InvalidateCmd{})
		px := gtx.Dp(size)
		if m := gtx.Constraints.Max; px > m.X || px > m.Y {
			if m.X < m.Y {
				px = m.X
			} else {
				px = m.Y
			}
		}
		if px > 0 {
			spinner(gtx, col, px)
		}
		return layout.Dimensions{Size: image.Pt(px, px)}
	}
}
