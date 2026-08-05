package lotusui

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// InputVariant selects the field chrome.
type InputVariant int

const (
	InputOutline InputVariant = iota // bordered panel field (default)
	InputSubtle                      // filled with BgSubtle, borderless
	InputFlushed                     // underline only, no fill, no radius
)

// inputMetrics is the shared size preset: editor sp ratio and the
// frame's paddings move together, like Button's.
func inputMetrics(sz Size) (ratio float32, v, h unit.Dp) {
	switch sz {
	case Size2XS:
		return 10.0 / 16.0, 4, 6
	case SizeXS:
		return 12.0 / 16.0, 5, 8
	case SizeSM:
		return 13.0 / 16.0, 6, 10
	case SizeLG:
		return 15.0 / 16.0, 10, 14
	case SizeXL:
		return 16.0 / 16.0, 12, 16
	case Size2XL:
		return 18.0 / 16.0, 14, 18
	}
	return 14.0 / 16.0, 8, 12
}

// inputFrame draws the text-input chrome for a variant: outline is
// the bordered panel field, subtle a borderless fill, flushed a bare
// underline. invalid always wins with danger ink on whatever edge the
// variant draws.
func inputFrame(th *Theme, gtx layout.Context, variant InputVariant, sz Size, invalid bool, content layout.Widget) layout.Dimensions {
	_, vPad, hPad := inputMetrics(sz)
	if variant == InputFlushed {
		hPad = 0
	}
	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	m := op.Record(gtx.Ops)
	dims := layout.Inset{Top: vPad, Bottom: vPad, Left: hPad, Right: hPad}.Layout(gtx, content)
	call := m.Stop()
	dims.Size.X = gtx.Constraints.Max.X
	danger := th.Palette.DangerScheme().Solid
	switch variant {
	case InputSubtle:
		r := ClampCorner(gtx.Dp(th.Radius.SM+2), dims.Size)
		defer clip.UniformRRect(image.Rectangle{Max: dims.Size}, r).Push(gtx.Ops).Pop()
		paint.Fill(gtx.Ops, th.Palette.BgSubtle)
		if invalid {
			widget.Border{Color: danger, Width: unit.Dp(1), CornerRadius: th.Radius.SM + 2}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Dimensions{Size: dims.Size}
			})
		}
	case InputFlushed:
		lineCol := th.Palette.Border
		if invalid {
			lineCol = danger
		}
		lh := gtx.Dp(unit.Dp(1))
		st := op.Offset(image.Pt(0, dims.Size.Y-lh)).Push(gtx.Ops)
		fillRect(gtx, image.Pt(dims.Size.X, lh), lineCol)
		st.Pop()
	default: // InputOutline
		r := ClampCorner(gtx.Dp(th.Radius.SM+2), dims.Size)
		seatShadow(gtx, dims.Size, r)
		defer clip.UniformRRect(image.Rectangle{Max: dims.Size}, r).Push(gtx.Ops).Pop()
		paint.Fill(gtx.Ops, th.Palette.BgPanel)
		border := th.Palette.Border
		if invalid {
			border = danger
		}
		widget.Border{Color: border, Width: unit.Dp(1), CornerRadius: th.Radius.SM + 2}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Dimensions{Size: dims.Size}
		})
	}
	call.Add(gtx.Ops)
	return dims
}

// InputFrameErr draws the default outline field chrome around any
// content — for composite fields that manage their own editor.
func InputFrameErr(th *Theme, gtx layout.Context, invalid bool, content layout.Widget) layout.Dimensions {
	return inputFrame(th, gtx, InputOutline, SizeMD, invalid, content)
}

