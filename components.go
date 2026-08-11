package lotusui

import (
	"image"
	"image/color"

	"gioui.org/font"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// maxContentDp is the fallback when Theme.PageMax is unset.
const maxContentDp = unit.Dp(920)

func fillRect(gtx layout.Context, sz image.Point, c color.NRGBA) layout.Dimensions {
	defer clip.Rect{Max: sz}.Push(gtx.Ops).Pop()
	paint.Fill(gtx.Ops, c)
	return layout.Dimensions{Size: sz}
}

// Fill paints the minimum constraint area in c.
func Fill(gtx layout.Context, c color.NRGBA) layout.Dimensions {
	return fillRect(gtx, gtx.Constraints.Min, c)
}

// LayoutPage centers content in a max-width column — one readable
// measure, generous whitespace. Reports the full available width so
// parent Flex layouts behave. Cap is th.PageMax (default 920dp), or
// th.PageMaxAt when set.
func LayoutPage(th *Theme, gtx layout.Context, content layout.Widget) layout.Dimensions {
	maxDp := th.PageMax
	if maxDp <= 0 {
		maxDp = maxContentDp
	}
	if th.PageMaxAt.Set() {
		maxDp = th.PageMaxAt.Resolve(th, gtx)
	}
	maxW := gtx.Dp(maxDp)
	contentW := gtx.Constraints.Max.X
	if contentW > maxW {
		contentW = maxW
	}
	offsetX := (gtx.Constraints.Max.X - contentW) / 2

	st := op.Offset(image.Pt(offsetX, 0)).Push(gtx.Ops)
	inner := gtx
	inner.Constraints.Min.X = contentW
	inner.Constraints.Max.X = contentW
	dims := content(inner)
	st.Pop()

	return layout.Dimensions{Size: image.Pt(gtx.Constraints.Max.X, dims.Size.Y)}
}

// cardShadow fakes a `0 2px 16px rgba(0,0,0,0.08)` box-shadow. Gio has
// no blur primitive, so three stacked translucent rounded rects — each
// one dp larger and fainter, biased downward — approximate the falloff.
// Cheap (three fills), and on the cool-gray page it reads as a real
// elevation, which is what lets the card border fade to near-nothing.
func cardShadow(gtx layout.Context, size image.Point, radius int) {
	// Each ring is the card grown UNIFORMLY by i dp and biased straight
	// down — rings stay concentric with the card, so their rounded
	// corners track the card's own radius exactly (a ring grown by g
	// keeps the same corner center when drawn at radius+g). Growing
	// asymmetrically instead is what made shadow corners visibly drift
	// off the card's curvature.
	drop := gtx.Dp(unit.Dp(2))
	for i := 3; i >= 1; i-- {
		grow := gtx.Dp(unit.Dp(i))
		alpha := uint8(16 - 4*i) // 4, 8, 12 — faintest ring outermost
		rect := image.Rect(-grow, -grow+drop, size.X+grow, size.Y+grow+drop)
		func() {
			defer clip.UniformRRect(rect, radius+grow).Push(gtx.Ops).Pop()
			paint.Fill(gtx.Ops, color.NRGBA{A: alpha})
		}()
	}
}

// SurfaceCard is the design language's standard grouping surface —
// the elevated Card at the default size. Kept as a convenience (and
// for Split's boxes); Card is the configurable family.
func SurfaceCard(th *Theme, gtx layout.Context, content layout.Widget) layout.Dimensions {
	return Card(th, CardProps{Variant: CardElevated}, content)(gtx)
}

// FloatingPanel is the floating full-height rounded panel the sidebar
// uses — white on the tinted page, soft radius, shadow elevation.
func FloatingPanel(th *Theme, gtx layout.Context, content layout.Widget) layout.Dimensions {
	r := gtx.Dp(th.Radius.LG)
	sz := gtx.Constraints.Max
	cardShadow(gtx, sz, r)
	defer clip.UniformRRect(image.Rectangle{Max: sz}, r).Push(gtx.Ops).Pop()
	paint.Fill(gtx.Ops, th.Palette.BgPanel)
	widget.Border{Color: th.Palette.Border, Width: unit.Dp(1), CornerRadius: th.Radius.LG}.Layout(gtx,
		func(gtx layout.Context) layout.Dimensions { return layout.Dimensions{Size: sz} })
	return content(gtx)
}

// seatShadow is the 1dp hairline drop shadow that SEATS a resting
// surface — the one grammar for chrome buttons, outline cards, input
// frames and raised tabs. Floating layers use cardShadow instead;
// borders never change with elevation.
func seatShadow(gtx layout.Context, size image.Point, r int) {
	sr := image.Rect(0, gtx.Dp(1), size.X, size.Y+gtx.Dp(1))
	paint.FillShape(gtx.Ops, color.NRGBA{A: 10}, clip.UniformRRect(sr, r).Op(gtx.Ops))
}

// Hairline draws a 1dp horizontal divider spanning the available width.
// Use only inside a Vertical Flex — as a Rigid child of a Horizontal
// Flex it would greedily claim all remaining width; use VerticalHairline
// there instead.
func Hairline(th *Theme) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		h := gtx.Dp(unit.Dp(1))
		return fillRect(gtx, gtx.Constraints.Constrain(image.Pt(gtx.Constraints.Max.X, h)), th.Palette.BorderSubtle)
	}
}

