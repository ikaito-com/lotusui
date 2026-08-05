package lotusui

import (
	"image"

	"gioui.org/font"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/widget"
)

// Accordion is the vertically stacked disclosure: titles that expand
// one panel of content at a time (set Multiple for independent
// panels). Expansion animates on the shared clock — content is
// measured at natural height and revealed by a growing clip, never
// reflowed mid-flight.
type Accordion struct {
	// Bordered wraps the whole accordion in a rounded border with
	// side padding — the boxed look.
	Bordered bool
	// Open indexes the expanded item in single mode; -1 = all closed.
	Open int
	// Multiple switches to independent panels tracked in Expanded.
	Multiple bool
	Expanded []bool
	btns     []widget.Clickable
	anims    []slideAnim
}

// AccordionItem is one disclosure: a title and its content.
type AccordionItem struct {
	Title   string
	Content layout.Widget
	// Disabled dims the title and ignores clicks.
	Disabled bool
}

// AccordionItemOf builds one item (build-time composition).
func AccordionItemOf(title string, content layout.Widget) AccordionItem {
	return AccordionItem{Title: title, Content: content}
}

// AccordionItems packs variadic items into a slice.
func AccordionItems(items ...AccordionItem) []AccordionItem {
	out := make([]AccordionItem, len(items))
	copy(out, items)
	return out
}

func (a *Accordion) Layout(th *Theme, gtx layout.Context, items ...AccordionItem) layout.Dimensions {
	if len(a.btns) != len(items) {
		a.btns = make([]widget.Clickable, len(items))
		a.anims = make([]slideAnim, len(items))
		a.Expanded = make([]bool, len(items))
		if !a.Multiple && a.Open == 0 && len(items) > 0 {
			// zero value opens nothing, matching the collapsed start
			a.Open = -1
		}
	}
	for i := range a.btns {
		if a.btns[i].Clicked(gtx) && !items[i].Disabled {
			if a.Multiple {
				a.Expanded[i] = !a.Expanded[i]
			} else if a.Open == i {
				a.Open = -1
			} else {
				a.Open = i
			}
		}
	}
	open := func(i int) bool {
		if a.Multiple {
			return a.Expanded[i]
		}
		return a.Open == i
	}

	var rows []layout.Widget
	for i, it := range items {
		i, it := i, it
		rows = append(rows, func(gtx layout.Context) layout.Dimensions {
			return a.btns[i].Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.X = gtx.Constraints.Max.X
				pointer.CursorPointer.Add(gtx.Ops)
				return layout.Inset{Top: 12, Bottom: 12}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					chev := IconChevronRight
					if open(i) {
						chev = IconChevronDown
					}
					ink, chevInk := th.Palette.Fg, th.Palette.FgSubtle
					if it.Disabled {
						ink, chevInk = th.Palette.FgDisabled, th.Palette.FgDisabled
					}
					return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							l := LabelBody(th, it.Title)
							l.Color = ink
							l.Font.Weight = font.Medium
							return l.Layout(gtx)
						}),
						layout.Rigid(SVGIcon(chev, 16, chevInk)),
					)
				})
			})
		})
		rows = append(rows, func(gtx layout.Context) layout.Dimensions {
			target := float32(0)
			if open(i) {
				target = 1
			}
			prog := a.anims[i].advance(gtx, target, th.Duration.Normal)
			if prog == 0 {
				return layout.Dimensions{Size: image.Pt(gtx.Constraints.Max.X, 0)}
			}
			// Measure the content at natural height, reveal prog of it.
			m := op.Record(gtx.Ops)
			cgtx := gtx
			cgtx.Constraints.Min = image.Point{}
			cgtx.Constraints.Max.Y = 1 << 20
			dims := layout.Inset{Bottom: 12}.Layout(cgtx, it.Content)
			call := m.Stop()
			h := int(float32(dims.Size.Y) * prog)
			defer clip.Rect(image.Rectangle{Max: image.Pt(gtx.Constraints.Max.X, h)}).Push(gtx.Ops).Pop()
			call.Add(gtx.Ops)
			return layout.Dimensions{Size: image.Pt(gtx.Constraints.Max.X, h)}
		})
		if i < len(items)-1 {
			rows = append(rows, Hairline(th))
		}
	}
	if !a.Bordered {
		return VStack(0, rows...)(gtx)
	}
	// Bordered: side padding inside a rounded outline.
	m := op.Record(gtx.Ops)
	dims := layout.Inset{Left: 16, Right: 16}.Layout(gtx, VStack(0, rows...))
	call := m.Stop()
	dims.Size.X = gtx.Constraints.Max.X
	r := gtx.Dp(th.Radius.MD)
	defer clip.UniformRRect(image.Rectangle{Max: dims.Size}, r).Push(gtx.Ops).Pop()
	widget.Border{Color: th.Palette.Border, Width: 1, CornerRadius: th.Radius.MD}.Layout(gtx,
		func(gtx layout.Context) layout.Dimensions { return layout.Dimensions{Size: dims.Size} })
	call.Add(gtx.Ops)
	return dims
}
