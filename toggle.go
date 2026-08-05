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

// Toggle is the pressed-state button: an icon or label that stays
// filled while On. State lives on the struct; clicking flips it.
type Toggle struct {
	On  bool
	btn widget.Clickable
}

// ToggleProps are the toggle's options — an Icon, a Label, or both.
type ToggleProps struct {
	Icon  string
	Label string
	// Content replaces Icon/Label with arbitrary content inside the
	// toggle's chrome.
	Content  layout.Widget
	Size     Size
	Outline  bool // bordered when off
	Disabled bool
}

func (t *Toggle) Layout(th *Theme, gtx layout.Context, o ToggleProps) layout.Dimensions {
	if t.btn.Clicked(gtx) && !o.Disabled {
		t.On = !t.On
	}
	if o.Disabled {
		gtx = gtx.Disabled()
	}
	return t.btn.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		hovered := t.btn.Hovered() && !o.Disabled
		ink := th.Palette.FgMuted
		if o.Disabled {
			ink = th.Palette.FgDisabled
		}
		bo := ButtonProps{Size: o.Size}
		ratio, vPad, hPad := bo.metrics()
		if o.Label == "" {
			hPad = vPad
		}
		m := op.Record(gtx.Ops)
		dims := layout.Inset{Top: vPad, Bottom: vPad, Left: hPad, Right: hPad}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			if o.Content != nil {
				return o.Content(gtx)
			}
			iconSz := unit.Dp(float32(gtx.Metric.SpToDp(Sp(th, ratio))) + 2)
			var row []layout.FlexChild
			if o.Icon != "" {
				row = append(row, layout.Rigid(SVGIcon(o.Icon, iconSz, ink)))
			}
			if o.Label != "" {
				if len(row) > 0 {
					row = append(row, layout.Rigid(HSpacer(th.Space.SM)))
				}
				row = append(row, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					l := LabelBody(th, o.Label)
					l.Color = ink
					l.Font.Weight = font.Medium
					return l.Layout(gtx)
				}))
			}
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx, row...)
		})
		call := m.Stop()
		r := gtx.Dp(th.Radius.SM + 2)
		defer clip.UniformRRect(image.Rectangle{Max: dims.Size}, r).Push(gtx.Ops).Pop()
		if !o.Disabled {
			pointer.CursorPointer.Add(gtx.Ops)
		}
		switch {
		case t.On:
			paint.Fill(gtx.Ops, th.Palette.BgEmphasized)
		case hovered:
			paint.Fill(gtx.Ops, th.Palette.BgSubtle)
		}
		if o.Outline && !t.On {
			widget.Border{Color: th.Palette.Border, Width: unit.Dp(1), CornerRadius: th.Radius.SM + 2}.Layout(gtx,
				func(gtx layout.Context) layout.Dimensions { return layout.Dimensions{Size: dims.Size} })
		}
		call.Add(gtx.Ops)
		return dims
	})
}

// ToggleOption is one toggle of a group — the choice contract lives
// with SelectOption. Icon renders before the label; Content replaces
// both with arbitrary content inside the toggle's chrome.
type ToggleOption struct {
	Label   string
	Value   string
	Icon    string
	Content layout.Widget
}

// Val is the option's stored value — its Label when Value is empty.
func (o ToggleOption) Val() string {
	if o.Value != "" {
		return o.Value
	}
	return o.Label
}

// ToggleItem is one label-only toggle option — Value equals Label.
func ToggleItem(label string) ToggleOption { return ToggleOption{Label: label} }

// ToggleItemValue is one toggle option with an explicit stored Value.
func ToggleItemValue(value, label string) ToggleOption {
	return ToggleOption{Label: label, Value: value}
}

// ToggleItems packs variadic options into a slice (build-time composition).
func ToggleItems(items ...ToggleOption) []ToggleOption {
	out := make([]ToggleOption, len(items))
	copy(out, items)
	return out
}

// ToggleOpts builds label-only toggle options — the shorthand when
// the label IS the datum. SelectOpts, RadioOpts and TabOpts are its
// siblings.
func ToggleOpts(labels ...string) []ToggleOption {
	out := make([]ToggleOption, len(labels))
	for i, l := range labels {
		out[i] = ToggleItem(l)
	}
	return out
}

