package lotusui

import (
	"gioui.org/gesture"
	"image"
	"image/color"
	"math"
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

// ScrollArea is shadcn Scroll Area — whole-content scrolling inside a
// bounded viewport WITHOUT layout.List's op.Record macros, with a
// macOS-style overlay Scrollbar (ScrollArea → ScrollBar composition).
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

	bar   Scrollbar
	hover bool
	// lastMax is the previous frame's scrollable distance, used to
	// bound what a gesture may consume THIS frame (content size is
	// only known after layout).
	lastMax  int
	measured bool
	tag      struct{}
	drag     gesture.Scroll
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
// page-level behaviour (vertical, shadowRoom inset for card shadows)
// plus a macOS overlay scrollbar (hover fade, MD thickness).
type ScrollAreaProps struct {
	// NoShadowRoom skips the card-shadow inset. Pane helpers that
	// already budget Card pad use this so content is not double-inset.
	NoShadowRoom bool
	// Scrollbar customizes the overlay thumb. Zero value = Hover + MD.
	Scrollbar ScrollbarProps
	// NoScrollbar hides the overlay thumb (wheel/drag still work via
	// the viewport; use when a parent chrome owns scrolling).
	NoScrollbar bool
}

// Layout lays content in a clipped, wheel-scrollable viewport with a
// macOS overlay scrollbar. content must not rely on Flexed filling
// "remaining" height — the measure pass is unbounded on the scroll
// axis (same rule as Scrollable).
func (s *ScrollArea) Layout(th *Theme, gtx layout.Context, content layout.Widget) layout.Dimensions {
	return s.LayoutWith(th, gtx, ScrollAreaProps{}, content)
}

// LayoutWith is Layout plus props (e.g. NoShadowRoom, Scrollbar).
func (s *ScrollArea) LayoutWith(th *Theme, gtx layout.Context, o ScrollAreaProps, content layout.Widget) layout.Dimensions {
	return scrollArea(th, s, gtx, o, content)
}

func scrollArea(th *Theme, s *ScrollArea, gtx layout.Context, o ScrollAreaProps, content layout.Widget) layout.Dimensions {
	viewport := gtx.Constraints.Max
	if viewport.X < 1 || viewport.Y < 1 {
		return layout.Dimensions{}
	}
	horiz := s.Horizontal

	cl := clip.Rect{Max: viewport}.Push(gtx.Ops)
	defer cl.Pop()

	event.Op(gtx.Ops, &s.tag)
	scrolled := false

	// Movement goes through gesture.Scroll — the same primitive
	// layout.List uses. A raw pointer.Scroll filter only ever sees
	// WHEEL events, so a touch drag scrolled nothing at all: the docs
	// were unscrollable on every phone. gesture.Scroll reads wheel,
	// drag and fling, and still honours the range so leftover scroll
	// chains to the parent.
	{
		rng := pointer.ScrollRange{}
		if s.measured {
			rng = pointer.ScrollRange{Min: -s.Offset, Max: s.lastMax - s.Offset}
			if rng.Min > 0 {
				rng.Min = 0
			}
			if rng.Max < 0 {
				rng.Max = 0
			}
		}
		axis := gesture.Vertical
		var xr, yr pointer.ScrollRange
		if horiz {
			axis, xr = gesture.Horizontal, rng
		} else {
			yr = rng
		}
		s.drag.Add(gtx.Ops)
		if d := s.drag.Update(gtx.Metric, gtx.Source, gtx.Now, axis, xr, yr); d != 0 {
			s.Offset += d
			scrolled = true
		}
		// A fling keeps moving after the finger lifts; keep frames
		// coming until it settles.
		if s.drag.State() == gesture.StateFlinging {
			gtx.Execute(op.InvalidateCmd{})
		}
	}

	for {
		// Consume only what is left to scroll. An unbounded range let a
		// gesture push Offset past the end for one frame — the content
		// painted overscrolled and snapped back next frame (elastic
		// glitch) — and swallowed scroll that should chain to a parent.
		// Before the first layout nothing is known, so nothing is taken.
		ev, ok := gtx.Event(pointer.Filter{
			Target: &s.tag,
			Kinds:  pointer.Enter | pointer.Leave,
		})
		if !ok {
			break
		}
		e, ok := ev.(pointer.Event)
		if !ok {
			continue
		}
		switch e.Kind {
		case pointer.Enter:
			s.hover = true
		case pointer.Leave:
			s.hover = false
		}
	}
	if scrolled {
		s.bar.poke(gtx)
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

	contentLen := dims.Size.Y
	viewLen := viewport.Y
	if horiz {
		contentLen = dims.Size.X
		viewLen = viewport.X
	}
	max := contentLen - viewLen
	if max < 0 {
		max = 0
	}
	if s.Offset < 0 {
		s.Offset = 0
	}
	if s.Offset > max {
		s.Offset = max
	}
	s.lastMax, s.measured = max, true

	if !o.NoScrollbar && max > 0 && contentLen > 0 {
		start := float32(s.Offset) / float32(contentLen)
		end := float32(s.Offset+viewLen) / float32(contentLen)
		if end > 1 {
			end = 1
		}
		bp := o.Scrollbar
		bp.Horizontal = horiz
		delta := s.bar.layoutOverlay(th, gtx, bp, viewport, start, end, s.hover)
		if delta != 0 {
			s.Offset += int(math.Round(float64(delta) * float64(contentLen)))
			if s.Offset < 0 {
				s.Offset = 0
			}
			if s.Offset > max {
				s.Offset = max
			}
			s.bar.poke(gtx)
		}
	}

	return layout.Dimensions{Size: viewport}
}

// ---- Scrollbar (shadcn ScrollBar · chakra hover/always + sizes) ----

// ScrollbarVariant selects visibility behavior (chakra ScrollArea
// variant). Hover is the macOS overlay default.
type ScrollbarVariant int

const (
	ScrollbarHover  ScrollbarVariant = iota // show on hover / scroll / drag; fade out
	ScrollbarAlways                         // stay visible while content overflows
)

// ScrollbarProps tunes one scrollbar paint. Zero value is macOS overlay:
// Hover fade, MD thickness, no track, FgSubtle thumb.
type ScrollbarProps struct {
	Variant ScrollbarVariant
	// Size sets thumb thickness (shared Size enum). MD ≈ 6dp.
	Size Size
	// Horizontal paints along X (shadcn orientation="horizontal").
	Horizontal bool
	// Color is an optional thumb ColorScale (.500 base → .600 hover).
	Color ColorScale
	// Scheme wins over Color for full manual thumb slots (Solid /
	// SolidHover, alpha-softened for overlay).
	Scheme *Scheme
	// ShowTrack paints a faint track behind the thumb (off = macOS).
	ShowTrack bool
}

// Scrollbar is durable drag / hover / fade state for one overlay
// scrollbar. Embed one per ScrollArea (ScrollArea already does) or
// lay out standalone when composing a custom viewport.
type Scrollbar struct {
	widget.Scrollbar
	vis       slideAnim
	wakeUntil time.Time
}

func (s *Scrollbar) poke(gtx layout.Context) {
	s.wakeUntil = gtx.Now.Add(900 * time.Millisecond)
	gtx.Execute(op.InvalidateCmd{})
}

func (o ScrollbarProps) thumbScheme(th *Theme) (rest, hover color.NRGBA) {
	if o.Scheme != nil {
		// Solid ladder reads as a thumb on light or dark panels;
		// Subtle/C100 would vanish on BgPanel.
		rest, hover = o.Scheme.Solid, o.Scheme.SolidHover
		rest.A = 180
		hover.A = 230
		return rest, hover
	}
	if o.Color != (ColorScale{}) {
		sc := o.Color.Scheme()
		rest, hover = sc.Solid, sc.SolidHover
		rest.A = 180
		hover.A = 230
		return rest, hover
	}
	// macOS-ish: muted ink, stronger when hovered.
	rest = th.Palette.FgSubtle
	rest.A = 140
	hover = th.Palette.FgMuted
	hover.A = 200
	return rest, hover
}

func scrollbarThickness(sz Size) unit.Dp {
	switch sz {
	case Size2XS:
		return 3
	case SizeXS:
		return 4
	case SizeSM:
		return 5
	case SizeLG:
		return 8
	case SizeXL:
		return 10
	case Size2XL:
		return 12
	}
	return 6 // MD
}

func scrollbarMinThumb(sz Size) unit.Dp {
	switch sz {
	case Size2XS, SizeXS:
		return 18
	case SizeSM:
		return 22
	case SizeLG:
		return 32
	case SizeXL, Size2XL:
		return 40
	}
	return 28 // MD
}

// Layout paints a scrollbar filling gtx.Constraints.Max as the track
// (major = Max along the scroll axis). Prefer ScrollArea for the usual
// overlay; use this when the app owns the track box.
//
// Returns the content-fraction delta from clicks/drags (multiply by
// content length in px and add to Offset).
func (s *Scrollbar) Layout(th *Theme, gtx layout.Context, o ScrollbarProps, viewportStart, viewportEnd float32) (layout.Dimensions, float32) {
	if viewportEnd-viewportStart >= 1 {
		return layout.Dimensions{Size: gtx.Constraints.Max}, 0
	}
	axis := layout.Vertical
	if o.Horizontal {
		axis = layout.Horizontal
	}
	return s.layoutTrack(th, gtx, o, axis, viewportStart, viewportEnd, true)
}

// layoutOverlay positions a macOS overlay thumb inside viewport and
// returns the content-fraction scroll delta.
func (s *Scrollbar) layoutOverlay(th *Theme, gtx layout.Context, o ScrollbarProps, viewport image.Point, start, end float32, areaHover bool) float32 {
	if end-start >= 1 {
		return 0
	}
	thick := gtx.Dp(scrollbarThickness(o.Size))
	pad := gtx.Dp(unit.Dp(3))
	var origin image.Point
	var track image.Point
	axis := layout.Vertical
	if o.Horizontal {
		axis = layout.Horizontal
		origin = image.Pt(pad, viewport.Y-thick-pad)
		track = image.Pt(viewport.X-2*pad, thick)
	} else {
		origin = image.Pt(viewport.X-thick-pad, pad)
		track = image.Pt(thick, viewport.Y-2*pad)
	}
	if track.X < 1 || track.Y < 1 {
		return 0
	}
	defer op.Offset(origin).Push(gtx.Ops).Pop()
	tgtx := gtx
	tgtx.Constraints.Min = track
	tgtx.Constraints.Max = track
	_, delta := s.layoutTrack(th, tgtx, o, axis, start, end, areaHover)
	return delta
}

func (s *Scrollbar) layoutTrack(th *Theme, gtx layout.Context, o ScrollbarProps, axis layout.Axis, start, end float32, areaHover bool) (layout.Dimensions, float32) {
	s.Update(gtx, axis, start, end)
	delta := s.ScrollDistance()
	if s.Dragging() {
		s.poke(gtx)
	}

	show := o.Variant == ScrollbarAlways ||
		areaHover ||
		s.IndicatorHovered() ||
		s.TrackHovered() ||
		s.Dragging() ||
		gtx.Now.Before(s.wakeUntil)
	target := float32(0)
	if show {
		target = 1
	}
	vis := s.vis.advance(gtx, target, th.Duration.Normal)
	if gtx.Now.Before(s.wakeUntil) {
		gtx.Execute(op.InvalidateCmd{})
	}
	if vis < 0.01 {
		// Still register hit targets at full opacity-zero so a fade-in
		// can start from the first hover frame.
		if !show {
			return layout.Dimensions{Size: gtx.Constraints.Max}, delta
		}
	}

	rest, hov := o.thumbScheme(th)
	col := rest
	if s.IndicatorHovered() || s.Dragging() {
		col = hov
	}
	col = fadeNRGBA(col, vis)

	convert := axis.Convert
	minAx := convert(gtx.Constraints.Min)
	maxAx := convert(gtx.Constraints.Max)
	trackLen := maxAx.X
	if trackLen < 1 {
		trackLen = minAx.X
	}
	thick := gtx.Dp(scrollbarThickness(o.Size))
	minThumb := gtx.Dp(scrollbarMinThumb(o.Size))

	// Full-area drag + track click (Gio scrollbar protocol).
	area := image.Rectangle{Max: gtx.Constraints.Max}
	func() {
		defer clip.Rect(area).Push(gtx.Ops).Pop()
		s.AddDrag(gtx.Ops)
		defer pointer.PassOp{}.Push(gtx.Ops).Pop()
		defer clip.Rect(area).Push(gtx.Ops).Pop()
		s.AddTrack(gtx.Ops)
		if o.ShowTrack && vis > 0.01 {
			tc := th.Palette.BgSubtle
			tc = fadeNRGBA(tc, vis*0.9)
			rr := thick / 2
			paint.FillShape(gtx.Ops, tc, clip.RRect{
				Rect: area, NW: rr, NE: rr, SE: rr, SW: rr,
			}.Op(gtx.Ops))
		}
	}()

	viewStart := int(math.Round(float64(start) * float64(trackLen)))
	viewEnd := int(math.Round(float64(end) * float64(trackLen)))
	thumbLen := viewEnd - viewStart
	if thumbLen < minThumb {
		thumbLen = minThumb
	}
	if viewStart+thumbLen > trackLen {
		viewStart = trackLen - thumbLen
	}
	if viewStart < 0 {
		viewStart = 0
	}
	thumbCross := thick
	thumbDims := convert(image.Pt(thumbLen, thumbCross))
	radius := thumbCross / 2

	defer op.Offset(convert(image.Pt(viewStart, 0))).Push(gtx.Ops).Pop()
	paint.FillShape(gtx.Ops, col, clip.RRect{
		Rect: image.Rectangle{Max: thumbDims},
		NW:   radius, NE: radius, SE: radius, SW: radius,
	}.Op(gtx.Ops))
	func() {
		defer pointer.PassOp{}.Push(gtx.Ops).Pop()
		defer clip.Rect(image.Rectangle{Max: thumbDims}).Push(gtx.Ops).Pop()
		s.AddIndicator(gtx.Ops)
	}()

	return layout.Dimensions{Size: gtx.Constraints.Max}, delta
}
