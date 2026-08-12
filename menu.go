package lotusui

import (
	"image"

	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
)

// The DropdownMenu family: a panel of actions. DropdownMenu is the CONTENT — the
// bordered panel that stacks items; DropdownMenuItem is one action row inside
// it. Where the panel appears is the caller's decision: inline in an
// actions box, or layered over the screen at window constraints (the
// same portal rule as Modal). A floating, anchored trigger — the
// popover half of the family — is on the roadmap.

// unboundedX is the "no known edge" threshold: a Max.X at or above it
// is Gio's infinite sentinel (a scroller, or a floating panel asking
// content to hug), never a real width to fill.
const unboundedX = 1 << 13

// fillWidth makes a row span its panel — but only against a REAL
// width. Filling an unbounded max would report a 16384px row, which is
// how the floating menus once painted a panel wider than the window.
func fillWidth(gtx layout.Context) layout.Context {
	if gtx.Constraints.Max.X < unboundedX {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
	}
	return gtx
}

// menuPanelWidth resolves a floating menu panel's width: measure the
// items at their intrinsic width (rows do not fill an unbounded max),
// then clamp into [minW, maxW] — shadcn's min-w, hug, cap.
func menuPanelWidth(th *Theme, gtx layout.Context, minW, maxW int, items ...layout.Widget) int {
	mgtx := gtx
	mgtx.Constraints = layout.Constraints{Max: image.Pt(1<<14, 1<<14)}
	w := MeasurePass(mgtx, DropdownMenu(th, items...)).Size.X
	if w < minW {
		w = minW
	}
	if w > maxW {
		w = maxW
	}
	if w < 1 {
		w = 1
	}
	return w
}

// DropdownMenu lays out items as the standard menu panel: BgPanel surface,
// border, soft radius. Vertical edge inset is Space.XS; the gap between
// items is half of that so adjacent hover pills don't read as
// double-spaced. Rows span the full panel width (same rhythm as Select).
func DropdownMenu(th *Theme, items ...layout.Widget) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		m := op.Record(gtx.Ops)
		dims := layout.Inset{Top: th.Space.XS, Bottom: th.Space.XS}.Layout(gtx,
			VStack(th.Space.XS/2, items...))
		call := m.Stop()
		r := gtx.Dp(th.Radius.MD)
		defer clip.UniformRRect(image.Rectangle{Max: dims.Size}, r).Push(gtx.Ops).Pop()
		paint.Fill(gtx.Ops, th.Palette.BgPanel)
		widget.Border{Color: th.Palette.Border, Width: unit.Dp(1), CornerRadius: th.Radius.MD}.Layout(gtx,
			func(gtx layout.Context) layout.Dimensions {
				return layout.Dimensions{Size: dims.Size}
			})
		call.Add(gtx.Ops)
		return dims
	}
}

// DropdownMenuItem is one action in a menu: a full-width interactive ROW —
// start-aligned label, pointer cursor, and the same hover-pill
// language as every other interactive row. danger marks a destructive
// action — danger INK on a danger-tinted hover, never a saturated
// fill.
func DropdownMenuItem(th *Theme, btn *widget.Clickable, label string, danger bool) layout.Widget {
	return menuRow(th, btn, label, menuRowCfg{danger: danger})
}

// DropdownMenuItemIcon is DropdownMenuItem with a leading icon, tinted
// with the row's own ink.
func DropdownMenuItemIcon(th *Theme, btn *widget.Clickable, icon, label string, danger bool) layout.Widget {
	return menuRow(th, btn, label, menuRowCfg{icon: icon, danger: danger})
}

// DropdownMenuShortcutItem is DropdownMenuItem with a right-aligned
// keyboard hint (⌘S, ⇧⌘D…) in muted ink. The hint is display only —
// binding the key is the caller's.
func DropdownMenuShortcutItem(th *Theme, btn *widget.Clickable, label, shortcut string, danger bool) layout.Widget {
	return menuRow(th, btn, label, menuRowCfg{shortcut: shortcut, danger: danger})
}

// DropdownMenuCheckboxItem is a toggleable row: a check-mark gutter,
// marked while checked. State lives with the caller — flip it on
// Clicked, immediate-mode style.
func DropdownMenuCheckboxItem(th *Theme, btn *widget.Clickable, label string, checked bool) layout.Widget {
	return menuRow(th, btn, label, menuRowCfg{gutter: true, marked: checked})
}

// DropdownMenuRadioItem is an exclusive-choice row: a dot gutter,
// marked on the selected row of the caller's group.
func DropdownMenuRadioItem(th *Theme, btn *widget.Clickable, label string, selected bool) layout.Widget {
	return menuRow(th, btn, label, menuRowCfg{gutter: true, marked: selected, dot: true})
}

