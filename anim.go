package lotusui

import (
	"math"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
)

// slideAnim is the app's ONE animation clock: frame-time-based eased
// progress toward a target — fast start, gentle landing, no keyframe
// bookkeeping. Every moving surface (hover, switch, dialog, strip)
// advances through this so all motion shares one feel; the Theme
// Duration scale picks how quickly each role settles.
//
// Gio only redraws on input/resize unless something requests frames:
// while unsettled we issue op.InvalidateCmd (see gioui.org/doc
// architecture/drawing — animation).
type slideAnim struct {
	prog float32
	last time.Time
}

// advance moves progress toward target over roughly d (exponential
// ease: ~98% settled at d). d<=0 snaps.
//
// The first call after construction only primes the clock and requests
// another frame — it does not invent a dt. Inventing ~16ms on the
// priming frame jumped the mix ~35% in one paint and made hover look
// like a flash.
func (a *slideAnim) advance(gtx layout.Context, target float32, d time.Duration) float32 {
	if d <= 0 {
		a.prog = target
		a.last = gtx.Now
		return a.prog
	}
	if a.last.IsZero() {
		a.last = gtx.Now
		if a.prog != target {
			gtx.Execute(op.InvalidateCmd{})
		}
		return a.prog
	}
	dt := float32(gtx.Now.Sub(a.last).Seconds())
	a.last = gtx.Now
	if dt < 0 {
		dt = 0
	} else if dt > 0.032 {
		// Cap so a long idle before the next event doesn't teleport.
		dt = 0.032
	}
	// 4/d ≈ reach ~98% within duration (e^{-4} ≈ 0.018).
	rate := float32(4.0 / d.Seconds())
	if rate < 1 {
		rate = 1
	}
	step := 1 - float32(math.Exp(float64(-dt*rate)))
	a.prog += (target - a.prog) * step
	if delta := target - a.prog; delta < 0.002 && delta > -0.002 {
		a.prog = target
	} else {
		gtx.Execute(op.InvalidateCmd{})
	}
	return a.prog
}
