package lotusui

import (
	"image"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

// AvatarProps are the avatar's options. Initials render centered in
// the circle; empty Initials fall back to the person glyph. Color
// re-tints the circle from a scale the pastel way (SoftScheme).
type AvatarProps struct {
	Initials string
	Size     Size
	Color    ColorScale
	Scheme   *Scheme
	// Badge renders a small status dot on the bottom-right rim,
	// ringed in the panel background.
	Badge *AvatarBadge
}

// AvatarBadge is the avatar's status dot. Color anchors its fill
// (zero = the theme accent); Icon renders inside, inverted.
type AvatarBadge struct {
	Color ColorScale
	Icon  string
}

func avatarDp(sz Size) int {
	switch sz {
	case Size2XS:
		return 20
	case SizeXS:
		return 24
	case SizeSM:
		return 28
	case SizeLG:
		return 40
	case SizeXL:
		return 48
	case Size2XL:
		return 56
	}
	return 32
}

// Avatar renders the circular identity mark: initials (or the person
// glyph) on a tinted circle, plus the status Badge when set. For the
// overlapping-group pattern use AvatarGroup.
func Avatar(th *Theme, o AvatarProps) layout.Widget {
	if o.Badge != nil {
		return avatarWithBadge(th, o)
	}
	return avatarCircle(th, o)
}

func avatarCircle(th *Theme, o AvatarProps) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		sc := th.Palette.Accent()
		if o.Color != (ColorScale{}) {
			sc = o.Color.SoftScheme()
		}
		if o.Scheme != nil {
			sc = *o.Scheme
		}
		d := gtx.Dp(unit.Dp(avatarDp(o.Size)))
		// Hostile constraints: the circle never overflows its box.
		if m := gtx.Constraints.Max; d > m.X || d > m.Y {
			if m.X < m.Y {
				d = m.X
			} else {
				d = m.Y
			}
		}
		sz := image.Pt(d, d)
		if d == 0 {
			return layout.Dimensions{}
		}
		defer clip.UniformRRect(image.Rectangle{Max: sz}, d/2).Push(gtx.Ops).Pop()
		paint.Fill(gtx.Ops, sc.Subtle)
		center := func(w layout.Widget) {
			cgtx := gtx
			cgtx.Constraints = layout.Constraints{Min: sz, Max: sz}
			layout.Center.Layout(cgtx, w)
		}
		if o.Initials == "" {
			center(SVGIcon(IconPerson, unit.Dp(avatarDp(o.Size)*5/9), sc.OnSubtle))
			return layout.Dimensions{Size: sz}
		}
		center(func(gtx layout.Context) layout.Dimensions {
			l := material.Label(th.Material, Sp(th, float32(avatarDp(o.Size))/40.0), o.Initials)
			l.Color = sc.OnSubtle
			l.Font.Weight = font.Medium
			l.MaxLines = 1
			return l.Layout(gtx)
		})
		return layout.Dimensions{Size: sz}
	}
}

