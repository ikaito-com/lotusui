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
	"gioui.org/widget/material"
)

// Tooltip attaches a hover label to any widget: hold the pointer over
// the child for a beat and the label floats in beneath it, inverted
// chrome (dark surface, light ink). The pointer area PASSES events
// through — the child keeps its own interactions.
//
// Layout may be called multiple times per frame on the same Tooltip;
// each call is a distinct site and only the hovered site paints.
//
// TooltipSide selects where the label floats relative to the child.
type TooltipSide int

const (
	TooltipBottom TooltipSide = iota
	TooltipTop
	TooltipLeft
	TooltipRight
)

type tipTrig struct{}

type Tooltip struct {
	// Side positions the label; bottom is the default.
	Side TooltipSide

	since  time.Time
	sites  layoutSites
	trigs  []*tipTrig
	over   []bool
	active int
}

// tooltipDelay is how long the pointer must rest before the label
// shows — long enough to never flicker during travel.
const tooltipDelay = 450 * time.Millisecond

func (t *Tooltip) nextSite(now time.Time) (idx int, tag *tipTrig) {
	idx = t.sites.next(now)
	for len(t.trigs) <= idx {
		t.trigs = append(t.trigs, new(tipTrig))
		t.over = append(t.over, false)
	}
	return idx, t.trigs[idx]
}

func (t *Tooltip) anyOver() bool {
	for _, o := range t.over {
		if o {
			return true
		}
	}
	return false
}

func (t *Tooltip) Layout(th *Theme, gtx layout.Context, text string, child layout.Widget) layout.Dimensions {
	// A throwaway pass has no site: measure the child only.
	if inMeasurePass() {
		return child(gtx)
	}
	idx, tag := t.nextSite(gtx.Now)

	for {
		ev, ok := gtx.Event(pointer.Filter{Target: tag, Kinds: pointer.Enter | pointer.Leave | pointer.Cancel})
		if !ok {
			break
		}
		if e, isPtr := ev.(pointer.Event); isPtr {
			switch e.Kind {
			case pointer.Enter:
				t.over[idx] = true
				t.active = idx
				t.since = gtx.Now
			default:
				t.over[idx] = false
			}
		}
	}

	dims := child(gtx)

	// A pass-through input area over the child: hover intent without
	// stealing the child's events.
	func() {
		defer pointer.PassOp{}.Push(gtx.Ops).Pop()
		defer clip.Rect(image.Rectangle{Max: dims.Size}).Push(gtx.Ops).Pop()
		event.Op(gtx.Ops, tag)
	}()

	if t.anyOver() && t.active == idx {
		if gtx.Now.Sub(t.since) < tooltipDelay {
			// One wake-up AT the reveal, not a frame every vsync
			// while the pointer rests: Gio schedules timed frames.
			gtx.Execute(op.InvalidateCmd{At: t.since.Add(tooltipDelay)})
		} else {
			Floating(gtx, func(gtx layout.Context) layout.Dimensions {
				m := op.Record(gtx.Ops)
				lgtx := gtx
				lgtx.Constraints.Min = image.Point{}
				ld := layout.Inset{Top: 6, Bottom: 6, Left: 10, Right: 10}.Layout(lgtx, func(gtx layout.Context) layout.Dimensions {
					l := material.Label(th.Material, Sp(th, 12.0/16.0), text)
					l.Color = th.Palette.FgInverted
					l.MaxLines = 1
					return l.Layout(gtx)
				})
				call := m.Stop()
				gap := gtx.Dp(6)
				var at image.Point
				switch t.Side {
				case TooltipTop:
					at = image.Pt((dims.Size.X-ld.Size.X)/2, -ld.Size.Y-gap)
				case TooltipLeft:
					at = image.Pt(-ld.Size.X-gap, (dims.Size.Y-ld.Size.Y)/2)
				case TooltipRight:
					at = image.Pt(dims.Size.X+gap, (dims.Size.Y-ld.Size.Y)/2)
				default:
					at = image.Pt((dims.Size.X-ld.Size.X)/2, dims.Size.Y+gap)
				}
				defer op.Offset(at).Push(gtx.Ops).Pop()
				// Clip then paint.Fill — Fill(gtx,…) paints Constraints.Min,
				// which is often 0×0 here and would leave only the ink visible.
				defer clip.UniformRRect(image.Rectangle{Max: ld.Size}, gtx.Dp(th.Radius.SM)).Push(gtx.Ops).Pop()
				paint.Fill(gtx.Ops, th.Palette.BgInverted)
				call.Add(gtx.Ops)
				return ld
			})
		}
	}
	return dims
}