// DropdownMenuCheckboxItemIcon is DropdownMenuCheckboxItem with a
// leading icon between the gutter and the label.
func DropdownMenuCheckboxItemIcon(th *Theme, btn *widget.Clickable, icon, label string, checked bool) layout.Widget {
	return menuRow(th, btn, label, menuRowCfg{gutter: true, marked: checked, icon: icon})
}

// DropdownMenuRadioItemIcon is DropdownMenuRadioItem with a leading
// icon between the gutter and the label.
func DropdownMenuRadioItemIcon(th *Theme, btn *widget.Clickable, icon, label string, selected bool) layout.Widget {
	return menuRow(th, btn, label, menuRowCfg{gutter: true, marked: selected, dot: true, icon: icon})
}

// DropdownMenuLabel is a non-interactive group heading inside a menu.
func DropdownMenuLabel(th *Theme, text string) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		gtx = fillWidth(gtx)
		return layout.Inset{Top: unit.Dp(6), Bottom: unit.Dp(2), Left: unit.Dp(12), Right: unit.Dp(12)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			l := LabelCaption(th, text)
			l.Color = th.Palette.FgSubtle
			return l.Layout(gtx)
		})
	}
}

// DropdownMenuSeparator is the hairline between menu groups.
func DropdownMenuSeparator(th *Theme) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		// A separator has no intrinsic width — against an unbounded max
		// it must measure as zero, or a hairline would decide how wide
		// the whole panel is.
		if gtx.Constraints.Max.X >= unboundedX {
			gtx.Constraints.Max.X = 0
		}
		gtx = fillWidth(gtx)
		return layout.Inset{Top: unit.Dp(4), Bottom: unit.Dp(4)}.Layout(gtx, Hairline(th))
	}
}

// DropdownMenuSub is the nested submenu: a trigger row with a
// trailing chevron whose side panel opens while the pointer rests on
// the row or the panel itself. One struct per submenu; use its Item
// method inside a menu's item list. If Item's widget is laid out more
// than once per frame, only the hovered site paints the side panel.
type DropdownMenuSub struct {
	// Width is the side panel's max width; zero means min 200dp and
	// grow with content (shadcn min-w). Non-zero: hug up to Width.
	Width unit.Dp

	sites  layoutSites
	rows   []*subSite
	active int
	// measureRow backs the row during a throwaway pass, which has no
	// site and must never touch a live site's state.
	measureRow subSite
}

type subSite struct {
	btn       widget.Clickable
	overPanel bool
}

func (s *DropdownMenuSub) siteAt(i int) *subSite {
	for len(s.rows) <= i {
		s.rows = append(s.rows, new(subSite))
	}
	return s.rows[i]
}

// Item renders the trigger row; while hovered (row or panel) the
// submenu floats beside it.
func (s *DropdownMenuSub) Item(th *Theme, label string, items ...layout.Widget) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		// A throwaway pass has no site: draw the row, never the panel.
		if inMeasurePass() {
			return menuRow(th, &s.measureRow.btn, label, menuRowCfg{chev: true})(gtx)
		}
		idx := s.sites.next(gtx.Now)
		site := s.siteAt(idx)
		for {
			ev, ok := gtx.Event(pointer.Filter{Target: site, Kinds: pointer.Enter | pointer.Leave | pointer.Cancel})
			if !ok {
				break
			}
			if e, isPtr := ev.(pointer.Event); isPtr {
				site.overPanel = e.Kind == pointer.Enter
			}
		}
		dims := menuRow(th, &site.btn, label, menuRowCfg{chev: true})(gtx)
		open := site.btn.Hovered() || site.overPanel
		if open {
			s.active = idx
		}
		if open && s.active == idx {
			Floating(gtx, func(gtx layout.Context) layout.Dimensions {
				defer op.Offset(image.Pt(dims.Size.X, 0)).Push(gtx.Ops).Pop()
				minW, maxW := gtx.Dp(200), 1<<14
				if s.Width != 0 {
					minW, maxW = 0, gtx.Dp(s.Width)
				}
				w := menuPanelWidth(th, gtx, minW, maxW, items...)
				gtx.Constraints = layout.Constraints{Min: image.Pt(w, 0), Max: image.Pt(w, 1<<14)}
				m := op.Record(gtx.Ops)
				d := DropdownMenu(th, items...)(gtx)
				call := m.Stop()
				cardShadow(gtx, d.Size, gtx.Dp(th.Radius.MD))
				call.Add(gtx.Ops)
				// Hover tracking over the panel keeps it open while the
				// pointer travels; pass-through so rows still act.
				defer pointer.PassOp{}.Push(gtx.Ops).Pop()
				defer clip.Rect(image.Rectangle{Max: d.Size}).Push(gtx.Ops).Pop()
				event.Op(gtx.Ops, site)
				return d
			})
		}
		return dims
	}
}

