package lotusui

import (
	"image"
	"math"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
)

// Skeleton is the loading placeholder: a rounded block pulsing gently
// between the theme's subtle fills on the shared clock. Give it the
// shape of the content it stands in for.
//
//	lotusui.Skeleton(th, 220, 16)  // a line of text
//	lotusui.Skeleton(th, 40, 40)   // an avatar
//
// It self-invalidates — mount it only while loading.
func Skeleton(th *Theme, w, h unit.Dp) layout.Widget {
	return skeleton(th, w, h, false)
}

// SkeletonCircle is the round form — avatars, icon slots.
func SkeletonCircle(th *Theme, d unit.Dp) layout.Widget {
	return skeleton(th, d, d, true)
}

func skeleton(th *Theme, w, h unit.Dp, round bool) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		gtx.Execute(op.InvalidateCmd{})
		sz := image.Pt(gtx.Dp(w), gtx.Dp(h))
		if w == 0 { // zero width = fill the available width
			sz.X = gtx.Constraints.Max.X
		}
		sz = gtx.Constraints.Constrain(sz)
		if sz.X == 0 || sz.Y == 0 {
			return layout.Dimensions{Size: sz}
		}
		// A slow sine between BgMuted and BgEmphasized: visible motion,
		// never distracting.
		t := float64(gtx.Now.UnixMilli()%1600) / 1600
		mix := (math.Sin(2*math.Pi*t) + 1) / 2
		a, b := th.Palette.BgMuted, th.Palette.BgEmphasized
		lerp := func(x, y uint8) uint8 { return uint8(float64(x) + (float64(y)-float64(x))*mix) }
		col := a
		col.R, col.G, col.B = lerp(a.R, b.R), lerp(a.G, b.G), lerp(a.B, b.B)
		r := gtx.Dp(th.Radius.SM)
		if round {
			r = sz.Y / 2
		}
		r = ClampCorner(r, sz)
		defer clip.UniformRRect(image.Rectangle{Max: sz}, r).Push(gtx.Ops).Pop()
		paint.Fill(gtx.Ops, col)
		return layout.Dimensions{Size: sz}
	}
}
