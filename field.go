package lotusui

import (
	"gioui.org/layout"
	"gioui.org/unit"
)

// Field is the form-field wrapper: label, helper text, error message
// and required marker around ANY control — an Input, a Dropdown, a
// Checkbox group, your own widget. It is pure structure (the
// lean-core rule: labels and messages are composition, not props of
// every control), so each part costs nothing when unused.
//
//	lotusui.Field(th, lotusui.FieldProps{
//		Label:    "Email",
//		Helper:   "We'll never share it.",
//		Required: true,
//	}, func(gtx C) D { return email.LayoutField(th, gtx, "you@example.com") })
type FieldProps struct {
	Label    string
	Helper   string // quiet guidance under the control
	Error    string // replaces Helper in danger ink when non-empty
	Required bool   // marks the label with a danger asterisk
}

// Field lays out the wrapper around control. When Error is set it
// replaces the helper line — one message slot, the error wins.
func Field(th *Theme, o FieldProps, control layout.Widget) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		var children []layout.FlexChild
		if o.Label != "" {
			label := func(gtx layout.Context) layout.Dimensions {
				if !o.Required {
					return SectionLabel(th, o.Label)(gtx)
				}
				return HStack(2,
					SectionLabel(th, o.Label),
					func(gtx layout.Context) layout.Dimensions {
						l := LabelCaption(th, "*")
						l.Color = th.Palette.Danger
						return l.Layout(gtx)
					},
				)(gtx)
			}
			children = append(children,
				layout.Rigid(label),
				layout.Rigid(Spacer(th.Space.XS)),
			)
		}
		children = append(children, layout.Rigid(control))
		switch {
		case o.Error != "":
			children = append(children,
				layout.Rigid(Spacer(th.Space.XS)),
				layout.Rigid(FieldError(th, o.Error)),
			)
		case o.Helper != "":
			children = append(children,
				layout.Rigid(Spacer(th.Space.XS)),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return LabelCaption(th, o.Helper).Layout(gtx)
				}),
			)
		}
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
	}
}

// FieldGroup stacks Field (or any) widgets with a vertical gap —
// shadcn's FieldGroup transposed to build-time composition. Zero gap
// uses Space.MD.
func FieldGroup(th *Theme, gap unit.Dp, fields ...layout.Widget) layout.Widget {
	if gap == 0 {
		gap = th.Space.MD
	}
	return VStack(gap, fields...)
}
