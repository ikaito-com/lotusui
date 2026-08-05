package lotusui

import (
	"image"

	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
)

// ScrollArea is shadcn Scroll Area — whole-content scrolling inside a
// bounded viewport WITHOUT layout.List's op.Record macros.
//
// Floating (Select, DropdownMenu, Popover, Tooltip, HoverCard) records
// into op.Defer so the portal paints above the frame. layout.List
// records every child; deferred ops then stay trapped inside the
// scroller and open panels look like they push the page. ScrollArea
// clips and offsets on the root ops stack so portals escape correctly.
//
// Prefer ScrollArea for screens that host Floating widgets. Prefer
// ListView for long uniform collections (virtualized). Scrollable is
// the older List-backed sibling — fine when content has no portals.
//
// State (Offset) must outlive the frame. The zero value scrolls
// vertically — do NOT use layout.Axis here: its zero value is
// Horizontal, which made bare ScrollArea{} sideways-scroll with
// Max.X unbounded (labels never wrap; the whole docs shell felt like
// a wide iframe).
type ScrollArea struct {
	Offset int
	// Horizontal scrolls on X. Zero value is vertical.
	Horizontal bool
	tag        struct{}
}

// Reset scrolls back to the start (route changes, filters).
func (s *ScrollArea) Reset() { s.Offset = 0 }

// ScrollTo jumps to a content offset in px (clamped on the next Layout).
func (s *ScrollArea) ScrollTo(offset int) {
	if offset < 0 {
		offset = 0
	}
	s.Offset = offset
}

// ScrollAreaProps tunes one Layout. The zero value matches Scrollable's
// page-level behaviour (vertical, shadowRoom inset for card shadows).
type ScrollAreaProps struct {
	// NoShadowRoom skips the card-shadow inset. Pane helpers that
	// already budget Card pad use this so content is not double-inset.
	NoShadowRoom bool
}

// Layout lays content in a clipped, wheel-scrollable viewport.
// content must not rely on Flexed filling "remaining" height — the
// measure pass is unbounded on the scroll axis (same rule as Scrollable).
func (s *ScrollArea) Layout(th *Theme, gtx layout.Context, content layout.Widget) layout.Dimensions {
	return s.LayoutWith(th, gtx, ScrollAreaProps{}, content)
}

// LayoutWith is Layout plus props (e.g. NoShadowRoom).
func (s *ScrollArea) LayoutWith(th *Theme, gtx layout.Context, o ScrollAreaProps, content layout.Widget) layout.Dimensions {
	return scrollArea(th, s, gtx, o, content)
}

func scrollArea(th *Theme, s *ScrollArea, gtx layout.Context, o ScrollAreaProps, content layout.Widget) layout.Dimensions {
	_ = th
	viewport := gtx.Constraints.Max
	if viewport.X < 1 || viewport.Y < 1 {
		return layout.Dimensions{}
	}
	horiz := s.Horizontal

	cl := clip.Rect{Max: viewport}.Push(gtx.Ops)
	defer cl.Pop()

	event.Op(gtx.Ops, &s.tag)
	for {
		f := pointer.Filter{Target: &s.tag, Kinds: pointer.Scroll}
		if horiz {
			f.ScrollX = pointer.ScrollRange{Min: -1e6, Max: 1e6}
		} else {
			f.ScrollY = pointer.ScrollRange{Min: -1e6, Max: 1e6}
		}
		ev, ok := gtx.Event(f)
		if !ok {
			break
		}
		if e, ok := ev.(pointer.Event); ok {
			if horiz {
				s.Offset += int(e.Scroll.X)
			} else {
				s.Offset += int(e.Scroll.Y)
			}
		}
	}

	off := image.Point{}
	if horiz {
		off.X = -s.Offset
	} else {
		off.Y = -s.Offset
	}
	trans := op.Offset(off).Push(gtx.Ops)

	cgtx := gtx
	cgtx.Constraints.Min = image.Point{}
	if horiz {
		cgtx.Constraints.Max.X = 1 << 20
	} else {
		cgtx.Constraints.Max.Y = 1 << 20
	}

	var dims layout.Dimensions
	if o.NoShadowRoom {
		dims = content(cgtx)
	} else {
		inset := layout.Inset{}
		if horiz {
			inset = layout.Inset{Left: shadowRoom, Right: shadowRoom}
		} else {
			inset = layout.UniformInset(shadowRoom)
		}
		dims = inset.Layout(cgtx, content)
	}
	trans.Pop()

	max := 0
	if horiz {
		max = dims.Size.X - viewport.X
	} else {
		max = dims.Size.Y - viewport.Y
	}
	if max < 0 {
		max = 0
	}
	if s.Offset < 0 {
		s.Offset = 0
	}
	if s.Offset > max {
		s.Offset = max
	}
	return layout.Dimensions{Size: viewport}
}
