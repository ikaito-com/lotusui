package lotusui

import (
	"image"

	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
)

// Switch is the toggle: a rounded track with a sliding
// thumb, accent-colored when on, animated on the app's shared clock.
// Value is the state; clicking flips it.
type Switch struct {
	Value    bool
	Size     Size
	Disabled bool // dimmed, swallows clicks — a fact, not a choice
	Invalid  bool // danger ring — save-time validation
	btn      widget.Clickable
	anim     slideAnim
}

// trackDp maps the shared size presets onto track width × height.
func (s *Switch) trackDp() (w, h unit.Dp) {
	switch s.Size {
	case Size2XS:
		return 20, 11
	case SizeXS:
		return 24, 13
	case SizeSM:
		return 28, 16
	case SizeLG:
		return 40, 22
	case SizeXL:
		return 48, 26
	case Size2XL:
		return 56, 30
	}
	return 32, 18
}

func (s *Switch) Layout(th *Theme, gtx layout.Context) layout.Dimensions {
	if s.Disabled {
		gtx = gtx.Disabled()
	}
	if s.btn.Clicked(gtx) && !s.Disabled {
		s.Value = !s.Value
	}
	target := float32(0)
	if s.Value {
		target = 1
	}
	p := s.anim.advance(gtx, target, th.Duration.Fast)

	wDp, hDp := s.trackDp()
	w, h := gtx.Dp(wDp), gtx.Dp(hDp)
	pad := gtx.Dp(unit.Dp(2))
	thumb := h - 2*pad

	return s.btn.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		sz := gtx.Constraints.Constrain(image.Pt(w, h))
		defer clip.UniformRRect(image.Rectangle{Max: sz}, h/2).Push(gtx.Ops).Pop()
		if !s.Disabled {
			pointer.CursorPointer.Add(gtx.Ops)
		}
		// Track: accent ink when on, quiet border-grey when off — the
		// tween blends the felt state via thumb position, not color.
		track := th.Palette.Border
		if s.Value {
			track = th.Palette.BrandFg
		}
		if s.Disabled {
			track = th.Palette.BgEmphasized
			if s.Value {
				// Brighter same-hue mute — BrandFg → .200, not darker.
				track = th.BrandScale.C200
			}
		}
		paint.Fill(gtx.Ops, track)
		if s.Invalid {
			widget.Border{Color: th.Palette.Danger, Width: unit.Dp(1), CornerRadius: hDp / 2}.Layout(gtx,
				func(gtx layout.Context) layout.Dimensions { return layout.Dimensions{Size: sz} })
		}
		x := pad + int(p*float32(w-thumb-2*pad))
		thumbRect := image.Rect(x, pad, x+thumb, pad+thumb)
		paint.FillShape(gtx.Ops, th.Palette.BgPanel, clip.Ellipse(thumbRect).Op(gtx.Ops))
		return layout.Dimensions{Size: sz}
	})
}
