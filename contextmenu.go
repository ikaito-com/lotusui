package lotusui

import (
	"image"
	"io"
	"strings"

	"gioui.org/io/clipboard"
	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/unit"
	"gioui.org/widget"
)

// The ContextMenu family: the menu that opens AT THE POINTER —
// right-click any content (Ctrl+click on macOS; ContextMenuPress is
// the one place that convention is answered) and the panel appears
// where you pressed, on the shared floating layer. Escape, a press
// anywhere else, or picking an item closes it. Rows are the same menu
// grammar as DropdownMenu, under this family's own names.

// ContextMenu wraps content in a pass-through pointer area — the
// child keeps every primary-button interaction (an editor its caret
// and selection) — and opens the panel on the platform's context
// gesture.
//
// Layout may be called multiple times per frame on one ContextMenu;
// each call is a distinct site. Only the site that saw the press
// paints the floating panel while open.
type ContextMenu struct {
	// KeepOpen suppresses close-on-selection — for menus of checkbox
	// or radio rows where picking is not leaving.
	KeepOpen bool
	// Width is the panel's max width; zero means min 224dp and grow
	// with content (shadcn min-w). Non-zero: hug up to Width.
	Width unit.Dp

	open   bool
	at     image.Point // press position, local to the active site
	sites  layoutSites
	areas  []*ctxSite
	active int

	dismiss dismisser
}

type ctxSite struct {
	tag      struct{ _ byte }
	closeTag struct{ _ byte }
}

func (c *ContextMenu) siteAt(i int) *ctxSite {
	for len(c.areas) <= i {
		c.areas = append(c.areas, new(ctxSite))
	}
	return c.areas[i]
}

// Layout draws content and watches it for the context gesture; items
// are the menu rows shown when it fires.
func (c *ContextMenu) Layout(th *Theme, gtx layout.Context, content layout.Widget, items ...layout.Widget) layout.Dimensions {
	// A throwaway pass has no site: measure the content alone, touch
	// no site state, never the floating panel.
	if inMeasurePass() {
		return content(gtx)
	}
	idx := c.sites.next(gtx.Now)
	site := c.siteAt(idx)

	// Dismiss once per frame (first site) so multi-Layout does not
	// race the catcher.
	if idx == 0 && c.open && c.dismiss.Dismissed(gtx) {
		c.open = false
	}
	if !c.KeepOpen && c.open && c.active == idx {
		// Close on selection: releases inside the panel arrive here
		// through a pass-through area, AFTER the row's own click.
		for {
			ev, ok := gtx.Event(pointer.Filter{Target: &site.closeTag, Kinds: pointer.Release})
			if !ok {
				break
			}
			if _, isPtr := ev.(pointer.Event); isPtr {
				c.open = false
			}
		}
	}
	for {
		ev, ok := gtx.Event(pointer.Filter{Target: &site.tag, Kinds: pointer.Press})
		if !ok {
			break
		}
		if e, isPtr := ev.(pointer.Event); isPtr && ContextMenuPress(e) {
			c.open = true
			c.active = idx
			c.at = image.Pt(int(e.Position.X+0.5), int(e.Position.Y+0.5))
		}
	}

	dims := content(gtx)

	// Pass-through input area over the content: primary events flow
	// to the child untouched; we only read the context gesture.
	func() {
		defer pointer.PassOp{}.Push(gtx.Ops).Pop()
		defer clip.Rect(image.Rectangle{Max: dims.Size}).Push(gtx.Ops).Pop()
		event.Op(gtx.Ops, &site.tag)
	}()

	if c.open && c.active == idx {
		avail := gtx.Constraints.Max
		Floating(gtx, func(gtx layout.Context) layout.Dimensions {
			c.dismiss.Add(gtx)
			minW, maxW := gtx.Dp(224), 1<<14
			if c.Width != 0 {
				minW, maxW = 0, gtx.Dp(c.Width)
			}
			gtx.Constraints = layout.Constraints{Min: image.Pt(minW, 0), Max: image.Pt(maxW, 1<<14)}
			m := op.Record(gtx.Ops)
			d := DropdownMenu(th, items...)(gtx)
			call := m.Stop()
			defer op.Offset(contextMenuPlace(c.at, d.Size, avail)).Push(gtx.Ops).Pop()
			cardShadow(gtx, d.Size, gtx.Dp(th.Radius.MD))
			call.Add(gtx.Ops)
			if !c.KeepOpen {
				defer pointer.PassOp{}.Push(gtx.Ops).Pop()
				defer clip.Rect(image.Rectangle{Max: d.Size}).Push(gtx.Ops).Pop()
				event.Op(gtx.Ops, &site.closeTag)
			}
			return d
		})
	}
	return dims
}

// contextMenuPlace positions a panel opened at press inside a box of
// avail extent: down-right of the pointer, flipped up/left — the
// native menu move — when the panel would overflow a KNOWN bound, and
// clamped to the bound when neither side fits. An unbounded axis (the
// 2^14 "infinite" sentinel or beyond, e.g. inside a scroller) reports
// no edge, so the panel simply opens down-right — a window edge is
// unknowable from local coordinates there.
func contextMenuPlace(press, panel, avail image.Point) image.Point {
	const unbounded = 1 << 13
	place := func(at, size, bound int) int {
		if bound >= unbounded || at+size <= bound {
			return at
		}
		if flipped := at - size; flipped >= 0 {
			return flipped
		}
		return max(bound-size, 0)
	}
	return image.Pt(place(press.X, panel.X, avail.X), place(press.Y, panel.Y, avail.Y))
}

// The row family, in this family's own vocabulary. Each is the
// DropdownMenu row of the same shape — one grammar, two menus.

