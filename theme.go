// Package lotusui is the ikaito design language for Gio (gioui.org)
// apps — desktop, mobile and web from one codebase: a semantic color
// system, an embedded DM Sans variable font, and prop-driven
// components — Button, Stack, Modal, TextField, Tabs, Split — built
// for immediate mode.
//
// The design decisions the package enforces: neutral grays by default,
// white cards floating on a tinted page, one accent color, and status
// colors that render as pastel background + deep ink, never saturated
// fill + white text.
//
// Customization flows through ONE object: every color a component
// paints comes from Theme.Palette's named tokens, every corner from
// Theme.Radius, every gap from Theme.Space, every motion step from
// Theme.Duration. All of it resolves at theme construction (NewTheme
// with options) — never per frame — so a fully custom look costs
// exactly as much as the default one.
package lotusui

import (
	"sync"

	"image/color"

	"gioui.org/font"
	"gioui.org/font/opentype"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	_ "embed"
)

//go:embed assets/DMSans-Variable.ttf
var uiFont []byte

// Palette is a theme's single source of truth for color: the standard
// semantic-token vocabulary (bg / fg / border ladders, brand slots,
// success / warning / info / danger status pairs) so anyone who has
// themed a modern component library already knows every name. Every
// color a lotusui component paints is one of these tokens — swap the
// Palette (NewTheme(WithPalette(p))) and every component follows;
// there are no hidden hard-coded colors.
type Palette struct {
	// Backgrounds, faint → prominent.
	Bg           color.NRGBA // the canvas: window/page background
	BgSubtle     color.NRGBA // slight elevation: chip fills, hover, track fills
	BgMuted      color.NRGBA // a step past subtle: pressed rows, quiet wells
	BgEmphasized color.NRGBA // the strongest neutral fill
	BgPanel      color.NRGBA // panels floating on the canvas: cards, sidebars, inputs
	BgInverted   color.NRGBA // dark surface for inverted moments (tooltips)

	// Borders, faint → prominent.
	BorderSubtle     color.NRGBA // 1dp dividers between content
	BorderMuted      color.NRGBA // between divider and outline
	Border           color.NRGBA // panel outlines
	BorderEmphasized color.NRGBA // outlines that must assert (focus, drag targets)

	// Foreground (text and icons), prominent → faint.
	Fg         color.NRGBA // headings, primary values
	FgMuted    color.NRGBA // body copy
	FgSubtle   color.NRGBA // captions, meta, section labels
	FgDisabled color.NRGBA // placeholders, disabled text
	FgInverted color.NRGBA // text/icons on BgInverted

	// The brand, in the standard color-palette slots.
	BrandSolid      color.NRGBA // solid brand fill — buttons, selections
	BrandSubtle     color.NRGBA // faint brand fill — subtle-variant backgrounds
	BrandFg         color.NRGBA // readable same-hue ink on light surfaces
	BrandEmphasized color.NRGBA // pressed/darker brand fill
	BrandContrast   color.NRGBA // text/icons on BrandSolid

	// Status pairs: an ink and its pastel background. Badges render as
	// tinted pill + deep ink text, never saturated fill + white text.
	Success        color.NRGBA
	SuccessBg      color.NRGBA
	Warning        color.NRGBA
	WarningBg      color.NRGBA
	Info           color.NRGBA
	InfoBg         color.NRGBA
	Danger         color.NRGBA
	DangerBg       color.NRGBA
	DangerContrast color.NRGBA // text/icons on solid Danger fills

	Overlay   color.NRGBA // the modal scrim
	FocusRing color.NRGBA // keyboard-focus ring around interactive controls
}

// RadiusScale is a theme's corner-radius scale.
type RadiusScale struct {
	SM unit.Dp
	MD unit.Dp
	LG unit.Dp
}

// SpaceScale is a theme's 8pt spacing scale.
type SpaceScale struct {
	XS unit.Dp
	SM unit.Dp
	MD unit.Dp
	LG unit.Dp
	XL unit.Dp
}

