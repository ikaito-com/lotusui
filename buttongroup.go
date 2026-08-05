package lotusui

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op"
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

// ButtonGroupProps are the group's options.
type ButtonGroupProps struct {
	// Vertical stacks the group top-to-bottom.
	Vertical bool
}

// ButtonGroupItem is one slot of the group: a button (Btn + Label +
// Props), a Separator, or an arbitrary Widget carried as-is (an
// Input — its own chrome stays). Flex gives the slot flexed weight.
type ButtonGroupItem struct {
	Btn   *widget.Clickable
	Label string
	Props ButtonProps
	// Separator renders the hairline divider between attached
	// neighbors — the split-button seam.
	Separator bool
	// Widget renders arbitrary content in the slot instead of a
	// button.
	Widget layout.Widget
	// Flex, when non-zero, lets the slot absorb remaining space.
	Flex float32
}

// ButtonGroup attaches its children edge-to-edge: shared borders,
// square inner corners, only the group's outer corners rounded.
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
		widgets := make([]layout.Widget, n)
		flexes := make([]float32, n)
		for i, it := range items {
			i, it := i, it
			flexes[i] = it.Flex
			switch {
			case it.Separator:
				widgets[i] = nil // drawn from the measured pass
			case it.Widget != nil:
				widgets[i] = it.Widget
			default:
				opts := it.Props
				if o.Vertical {
					opts.Attached = AttachedEdges{Top: i > 0, Bottom: i < n-1}
				} else {
					opts.Attached = AttachedEdges{Start: i > 0, End: i < n-1}
				}
				widgets[i] = Button(th, it.Btn, it.Label, opts)
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
			m := op.Record(gtx.Ops)
			cgtx := gtx
			cgtx.Constraints.Min = image.Point{}
			slots[i] = slot{dims: w(cgtx)}
			slots[i].call = m.Stop()
			if o.Vertical {
				rigidSum += slots[i].dims.Size.Y
			} else {
				rigidSum += slots[i].dims.Size.X
			}
		}
		if totalFlex > 0 {
			room := axisMax - rigidSum + overlap*(n-1)
			if room < 0 {
				room = 0
			}
			for i, w := range widgets {
				if flexes[i] <= 0 || items[i].Separator {
					continue
				}
				m := op.Record(gtx.Ops)
				cgtx := gtx
				share := int(float32(room) * flexes[i] / totalFlex)
				if o.Vertical {
					cgtx.Constraints = layout.Constraints{Min: image.Pt(0, share), Max: image.Pt(gtx.Constraints.Max.X, share)}
				} else {
					cgtx.Constraints = layout.Constraints{Min: image.Pt(share, 0), Max: image.Pt(share, gtx.Constraints.Max.Y)}
				}
				slots[i] = slot{dims: w(cgtx)}
				slots[i].call = m.Stop()
			}
		}

		// Cross-axis extent = the tallest (widest) child.
		cross := 0
		for i := range slots {
			c := slots[i].dims.Size.Y
			if o.Vertical {
				c = slots[i].dims.Size.X
			}
			if c > cross {
				cross = c
			}
		}

		// Place: each child centered on the cross axis, neighbors
		// overlapping by 1dp along the main axis.
		pos := 0
		for i := range slots {
			if items[i].Separator {
				// The seam: a hairline spanning the cross axis, sitting
				// exactly on the overlap.
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
				fgtx.Constraints = layout.Constraints{Min: sz, Max: sz}
				Fill(fgtx, th.Palette.Border)
				off.Pop()
				pos += sepW - overlap
				continue
			}
			d := slots[i].dims.Size
			var at image.Point
			if o.Vertical {
				at = image.Pt((cross-d.X)/2, pos)
				pos += d.Y - overlap
			} else {
				at = image.Pt(pos, (cross-d.Y)/2)
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

// ButtonGroupWidget packs an arbitrary widget slot with optional flex weight.
func ButtonGroupWidget(w layout.Widget, flex float32) ButtonGroupItem {
	return ButtonGroupItem{Widget: w, Flex: flex}
}
