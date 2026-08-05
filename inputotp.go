package lotusui

import (
	"image"
	"image/color"
	"strings"

	"gioui.org/font"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// InputOTP is the one-time-code input: character slots rendered as
// attached boxes driven by ONE hidden editor — click anywhere on the
// row, type or paste, and the code fills left to right. The active
// slot (where the next character lands) carries the focus ring while
// the group is focused.
type InputOTP struct {
	// Length is the code length; zero means 6.
	Length int
	// Groups splits the slots into separated groups (e.g. {3, 3} =
	// two groups of three with a dash between); nil = one group.
	Groups []int
	// Filter is the allow-list; empty accepts everything —
	// "0123456789" is the digits-only pattern.
	Filter   string
	Disabled bool
	Invalid  bool
	editor   widget.Editor
}

// Value returns the code typed so far.
func (o *InputOTP) Value() string { return o.editor.Text() }

// SetValue replaces the code.
func (o *InputOTP) SetValue(s string) { o.editor.SetText(s) }

func (o *InputOTP) length() int {
	if o.Length > 0 {
		return o.Length
	}
	return 6
}

func (o *InputOTP) Layout(th *Theme, gtx layout.Context) layout.Dimensions {
	n := o.length()
	o.editor.SingleLine = true
	o.editor.Filter = o.Filter
	o.editor.MaxLen = n
	o.editor.ReadOnly = o.Disabled
	// A digits-only code asks phones for the number pad — Gio's own
	// input hint, so the on-screen keyboard is right on mobile.
	o.editor.InputHint = key.HintAny
	if o.Filter != "" && strings.Trim(o.Filter, "0123456789") == "" {
		o.editor.InputHint = key.HintNumeric
	}

	slotW, slotH := gtx.Dp(38), gtx.Dp(42)
	sepGap := gtx.Dp(18)
	overlap := gtx.Dp(1)

	groups := o.Groups
	if len(groups) == 0 {
		groups = []int{n}
	}

	// Geometry first: slot origins, then total size.
	type box struct {
		at          image.Point
		first, last bool
	}
	boxes := make([]box, 0, n)
	x, idx := 0, 0
	for gi, g := range groups {
		for j := 0; j < g && idx < n; j++ {
			boxes = append(boxes, box{at: image.Pt(x, 0), first: j == 0, last: j == g-1 || idx == n-1})
			x += slotW - overlap
			idx++
		}
		x += overlap // the group's last slot keeps its full width
		if gi < len(groups)-1 {
			x += sepGap
		}
	}
	total := image.Pt(x, slotH)
	// Hostile constraints: never paint or report beyond Max.
	if total.X > gtx.Constraints.Max.X {
		total.X = gtx.Constraints.Max.X
	}
	if total.Y > gtx.Constraints.Max.Y {
		total.Y = gtx.Constraints.Max.Y
	}
	defer clip.Rect(image.Rectangle{Max: total}).Push(gtx.Ops).Pop()

	// EVENTS FIRST, PAINT LAST. Gio commits a keystroke inside the
	// editor's own layout, so anything read before that call — the
	// text, the focus — describes the PREVIOUS frame, and the typed
	// character would surface one frame late. The editor is invisible
	// (its ink and selection are transparent; the slots below are the
	// visible form), so it is laid out here, into a macro, and added
	// at the end where its input area sits on top of the row.
	var edCall op.CallOp
	hasEditor := !o.Disabled
	if hasEditor {
		m := op.Record(gtx.Ops)
		func() {
			defer clip.Rect(image.Rectangle{Max: total}).Push(gtx.Ops).Pop()
			cgtx := gtx
			cgtx.Constraints = layout.Constraints{Min: total, Max: total}
			ed := material.Editor(th.Material, &o.editor, "")
			ed.TextSize = Sp(th, 1)
			ed.Color = color.NRGBA{}
			ed.SelectionColor = color.NRGBA{}
			ed.Layout(cgtx)
		}()
		edCall = m.Stop()
	}

	// Now the state is THIS frame's. The ring follows the editor's own
	// caret (rune indices from Selection), so arrow keys and clicks
	// move it exactly as they move the real insertion point.
	chars := []rune(o.editor.Text())
	if len(chars) > n {
		chars = chars[:n]
	}
	_, caret := o.editor.Selection()
	if caret > len(chars) {
		caret = len(chars)
	}
	if caret >= n {
		caret = n - 1
	}
	focused := gtx.Focused(&o.editor) && !o.Disabled

	// The slots: attached boxes, outer corners rounded per group.
	danger := th.Palette.DangerScheme().Solid
	r := gtx.Dp(th.Radius.SM + 2)
	// The focus ring is remembered here and drawn AFTER the loop:
	// neighbors overlap by 1dp to collapse the shared border, so a
	// ring painted in-loop loses a sliver to the NEXT slot's fill and
	// reads thinner on its junction edge.
	var ringAt image.Point
	var ringRR clip.RRect
	drawRing := false
	for i, b := range boxes {
		off := op.Offset(b.at).Push(gtx.Ops)
		rr := clip.RRect{Rect: image.Rectangle{Max: image.Pt(slotW, slotH)}}
		if b.first {
			rr.NW, rr.SW = r, r
		}
		if b.last {
			rr.NE, rr.SE = r, r
		}
		func() {
			defer rr.Push(gtx.Ops).Pop()
			paint.Fill(gtx.Ops, th.Palette.BgPanel)
			borderCol := th.Palette.Border
			if o.Invalid {
				borderCol = danger
			}
			paint.FillShape(gtx.Ops, borderCol, clip.Stroke{Path: rr.Path(gtx.Ops), Width: float32(gtx.Dp(1)) * 2}.Op())
			// The character, centered.
			if i < len(chars) {
				cgtx := gtx
				cgtx.Constraints = layout.Constraints{Min: image.Pt(slotW, slotH), Max: image.Pt(slotW, slotH)}
				layout.Center.Layout(cgtx, func(gtx layout.Context) layout.Dimensions {
					l := material.Label(th.Material, Sp(th, 1), string(chars[i]))
					l.Color = th.Palette.Fg
					if o.Disabled {
						l.Color = th.Palette.FgDisabled
					}
					l.Font.Weight = font.Medium
					return l.Layout(gtx)
				})
			}
		}()
		if focused && i == caret {
			ringAt, ringRR, drawRing = b.at, rr, true
		}
		off.Pop()
	}
	// The active slot's ring, over every slot: it strokes the SAME
	// rounded shape the slot clips to, at full weight on all four
	// sides — including the edges neighbors overlap.
	if drawRing {
		off := op.Offset(ringAt).Push(gtx.Ops)
		func() {
			defer ringRR.Push(gtx.Ops).Pop()
			paint.FillShape(gtx.Ops, th.Palette.FocusRing,
				clip.Stroke{Path: ringRR.Path(gtx.Ops), Width: float32(gtx.Dp(2)) * 2}.Op())
		}()
		off.Pop()
	}

	// Group separators: a small dash centered in each gap.
	x, idx = 0, 0
	for gi, g := range groups {
		x += g*(slotW-overlap) + overlap
		if gi < len(groups)-1 {
			dw, dh := gtx.Dp(8), gtx.Dp(2)
			off := op.Offset(image.Pt(x+(sepGap-dw)/2, (slotH-dh)/2)).Push(gtx.Ops)
			fgtx := gtx
			fgtx.Constraints = layout.Constraints{Min: image.Pt(dw, dh), Max: image.Pt(dw, dh)}
			Fill(fgtx, th.Palette.BgEmphasized)
			off.Pop()
			x += sepGap
		}
		idx += g
	}

	// The invisible editor, laid out above: adding it here puts its
	// input area over the whole row, so a click anywhere focuses it
	// and typing or pasting fills the slots.
	if hasEditor {
		edCall.Add(gtx.Ops)
	}

	return layout.Dimensions{Size: total}
}