// Radius holds the default corner-radius scale. Treat it as read-only:
// per-app customization goes through NewTheme(WithRadius(...)), which
// components read as th.Radius.
var Radius = RadiusScale{SM: 6, MD: 10, LG: 12}

// Space holds the default 8pt spacing scale. Treat it as read-only:
// per-app customization goes through NewTheme(WithSpace(...)), which
// components read as th.Space.
var Space = SpaceScale{XS: 4, SM: 8, MD: 16, LG: 24, XL: 32}

func rgb(r, g, b uint8) color.NRGBA { return color.NRGBA{R: r, G: g, B: b, A: 0xFF} }

// DefaultPalette is the ikaito color scheme: the white-dominant gray
// scale shared across ikaito desktop tools, with the soft-lavender
// brand accent.
var DefaultPalette = Palette{
	// A soft cool gray page behind pure-white cards — white cards
	// floating on a tinted page with a soft shadow is most of what reads
	// as "designed" at a glance. The sidebar keeps Card white, so the
	// chrome separates from the content column without a heavy border.
	// Every grey leans a TINY bit blue — a nod toward the lavender
	// accent, so chrome and accent read as one temperature. Change the
	// cast here, never per-view.
	Bg:           rgb(0xF3, 0xF4, 0xFA),
	BgSubtle:     rgb(0xF4, 0xF5, 0xFB),
	BgMuted:      rgb(0xE9, 0xEB, 0xF4),
	BgEmphasized: rgb(0xDD, 0xE0, 0xEC),
	BgPanel:      rgb(0xFF, 0xFF, 0xFF),
	BgInverted:   rgb(0x23, 0x24, 0x2F),

	BorderSubtle:     rgb(0xEC, 0xEE, 0xF6),
	BorderMuted:      rgb(0xE9, 0xEB, 0xF4),
	Border:           rgb(0xE6, 0xE8, 0xF2),
	BorderEmphasized: rgb(0xD4, 0xD8, 0xE7),

	Fg:         rgb(0x23, 0x24, 0x2F),
	FgMuted:    rgb(0x47, 0x4A, 0x5A),
	FgSubtle:   rgb(0x6F, 0x72, 0x88),
	FgDisabled: rgb(0xAF, 0xB2, 0xC6),
	FgInverted: rgb(0xFF, 0xFF, 0xFF),

	// The brand is a soft lavender. A pastel can't be used for text on
	// white — it's near-invisible — so it splits into slots: BrandSolid
	// is the FILL (buttons, selections, with deep-indigo BrandContrast
	// text on top), BrandFg the readable dark sibling of the same hue.
	BrandSolid:      rgb(0xE0, 0xD9, 0xFF),
	BrandSubtle:     rgb(0xEC, 0xE9, 0xFC),
	BrandFg:         rgb(0x4C, 0x3F, 0xD6),
	BrandEmphasized: rgb(0xC9, 0xBD, 0xF7),
	BrandContrast:   rgb(0x2F, 0x25, 0x66),

	Success:   rgb(0x11, 0x7A, 0x3D),
	SuccessBg: rgb(0xD9, 0xF3, 0xE4),
	// The amber ink is darker than a typical warning yellow ON PURPOSE:
	// it's the deepest step that still reads as amber while holding
	// WCAG AA (≥4.5:1) on both WarningBg and white — verified by the
	// CLI's contrast checker (`lotusui theme -strict`).
	Warning:        rgb(0x8C, 0x5F, 0x00),
	WarningBg:      rgb(0xFD, 0xF3, 0xDD),
	Info:           rgb(0x2A, 0x62, 0xB5),
	InfoBg:         rgb(0xE1, 0xEC, 0xFB),
	Danger:         rgb(0xC8, 0x2E, 0x2E),
	DangerBg:       rgb(0xFD, 0xEC, 0xEC),
	DangerContrast: rgb(0xFF, 0xFF, 0xFF),

	Overlay:   color.NRGBA{A: 110},
	FocusRing: rgb(0x4C, 0x3F, 0xD6),
}

