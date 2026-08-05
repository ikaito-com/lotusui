package lotusui

import (
	"image"

	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
)

// Dialog is the one overlay primitive, modeled on Chakra-UI's Dialog
// : a dimmed scrim over the ENTIRE window with a
// width-capped SurfaceCard centered on it. Like Chakra, it wraps
// arbitrary content and knows nothing about what's inside; visibility
// stays with the caller (the isOpen/onClose contract) — don't call
// Layout when closed.
//
// onClose mirrors Chakra's closeOnOverlayClick: pass the caller's close
// func to let a backdrop click dismiss (glance-and-dismiss modals), or
// nil to absorb outside clicks without dismissing (typed confirmations —
// anywhere an accidental dismissal costs more than it saves).
//
// Lay the modal out at WINDOW constraints (your app shell's top
// layer / portal) — laid out inside a content column it would inherit
// that column's constraints, which is exactly the "scrim only covers
// part of the window, card off-center" bug this type exists to end.
type Dialog struct {
	// Size picks a width preset (Size2XS 280dp … Size2XL 840dp; SizeMD,
	// the default, is 480dp). Width, when non-zero, overrides it.
	// Sizes / Widths, when Set(), override Size / Width per breakpoint.
	Size   Size
	Sizes  ResponsiveSize
	Width  unit.Dp
	Widths ResponsiveDp
	// HideClose suppresses the corner ✕ on dismissable dialogs —
	// backdrop and Escape still work through onClose.
	HideClose bool

	scrimClick widget.Clickable
	cardClick  widget.Clickable
	closeBtn   widget.Clickable
	anim       slideAnim
}

// Appear restarts the entrance animation — call it when your isOpen
// transitions from closed to open (not every frame). The zero value
// already animates its first open.
func (m *Dialog) Appear() { m.anim.prog = 0 }

const dialogDefaultWidth = unit.Dp(480)

// dialogWidthDp maps a Size preset to dialog card width (dp).
func dialogWidthDp(sz Size) unit.Dp {
	switch sz {
	case Size2XS:
		return 280
	case SizeXS:
		return 320
	case SizeSM:
		return 400
	case SizeLG:
		return 600
	case SizeXL:
		return 720
	case Size2XL:
		return 840
	default:
		return dialogDefaultWidth
	}
}

func (m *Dialog) resolveWidth(th *Theme, gtx layout.Context) unit.Dp {
	idx := th.BreakpointIndex(gtx)
	if m.Widths.Set() {
		return m.Widths.ResolveAt(th, idx)
	}
	if m.Width != 0 {
		return m.Width
	}
	sz := m.Size
	if m.Sizes.Set() {
		sz = m.Sizes.ResolveAt(th, idx)
	}
	return dialogWidthDp(sz)
}

func (m *Dialog) Layout(th *Theme, gtx layout.Context, onClose func(), content layout.Widget) layout.Dimensions {
	// Clicked must be drained every frame even when onClose is nil, or
	// queued clicks would fire the moment a dismissable modal reuses the
	// widget. The scrim's click area is what absorbs pointer input so
	// nothing falls through to the screen underneath.
	if m.scrimClick.Clicked(gtx) && onClose != nil {
		onClose()
		return layout.Dimensions{}
	}
	// The corner ✕ mirrors the backdrop contract: dismissable dialogs
	// get the standard affordance, absorb-only dialogs get none.
	if m.closeBtn.Clicked(gtx) && onClose != nil {
		onClose()
		return layout.Dimensions{}
	}
	// Escape dismisses exactly like a backdrop click — same onClose
	// contract, so absorb-only modals (onClose nil) ignore it too.
	for {
		ev, ok := gtx.Event(key.Filter{Name: key.NameEscape})
		if !ok {
			break
		}
		if e, isKey := ev.(key.Event); isKey && e.State == key.Press && onClose != nil {
			onClose()
			return layout.Dimensions{}
		}
	}
	// Entrance motion on the shared clock: the zero value animates on
	// its first open; call Appear() on later open TRANSITIONS to play
	// it again. Detecting "fresh open" by wall-clock gaps is wrong in
	// an event-driven renderer — idle means no frames, and the first
	// frame after idle must not look like a re-open.
	appear := m.anim.advance(gtx, 1, th.Duration.Slow)
	// Drained and ignored: the card registers its own click area so a
	// click on the modal's body (title, text, padding) is absorbed
	// instead of falling through to the scrim and dismissing — only a
	// genuine BACKDROP click closes.
	_ = m.cardClick.Clicked(gtx)
	// The scrim and the centering both need the FULL area this modal was
	// given, and both size themselves off Constraints.Min — which every
	// parent hands down as zero or content-sized. So no Stack here:
	// claim Max explicitly and paint the two layers sequentially into
	// the same ops — scrim first, card on top (later paint also wins
	// hit-testing, so the card's buttons sit above the scrim's click
	// area).
	gtx.Constraints.Min = gtx.Constraints.Max
	m.scrimClick.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		scrim := th.Palette.Overlay
		scrim.A = uint8(float32(scrim.A) * appear)
		return Fill(gtx, scrim)
	})
	rise := op.Offset(image.Pt(0, int(float32(gtx.Dp(12))*(1-appear)))).Push(gtx.Ops)
	defer rise.Pop()
	layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		w := m.resolveWidth(th, gtx)
		maxW := gtx.Dp(w)
		if gtx.Constraints.Max.X < maxW {
			maxW = gtx.Constraints.Max.X
		}
		gtx.Constraints.Max.X, gtx.Constraints.Min.X = maxW, maxW
		// A tall modal stops short of the window edges instead of
		// touching them; taller content must scroll internally.
		if margin := 2 * gtx.Dp(th.Space.LG); gtx.Constraints.Max.Y > margin {
			gtx.Constraints.Max.Y -= margin
		}
		return m.cardClick.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			dims := SurfaceCard(th, gtx, content)
			if onClose != nil && !m.HideClose {
				sz := gtx.Dp(28)
				st := op.Offset(image.Pt(dims.Size.X-sz-gtx.Dp(8), gtx.Dp(8))).Push(gtx.Ops)
				cgtx := gtx
				cgtx.Constraints = layout.Constraints{Max: image.Pt(sz, sz)}
				SVGIconButtonTint(th, &m.closeBtn, IconRefuse, 14, false, th.Palette.FgSubtle)(cgtx)
				st.Pop()
			}
			return dims
		})
	})
	return layout.Dimensions{Size: gtx.Constraints.Max}
}
