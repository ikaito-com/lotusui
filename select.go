package lotusui

import (
	"image"
	"image/color"

	"gioui.org/font"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// Select is the select control: a bordered trigger showing the
// current choice with a chevron; clicking it opens a FLOATING panel
// of options over the content beneath (the shared portal primitive),
// with the selected row check-marked. Picking an option, pressing
// anywhere else, or Escape closes it.
// SelectOption is one choice: what the user READS (Label) and what
// the app STORES (Value) — HTML's `<option value="…">`.
//
// # The choice contract
//
// Every choice component in lotusui — Select, RadioGroup, Tabs,
// ToggleGroup — follows the same rules, with its OWN option type
// carrying only what that component renders (RadioOption adds a
// Description, TabOption an Icon, and so on):
//
//   - Val() is the stored value: Value, or Label when Value is empty
//     (HTML's rule).
//   - The component exposes Value() / SetValue(v) / Clear() /
//     Chosen(). The cursor is UNEXPORTED: an index is a fact about
//     one list's current order, so it must never reach app state,
//     storage or a wire. Reordering or rewording a list can then
//     never change what stored data means.
//   - The zero value selects the FIRST option, like a <select> with
//     no `selected` attribute.
//   - SetValue with an unknown value CLEARS the choice rather than
//     falling back to option 0, so a stored value that no longer
//     exists shows the placeholder instead of silently meaning
//     something else.
type SelectOption struct {
	Label string
	Value string
	// Meta is optional secondary text on the far right of the option
	// row (a count, shortcut, …). Empty omits it. The selected check
	// still sits after Meta.
	Meta string
	// Icon is a leading icon on the option row (and closed trigger
	// when Content is nil). Empty omits it.
	Icon string
	// Content replaces Icon+Label in the panel row and closed trigger
	// with arbitrary build-time widgets (multiline plan cards, …).
	// Label/Value still own the choice contract. Build Content when
	// the options list is built — never inside Layout.
	Content layout.Widget
}

// SelectItem is one label-only option — Value equals Label.
func SelectItem(label string) SelectOption { return SelectOption{Label: label} }

// SelectItemValue is one option with an explicit stored Value.
func SelectItemValue(value, label string) SelectOption {
	return SelectOption{Label: label, Value: value}
}

// SelectItems packs variadic options into a slice (build-time composition).
func SelectItems(items ...SelectOption) []SelectOption {
	out := make([]SelectOption, len(items))
	copy(out, items)
	return out
}

// SelectGrouped is one labeled group of options (ctor — the type is SelectGroup).
func SelectGrouped(label string, items ...SelectOption) SelectGroup {
	return SelectGroup{Label: label, Options: SelectItems(items...)}
}

// SelectGroups packs variadic groups into a slice.
func SelectGroups(groups ...SelectGroup) []SelectGroup {
	out := make([]SelectGroup, len(groups))
	copy(out, groups)
	return out
}

// SelectOpts builds label-only options, where each option's value IS
// its label — the shorthand for lists whose text is already the datum
// ("10", "25", "us-east-1"). Give explicit Values whenever the label
// is prose the product might reword. RadioOpts, TabOpts and
// ToggleOpts are the same shorthand for the other components.
func SelectOpts(labels ...string) []SelectOption {
	out := make([]SelectOption, len(labels))
	for i, l := range labels {
		out[i] = SelectItem(l)
	}
	return out
}

// Val is an option's stored value — its Label when Value is empty.
func (o SelectOption) Val() string {
	if o.Value != "" {
		return o.Value
	}
	return o.Label
}

// option is what every component's option type satisfies: the choice
// contract is shared as BEHAVIOUR, not as one struct, so each
// component's list carries only the fields it renders.
type option interface{ Val() string }

// chooseValue is the shared cursor move: the index of the option
// carrying v, or -1 when none does (the "unknown value clears" rule).
func chooseValue[T option](opts []T, v string) int {
	for i, o := range opts {
		if o.Val() == v {
			return i
		}
	}
	return -1
}

// valueAt reads the value at a cursor, "" when it points nowhere.
func valueAt[T option](opts []T, i int) string {
	if i < 0 || i >= len(opts) {
		return ""
	}
	return opts[i].Val()
}

// SelectGroup labels a run of options — rendered as a muted group
// label above its options, with a separator between groups.
type SelectGroup struct {
	Label   string
	Options []SelectOption
}

type Select struct {
	Options []SelectOption
	// Groups, when set, wins over Options: options render under muted
	// group labels with separators between groups, flattened in order.
	Groups []SelectGroup
	// Size is the shared size preset for the trigger frame.
	Size Size
	// Placeholder shows (in disabled ink) while nothing is chosen —
	// the "Choose one…" state Clear() puts a form in.
	Placeholder string
	// Disabled freezes the control on its current choice: no pointer
	// cursor, no opening, dimmed value — for identity fields that
	// cannot change after creation.
	Disabled bool
	// Invalid renders the trigger in danger chrome — pair it with a
	// Field error message.
	Invalid bool
	// Attached marks sides that sit against a neighbor in a
	// ButtonGroup: those corners render square and the seat shadow
	// drops.
	Attached AttachedEdges
	// AlignItemWithTrigger positions the OPEN panel so the selected
	// row sits directly over the trigger (the native-select feel),
	// instead of dropping below it. Long lists still scroll; the
	// alignment accounts for the scrolled-away rows.
	AlignItemWithTrigger bool
	// sel is the cursor into the flattened options — UNEXPORTED on
	// purpose: an index is a fact about THIS list's current order, so
	// it must never reach app state, storage or a wire. Callers speak
	// Value()/SetValue(); the zero value selects the first option,
	// exactly like a <select> with no `selected` attribute.
	sel     int
	open    bool
	btn     widget.Clickable
	optBtns []widget.Clickable
	list    widget.List
	dismiss dismisser
	sites   layoutSites
}

// Value is the chosen option's value ("" when nothing is chosen).
func (d *Select) Value() string { return valueAt(d.flatOptions(), d.sel) }

// SetValue chooses the option carrying v. An unknown value clears the
// choice, so a stored value that no longer exists shows the
// placeholder instead of silently meaning something else.
func (d *Select) SetValue(v string) { d.sel = chooseValue(d.flatOptions(), v) }

// Clear drops the choice back to the placeholder state.
func (d *Select) Clear() { d.sel = -1 }

// Chosen reports whether an option is currently selected.
func (d *Select) Chosen() bool { return d.Value() != "" }

// maxVisibleOptions bounds the panel: longer option lists scroll.
const maxVisibleOptions = 7

// flatOptions is the option list the panel renders and Selected
// indexes — Groups flattened in order, or Options.
func (d *Select) flatOptions() []SelectOption {
	if len(d.Groups) == 0 {
		return d.Options
	}
	var out []SelectOption
	for _, g := range d.Groups {
		out = append(out, g.Options...)
	}
	return out
}

// panelRow is one row of the open panel: an option (flat index), a
// group label, or a separator.
type panelRow struct {
	kind int // 0 option, 1 label, 2 separator
	text string
	meta string
	idx  int
}

func (d *Select) panelRows() []panelRow {
	if len(d.Groups) == 0 {
		rows := make([]panelRow, len(d.Options))
		for i, o := range d.Options {
			rows[i] = panelRow{kind: 0, text: o.Label, meta: o.Meta, idx: i}
		}
		return rows
	}
	var rows []panelRow
	flat := 0
	for gi, g := range d.Groups {
		if gi > 0 {
			rows = append(rows, panelRow{kind: 2})
		}
		if g.Label != "" {
			rows = append(rows, panelRow{kind: 1, text: g.Label})
		}
		for _, o := range g.Options {
			rows = append(rows, panelRow{kind: 0, text: o.Label, meta: o.Meta, idx: flat})
			flat++
		}
	}
	return rows
}

func (d *Select) Layout(th *Theme, gtx layout.Context, label string) layout.Dimensions {
	idx := d.sites.next(gtx.Now)
	if d.Disabled {
		d.open = false
	}
	if idx == 0 && d.open && d.dismiss.Dismissed(gtx) {
		d.open = false
	}
	if d.btn.Clicked(gtx) && !d.Disabled {
		d.open = !d.open
		if d.open {
			// Open aligned to the selection: scroll the panel so the
			// selected row is in view, the way a native select opens
			// at its current value.
			if first := d.sel - 2; first > 0 {
				d.list.Position.First = first
			} else {
				d.list.Position.First = 0
			}
			d.list.Position.Offset = 0
		}
	}
	opts := d.flatOptions()
	if len(d.optBtns) != len(opts) {
		d.optBtns = make([]widget.Clickable, len(opts))
	}
	for i := range d.optBtns {
		if d.optBtns[i].Clicked(gtx) {
			d.sel = i
			d.open = false
		}
	}

	trigger := func(gtx layout.Context) layout.Dimensions {
		dims := d.btn.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return inputFrame(th, gtx, InputOutline, d.Size, d.Invalid, d.Attached, func(gtx layout.Context) layout.Dimensions {
				defer clip.Rect(image.Rectangle{Max: gtx.Constraints.Max}).Push(gtx.Ops).Pop()
				if !d.Disabled {
					pointer.CursorPointer.Add(gtx.Ops)
				}
				ratio, _, _ := inputMetrics(d.Size)
				var chosen *SelectOption
				txt, ink, meta := d.Placeholder, th.Palette.FgDisabled, ""
				if opts := d.flatOptions(); d.sel >= 0 && d.sel < len(opts) {
					chosen = &opts[d.sel]
					txt, ink, meta = chosen.Label, th.Palette.Fg, chosen.Meta
				}
				if d.Disabled {
					ink = th.Palette.FgDisabled
				}
				return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						if chosen != nil && chosen.Content != nil {
							return chosen.Content(gtx)
						}
						return d.layoutOptionLabel(th, gtx, ratio, ink, txt, chosen, false)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if meta == "" || (chosen != nil && chosen.Content != nil) {
							return layout.Dimensions{}
						}
						mInk := th.Palette.FgSubtle
						if d.Disabled {
							mInk = th.Palette.FgDisabled
						}
						// Right gap before the chevron — Meta must not
						// sit flush against the trailing icon.
						return layout.Inset{Left: unit.Dp(8), Right: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							l := material.Label(th.Material, Sp(th, ratio), meta)
							l.Color = mInk
							l.MaxLines = 1
							return l.Layout(gtx)
						})
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if d.Disabled {
							return layout.Dimensions{}
						}
						return SVGIcon(IconChevronDown, 16, th.Palette.FgSubtle)(gtx)
					}),
				)
			})
		})
		// One Select, one panel — multi-Layout must not stack panels.
		if d.open && idx == 0 {
			d.layoutPanel(th, gtx, dims.Size)
		}
		return dims
	}
	if label == "" {
		return trigger(gtx)
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(SectionLabel(th, label)),
		layout.Rigid(Spacer(th.Space.XS)),
		layout.Rigid(trigger),
	)
}