// DefaultDarkPalette is the built-in dark theme: the same lavender
// brand on a deep cool-gray canvas. Dark mode in lotusui is not a
// mode at all — it is a Palette like any other, applied with
// NewTheme(WithPalette(DefaultDarkPalette)); an app that follows the
// system appearance simply constructs both themes at startup and
// swaps the pointer. Every ladder keeps the light palette's ORDER
// (faint → prominent) — components never know which world they're in.
var DefaultDarkPalette = Palette{
	// Panels sit slightly LIGHTER than the canvas — elevation still
	// reads as "closer to the light", just as white cards do on the
	// tinted light page. The same tiny blue lean keeps chrome and
	// accent at one temperature.
	Bg:           rgb(0x16, 0x17, 0x1E),
	BgSubtle:     rgb(0x1C, 0x1E, 0x27),
	BgMuted:      rgb(0x23, 0x26, 0x31),
	BgEmphasized: rgb(0x2C, 0x30, 0x40),
	BgPanel:      rgb(0x1E, 0x20, 0x29),
	BgInverted:   rgb(0xF3, 0xF4, 0xFA),

	BorderSubtle:     rgb(0x26, 0x29, 0x35),
	BorderMuted:      rgb(0x2C, 0x30, 0x40),
	Border:           rgb(0x33, 0x37, 0x48),
	BorderEmphasized: rgb(0x43, 0x49, 0x60),

	Fg:         rgb(0xEC, 0xED, 0xF5),
	FgMuted:    rgb(0xBF, 0xC2, 0xD4),
	FgSubtle:   rgb(0x8B, 0x8F, 0xA8),
	FgDisabled: rgb(0x56, 0x5A, 0x70),
	FgInverted: rgb(0x23, 0x24, 0x2F),

	// The lavender fill carries over VERBATIM — a pastel pops on dark
	// even better than on white — while the readable ink flips to a
	// light sibling of the hue.
	BrandSolid:      rgb(0xE0, 0xD9, 0xFF),
	BrandSubtle:     rgb(0x2A, 0x27, 0x45),
	BrandFg:         rgb(0xB7, 0xA9, 0xFF),
	BrandEmphasized: rgb(0xC9, 0xBD, 0xF7),
	BrandContrast:   rgb(0x2F, 0x25, 0x66),

	// Status pairs invert their construction: light ink on a deep
	// tinted well, instead of deep ink on a pastel.
	Success:        rgb(0x6E, 0xD9, 0x9B),
	SuccessBg:      rgb(0x12, 0x29, 0x1C),
	Warning:        rgb(0xE5, 0xB4, 0x60),
	WarningBg:      rgb(0x2E, 0x24, 0x13),
	Info:           rgb(0x7F, 0xB1, 0xF5),
	InfoBg:         rgb(0x14, 0x22, 0x3A),
	Danger:         rgb(0xF0, 0x6E, 0x6E),
	DangerBg:       rgb(0x37, 0x1A, 0x1A),
	DangerContrast: rgb(0x2B, 0x0A, 0x0A),

	Overlay:   color.NRGBA{A: 150},
	FocusRing: rgb(0xB7, 0xA9, 0xFF),
}