// Input is the single-line text input: quiet label above a bordered
// field with a hint. Domain rules stay in YOUR app — the component
// only offers the generic input-layer mechanisms:
//
//   - Filter is Gio's allow-list: characters outside it are never
//     inserted at all, the idiomatic way to reject input (no red
//     flash, no beep — the character simply doesn't appear).
//   - Transform rewrites the text in the same frame it changes (e.g.
//     strings.ToLower for lowercase-only identifiers) — a typed "A"
//     appears instantly as "a". When the rewrite preserves length,
//     the caret and selection survive unchanged.
//   - Error, set by your save-time validation, turns the border red
//     and renders the message below the field.
type Input struct {
	Editor    widget.Editor
	Variant   InputVariant        // Outline (default), Subtle, Flushed
	Size      Size                // the shared size presets
	Filter    string              // allow-list; "" accepts everything
	Transform func(string) string // same-frame rewrite; nil = none
	Error     string
	Disabled  bool // read-only, dimmed, no caret
	// Start and End render INSIDE the field, beside the editor — an
	// icon, a small button (a clear ✕, a visibility eye), any widget.
	// nil costs nothing: pure composition, the lean-core rule.
	Start, End layout.Widget
	// Top and Bottom render as full-width rows INSIDE the frame,
	// above/below the editor line — the block addons (a header, a
	// status row with actions).
	Top, Bottom layout.Widget
}

// ApplyConstraints enforces the input-layer rules (allow-list +
// transform). Layout calls it automatically; call it yourself only
// when a composite field renders f.Editor directly inside its own
// frame.
func (f *Input) ApplyConstraints() {
	f.Editor.SingleLine = true
	f.Editor.Filter = f.Filter
	f.applyTransform()
}

// applyTransform rewrites the editor's text in place, preserving the
// caret when the rewrite preserves length.
func (f *Input) applyTransform() {
	if f.Transform == nil {
		return
	}
	t := f.Editor.Text()
	if nt := f.Transform(t); nt != t {
		start, end := f.Editor.Selection()
		f.Editor.SetText(nt)
		if len(nt) == len(t) {
			f.Editor.SetCaret(start, end)
		}
	}
}

// editorPass lays the editor out and rewrites its text in the SAME
// frame the character arrived. Gio commits a keystroke inside the
// editor's own layout, so a Transform applied before it always
// rewrites the PREVIOUS frame's text — a typed "A" would show for one
// frame before folding to "a". When the rewrite changes the text the
// recorded pass is discarded and re-laid out; that second pass only
// happens on frames that actually rewrite, and it processes no
// events (the first pass drained them).
func (f *Input) editorPass(ed material.EditorStyle) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		if f.Transform == nil {
			return ed.Layout(gtx)
		}
		m := op.Record(gtx.Ops)
		dims := ed.Layout(gtx)
		call := m.Stop()
		before := f.Editor.Text()
		f.applyTransform()
		if f.Editor.Text() != before {
			m = op.Record(gtx.Ops)
			dims = ed.Layout(gtx)
			call = m.Stop()
		}
		call.Add(gtx.Ops)
		return dims
	}
}

// LayoutField renders the BARE field — no label, no messages. Wrap it
// in Field for labels/helper/error structure, or use Layout for the
// common label+error shorthand.
func (f *Input) LayoutField(th *Theme, gtx layout.Context, hint string) layout.Dimensions {
	f.ApplyConstraints()
	f.Editor.ReadOnly = f.Disabled
	if f.Disabled {
		gtx = gtx.Disabled()
	}
	ratio, _, _ := inputMetrics(f.Size)
	line := func(gtx layout.Context) layout.Dimensions {
		ed := material.Editor(th.Material, &f.Editor, hint)
		ed.TextSize = Sp(th, ratio)
		ed.Color = th.Palette.Fg
		if f.Disabled {
			ed.Color = th.Palette.FgDisabled
		}
		ed.HintColor = th.Palette.FgDisabled
		row := []layout.FlexChild{}
		if f.Start != nil {
			row = append(row, layout.Rigid(f.Start), layout.Rigid(HSpacer(th.Space.SM)))
		}
		row = append(row, layout.Flexed(1, f.editorPass(ed)))
		if f.End != nil {
			row = append(row, layout.Rigid(HSpacer(th.Space.SM)), layout.Rigid(f.End))
		}
		return layout.Flex{Alignment: layout.Middle}.Layout(gtx, row...)
	}
	content := line
	if f.Top != nil || f.Bottom != nil {
		content = func(gtx layout.Context) layout.Dimensions {
			var rows []layout.Widget
			if f.Top != nil {
				rows = append(rows, f.Top)
			}
			rows = append(rows, line)
			if f.Bottom != nil {
				rows = append(rows, f.Bottom)
			}
			return VStack(th.Space.SM, rows...)(gtx)
		}
	}
	return inputFrame(th, gtx, f.Variant, f.Size, f.Error != "", content)
}