type menuRowCfg struct {
	icon     string
	shortcut string
	danger   bool
	gutter   bool // reserve the check/dot gutter
	marked   bool // paint the mark
	dot      bool // radio dot instead of check
	chev     bool // trailing chevron — the submenu trigger
}

func menuRow(th *Theme, btn *widget.Clickable, label string, cfg menuRowCfg) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return btn.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			// hug: no real width to fill (the panel is measuring), so the
			// label stays Rigid and the row reports its natural width.
			hug := gtx.Constraints.Max.X >= unboundedX
			gtx = fillWidth(gtx)
			ink := th.Palette.Fg
			if cfg.danger {
				ink = th.Palette.Danger
			}
			m := op.Record(gtx.Ops)
			dims := layout.Inset{Top: unit.Dp(10), Bottom: unit.Dp(10), Left: unit.Dp(12), Right: unit.Dp(12)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				var row []layout.FlexChild
				if cfg.gutter {
					row = append(row, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						sz := gtx.Dp(16)
						if !cfg.marked {
							return layout.Dimensions{Size: image.Pt(sz, 0)}
						}
						if cfg.dot {
							// The radio dot: a filled circle in the row ink.
							d := gtx.Dp(6)
							off := op.Offset(image.Pt((sz-d)/2, (gtx.Dp(16)-d)/2)).Push(gtx.Ops)
							defer clip.UniformRRect(image.Rectangle{Max: image.Pt(d, d)}, d/2).Push(gtx.Ops).Pop()
							paint.Fill(gtx.Ops, ink)
							off.Pop()
							return layout.Dimensions{Size: image.Pt(sz, gtx.Dp(16))}
						}
						return SVGIcon(IconAccept, 16, ink)(gtx)
					}), layout.Rigid(HSpacer(th.Space.SM)))
				}
				if cfg.icon != "" {
					row = append(row, layout.Rigid(SVGIcon(cfg.icon, 16, ink)), layout.Rigid(HSpacer(th.Space.SM)))
				}
				labelW := func(gtx layout.Context) layout.Dimensions {
					l := LabelBody(th, label)
					l.Color = ink
					l.MaxLines = 1
					return l.Layout(gtx)
				}
				if hug {
					// Natural width: the label sizes to its text, and the
					// trailing hint keeps the gap it will hold once the
					// panel is laid out at this measured width.
					row = append(row, layout.Rigid(labelW))
					if cfg.shortcut != "" || cfg.chev {
						row = append(row, layout.Rigid(HSpacer(th.Space.LG)))
					}
				} else {
					// The label takes the slack, so the hint sits flush right.
					row = append(row, layout.Flexed(1, labelW))
				}
				if cfg.shortcut != "" {
					row = append(row, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						l := LabelCaption(th, cfg.shortcut)
						l.Color = th.Palette.FgSubtle
						return l.Layout(gtx)
					}))
				}
				if cfg.chev {
					row = append(row, layout.Rigid(SVGIcon(IconChevronRight, 14, th.Palette.FgSubtle)))
				}
				return layout.Flex{Alignment: layout.Middle}.Layout(gtx, row...)
			})
			call := m.Stop()
			if !hug {
				dims.Size.X = gtx.Constraints.Max.X
			}
			defer clip.UniformRRect(image.Rectangle{Max: dims.Size}, gtx.Dp(th.Radius.SM)).Push(gtx.Ops).Pop()
			pointer.CursorPointer.Add(gtx.Ops)
			if btn.Hovered() {
				fill := th.Palette.BgSubtle
				if cfg.danger {
					fill = th.Palette.DangerBg
				}
				paint.Fill(gtx.Ops, fill)
			}
			call.Add(gtx.Ops)
			return dims
		})
	}
}