// Scheme is a semantic color scheme: how one color role renders
// across the Button variants (and any future colored surface),
// INCLUDING its interaction steps — base, hover, active — the
// .500/.600/.700 laddering of a color scale. The theme exposes three:
// Accent (the brand), Neutral (the greys), Danger (destructive only).
// Schemes are DERIVED — from Palette tokens or a ColorScale's steps —
// so a custom palette or scale propagates to every variant and every
// interaction state with no extra configuration. Zero Hover/Active
// fields fall back to an arithmetic shade of the base, so hand-built
// schemes keep working.
type Scheme struct {
	Solid       color.NRGBA // solid variant background
	SolidHover  color.NRGBA // …while hovered (a scale step darker)
	SolidActive color.NRGBA // …while pressed (one more step)
	OnSolid     color.NRGBA // text/icons on Solid

	Subtle       color.NRGBA // subtle variant background
	SubtleHover  color.NRGBA // …while hovered
	SubtleActive color.NRGBA // …while pressed
	OnSubtle     color.NRGBA // text on Subtle / Outline / Ghost

	Outline color.NRGBA // outline variant border
}

// Accent is the brand scheme — the app's one accent. Its interaction
// steps walk the palette's own brand ladder: Subtle → Solid →
// Emphasized, so hover and press feel like the same hue deepening.
func (p Palette) Accent() Scheme {
	return Scheme{
		Solid: p.BrandSolid, SolidHover: p.BrandEmphasized, SolidActive: shade(p.BrandEmphasized),
		OnSolid: p.BrandContrast,
		Subtle:  p.BrandSubtle, SubtleHover: p.BrandSolid, SubtleActive: p.BrandEmphasized,
		OnSubtle: p.BrandFg,
		Outline:  p.BrandFg,
	}
}

// Neutral is the grey scheme; its steps walk the background ladder
// (BgSubtle → BgMuted → BgEmphasized).
func (p Palette) Neutral() Scheme {
	return Scheme{
		Solid: p.BgSubtle, SolidHover: p.BgMuted, SolidActive: p.BgEmphasized,
		OnSolid: p.Fg,
		Subtle:  p.BgSubtle, SubtleHover: p.BgMuted, SubtleActive: p.BgEmphasized,
		OnSubtle: p.Fg,
		Outline:  p.Border,
	}
}

// DangerScheme is the destructive scheme — Delete-class actions
// exclusively.
func (p Palette) DangerScheme() Scheme {
	return Scheme{
		Solid: p.Danger, SolidHover: shade(p.Danger), SolidActive: shade(shade(p.Danger)),
		OnSolid: p.DangerContrast,
		Subtle:  p.DangerBg, SubtleHover: shade(p.DangerBg), SubtleActive: shade(shade(p.DangerBg)),
		OnSubtle: p.Danger,
		Outline:  p.Danger,
	}
}

// Theme bundles the material theme (used for widget defaults like
// buttons and text fields) with the design tokens lotusui components
// read: Palette for color, Radius and Space for shape and rhythm.
// Build one with NewTheme at startup and pass it everywhere; all
// customization is resolved here, never per frame.
type Theme struct {
	Material *material.Theme
	Palette  Palette
	Radius   RadiusScale
	Space    SpaceScale
	Duration DurationScale

	// Breakpoints are named min-widths (dp) for responsive layout
	// props — Columns, gaps, Show, PageMax. DefaultBreakpoints unless
	// WithBreakpoints / ParseBreakpointsJSON.
	Breakpoints Breakpoints

	// PageMax caps LayoutPage's readable column (default 920dp).
	// PageMaxAt, when set, overrides PageMax per breakpoint.
	PageMax   unit.Dp
	PageMaxAt ResponsiveDp

	// BrandScale is the palette's brand hue graded 50…900 —
	// automatically derived from Palette.BrandFg at construction, so
	// defining ONE custom brand color gives the whole app a full scale
	// (th.BrandScale.C50 tints, th.BrandScale.Scheme() buttons, …)
	// with no extra configuration.
	BrandScale ColorScale

	// baseFaces are the faces the shapers are built from (the embedded
	// DM Sans, or WithFaces' replacement) — parsed once at NewTheme.
	baseFaces []font.FontFace

	// hovers keys per-Clickable hover clocks (Duration.Fast). UI
	// thread only — Gio frames are single-threaded.
	hovers map[*widget.Clickable]*slideAnim

	pendingMu     sync.Mutex
	pendingShaper *text.Shaper
}

