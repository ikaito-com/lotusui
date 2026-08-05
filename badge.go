package lotusui

import (
	"image"
	"image/color"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// BadgeVariant selects how the badge's scheme paints it.
type BadgeVariant int

const (
	BadgeDefault     BadgeVariant = iota // solid primary fill + contrast text
	BadgeSecondary                       // faint neutral fill + ink
	BadgeDestructive                     // destructive tint + ink
	BadgeOutline                         // border + ink, no fill
	BadgeGhost                           // ink only — no fill, no border
)

// BadgeProps are the badge's props. Scheme defaults to the theme's
// Neutral; Bg/Fg, when both set, override the scheme entirely — the
// escape hatch for status pairs (SuccessBg/Success and friends),
// which are tokens, not schemes.
type BadgeProps struct {
	Variant BadgeVariant
	Size    Size // the shared size presets
	// Color re-colors the badge from a ColorScale, the PASTEL way
	// (SoftScheme: tinted fill + deep same-hue ink) — the badge
	// doctrine holds for any color the caller picks.
	Color  ColorScale
	Scheme *Scheme
	Bg, Fg color.NRGBA
	// Icon names an embedded icon rendered before the text, tinted
	// with the badge's ink.
	Icon string
	// Start/End render arbitrary widgets before/after the text — a
	// Spinner, a dot. The widget owns its own color.
	Start, End layout.Widget
}

// Badge renders a small rounded label — counts, statuses, health.
// The status doctrine holds whatever the colors: tinted pill + deep
// ink, never saturated fill + white text (BadgeDefault exists for
// schemes whose Solid/OnSolid pair honors contrast).
func Badge(th *Theme, text string, o BadgeProps) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		sc := th.Palette.Accent()
		if o.Color != (ColorScale{}) {
			soft := o.Color.SoftScheme()
			sc = soft
			if o.Variant == BadgeDefault {
				// Colored default badges keep the tinted-pill doctrine:
				// pastel fill + deep ink, never .500 + white.
				sc.Solid, sc.OnSolid = soft.Subtle, soft.OnSubtle
			}
		}
		if o.Scheme != nil {
			sc = *o.Scheme
		}
		bg, fg := sc.Solid, sc.OnSolid
		var borderCol color.NRGBA
		switch o.Variant {
		case BadgeSecondary:
			bg, fg = th.Palette.Neutral().Subtle, th.Palette.Neutral().OnSubtle
		case BadgeDestructive:
			d := th.Palette.DangerScheme()
			bg, fg = d.Subtle, d.OnSubtle
		case BadgeOutline:
			bg, fg, borderCol = color.NRGBA{}, sc.OnSubtle, sc.Outline
		case BadgeGhost:
			bg, fg = color.NRGBA{}, th.Palette.FgMuted
		}
		if o.Bg != (color.NRGBA{}) && o.Fg != (color.NRGBA{}) {
			bg, fg, borderCol = o.Bg, o.Fg, color.NRGBA{}
		}

		ratio, hp, vp := float32(12.0/16.0), unit.Dp(8), unit.Dp(3)
		switch o.Size {
		case Size2XS:
			ratio, hp, vp = 9.0/16.0, 5, 1
		case SizeXS:
			ratio, hp, vp = 10.0/16.0, 6, 2
		case SizeSM:
			ratio, hp, vp = 11.0/16.0, 7, 2
		case SizeLG:
			ratio, hp, vp = 13.0/16.0, 10, 4
		case SizeXL:
			ratio, hp, vp = 14.0/16.0, 12, 5
		case Size2XL:
			ratio, hp, vp = 15.0/16.0, 14, 6
		}
		lbl := material.Label(th.Material, Sp(th, ratio), text)
		lbl.Color = fg
		lbl.Font.Weight = font.Medium

		// The badge hugs its text — never inherit a stretched Min.
		lgtx := gtx
		lgtx.Constraints.Min = image.Point{}
		m := op.Record(gtx.Ops)
		var dims layout.Dimensions
		if o.Icon != "" || o.Start != nil || o.End != nil {
			iconSz := unit.Dp(float32(gtx.Metric.SpToDp(Sp(th, ratio))) + 1)
			var row []layout.FlexChild
			if o.Start != nil {
				row = append(row, layout.Rigid(o.Start), layout.Rigid(HSpacer(4)))
			}
			if o.Icon != "" {
				row = append(row, layout.Rigid(SVGIcon(o.Icon, iconSz, fg)), layout.Rigid(HSpacer(4)))
			}
			row = append(row, layout.Rigid(lbl.Layout))
			if o.End != nil {
				row = append(row, layout.Rigid(HSpacer(4)), layout.Rigid(o.End))
			}
			dims = layout.Flex{Alignment: layout.Middle}.Layout(lgtx, row...)
		} else {
			dims = lbl.Layout(lgtx)
		}
		call := m.Stop()

		hPad := gtx.Dp(hp)
		vPad := gtx.Dp(vp)
		sz := gtx.Constraints.Constrain(image.Pt(dims.Size.X+2*hPad, dims.Size.Y+2*vPad))
		// A rounded RECT, not a pill — the badge reads as a compact
		// label, not a capsule.
		r := ClampCorner(gtx.Dp(6), sz)
		defer clip.UniformRRect(image.Rectangle{Max: sz}, r).Push(gtx.Ops).Pop()
		if bg != (color.NRGBA{}) {
			paint.Fill(gtx.Ops, bg)
		}
		if borderCol != (color.NRGBA{}) {
			widget.Border{Color: borderCol, Width: unit.Dp(1), CornerRadius: unit.Dp(6)}.Layout(gtx,
				func(gtx layout.Context) layout.Dimensions {
					return layout.Dimensions{Size: sz}
				})
		}
		st := op.Offset(image.Pt(hPad, vPad)).Push(gtx.Ops)
		call.Add(gtx.Ops)
		st.Pop()
		return layout.Dimensions{Size: sz}
	}
}
