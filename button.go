package lotusui

import (
	"image"
	"image/color"
	"math"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// The Button component: one component, styled by VARIANT × SIZE,
// with disabled and loading states. Each variant carries its role's
// scheme (Default→accent, Secondary→neutral, Destructive→danger) —
// or override with any custom Scheme, e.g. one derived from a
// ColorScale — so "what the accent is" lives in one place.
//
//	Button(th, &btn, "Save", ButtonProps{})                             // solid primary
//	Button(th, &btn, "Cancel", ButtonProps{Variant: ButtonSecondary})
//	Button(th, &btn, "Delete", ButtonProps{Variant: ButtonDestructive})
//	Button(th, &btn, "Save", ButtonProps{Loading: saving})
//	Button(th, &btn, "Go", ButtonProps{Scheme: &teal})                  // any scheme

// ButtonVariant selects the button's role and chrome.
type ButtonVariant int

const (
	ButtonDefault     ButtonVariant = iota // solid primary — the call to action
	ButtonSecondary                        // muted solid for auxiliary actions
	ButtonDestructive                      // solid destructive — Delete-class only
	ButtonOutline                          // bordered, quiet ink, fills faintly on hover
	ButtonGhost                            // no chrome until hovered
	ButtonLink                             // primary ink only; underlines on hover
)

// Size selects a component's size preset — shared by Button,
// Checkbox and Switch, so "SizeSM" means the same step everywhere.
type Size int

const (
	SizeMD Size = iota // the default
	Size2XS
	SizeXS
	SizeSM
	SizeLG
	SizeXL
	Size2XL
)

// ButtonProps are the button's props. A button's label is always
// centered — a full-width action with a start-aligned label is a menu
// row, not a button; use DropdownMenuItem for those.
type ButtonProps struct {
	Variant ButtonVariant
	Size    Size
	// Color re-colors the variant from a ColorScale — the named scales
	// (Teal, Pink, …) or any ScaleFrom(anchor). One anchor in, the
	// whole interaction ladder out (.500 base, .600 hover, .700
	// pressed), so a custom color can never fall out of step with its
	// hover and pressed states.
	Color ColorScale
	// Scheme, when set, wins over Color and the variant's role scheme:
	// full manual control of every slot — e.g. Teal.SoftScheme() for
	// the pastel construction, or a hand-built Scheme.
	Scheme *Scheme
	// IconStart / IconEnd name embedded icons (typed constants or your
	// registered ones) rendered beside the label, tinted with the
	// button's own ink.
	IconStart, IconEnd string
	// Rounded renders the pill form — full-round corners.
	Rounded bool
	// Attached marks the sides that sit against a neighbor in a
	// ButtonGroup: those corners render square and the seat shadow
	// drops. ButtonGroup sets this for its children.
	Attached AttachedEdges
	Loading  bool // shows a spinner (label keeps the width), input off
	// LoadingText, when set, replaces the label while Loading instead
	// of hiding it under the spinner.
	LoadingText string
	Disabled    bool
}

func (o ButtonProps) scheme(th *Theme) Scheme {
	if o.Scheme != nil {
		return *o.Scheme
	}
	if o.Color != (ColorScale{}) {
		return o.Color.Scheme()
	}
	switch o.Variant {
	case ButtonSecondary, ButtonOutline, ButtonGhost:
		return th.Palette.Neutral()
	case ButtonDestructive:
		return th.Palette.DangerScheme()
	}
	return th.Palette.Accent()
}

// disabledColors resolves the quiet same-hue palette for a disabled
// button: a BRIGHTER step of the button's own scale (.200/.300 fill
// for saturated solids; Subtle for soft pastels; lighter ink for
// outline/ghost/link) — never a darker press-step, never a foreign grey.
func (o ButtonProps) disabledColors(th *Theme, sc Scheme, fg color.NRGBA) (solid, subtle, ink, outline color.NRGBA) {
	solid, subtle, ink, outline = sc.Solid, sc.Subtle, fg, sc.Outline
	if s, ok := o.colorScale(); ok {
		// Saturated Color.Scheme ladder: .500 → brighter .200 fill.
		solid, subtle = s.C200, s.C50
		outline = s.C300
		switch o.Variant {
		case ButtonDefault, ButtonSecondary, ButtonDestructive:
			ink = s.C600
		default:
			ink = s.C300
		}
		return solid, subtle, ink, outline
	}
	// Soft schemes (dark ink on a light fill): lift to Subtle — a
	// brighter tint of the same hue (SoftScheme's .50).
	if softScheme(sc) {
		solid = sc.Subtle
		if solid == (color.NRGBA{}) || lum(solid) <= lum(sc.Solid) {
			// Neutral's Subtle equals Solid — step up to the panel white.
			solid = th.Palette.BgPanel
		}
		ink = sc.OnSubtle
		if ink == (color.NRGBA{}) {
			ink = sc.OnSolid
		}
		subtle = sc.Subtle
		outline = sc.Outline
		switch o.Variant {
		case ButtonDefault, ButtonSecondary, ButtonDestructive:
			// soft chip: brighter fill, softer OnSubtle ink
		case ButtonLink:
			ink = th.BrandScale.C300
		default:
			ink = th.Palette.FgMuted
		}
		return solid, subtle, ink, outline
	}
	// Saturated hand schemes / danger without a scale: mute into the
	// brighter Subtle family of the same scheme.
	solid = sc.Subtle
	if solid == (color.NRGBA{}) {
		solid = th.Palette.BgMuted
	}
	ink = sc.OnSubtle
	if ink == (color.NRGBA{}) {
		ink = th.Palette.FgMuted
	}
	subtle = sc.Subtle
	outline = sc.SubtleHover
	if outline == (color.NRGBA{}) {
		outline = sc.Outline
	}
	return solid, subtle, ink, outline
}

// colorScale is the ColorScale the button is painted from, when the
// Color prop set one. Role schemes (accent / neutral / danger) and
// hand-built Scheme values have no recoverable scale — disabledColors
// walks their Soft/Subtle slots instead.
func (o ButtonProps) colorScale() (ColorScale, bool) {
	if o.Color != (ColorScale{}) {
		return o.Color, true
	}
	return ColorScale{}, false
}

// softScheme reports a pastel scheme: dark ink on a light solid fill
// (SoftScheme), as opposed to white-on-saturated (Scheme).
func softScheme(sc Scheme) bool {
	return lum(sc.OnSolid) < lum(sc.Solid)
}

func lum(c color.NRGBA) int { return int(c.R) + int(c.G) + int(c.B) }

// metrics returns the size preset: label sp ratio and paddings.
func (o ButtonProps) metrics() (ratio float32, v, h unit.Dp) {
	switch o.Size {
	case Size2XS:
		return 11.0 / 16.0, 3, 8
	case SizeXS:
		return 12.0 / 16.0, 4, 10
	case SizeSM:
		return 13.0 / 16.0, 6, 12
	case SizeLG:
		return 14.0 / 16.0, 10, 24
	case SizeXL:
		return 16.0 / 16.0, 12, 28
	case Size2XL:
		return 18.0 / 16.0, 14, 32
	}
	return 14.0 / 16.0, 8, 16
}

// Button lays out one button. The MD size matches the app's button
// spec (8dp corners, 14sp semibold, 12/20 padding).
func Button(th *Theme, btn *widget.Clickable, label string, o ButtonProps) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		// Loading disables the button's INPUT, not its motion: the
		// spinner's frame requests must go through the still-enabled
		// parent context, because a disabled context blocks commands.
		if o.Loading {
			gtx.Execute(op.InvalidateCmd{})
		}
		if o.Disabled || o.Loading {
			gtx = gtx.Disabled()
		}
		sc := o.scheme(th)
		ratio, vPad, hPad := o.metrics()
		if label == "" && o.IconStart != "" {
			// Icon-only button: square padding, no label slot.
			hPad = vPad
		}
		return btn.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			// Interaction state is read HERE, inside the closure:
			// Clickable.Layout drains the event queue before calling it,
			// so this frame paints the state this frame's events
			// produced. Reading before Layout renders one event late —
			// a release only showed once the pointer moved again.
			hovered := btn.Hovered() && !o.Loading && !o.Disabled
			pressed := btn.Pressed() && !o.Loading && !o.Disabled
			// Hover eases on Duration.Fast; pressed snaps paint to Active
			// while keeping the clock at 1 for a smooth release.
			target := float32(0)
			if hovered || pressed {
				target = 1
			}
			mix := th.hoverToward(gtx, btn, target)
			// The interaction ladder: base → Hover → Active are scale
			// steps (.500/.600/.700 in web terms). Hand-built schemes
			// without the step fields fall back to an arithmetic shade.
			solidBg := steppedMix(sc.Solid, sc.SolidHover, sc.SolidActive, mix, pressed)
			subtleBg := steppedMix(sc.Subtle, sc.SubtleHover, sc.SubtleActive, mix, pressed)
			fg := sc.OnSolid
			outline := sc.Outline
			switch o.Variant {
			case ButtonDefault, ButtonSecondary, ButtonDestructive:
			case ButtonLink:
				fg = th.Palette.Accent().OnSubtle // primary ink, like a link
			default:
				fg = sc.OnSubtle
			}
			// Disabled walks a BRIGHTER step of the button's own scale —
			// same hue, quieter chrome — never a darker press-step and
			// never a foreign grey. Soft fills lift to Subtle; saturated
			// Color fills mute to .200; outline/ghost/link ink softens.
			if o.Disabled {
				solidBg, subtleBg, fg, outline = o.disabledColors(th, sc, fg)
			}
			// Pressing nudges the button 1dp down — the tactile
			// micro-interaction; dims are unaffected so layout is stable.
			if pressed {
				defer op.Offset(image.Pt(0, gtx.Dp(1))).Push(gtx.Ops).Pop()
			}
			m := op.Record(gtx.Ops)
			content := layout.Inset{Top: vPad, Bottom: vPad, Left: hPad, Right: hPad}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				txt := label
				if o.Loading && o.LoadingText != "" {
					txt = o.LoadingText
				}
				l := material.Label(th.Material, Sp(th, ratio), txt)
				l.Color = fg
				l.Font.Weight = font.Medium
				l.MaxLines = 1
				l.Alignment = text.Middle
				iconSz := unit.Dp(float32(gtx.Metric.SpToDp(Sp(th, ratio))) + 2)
				if txt == "" && o.IconStart != "" && !o.Loading {
					// Icon-only: the icon IS the content.
					return SVGIcon(o.IconStart, iconSz, fg)(gtx)
				}
				var row []layout.FlexChild
				if o.Loading && o.LoadingText != "" {
					row = append(row, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						sp := gtx.Dp(14)
						spinner(gtx, fg, sp)
						return layout.Dimensions{Size: image.Pt(sp, sp)}
					}), layout.Rigid(HSpacer(th.Space.SM)))
				} else if o.IconStart != "" {
					row = append(row, layout.Rigid(SVGIcon(o.IconStart, iconSz, fg)), layout.Rigid(HSpacer(th.Space.SM)))
				}
				if len(row) > 0 || o.IconEnd != "" {
					row = append(row, layout.Rigid(l.Layout))
					if o.IconEnd != "" && !o.Loading {
						row = append(row, layout.Rigid(HSpacer(th.Space.SM)), layout.Rigid(SVGIcon(o.IconEnd, iconSz, fg)))
					}
					return layout.Flex{Alignment: layout.Middle}.Layout(gtx, row...)
				}
				if !o.Loading {
					return l.Layout(gtx)
				}
				// Loading without LoadingText: the label keeps the
				// button's width but paints invisible; the spinner draws
				// centered over it.
				l.Color.A = 0
				tdims := l.Layout(gtx)
				sp := gtx.Dp(16)
				off := op.Offset(image.Pt((tdims.Size.X-sp)/2, (tdims.Size.Y-sp)/2)).Push(gtx.Ops)
				spinner(gtx, fg, sp)
				off.Pop()
				return tdims
			})
			call := m.Stop()

			// Expand to Min (ButtonGroup stretches siblings to one
			// height) and center the label/icon in the spare room —
			// shadcn's items-stretch + flex items-center.
			sz := gtx.Constraints.Constrain(content.Size)
			ox, oy := (sz.X-content.Size.X)/2, (sz.Y-content.Size.Y)/2
			dims := layout.Dimensions{Size: sz, Baseline: content.Baseline + oy}

			r := gtx.Dp(th.Radius.MD)
			if o.Rounded {
				r = dims.Size.Y / 2
			}
			r = ClampCorner(r, dims.Size)
			rr := attachedRRect(dims.Size, r, o.Attached)
			attached := o.Attached != (AttachedEdges{})
			switch o.Variant {
			case ButtonDefault, ButtonSecondary, ButtonDestructive, ButtonOutline:
				if !attached {
					seatShadow(gtx, dims.Size, r)
				}
			}
			defer rr.Push(gtx.Ops).Pop()
			if !o.Disabled && !o.Loading {
				pointer.CursorPointer.Add(gtx.Ops)
			}
			border := func(col color.NRGBA) {
				if attached {
					// The clipped-corner border: stroke the same shape the
					// clip uses, so square corners stay fully drawn.
					paint.FillShape(gtx.Ops, col, clip.Stroke{Path: rr.Path(gtx.Ops), Width: float32(gtx.Dp(1)) * 2}.Op())
					return
				}
				// The border follows the SAME radius the fill was clipped
				// to (r is already clamped), so a Rounded button's outline
				// is a pill too instead of MD corners inside a capsule.
				widget.Border{Color: col, Width: unit.Dp(1), CornerRadius: unit.Dp(float32(r) / gtx.Metric.PxPerDp)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Dimensions{Size: dims.Size}
				})
			}
			switch o.Variant {
			case ButtonDefault, ButtonSecondary, ButtonDestructive:
				paint.Fill(gtx.Ops, solidBg)
			case ButtonOutline:
				if pressed {
					paint.Fill(gtx.Ops, subtleBg)
				} else if mix > 0.01 {
					// Fade the hover fill in via alpha (Gio MulAlpha /
					// PushOpacity model) — never lerp RGB from
					// transparent black, which flashes dark mid-fade.
					paint.Fill(gtx.Ops, fadeNRGBA(subtleTarget(sc), mix))
				}
				border(outline)
			case ButtonGhost:
				if pressed {
					paint.Fill(gtx.Ops, subtleBg)
				} else if mix > 0.01 {
					paint.Fill(gtx.Ops, fadeNRGBA(subtleTarget(sc), mix))
				}
			case ButtonLink:
				// ink only; hover underlines, like the link it imitates.
				if pressed || mix > 0.01 {
					ulFg := fg
					if !pressed {
						ulFg.A = uint8(float32(fg.A)*mix + 0.5)
					}
					lh := gtx.Dp(1)
					ul := image.Rect(gtx.Dp(hPad)+ox, dims.Size.Y-gtx.Dp(vPad)-oy+lh, dims.Size.X-gtx.Dp(hPad)-ox, dims.Size.Y-gtx.Dp(vPad)-oy+2*lh)
					paint.FillShape(gtx.Ops, ulFg, clip.Rect(ul).Op())
				}
			}
			if gtx.Focused(btn) {
				// Keyboard focus: a visible ring in the theme's
				// FocusRing token — never color alone. The ring follows
				// the SAME clamped radius (and attached corners) the
				// fill was clipped to, so pills get a pill ring and
				// attached edges stay square.
				if attached {
					paint.FillShape(gtx.Ops, th.Palette.FocusRing, clip.Stroke{Path: rr.Path(gtx.Ops), Width: float32(gtx.Dp(2)) * 2}.Op())
				} else {
					widget.Border{Color: th.Palette.FocusRing, Width: unit.Dp(2), CornerRadius: unit.Dp(float32(r) / gtx.Metric.PxPerDp)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Dimensions{Size: dims.Size}
					})
				}
			}
			off := op.Offset(image.Pt(ox, oy)).Push(gtx.Ops)
			call.Add(gtx.Ops)
			off.Pop()
			return dims
		})
	}
}

