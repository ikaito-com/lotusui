package lotusui

import (
	"image"
	"testing"
)

func TestHoverCardOffsetSides(t *testing.T) {
	h := HoverCard{}
	anchor := image.Pt(40, 20)
	card := image.Pt(200, 80)
	gap := 6

	// Zero Align is Center (shadcn).
	h.Side = HoverCardBottom
	if got := h.offset(anchor, card, gap); got != image.Pt((40-200)/2, 26) {
		t.Fatalf("bottom center (zero Align): got %v", got)
	}
	h.Align = PopoverStart
	if got := h.offset(anchor, card, gap); got != image.Pt(0, 26) {
		t.Fatalf("bottom start: got %v", got)
	}
	h.Side = HoverCardTop
	if got := h.offset(anchor, card, gap); got != image.Pt(0, -86) {
		t.Fatalf("top start: got %v", got)
	}
	h.Align = PopoverCenter
	h.Side = HoverCardLeft
	if got := h.offset(anchor, card, gap); got != image.Pt(-206, (20-80)/2) {
		t.Fatalf("left center: got %v", got)
	}
	h.Side = HoverCardRight
	if got := h.offset(anchor, card, gap); got != image.Pt(46, (20-80)/2) {
		t.Fatalf("right center: got %v", got)
	}
}