// avatarWithBadge draws the circle, then the badge OVER the rim — a
// BgPanel ring under a filled dot, bottom-right.
func avatarWithBadge(th *Theme, o AvatarProps) layout.Widget {
	base := avatarCircle(th, o)
	return func(gtx layout.Context) layout.Dimensions {
		dims := base(gtx)
		d := dims.Size.X
		bd := d * 5 / 16 // badge diameter
		if bd < gtx.Dp(8) {
			bd = gtx.Dp(8)
		}
		fill := th.Palette.Accent().Solid
		if o.Badge.Color != (ColorScale{}) {
			fill = o.Badge.Color.C500
		}
		// Bottom-right, ringed: a BgPanel disc 2dp larger underneath.
		ring := gtx.Dp(2)
		at := image.Pt(d-bd, d-bd)
		off := op.Offset(at.Sub(image.Pt(ring, ring))).Push(gtx.Ops)
		rd := bd + 2*ring
		func() {
			defer clip.UniformRRect(image.Rectangle{Max: image.Pt(rd, rd)}, rd/2).Push(gtx.Ops).Pop()
			paint.Fill(gtx.Ops, th.Palette.BgPanel)
		}()
		off.Pop()
		off = op.Offset(at).Push(gtx.Ops)
		func() {
			defer clip.UniformRRect(image.Rectangle{Max: image.Pt(bd, bd)}, bd/2).Push(gtx.Ops).Pop()
			paint.Fill(gtx.Ops, fill)
			if o.Badge.Icon != "" {
				cgtx := gtx
				cgtx.Constraints = layout.Constraints{Min: image.Pt(bd, bd), Max: image.Pt(bd, bd)}
				layout.Center.Layout(cgtx, SVGIcon(o.Badge.Icon, unit.Dp(float32(bd)*3/4/cgtx.Metric.PxPerDp), th.Palette.FgInverted))
			}
		}()
		off.Pop()
		return dims
	}
}

// AvatarGroupProps are the overlapping-group's options.
type AvatarGroupProps struct {
	Size Size
	// Count renders a trailing "+N" bubble; CountIcon an icon bubble.
	Count     string
	CountIcon string
}

// AvatarGroup overlaps avatars right-over-left, each ringed in the
// panel background, with an optional trailing count bubble.
func AvatarGroup(th *Theme, o AvatarGroupProps, avatars ...AvatarProps) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		d := gtx.Dp(unit.Dp(avatarDp(o.Size)))
		overlap := d / 4
		ring := gtx.Dp(2)
		n := len(avatars)
		extra := 0
		if o.Count != "" || o.CountIcon != "" {
			extra = 1
		}
		total := image.Pt(d+(n+extra-1)*(d-overlap), d)
		if n+extra == 0 {
			return layout.Dimensions{}
		}
		x := 0
		cell := func(w layout.Widget) {
			off := op.Offset(image.Pt(x-ring, -ring)).Push(gtx.Ops)
			rd := d + 2*ring
			func() {
				defer clip.UniformRRect(image.Rectangle{Max: image.Pt(rd, rd)}, rd/2).Push(gtx.Ops).Pop()
				if x > 0 {
					paint.Fill(gtx.Ops, th.Palette.BgPanel)
				}
			}()
			off.Pop()
			off = op.Offset(image.Pt(x, 0)).Push(gtx.Ops)
			cgtx := gtx
			cgtx.Constraints = layout.Constraints{Min: image.Pt(d, d), Max: image.Pt(d, d)}
			w(cgtx)
			off.Pop()
			x += d - overlap
		}
		for _, a := range avatars {
			a := a
			a.Size = o.Size
			cell(Avatar(th, a))
		}
		if extra == 1 {
			cell(func(gtx layout.Context) layout.Dimensions {
				sz := image.Pt(d, d)
				defer clip.UniformRRect(image.Rectangle{Max: sz}, d/2).Push(gtx.Ops).Pop()
				paint.Fill(gtx.Ops, th.Palette.BgSubtle)
				cgtx := gtx
				cgtx.Constraints = layout.Constraints{Min: sz, Max: sz}
				if o.CountIcon != "" {
					layout.Center.Layout(cgtx, SVGIcon(o.CountIcon, unit.Dp(avatarDp(o.Size)*5/9), th.Palette.FgMuted))
				} else {
					layout.Center.Layout(cgtx, func(gtx layout.Context) layout.Dimensions {
						l := material.Label(th.Material, Sp(th, float32(avatarDp(o.Size))/40.0), o.Count)
						l.Color = th.Palette.FgMuted
						l.Font.Weight = font.Medium
						l.MaxLines = 1
						return l.Layout(gtx)
					})
				}
				return layout.Dimensions{Size: sz}
			})
		}
		return layout.Dimensions{Size: total}
	}
}