// ClampCorner caps a corner radius at half the smaller side — past
// that, a rounded rect's corners grow spurs instead of staying a
// clean capsule. Every fixed-radius chrome that can meet an arbitrary
// size goes through this.
func ClampCorner(r int, sz image.Point) int {
	if m := sz.X / 2; r > m {
		r = m
	}
	if m := sz.Y / 2; r > m {
		r = m
	}
	return r
}

// VerticalHairline draws a 1dp vertical divider spanning the available
// height. In an UNBOUNDED context (a Rigid child of a row measured at
// natural height) it falls back to one line-height — a divider must
// never be the child that decides the row is a mile tall.
func VerticalHairline(th *Theme) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		w := gtx.Dp(unit.Dp(1))
		h := gtx.Constraints.Max.Y
		if h > gtx.Constraints.Min.Y && h > 1<<14 {
			h = gtx.Constraints.Min.Y
			if h == 0 {
				h = gtx.Dp(16)
			}
		}
		return fillRect(gtx, gtx.Constraints.Constrain(image.Pt(w, h)), th.Palette.BorderSubtle)
	}
}

// SectionLabel renders a small, quiet caption used to introduce a group
// of related content — hierarchy from weight and color, not bold
// shouting.
func SectionLabel(th *Theme, text string) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		lbl := material.Label(th.Material, Sp(th, 12.0/16.0), text)
		lbl.Color = th.Palette.FgSubtle
		lbl.Font.Weight = font.Medium
		return lbl.Layout(gtx)
	}
}

// SVGIconButton is a small icon-only clickable using an embedded
// full-color SVG icon, with the pointer cursor and a comfortable tap
// inset. active draws the Surface2 pill behind it — the same active
// language as rows — so toggle buttons (the gear) visibly read as ON;
// hover shows the same tint as affordance.
func SVGIconButton(th *Theme, btn *widget.Clickable, icon string, size unit.Dp, active bool) layout.Widget {
	return SVGIconButtonTint(th, btn, icon, size, active, color.NRGBA{})
}

// SVGIconButtonTint is SVGIconButton for MONO icons (the plain fluent
// set), whose currentColor takes the given tint — e.g. the quiet
// TextDim pencil. Full-color icons ignore the tint.
func SVGIconButtonTint(th *Theme, btn *widget.Clickable, icon string, size unit.Dp, active bool, tint color.NRGBA) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return btn.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			m := op.Record(gtx.Ops)
			dims := layout.UniformInset(unit.Dp(6)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				d := SVGIcon(icon, size, tint)(gtx)
				defer clip.Rect(image.Rectangle{Max: d.Size}).Push(gtx.Ops).Pop()
				pointer.CursorPointer.Add(gtx.Ops)
				return d
			})
			call := m.Stop()
			dims.Size = gtx.Constraints.Constrain(dims.Size)
			if active || btn.Hovered() {
				// Active reads one step DARKER than a row's active tint,
				// so a toggled icon stays visible on a tinted row.
				fill := th.Palette.BgSubtle
				if active {
					fill = th.Palette.Border
				}
				defer clip.UniformRRect(image.Rectangle{Max: dims.Size}, gtx.Dp(th.Radius.MD)).Push(gtx.Ops).Pop()
				paint.Fill(gtx.Ops, fill)
			}
			call.Add(gtx.Ops)
			return dims
		})
	}
}

