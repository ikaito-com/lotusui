package lotusui

import (
	"gioui.org/font"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
)

// Breadcrumb renders the path to the current page: muted ancestor
// labels separated by chevrons, the LAST label in full ink as the
// current page. btns, when non-nil and index-aligned, makes ancestor
// labels clickable (poll Clicked on your side); the current page is
// never interactive.
//
// Prefer BreadcrumbNav for trails that must collapse like shadcn
// (ellipsis menu for the middle, ItemsToDisplay, narrow truncate).
// The composable pieces — BreadcrumbLink, BreadcrumbPage,
// BreadcrumbSep, BreadcrumbEllipsis — build custom trails by hand.
func Breadcrumb(th *Theme, btns []*widget.Clickable, labels ...string) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		var row []layout.FlexChild
		for i, lb := range labels {
			i, lb := i, lb
			last := i == len(labels)-1
			if i > 0 {
				row = append(row,
					layout.Rigid(HSpacer(th.Space.XS)),
					layout.Rigid(SVGIcon(IconChevronRight, 13, th.Palette.FgDisabled)),
					layout.Rigid(HSpacer(th.Space.XS)),
				)
			}
			item := func(gtx layout.Context) layout.Dimensions {
				l := LabelBody(th, lb)
				l.Color = th.Palette.FgSubtle
				if last {
					l.Color = th.Palette.Fg
					l.Font.Weight = font.Medium
				}
				return l.Layout(gtx)
			}
			if !last && btns != nil && i < len(btns) && btns[i] != nil {
				btn := btns[i]
				clickable := item
				item = func(gtx layout.Context) layout.Dimensions {
					return btn.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						pointer.CursorPointer.Add(gtx.Ops)
						return clickable(gtx)
					})
				}
			}
			row = append(row, layout.Rigid(item))
		}
		return layout.Flex{Alignment: layout.Middle}.Layout(gtx, row...)
	}
}

// BreadcrumbSeg is one segment of a BreadcrumbNav trail.
type BreadcrumbSeg struct {
	Label string
	// Btn makes the segment a link (and a menu row when collapsed into
	// the ellipsis). Nil is fine for the current page (last seg).
	Btn *widget.Clickable
}

// BreadcrumbSegOf is one labeled segment (no clickable).
func BreadcrumbSegOf(label string) BreadcrumbSeg { return BreadcrumbSeg{Label: label} }

// BreadcrumbSegLink is one clickable segment.
func BreadcrumbSegLink(btn *widget.Clickable, label string) BreadcrumbSeg {
	return BreadcrumbSeg{Label: label, Btn: btn}
}

// BreadcrumbSegs packs variadic segments (build-time composition).
func BreadcrumbSegs(segs ...BreadcrumbSeg) []BreadcrumbSeg {
	out := make([]BreadcrumbSeg, len(segs))
	copy(out, segs)
	return out
}

// BreadcrumbNav is the shadcn responsive trail: always the first
// segment, an ellipsis DropdownMenu for the collapsed middle when the
// path is longer than ItemsToDisplay (default 3 — shadcn), then the
// trailing visible segments. Below the theme's "md" breakpoint,
// ItemsToDisplay drops to 2 (first + last) and labels truncate like
// shadcn's max-w-20. Mobile Drawer is "from the web"; the ellipsis
// still opens a DropdownMenu on every width.
type BreadcrumbNav struct {
	// ItemsToDisplay is how many trail ends stay visible when the
	// path is long (shadcn default 3 → first + last two). Zero means 3.
	// Below md the effective value is min(ItemsToDisplay, 2).
	ItemsToDisplay int

	menu     DropdownMenuTrigger
	menuBtns []widget.Clickable // for middle segs that have no Btn
}