// ThemeOption customizes NewTheme. Options are applied once, at
// construction — theming has zero per-frame cost.
type ThemeOption func(*Theme)

// WithPalette replaces the color tokens. Components derive every fill,
// ink, border and scheme from these, so one palette swap restyles the
// whole library consistently.
func WithPalette(p Palette) ThemeOption {
	return func(t *Theme) { t.Palette = p }
}

// WithRadius replaces the corner-radius scale.
func WithRadius(r RadiusScale) ThemeOption {
	return func(t *Theme) { t.Radius = r }
}

// WithSpace replaces the spacing scale.
func WithSpace(s SpaceScale) ThemeOption {
	return func(t *Theme) { t.Space = s }
}

// WithPageMax sets LayoutPage's readable column cap (default 920dp).
func WithPageMax(max unit.Dp) ThemeOption {
	return func(t *Theme) { t.PageMax = max }
}

// WithTextSize sets the base text size the whole type scale derives
// from (default 16sp; see Sp).
func WithTextSize(sz unit.Sp) ThemeOption {
	return func(t *Theme) { t.Material.TextSize = sz }
}

// WithFaces replaces the embedded DM Sans as the primary font
// collection — the brand font for an app that isn't ikaito's. System
// fallback still arrives via UpgradeShaperAsync.
func WithFaces(faces []font.FontFace) ThemeOption {
	return func(t *Theme) { t.baseFaces = faces }
}

// Sp scales th.TextSize by ratio, producing a derived type-scale value —
// e.g. Sp(th, 12.0/16.0) for a 12sp label when the base size is 16sp.
// This keeps every label's size relative to one base instead of
// hard-coded magic numbers scattered through views.
func Sp(th *Theme, ratio float32) unit.Sp {
	return unit.Sp(float32(th.Material.TextSize) * ratio)
}

// ── Type scale ────────────────────────────────────────────────────────────
//
// Hierarchy comes from weight (500/600/700) more than size: 13sp gray
// meta text, 14-15sp body, 16-17sp semibold titles, with genuinely big
// text reserved for the one hero question on a screen. Use these instead
// of material.H*/Body*/Caption everywhere.

// LabelHero is the one big text on a screen — a screen's H1. 20sp.
func LabelHero(th *Theme, txt string) material.LabelStyle {
	l := material.Label(th.Material, Sp(th, 20.0/16.0), txt)
	l.Font.Weight = 600
	l.Color = th.Palette.Fg
	return l
}

// LabelTitle is a section/modal/card-group title — 16sp semibold.
func LabelTitle(th *Theme, txt string) material.LabelStyle {
	l := material.Label(th.Material, Sp(th, 16.0/16.0), txt)
	l.Font.Weight = 600
	l.Color = th.Palette.Fg
	return l
}

// LabelCardTitle is a card's own heading — 14sp semibold.
func LabelCardTitle(th *Theme, txt string) material.LabelStyle {
	l := material.Label(th.Material, Sp(th, 14.0/16.0), txt)
	l.Font.Weight = 600
	l.Color = th.Palette.Fg
	return l
}

// LabelBody is primary row/content text — 14sp regular.
func LabelBody(th *Theme, txt string) material.LabelStyle {
	l := material.Label(th.Material, Sp(th, 14.0/16.0), txt)
	l.Color = th.Palette.FgMuted
	return l
}

// LabelMeta is secondary/explanatory text — 13sp.
func LabelMeta(th *Theme, txt string) material.LabelStyle {
	l := material.Label(th.Material, Sp(th, 13.0/16.0), txt)
	l.Color = th.Palette.FgSubtle
	return l
}

// LabelCaption is the smallest text — timestamps, fine print, 12sp.
func LabelCaption(th *Theme, txt string) material.LabelStyle {
	l := material.Label(th.Material, Sp(th, 12.0/16.0), txt)
	l.Color = th.Palette.FgSubtle
	return l
}

