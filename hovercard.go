package lotusui

import (
	"image"
	"time"

	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
)

// HoverCard is the hover-preview panel: rest the pointer on a trigger
// and a content card floats in on the portal layer — for sighted
// users to preview what's behind a link (shadcn Hover Card). Unlike
// Tooltip (inverted chrome, a string label), HoverCard is a BgPanel
// card with arbitrary content and stays open while the pointer is
// over the card itself.
//
// Layout may be called multiple times per frame on the same HoverCard
// (one jargon tip shared across every "GB" in a table). Each call is a
// distinct trigger site; only the hovered site paints the floating
// card — never one open state spawning N stacked panels.
//
//	var tip lotusui.HoverCard
//	tip.Layout(th, gtx, cardBody, trigger)
//
// HoverCardSide selects where the card floats relative to the trigger.
type HoverCardSide int

const (
	HoverCardBottom HoverCardSide = iota // default
	HoverCardTop
	HoverCardLeft
	HoverCardRight
)

// Defaults match Radix Hover Card: a deliberate open rest, a short
// close grace so the pointer can travel the gap onto the card.
const (
	hoverCardOpenDelay  = 700 * time.Millisecond
	hoverCardCloseDelay = 300 * time.Millisecond
	hoverCardWidth      = unit.Dp(320) // shadcn's w-80
)

// hoverTrig is a heap-stable event tag for one Layout call site.
type hoverTrig struct{}

type HoverCard struct {
	// Side positions the card; bottom is the default.
	Side HoverCardSide
	// Align positions the card along the side. Zero is PopoverCenter
	// (shadcn HoverCardContent). A start-aligned card on a short
	// trigger looks detached — the bulk sits far beside the anchor.
	Align PopoverAlign
	// OpenDelay / CloseDelay override the defaults when non-zero.
	OpenDelay, CloseDelay time.Duration
	// Width is the maximum card width; zero means 320dp (shadcn w-80).
	// The card hugs its content up to that max — a short tip on a short
	// trigger (e.g. "GB") stays clipped to the word instead of sitting
	// in an empty 320dp slab that only looks centered in chrome.
	Width unit.Dp
	// Disabled prevents opening (chakra).
	Disabled bool

	overCard        bool
	open            bool
	openAt, closeAt time.Time
	// panel is the event tag for the floating card's hover area.
	panel struct{}

	// Per-frame Layout sites sharing this HoverCard.
	sites  layoutSites
	trigs  []*hoverTrig
	over   []bool
	active int // site index that owns the open card
}

func (h *HoverCard) openDelay() time.Duration {
	if h.OpenDelay > 0 {
		return h.OpenDelay
	}
	return hoverCardOpenDelay
}

func (h *HoverCard) closeDelay() time.Duration {
	if h.CloseDelay > 0 {
		return h.CloseDelay
	}
	return hoverCardCloseDelay
}

func (h *HoverCard) widthDp() unit.Dp {
	if h.Width != 0 {
		return h.Width
	}
	return hoverCardWidth
}

func (h *HoverCard) nextSite(now time.Time) (idx int, tag *hoverTrig) {
	idx = h.sites.next(now)
	for len(h.trigs) <= idx {
		h.trigs = append(h.trigs, new(hoverTrig))
		h.over = append(h.over, false)
	}
	return idx, h.trigs[idx]
}

func (h *HoverCard) anyOverTrig() bool {
	for _, o := range h.over {
		if o {
			return true
		}
	}
	return false
}

// Layout wraps trigger; content is the card body. The trigger keeps
// its own interactions (pass-through hover). The card opens after
// OpenDelay and closes after CloseDelay once the pointer has left
// both the trigger and the card.
func (h *HoverCard) Layout(th *Theme, gtx layout.Context, content, trigger layout.Widget) layout.Dimensions {
	// A THROWAWAY pass has no site: measure the trigger and touch no
	// state. The card floats, so it never contributes to the size.
	if inMeasurePass() {
		return trigger(gtx)
	}
	idx, tag := h.nextSite(gtx.Now)

	for {
		ev, ok := gtx.Event(pointer.Filter{Target: tag, Kinds: pointer.Enter | pointer.Leave | pointer.Cancel})
		if !ok {
			break
		}
		if e, isPtr := ev.(pointer.Event); isPtr {
			switch e.Kind {
			case pointer.Enter:
				h.over[idx] = true
				h.active = idx
				h.closeAt = time.Time{}
				if !h.open {
					h.openAt = gtx.Now
				}
			default:
				h.over[idx] = false
				if h.open && !h.overCard && !h.anyOverTrig() {
					h.closeAt = gtx.Now
				}
			}
		}
	}
	for {
		ev, ok := gtx.Event(pointer.Filter{Target: &h.panel, Kinds: pointer.Enter | pointer.Leave | pointer.Cancel})
		if !ok {
			break
		}
		if e, isPtr := ev.(pointer.Event); isPtr {
			switch e.Kind {
			case pointer.Enter:
				h.overCard = true
				h.closeAt = time.Time{}
			default:
				h.overCard = false
				if h.open && !h.anyOverTrig() {
					h.closeAt = gtx.Now
				}
			}
		}
	}

	dims := trigger(gtx)

	// Pass-through input area over the trigger: hover intent without
	// stealing the child's events. Tag is per call site.
	func() {
		defer pointer.PassOp{}.Push(gtx.Ops).Pop()
		defer clip.Rect(image.Rectangle{Max: dims.Size}).Push(gtx.Ops).Pop()
		event.Op(gtx.Ops, tag)
	}()

	h.advance(gtx)

	// Only the hovered site paints — shared HoverCards across N
	// occurrences must not stack N identical panels.
	if h.open && !h.Disabled && h.active == idx {
		h.layoutCard(th, gtx, dims.Size, content)
	}
	return dims
}

