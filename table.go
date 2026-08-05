package lotusui

import (
	"gioui.org/font"
	"gioui.org/layout"
)

// TableProps are the table's options.
type TableProps struct {
	// Widths are per-column flex weights; nil = equal columns.
	Widths []float32
	// Footer renders a final emphasized row (totals) above the caption.
	Footer []string
	// Caption renders muted under the table.
	Caption string
}

// Table renders tabular data: a muted header row, hairlines between
// body rows, cells as arbitrary widgets. For plain text use
// TableText. Long collections belong in ListView; Table is for the
// bounded, comparison-shaped data a screen actually shows.
func Table(th *Theme, o TableProps, header []string, rows [][]layout.Widget) layout.Widget {
	weight := func(col int) float32 {
		if col < len(o.Widths) && o.Widths[col] > 0 {
			return o.Widths[col]
		}
		return 1
	}
	rowOf := func(cells []layout.Widget, pad layout.Inset) layout.Widget {
		return func(gtx layout.Context) layout.Dimensions {
			var fl []layout.FlexChild
			for c, cell := range cells {
				c, cell := c, cell
				fl = append(fl, layout.Flexed(weight(c), func(gtx layout.Context) layout.Dimensions {
					return pad.Layout(gtx, cell)
				}))
			}
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx, fl...)
		}
	}
	return func(gtx layout.Context) layout.Dimensions {
		pad := layout.Inset{Top: 10, Bottom: 10, Left: 8, Right: 8}
		var body []layout.Widget
		if len(header) > 0 {
			var hs []layout.Widget
			for _, h := range header {
				h := h
				hs = append(hs, func(gtx layout.Context) layout.Dimensions {
					l := LabelCaption(th, h)
					l.Color = th.Palette.FgSubtle
					l.Font.Weight = font.Medium
					return l.Layout(gtx)
				})
			}
			body = append(body, rowOf(hs, pad), Hairline(th))
		}
		for i, r := range rows {
			if i > 0 {
				body = append(body, Hairline(th))
			}
			body = append(body, rowOf(r, pad))
		}
		if len(o.Footer) > 0 {
			var fs []layout.Widget
			for _, f := range o.Footer {
				f := f
				fs = append(fs, func(gtx layout.Context) layout.Dimensions {
					l := LabelBody(th, f)
					l.Color = th.Palette.Fg
					l.Font.Weight = font.Medium
					return l.Layout(gtx)
				})
			}
			body = append(body, Hairline(th), rowOf(fs, pad))
		}
		if o.Caption != "" {
			body = append(body, Spacer(th.Space.SM), func(gtx layout.Context) layout.Dimensions {
				l := LabelCaption(th, o.Caption)
				l.Color = th.Palette.FgSubtle
				return l.Layout(gtx)
			})
		}
		return VStack(0, body...)(gtx)
	}
}

// TableText is Table for plain string cells.
func TableText(th *Theme, o TableProps, header []string, rows [][]string) layout.Widget {
	wrows := make([][]layout.Widget, len(rows))
	for i, r := range rows {
		cells := make([]layout.Widget, len(r))
		for c, txt := range r {
			txt := txt
			cells[c] = func(gtx layout.Context) layout.Dimensions {
				l := LabelBody(th, txt)
				l.Color = th.Palette.Fg
				l.MaxLines = 1
				return l.Layout(gtx)
			}
		}
		wrows[i] = cells
	}
	return Table(th, o, header, wrows)
}
