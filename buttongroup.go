package lotusui

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/unit"
	"gioui.org/widget"
)

// AttachedEdges marks the sides of a control that sit flush against a
// neighbor — those corners render square and the shared border
// collapses. ButtonGroup sets it on its children; set it by hand only
// when hand-building attached rows.
type AttachedEdges struct {
	Start, End, Top, Bottom bool
}

// attachedRRect builds a rounded rect with attached sides squared —
// the shared chrome for Button and Input inside a ButtonGroup.
func attachedRRect(sz image.Point, r int, a AttachedEdges) clip.RRect {
	rr := clip.RRect{Rect: image.Rectangle{Max: sz}, NW: r, NE: r, SE: r, SW: r}
	if a.Start {
		rr.NW, rr.SW = 0, 0
	}
	if a.End {
		rr.NE, rr.SE = 0, 0
	}
	if a.Top {
		rr.NW, rr.NE = 0, 0
	}
	if a.Bottom {
		rr.SW, rr.SE = 0, 0
	}
	return rr
}

// ButtonGroupProps are the group's options.
type ButtonGroupProps struct {
	// Vertical stacks the group top-to-bottom.
	Vertical bool
}

// ButtonGroupItem is one slot of the group: a button (Btn + Label +
// Props), an Input (fused chrome via Attached), a Separator, or an
// arbitrary Widget. Flex gives the slot flexed weight.
type ButtonGroupItem struct {
	Btn   *widget.Clickable
	Label string
	Props ButtonProps
	// Separator renders the hairline divider between attached
	// neighbors — the split-button seam.
	Separator bool
	// Input, when set, is laid out as a field slot; the group applies
	// AttachedEdges so the field fuses with neighboring buttons
	// (shadcn ButtonGroup + Input). Prefer this over Widget for inputs.
	Input *Input
	Hint  string // placeholder when Input is set
	// Select, when set, is laid out as a trigger slot with AttachedEdges
	// applied (shadcn ButtonGroup + Select).
	Select *Select
	// Widget renders arbitrary content in the slot instead of a
	// button. Chrome is not auto-attached — use Input/Select for fields.
	Widget layout.Widget
	// Flex, when non-zero, lets the slot absorb remaining space.
	Flex float32
}

