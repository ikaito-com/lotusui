package lotusui

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/unit"
	"gioui.org/widget"
)

// Split arranges a screen's main panes as a CAROUSEL. All the screen's boxes live on one horizontal strip behind
// an overflow-hidden viewport; the screen's state (depth) decides what
// the viewport shows:
//
//	depth 0 — box 0 takes the full width, hiding the others laterally
//	depth 1 — boxes 0 and 1 side by side, half each
//	depth n — boxes n-1 and n side by side; everything before has
//	          translated off to the left
//
// Two eased values drive every transition: box 0's width fraction
// (full ↔ half, the expand/shrink) and the strip's X translation (the
// carousel move). Boxes are always laid out at their natural width and
// only ever REVEALED or HIDDEN by the viewport's edges — content never
// re-flows mid-animation, so a disappearing box stays visually
// identical while it slides away.
type Split struct {
	frac   slideAnim // box 0's width fraction of the strip (1 = full, 0.5 = half)
	pos    slideAnim // which box index sits at the viewport's left edge
	inited bool
}

// SplitBox is the standard wrapper for a box inside a Split: the
// rounded surface card around the pane's own content, so appearing
// and disappearing panes look identical across an app.
func SplitBox(th *Theme, content layout.Widget) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return SurfaceCard(th, gtx, content)
	}
}

// VSlide is Split's VERTICAL sibling, for full-screen pivots (the
// expanded item view): two REAL screens, the over-screen sliding up
// from the bottom while the base screen is pushed up and away — the
// same grammar as the horizontal strip, turned 90°. Settled, only the
// visible screen is laid out at all. Mid-flight, both screens are laid
// out with their normal settled constraints but a DISABLED input
// context — hover pills, carets, anything input-dependent is uniformly
// off — and their recorded frames are merely translated inside an
// overflow-hidden viewport: verbatim pixels sliding, so the moving
// content can never reflow, compress, or flicker.
type VSlide struct {
	anim  slideAnim
	scrim widget.Clickable
}

// Moving reports whether the slide is mid-flight — neither screen has
// settled ownership of the viewport.
func (s *VSlide) Moving() bool { return s.anim.prog > 0 && s.anim.prog < 1 }

func (s *VSlide) Layout(gtx layout.Context, th *Theme, open bool, base, over layout.Widget) layout.Dimensions {
	target := float32(0)
	if open {
		target = 1
	}
	p := s.anim.advance(gtx, target, th.Duration.Slow)
	switch {
	case p <= 0:
		return base(gtx)
	case p >= 1:
		gtx.Constraints.Min = gtx.Constraints.Max
		return over(gtx)
	}
	h := gtx.Constraints.Max.Y
	dgtx := gtx.Disabled() // frozen frames: no input-driven visual can change mid-flight
	mBase := op.Record(gtx.Ops)
	base(dgtx)
	baseCall := mBase.Stop()
	ogtx := dgtx
	ogtx.Constraints.Min = ogtx.Constraints.Max
	mOver := op.Record(gtx.Ops)
	Fill(ogtx, th.Palette.Bg) // opaque: the arriving screen hides the departing one
	over(ogtx)
	overCall := mOver.Stop()

	defer clip.Rect(image.Rectangle{Max: gtx.Constraints.Max}).Push(gtx.Ops).Pop()
	off := op.Offset(image.Pt(0, -int(float32(h)*p))).Push(gtx.Ops)
	baseCall.Add(gtx.Ops)
	off.Pop()
	off = op.Offset(image.Pt(0, int(float32(h)*(1-p)))).Push(gtx.Ops)
	overCall.Add(gtx.Ops)
	off.Pop()
	// Topmost: absorb clicks so neither mid-flight screen is
	// interactive during the transition.
	_ = s.scrim.Clicked(gtx)
	s.scrim.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: gtx.Constraints.Max}
	})
	return layout.Dimensions{Size: gtx.Constraints.Max}
}

// LayoutSolo is the strip's FULL-WIDTH pivot: two boxes, one visible
// at a time — solo=false shows box0, solo=true slides the strip left
// so box1 occupies the whole viewport. Pure translation of
// natural-width frames (the no-reflow doctrine): both boxes keep full
// width for the entire flight.
func (s *Split) LayoutSolo(gtx layout.Context, gap unit.Dp, solo bool, box0, box1 layout.Widget) layout.Dimensions {
	if !s.inited {
		s.frac.prog, s.pos.prog, s.inited = 1, 0, true
	}
	target := float32(0)
	if solo {
		target = 1
	}
	p := s.pos.advance(gtx, target, Duration.Normal)
	switch {
	case p <= 0:
		return box0(gtx)
	case p >= 1:
		return box1(gtx)
	}
	gapPx := gtx.Dp(gap)
	w := gtx.Constraints.Max.X
	room := gtx.Dp(shadowRoom)
	m0 := op.Record(gtx.Ops)
	d0 := box0(gtx)
	c0 := m0.Stop()
	m1 := op.Record(gtx.Ops)
	box1(gtx)
	c1 := m1.Stop()
	h := d0.Size.Y
	defer clip.Rect(image.Rectangle{Min: image.Pt(-room, -room), Max: image.Pt(w+room, gtx.Constraints.Max.Y+room)}).Push(gtx.Ops).Pop()
	shift := int(p * float32(w+gapPx))
	off := op.Offset(image.Pt(-shift, 0)).Push(gtx.Ops)
	c0.Add(gtx.Ops)
	off.Pop()
	off = op.Offset(image.Pt(w+gapPx-shift, 0)).Push(gtx.Ops)
	c1.Add(gtx.Ops)
	off.Pop()
	return layout.Dimensions{Size: image.Pt(w, h)}
}

