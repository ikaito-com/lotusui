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
	"gioui.org/widget/material"

	"image/color"
)

// TabsVariant selects the tab look.
type TabsVariant int

const (
	// TabsDefault: the whole strip sits in a muted rounded well; the
	// active tab is a raised panel inside it — the standard look.
	TabsDefault TabsVariant = iota
	// TabsLine: quiet labels over an underline that marks the active
	// tab — the classic tab strip.
	TabsLine
	// TabsSubtle: pill-styled labels, the active one filled BgSubtle —
	// the same active language as rows and toggles everywhere else.
	TabsSubtle
)

// TabOption is one tab — the choice contract lives with SelectOption;
// Icon renders before the label and Disabled dims the tab and drops
// its cursor (both replacing the index-aligned slices Tabs used to
// carry).
type TabOption struct {
	Label    string
	Value    string
	Icon     string
	Disabled bool
}

// Val is the option's stored value — its Label when Value is empty.
func (o TabOption) Val() string {
	if o.Value != "" {
		return o.Value
	}
	return o.Label
}

// TabItem is one label-only tab — Value equals Label.
func TabItem(label string) TabOption { return TabOption{Label: label} }

// TabItemValue is one tab with an explicit stored Value.
func TabItemValue(value, label string) TabOption {
	return TabOption{Label: label, Value: value}
}

// TabItems packs variadic tabs into a slice (build-time composition).
func TabItems(items ...TabOption) []TabOption {
	out := make([]TabOption, len(items))
	copy(out, items)
	return out
}

// TabOpts builds label-only tabs — the shorthand when the label IS
// the datum.
func TabOpts(labels ...string) []TabOption {
	out := make([]TabOption, len(labels))
	for i, l := range labels {
		out[i] = TabItem(l)
	}
	return out
}

// Tabs is a row of tab labels; Variant picks the pill or underline
// look. Options carry each tab's Label, Value, Icon and Disabled
// flag, and the selection is read and written as a VALUE — the
// cursor is unexported, because a tab's index is a fact about this
// strip's current order, not something an app should store or route
// on.
//
//	tabs := lotusui.Tabs{Options: lotusui.TabOpts("Account", "Password")}
//	tabs.Update(gtx)          // BEFORE anything reads the selection
//	switch tabs.Value() { ... }
//	tabs.Layout(th, gtx)
type Tabs struct {
	Options []TabOption
	Variant TabsVariant
	// Vertical stacks the tab strip in a column — pair it with your
	// content beside it.
	Vertical bool
	// sel is the cursor into Options — UNEXPORTED: callers speak
	// Value()/SetValue(). The zero value selects the first tab.
	sel  int
	btns []widget.Clickable
}

// Value is the selected tab's value ("" when nothing is selected).
func (t *Tabs) Value() string { return valueAt(t.Options, t.sel) }

// SetValue selects the tab carrying v. An unknown value selects
// nothing rather than silently falling back to the first tab.
func (t *Tabs) SetValue(v string) { t.sel = chooseValue(t.Options, v) }

// Clear leaves no tab selected.
func (t *Tabs) Clear() { t.sel = -1 }

// Chosen reports whether a tab is currently selected.
func (t *Tabs) Chosen() bool { return t.Value() != "" }

func (t *Tabs) disabled(i int) bool {
	return i < len(t.Options) && t.Options[i].Disabled
}

func (t *Tabs) icon(i int) string {
	if i < len(t.Options) {
		return t.Options[i].Icon
	}
	return ""
}

// tabLabel renders one tab's content: optional icon + label, sharing
// the label's ink.
func (t *Tabs) tabLabel(gtx layout.Context, i int, l material.LabelStyle) layout.Dimensions {
	if ic := t.icon(i); ic != "" {
		return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(SVGIcon(ic, 15, l.Color)),
			layout.Rigid(HSpacer(6)),
			layout.Rigid(l.Layout),
		)
	}
	return l.Layout(gtx)
}