func (f *Input) Layout(th *Theme, gtx layout.Context, label, hint string) layout.Dimensions {
	return Field(th, FieldProps{Label: label, Error: f.Error}, func(gtx layout.Context) layout.Dimensions {
		return f.LayoutField(th, gtx, hint)
	})(gtx)
}

// LayoutSuffix is Layout with an optional NON-EDITABLE suffix segment
// beside the input (a fixed domain, a generated qualifier): two boxes
// side by side, the right one quiet and fixed — the base text is the
// user's, the suffix is your app's truth.
func (f *Input) LayoutSuffix(th *Theme, gtx layout.Context, label, hint, suffix string) layout.Dimensions {
	f.ApplyConstraints()
	f.Editor.ReadOnly = f.Disabled
	if f.Disabled {
		gtx = gtx.Disabled()
	}
	row := func(gtx layout.Context) layout.Dimensions {
		return f.LayoutField(th, gtx, hint)
	}
	if suffix != "" {
		// ONE input look: a single frame whose right segment is the
		// non-editable suffix — tinted, sharing the outer radius, the
		// junction seamless (no inner border, no gap).
		row = func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
			// Measure the suffix at its NATURAL size — with the row's
			// stretched Min it would report the full row width and its
			// segment would cover the editor.
			lgtx := gtx
			lgtx.Constraints.Min = image.Point{}
			lm := op.Record(gtx.Ops)
			sl := LabelBody(th, suffix)
			sl.Color = th.Palette.FgSubtle
			lDims := layout.Inset{Left: unit.Dp(10), Right: unit.Dp(12)}.Layout(lgtx, sl.Layout)
			lCall := lm.Stop()

			m := op.Record(gtx.Ops)
			dims := layout.Inset{Top: unit.Dp(10), Bottom: unit.Dp(10), Left: unit.Dp(12)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				if gtx.Constraints.Max.X > lDims.Size.X {
					gtx.Constraints.Max.X -= lDims.Size.X
				}
				gtx.Constraints.Min.X = gtx.Constraints.Max.X
				ed := material.Editor(th.Material, &f.Editor, hint)
				ed.TextSize = Sp(th, 14.0/16.0)
				ed.Color = th.Palette.Fg
				ed.HintColor = th.Palette.FgDisabled
				return ed.Layout(gtx)
			})
			call := m.Stop()
			dims.Size.X = gtx.Constraints.Max.X
			r := gtx.Dp(th.Radius.SM + 2)
			defer clip.UniformRRect(image.Rectangle{Max: dims.Size}, r).Push(gtx.Ops).Pop()
			paint.Fill(gtx.Ops, th.Palette.BgPanel)
			seg := image.Rect(dims.Size.X-lDims.Size.X, 0, dims.Size.X, dims.Size.Y)
			paint.FillShape(gtx.Ops, th.Palette.BgSubtle, clip.RRect{Rect: seg, NE: r, SE: r}.Op(gtx.Ops))
			borderCol := th.Palette.Border
			if f.Error != "" {
				borderCol = th.Palette.DangerScheme().Solid
			}
			widget.Border{Color: borderCol, Width: unit.Dp(1), CornerRadius: th.Radius.SM + 2}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Dimensions{Size: dims.Size}
			})
			call.Add(gtx.Ops)
			off := op.Offset(image.Pt(seg.Min.X, (dims.Size.Y-lDims.Size.Y)/2)).Push(gtx.Ops)
			lCall.Add(gtx.Ops)
			off.Pop()
			return dims
		}
	}
	return Field(th, FieldProps{Label: label, Error: f.Error}, row)(gtx)
}

// FieldError is the one validation-message look: quiet, danger ink,
// below the field it belongs to.
func FieldError(th *Theme, msg string) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		l := LabelMeta(th, msg)
		l.Color = th.Palette.DangerScheme().Solid
		return l.Layout(gtx)
	}
}
