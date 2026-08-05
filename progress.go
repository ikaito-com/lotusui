package lotusui

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

// Progress renders a determinate progress bar: a track in the subtle
// fill, the filled portion in the brand solid. value is clamped to
// [0, 1]; pass a NEGATIVE value for the indeterminate state — a
// segment sweeping the track on the shared clock (self-invalidating;
// mount it only while something is actually in flight).
func Progress(th *Theme, value float32) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		w := gtx.Constraints.Max.X
		h := gtx.Dp(8)
		if h > gtx.Constraints.Max.Y {
			h = gtx.Constraints.Max.Y
		}
		if h == 0 || w == 0 {
			return layout.Dimensions{Size: gtx.Constraints.Constrain(image.Pt(w, h))}
		}
		r := h / 2
		defer clip.UniformRRect(image.Rectangle{Max: image.Pt(w, h)}, r).Push(gtx.Ops).Pop()
		paint.Fill(gtx.Ops, th.Palette.BgMuted)

		if value < 0 {
			// Indeterminate: a third-width segment sweeping left→right.
			gtx.Execute(op.InvalidateCmd{})
			t := float64(gtx.Now.UnixMilli()%1200) / 1200
			seg := w / 3
			x := int(float64(w+seg)*t) - seg
			fill := clip.UniformRRect(image.Rect(x, 0, x+seg, h), ClampCorner(r, image.Pt(seg, h)))
			paint.FillShape(gtx.Ops, th.Palette.BrandSolid, fill.Op(gtx.Ops))
			return layout.Dimensions{Size: image.Pt(w, h)}
		}
		if value > 1 {
			value = 1
		}
		fw := int(float32(w) * value)
		if fw > 0 {
			fill := clip.UniformRRect(image.Rect(0, 0, fw, h), ClampCorner(r, image.Pt(fw, h)))
			paint.FillShape(gtx.Ops, th.Palette.BrandSolid, fill.Op(gtx.Ops))
		}
		return layout.Dimensions{Size: image.Pt(w, h)}
	}
}
