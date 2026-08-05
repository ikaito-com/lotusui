package lotusui

import (
	"image"

	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
)

// RadioOption is one radio choice — the choice contract lives with
// SelectOption; Description renders muted under the label and
// Disabled makes this one option unclickable (both replacing the
// index-aligned slices RadioGroup used to carry).
type RadioOption struct {
	Label       string
	Value       string
	Description string
	Disabled    bool
}

// Val is the option's stored value — its Label when Value is empty.
func (o RadioOption) Val() string {
	if o.Value != "" {
		return o.Value
	}
	return o.Label
}

// RadioItem is one label-only radio option — Value equals Label.
func RadioItem(label string) RadioOption { return RadioOption{Label: label} }

// RadioItemValue is one radio option with an explicit stored Value.
func RadioItemValue(value, label string) RadioOption {
	return RadioOption{Label: label, Value: value}
}

// RadioItems packs variadic options into a slice (build-time composition).
func RadioItems(items ...RadioOption) []RadioOption {
	out := make([]RadioOption, len(items))
	copy(out, items)
	return out
}

// RadioOpts builds label-only radio options — the shorthand when the
// label IS the datum.
func RadioOpts(labels ...string) []RadioOption {
	out := make([]RadioOption, len(labels))
	for i, l := range labels {
		out[i] = RadioItem(l)
	}
	return out
}

// RadioGroup is the exclusive choice: labeled circles stacked
// vertically, exactly one selected. Options carry their own
// Description and Disabled flags, and the choice is read and written
// as a VALUE — the cursor is unexported on purpose, because an index
// is a fact about this list's current order and must never reach app
// state, storage or a wire.
//
//	plan := lotusui.RadioGroup{Options: []lotusui.RadioOption{
//		{Label: "Starter", Value: "starter", Description: "For side projects."},
//		{Label: "Pro", Value: "pro", Description: "For teams shipping daily."},
//	}}
//	plan.SetValue(stored) // an unknown value clears, never picks option 0
//	plan.Layout(th, gtx)
//	stored = plan.Value()
//
// Clicks are processed at the top of Layout, so the click's own frame
// renders the new selection.
type RadioGroup struct {
	Options []RadioOption
	Size    Size
	// Invalid renders danger chrome on every circle.
	Invalid bool
	// sel is the cursor into Options — UNEXPORTED: callers speak
	// Value()/SetValue(). The zero value selects the first option.
	sel  int
	btns []widget.Clickable
}

// Value is the chosen option's value ("" when nothing is chosen).
func (r *RadioGroup) Value() string { return valueAt(r.Options, r.sel) }

// SetValue chooses the option carrying v. An unknown value clears the
// choice, so a stored value that no longer exists leaves nothing
// selected instead of silently meaning something else.
func (r *RadioGroup) SetValue(v string) { r.sel = chooseValue(r.Options, v) }

// Clear drops the choice — nothing selected.
func (r *RadioGroup) Clear() { r.sel = -1 }

// Chosen reports whether an option is currently selected.
func (r *RadioGroup) Chosen() bool { return r.Value() != "" }

func (r *RadioGroup) circleDp(base int) unit.Dp {
	switch r.Size {
	case Size2XS:
		return unit.Dp(base - 4)
	case SizeXS:
		return unit.Dp(base - 3)
	case SizeSM:
		return unit.Dp(base - 2)
	case SizeLG:
		return unit.Dp(base + 3)
	case SizeXL:
		return unit.Dp(base + 5)
	case Size2XL:
		return unit.Dp(base + 7)
	}
	return unit.Dp(base)
}

func (r *RadioGroup) Layout(th *Theme, gtx layout.Context) layout.Dimensions {
	if len(r.btns) != len(r.Options) {
		r.btns = make([]widget.Clickable, len(r.Options))
	}
	for i := range r.btns {
		if r.btns[i].Clicked(gtx) && !r.Options[i].Disabled {
			r.sel = i
		}
	}
	var rows []layout.Widget
	for i, o := range r.Options {
		i, o := i, o
		rows = append(rows, func(gtx layout.Context) layout.Dimensions {
			return r.btns[i].Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				if !o.Disabled {
					pointer.CursorPointer.Add(gtx.Ops)
				}
				return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						d := gtx.Dp(r.circleDp(16))
						ring := th.Palette.BorderEmphasized
						if r.Invalid {
							ring = th.Palette.Danger
						}
						if o.Disabled {
							ring = th.Palette.Border
						}
						sz := image.Pt(d, d)
						defer clip.UniformRRect(image.Rectangle{Max: sz}, d/2).Push(gtx.Ops).Pop()
						// The ring is painted as two discs — a border
						// stroke degenerates when its corner radius
						// exceeds half the shape, so circles never use
						// widget.Border.
						paint.Fill(gtx.Ops, ring)
						bw := gtx.Dp(1)
						inner := image.Rect(bw, bw, d-bw, d-bw)
						paint.FillShape(gtx.Ops, th.Palette.BgPanel, clip.UniformRRect(inner, inner.Dx()/2).Op(gtx.Ops))
						if r.sel == i {
							// The dot: brand ink, centered, ~40% of the circle.
							dot := d * 2 / 5
							off := (d - dot) / 2
							dr := image.Rect(off, off, off+dot, off+dot)
							ink := th.Palette.BrandFg
							if o.Disabled {
								// Brighter same-hue mute — never darker/grey.
								ink = th.BrandScale.C300
							}
							paint.FillShape(gtx.Ops, ink, clip.UniformRRect(dr, dot/2).Op(gtx.Ops))
						}
						return layout.Dimensions{Size: sz}
					}),
					layout.Rigid(HSpacer(th.Space.SM)),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						l := LabelBody(th, o.Label)
						l.Color = th.Palette.Fg
						if o.Disabled {
							l.Color = th.Palette.FgDisabled
						}
						if o.Description == "" {
							return l.Layout(gtx)
						}
						d := LabelCaption(th, o.Description)
						d.Color = th.Palette.FgSubtle
						return VStack(2, l.Layout, d.Layout)(gtx)
					}),
				)
			})
		})
	}
	return VStack(th.Space.SM, rows...)(gtx)
}
