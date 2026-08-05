package lotusui

import (
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// Textarea is the multi-line sibling of Input: the same frame chrome
// and size presets, a wrapping editor inside, with a minimum height
// in rows. Behavior mechanisms mirror Input: Error for danger chrome,
// Disabled for read-only.
type Textarea struct {
	Editor   widget.Editor
	Size     Size
	Variant  InputVariant
	Error    string
	Disabled bool
	// Rows is the minimum visible line count; zero means 3.
	Rows int
}

// LayoutField renders the bare textarea; wrap it in Field for label,
// helper and error structure.
func (t *Textarea) LayoutField(th *Theme, gtx layout.Context, hint string) layout.Dimensions {
	t.Editor.SingleLine = false
	t.Editor.ReadOnly = t.Disabled
	if t.Disabled {
		gtx = gtx.Disabled()
	}
	rows := t.Rows
	if rows == 0 {
		rows = 3
	}
	ratio, _, _ := inputMetrics(t.Size)
	return inputFrame(th, gtx, t.Variant, t.Size, t.Error != "", AttachedEdges{}, func(gtx layout.Context) layout.Dimensions {
		lineH := gtx.Sp(Sp(th, ratio)) * 13 / 10
		gtx.Constraints.Min.Y = rows * lineH
		ed := material.Editor(th.Material, &t.Editor, hint)
		ed.TextSize = Sp(th, ratio)
		ed.Color = th.Palette.Fg
		if t.Disabled {
			ed.Color = th.Palette.FgDisabled
		}
		ed.HintColor = th.Palette.FgDisabled
		return ed.Layout(gtx)
	})
}

// Layout is LayoutField wrapped in a Field with label and the
// textarea's own error.
func (t *Textarea) Layout(th *Theme, gtx layout.Context, label, hint string) layout.Dimensions {
	return Field(th, FieldProps{Label: label, Error: t.Error}, func(gtx layout.Context) layout.Dimensions {
		return t.LayoutField(th, gtx, hint)
	})(gtx)
}