// Update processes tab clicks. It MUST run before anything reads the
// selection in the same frame — otherwise the click's own frame
// renders the stale tab and the switch only shows on the next
// incidental event, which reads as lag. This bug shipped TWICE
// (Create/Select, twice in the original app), so Layout deliberately
// does NOT process clicks anymore: a consumer that forgets Update
// gets DEAD tabs — impossible to miss in the first manual test —
// instead of subtle lag that survives review. Call Update in the
// frame's handler phase (or immediately before the Value() read).
func (t *Tabs) Update(gtx layout.Context) {
	if len(t.btns) != len(t.Options) {
		t.btns = make([]widget.Clickable, len(t.Options))
	}
	for i := range t.btns {
		if t.btns[i].Clicked(gtx) && !t.disabled(i) {
			t.sel = i
		}
	}
}

func (t *Tabs) Layout(th *Theme, gtx layout.Context) layout.Dimensions {
	// NO click processing here — see Update. Rendering with a stale
	// btns slice before the first Update would panic, so size only.
	if len(t.btns) != len(t.Options) {
		t.btns = make([]widget.Clickable, len(t.Options))
	}
	gap := th.Space.SM
	if t.Variant == TabsDefault {
		gap = th.Space.XS
	}
	tabAt := func(i int) layout.Widget {
		lb := t.Options[i].Label
		return func(gtx layout.Context) layout.Dimensions {
			switch t.Variant {
			case TabsLine:
				return t.layoutLine(th, gtx, i, lb)
			case TabsDefault:
				return t.layoutEnclosed(th, gtx, i, lb)
			}
			return t.layoutSubtle(th, gtx, i, lb)
		}
	}

	var strip layout.Widget
	if t.Vertical {
		children := make([]layout.FlexChild, 0, len(t.Options)*2)
		for i := range t.Options {
			if i > 0 {
				children = append(children, layout.Rigid(Spacer(gap)))
			}
			children = append(children, layout.Rigid(tabAt(i)))
		}
		strip = func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
		}
	} else {
		// Horizontal: Wrap at intrinsic tab widths — never squeeze a
		// label into a 1-character column the way Flex+Rigid does
		// under a narrow Max.X (Split half-pane: Changes / Staging /
		// Production).
		widgets := make([]layout.Widget, len(t.Options))
		for i := range t.Options {
			widgets[i] = tabAt(i)
		}
		strip = Wrap(gap, layout.Middle, widgets...)
	}

	if t.Variant != TabsDefault {
		return strip(gtx)
	}
	// Enclosed: the strip sits in a subtle rounded well that grows
	// with wrapped height.
	m := op.Record(gtx.Ops)
	dims := layout.UniformInset(unit.Dp(3)).Layout(gtx, strip)
	call := m.Stop()
	defer clip.UniformRRect(image.Rectangle{Max: dims.Size}, gtx.Dp(th.Radius.MD+2)).Push(gtx.Ops).Pop()
	paint.Fill(gtx.Ops, th.Palette.BgSubtle)
	call.Add(gtx.Ops)
	return dims
}

// layoutSubtle renders one pill-styled tab.
func (t *Tabs) layoutSubtle(th *Theme, gtx layout.Context, i int, lb string) layout.Dimensions {
	return t.btns[i].Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		active := t.sel == i
		m := op.Record(gtx.Ops)
		dims := layout.Inset{Top: unit.Dp(8), Bottom: unit.Dp(8), Left: unit.Dp(14), Right: unit.Dp(14)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			l := LabelBody(th, lb)
			switch {
			case t.disabled(i):
				l.Color = th.Palette.FgDisabled
			case active:
				l.Color = th.Palette.Fg
				l.Font.Weight = font.SemiBold
			default:
				l.Color = th.Palette.FgSubtle
			}
			return t.tabLabel(gtx, i, l)
		})
		call := m.Stop()
		defer clip.UniformRRect(image.Rectangle{Max: dims.Size}, gtx.Dp(th.Radius.MD)).Push(gtx.Ops).Pop()
		if !t.disabled(i) {
			pointer.CursorPointer.Add(gtx.Ops)
			if active || t.btns[i].Hovered() {
				paint.Fill(gtx.Ops, th.Palette.BgSubtle)
			}
		}
		call.Add(gtx.Ops)
		return dims
	})
}