// NewTheme builds a Theme: the ikaito defaults, then any options.
//
//	th := lotusui.NewTheme()                              // ikaito look
//	th := lotusui.NewTheme(lotusui.WithPalette(myBrand))  // custom look
func NewTheme(opts ...ThemeOption) *Theme {
	m := material.NewTheme()
	t := &Theme{
		Material:    m,
		Palette:     DefaultPalette,
		Radius:      Radius,
		Space:       Space,
		Duration:    Duration,
		Breakpoints: DefaultBreakpoints,
		PageMax:     920,
	}
	m.TextSize = unit.Sp(16)
	m.FingerSize = unit.Dp(44)
	for _, o := range opts {
		o(t)
	}
	// Derived values come AFTER options, so they always reflect the
	// final palette — a custom brand automatically grades its scale.
	t.BrandScale = ScaleFrom(t.Palette.BrandFg)
	if t.baseFaces == nil {
		faces, err := opentype.ParseCollection(uiFont)
		if err != nil || len(faces) == 0 {
			panic("lotusui: failed to load DM Sans font: " + errString(err))
		}
		t.baseFaces = faces
	}
	// Two-phase shaping: NewShaper with system fallback enumerates every
	// font on the machine — hundreds of ms before the first frame could
	// paint. Phase 1 (here) shapes with the base collection alone so the
	// window appears instantly; phase 2 (UpgradeShaperAsync) builds the
	// full shaper in the background and swaps it in a few hundred ms
	// later — symbols outside the base coverage (✓ ⚠ ●) simply sharpen
	// on the next invalidate.
	m.Shaper = text.NewShaper(text.NoSystemFonts(), text.WithCollection(t.baseFaces))
	m.Palette = material.Palette{
		Bg:         t.Palette.Bg,
		Fg:         t.Palette.Fg,
		ContrastBg: t.Palette.BrandSolid,
		ContrastFg: t.Palette.BrandContrast,
	}
	return t
}

// UpgradeShaperAsync builds the full text shaper (system-font fallback
// included) off the startup path and hands it to ApplyPendingShaper.
// invalidate is Window.Invalidate — safe from any goroutine.
func (t *Theme) UpgradeShaperAsync(invalidate func()) {
	base := t.baseFaces
	go func() {
		full := text.NewShaper(text.WithCollection(base))
		t.pendingMu.Lock()
		t.pendingShaper = full
		t.pendingMu.Unlock()
		invalidate()
	}()
}

// SetExtraFaces rebuilds the shaper with additional runtime-loaded font
// faces (e.g. a CJK face fetched from a server) appended after the
// theme's base collection, system-font fallback included. Safe from any
// goroutine; the swap lands via ApplyPendingShaper on the next frame.
// Apps that call this should NOT also call UpgradeShaperAsync — the two
// would race to set the pending shaper and the loser's faces would be
// dropped.
func (t *Theme) SetExtraFaces(faces []font.FontFace, invalidate func()) {
	base := t.baseFaces
	go func() {
		full := text.NewShaper(text.WithCollection(append(base[:len(base):len(base)], faces...)))
		t.pendingMu.Lock()
		t.pendingShaper = full
		t.pendingMu.Unlock()
		if invalidate != nil {
			invalidate()
		}
	}()
}

// ApplyPendingShaper swaps the upgraded shaper in — called at the top of
// every frame (UI goroutine), so the swap can never race a render.
func (t *Theme) ApplyPendingShaper() {
	t.pendingMu.Lock()
	if t.pendingShaper != nil {
		t.Material.Shaper = t.pendingShaper
		t.pendingShaper = nil
	}
	t.pendingMu.Unlock()
}

func errString(err error) string {
	if err == nil {
		return "no font faces found"
	}
	return err.Error()
}
