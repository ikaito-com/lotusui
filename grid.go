package lotusui

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

// Grid is the 2D layout primitive: equal-width tracks (repeat(N, 1fr)
// in web terms), items auto-flowing row-major, each optionally spanning
// columns and rows. Row heights derive from content; an item spanning
// rows stretches across them.
//
// Columns is the fixed track count. Cols, when set (Cols(n).At(…)),
// wins and resolves against Theme.Breakpoints each frame.
type Grid struct {
	Columns int           // number of equal-width tracks when Cols unset; min 1
	Cols    ResponsiveInt // stepped columns; when Set(), overrides Columns
	Gap     unit.Dp       // both axes, unless overridden below
	RowGap  unit.Dp
	ColGap  unit.Dp
	Gaps    ResponsiveDp // when Set(), both axes (unless RowGap/ColGap set after resolve)
}

// GridItem is one grid cell: a widget plus how many tracks it spans.
// Zero spans mean 1. ColSpans / RowSpans override scalars when Set().
type GridItem struct {
	ColSpan, RowSpan   int
	ColSpans, RowSpans ResponsiveInt
	W                  layout.Widget
}

// Cell wraps a widget as a 1×1 grid item.
func Cell(w layout.Widget) GridItem { return GridItem{W: w} }

// Span wraps a widget as an item spanning cols columns.
func Span(cols int, w layout.Widget) GridItem { return GridItem{ColSpan: cols, W: w} }

func (g Grid) resolveCols(th *Theme, idx int) int {
	cols := g.Columns
	if g.Cols.Set() {
		cols = g.Cols.ResolveAt(th, idx)
	}
	if cols < 1 {
		cols = 1
	}
	return cols
}

func (g Grid) resolveGaps(th *Theme, idx int) (rowGap, colGap unit.Dp) {
	rowGap, colGap = g.RowGap, g.ColGap
	if g.Gaps.Set() {
		d := g.Gaps.ResolveAt(th, idx)
		if rowGap == 0 {
			rowGap = d
		}
		if colGap == 0 {
			colGap = d
		}
	}
	if rowGap == 0 {
		rowGap = g.Gap
	}
	if colGap == 0 {
		colGap = g.Gap
	}
	return rowGap, colGap
}

func (it GridItem) resolveSpans(th *Theme, idx int, cols int) (cs, rs int) {
	cs, rs = it.ColSpan, it.RowSpan
	if it.ColSpans.Set() {
		cs = it.ColSpans.ResolveAt(th, idx)
	}
	if it.RowSpans.Set() {
		rs = it.RowSpans.ResolveAt(th, idx)
	}
	if cs < 1 {
		cs = 1
	}
	if cs > cols {
		cs = cols
	}
	if rs < 1 {
		rs = 1
	}
	return cs, rs
}

// Layout places items. th supplies Breakpoints for Cols / Gaps / spans.
func (g Grid) Layout(th *Theme, gtx layout.Context, items ...GridItem) layout.Dimensions {
	idx := th.BreakpointIndex(gtx)
	cols := g.resolveCols(th, idx)
	rowGap, colGap := g.resolveGaps(th, idx)
	rg, cg := gtx.Dp(rowGap), gtx.Dp(colGap)
	cellW := (gtx.Constraints.Max.X - cg*(cols-1)) / cols
	if cellW < 0 {
		cellW = 0
	}

	// Row-major auto-placement with an occupancy matrix, exactly the
	// web's auto-flow: each item takes the first slot wide and deep
	// enough for its spans.
	type placed struct {
		item       GridItem
		row, col   int
		cs, rs     int
		w, natural int
	}
	var occupied []([]bool)
	rowOf := func(r int) []bool {
		for len(occupied) <= r {
			occupied = append(occupied, make([]bool, cols))
		}
		return occupied[r]
	}
	fits := func(r, c, cs, rs int) bool {
		for dr := 0; dr < rs; dr++ {
			row := rowOf(r + dr)
			for dc := 0; dc < cs; dc++ {
				if row[c+dc] {
					return false
				}
			}
		}
		return true
	}
	mark := func(r, c, cs, rs int) {
		for dr := 0; dr < rs; dr++ {
			row := rowOf(r + dr)
			for dc := 0; dc < cs; dc++ {
				row[c+dc] = true
			}
		}
	}

	scratch := new(op.Ops)
	var cells []placed
	for _, it := range items {
		cs, rs := it.resolveSpans(th, idx, cols)
		r, c := 0, 0
	place:
		for {
			for c = 0; c+cs <= cols; c++ {
				if fits(r, c, cs, rs) {
					break place
				}
			}
			r++
		}
		mark(r, c, cs, rs)
		w := cs*cellW + (cs-1)*cg
		// Measure at span width, unbounded height — never into the frame.
		// Disabled + bracketed: a throwaway pass must neither eat the
		// frame's events nor claim a floating site (see MeasurePass).
		scratch.Reset()
		beginMeasurePass()
		m := gtx.Disabled()
		m.Ops = scratch
		m.Constraints = layout.Constraints{Min: image.Pt(w, 0), Max: image.Pt(w, 1<<20)}
		nat := it.W(m).Size.Y
		endMeasurePass()
		cells = append(cells, placed{item: it, row: r, col: c, cs: cs, rs: rs, w: w, natural: nat})
	}

	// Row heights: single-row items set each row's minimum; spanning
	// items then grow their LAST row when the spanned span is short.
	rowH := make([]int, len(occupied))
	for _, p := range cells {
		if p.rs == 1 && p.natural > rowH[p.row] {
			rowH[p.row] = p.natural
		}
	}
	for _, p := range cells {
		if p.rs == 1 {
			continue
		}
		have := rg * (p.rs - 1)
		for dr := 0; dr < p.rs; dr++ {
			have += rowH[p.row+dr]
		}
		if p.natural > have {
			rowH[p.row+p.rs-1] += p.natural - have
		}
	}
	rowY := make([]int, len(rowH)+1)
	for i, h := range rowH {
		rowY[i+1] = rowY[i] + h + rg
	}

	// Render: every item stretches to its rows' full height (Min.Y),
	// so cards in a row stay equal and row-spanning panels fill.
	for _, p := range cells {
		h := rg * (p.rs - 1)
		for dr := 0; dr < p.rs; dr++ {
			h += rowH[p.row+dr]
		}
		x := p.col * (cellW + cg)
		st := op.Offset(image.Pt(x, rowY[p.row])).Push(gtx.Ops)
		cgtx := gtx
		cgtx.Constraints = layout.Constraints{Min: image.Pt(p.w, h), Max: image.Pt(p.w, h)}
		p.item.W(cgtx)
		st.Pop()
	}
	total := 0
	if len(rowH) > 0 {
		total = rowY[len(rowH)] - rg
	}
	return layout.Dimensions{Size: gtx.Constraints.Constrain(image.Pt(gtx.Constraints.Max.X, total))}
}