// advance resolves open/close against the delay clocks and schedules
// the next wake-up — one frame AT the deadline, not every vsync.
func (h *HoverCard) advance(gtx layout.Context) {
	if h.Disabled {
		h.open = false
		return
	}
	if h.open {
		if !h.closeAt.IsZero() && !h.anyOverTrig() && !h.overCard {
			if gtx.Now.Sub(h.closeAt) >= h.closeDelay() {
				h.open = false
				h.closeAt = time.Time{}
				return
			}
			gtx.Execute(op.InvalidateCmd{At: h.closeAt.Add(h.closeDelay())})
		}
		return
	}
	if h.anyOverTrig() && !h.openAt.IsZero() {
		if gtx.Now.Sub(h.openAt) >= h.openDelay() {
			h.open = true
			h.openAt = time.Time{}
			return
		}
		gtx.Execute(op.InvalidateCmd{At: h.openAt.Add(h.openDelay())})
	}
}

func (h *HoverCard) layoutCard(th *Theme, gtx layout.Context, anchor image.Point, content layout.Widget) {
	Floating(gtx, func(gtx layout.Context) layout.Dimensions {
		maxW := gtx.Dp(h.widthDp())
		gtx.Constraints = layout.Constraints{
			Min: image.Point{},
			Max: image.Pt(maxW, gtx.Dp(480)),
		}
		m := op.Record(gtx.Ops)
		dims := layout.UniformInset(th.Space.MD).Layout(gtx, content)
		call := m.Stop()

		// Hug content; never stretch to Width when the tip is narrower.
		size := dims.Size
		if size.X > maxW {
			size.X = maxW
		}
		gap := gtx.Dp(6)
		at := h.offset(anchor, size, gap)
		defer op.Offset(at).Push(gtx.Ops).Pop()

		r := gtx.Dp(th.Radius.MD)
		cardShadow(gtx, size, r)
		defer clip.UniformRRect(image.Rectangle{Max: size}, r).Push(gtx.Ops).Pop()
		paint.Fill(gtx.Ops, th.Palette.BgPanel)
		widget.Border{Color: th.Palette.Border, Width: unit.Dp(1), CornerRadius: th.Radius.MD}.Layout(gtx,
			func(gtx layout.Context) layout.Dimensions { return layout.Dimensions{Size: size} })
		call.Add(gtx.Ops)

		// Hover tracking over the card keeps it open while the pointer
		// travels onto it; pass-through so content stays interactive.
		defer pointer.PassOp{}.Push(gtx.Ops).Pop()
		defer clip.Rect(image.Rectangle{Max: size}).Push(gtx.Ops).Pop()
		event.Op(gtx.Ops, &h.panel)

		return layout.Dimensions{Size: size}
	})
}

func (h *HoverCard) offset(anchor, card image.Point, gap int) image.Point {
	ax := func() int {
		switch h.Align {
		case PopoverStart:
			return 0
		case PopoverEnd:
			return anchor.X - card.X
		default: // PopoverCenter (zero)
			return (anchor.X - card.X) / 2
		}
	}
	ay := func() int {
		switch h.Align {
		case PopoverStart:
			return 0
		case PopoverEnd:
			return anchor.Y - card.Y
		default: // PopoverCenter (zero)
			return (anchor.Y - card.Y) / 2
		}
	}
	switch h.Side {
	case HoverCardTop:
		return image.Pt(ax(), -card.Y-gap)
	case HoverCardLeft:
		return image.Pt(-card.X-gap, ay())
	case HoverCardRight:
		return image.Pt(anchor.X+gap, ay())
	default:
		return image.Pt(ax(), anchor.Y+gap)
	}
}