// ButtonGroup attaches its children edge-to-edge: shared borders,
// square inner corners, only the group's outer corners rounded.
// Children stretch to one cross-axis size (shadcn items-stretch).
// Neighbors overlap by 1dp so their borders collapse to one line.
func ButtonGroup(th *Theme, o ButtonGroupProps, items ...ButtonGroupItem) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		n := len(items)
		if n == 0 {
			return layout.Dimensions{}
		}
		overlap := gtx.Dp(1)

		// Attach edges facing a neighbor (separators don't break the
		// attachment — they ARE the seam).
		attached := func(i int) AttachedEdges {
			if o.Vertical {
				return AttachedEdges{Top: i > 0, Bottom: i < n-1}
			}
			return AttachedEdges{Start: i > 0, End: i < n-1}
		}
		widgets := make([]layout.Widget, n)
		flexes := make([]float32, n)
		for i, it := range items {
			i, it := i, it
			flexes[i] = it.Flex
			a := attached(i)
			switch {
			case it.Separator:
				widgets[i] = nil // drawn from the measured pass
			case it.Btn != nil:
				opts := it.Props
				opts.Attached = a
				widgets[i] = Button(th, it.Btn, it.Label, opts)
			case it.Input != nil:
				in, hint := it.Input, it.Hint
				in.Attached = a
				widgets[i] = func(gtx layout.Context) layout.Dimensions {
					return in.LayoutField(th, gtx, hint)
				}
			case it.Select != nil:
				sel := it.Select
				sel.Attached = a
				widgets[i] = func(gtx layout.Context) layout.Dimensions {
					return sel.Layout(th, gtx, "")
				}
			case it.Widget != nil:
				widgets[i] = it.Widget
			}
		}

		// Measure: rigids at natural size, flexed slots split what
		// remains on the main axis.
		axisMax := gtx.Constraints.Max.X
		if o.Vertical {
			axisMax = gtx.Constraints.Max.Y
		}
		type slot struct {
			call op.CallOp
			dims layout.Dimensions
			main int // natural / flexed main-axis size
		}
		slots := make([]slot, n)
		var totalFlex float32
		rigidSum := 0
		sepW := gtx.Dp(unit.Dp(1))
		for i, w := range widgets {
			if items[i].Separator {
				rigidSum += sepW
				continue
			}
			if flexes[i] > 0 {
				totalFlex += flexes[i]
				continue
			}
			// Natural size only — every slot is laid out again at exact
			// size further down, so this recording is discarded.
			beginMeasurePass()
			m := op.Record(gtx.Ops)
			cgtx := gtx
			cgtx.Constraints.Min = image.Point{}
			d := w(cgtx)
			slots[i] = slot{dims: d, call: m.Stop()}
			endMeasurePass()
			if o.Vertical {
				slots[i].main = d.Size.Y
				rigidSum += d.Size.Y
			} else {
				slots[i].main = d.Size.X
				rigidSum += d.Size.X
			}
		}
		if totalFlex > 0 {
			room := axisMax - rigidSum + overlap*(n-1)
			if room < 0 {
				room = 0
			}
			for i := range widgets {
				if flexes[i] <= 0 || items[i].Separator {
					continue
				}
				share := int(float32(room) * flexes[i] / totalFlex)
				slots[i].main = share
			}
		}

		// Cross-axis extent = the tallest (widest) natural child.
		// Flexed slots are measured once at their main size to learn
		// their natural cross before the stretch pass.
		cross := 0
		for i, w := range widgets {
			if items[i].Separator || w == nil {
				continue
			}
			if flexes[i] > 0 {
				beginMeasurePass()
				m := op.Record(gtx.Ops)
				cgtx := gtx
				if o.Vertical {
					cgtx.Constraints = layout.Constraints{
						Min: image.Pt(0, slots[i].main),
						Max: image.Pt(gtx.Constraints.Max.X, slots[i].main),
					}
				} else {
					cgtx.Constraints = layout.Constraints{
						Min: image.Pt(slots[i].main, 0),
						Max: image.Pt(slots[i].main, gtx.Constraints.Max.Y),
					}
				}
				d := w(cgtx)
				slots[i] = slot{dims: d, call: m.Stop(), main: slots[i].main}
				endMeasurePass()
			}
			c := slots[i].dims.Size.Y
			if o.Vertical {
				c = slots[i].dims.Size.X
			}
			if c > cross {
				cross = c
			}
		}

		// Stretch: re-layout every child at Exact(main × cross) so
		// chrome fills the shared bar (items-stretch).
		for i, w := range widgets {
			if items[i].Separator || w == nil {
				continue
			}
			m := op.Record(gtx.Ops)
			cgtx := gtx
			if o.Vertical {
				cgtx.Constraints = layout.Exact(image.Pt(cross, slots[i].main))
			} else {
				cgtx.Constraints = layout.Exact(image.Pt(slots[i].main, cross))
			}
			d := w(cgtx)
			slots[i] = slot{dims: d, call: m.Stop(), main: slots[i].main}
		}

		// Place: neighbors overlapping by 1dp along the main axis.
		pos := 0
		for i := range slots {
			if items[i].Separator {
				var sz image.Point
				if o.Vertical {
					sz = image.Pt(cross, sepW)
				} else {
					sz = image.Pt(sepW, cross)
				}
				var at image.Point
				if o.Vertical {
					at = image.Pt(0, pos)
				} else {
					at = image.Pt(pos, 0)
				}
				off := op.Offset(at).Push(gtx.Ops)
				fgtx := gtx
				fgtx.Constraints = layout.Exact(sz)
				Fill(fgtx, th.Palette.Border)
				off.Pop()
				pos += sepW - overlap
				continue
			}
			d := slots[i].dims.Size
			var at image.Point
			if o.Vertical {
				at = image.Pt(0, pos)
				pos += d.Y - overlap
			} else {
				at = image.Pt(pos, 0)
				pos += d.X - overlap
			}
			off := op.Offset(at).Push(gtx.Ops)
			slots[i].call.Add(gtx.Ops)
			off.Pop()
		}
		pos += overlap

		if o.Vertical {
			return layout.Dimensions{Size: image.Pt(cross, pos)}
		}
		return layout.Dimensions{Size: image.Pt(pos, cross)}
	}
}

// ButtonGroupSeparator is the ready-made separator item.
func ButtonGroupSeparator() ButtonGroupItem { return ButtonGroupItem{Separator: true} }

// ButtonGroupButton packs a labeled button slot (build-time composition).
func ButtonGroupButton(btn *widget.Clickable, label string, p ButtonProps) ButtonGroupItem {
	return ButtonGroupItem{Btn: btn, Label: label, Props: p}
}

// ButtonGroupInput packs an Input field slot; the group applies
// AttachedEdges so the field fuses with neighboring buttons.
func ButtonGroupInput(in *Input, hint string, flex float32) ButtonGroupItem {
	return ButtonGroupItem{Input: in, Hint: hint, Flex: flex}
}

// ButtonGroupSelect packs a Select trigger slot; the group applies
// AttachedEdges so the trigger fuses with neighboring controls.
func ButtonGroupSelect(sel *Select, flex float32) ButtonGroupItem {
	return ButtonGroupItem{Select: sel, Flex: flex}
}

// ButtonGroupWidget packs an arbitrary widget slot with optional flex weight.
func ButtonGroupWidget(w layout.Widget, flex float32) ButtonGroupItem {
	return ButtonGroupItem{Widget: w, Flex: flex}
}
