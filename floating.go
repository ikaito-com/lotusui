package lotusui

import (
	"image"
	"time"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
)

// layoutSites numbers Layout calls that share one widget within a
// single frame (gtx.Now is constant for the frame). Used so a shared
// HoverCard/Tooltip/menu/… does not paint N stacked floating panels.
type layoutSites struct {
	epoch time.Time
	seq   int
}

func (s *layoutSites) next(now time.Time) int {
	if s.epoch != now {
		s.epoch = now
		s.seq = 0
	}
	i := s.seq
	s.seq++
	return i
}

// The floating layer — the shared portal primitive behind Select's
// dropdown, menu triggers, tooltips, hover cards and popovers.
//
// Gio's op.Defer is the mechanism: deferred ops run after everything
// else in the frame, keep the CALLER'S transform, and reset all other
// state (including clips — a dropdown must escape its parents'
// clipping). Hit-testing follows paint order, so the floating layer
// also wins the pointer over everything beneath it.

// Floating paints content above everything else this frame, anchored
// at the caller's current position.
func Floating(gtx layout.Context, content layout.Widget) {
	m := op.Record(gtx.Ops)
	content(gtx)
	op.Defer(gtx.Ops, m.Stop())
}

// dismisser implements the floating layer's "anywhere else closes it"
// contract: a huge press-catcher painted UNDER the panel (the panel's
// own widgets, added after, stay on top of it), plus Escape. One
// dismisser per floating owner; its address is the event tag.
type dismisser struct{}

// Dismissed drains this frame's events and reports whether the layer
// should close: a press outside the panel, or Escape.
func (d *dismisser) Dismissed(gtx layout.Context) bool {
	dismissed := false
	for {
		ev, ok := gtx.Event(
			pointer.Filter{Target: d, Kinds: pointer.Press},
			key.Filter{Name: key.NameEscape},
		)
		if !ok {
			break
		}
		switch ev.(type) {
		case pointer.Event, key.Event:
			dismissed = true
		}
	}
	return dismissed
}

// Add registers the catcher: an effectively unbounded input area at
// the floating layer's base. Call it FIRST inside Floating's content,
// before painting the panel.
func (d *dismisser) Add(gtx layout.Context) {
	defer clip.Rect(image.Rect(-1<<14, -1<<14, 1<<14, 1<<14)).Push(gtx.Ops).Pop()
	event.Op(gtx.Ops, d)
}
