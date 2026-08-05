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
type slideAnim struct {
	prog float32
	last time.Time
}

// advance moves progress toward target over roughly d (exponential
// ease: ~98% settled at d). d<=0 snaps. dt is capped so the first
// frame after a long idle doesn't teleport to the end.
func (a *slideAnim) advance(gtx layout.Context, target float32, d time.Duration) float32 {
	if d <= 0 {
		a.prog = target
		a.last = gtx.Now
		return a.prog
	}
	dt := float32(0.016)
	if !a.last.IsZero() {
		dt = float32(gtx.Now.Sub(a.last).Seconds())
		if dt > 0.032 {
			dt = 0.032
		}
	}
	a.last = gtx.Now
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
