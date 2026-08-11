package lotusui

import (
	"image"

	"gioui.org/f32"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
)

// Checkbox is the checkbox: a rounded box with a check
// mark when on, a label beside it. Disabled renders quietly and
// swallows clicks — for states that are facts, not choices (a
// permission the plan doesn't include). Callers read Clicked and set Value.
type Checkbox struct {
	Value bool
	// Indeterminate renders a dash instead of a check — the classic
	// parent of a partially-selected list. It is a display state; the
	// caller decides what a click on it means.
	Indeterminate bool
	// Invalid turns the unchecked border danger red (save-time
	// validation, same contract as Input.Error).
	Invalid  bool
	Disabled bool
	Size     Size
	btn      widget.Clickable
}

// boxDp maps the shared size presets onto the check square.
func (c *Checkbox) boxDp() unit.Dp {
	switch c.Size {
	case Size2XS:
		return 11
	case SizeXS:
		return 13
	case SizeSM:
		return 14
	case SizeLG:
		return 20
	case SizeXL:
		return 22
	case Size2XL:
		return 24
	}
	return 16
}

// Clicked reports a click this frame — always false while Disabled.
func (c *Checkbox) Clicked(gtx layout.Context) bool {
	clicked := c.btn.Clicked(gtx)
	return clicked && !c.Disabled
}

func (c *Checkbox) Layout(th *Theme, gtx layout.Context, label string) layout.Dimensions {
	return c.btn.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				sz := gtx.Dp(c.boxDp())
				box := image.Rectangle{Max: image.Pt(sz, sz)}
				r := gtx.Dp(th.Radius.XS)
				defer clip.UniformRRect(box, r).Push(gtx.Ops).Pop()
				if !c.Disabled {
					pointer.CursorPointer.Add(gtx.Ops)
				}
				fill := th.Palette.BgPanel
				border := th.Palette.Border
				mark := th.Palette.BgPanel
				if c.Invalid && !c.Value && !c.Indeterminate {
					border = th.Palette.Danger
				}
				if c.Value || c.Indeterminate {
					fill = th.Palette.BrandFg
					border = th.Palette.BrandFg
					if c.Disabled {
						// Brighter same-hue mute — BrandFg → .200 with
						// mid mark ink, never a darker/grey step.
						fill = th.BrandScale.C200
						border = fill
						mark = th.BrandScale.C600
					}
				} else if c.Disabled {
					border = th.Palette.BorderMuted
				}
				paint.Fill(gtx.Ops, fill)
				widget.Border{Color: border, Width: unit.Dp(1), CornerRadius: th.Radius.XS}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Dimensions{Size: box.Max}
				})
				if c.Indeterminate {
					// The dash: a centered bar, drawn (not typed) so it
					// scales with the box.
					bw, bh := sz*10/18, gtx.Dp(unit.Dp(2))
					bar := image.Rect((sz-bw)/2, (sz-bh)/2, (sz+bw)/2, (sz+bh)/2)
					paint.FillShape(gtx.Ops, mark, clip.Rect(bar).Op())
				} else if c.Value {
					// The check is DRAWN, not typed: a ✓ glyph depends on
					// font fallback, which wasm builds don't have — a path
					// renders identically on every target and scales with
					// the box.
					sw := float32(gtx.Dp(2))
					f := float32(sz)
					var path clip.Path
					path.Begin(gtx.Ops)
					path.MoveTo(f32.Pt(f*0.24, f*0.54))
					path.LineTo(f32.Pt(f*0.42, f*0.72))
					path.LineTo(f32.Pt(f*0.76, f*0.30))
					paint.FillShape(gtx.Ops, mark,
						clip.Stroke{Path: path.End(), Width: sw}.Op())
				}
				return layout.Dimensions{Size: box.Max}
			}),
			layout.Rigid(HSpacer(th.Space.SM)),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				l := LabelBody(th, label)
				l.Color = th.Palette.Fg
				if c.Disabled {
					l.Color = th.Palette.FgDisabled
				}
				if !c.Disabled {
					dims := l.Layout(gtx)
					defer clip.Rect(image.Rectangle{Max: dims.Size}).Push(gtx.Ops).Pop()
					pointer.CursorPointer.Add(gtx.Ops)
					return dims
				}
				return l.Layout(gtx)
			}),
		)
	})
}