// Layout renders the strip at the given depth. boxes beyond the depth
// still get their slot on the strip (off-screen right); boxes before
// depth-1 sit off-screen left. Pass every box the screen can show —
// including ones currently hidden — so mid-animation frames render
// their content sliding, never popping.
func (s *Split) Layout(gtx layout.Context, gap unit.Dp, depth int, boxes ...layout.Widget) layout.Dimensions {
	if len(boxes) == 0 {
		return layout.Dimensions{}
	}
	if !s.inited {
		s.frac.prog, s.pos.prog, s.inited = 1, 0, true
	}
	if depth > len(boxes)-1 {
		depth = len(boxes) - 1
	}
	targetFrac := float32(1)
	if depth > 0 {
		targetFrac = 0.5
	}
	targetPos := float32(0)
	if depth > 1 {
		targetPos = float32(depth - 1)
	}

	// The shared animation clock drives both eased values, so the strip
	// moves with the same feel as every other motion in the app.
	frac := s.frac.advance(gtx, targetFrac, Duration.Normal)
	pos := s.pos.advance(gtx, targetPos, Duration.Normal)

	if frac >= 1 && pos <= 0 {
		return boxes[0](gtx) // settled home: no strip, no clip, free shadows
	}

	gapPx := gtx.Dp(gap)
	avail := gtx.Constraints.Max.X
	inner := avail - gapPx
	if inner < 0 {
		inner = 0
	}
	wHalf := inner / 2
	// Box 0's animated width; the reveal of box 1 is its complement.
	w1visible := int(float32(inner)*(1-frac) + 0.5)
	w0 := inner - w1visible
	// The gap right of box 0 fades in with the reveal so the closed
	// state has no phantom spacing.
	g01 := gapPx
	if f := (1 - frac) * 2; f < 1 {
		g01 = int(float32(gapPx)*f + 0.5)
	}

	// Strip geometry: every box at its natural width, in sequence.
	xs := make([]int, len(boxes))
	ws := make([]int, len(boxes))
	x := 0
	for i := range boxes {
		w := wHalf
		g := gapPx
		if i == 0 {
			w, g = w0, g01
		}
		xs[i], ws[i] = x, w
		x += w + g
	}
	// The strip's translation: bring box floor(pos) (interpolated
	// toward the next) to the viewport's left edge.
	i0 := int(pos)
	if i0 > len(xs)-1 {
		i0 = len(xs) - 1
	}
	tx := float32(xs[i0])
	if f := pos - float32(i0); f > 0 && i0+1 < len(xs) {
		tx += f * float32(xs[i0+1]-xs[i0])
	}

	// Measure every box that intersects the viewport (with shadowRoom
	// slack so resting shadows are never flat-cut), then draw them
	// inside the overflow-hidden clip.
	room := gtx.Dp(shadowRoom)
	type placed struct {
		call op.CallOp
		x    int
	}
	var visible []placed
	maxH := 0
	for i, box := range boxes {
		vx := xs[i] - int(tx+0.5)
		if vx+ws[i] < -room || vx > avail+room {
			continue
		}
		bgtx := gtx
		bgtx.Constraints.Min.X, bgtx.Constraints.Max.X = ws[i], ws[i]
		// Keep Min.Y so callers can ask for full-height panes (Changes /
		// Projects carousel flush with the sidebar). Cap if the parent
		// handed a min taller than this box's max.
		if bgtx.Constraints.Min.Y > bgtx.Constraints.Max.Y {
			bgtx.Constraints.Min.Y = bgtx.Constraints.Max.Y
		}
		m := op.Record(gtx.Ops)
		d := box(bgtx)
		visible = append(visible, placed{call: m.Stop(), x: vx})
		if d.Size.Y > maxH {
			maxH = d.Size.Y
		}
	}
	defer clip.Rect(image.Rect(-room, -room, avail+room, maxH+room)).Push(gtx.Ops).Pop()
	for _, p := range visible {
		st := op.Offset(image.Pt(p.x, 0)).Push(gtx.Ops)
		p.call.Add(gtx.Ops)
		st.Pop()
	}
	return layout.Dimensions{Size: gtx.Constraints.Constrain(image.Pt(avail, maxH))}
}