// ContextMenuItem is one action in the menu — a full-width row in the
// shared menu grammar. danger marks a destructive action.
func ContextMenuItem(th *Theme, btn *widget.Clickable, label string, danger bool) layout.Widget {
	return DropdownMenuItem(th, btn, label, danger)
}

// ContextMenuItemIcon is ContextMenuItem with a leading icon, tinted
// with the row's own ink.
func ContextMenuItemIcon(th *Theme, btn *widget.Clickable, icon, label string, danger bool) layout.Widget {
	return DropdownMenuItemIcon(th, btn, icon, label, danger)
}

// ContextMenuShortcutItem is ContextMenuItem with a right-aligned
// keyboard hint (⌘C, ⇧⌘S…) in muted ink. The hint is display only —
// binding the key is the caller's; ShortcutHint spells the platform
// modifier.
func ContextMenuShortcutItem(th *Theme, btn *widget.Clickable, label, shortcut string, danger bool) layout.Widget {
	return DropdownMenuShortcutItem(th, btn, label, shortcut, danger)
}

// ContextMenuCheckboxItem is a toggleable row: a check-mark gutter,
// marked while checked. State lives with the caller — flip it on
// Clicked, immediate-mode style.
func ContextMenuCheckboxItem(th *Theme, btn *widget.Clickable, label string, checked bool) layout.Widget {
	return DropdownMenuCheckboxItem(th, btn, label, checked)
}

// ContextMenuRadioItem is an exclusive-choice row: a dot gutter,
// marked on the selected row of the caller's group.
func ContextMenuRadioItem(th *Theme, btn *widget.Clickable, label string, selected bool) layout.Widget {
	return DropdownMenuRadioItem(th, btn, label, selected)
}

// ContextMenuCheckboxItemIcon is ContextMenuCheckboxItem with a
// leading icon between the gutter and the label.
func ContextMenuCheckboxItemIcon(th *Theme, btn *widget.Clickable, icon, label string, checked bool) layout.Widget {
	return DropdownMenuCheckboxItemIcon(th, btn, icon, label, checked)
}

// ContextMenuRadioItemIcon is ContextMenuRadioItem with a leading
// icon between the gutter and the label.
func ContextMenuRadioItemIcon(th *Theme, btn *widget.Clickable, icon, label string, selected bool) layout.Widget {
	return DropdownMenuRadioItemIcon(th, btn, icon, label, selected)
}

// ContextMenuLabel is a non-interactive group heading inside the menu.
func ContextMenuLabel(th *Theme, text string) layout.Widget {
	return DropdownMenuLabel(th, text)
}

// ContextMenuSeparator is the hairline between menu groups.
func ContextMenuSeparator(th *Theme) layout.Widget {
	return DropdownMenuSeparator(th)
}

// ContextMenuSub is the nested submenu — a trigger row whose side
// panel opens on hover. One struct per submenu; use its Item method
// inside the items list.
type ContextMenuSub = DropdownMenuSub

// writeClipboard puts s on the system clipboard (the CodeBlock Copy
// pattern).
func writeClipboard(gtx layout.Context, s string) {
	gtx.Execute(clipboard.WriteCmd{
		Type: "application/text",
		Data: io.NopCloser(strings.NewReader(s)),
	})
}

// EditorContextMenu is the read-only-editor convenience — a lotusui
// extension (no shadcn counterpart). Wrap a laid-out widget.Editor (a
// log pane, a diff view) and the context gesture opens Copy / Copy
// all / Select all at the pointer: the discoverable mouse path to the
// clipboard. Gio already binds the copy SHORTCUT itself — ⌘C / Ctrl+C
// works in a focused read-only editor — so this adds only the menu.
// Copy shows while a selection exists and carries the platform hint.
//
//	var logMenu lotusui.EditorContextMenu // beside your widget.Editor
//	logMenu.Layout(th, gtx, &logEd, laidOutEditor)
type EditorContextMenu struct {
	// Menu is the underlying ContextMenu — reach in for Width.
	Menu ContextMenu

	copyBtn, copyAllBtn, selectAllBtn widget.Clickable
}

// Layout wraps content — the widget that renders ed — in the
// context-menu area.
func (m *EditorContextMenu) Layout(th *Theme, gtx layout.Context, ed *widget.Editor, content layout.Widget) layout.Dimensions {
	// A throwaway pass has no site and must touch no state.
	if inMeasurePass() {
		return content(gtx)
	}
	if m.copyBtn.Clicked(gtx) {
		writeClipboard(gtx, ed.SelectedText())
	}
	if m.copyAllBtn.Clicked(gtx) {
		writeClipboard(gtx, ed.Text())
	}
	if m.selectAllBtn.Clicked(gtx) {
		ed.SetCaret(0, ed.Len())
	}

	// Snapshot the selection BEFORE the editor lays out: on macOS the
	// Ctrl+primary context press is ALSO a primary click to the
	// editor, which collapses the selection in this same frame. When
	// this frame's press opened the menu, restore the snapshot after —
	// Copy must act on the text the user right-clicked.
	selStart, selEnd := ed.Selection()
	wasOpen := m.Menu.open

	items := make([]layout.Widget, 0, 3)
	if selStart != selEnd {
		items = append(items, ContextMenuShortcutItem(th, &m.copyBtn, "Copy", ShortcutHint("C"), false))
	}
	items = append(items,
		ContextMenuItem(th, &m.copyAllBtn, "Copy all", false),
		ContextMenuShortcutItem(th, &m.selectAllBtn, "Select all", ShortcutHint("A"), false),
	)
	dims := m.Menu.Layout(th, gtx, content, items...)
	if !wasOpen && m.Menu.open && selStart != selEnd {
		ed.SetCaret(selStart, selEnd)
	}
	return dims
}
