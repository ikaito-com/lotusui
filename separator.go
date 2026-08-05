package lotusui

import (
	"gioui.org/layout"
)

// Separator is the semantic divider: a 1dp hairline in the theme's
// subtle border ink, horizontal by default. It is the component form
// of Hairline/VerticalHairline — use it wherever content needs a
// visual boundary that is not a box.
func Separator(th *Theme) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
		return Hairline(th)(gtx)
	}
}

// SeparatorVertical divides horizontal siblings — a 1dp vertical rule
// spanning the row's height.
func SeparatorVertical(th *Theme) layout.Widget {
	return VerticalHairline(th)
}