// TopBar is the screen header, a composable: a FIXED-height
// row (so the title never shifts vertically between screen states),
// the title dead-center via ScreenTitle, an optional leading widget
// (the back button) overlaid at the left — and NO background: it sits
// directly on the page.
func TopBar(th *Theme, title string, leading layout.Widget) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		h := gtx.Dp(unit.Dp(44))
		sz := gtx.Constraints.Constrain(image.Pt(gtx.Constraints.Max.X, h))
		gtx.Constraints.Min, gtx.Constraints.Max = sz, sz
		children := []layout.StackChild{
			layout.Expanded(ScreenTitle(th, title)),
		}
		if leading != nil {
			children = append(children, layout.Expanded(func(gtx layout.Context) layout.Dimensions {
				return layout.W.Layout(gtx, leading)
			}))
		}
		return layout.Stack{}.Layout(gtx, children...)
	}
}

// TitleWithIcons is an in-content / section title with trailing action
// icons — the counterpart to TopBar (which is screen chrome with an
// optional leading control). Title is LabelTitle on the left (Flexed);
// icons are Rigid with Space.XS between them only. Empty icons → title alone.
func TitleWithIcons(th *Theme, title string, icons ...layout.Widget) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		row := []layout.FlexChild{
			layout.Flexed(1, LabelTitle(th, title).Layout),
		}
		for i, ic := range icons {
			if i > 0 {
				row = append(row, layout.Rigid(HSpacer(th.Space.XS)))
			}
			row = append(row, layout.Rigid(ic))
		}
		return layout.Flex{Alignment: layout.Middle}.Layout(gtx, row...)
	}
}

// ScreenTitle renders a screen's name the app's one way — every screen
// reuses this: small (14sp semibold, a macOS-toolbar-title weight) and
// horizontally centered over the CONTENT column — it spans whatever
// width its parent hands it, so inside LayoutPage's constrained column
// it centers on the content, not the window.
func ScreenTitle(th *Theme, text string) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
		return layout.Center.Layout(gtx, LabelCardTitle(th, text).Layout)
	}
}

// FullWidth makes a button (or any widget) span the whole column.
func FullWidth(w layout.Widget) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
		return w(gtx)
	}
}

// RightAligned spans the full row width and anchors w to its right
// edge — the alignment for modal button rows. (A zero-size Flexed
// spacer does NOT work for this: Flex advances by each child's
// REPORTED size, so an empty filler collapses and the buttons drift
// left.)
func RightAligned(w layout.Widget) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
		return layout.E.Layout(gtx, w)
	}
}

// Spacer renders a fixed-height vertical gap — the app's default rhythm
// between stacked elements.
func Spacer(h unit.Dp) layout.Widget {
	return layout.Spacer{Height: h}.Layout
}

// HSpacer renders a fixed-width horizontal gap.
func HSpacer(w unit.Dp) layout.Widget {
	return layout.Spacer{Width: w}.Layout
}

// There are deliberately NO PrimaryButton/SecondaryButton/DangerButton
// helpers — one Button, composed via
// ButtonProps{Variant, Color, Loading, Disabled}. The house pairings:
// call-to-action = ButtonProps{} (solid accent); auxiliary = subtle
// neutral; destructive = solid danger.

// Disabled conditionally disables gtx — a small readability helper used
// to gate buttons on state.
func Disabled(gtx layout.Context, when bool) layout.Context {
	if when {
		return gtx.Disabled()
	}
	return gtx
}
