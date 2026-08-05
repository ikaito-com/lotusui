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

// CardVariant selects the card chrome.
type CardVariant int

const (
	CardOutline  CardVariant = iota // bordered panel, no shadow (default)
	CardElevated                    // panel floating on a soft shadow
	CardSubtle                      // subtle fill, no border, no shadow
)

// CardProps are the card's props. Size scales the content padding —
// same Size enum as every other component; no separate pad knob.
type CardProps struct {
	Variant CardVariant
	Size    Size
}

// Pad is the content inset Card applies for Size (MD → 20dp). Apps that
// budget Max.Y around a fill/scroll child use this instead of hardcoding
// 20 — Card shrinks Max.X by 2×Pad but does not shrink Max.Y.
func (o CardProps) Pad() unit.Dp {
	switch o.Size {
	case Size2XS:
		return 6
	case SizeXS:
		return 10
	case SizeSM:
		return 14
	case SizeLG:
		return 26
	case SizeXL:
		return 32
	case Size2XL:
		return 38
	}
	return 20
}

// CardHeader stacks a title, optional description, and optional action
// (shadcn CardHeader / CardTitle / CardDescription / CardAction as one
// build-time block). Empty strings and nil action are omitted.
func CardHeader(th *Theme, title, description string, action layout.Widget) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		titleW := func(gtx layout.Context) layout.Dimensions {
			if title == "" {
				return layout.Dimensions{}
			}
			return LabelTitle(th, title).Layout(gtx)
		}
		descW := func(gtx layout.Context) layout.Dimensions {
			if description == "" {
				return layout.Dimensions{}
			}
			l := LabelMeta(th, description)
			l.Color = th.Palette.FgMuted
			return l.Layout(gtx)
		}
		if action == nil {
			return VStack(th.Space.XS, titleW, descW)(gtx)
		}
		return layout.Flex{Alignment: layout.Start}.Layout(gtx,
			layout.Flexed(1, VStack(th.Space.XS, titleW, descW)),
			layout.Rigid(HSpacer(th.Space.SM)),
			layout.Rigid(action),
		)
	}
}

// CardFooter is the trailing actions row inside a Card (shadcn CardFooter).
func CardFooter(th *Theme, content layout.Widget) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{Top: th.Space.MD}.Layout(gtx, content)
	}
}

// Card is the grouping surface: a rounded panel around arbitrary
// content — header, body and footer are composition (stack them),
// not slots. It honors gtx.Constraints.Min.Y: a card asked to be at
// least as tall as its row siblings stretches its chrome to match
// (content stays top-aligned), which is what makes equal-height card
// rows possible (see SimpleGrid's measure pass).
//
// Content Min.Y is zeroed before the child layouts; re-assert
// Min.Y = Max.Y on the content if you need fill layout inside (see
// SplitBoxFillScroll). Max.X is reduced by 2× Pad(); Max.Y is not —
// children that fill Max.Y can paint under the pad unless the caller
// (or a Split pane helper) subtracts Pad().
func Card(th *Theme, o CardProps, content layout.Widget) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		pad := gtx.Dp(o.Pad())
		m := op.Record(gtx.Ops)
		inner := gtx
		inner.Constraints.Max.X -= 2 * pad
		if inner.Constraints.Max.X < 0 {
			inner.Constraints.Max.X = 0
		}
		inner.Constraints.Min.X = inner.Constraints.Max.X
		inner.Constraints.Min.Y = 0
		dims := content(inner)
		call := m.Stop()

		cardSize := image.Pt(gtx.Constraints.Max.X, dims.Size.Y+2*pad)
		if cardSize.Y < gtx.Constraints.Min.Y {
			cardSize.Y = gtx.Constraints.Min.Y
		}
		cardSize = gtx.Constraints.Constrain(cardSize)
		r := gtx.Dp(th.Radius.LG)
		// ONE surface grammar: every bordered card carries the SAME
		// border; elevation only deepens the shadow. Outline gets the
		// hairline seat, Elevated the full three-ring shadow — borders
		// never change between them.
		switch o.Variant {
		case CardElevated:
			cardShadow(gtx, cardSize, r)
		case CardOutline:
			seatShadow(gtx, cardSize, r)
		}
		defer clip.UniformRRect(image.Rectangle{Max: cardSize}, r).Push(gtx.Ops).Pop()
		switch o.Variant {
		case CardSubtle:
			paint.Fill(gtx.Ops, th.Palette.BgSubtle)
		default:
			paint.Fill(gtx.Ops, th.Palette.BgPanel)
			widget.Border{Color: th.Palette.Border, Width: unit.Dp(1), CornerRadius: th.Radius.LG}.Layout(gtx,
				func(gtx layout.Context) layout.Dimensions {
					return layout.Dimensions{Size: cardSize}
				})
		}

		st := op.Offset(image.Pt(pad, pad)).Push(gtx.Ops)
		call.Add(gtx.Ops)
		st.Pop()

		return layout.Dimensions{Size: cardSize}
	}
}
