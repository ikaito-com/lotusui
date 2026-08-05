package lotusui

import (
	"image"

	"gioui.org/font"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
)

// Item is shadcn's list-row primitive: a flex container for media,
// title, description and actions. Compose slots on ItemProps with
// ItemMedia / ItemContent / ItemTitle / ItemDescription / ItemActions
// (and optional Header/Footer). Group rows with ItemGroup /
// ItemSeparator.
//
// Build-time composition — slots are layout.Widget values assembled
// once, not React children. Pair with SelectOption.Content for
// multiline plan cards.

// ItemVariant selects the row chrome.
type ItemVariant int

const (
	ItemDefault ItemVariant = iota // transparent; hover when clickable
	ItemOutline                    // bordered panel
	ItemMuted                      // BgSubtle fill
)

// ItemMediaVariant selects the leading media well.
type ItemMediaVariant int

const (
	ItemMediaDefault ItemMediaVariant = iota // bare content (avatar, custom)
	ItemMediaIcon                            // muted square well around an icon
	ItemMediaImage                           // clipped rounded image slot
)

// ItemProps are the row's options and slots. Size uses the shared
// enum; MD is shadcn's default, SM → sm, XS → xs.
type ItemProps struct {
	Variant ItemVariant
	Size    Size
	// Btn, when set, makes the whole row a clickable HoverRow-style
	// target (shadcn link/render). Poll Clicked on your side.
	Btn *widget.Clickable

	Header  layout.Widget // full-width above the main row
	Media   layout.Widget
	Content layout.Widget
	Actions layout.Widget
	Footer  layout.Widget // full-width below the main row
}

func itemPad(sz Size) (v, h unit.Dp) {
	switch sz {
	case Size2XS, SizeXS:
		return 6, 8
	case SizeSM:
		return 8, 10
	case SizeLG, SizeXL, Size2XL:
		return 14, 16
	}
	return 10, 12
}

// Item lays out a composed row from ItemProps slots.
func Item(th *Theme, o ItemProps) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		vPad, hPad := itemPad(o.Size)
		body := func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
			var stack []layout.Widget
			if o.Header != nil {
				stack = append(stack, o.Header)
			}
			stack = append(stack, func(gtx layout.Context) layout.Dimensions {
				var children []layout.FlexChild
				if o.Media != nil {
					children = append(children, layout.Rigid(o.Media))
					children = append(children, layout.Rigid(HSpacer(th.Space.SM)))
				}
				if o.Content != nil {
					children = append(children, layout.Flexed(1, o.Content))
				} else {
					children = append(children, layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return layout.Dimensions{Size: image.Pt(0, 0)}
					}))
				}
				if o.Actions != nil {
					children = append(children, layout.Rigid(HSpacer(th.Space.SM)))
					children = append(children, layout.Rigid(o.Actions))
				}
				return layout.Flex{Alignment: layout.Middle}.Layout(gtx, children...)
			})
			if o.Footer != nil {
				stack = append(stack, o.Footer)
			}
			if len(stack) == 1 {
				return stack[0](gtx)
			}
			return VStack(th.Space.SM, stack...)(gtx)
		}
		inset := layout.Inset{Top: vPad, Bottom: vPad, Left: hPad, Right: hPad}
		paintRow := func(gtx layout.Context) layout.Dimensions {
			m := op.Record(gtx.Ops)
			dims := inset.Layout(gtx, body)
			call := m.Stop()
			r := ClampCorner(gtx.Dp(th.Radius.MD), dims.Size)
			defer clip.UniformRRect(image.Rectangle{Max: dims.Size}, r).Push(gtx.Ops).Pop()
			switch o.Variant {
			case ItemMuted:
				paint.Fill(gtx.Ops, th.Palette.BgSubtle)
			case ItemOutline:
				paint.Fill(gtx.Ops, th.Palette.BgPanel)
				widget.Border{
					Color:        th.Palette.Border,
					Width:        unit.Dp(1),
					CornerRadius: th.Radius.MD,
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Dimensions{Size: dims.Size}
				})
			}
			if o.Btn != nil && o.Btn.Hovered() && o.Variant == ItemDefault {
				paint.Fill(gtx.Ops, th.Palette.BgSubtle)
			}
			call.Add(gtx.Ops)
			return dims
		}
		if o.Btn == nil {
			return paintRow(gtx)
		}
		return o.Btn.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			pointer.CursorPointer.Add(gtx.Ops)
			return paintRow(gtx)
		})
	}
}

// ItemGroup stacks items (and separators) full-width.
func ItemGroup(th *Theme, items ...layout.Widget) layout.Widget {
	return VStack(0, items...)
}

// ItemSeparator is a hairline between items in a group.
func ItemSeparator(th *Theme) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{Top: th.Space.XS, Bottom: th.Space.XS}.Layout(gtx, Hairline(th))
	}
}

// ItemMedia is the leading slot. Icon variant wraps content in a muted
// square; Image clips to a rounded rect; Default is bare.
func ItemMedia(th *Theme, variant ItemMediaVariant, content layout.Widget) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		switch variant {
		case ItemMediaIcon:
			sz := gtx.Dp(36)
			gtx.Constraints = layout.Exact(image.Pt(sz, sz))
			m := op.Record(gtx.Ops)
			layout.Center.Layout(gtx, content)
			call := m.Stop()
			r := gtx.Dp(th.Radius.SM)
			defer clip.UniformRRect(image.Rectangle{Max: image.Pt(sz, sz)}, r).Push(gtx.Ops).Pop()
			paint.Fill(gtx.Ops, th.Palette.BgMuted)
			call.Add(gtx.Ops)
			return layout.Dimensions{Size: image.Pt(sz, sz)}
		case ItemMediaImage:
			sz := gtx.Dp(40)
			gtx.Constraints = layout.Exact(image.Pt(sz, sz))
			defer clip.UniformRRect(image.Rectangle{Max: image.Pt(sz, sz)}, gtx.Dp(th.Radius.SM)).Push(gtx.Ops).Pop()
			return content(gtx)
		}
		return content(gtx)
	}
}

// ItemContent stacks title/description (and anything else) in a column.
func ItemContent(th *Theme, parts ...layout.Widget) layout.Widget {
	return VStack(unit.Dp(2), parts...)
}

// ItemTitle is the primary line — one line, medium weight.
func ItemTitle(th *Theme, text string) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		l := LabelBody(th, text)
		l.Font.Weight = font.Medium
		l.MaxLines = 1
		return l.Layout(gtx)
	}
}

// ItemDescription is the secondary line — muted, up to two lines.
func ItemDescription(th *Theme, text string) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		l := LabelCaption(th, text)
		l.Color = th.Palette.FgMuted
		l.MaxLines = 2
		return l.Layout(gtx)
	}
}

// ItemActions is the trailing slot (buttons, chevron, …).
func ItemActions(th *Theme, parts ...layout.Widget) layout.Widget {
	return HStack(th.Space.XS, parts...)
}

// ItemHeader spans the full row above the main columns (title + action).
func ItemHeader(th *Theme, left, right layout.Widget) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		if right == nil {
			return left(gtx)
		}
		return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
			layout.Flexed(1, left),
			layout.Rigid(right),
		)
	}
}

// ItemFooter spans the full row below the main columns.
func ItemFooter(th *Theme, left, right layout.Widget) layout.Widget {
	return ItemHeader(th, left, right)
}