// SimpleGridProps configures SimpleGrid. When Columns is Set(), column
// count resolves from Theme.Breakpoints (Chakra columns={{…}}).
// Otherwise columns derive continuously from MinChildWidth / MaxCols.
type SimpleGridProps struct {
	MinChildWidth unit.Dp
	MaxCols       int
	Columns       ResponsiveInt
	Gap           unit.Dp
	Gaps          ResponsiveDp
}

// SimpleGrid lays out one cell per item. Continuous mode (Columns
// unset): column count = min(MaxCols, floor(Max.X / MinChildWidth)).
// Stepped mode (Columns set): Resolve against Theme.Breakpoints.
// Rows share the tallest cell's height; a trailing partial row is
// padded so cells don't stretch.
func SimpleGrid[T any](th *Theme, gtx layout.Context, items []T, p SimpleGridProps, cell func(layout.Context, T) layout.Dimensions) layout.Dimensions {
	idx := th.BreakpointIndex(gtx)
	gap := p.Gap
	if p.Gaps.Set() {
		gap = p.Gaps.ResolveAt(th, idx)
	}
	var cols int
	if p.Columns.Set() {
		cols = p.Columns.ResolveAt(th, idx)
		if cols < 1 {
			cols = 1
		}
	} else {
		minW := p.MinChildWidth
		if minW < 1 {
			minW = 1
		}
		cols = gtx.Constraints.Max.X / gtx.Dp(minW)
		if cols < 1 {
			cols = 1
		}
		if p.MaxCols > 0 && cols > p.MaxCols {
			cols = p.MaxCols
		}
	}

	// One scratch Ops serves every measure in this grid — a fresh Ops
	// per cell would put len(items) heap allocations in every frame.
	scratch := new(op.Ops)

	var rows []layout.FlexChild
	for i := 0; i < len(items); i += cols {
		row := items[i:min(i+cols, len(items))]
		if i > 0 {
			rows = append(rows, layout.Rigid(Spacer(gap)))
		}
		rows = append(rows, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layoutSimpleRow(gtx, cols, gap, row, cell, scratch)
		}))
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, rows...)
}

// layoutSimpleRow lays out one grid row of equal-width, equal-height
// cells: a throwaway measure pass finds the tallest cell, then the
// real pass hands every cell that height as its minimum.
func layoutSimpleRow[T any](gtx layout.Context, cols int, gap unit.Dp, items []T, cell func(layout.Context, T) layout.Dimensions, scratch *op.Ops) layout.Dimensions {
	gapPx := gtx.Dp(gap)
	cellW := (gtx.Constraints.Max.X - gapPx*(cols-1)) / cols
	if cellW < 0 {
		cellW = 0
	}

	maxH := 0
	for _, it := range items {
		scratch.Reset()
		beginMeasurePass()
		m := gtx.Disabled()
		m.Ops = scratch // measure only — never added to the frame
		m.Constraints.Min = image.Pt(cellW, 0)
		m.Constraints.Max.X = cellW
		if d := cell(m, it); d.Size.Y > maxH {
			maxH = d.Size.Y
		}
		endMeasurePass()
	}

	var children []layout.FlexChild
	for i, it := range items {
		it := it
		if i > 0 {
			children = append(children, layout.Rigid(HSpacer(gap)))
		}
		children = append(children, layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.Y = maxH
			return cell(gtx, it)
		}))
	}
	// Pad the row with empty flexed space if it's not full, so cells don't stretch.
	for i := len(items); i < cols; i++ {
		children = append(children, layout.Rigid(HSpacer(gap)), layout.Flexed(1, func(gtx layout.Context) layout.Dimensions { return layout.Dimensions{} }))
	}
	return layout.Flex{}.Layout(gtx, children...)
}