// Layout paints the trail for segs. The last segment is always the
// current page (BreadcrumbPage). Poll each seg's Btn.Clicked for
// navigation — including rows that only appear inside the ellipsis menu.
func (n *BreadcrumbNav) Layout(th *Theme, gtx layout.Context, segs ...BreadcrumbSeg) layout.Dimensions {
	if len(segs) == 0 {
		return layout.Dimensions{}
	}
	display := n.ItemsToDisplay
	if display <= 0 {
		display = 3
	}
	// shadcn breadcrumb-responsive: below md, fewer visible crumbs + truncate.
	truncate := false
	md := th.Breakpoints
	if md.Len() == 0 {
		md = DefaultBreakpoints
	}
	if mdIdx := md.IndexOf("md"); mdIdx >= 0 && th.BreakpointIndex(gtx) < mdIdx {
		if display > 2 {
			display = 2
		}
		truncate = true
	}
	n.menu.Variant = ButtonGhost
	n.menu.Icon = IconMoreHorizontal
	n.menu.Size = SizeSM
	n.menu.Align = PopoverStart
	if n.menu.Width == 0 {
		n.menu.Width = 200
	}

	sep := BreadcrumbSep(th, "")
	var row []layout.FlexChild
	add := func(w layout.Widget) {
		if len(row) > 0 {
			row = append(row, layout.Rigid(sep))
		}
		row = append(row, layout.Rigid(w))
	}

	add(n.segWidget(th, segs[0], false, truncate))

	if len(segs) > display {
		// Middle: segs[1 : len-(display-1)] → menu (shadcn items.slice(1, -2) when display=3).
		end := len(segs) - (display - 1)
		if end < 1 {
			end = 1
		}
		middle := segs[1:end]
		for len(n.menuBtns) < len(middle) {
			n.menuBtns = append(n.menuBtns, widget.Clickable{})
		}
		items := make([]layout.Widget, len(middle))
		for i, s := range middle {
			i, s := i, s
			btn := s.Btn
			if btn == nil {
				btn = &n.menuBtns[i]
			}
			items[i] = DropdownMenuItem(th, btn, s.Label, false)
		}
		add(func(gtx layout.Context) layout.Dimensions {
			return n.menu.Layout(th, gtx, "", items...)
		})
		for i := end; i < len(segs); i++ {
			add(n.segWidget(th, segs[i], i == len(segs)-1, truncate))
		}
	} else {
		for i := 1; i < len(segs); i++ {
			add(n.segWidget(th, segs[i], i == len(segs)-1, truncate))
		}
	}
	return layout.Flex{Alignment: layout.Middle}.Layout(gtx, row...)
}

func (n *BreadcrumbNav) segWidget(th *Theme, s BreadcrumbSeg, page, truncate bool) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		if truncate {
			max := gtx.Dp(unit.Dp(80)) // shadcn max-w-20
			if gtx.Constraints.Max.X > max || gtx.Constraints.Max.X == 0 {
				gtx.Constraints.Max.X = max
			}
		}
		if page || s.Btn == nil {
			return breadcrumbPageTrunc(th, gtx, s.Label)
		}
		return s.Btn.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			pointer.CursorPointer.Add(gtx.Ops)
			l := LabelBody(th, s.Label)
			l.Color = th.Palette.FgSubtle
			l.MaxLines = 1
			return l.Layout(gtx)
		})
	}
}

func breadcrumbPageTrunc(th *Theme, gtx layout.Context, label string) layout.Dimensions {
	l := LabelBody(th, label)
	l.Color = th.Palette.Fg
	l.Font.Weight = font.Medium
	l.MaxLines = 1
	return l.Layout(gtx)
}

// BreadcrumbLink is one clickable ancestor: muted ink, pointer
// cursor. Poll btn.Clicked on your side.
func BreadcrumbLink(th *Theme, btn *widget.Clickable, label string) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return btn.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			pointer.CursorPointer.Add(gtx.Ops)
			l := LabelBody(th, label)
			l.Color = th.Palette.FgSubtle
			return l.Layout(gtx)
		})
	}
}

// BreadcrumbPage is the current page: full ink, medium weight, never
// interactive.
func BreadcrumbPage(th *Theme, label string) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		l := LabelBody(th, label)
		l.Color = th.Palette.Fg
		l.Font.Weight = font.Medium
		return l.Layout(gtx)
	}
}

// BreadcrumbSep is the between-items glyph; empty icon means the
// default chevron. IconDot gives the dotted trail.
func BreadcrumbSep(th *Theme, icon string) layout.Widget {
	if icon == "" {
		icon = IconChevronRight
	}
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{Left: th.Space.XS, Right: th.Space.XS}.Layout(gtx,
			SVGIcon(icon, 13, th.Palette.FgDisabled))
	}
}

// BreadcrumbEllipsis marks collapsed depth: the more-horizontal glyph
// in muted ink (non-interactive). For the interactive shadcn pattern
// use BreadcrumbNav or DropdownMenuTrigger{Icon: IconMoreHorizontal,
// Variant: ButtonGhost, Align: PopoverStart}.
func BreadcrumbEllipsis(th *Theme) layout.Widget {
	return SVGIcon(IconMoreHorizontal, 16, th.Palette.FgSubtle)
}