// layoutOptionLabel is Icon+Label (or Label alone) for the string path.
func (d *Select) layoutOptionLabel(th *Theme, gtx layout.Context, ratio float32, ink color.NRGBA, text string, o *SelectOption, selected bool) layout.Dimensions {
	icon := ""
	if o != nil {
		icon = o.Icon
	}
	if icon == "" {
		l := material.Label(th.Material, Sp(th, ratio), text)
		l.Color = ink
		l.MaxLines = 1
		if selected {
			l.Font.Weight = font.Medium
		}
		return l.Layout(gtx)
	}
	iconSz := unit.Dp(float32(gtx.Metric.SpToDp(Sp(th, ratio))) + 2)
	return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(SVGIcon(icon, iconSz, ink)),
		layout.Rigid(HSpacer(th.Space.SM)),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			l := material.Label(th.Material, Sp(th, ratio), text)
			l.Color = ink
			l.MaxLines = 1
			if selected {
				l.Font.Weight = font.Medium
			}
			return l.Layout(gtx)
		}),
	)
}

// layoutPanel paints the floating options panel: anchored under the
// trigger, trigger-wide, capped by a dp budget (~maxVisibleOptions
// default rows), then scrolling. Variable-height Content rows share
// that budget — fewer tall rows fit, which is correct.
func (d *Select) layoutPanel(th *Theme, gtx layout.Context, trigger image.Point) {
	Floating(gtx, func(gtx layout.Context) layout.Dimensions {
		// The press-catcher goes in FIRST: everything the panel draws
		// after sits on top of it.
		d.dismiss.Add(gtx)

		// Vertical edge inset is Space.XS; the gap BETWEEN options is
		// half of that so adjacent hover pills don't read as double-spaced
		// against other UI gaps. Horizontal inset is NOT applied here —
		// HoverRow fills the panel width so the hover/selected pill reads
		// as a full-width row; its own SM padding keeps the label off the edges.
		_, vPad, _ := inputMetrics(d.Size)
		rowH := gtx.Dp(unit.Dp(22) + 2*vPad)
		edge := gtx.Dp(th.Space.XS)
		gap := gtx.Dp(th.Space.XS / 2)
		maxH := maxVisibleOptions*rowH + (maxVisibleOptions-1)*gap + 2*edge

		// Default anchor: 4dp below the trigger. Aligned mode instead
		// overlays the panel so the selected row's center matches the
		// trigger's center — uniform string rows only; Content rows have
		// variable height so we keep the drop-below anchor.
		offY := trigger.Y + gtx.Dp(4)
		if d.AlignItemWithTrigger && !d.hasContentOptions() {
			selRow := 0
			for i, r := range d.panelRows() {
				if r.kind == 0 && r.idx == d.sel {
					selRow = i
					break
				}
			}
			visible := selRow - d.list.Position.First
			if visible < 0 {
				visible = 0
			}
			if visible > maxVisibleOptions-1 {
				visible = maxVisibleOptions - 1
			}
			offY = trigger.Y/2 - (edge + visible*(rowH+gap) + rowH/2)
		}
		defer op.Offset(image.Pt(0, offY)).Push(gtx.Ops).Pop()

		gtx.Constraints = layout.Constraints{
			Min: image.Pt(trigger.X, 0),
			Max: image.Pt(trigger.X, maxH),
		}

		rows := d.panelRows()
		m := op.Record(gtx.Ops)
		dims := layout.Inset{Top: th.Space.XS, Bottom: th.Space.XS}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			d.list.Axis = layout.Vertical
			return material.List(th.Material, &d.list).Layout(gtx, len(rows), func(gtx layout.Context, i int) layout.Dimensions {
				if i == 0 {
					return d.layoutRow(th, gtx, rows[i])
				}
				return layout.Inset{Top: th.Space.XS / 2}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return d.layoutRow(th, gtx, rows[i])
				})
			})
		})
		call := m.Stop()

		size := image.Pt(trigger.X, dims.Size.Y)
		r := gtx.Dp(th.Radius.MD)
		cardShadow(gtx, size, r)
		defer clip.UniformRRect(image.Rectangle{Max: size}, r).Push(gtx.Ops).Pop()
		paint.Fill(gtx.Ops, th.Palette.BgPanel)
		widget.Border{Color: th.Palette.Border, Width: unit.Dp(1), CornerRadius: th.Radius.MD}.Layout(gtx,
			func(gtx layout.Context) layout.Dimensions { return layout.Dimensions{Size: size} })
		call.Add(gtx.Ops)
		return layout.Dimensions{Size: size}
	})
}