// layoutEnclosed renders one tab inside the well: the active one is a
// raised panel.
func (t *Tabs) layoutEnclosed(th *Theme, gtx layout.Context, i int, lb string) layout.Dimensions {
	return t.btns[i].Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		active := t.sel == i
		m := op.Record(gtx.Ops)
		dims := layout.Inset{Top: unit.Dp(7), Bottom: unit.Dp(7), Left: unit.Dp(14), Right: unit.Dp(14)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			l := LabelBody(th, lb)
			switch {
			case t.disabled(i):
				l.Color = th.Palette.FgDisabled
			case active:
				l.Color = th.Palette.Fg
				l.Font.Weight = font.SemiBold
			default:
				l.Color = th.Palette.FgSubtle
			}
			return t.tabLabel(gtx, i, l)
		})
		call := m.Stop()
		if active {
			// The seat sits OUTSIDE the clip — a 1dp drop below the tab.
			seatShadow(gtx, dims.Size, gtx.Dp(th.Radius.MD))
		}
		defer clip.UniformRRect(image.Rectangle{Max: dims.Size}, gtx.Dp(th.Radius.MD)).Push(gtx.Ops).Pop()
		if !t.disabled(i) {
			pointer.CursorPointer.Add(gtx.Ops)
		}
		if active {
			paint.Fill(gtx.Ops, th.Palette.BgPanel)
			widget.Border{Color: th.Palette.BorderSubtle, Width: unit.Dp(1), CornerRadius: th.Radius.MD}.Layout(gtx,
				func(gtx layout.Context) layout.Dimensions { return layout.Dimensions{Size: dims.Size} })
		} else if !t.disabled(i) && t.btns[i].Hovered() {
			paint.Fill(gtx.Ops, th.Palette.BgMuted)
		}
		call.Add(gtx.Ops)
		return dims
	})
}

// layoutLine renders one TabsLine tab: label with a 2dp underline —
// BrandFg under the active tab, transparent under the rest, hover
// pre-shadows it in BorderEmphasized. Disabled dims the label and
// drops the bar, pointer, and hover — same quiet contract as the
// other tab variants.
func (t *Tabs) layoutLine(th *Theme, gtx layout.Context, i int, lb string) layout.Dimensions {
	return t.btns[i].Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		active := t.sel == i
		off := t.disabled(i)
		bar := color.NRGBA{}
		switch {
		case off:
			// no underline while disabled
		case active:
			bar = th.Palette.BrandFg
		case t.btns[i].Hovered():
			bar = th.Palette.BorderEmphasized
		}
		m := op.Record(gtx.Ops)
		dims := layout.Inset{Top: unit.Dp(8), Bottom: unit.Dp(8), Left: unit.Dp(4), Right: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			l := LabelBody(th, lb)
			switch {
			case off:
				l.Color = th.Palette.FgDisabled
			case active:
				l.Color = th.Palette.Fg
				l.Font.Weight = font.SemiBold
			default:
				l.Color = th.Palette.FgSubtle
			}
			return t.tabLabel(gtx, i, l)
		})
		call := m.Stop()
		defer clip.Rect(image.Rectangle{Max: dims.Size}).Push(gtx.Ops).Pop()
		if !off {
			pointer.CursorPointer.Add(gtx.Ops)
		}
		call.Add(gtx.Ops)
		h := gtx.Dp(unit.Dp(2))
		barRect := image.Rect(0, dims.Size.Y-h, dims.Size.X, dims.Size.Y)
		defer clip.Rect(barRect).Push(gtx.Ops).Pop()
		paint.Fill(gtx.Ops, bar)
		return dims
	})
}
