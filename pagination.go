package lotusui

import (
	"strconv"

	"gioui.org/layout"
	"gioui.org/widget"
)

// Pagination is the page selector: previous/next plus numbered pages,
// long ranges elided around the current page. Page is 1-based and
// lives on the struct; clicks are processed at the top of Layout.
type Pagination struct {
	Page int
	// Simple renders numbered links only — no previous/next, no
	// elision.
	Simple     bool
	prev, next widget.Clickable
	nums       []widget.Clickable
}

// pageWindow returns the visible page numbers (0 = ellipsis).
func pageWindow(page, total int) []int {
	if total <= 7 {
		out := make([]int, total)
		for i := range out {
			out[i] = i + 1
		}
		return out
	}
	switch {
	case page <= 4:
		return []int{1, 2, 3, 4, 5, 0, total}
	case page >= total-3:
		return []int{1, 0, total - 4, total - 3, total - 2, total - 1, total}
	default:
		return []int{1, 0, page - 1, page, page + 1, 0, total}
	}
}

func (p *Pagination) Layout(th *Theme, gtx layout.Context, total int) layout.Dimensions {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.Page > total {
		p.Page = total
	}
	if len(p.nums) != total {
		p.nums = make([]widget.Clickable, total)
	}
	if p.prev.Clicked(gtx) && p.Page > 1 {
		p.Page--
	}
	if p.next.Clicked(gtx) && p.Page < total {
		p.Page++
	}
	for i := range p.nums {
		if p.nums[i].Clicked(gtx) {
			p.Page = i + 1
		}
	}

	var row []layout.FlexChild
	add := func(w layout.Widget) {
		if len(row) > 0 {
			row = append(row, layout.Rigid(HSpacer(4)))
		}
		row = append(row, layout.Rigid(w))
	}
	if p.Simple {
		for n := 1; n <= total; n++ {
			n := n
			v := ButtonGhost
			if n == p.Page {
				v = ButtonOutline
			}
			add(Button(th, &p.nums[n-1], strconv.Itoa(n), ButtonProps{Variant: v, Size: SizeSM}))
		}
		return layout.Flex{Alignment: layout.Middle}.Layout(gtx, row...)
	}
	add(Button(th, &p.prev, "‹", ButtonProps{Variant: ButtonGhost, Size: SizeSM, Disabled: p.Page <= 1}))
	for _, n := range pageWindow(p.Page, total) {
		n := n
		if n == 0 {
			add(func(gtx layout.Context) layout.Dimensions {
				l := LabelBody(th, "…")
				l.Color = th.Palette.FgSubtle
				return layout.Inset{Left: 6, Right: 6}.Layout(gtx, l.Layout)
			})
			continue
		}
		v := ButtonGhost
		if n == p.Page {
			v = ButtonOutline
		}
		add(Button(th, &p.nums[n-1], strconv.Itoa(n), ButtonProps{Variant: v, Size: SizeSM}))
	}
	add(Button(th, &p.next, "›", ButtonProps{Variant: ButtonGhost, Size: SizeSM, Disabled: p.Page >= total}))
	return layout.Flex{Alignment: layout.Middle}.Layout(gtx, row...)
}