// DropdownMenuTrigger is the anchored, floating form of the family —
// the standard shape: a trigger button that opens the panel on the
// portal layer, 4dp below itself. Rows keep their caller-owned
// clickables; the panel closes on selection (any release inside),
// on a press anywhere else, and on Escape. Checkbox and radio rows
// usually WANT the menu to stay open — hold KeepOpen for those.
//
// Layout may be called multiple times per frame on one trigger; each
// call is a distinct site with its own button. Only the active site
// paints the floating panel while Open.
type DropdownMenuTrigger struct {
	Open bool
	// KeepOpen suppresses close-on-selection — for menus of checkbox
	// or radio rows where picking is not leaving.
	KeepOpen bool
	// Width is the panel's max width; zero means min 224dp and grow
	// with content (shadcn min-w). Non-zero: hug up to Width.
	Width unit.Dp
	// Variant styles the trigger button; the zero value renders
	// outline (the family default), so pass ButtonGhost or ButtonLink
	// for quieter triggers.
	Variant ButtonVariant
	// Icon names an embedded icon on the trigger. Empty label + Icon
	// = icon-only (the BreadcrumbEllipsis-as-menu-trigger pattern).
	// With a label it sits at IconStart.
	Icon string
	// Size is the shared size preset for the trigger button.
	Size Size
	// Align positions the panel against the trigger edge. Zero is
	// PopoverCenter (same enum as Popover); set PopoverStart for a
	// leading-edge menu, PopoverEnd for a trailing-edge one.
	Align PopoverAlign

	sites   layoutSites
	trigs   []*menuTrigSite
	active  int
	dismiss dismisser
	// measureBtn backs the trigger during a throwaway pass, which has
	// no site and must never touch a live site's state.
	measureBtn widget.Clickable
}

type menuTrigSite struct {
	btn      widget.Clickable
	closeTag struct{ _ byte }
}

func (t *DropdownMenuTrigger) siteAt(i int) *menuTrigSite {
	for len(t.trigs) <= i {
		t.trigs = append(t.trigs, new(menuTrigSite))
	}
	return t.trigs[i]
}

// triggerBtn renders the trigger button; shared so a measure pass
// reports exactly the size the live pass will paint.
func (t *DropdownMenuTrigger) triggerBtn(th *Theme, btn *widget.Clickable, label string) layout.Widget {
	v := t.Variant
	if v == ButtonDefault {
		v = ButtonOutline
	}
	return Button(th, btn, label, ButtonProps{Variant: v, Size: t.Size, IconStart: t.Icon})
}

func (t *DropdownMenuTrigger) Layout(th *Theme, gtx layout.Context, label string, items ...layout.Widget) layout.Dimensions {
	// A throwaway pass has no site: measure the trigger button, never
	// the floating panel.
	if inMeasurePass() {
		return t.triggerBtn(th, &t.measureBtn, label)(gtx)
	}
	idx := t.sites.next(gtx.Now)
	site := t.siteAt(idx)

	// Dismiss once per frame (first site) so multi-Layout does not
	// race the catcher.
	if idx == 0 && t.Open && t.dismiss.Dismissed(gtx) {
		t.Open = false
	}
	if !t.KeepOpen && t.Open && t.active == idx {
		// Close on selection: releases inside the panel arrive here
		// through a pass-through area, AFTER the row's own click.
		for {
			ev, ok := gtx.Event(pointer.Filter{Target: &site.closeTag, Kinds: pointer.Release})
			if !ok {
				break
			}
			if _, isPtr := ev.(pointer.Event); isPtr {
				t.Open = false
			}
		}
	}
	if site.btn.Clicked(gtx) {
		if t.Open && t.active == idx {
			t.Open = false
		} else {
			t.Open = true
			t.active = idx
		}
	}
	dims := t.triggerBtn(th, &site.btn, label)(gtx)
	if t.Open && t.active == idx {
		Floating(gtx, func(gtx layout.Context) layout.Dimensions {
			t.dismiss.Add(gtx)
			minW, maxW := gtx.Dp(224), 1<<14
			if t.Width != 0 {
				minW, maxW = 0, gtx.Dp(t.Width)
			}
			w := menuPanelWidth(th, gtx, minW, maxW, items...)
			gtx.Constraints = layout.Constraints{Min: image.Pt(w, 0), Max: image.Pt(w, 1<<14)}
			m := op.Record(gtx.Ops)
			d := DropdownMenu(th, items...)(gtx)
			call := m.Stop()
			x := (dims.Size.X - d.Size.X) / 2 // PopoverCenter (zero)
			switch t.Align {
			case PopoverStart:
				x = 0
			case PopoverEnd:
				x = dims.Size.X - d.Size.X
			}
			defer op.Offset(image.Pt(x, dims.Size.Y+gtx.Dp(4))).Push(gtx.Ops).Pop()
			cardShadow(gtx, d.Size, gtx.Dp(th.Radius.MD))
			call.Add(gtx.Ops)
			if !t.KeepOpen {
				defer pointer.PassOp{}.Push(gtx.Ops).Pop()
				defer clip.Rect(image.Rectangle{Max: d.Size}).Push(gtx.Ops).Pop()
				event.Op(gtx.Ops, &site.closeTag)
			}
			return d
		})
	}
	return dims
}