func (d *Select) hasContentOptions() bool {
	for _, o := range d.flatOptions() {
		if o.Content != nil {
			return true
		}
	}
	return false
}

// layoutRow renders one panel row: an option (hover pill + check on
// the selected one), a muted group label, or a separator.
func (d *Select) layoutRow(th *Theme, gtx layout.Context, row panelRow) layout.Dimensions {
	switch row.kind {
	case 1:
		return layout.Inset{Top: 6, Bottom: 2, Left: 8, Right: 8}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			l := LabelCaption(th, row.text)
			l.Color = th.Palette.FgSubtle
			return l.Layout(gtx)
		})
	case 2:
		return layout.Inset{Top: 4, Bottom: 4}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
			return Hairline(th)(gtx)
		})
	}
	return d.layoutOption(th, gtx, row.idx, row.text, row.meta)
}

// layoutOption renders one option row: Label/Icon/Content, optional Meta,
// the hover pill, and the check mark on the selected row.
func (d *Select) layoutOption(th *Theme, gtx layout.Context, i int, text, meta string) layout.Dimensions {
	ratio, _, _ := inputMetrics(d.Size)
	opts := d.flatOptions()
	var o *SelectOption
	if i >= 0 && i < len(opts) {
		o = &opts[i]
		text, meta = o.Label, o.Meta
	}
	return HoverRow(th, &d.optBtns[i], i == d.sel, func(gtx layout.Context) layout.Dimensions {
		ink := th.Palette.Fg
		return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				if o != nil && o.Content != nil {
					return o.Content(gtx)
				}
				return d.layoutOptionLabel(th, gtx, ratio, ink, text, o, i == d.sel)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if meta == "" || (o != nil && o.Content != nil) {
					return layout.Dimensions{}
				}
				// Right gap before the check / reserved slot — Meta
				// must not sit flush against the trailing icon.
				return layout.Inset{Left: unit.Dp(8), Right: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					l := material.Label(th.Material, Sp(th, ratio), meta)
					l.Color = th.Palette.FgSubtle
					l.MaxLines = 1
					return l.Layout(gtx)
				})
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				// The check sits on the RIGHT of the row, only on the
				// selected option.
				if i == d.sel {
					return SVGIcon(IconAccept, 16, ink)(gtx)
				}
				return layout.Dimensions{Size: image.Pt(gtx.Dp(16), 0)}
			}),
		)
	})(gtx)
}
