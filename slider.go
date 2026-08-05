package lotusui

import (
	"image"

	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

// Slider is the draggable value control: a track, the filled range in
// the brand solid, a round thumb. Value is a fraction in [0, 1] —
// map it to your domain at the call site; Step, when non-zero, snaps
// the fraction to multiples (e.g. 0.25 for quarters).
//
// Values, when non-empty, switches to MULTI-THUMB mode: one thumb per
// entry, kept ordered, the fill spanning first to last (two entries
// is the range slider). Vertical rotates the axis — cap the height
// from the call site; the value grows upward.
type Slider struct {
	Value    float32
	Values   []float32
	Step     float32
	Vertical bool
	Disabled bool
	track    int // last laid-out track length, for px→fraction
	dragging bool
	active   int // the thumb a multi drag owns
}

func (s *Slider) clamp(v float32) float32 {
	if s.Step > 0 {
		steps := float32(int(v/s.Step + 0.5))
		v = steps * s.Step
	}
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// frac converts a pointer position along the axis to a fraction.
func (s *Slider) frac(pos float32, thumb int) float32 {
	f := (pos - float32(thumb)/2) / float32(s.track-thumb)
	if s.Vertical {
		f = 1 - f
	}
	return s.clamp(f)
}

func (s *Slider) Layout(th *Theme, gtx layout.Context) layout.Dimensions {
	thumb := gtx.Dp(16)
	trackH := gtx.Dp(6)
	multi := len(s.Values) > 0

	length := gtx.Constraints.Max.X
	if s.Vertical {
		length = gtx.Constraints.Max.Y
		if length > 1<<15 { // unbounded: a sane default height
			length = gtx.Dp(160)
		}
	}
	breadth := thumb + gtx.Dp(4)

	// Events first: press or drag sets the value from the pointer's
	// axis position — the same frame paints it. A multi press claims
	// the nearest thumb and owns it for the drag.
	if !s.Disabled {
		for {
			ev, ok := gtx.Event(pointer.Filter{Target: s, Kinds: pointer.Press | pointer.Drag | pointer.Release | pointer.Cancel})
			if !ok {
				break
			}
			e, isPtr := ev.(pointer.Event)
			if !isPtr {
				continue
			}
			pos := e.Position.X
			if s.Vertical {
				pos = e.Position.Y
			}
			switch e.Kind {
			case pointer.Press:
				s.dragging = true
				if multi && s.track > 0 {
					f := s.frac(pos, thumb)
					s.active = 0
					best := float32(2)
					for i, v := range s.Values {
						d := v - f
						if d < 0 {
							d = -d
						}
						if d < best {
							best, s.active = d, i
						}
					}
				}
			case pointer.Release, pointer.Cancel:
				s.dragging = false
			}
			if (e.Kind == pointer.Press || e.Kind == pointer.Drag) && s.track > 0 {
				f := s.frac(pos, thumb)
				if multi {
					// The dragged thumb stays between its neighbors.
					if s.active > 0 && f < s.Values[s.active-1] {
						f = s.Values[s.active-1]
					}
					if s.active < len(s.Values)-1 && f > s.Values[s.active+1] {
						f = s.Values[s.active+1]
					}
					s.Values[s.active] = f
				} else {
					s.Value = f
				}
			}
		}
	}
	s.track = length

	size := image.Pt(length, breadth)
	if s.Vertical {
		size = image.Pt(breadth, length)
	}
	defer clip.Rect(image.Rectangle{Max: size}).Push(gtx.Ops).Pop()
	event.Op(gtx.Ops, s)
	if !s.Disabled {
		pointer.CursorPointer.Add(gtx.Ops)
	}

	fillCol, thumbCol := th.Palette.BrandSolid, th.Palette.BrandFg
	if s.Disabled {
		// Brighter same-hue mute — Soft fill lifts to BrandSubtle;
		// thumb softens toward .300 (never darker/grey).
		fillCol, thumbCol = th.Palette.BrandSubtle, th.BrandScale.C300
	}

	// px positions the leading edge of a thumb at fraction v.
	px := func(v float32) int {
		if s.Vertical {
			v = 1 - v
		}
		return int(v * float32(length-thumb))
	}

	// along builds an a..b rectangle on the axis at track breadth.
	along := func(a, b int) image.Rectangle {
		ty := (breadth - trackH) / 2
		if s.Vertical {
			return image.Rect(ty, a, ty+trackH, b)
		}
		return image.Rect(a, ty, b, ty+trackH)
	}
	paint.FillShape(gtx.Ops, th.Palette.BgMuted, clip.UniformRRect(along(0, length), trackH/2).Op(gtx.Ops))

	var lo, hi int
	switch {
	case multi:
		lo, hi = px(s.Values[0])+thumb/2, px(s.Values[len(s.Values)-1])+thumb/2
	case s.Vertical:
		lo, hi = px(s.Value)+thumb/2, length
	default:
		lo, hi = 0, px(s.Value)+thumb/2
	}
	if lo > hi {
		lo, hi = hi, lo
	}
	paint.FillShape(gtx.Ops, fillCol, clip.UniformRRect(along(lo, hi), trackH/2).Op(gtx.Ops))

	drawThumb := func(v float32) {
		t := px(v)
		var tr image.Rectangle
		if s.Vertical {
			tr = image.Rect((breadth-thumb)/2, t, (breadth+thumb)/2, t+thumb)
		} else {
			tr = image.Rect(t, (breadth-thumb)/2, t+thumb, (breadth+thumb)/2)
		}
		paint.FillShape(gtx.Ops, thumbCol, clip.UniformRRect(tr, thumb/2).Op(gtx.Ops))
		inner := tr.Inset(gtx.Dp(2))
		paint.FillShape(gtx.Ops, th.Palette.BgPanel, clip.UniformRRect(inner, inner.Dx()/2).Op(gtx.Ops))
	}
	if multi {
		for _, v := range s.Values {
			drawThumb(v)
		}
	} else {
		drawThumb(s.Value)
	}

	return layout.Dimensions{Size: size}
}