// ToggleGroup coordinates a row of toggles over a shared Option
// list: single-select by default (radio semantics — clicking the
// selected item clears it), or independent choices with Multiple.
// Selection is read and written as VALUES; the cursor and the
// per-option bools are unexported, because a position in this list
// is not something an app should store.
//
//	view := lotusui.ToggleGroup{Options: lotusui.ToggleOpts("All", "Missed"), Outline: true}
//	view.SetValue(stored)
//	view.Layout(th, gtx, lotusui.SizeMD)
//	stored = view.Value()
//
//	marks := lotusui.ToggleGroup{Multiple: true, Options: []lotusui.ToggleOption{
//		{Label: "Bold", Value: "bold", Icon: lotusui.IconTextBold},
//		{Label: "Italic", Value: "italic", Icon: lotusui.IconTextItalic},
//	}}
//	marks.SetValues([]string{"bold"})
//	active := marks.Values() // in Options order
type ToggleGroup struct {
	Options  []ToggleOption
	Multiple bool
	// Outline borders every off item; Disabled dims the whole group;
	// Vertical stacks instead of rowing; Spacing overrides the 2dp
	// gap between items.
	Outline  bool
	Disabled bool
	Vertical bool
	Spacing  unit.Dp
	// sel is the single-select cursor, on the multi-select state —
	// both UNEXPORTED, both aligned to Options.
	sel   int
	on    []bool
	items []Toggle
}

// sync sizes the per-option state to the current Options.
func (g *ToggleGroup) sync() {
	if len(g.items) != len(g.Options) {
		g.items = make([]Toggle, len(g.Options))
	}
	if len(g.on) != len(g.Options) {
		on := make([]bool, len(g.Options))
		copy(on, g.on)
		g.on = on
	}
}

// Value is the selected option's value in single-select mode ("" when
// nothing is selected, and always "" when Multiple is set — use
// Values there).
func (g *ToggleGroup) Value() string {
	if g.Multiple {
		return ""
	}
	return valueAt(g.Options, g.sel)
}

// SetValue selects the option carrying v (single-select). An unknown
// value selects nothing rather than falling back to the first option.
func (g *ToggleGroup) SetValue(v string) { g.sel = chooseValue(g.Options, v) }

// Clear leaves nothing selected, in either mode.
func (g *ToggleGroup) Clear() {
	g.sel = -1
	g.sync()
	for i := range g.on {
		g.on[i] = false
	}
}

// Chosen reports whether anything is selected.
func (g *ToggleGroup) Chosen() bool { return len(g.Values()) > 0 }

// Values are the selected options' values in Options order — one
// entry at most in single-select mode.
func (g *ToggleGroup) Values() []string {
	if !g.Multiple {
		if v := valueAt(g.Options, g.sel); v != "" {
			return []string{v}
		}
		return nil
	}
	g.sync()
	var out []string
	for i, o := range g.Options {
		if g.on[i] {
			out = append(out, o.Val())
		}
	}
	return out
}

// SetValues selects exactly the options carrying vs (multi-select);
// unknown values are ignored. In single-select mode the first
// recognised value wins.
func (g *ToggleGroup) SetValues(vs []string) {
	if !g.Multiple {
		g.sel = -1
		for _, v := range vs {
			if i := chooseValue(g.Options, v); i >= 0 {
				g.sel = i
				return
			}
		}
		return
	}
	g.sync()
	for i := range g.on {
		g.on[i] = false
	}
	for _, v := range vs {
		if i := chooseValue(g.Options, v); i >= 0 {
			g.on[i] = true
		}
	}
}

func (g *ToggleGroup) Layout(th *Theme, gtx layout.Context, size Size) layout.Dimensions {
	g.sync()
	gap := unit.Dp(2)
	if g.Spacing != 0 {
		gap = g.Spacing
	}
	var row []layout.FlexChild
	for i, o := range g.Options {
		i, o := i, o
		if i > 0 {
			if g.Vertical {
				row = append(row, layout.Rigid(Spacer(gap)))
			} else {
				row = append(row, layout.Rigid(HSpacer(gap)))
			}
		}
		row = append(row, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			// Sync the item's On from the group state, then render; the
			// item flips itself on click and the group reconciles.
			if g.Multiple {
				g.items[i].On = g.on[i]
			} else {
				g.items[i].On = g.sel == i
			}
			d := g.items[i].Layout(th, gtx, ToggleProps{Icon: o.Icon, Label: o.Label, Content: o.Content,
				Size: size, Outline: g.Outline, Disabled: g.Disabled})
			if g.Multiple {
				g.on[i] = g.items[i].On
			} else if g.items[i].On {
				g.sel = i
			} else if g.sel == i && !g.items[i].On {
				g.sel = -1
			}
			return d
		}))
	}
	if g.Vertical {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx, row...)
	}
	return layout.Flex{Alignment: layout.Middle}.Layout(gtx, row...)
}
