package lotusui

import (
	"image"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
)

// AlertVariant selects the alert's role.
type AlertVariant int

const (
	AlertDefault     AlertVariant = iota // neutral callout
	AlertDestructive                     // danger ink on a danger tint
)

// AlertProps are the alert's options. Icon defaults by variant
// (info / warning); empty Description renders title-only.
type AlertProps struct {
	Variant     AlertVariant
	Icon        string // override the variant's default icon
	Title       string
	Description string
	// Action renders at the alert's end — typically a small Button.
	Action layout.Widget
	// Color re-tints the whole alert from a scale, the pastel way
	// (tinted well + deep same-hue ink) — the amber "subscription
	// expiring" callout, in any hue, without breaking readability.
	Color ColorScale
}

// Alert is the static callout: an icon, a title, and an optional
// description in a bordered, tinted box. It informs — for actions use
// Dialog or AlertDialog.
func Alert(th *Theme, o AlertProps) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		icon := o.Icon
		bg, border, ink, body := th.Palette.BgPanel, th.Palette.Border, th.Palette.Fg, th.Palette.FgMuted
		if o.Color != (ColorScale{}) {
			soft := o.Color.SoftScheme()
			bg, border, ink, body = soft.Subtle, soft.Outline, soft.OnSubtle, soft.OnSubtle
			if icon == "" {
				icon = IconWarning
			}
		}
		if o.Variant == AlertDestructive {
			bg, border, ink, body = th.Palette.DangerBg, th.Palette.Danger, th.Palette.Danger, th.Palette.Danger
			if icon == "" {
				icon = IconWarning
			}
		}
		if icon == "" {
			icon = IconInfo
		}
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
		m := op.Record(gtx.Ops)
		dims := layout.UniformInset(th.Space.MD).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{}.Layout(gtx,
				layout.Rigid(SVGIcon(icon, 18, ink)),
				layout.Rigid(HSpacer(th.Space.SM)),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					title := LabelBody(th, o.Title)
					title.Color = ink
					title.Font.Weight = font.Medium
					if o.Description == "" {
						return title.Layout(gtx)
					}
					desc := LabelCaption(th, o.Description)
					desc.Color = body
					return VStack(4, title.Layout, desc.Layout)(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if o.Action == nil {
						return layout.Dimensions{}
					}
					return layout.Inset{Left: th.Space.SM}.Layout(gtx, o.Action)
				}),
			)
		})
		call := m.Stop()
		dims.Size.X = gtx.Constraints.Max.X
		r := gtx.Dp(th.Radius.MD)
		defer clip.UniformRRect(image.Rectangle{Max: dims.Size}, r).Push(gtx.Ops).Pop()
		paint.Fill(gtx.Ops, bg)
		widget.Border{Color: border, Width: unit.Dp(1), CornerRadius: th.Radius.MD}.Layout(gtx,
			func(gtx layout.Context) layout.Dimensions { return layout.Dimensions{Size: dims.Size} })
		call.Add(gtx.Ops)
		return dims
	}
}