// steppedMix resolves an interaction ladder with a 0…1 hover mix;
// pressed snaps to Active. Zero step fields fall back to shading.
func steppedMix(base, hover, active color.NRGBA, mix float32, pressed bool) color.NRGBA {
	zero := color.NRGBA{}
	if pressed {
		if active != zero {
			return active
		}
		return shade(shade(base))
	}
	h := hover
	if h == zero {
		h = shade(base)
	}
	return lerpNRGBA(base, h, mix)
}

// subtleTarget is the full hover fill for outline/ghost (transparent→this).
func subtleTarget(sc Scheme) color.NRGBA {
	if sc.SubtleHover != (color.NRGBA{}) {
		return sc.SubtleHover
	}
	if sc.Subtle != (color.NRGBA{}) {
		return shade(sc.Subtle)
	}
	return shade(sc.Solid)
}

// stepped resolves an interaction ladder: pressed wins over hovered;
// zero step fields (hand-built schemes) fall back to shading the base.
func stepped(base, hover, active color.NRGBA, hovered, pressed bool) color.NRGBA {
	mix := float32(0)
	if hovered {
		mix = 1
	}
	return steppedMix(base, hover, active, mix, pressed)
}

// shade darkens a fill slightly — the fallback interaction feedback
// for schemes built without explicit Hover/Active steps.
func shade(c color.NRGBA) color.NRGBA {
	dim := func(v uint8) uint8 {
		return uint8(int(v) * 94 / 100)
	}
	return color.NRGBA{R: dim(c.R), G: dim(c.G), B: dim(c.B), A: c.A}
}

// spinner draws the loading indicator: a three-quarter arc rotating
// with the frame clock, stroked in the button's own text color.
func spinner(gtx layout.Context, col color.NRGBA, px int) {
	r := float32(px) / 2
	sw := float32(gtx.Dp(2))
	t := float64(gtx.Now.UnixMilli()%900) / 900
	start := t * 2 * math.Pi
	center := f32.Pt(r, r)
	rad := r - sw // stroke centerline radius: the 2dp stroke stays inside the box
	pen := f32.Pt(
		center.X+rad*float32(math.Cos(start)),
		center.Y+rad*float32(math.Sin(start)),
	)
	var p clip.Path
	p.Begin(gtx.Ops)
	p.MoveTo(pen)
	// Arc's focus points are RELATIVE to the pen; for a circle both
	// foci sit at the center, so pass center-minus-pen twice.
	f := f32.Pt(center.X-pen.X, center.Y-pen.Y)
	p.Arc(f, f, 2*math.Pi*0.75)
	paint.FillShape(gtx.Ops, col, clip.Stroke{Path: p.End(), Width: sw}.Op())
	// Frame scheduling happens in Button, on the enabled context — a
	// disabled gtx (which a loading button always has) drops commands.
}
