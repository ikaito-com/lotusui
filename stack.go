package lotusui

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

// Stack lays out children along one axis with a uniform gap between
// them — Chakra UI's Stack, as a Gio widget. Children size themselves;
// the gap is the component's whole job, so screens stop hand-weaving
// Spacer children between rows.
type Stack struct {
	Axis  layout.Axis      // layout.Vertical (default) or layout.Horizontal
	Gap   unit.Dp          // uniform spacing between children
	Align layout.Alignment // cross-axis alignment; layout.Start is the zero value
}

func (s Stack) Layout(gtx layout.Context, children ...layout.Widget) layout.Dimensions {
	flexChildren := make([]layout.FlexChild, 0, len(children)*2)
	for i, child := range children {
		if i > 0 && s.Gap > 0 {
			if s.Axis == layout.Horizontal {
				flexChildren = append(flexChildren, layout.Rigid(HSpacer(s.Gap)))
			} else {
				flexChildren = append(flexChildren, layout.Rigid(Spacer(s.Gap)))
			}
		}
		flexChildren = append(flexChildren, layout.Rigid(child))
	}
	return layout.Flex{Axis: s.Axis, Alignment: s.Align}.Layout(gtx, flexChildren...)
}

// VStack stacks children vertically with a uniform gap — Chakra's
// VStack. Children keep their natural width (cross-axis Start).
func VStack(gap unit.Dp, children ...layout.Widget) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return Stack{Axis: layout.Vertical, Gap: gap}.Layout(gtx, children...)
	}
}

// HStack stacks children horizontally with a uniform gap — Chakra's
// HStack. Like Chakra's, children center on the cross axis, so a
// switch and its label sit on one visual line.
//
// HStack never wraps: under a narrow Max.X, Rigid children are squeezed
// and labels can wrap per character. Use Wrap when chips/labels must
// flow to the next line at their intrinsic width.
func HStack(gap unit.Dp, children ...layout.Widget) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return Stack{Axis: layout.Horizontal, Gap: gap, Align: layout.Middle}.Layout(gtx, children...)
	}
}

// Wrap lays children left-to-right and wraps to the next line when the
// next child would exceed Max.X — Chakra's Wrap / CSS flex-wrap. Each
// child is measured at its intrinsic width (unconstrained Max.X); a
// child wider than Max.X alone is laid out at Max.X. Gap is used both
// between items on a line and between lines. Align is the cross-axis
// alignment within each line (Middle for text + icons).
//
// Prefer Wrap over HStack for flowing chips/badges; prefer SimpleGrid
// when you want equal cells and a fixed column count.
func Wrap(gap unit.Dp, align layout.Alignment, children ...layout.Widget) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return (wrapLayout{Gap: gap, Align: align}).Layout(gtx, children...)
	}
}

type wrapLayout struct {
	Gap   unit.Dp
	Align layout.Alignment
}

func (w wrapLayout) Layout(gtx layout.Context, children ...layout.Widget) layout.Dimensions {
	n := len(children)
	if n == 0 {
		return layout.Dimensions{}
	}
	maxW := gtx.Constraints.Max.X
	gapPx := 0
	if w.Gap > 0 {
		gapPx = gtx.Dp(w.Gap)
	}

	// Measure each child at intrinsic width on scratch ops — never
	// squeeze into a 1-character column the way HStack/Flex can.
	sizes := make([]image.Point, n)
	for i, child := range children {
		var mop op.Ops
		mgtx := gtx
		mgtx.Ops = &mop
		mgtx.Constraints.Min = image.Point{}
		mgtx.Constraints.Max.X = 1 << 30
		d := child(mgtx)
		if maxW > 0 && d.Size.X > maxW {
			mop.Reset()
			mgtx.Constraints.Max.X = maxW
			d = child(mgtx)
		}
		sizes[i] = d.Size
	}

	type line struct {
		start, end int // end exclusive
		width      int
		height     int
	}
	var lines []line
	lineStart := 0
	lineW, lineH := 0, 0
	for i, sz := range sizes {
		need := sz.X
		if i > lineStart {
			need += gapPx
		}
		if i > lineStart && maxW > 0 && lineW+need > maxW {
			lines = append(lines, line{start: lineStart, end: i, width: lineW, height: lineH})
			lineStart = i
			lineW, lineH = sz.X, sz.Y
			continue
		}
		if i > lineStart {
			lineW += gapPx
		}
		lineW += sz.X
		if sz.Y > lineH {
			lineH = sz.Y
		}
	}
	lines = append(lines, line{start: lineStart, end: n, width: lineW, height: lineH})

	// Paint with the same constraints used when measuring.
	y, totalW := 0, 0
	for li, ln := range lines {
		if li > 0 {
			y += gapPx
		}
		x := 0
		for i := ln.start; i < ln.end; i++ {
			if i > ln.start {
				x += gapPx
			}
			cgtx := gtx
			cgtx.Constraints.Min = image.Point{}
			cgtx.Constraints.Max.X = sizes[i].X
			if maxW > 0 && sizes[i].X >= maxW {
				cgtx.Constraints.Max.X = maxW
			}
			cross := 0
			switch w.Align {
			case layout.Middle:
				cross = (ln.height - sizes[i].Y) / 2
			case layout.End:
				cross = ln.height - sizes[i].Y
			}
			st := op.Offset(image.Pt(x, y+cross)).Push(gtx.Ops)
			d := children[i](cgtx)
			st.Pop()
			x += d.Size.X
		}
		if ln.width > totalW {
			totalW = ln.width
		}
		y += ln.height
	}
	size := image.Pt(totalW, y)
	if maxW > 0 && size.X > maxW {
		size.X = maxW
	}
	size = gtx.Constraints.Constrain(size)
	return layout.Dimensions{Size: size}
}
