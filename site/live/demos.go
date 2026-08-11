// Package live holds the addressable component demos shared by the
// gallery harness and the lotusui docs app.
package live

import (
	"fmt"
	"image"
	"image/color"
	"strconv"
	"strings"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"

	lotusui "github.com/ikaito-com/lotusui"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

type Demo struct {
	slug       string
	render     func(th *lotusui.Theme, gtx C) D
	overlay    func(th *lotusui.Theme, gtx C) D
	applyState func(state string)
	// edges reports an internally-scrollable demo's scroll position
	// (atStart, atEnd) — the overlay page uses it to chain wheel
	// events to the page exactly when the inner scroll is exhausted.
	edges func() (atStart, atEnd bool)
}

var Demos = []Demo{
	{slug: "palette", render: paletteDemo},
	{slug: "scales", render: scalesDemo},
	{slug: "typography", render: typographyDemo},
	{slug: "stack", render: stackDemo},
	{slug: "wrap", render: wrapDemo},
	{slug: "layout", render: layoutChromeDemo},
	{slug: "button", render: buttonDemo},
	{slug: "input", render: inputDemo},
	{slug: "item", render: itemDemo},
	{slug: "checkbox", render: checkboxDemo},
	{slug: "code-block", render: codeBlockDemo},
	{slug: "example", render: exampleDemo},
	{slug: "field", render: fieldDemo},
	{slug: "switch", render: switchDemo, applyState: switchState},
	{slug: "select", render: selectDemo},
	{slug: "scroll-area", render: scrollAreaDemo},
	{slug: "tabs", render: tabsDemo, applyState: tabsState},
	{slug: "dialog", render: dialogDemo, overlay: dialogOverlay, applyState: dialogState},
	{slug: "list", render: listDemo, edges: listEdges},
	{slug: "badge", render: badgeDemo},
	{slug: "dropdown-menu", render: menuDemo},
	{slug: "context-menu", render: contextMenuDemo},
	{slug: "desktop-download", render: desktopDownloadDemo},
	{slug: "accordion", render: accordionDemo},
	{slug: "alert", render: alertDemo},
	{slug: "alert-dialog", render: alertDialogDemo, overlay: alertDialogOverlay},
	{slug: "annotated-text", render: annotatedTextDemo},
	{slug: "avatar", render: avatarDemo},
	{slug: "breadcrumb", render: breadcrumbDemo},
	{slug: "button-group", render: buttonGroupDemo},
	{slug: "input-otp", render: inputOTPDemo},
	{slug: "pagination", render: paginationDemo},
	{slug: "popover", render: popoverDemo},
	{slug: "hover-card", render: hoverCardDemo},
	{slug: "progress", render: progressDemo},
	{slug: "radio-group", render: radioDemo},
	{slug: "separator", render: separatorDemo},
	{slug: "skeleton", render: skeletonDemo},
	{slug: "slider", render: sliderDemo},
	{slug: "spinner", render: spinnerDemo},
	{slug: "table", render: tableDemo},
	{slug: "textarea", render: textareaDemo},
	{slug: "toast", render: toastDemo, overlay: toastOverlay},
	{slug: "toggle", render: toggleDemo},
	{slug: "tooltip", render: tooltipDemo},
	{slug: "card", render: cardDemo},
	{slug: "grid", render: gridDemo},
	{slug: "simplegrid", render: simpleGridDemo},
	{slug: "split", render: splitDemo, applyState: splitState},
	{slug: "icons", render: iconsDemo},
	{slug: "showcase", render: showcaseDemo},
	{slug: "showcase-colors", render: showcaseColorsDemo},
	{slug: "showcase-devices", render: showcaseDevicesDemo},
	// blank paints NOTHING — the transparency probe for overlay-style
	// embedding (and a harmless addressable no-op state).
	{slug: "blank", render: func(th *lotusui.Theme, gtx C) D { return D{Size: image.Pt(10, 10)} }},
}

var AppliedRoute = "\x00" // never equal to a real route, so the first frame applies state

// Overlay mode: the docs page runs ONE gallery instance in a
// transparent iframe spanning the whole article and tells it where
// every demo box sits. Each frame renders every region — one Go
// runtime, all demos live.
type overlayRegion struct {
	slug, state string
	rect        image.Rectangle // device px, relative to the canvas
}

// VisibleRegions, when non-empty, is the set of region indexes worth
// laying out this frame — the docs page keeps it current on scroll,
// so offscreen demos cost nothing whatever the page length.
var VisibleRegions map[int]bool

var (
	OverlayRegions []overlayRegion
	// pending measure requests: specs like "button/2" with box widths
	// (device px); answered inside the frame loop where a real gtx
	// exists, then sent back to the page.
	PendingMeasureSpecs  []string
	PendingMeasureWidths []int
)

// DemoState is the parsed sub-state. A NUMERIC state ("#button/2")
// shows only that section of the demo — how the docs give every
// feature its own focused live example from the one gallery binary.
var DemoState string

// OverlayOwner is the region key ("slug/N") whose interaction opened
// the current overlay (dialog, toast). In the multi-region strip
// every region shares the demo's state, so without an owner a dialog
// opened in one box would paint its scrim in every box of the page.
var OverlayOwner string

// CurrentSlug is the demo being laid out right now — renderOverlay
// sets it per region, render sets it once. OwnOverlay stamps the
// fully qualified owner key from inside a section closure.
var CurrentSlug string

func OwnOverlay() { OverlayOwner = CurrentSlug + "/" + DemoState }

// parseRoute splits "#slug/state&palette=name&look=name" — the params
// make any demo state deep-linkable in any theme (either axis, any
// order).
func ParseRoute(r string) (slug, state, palette, look string) {
	r = strings.TrimPrefix(r, "#")
	parts := strings.Split(r, "&")
	for _, kv := range parts[1:] {
		if v, ok := strings.CutPrefix(kv, "palette="); ok {
			palette = v
		}
		if v, ok := strings.CutPrefix(kv, "look="); ok {
			look = v
		}
	}
	slug, state, _ = strings.Cut(parts[0], "/")
	return slug, state, palette, look
}

func Lookup(slug string) Demo {
	for _, d := range Demos {
		if d.slug == slug {
			return d
		}
	}
	return Demos[0]
}

// Has reports whether a demo slug is registered (Lookup alone falls
// back to the first demo for unknown routes).
func Has(slug string) bool {
	for _, d := range Demos {
		if d.slug == slug {
			return true
		}
	}
	return false
}

// RenderSection lays out one focused demo ("button/2") for docsapp
// section Previews. Unlike Render, it does not read CurrentRoute and
// does not fill the panel — callers supply chrome. The demo is the
// same in-process widget the gallery uses; there is nothing to load.
func RenderSection(th *lotusui.Theme, gtx C, demo string) D {
	if demo == "" {
		return D{}
	}
	slug, state, _, _ := ParseRoute(demo)
	d := Lookup(slug)
	savedState, savedSlug := DemoState, CurrentSlug
	DemoState = state
	CurrentSlug = slug
	defer func() {
		DemoState = savedState
		CurrentSlug = savedSlug
	}()

	// Apply addressable state once per focused section (slug+state),
	// not once per slug — otherwise later sections of the same demo
	// (tabs/1 after tabs/0) re-apply and reset interactive widgets
	// every frame (Account/Password flicker).
	key := slug + "/" + state
	if _, ok := embedApplied[key]; !ok {
		embedApplied[key] = state
		if d.applyState != nil {
			d.applyState(state)
		}
	}

	cgtx := gtx
	cgtx.Constraints.Min.Y = 0
	cgtx.Constraints.Max.Y = 1 << 20
	dims := d.render(th, cgtx)
	if d.overlay != nil && Embed {
		ov := d.overlay
		PendingOverlay = func(th *lotusui.Theme, gtx C) { ov(th, gtx) }
	} else if d.overlay != nil {
		d.overlay(th, gtx)
	}
	return dims
}

// embedApplied tracks which focused states have been applied so
// interaction is not reset every frame.
var embedApplied = map[string]string{}

// renderOverlay paints every region at its rect: panel fill, the
// demo's section, and its overlay (clipped to the region — a modal
// scrim covers its box, not the article).
// setOverlayRegions installs a region layout. Addressable states are
// applied only for specs that are NEW since the previous layout — a
// pure re-position (resize, font settle) must never re-apply states,
// or user interaction would reset and entrance animations replay.
func SetOverlayRegions(rs []overlayRegion) {
	prev := map[string]bool{}
	for _, r := range OverlayRegions {
		prev[r.slug+"/"+r.state] = true
	}
	OverlayRegions = rs
	for _, r := range rs {
		if prev[r.slug+"/"+r.state] {
			continue
		}
		if d := Lookup(r.slug); d.applyState != nil {
			d.applyState(r.state)
		}
	}
}

// reportRegionEdges pushes each scrollable region's edge state to the
// iframe-side wheel router (JS), only when something changed.
var LastEdges []int8

func ReportRegionEdges() {
	if len(OverlayRegions) == 0 {
		return
	}
	if len(LastEdges) != len(OverlayRegions) {
		LastEdges = make([]int8, len(OverlayRegions))
		for i := range LastEdges {
			LastEdges[i] = -1
		}
	}
	for i, r := range OverlayRegions {
		d := Lookup(r.slug)
		var enc int8
		if d.edges != nil {
			atStart, atEnd := d.edges()
			enc = 1 // scrollable
			if atStart {
				enc |= 2
			}
			if atEnd {
				enc |= 4
			}
		}
		if enc != LastEdges[i] {
			LastEdges[i] = enc
			SetRegionScroll(i, enc)
		}
	}
}

func RenderOverlay(th *lotusui.Theme, gtx C) D {
	for i, r := range OverlayRegions {
		// Cull: only the slots the page reports visible are laid out —
		// the page's length never raises the frame cost.
		if len(VisibleRegions) > 0 && !VisibleRegions[i] {
			continue
		}
		d := Lookup(r.slug)
		DemoState = r.state
		CurrentSlug = r.slug
		st := op.Offset(r.rect.Min).Push(gtx.Ops)
		cl := clip.Rect(image.Rectangle{Max: r.rect.Size()}).Push(gtx.Ops)
		cgtx := gtx
		cgtx.Constraints = layout.Constraints{Min: r.rect.Size(), Max: r.rect.Size()}
		lotusui.Fill(cgtx, th.Palette.BgPanel)
		// Top-align (layout.N): Floating menus/Selects need room BELOW
		// the trigger inside the demobox. Centering wasted half the
		// DemoH above the control and clipped open panels.
		layout.N.Layout(cgtx, func(gtx C) D {
			return layout.UniformInset(lotusui.Space.LG).Layout(gtx, func(gtx C) D {
				return d.render(th, gtx)
			})
		})
		// The owner gate is SLUG-scoped: a dialog opened from one box
		// must not paint its scrim in the page's other boxes, but an
		// owner left over from another page never gates anything.
		gated := OverlayOwner != "" && strings.HasPrefix(OverlayOwner, d.slug+"/") && OverlayOwner != d.slug+"/"+r.state
		if d.overlay != nil && !gated {
			ogtx := gtx
			ogtx.Constraints = layout.Constraints{Min: r.rect.Size(), Max: r.rect.Size()}
			d.overlay(th, ogtx)
		}
		cl.Pop()
		st.Pop()
	}
	return D{Size: gtx.Constraints.Max}
}

func Render(th *lotusui.Theme, gtx C) D {
	if len(OverlayRegions) > 0 {
		return RenderOverlay(th, gtx)
	}
	route := CurrentRoute()
	slug, state, palette, look := ParseRoute(route)
	d := Lookup(slug)
	CurrentSlug = slug
	if route != AppliedRoute {
		AppliedRoute = route
		if palette != "" {
			CurPalette = palette
		}
		if look != "" {
			CurLook = look
		}
		if palette != "" || look != "" {
			OnThemeChange()
		}
		DemoState = state
		if d.applyState != nil {
			d.applyState(state)
		}
	}
	// The demo IS the example box: white background, content at its
	// natural height. The docs page sizes the iframe to match, so there
	// is one box and one scrollbar — the page's. Embed hosts (docsapp)
	// paint their own page chrome and per-section preview cards.
	if d.slug != "blank" && !Embed {
		lotusui.Fill(gtx, th.Palette.BgPanel)
	}
	// Lay the demo out at UNBOUNDED height: components clamp to their
	// constraints, so a viewport-bounded pass could never report a
	// height larger than the iframe already is — content would squeeze
	// instead of asking for room. Unbounded, dims is the demo's natural
	// height; the page grows the iframe to it and the next frame's
	// canvas matches.
	cgtx := gtx
	cgtx.Constraints.Min.Y = 0
	cgtx.Constraints.Max.Y = 1 << 20
	inset := lotusui.Space.LG
	if Embed {
		inset = 0 // preview cards carry their own padding
	}
	dims := layout.UniformInset(inset).Layout(cgtx, func(gtx C) D {
		return d.render(th, gtx)
	})
	if d.overlay != nil {
		if Embed {
			// Host paints at the window root so scrims cover chrome too.
			ov := d.overlay
			PendingOverlay = func(th *lotusui.Theme, gtx C) { ov(th, gtx) }
		} else {
			d.overlay(th, gtx) // overlays stay at WINDOW constraints
		}
	}
	return dims
}

// PendingOverlay is set by Render when Embed is true; the host paints
// it once at the window root (after chrome) via PaintOverlay.
var PendingOverlay func(th *lotusui.Theme, gtx C)

// PaintOverlay paints and clears PendingOverlay. No-op when unset.
func PaintOverlay(th *lotusui.Theme, gtx C) {
	if PendingOverlay == nil {
		return
	}
	fn := PendingOverlay
	PendingOverlay = nil
	fn(th, gtx)
}

// ResetEmbed clears embed-host frame state (preview heights, pending
// overlay, applied section states) — call when navigating between docs pages.
func ResetEmbed() {
	PendingOverlay = nil
	embedPrevH = map[string]int{}
	embedApplied = map[string]string{}
}

// section pairs a quiet SectionLabel with one block of states, both
// centered on the cross axis — the example box centers the block, and
// the label centers over the content instead of hanging off its left.
// When Embed, each block sits in a white rounded preview card (the
// frame the static docs used to get from .exbox / the iframe).
func section(th *lotusui.Theme, title string, w layout.Widget) layout.Widget {
	return func(gtx C) D {
		// Embed hosts (docsapp) supply exbox chrome — never nest
		// embedPreview cards or section titles inside the Preview.
		if Embed {
			return w(gtx)
		}
		return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(lotusui.SectionLabel(th, title)),
			layout.Rigid(lotusui.Spacer(lotusui.Space.SM)),
			layout.Rigid(w),
		)
	}
}

// embedPrevH remembers each preview card's height so we can paint the
// white panel BEHIND content without op.Record (which would trap
// Floating / Select inside a macro and make the panel push layout).
var embedPrevH = map[string]int{}

func embedPreview(th *lotusui.Theme, gtx C, key string, content layout.Widget) D {
	pad := gtx.Dp(lotusui.Space.LG)
	r := gtx.Dp(th.Radius.LG)
	maxX := gtx.Constraints.Max.X

	totalH := embedPrevH[key]
	if totalH < 2*pad+1 {
		totalH = 2*pad + gtx.Dp(unit.Dp(48))
	}
	sz := gtx.Constraints.Constrain(image.Pt(maxX, totalH))

	// Chrome first (last frame's size), then content directly on the
	// root ops — Select/Menu Floating must not sit inside op.Record.
	cl := clip.UniformRRect(image.Rectangle{Max: sz}, r).Push(gtx.Ops)
	paint.Fill(gtx.Ops, th.Palette.BgPanel)
	cl.Pop()
	widget.Border{
		Color:        th.Palette.Border,
		Width:        unit.Dp(1),
		CornerRadius: th.Radius.LG,
	}.Layout(gtx, func(gtx C) D { return D{Size: sz} })

	trans := op.Offset(image.Pt(pad, pad)).Push(gtx.Ops)
	cgtx := gtx
	cgtx.Constraints.Min = image.Point{}
	cgtx.Constraints.Max.X = maxX - 2*pad
	if cgtx.Constraints.Max.X < 0 {
		cgtx.Constraints.Max.X = 0
	}
	dims := content(cgtx)
	trans.Pop()

	totalH = dims.Size.Y + 2*pad
	embedPrevH[key] = totalH
	return D{Size: image.Pt(maxX, totalH)}
}

// card stacks a demo's sections directly on the white example
// surface — the docs page's box provides the frame, so the demo adds
// no chrome of its own. A numeric DemoState narrows to that single
// section, so each docs feature section embeds exactly its example.
func card(th *lotusui.Theme, gtx C, sections ...layout.Widget) D {
	if n, err := strconv.Atoi(DemoState); err == nil && n >= 0 && n < len(sections) {
		sections = sections[n : n+1]
	}
	gap := lotusui.Space.LG
	if Embed {
		gap = lotusui.Space.XL
	}
	return lotusui.VStack(gap, sections...)(gtx)
}

// ---- palette ----

func paletteDemo(th *lotusui.Theme, gtx C) D {
	p := th.Palette
	type sw struct {
		name string
		col  color.NRGBA
	}
	groups := []struct {
		title string
		sws   []sw
	}{
		{"Backgrounds", []sw{{"Bg", p.Bg}, {"BgSubtle", p.BgSubtle}, {"BgMuted", p.BgMuted}, {"BgEmphasized", p.BgEmphasized}, {"BgPanel", p.BgPanel}, {"BgInverted", p.BgInverted}}},
		{"Borders", []sw{{"BorderSubtle", p.BorderSubtle}, {"BorderMuted", p.BorderMuted}, {"Border", p.Border}, {"BorderEmphasized", p.BorderEmphasized}}},
		{"Foreground", []sw{{"Fg", p.Fg}, {"FgMuted", p.FgMuted}, {"FgSubtle", p.FgSubtle}, {"FgDisabled", p.FgDisabled}, {"FgInverted", p.FgInverted}}},
		{"Brand", []sw{{"BrandSolid", p.BrandSolid}, {"BrandSubtle", p.BrandSubtle}, {"BrandFg", p.BrandFg}, {"BrandEmphasized", p.BrandEmphasized}, {"BrandContrast", p.BrandContrast}}},
		{"Status", []sw{{"Success", p.Success}, {"SuccessBg", p.SuccessBg}, {"Warning", p.Warning}, {"WarningBg", p.WarningBg}, {"Info", p.Info}, {"InfoBg", p.InfoBg}, {"Danger", p.Danger}, {"DangerBg", p.DangerBg}, {"DangerContrast", p.DangerContrast}}},
	}
	var sections []layout.Widget
	for _, g := range groups {
		g := g
		sections = append(sections, section(th, g.title, func(gtx C) D {
			return lotusui.SimpleGrid(th, gtx, g.sws, lotusui.SimpleGridProps{
				MinChildWidth: 110, MaxCols: 6, Gap: lotusui.Space.SM,
			}, func(gtx C, s sw) D {
				return lotusui.VStack(lotusui.Space.XS,
					func(gtx C) D {
						gtx.Constraints.Min = image.Pt(gtx.Constraints.Max.X, gtx.Dp(40))
						return lotusui.Fill(gtx, s.col)
					},
					lotusui.LabelCaption(th, s.name).Layout,
				)(gtx)
			})
		}))
	}
	return card(th, gtx, sections...)
}

// ---- color scales ----

func scalesDemo(th *lotusui.Theme, gtx C) D {
	scales := []struct {
		name string
		s    lotusui.ColorScale
	}{
		{"Gray", lotusui.Gray}, {"Red", lotusui.Red}, {"Orange", lotusui.Orange},
		{"Yellow", lotusui.Yellow}, {"Green", lotusui.Green}, {"Teal", lotusui.Teal},
		{"Blue", lotusui.Blue}, {"Cyan", lotusui.Cyan}, {"Purple", lotusui.Purple},
		{"Pink", lotusui.Pink},
		{"ScaleFrom(Brand)", lotusui.ScaleFrom(th.Palette.BrandSolid)},
	}
	var rows []layout.Widget
	for _, sc := range scales {
		sc := sc
		steps := []color.NRGBA{sc.s.C50, sc.s.C100, sc.s.C200, sc.s.C300, sc.s.C400, sc.s.C500, sc.s.C600, sc.s.C700, sc.s.C800, sc.s.C900}
		var cells []layout.Widget
		for _, c := range steps {
			c := c
			cells = append(cells, func(gtx C) D {
				sz := image.Pt(gtx.Dp(34), gtx.Dp(26))
				gtx.Constraints.Min, gtx.Constraints.Max = sz, sz
				return lotusui.Fill(gtx, c)
			})
		}
		rows = append(rows, lotusui.VStack(lotusui.Space.XS,
			lotusui.LabelCaption(th, sc.name).Layout,
			lotusui.HStack(2, cells...),
		))
	}
	return card(th, gtx,
		section(th, "50 → 900, graded from one anchor each", lotusui.VStack(lotusui.Space.MD, rows...)),
	)
}

// ---- typography ----

func typographyDemo(th *lotusui.Theme, gtx C) D {
	return card(th, gtx,
		lotusui.LabelHero(th, "LabelHero — the one big text on a screen (20sp)").Layout,
		lotusui.LabelTitle(th, "LabelTitle — section and modal titles (16sp semibold)").Layout,
		lotusui.LabelCardTitle(th, "LabelCardTitle — a card's own heading (14sp semibold)").Layout,
		lotusui.LabelBody(th, "LabelBody — primary row and content text (14sp regular)").Layout,
		lotusui.LabelMeta(th, "LabelMeta — secondary, explanatory text (13sp)").Layout,
		lotusui.LabelCaption(th, "LabelCaption — timestamps and fine print (12sp)").Layout,
		lotusui.SectionLabel(th, "SectionLabel — a quiet caption introducing a group"),
	)
}

// ---- stack / layout ----

func stackDemo(th *lotusui.Theme, gtx C) D {
	chip := func(txt string) layout.Widget {
		return lotusui.Badge(th, txt, lotusui.BadgeProps{Bg: th.Palette.BgSubtle, Fg: th.Palette.FgMuted})
	}
	return card(th, gtx,
		section(th, "VStack — vertical, uniform gap", lotusui.VStack(lotusui.Space.SM,
			chip("first"), chip("second"), chip("third"))),
		section(th, "HStack — horizontal, uniform gap", lotusui.HStack(lotusui.Space.SM,
			chip("one"), chip("two"), chip("three"), chip("four"))),
	)
}

func wrapDemo(th *lotusui.Theme, gtx C) D {
	chip := func(txt string) layout.Widget {
		return lotusui.Badge(th, txt, lotusui.BadgeProps{Bg: th.Palette.BgSubtle, Fg: th.Palette.FgMuted})
	}
	// Full box width — no artificial Max.X. Narrow the browser: chips
	// reflow to more lines; widen: they climb back onto fewer rows.
	return card(th, gtx,
		section(th, "Wrap — resize the window; chips reflow", lotusui.Wrap(lotusui.Space.SM, layout.Middle,
			chip("Design"), chip("Engineering"), chip("Product"),
			chip("Marketing"), chip("Sales"), chip("Support"),
			chip("Finance"), chip("Legal"), chip("Operations"),
			chip("Research"), chip("Security"), chip("Success"),
		)),
	)
}

// ---- layout chrome ----

var layoutChrome struct {
	add1, add2, more widget.Clickable
}

func layoutChromeDemo(th *lotusui.Theme, gtx C) D {
	return card(th, gtx,
		section(th, "TitleWithIcons — one trailing action", lotusui.TitleWithIcons(th, "Documents",
			lotusui.SVGIconButton(th, &layoutChrome.add1, lotusui.IconAdd, 20, false),
		)),
		section(th, "TitleWithIcons — two trailing actions", lotusui.TitleWithIcons(th, "Documents",
			lotusui.SVGIconButton(th, &layoutChrome.add2, lotusui.IconAdd, 20, false),
			lotusui.SVGIconButton(th, &layoutChrome.more, lotusui.IconSettings, 20, false),
		)),
	)
}

// ---- button ----

var btn struct {
	def, secondary, destructive, outline, ghost, link widget.Clickable
	iconOnly, rounded, roundedOut                     widget.Clickable
	groupCancel, groupSave                            widget.Clickable
	xs2, xs, sm, md, lg, xl, xl2                      widget.Clickable
	teal, pink, tealSoft                              widget.Clickable
	loading, disabled                                 widget.Clickable
}

var tealSoftScheme = lotusui.Teal.SoftScheme()

var btnMore struct {
	basic, iconL, iconR, loadText widget.Clickable
}

func buttonDemo(th *lotusui.Theme, gtx C) D {
	return card(th, gtx,
		section(th, "Button", lotusui.HStack(lotusui.Space.SM,
			lotusui.Button(th, &btnMore.basic, "Button", lotusui.ButtonProps{}),
		)),
		section(th, "Secondary", lotusui.HStack(lotusui.Space.SM,
			lotusui.Button(th, &btn.secondary, "Secondary", lotusui.ButtonProps{Variant: lotusui.ButtonSecondary}),
		)),
		section(th, "Destructive", lotusui.HStack(lotusui.Space.SM,
			lotusui.Button(th, &btn.destructive, "Destructive", lotusui.ButtonProps{Variant: lotusui.ButtonDestructive}),
		)),
		section(th, "Outline", lotusui.HStack(lotusui.Space.SM,
			lotusui.Button(th, &btn.outline, "Outline", lotusui.ButtonProps{Variant: lotusui.ButtonOutline}),
		)),
		section(th, "Ghost", lotusui.HStack(lotusui.Space.SM,
			lotusui.Button(th, &btn.ghost, "Ghost", lotusui.ButtonProps{Variant: lotusui.ButtonGhost}),
		)),
		section(th, "Link", lotusui.HStack(lotusui.Space.SM,
			lotusui.Button(th, &btn.link, "Link", lotusui.ButtonProps{Variant: lotusui.ButtonLink}),
		)),
		section(th, "Icon", lotusui.HStack(lotusui.Space.SM,
			lotusui.Button(th, &btn.iconOnly, "", lotusui.ButtonProps{Variant: lotusui.ButtonOutline, IconStart: lotusui.IconSettings}),
		)),
		section(th, "With icon", lotusui.HStack(lotusui.Space.SM,
			lotusui.Button(th, &btnMore.iconL, "Add item", lotusui.ButtonProps{IconStart: lotusui.IconAdd}),
			lotusui.Button(th, &btnMore.iconR, "Continue", lotusui.ButtonProps{Variant: lotusui.ButtonOutline, IconEnd: lotusui.IconExpand}),
		)),
		section(th, "Loading", lotusui.HStack(lotusui.Space.SM,
			lotusui.Button(th, &btn.loading, "Loading", lotusui.ButtonProps{Loading: true}),
			lotusui.Button(th, &btnMore.loadText, "Please wait", lotusui.ButtonProps{Loading: true, LoadingText: "Please wait"}),
		)),
		section(th, "Sizes — 2XS to 2XL", lotusui.HStack(lotusui.Space.SM,
			lotusui.Button(th, &btn.xs2, "2XS", lotusui.ButtonProps{Size: lotusui.Size2XS}),
			lotusui.Button(th, &btn.xs, "XS", lotusui.ButtonProps{Size: lotusui.SizeXS}),
			lotusui.Button(th, &btn.sm, "SM", lotusui.ButtonProps{Size: lotusui.SizeSM}),
			lotusui.Button(th, &btn.md, "MD", lotusui.ButtonProps{}),
			lotusui.Button(th, &btn.lg, "LG", lotusui.ButtonProps{Size: lotusui.SizeLG}),
			lotusui.Button(th, &btn.xl, "XL", lotusui.ButtonProps{Size: lotusui.SizeXL}),
			lotusui.Button(th, &btn.xl2, "2XL", lotusui.ButtonProps{Size: lotusui.Size2XL}),
		)),
		section(th, "Color — a ColorScale re-colors the variant", lotusui.HStack(lotusui.Space.SM,
			lotusui.Button(th, &btn.teal, "Teal", lotusui.ButtonProps{Color: lotusui.Teal}),
			lotusui.Button(th, &btn.pink, "Pink", lotusui.ButtonProps{Color: lotusui.Pink}),
		)),
		section(th, "Scheme — full manual control of every slot", lotusui.HStack(lotusui.Space.SM,
			lotusui.Button(th, &btn.tealSoft, "Teal.SoftScheme()", lotusui.ButtonProps{Scheme: &tealSoftScheme}),
		)),
		section(th, "Disabled", lotusui.HStack(lotusui.Space.SM,
			lotusui.Button(th, &btn.disabled, "Disabled", lotusui.ButtonProps{Disabled: true}),
		)),
		section(th, "Group — HStack IS the group", lotusui.HStack(lotusui.Space.SM,
			lotusui.Button(th, &btn.groupCancel, "Cancel", lotusui.ButtonProps{Variant: lotusui.ButtonOutline}),
			lotusui.Button(th, &btn.groupSave, "Save", lotusui.ButtonProps{}),
		)),
	)
}

// ---- menu ----

var menuBtns struct {
	a, b, c, d, e, f, g widget.Clickable
	chk1, chk2          widget.Clickable
	cxProfile           widget.Clickable
	cxSettings          widget.Clickable
	cxUrls, cxDelete    widget.Clickable
	showUrls            bool
	trg                 [10]lotusui.DropdownMenuTrigger
	sub                 lotusui.DropdownMenuSub
	subA, subB, subC    widget.Clickable
	subSave, subShort   widget.Clickable
	subName, subDev     widget.Clickable
	r1, r2, r3          widget.Clickable
	showStatus          bool
	showActivity        bool
	position            int
	nMail, nSms, nPush  widget.Clickable
	notifMail           bool
	notifSms            bool
	notifPush           bool
	pmCard, pmPp, pmBk  widget.Clickable
	payment             int
}

func menuDemo(th *lotusui.Theme, gtx C) D {
	if menuBtns.chk1.Clicked(gtx) {
		menuBtns.showStatus = !menuBtns.showStatus
	}
	if menuBtns.chk2.Clicked(gtx) {
		menuBtns.showActivity = !menuBtns.showActivity
	}
	if menuBtns.cxUrls.Clicked(gtx) {
		menuBtns.showUrls = !menuBtns.showUrls
	}
	for i, b := range []*widget.Clickable{&menuBtns.r1, &menuBtns.r2, &menuBtns.r3} {
		if b.Clicked(gtx) {
			menuBtns.position = i
		}
	}
	menuBtns.trg[3].KeepOpen = true // checkboxes: picking is not leaving
	menuBtns.trg[4].KeepOpen = true // radio group too
	menuBtns.trg[7].KeepOpen = true
	menuBtns.trg[8].KeepOpen = true
	if menuBtns.nMail.Clicked(gtx) {
		menuBtns.notifMail = !menuBtns.notifMail
	}
	if menuBtns.nSms.Clicked(gtx) {
		menuBtns.notifSms = !menuBtns.notifSms
	}
	if menuBtns.nPush.Clicked(gtx) {
		menuBtns.notifPush = !menuBtns.notifPush
	}
	for i, b := range []*widget.Clickable{&menuBtns.pmCard, &menuBtns.pmPp, &menuBtns.pmBk} {
		if b.Clicked(gtx) {
			menuBtns.payment = i
		}
	}
	return card(th, gtx,
		section(th, "Usage — the trigger opens the floating panel", func(gtx C) D {
			return menuBtns.trg[0].Layout(th, gtx, "Open",
				lotusui.DropdownMenuLabel(th, "My Account"),
				lotusui.DropdownMenuItem(th, &menuBtns.a, "Profile", false),
				lotusui.DropdownMenuItem(th, &menuBtns.b, "Billing", false),
				lotusui.DropdownMenuSeparator(th),
				lotusui.DropdownMenuItem(th, &menuBtns.c, "Team", false),
				lotusui.DropdownMenuItem(th, &menuBtns.g, "Subscription", false),
			)
		}),
		section(th, "Icons", func(gtx C) D {
			return menuBtns.trg[1].Layout(th, gtx, "Open",
				lotusui.DropdownMenuItemIcon(th, &menuBtns.d, lotusui.IconEdit, "Rename", false),
				lotusui.DropdownMenuItemIcon(th, &menuBtns.e, lotusui.IconSettings, "Settings", false),
			)
		}),
		section(th, "Shortcuts — display-only hints", func(gtx C) D {
			return menuBtns.trg[2].Layout(th, gtx, "Open",
				lotusui.DropdownMenuShortcutItem(th, &menuBtns.a, "Save", "⌘S", false),
				lotusui.DropdownMenuShortcutItem(th, &menuBtns.b, "Duplicate", "⇧⌘D", false),
			)
		}),
		section(th, "Checkboxes — the menu stays open while picking", func(gtx C) D {
			return menuBtns.trg[3].Layout(th, gtx, "Open",
				lotusui.DropdownMenuCheckboxItem(th, &menuBtns.chk1, "Show status bar", menuBtns.showStatus),
				lotusui.DropdownMenuCheckboxItem(th, &menuBtns.chk2, "Show activity", menuBtns.showActivity),
			)
		}),
		section(th, "Checkboxes with icons", func(gtx C) D {
			menuBtns.trg[7].Width = 240
			return menuBtns.trg[7].Layout(th, gtx, "Notifications",
				lotusui.DropdownMenuLabel(th, "Notification Preferences"),
				lotusui.DropdownMenuCheckboxItemIcon(th, &menuBtns.nMail, lotusui.IconMail, "Email notifications", menuBtns.notifMail),
				lotusui.DropdownMenuCheckboxItemIcon(th, &menuBtns.nSms, lotusui.IconMessage, "SMS notifications", menuBtns.notifSms),
				lotusui.DropdownMenuCheckboxItemIcon(th, &menuBtns.nPush, lotusui.IconBell, "Push notifications", menuBtns.notifPush),
			)
		}),
		section(th, "Radio group with icons", func(gtx C) D {
			menuBtns.trg[8].Width = 240
			return menuBtns.trg[8].Layout(th, gtx, "Payment Method",
				lotusui.DropdownMenuLabel(th, "Select Payment Method"),
				lotusui.DropdownMenuRadioItemIcon(th, &menuBtns.pmCard, lotusui.IconCreditCard, "Credit Card", menuBtns.payment == 0),
				lotusui.DropdownMenuRadioItemIcon(th, &menuBtns.pmPp, lotusui.IconWallet, "PayPal", menuBtns.payment == 1),
				lotusui.DropdownMenuRadioItemIcon(th, &menuBtns.pmBk, lotusui.IconBuilding, "Bank Transfer", menuBtns.payment == 2),
			)
		}),
		section(th, "Radio group — one selected, caller-owned", func(gtx C) D {
			return menuBtns.trg[4].Layout(th, gtx, "Open",
				lotusui.DropdownMenuLabel(th, "Panel position"),
				lotusui.DropdownMenuRadioItem(th, &menuBtns.r1, "Top", menuBtns.position == 0),
				lotusui.DropdownMenuRadioItem(th, &menuBtns.r2, "Bottom", menuBtns.position == 1),
				lotusui.DropdownMenuRadioItem(th, &menuBtns.r3, "Right", menuBtns.position == 2),
			)
		}),
		section(th, "Destructive", func(gtx C) D {
			return menuBtns.trg[5].Layout(th, gtx, "Open",
				lotusui.DropdownMenuItem(th, &menuBtns.f, "Delete workspace…", true),
			)
		}),
		section(th, "Submenu — a side panel on hover", func(gtx C) D {
			menuBtns.trg[9].KeepOpen = true
			return menuBtns.trg[9].Layout(th, gtx, "Open",
				lotusui.DropdownMenuItem(th, &menuBtns.subA, "New Tab", false),
				lotusui.DropdownMenuItem(th, &menuBtns.subB, "New Window", false),
				lotusui.DropdownMenuSeparator(th),
				menuBtns.sub.Item(th, "More Tools",
					lotusui.DropdownMenuItem(th, &menuBtns.subSave, "Save Page As…", false),
					lotusui.DropdownMenuItem(th, &menuBtns.subShort, "Create Shortcut…", false),
					lotusui.DropdownMenuItem(th, &menuBtns.subName, "Name Window…", false),
					lotusui.DropdownMenuSeparator(th),
					lotusui.DropdownMenuItem(th, &menuBtns.subDev, "Developer Tools", false),
				),
				lotusui.DropdownMenuSeparator(th),
				lotusui.DropdownMenuItem(th, &menuBtns.subC, "Close Window", false),
			)
		}),
		section(th, "Complex — groups, icons, shortcuts, toggles together", func(gtx C) D {
			menuBtns.trg[6].Width = 260
			return menuBtns.trg[6].Layout(th, gtx, "Open",
				lotusui.DropdownMenuLabel(th, "My Account"),
				lotusui.DropdownMenuShortcutItem(th, &menuBtns.cxProfile, "Profile", "⇧⌘P", false),
				lotusui.DropdownMenuItemIcon(th, &menuBtns.cxSettings, lotusui.IconSettings, "Settings", false),
				lotusui.DropdownMenuSeparator(th),
				lotusui.DropdownMenuCheckboxItem(th, &menuBtns.cxUrls, "Show full URLs", menuBtns.showUrls),
				lotusui.DropdownMenuSeparator(th),
				lotusui.DropdownMenuItem(th, &menuBtns.cxDelete, "Delete account", true),
			)
		}),
	)
}

// ---- input ----

var tf struct {
	plain, username, errored, domain, off lotusui.Input
	sm, md, lg, xs2, xs, xl, xl2          lotusui.Input
	subtle, flushed                       lotusui.Input
	search, secret, clearable, bio        lotusui.Input
	email, withBtn                        lotusui.Input
	first, last, city, zip, token         lotusui.Input
	amount, url2, kbdSearch               lotusui.Input
	searching, copyURL, blockName         lotusui.Input
	subBtn                                widget.Clickable
	clearBtn, eyeBtn                      widget.Clickable
	copyBtn2                              widget.Clickable
	reveal                                bool
	inited                                bool
}

// addonText is the muted in-frame text addon ($, USD, https://).
func addonText(th *lotusui.Theme, s string) layout.Widget {
	return func(gtx C) D {
		l := lotusui.LabelBody(th, s)
		l.Color = th.Palette.FgSubtle
		return l.Layout(gtx)
	}
}

func inputDemo(th *lotusui.Theme, gtx C) D {
	if !tf.inited {
		tf.inited = true
		// Your app's rules, expressed with the generic mechanisms:
		// an allow-list plus a same-frame fold to lowercase.
		tf.username.Filter = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-"
		tf.username.Transform = strings.ToLower
		tf.errored.Error = "this name is already taken"
		tf.errored.Editor.SetText("taken-name")
		tf.off.Disabled = true
		tf.off.Editor.SetText("fixed at creation")
		tf.sm.Size = lotusui.SizeSM
		tf.lg.Size = lotusui.SizeLG
		tf.xs2.Size = lotusui.Size2XS
		tf.xs.Size = lotusui.SizeXS
		tf.xl.Size = lotusui.SizeXL
		tf.xl2.Size = lotusui.Size2XL
		tf.secret.Editor.SetText("correct-horse-battery")
		tf.subtle.Variant = lotusui.InputSubtle
		tf.flushed.Variant = lotusui.InputFlushed
		// Cap the bio at 80 chars WITH the generic Transform — the
		// counter below is Field.Helper, recomputed each frame.
		tf.bio.Transform = func(t string) string {
			if len(t) > 80 {
				return t[:80]
			}
			return t
		}
	}
	// Composed enrichments: a clear button and a visibility toggle are
	// just End widgets — the core knows nothing about them.
	if tf.clearBtn.Clicked(gtx) {
		tf.clearable.Editor.SetText("")
	}
	if tf.eyeBtn.Clicked(gtx) {
		tf.reveal = !tf.reveal
	}
	tf.search.Start = lotusui.SVGIcon(lotusui.IconEye, 16, th.Palette.FgSubtle)
	tf.clearable.End = lotusui.SVGIconButtonTint(th, &tf.clearBtn, lotusui.IconRemove, 14, false, th.Palette.FgSubtle)
	eye := lotusui.IconEyeOff
	// Masking is Gio's own Editor.Mask, reached through the exported
	// Editor — the eye toggles it, so the reveal is real.
	tf.secret.Editor.Mask = '•'
	if tf.reveal {
		eye = lotusui.IconEye
		tf.secret.Editor.Mask = 0
	}
	tf.secret.End = lotusui.SVGIconButtonTint(th, &tf.eyeBtn, eye, 16, false, th.Palette.FgSubtle)

	return card(th, gtx,
		section(th, "Input", func(gtx C) D {
			return tf.plain.LayoutField(th, gtx, "Email")
		}),
		section(th, "Disabled", func(gtx C) D {
			return tf.off.LayoutField(th, gtx, "")
		}),
		section(th, "With label", func(gtx C) D {
			return lotusui.Field(th, lotusui.FieldProps{Label: "Email"},
				func(gtx C) D { return tf.email.LayoutField(th, gtx, "you@example.com") })(gtx)
		}),
		section(th, "With button", func(gtx C) D {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Flexed(1, func(gtx C) D { return tf.withBtn.LayoutField(th, gtx, "you@example.com") }),
				layout.Rigid(lotusui.HSpacer(th.Space.SM)),
				layout.Rigid(lotusui.Button(th, &tf.subBtn, "Subscribe", lotusui.ButtonProps{Variant: lotusui.ButtonOutline})),
			)
		}),
		section(th, "Variants — outline, subtle, flushed", lotusui.VStack(th.Space.SM,
			func(gtx C) D { return tf.plain.LayoutField(th, gtx, "Outline (default)") },
			func(gtx C) D { return tf.subtle.LayoutField(th, gtx, "Subtle") },
			func(gtx C) D { return tf.flushed.LayoutField(th, gtx, "Flushed") },
		)),
		section(th, "Sizes — 2XS to 2XL", lotusui.VStack(th.Space.SM,
			func(gtx C) D { return tf.xs2.LayoutField(th, gtx, "2XS") },
			func(gtx C) D { return tf.xs.LayoutField(th, gtx, "XS") },
			func(gtx C) D { return tf.sm.LayoutField(th, gtx, "SM") },
			func(gtx C) D { return tf.md.LayoutField(th, gtx, "MD") },
			func(gtx C) D { return tf.lg.LayoutField(th, gtx, "LG") },
			func(gtx C) D { return tf.xl.LayoutField(th, gtx, "XL") },
			func(gtx C) D { return tf.xl2.LayoutField(th, gtx, "2XL") },
		)),
		section(th, "Start and End elements", lotusui.VStack(th.Space.SM,
			func(gtx C) D { return tf.search.LayoutField(th, gtx, "Search…") },
			func(gtx C) D { return tf.secret.LayoutField(th, gtx, "Password") },
		)),
		section(th, "Suffix addon", func(gtx C) D {
			return tf.domain.LayoutSuffix(th, gtx, "Subdomain", "yourname", ".example.com")
		}),
		section(th, "Invalid — Field error + danger chrome", func(gtx C) D {
			return tf.errored.Layout(th, gtx, "Workspace name", "my-workspace")
		}),
		section(th, "Field — label, helper, required", func(gtx C) D {
			return lotusui.Field(th, lotusui.FieldProps{Label: "Email", Helper: "We'll never share it.", Required: true},
				func(gtx C) D { return tf.email.LayoutField(th, gtx, "you@example.com") })(gtx)
		}),
		section(th, "Filter and Transform — typed A appears as a", func(gtx C) D {
			return tf.username.Layout(th, gtx, "Username", "lowercase-only")
		}),
		section(th, "Clear button — an End widget, composed", func(gtx C) D {
			return tf.clearable.LayoutField(th, gtx, "Type, then clear")
		}),
		section(th, "Character counter — Field.Helper, composed", func(gtx C) D {
			n := len(tf.bio.Editor.Text())
			return lotusui.Field(th, lotusui.FieldProps{Label: "Bio", Helper: fmt.Sprintf("%d / 80", n)},
				func(gtx C) D { return tf.bio.LayoutField(th, gtx, "A line about you") })(gtx)
		}),
		section(th, "Grid — inputs side by side", func(gtx C) D {
			return layout.Flex{}.Layout(gtx,
				layout.Flexed(1, lotusui.Field(th, lotusui.FieldProps{Label: "First name"},
					func(gtx C) D { return tf.first.LayoutField(th, gtx, "Ada") })),
				layout.Rigid(lotusui.HSpacer(th.Space.MD)),
				layout.Flexed(1, lotusui.Field(th, lotusui.FieldProps{Label: "Last name"},
					func(gtx C) D { return tf.last.LayoutField(th, gtx, "Lovelace") })),
			)
		}),
		section(th, "With badge — compose the label row", func(gtx C) D {
			return lotusui.VStack(th.Space.XS,
				lotusui.HStack(th.Space.SM,
					lotusui.SectionLabel(th, "API token"),
					lotusui.Badge(th, "Beta", lotusui.BadgeProps{Variant: lotusui.BadgeSecondary, Size: lotusui.SizeXS}),
				),
				func(gtx C) D { return tf.token.LayoutField(th, gtx, "tok_…") },
			)(gtx)
		}),
		section(th, "Field group — stacked fields read as one form", func(gtx C) D {
			return lotusui.Card(th, lotusui.CardProps{}, lotusui.VStack(th.Space.MD,
				lotusui.Field(th, lotusui.FieldProps{Label: "City"},
					func(gtx C) D { return tf.city.LayoutField(th, gtx, "Lyon") }),
				lotusui.Field(th, lotusui.FieldProps{Label: "Postal code"},
					func(gtx C) D { return tf.zip.LayoutField(th, gtx, "69000") }),
			))(gtx)
		}),
		section(th, "Text addons — muted segments in the frame", lotusui.VStack(th.Space.SM,
			func(gtx C) D {
				tf.amount.Start, tf.amount.End = addonText(th, "$"), addonText(th, "USD")
				return tf.amount.LayoutField(th, gtx, "0.00")
			},
			func(gtx C) D {
				tf.url2.Start, tf.url2.End = addonText(th, "https://"), addonText(th, ".com")
				return tf.url2.LayoutField(th, gtx, "example")
			},
		)),
		section(th, "Kbd — the shortcut hint", func(gtx C) D {
			tf.kbdSearch.Start = lotusui.SVGIcon(lotusui.IconSearch, 16, th.Palette.FgSubtle)
			tf.kbdSearch.End = lotusui.Kbd(th, "⌘K")
			return tf.kbdSearch.LayoutField(th, gtx, "Search...")
		}),
		section(th, "Spinner — the in-flight field", func(gtx C) D {
			tf.searching.End = lotusui.Spinner(th, 14)
			return tf.searching.LayoutField(th, gtx, "Searching...")
		}),
		section(th, "Inline button", func(gtx C) D {
			tf.copyURL.End = lotusui.SVGIconButtonTint(th, &tf.copyBtn2, lotusui.IconFile, 14, false, th.Palette.FgSubtle)
			if tf.copyURL.Editor.Text() == "" {
				tf.copyURL.Editor.SetText("https://lotusui.com")
			}
			return tf.copyURL.LayoutField(th, gtx, "")
		}),
		section(th, "Block addons — rows inside the frame", func(gtx C) D {
			tf.blockName.Top = func(gtx C) D {
				l := lotusui.LabelCaption(th, "Full Name")
				l.Color = th.Palette.FgSubtle
				return l.Layout(gtx)
			}
			return tf.blockName.LayoutField(th, gtx, "Enter your name")
		}),
	)
}

// ---- checkbox ----

var cb struct {
	a, b, off, parent, invalid lotusui.Checkbox
	g1, g2, g3                 lotusui.Checkbox
	t1, t2, t3                 lotusui.Checkbox
	sm, md, lg                 lotusui.Checkbox
	xs2, xs, xl, xl2           lotusui.Checkbox
	inited                     bool
}

func checkboxDemo(th *lotusui.Theme, gtx C) D {
	if !cb.inited {
		cb.inited = true
		cb.a.Value = true
		cb.off.Value = true
		cb.off.Disabled = true
		cb.parent.Indeterminate = true
		cb.invalid.Invalid = true
		cb.sm.Size = lotusui.SizeSM
		cb.lg.Size = lotusui.SizeLG
		cb.xs2.Size = lotusui.Size2XS
		cb.xs.Size = lotusui.SizeXS
		cb.xl.Size = lotusui.SizeXL
		cb.xl2.Size = lotusui.Size2XL
	}
	if cb.a.Clicked(gtx) {
		cb.a.Value = !cb.a.Value
	}
	if cb.b.Clicked(gtx) {
		cb.b.Value = !cb.b.Value
	}
	for _, c := range []*lotusui.Checkbox{&cb.g1, &cb.g2, &cb.g3, &cb.t1, &cb.t2, &cb.t3} {
		if c.Clicked(gtx) {
			c.Value = !c.Value
		}
	}
	return card(th, gtx,
		section(th, "Checkbox", lotusui.VStack(lotusui.Space.SM,
			func(gtx C) D { return cb.a.Layout(th, gtx, "Accept terms and conditions") },
		)),
		section(th, "With text", func(gtx C) D {
			return lotusui.VStack(lotusui.Space.XS,
				func(gtx C) D { return cb.b.Layout(th, gtx, "Accept terms and conditions") },
				func(gtx C) D {
					return layout.Inset{Left: 26}.Layout(gtx,
						lotusui.LabelCaption(th, "By clicking this checkbox, you agree to the terms and conditions.").Layout)
				},
			)(gtx)
		}),
		section(th, "Group — a list of caller-owned values", lotusui.VStack(lotusui.Space.SM,
			func(gtx C) D { return cb.g1.Layout(th, gtx, "Recents") },
			func(gtx C) D { return cb.g2.Layout(th, gtx, "Home") },
			func(gtx C) D { return cb.g3.Layout(th, gtx, "Applications") },
		)),
		section(th, "Disabled — a fact, not a choice", func(gtx C) D {
			return cb.off.Layout(th, gtx, "Included in your plan")
		}),
		section(th, "Indeterminate and invalid", lotusui.VStack(lotusui.Space.SM,
			func(gtx C) D { return cb.parent.Layout(th, gtx, "Select all (some selected)") },
			func(gtx C) D { return cb.invalid.Layout(th, gtx, "You must accept the terms") },
		)),
		section(th, "In a table — row selection is composition", func(gtx C) D {
			cell := func(w layout.Widget) layout.Widget { return w }
			label := func(t string) layout.Widget {
				return func(gtx C) D {
					l := lotusui.LabelBody(th, t)
					l.Color = th.Palette.Fg
					return l.Layout(gtx)
				}
			}
			return lotusui.Table(th, lotusui.TableProps{Widths: []float32{0.5, 2, 1}},
				[]string{"", "Repository", "Visibility"},
				[][]layout.Widget{
					{cell(func(gtx C) D { return cb.t1.Layout(th, gtx, "") }), label("lotusui"), label("Public")},
					{cell(func(gtx C) D { return cb.t2.Layout(th, gtx, "") }), label("vaultalia"), label("Private")},
					{cell(func(gtx C) D { return cb.t3.Layout(th, gtx, "") }), label("mandarin"), label("Private")},
				})(gtx)
		}),
		section(th, "Sizes — 2XS to 2XL", lotusui.HStack(lotusui.Space.MD,
			func(gtx C) D { return cb.xs2.Layout(th, gtx, "2XS") },
			func(gtx C) D { return cb.xs.Layout(th, gtx, "XS") },
			func(gtx C) D { return cb.sm.Layout(th, gtx, "SM") },
			func(gtx C) D { return cb.md.Layout(th, gtx, "MD") },
			func(gtx C) D { return cb.lg.Layout(th, gtx, "LG") },
			func(gtx C) D { return cb.xl.Layout(th, gtx, "XL") },
			func(gtx C) D { return cb.xl2.Layout(th, gtx, "2XL") },
		)),
	)
}

// ---- code-block ----

var codeBlk struct {
	copy widget.Clickable
}

func codeBlockDemo(th *lotusui.Theme, gtx C) D {
	plain := "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"hello\")\n}"
	hi := [][]lotusui.CodeSpan{
		{{Text: "package ", Color: th.Palette.BrandFg, Bold: true}, {Text: "main", Color: th.Palette.Fg}},
		{},
		{{Text: "import ", Color: th.Palette.BrandFg, Bold: true}, {Text: "\"fmt\"", Color: th.Palette.Success}},
		{},
		{{Text: "func ", Color: th.Palette.BrandFg, Bold: true}, {Text: "main", Color: th.Palette.Fg}, {Text: "() {", Color: th.Palette.FgSubtle}},
		{{Text: "\tfmt.Println(", Color: th.Palette.Fg}, {Text: "\"hello\"", Color: th.Palette.Success}, {Text: ")", Color: th.Palette.Fg}},
		{{Text: "}", Color: th.Palette.FgSubtle}},
	}
	return card(th, gtx,
		section(th, "Usage", lotusui.CodeBlock(th, lotusui.CodeBlockProps{
			Lang: "go", Plain: `fmt.Println("hello")`,
		})),
		section(th, "Highlighted spans", lotusui.CodeBlock(th, lotusui.CodeBlockProps{
			Lang: "go", Lines: hi, Plain: plain,
		})),
		section(th, "Copy", lotusui.CodeBlock(th, lotusui.CodeBlockProps{
			Lang: "go", Plain: plain, Copy: &codeBlk.copy,
		})),
		section(th, "Nested", func(gtx C) D {
			return lotusui.Card(th, lotusui.CardProps{},
				lotusui.CodeBlock(th, lotusui.CodeBlockProps{
					Nested: true, Lang: "go", Plain: `x := 1`,
				}),
			)(gtx)
		}),
	)
}

// ---- example ----

var exDemo struct {
	a, b, c lotusui.Example
	copy    widget.Clickable
}

func exampleDemo(th *lotusui.Theme, gtx C) D {
	src := `fmt.Println("hello")`
	preview := func(gtx C) D {
		return layout.UniformInset(lotusui.Space.MD).Layout(gtx, func(gtx C) D {
			return lotusui.LabelBody(th, "Live preview body").Layout(gtx)
		})
	}
	code := lotusui.CodeBlock(th, lotusui.CodeBlockProps{
		Nested: true, Lang: "go", Plain: src,
		Lines: [][]lotusui.CodeSpan{
			{{Text: "fmt.Println(", Color: th.Palette.Fg}, {Text: `"hello"`, Color: th.Palette.Success}, {Text: ")", Color: th.Palette.Fg}},
		},
		Copy: &exDemo.copy,
	})
	return card(th, gtx,
		section(th, "Usage", func(gtx C) D {
			return exDemo.a.Layout(th, gtx, lotusui.ExampleProps{Preview: preview, Code: code})
		}),
		section(th, "Preview only", func(gtx C) D {
			return exDemo.b.Layout(th, gtx, lotusui.ExampleProps{Preview: preview})
		}),
		section(th, "With CodeBlock", func(gtx C) D {
			return exDemo.c.Layout(th, gtx, lotusui.ExampleProps{Preview: preview, Code: code})
		}),
	)
}

// ---- scroll-area ----

var saDemo struct {
	vert, horiz, float     lotusui.ScrollArea
	always, sizeSM, sizeLG lotusui.ScrollArea
	color, track           lotusui.ScrollArea
	env                    lotusui.Select
	inited                 bool
}

func scrollAreaDemo(th *lotusui.Theme, gtx C) D {
	if !saDemo.inited {
		saDemo.inited = true
		saDemo.env.Options = lotusui.SelectOpts("Development", "Staging", "Production")
		saDemo.env.SetValue("Development")
	}
	tags := make([]layout.Widget, 12)
	for i := range tags {
		i := i
		tags[i] = lotusui.Badge(th, "Tag "+strconv.Itoa(i+1), lotusui.BadgeProps{Variant: lotusui.BadgeOutline})
	}
	tall := make([]layout.Widget, 16)
	for i := range tall {
		i := i
		tall[i] = lotusui.LabelBody(th, "Row "+strconv.Itoa(i+1)+" — scroll the viewport").Layout
	}
	pane := func(sa *lotusui.ScrollArea, props lotusui.ScrollAreaProps, body layout.Widget) layout.Widget {
		return func(gtx C) D {
			gtx.Constraints.Max.Y = gtx.Dp(unit.Dp(200))
			gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
			props.NoShadowRoom = true
			return lotusui.Card(th, lotusui.CardProps{}, func(gtx C) D {
				return sa.LayoutWith(th, gtx, props, body)
			})(gtx)
		}
	}
	return card(th, gtx,
		section(th, "Usage", pane(&saDemo.vert, lotusui.ScrollAreaProps{},
			lotusui.VStack(lotusui.Space.SM, tall...))),
		section(th, "Horizontal", func(gtx C) D {
			gtx.Constraints.Max.Y = gtx.Dp(unit.Dp(72))
			gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
			saDemo.horiz.Horizontal = true
			return lotusui.Card(th, lotusui.CardProps{}, func(gtx C) D {
				return saDemo.horiz.LayoutWith(th, gtx, lotusui.ScrollAreaProps{NoShadowRoom: true},
					lotusui.HStack(lotusui.Space.SM, tags...))
			})(gtx)
		}),
		section(th, "Floating inside", func(gtx C) D {
			gtx.Constraints.Max.Y = gtx.Dp(unit.Dp(240))
			gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
			return lotusui.Card(th, lotusui.CardProps{}, func(gtx C) D {
				return saDemo.float.LayoutWith(th, gtx, lotusui.ScrollAreaProps{NoShadowRoom: true},
					lotusui.VStack(lotusui.Space.MD,
						lotusui.LabelBody(th, "Scroll, then open the select — the panel escapes the viewport.").Layout,
						func(gtx C) D { return saDemo.env.Layout(th, gtx, "Environment") },
						lotusui.LabelMeta(th, "More content below keeps the area scrollable.").Layout,
						lotusui.LabelBody(th, "Extra row A").Layout,
						lotusui.LabelBody(th, "Extra row B").Layout,
						lotusui.LabelBody(th, "Extra row C").Layout,
						lotusui.LabelBody(th, "Extra row D").Layout,
					))
			})(gtx)
		}),
		section(th, "Always visible", pane(&saDemo.always, lotusui.ScrollAreaProps{
			Scrollbar: lotusui.ScrollbarProps{Variant: lotusui.ScrollbarAlways},
		}, lotusui.VStack(lotusui.Space.SM, tall...))),
		section(th, "Sizes", func(gtx C) D {
			return lotusui.HStack(lotusui.Space.MD,
				pane(&saDemo.sizeSM, lotusui.ScrollAreaProps{
					Scrollbar: lotusui.ScrollbarProps{Variant: lotusui.ScrollbarAlways, Size: lotusui.SizeSM},
				}, lotusui.VStack(lotusui.Space.SM, tall...)),
				pane(&saDemo.sizeLG, lotusui.ScrollAreaProps{
					Scrollbar: lotusui.ScrollbarProps{Variant: lotusui.ScrollbarAlways, Size: lotusui.SizeLG},
				}, lotusui.VStack(lotusui.Space.SM, tall...)),
			)(gtx)
		}),
		section(th, "Thumb color", pane(&saDemo.color, lotusui.ScrollAreaProps{
			Scrollbar: lotusui.ScrollbarProps{
				Variant: lotusui.ScrollbarAlways,
				Color:   lotusui.Teal,
			},
		}, lotusui.VStack(lotusui.Space.SM, tall...))),
		section(th, "Show track", pane(&saDemo.track, lotusui.ScrollAreaProps{
			Scrollbar: lotusui.ScrollbarProps{
				Variant:   lotusui.ScrollbarAlways,
				ShowTrack: true,
			},
		}, lotusui.VStack(lotusui.Space.SM, tall...))),
	)
}

// ---- field ----

var fld struct {
	email, name lotusui.Input
	role        lotusui.Select
	inited      bool
}

func fieldDemo(th *lotusui.Theme, gtx C) D {
	if !fld.inited {
		fld.inited = true
		fld.name.Error = "this name is already taken"
		fld.name.Editor.SetText("taken-name")
		fld.role.Options = lotusui.SelectOpts("Viewer", "Editor", "Admin")
		fld.role.Clear()
		fld.role.Placeholder = "Choose a role…"
	}
	return card(th, gtx,
		section(th, "Field — label above any control", func(gtx C) D {
			return lotusui.Field(th, lotusui.FieldProps{Label: "Email"}, func(gtx C) D {
				return fld.email.LayoutField(th, gtx, "you@example.com")
			})(gtx)
		}),
		section(th, "Helper text", func(gtx C) D {
			return lotusui.Field(th, lotusui.FieldProps{Label: "Email", Helper: "We'll never share it."}, func(gtx C) D {
				return fld.email.LayoutField(th, gtx, "you@example.com")
			})(gtx)
		}),
		section(th, "Required", func(gtx C) D {
			return lotusui.Field(th, lotusui.FieldProps{Label: "Email", Required: true}, func(gtx C) D {
				return fld.email.LayoutField(th, gtx, "you@example.com")
			})(gtx)
		}),
		section(th, "Error", func(gtx C) D {
			return lotusui.Field(th, lotusui.FieldProps{Label: "Workspace name", Error: fld.name.Error}, func(gtx C) D {
				return fld.name.LayoutField(th, gtx, "my-workspace")
			})(gtx)
		}),
		section(th, "Any control — a Select in a Field", func(gtx C) D {
			return lotusui.Field(th, lotusui.FieldProps{Label: "Role"}, func(gtx C) D {
				return fld.role.Layout(th, gtx, "")
			})(gtx)
		}),
	)
}

// ---- switch ----

var sws struct {
	a, b, sm, md, lg lotusui.Switch
	xs2, xs, xl, xl2 lotusui.Switch
	cc1, cc2         lotusui.Switch
	offOn, off, bad  lotusui.Switch
	inited           bool
}

func switchState(state string) { sws.a.Value = state == "on" }

func switchDemo(th *lotusui.Theme, gtx C) D {
	if !sws.inited {
		sws.inited = true
		sws.sm.Size = lotusui.SizeSM
		sws.lg.Size = lotusui.SizeLG
		sws.xs2.Size = lotusui.Size2XS
		sws.xs.Size = lotusui.SizeXS
		sws.xl.Size = lotusui.SizeXL
		sws.xl2.Size = lotusui.Size2XL
		sws.offOn.Value = true
		sws.offOn.Disabled = true
		sws.off.Disabled = true
		sws.bad.Invalid = true
	}
	row := func(s *lotusui.Switch, label string) layout.Widget {
		return lotusui.HStack(lotusui.Space.SM,
			func(gtx C) D { return s.Layout(th, gtx) },
			lotusui.LabelBody(th, label).Layout,
		)
	}
	return card(th, gtx,
		section(th, "Switch — animated on the shared clock", lotusui.VStack(lotusui.Space.MD,
			row(&sws.a, "Airplane Mode"),
		)),
		section(th, "Sizes — 2XS to 2XL", lotusui.HStack(lotusui.Space.MD,
			func(gtx C) D { return sws.xs2.Layout(th, gtx) },
			func(gtx C) D { return sws.xs.Layout(th, gtx) },
			func(gtx C) D { return sws.sm.Layout(th, gtx) },
			func(gtx C) D { return sws.md.Layout(th, gtx) },
			func(gtx C) D { return sws.lg.Layout(th, gtx) },
			func(gtx C) D { return sws.xl.Layout(th, gtx) },
			func(gtx C) D { return sws.xl2.Layout(th, gtx) },
		)),
		section(th, "Disabled — on and off, both facts", lotusui.HStack(lotusui.Space.MD,
			row(&sws.offOn, "Included in your plan"),
			row(&sws.off, "Requires an upgrade"),
		)),
		section(th, "Invalid", row(&sws.bad, "You must enable this to continue")),
		section(th, "Choice card — settings rows in a Card", func(gtx C) D {
			line := func(s *lotusui.Switch, title, sub string) layout.Widget {
				return func(gtx C) D {
					return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
						layout.Flexed(1, lotusui.VStack(2,
							lotusui.LabelBody(th, title).Layout,
							lotusui.LabelCaption(th, sub).Layout,
						)),
						layout.Rigid(func(gtx C) D { return s.Layout(th, gtx) }),
					)
				}
			}
			return lotusui.Card(th, lotusui.CardProps{}, lotusui.VStack(th.Space.MD,
				line(&sws.cc1, "Marketing emails", "Receive emails about new products, features, and more."),
				lotusui.Hairline(th),
				line(&sws.cc2, "Security emails", "Receive emails about your account security."),
			))(gtx)
		}),
	)
}

// ---- dropdown ----

var dd struct {
	env, frozen, fresh lotusui.Select
	small, large, long lotusui.Select
	s2xs, sxs, smd     lotusui.Select
	sxl, s2xl          lotusui.Select
	aligned            lotusui.Select
	grouped            lotusui.Select
	meta               lotusui.Select
	icons              lotusui.Select
	plan               lotusui.Select
	inited             bool
	planCleared        bool
}

func selectDemo(th *lotusui.Theme, gtx C) D {
	planRow := func(name, desc string) layout.Widget {
		return lotusui.ItemContent(th,
			lotusui.ItemTitle(th, name),
			lotusui.ItemDescription(th, desc),
		)
	}
	if !dd.inited {
		dd.inited = true
		dd.env.Options = lotusui.SelectOpts("Apple", "Banana", "Blueberry", "Grapes", "Pineapple")
		dd.env.Clear()
		dd.env.Placeholder = "Select a fruit"
		dd.grouped.Groups = lotusui.SelectGroups(
			lotusui.SelectGrouped("Fruits", lotusui.SelectOpts("Apple", "Banana", "Cherry")...),
			lotusui.SelectGrouped("Vegetables", lotusui.SelectOpts("Carrot", "Leek", "Spinach")...),
		)
		dd.grouped.Clear()
		dd.grouped.Placeholder = "Pick a produce…"
		dd.aligned.Options = lotusui.SelectOpts("Inter", "Lora", "Nunito", "DM Sans", "Baloo")
		dd.aligned.SetValue("DM Sans")
		dd.aligned.AlignItemWithTrigger = true
		dd.long.Options = lotusui.SelectOpts(
			"(GMT-11:00) Midway Island", "(GMT-10:00) Hawaii", "(GMT-8:00) Alaska",
			"(GMT-7:00) Pacific Time", "(GMT-6:00) Mountain Time", "(GMT-5:00) Central Time",
			"(GMT-4:00) Eastern Time", "(GMT-3:30) Newfoundland", "(GMT-3:00) Buenos Aires",
			"(GMT+0:00) London", "(GMT+1:00) Paris", "(GMT+2:00) Athens",
			"(GMT+5:30) Mumbai", "(GMT+8:00) Singapore", "(GMT+9:00) Tokyo",
		)
		dd.long.SetValue("(GMT+1:00) Paris")
		dd.meta.Options = []lotusui.SelectOption{
			{Label: "roteland", Value: "r", Meta: "1"},
			{Label: "test", Value: "t", Meta: "2"},
		}
		dd.meta.SetValue("t")
		dd.icons.Options = lotusui.SelectItems(
			lotusui.SelectOption{Label: "Line", Value: "line", Icon: lotusui.IconEdit},
			lotusui.SelectOption{Label: "Bar", Value: "bar", Icon: lotusui.IconSettings},
			lotusui.SelectOption{Label: "Pie", Value: "pie", Icon: lotusui.IconBell},
		)
		dd.icons.Clear()
		dd.icons.Placeholder = "Select a chart"
		dd.frozen.Options = lotusui.SelectOpts("PostgreSQL")
		dd.frozen.Disabled = true
		dd.fresh.Options = lotusui.SelectOpts("Small", "Medium", "Large")
		dd.fresh.Clear()
		dd.fresh.Placeholder = "Choose a size…"
		dd.fresh.Invalid = true
		dd.small.Options = lotusui.SelectOpts("SM option")
		dd.small.Size = lotusui.SizeSM
		dd.large.Options = lotusui.SelectOpts("LG option")
		dd.large.Size = lotusui.SizeLG
		for _, p := range []struct {
			s  *lotusui.Select
			sz lotusui.Size
			l  string
		}{
			{&dd.s2xs, lotusui.Size2XS, "2XS option"}, {&dd.sxs, lotusui.SizeXS, "XS option"},
			{&dd.smd, lotusui.SizeMD, "MD option"}, {&dd.sxl, lotusui.SizeXL, "XL option"},
			{&dd.s2xl, lotusui.Size2XL, "2XL option"},
		} {
			p.s.Options = lotusui.SelectOpts(p.l)
			p.s.Size = p.sz
		}
	}
	// Plan rows capture th so palette swaps stay live (same pattern as Accordion demos).
	dd.plan.Placeholder = "Select a plan"
	dd.plan.Options = lotusui.SelectItems(
		lotusui.SelectOption{Label: "Starter", Value: "starter", Content: planRow("Starter", "Perfect for individuals getting started.")},
		lotusui.SelectOption{Label: "Professional", Value: "pro", Content: planRow("Professional", "Ideal for growing teams and businesses.")},
		lotusui.SelectOption{Label: "Enterprise", Value: "ent", Content: planRow("Enterprise", "Advanced features for large organizations.")},
	)
	if !dd.planCleared {
		dd.plan.Clear()
		dd.planCleared = true
	} else if v := dd.plan.Value(); v != "" {
		dd.plan.SetValue(v) // re-resolve after Options rebuild
	}
	return card(th, gtx,
		section(th, "Select — a floating panel over the content", func(gtx C) D {
			return dd.env.Layout(th, gtx, "Fruit")
		}),
		section(th, "Align item with trigger — the native-select feel", func(gtx C) D {
			return dd.aligned.Layout(th, gtx, "Font")
		}),
		section(th, "Groups — labels and separators", func(gtx C) D {
			return dd.grouped.Layout(th, gtx, "Produce")
		}),
		section(th, "Scrollable — opens aligned to the selection", func(gtx C) D {
			return dd.long.Layout(th, gtx, "Timezone")
		}),
		section(th, "Meta — secondary text on the far right", func(gtx C) D {
			return dd.meta.Layout(th, gtx, "Project")
		}),
		section(th, "Icons — leading icon on the option row", func(gtx C) D {
			return dd.icons.Layout(th, gtx, "Chart")
		}),
		section(th, "Subscription plan — multiline Content", func(gtx C) D {
			return dd.plan.Layout(th, gtx, "Plan")
		}),
		section(th, "Disabled — identity fixed after creation", func(gtx C) D {
			return dd.frozen.Layout(th, gtx, "Engine")
		}),
		section(th, "Invalid", func(gtx C) D {
			return dd.fresh.Layout(th, gtx, "Size")
		}),
		section(th, "Sizes — 2XS to 2XL", lotusui.VStack(th.Space.SM,
			func(gtx C) D { return dd.s2xs.Layout(th, gtx, "") },
			func(gtx C) D { return dd.sxs.Layout(th, gtx, "") },
			func(gtx C) D { return dd.small.Layout(th, gtx, "") },
			func(gtx C) D { return dd.smd.Layout(th, gtx, "") },
			func(gtx C) D { return dd.large.Layout(th, gtx, "") },
			func(gtx C) D { return dd.sxl.Layout(th, gtx, "") },
			func(gtx C) D { return dd.s2xl.Layout(th, gtx, "") },
		)),
	)
}

// ---- tabs ----

var (
	overview   = func() []lotusui.TabOption { return lotusui.TabOpts("Overview", "Activity", "Settings") }
	tabs       = lotusui.Tabs{Options: lotusui.TabOpts("Account", "Password")}
	lineTabs   = lotusui.Tabs{Variant: lotusui.TabsLine, Options: overview()}
	subtleTabs = lotusui.Tabs{Variant: lotusui.TabsSubtle, Options: overview()}
	vertTabs   = lotusui.Tabs{Vertical: true, Options: overview()}
	iconTabs   = lotusui.Tabs{Options: []lotusui.TabOption{
		{Label: "Files", Icon: lotusui.IconFile},
		{Label: "Changes", Icon: lotusui.IconChanges},
		{Label: "Settings", Icon: lotusui.IconSettings},
	}}
	disTabs = lotusui.Tabs{Options: []lotusui.TabOption{
		{Label: "Overview"},
		{Label: "Archived", Disabled: true},
		{Label: "Settings"},
	}}
	wrapTabs = lotusui.Tabs{Options: lotusui.TabOpts(
		"Changes", "Staging", "Production", "Reviews", "Approvals", "History",
	)}
)

func tabsState(state string) {
	// Only the Usage demo (tabs/0) owns the shared Account/Password
	// strip. Other section indices use different Tabs instances —
	// never reset the interactive strip when they mount.
	if state == "0" || state == "" {
		tabs.SetValue("Account")
	}
}

func tabsDemo(th *lotusui.Theme, gtx C) D {
	tabs.Update(gtx)
	accountBody := lotusui.Card(th, lotusui.CardProps{}, lotusui.VStack(th.Space.MD,
		func(gtx C) D {
			l := lotusui.LabelBody(th, "Make changes to your account here. Click save when you're done.")
			l.Color = th.Palette.FgMuted
			return l.Layout(gtx)
		},
		lotusui.Field(th, lotusui.FieldProps{Label: "Name"}, func(gtx C) D {
			return tabAcct.name.LayoutField(th, gtx, "Pedro Duarte")
		}),
		lotusui.Field(th, lotusui.FieldProps{Label: "Username"}, func(gtx C) D {
			return tabAcct.username.LayoutField(th, gtx, "@peduarte")
		}),
		func(gtx C) D {
			return lotusui.Button(th, &tabAcct.save, "Save changes", lotusui.ButtonProps{})(gtx)
		},
	))
	passwordBody := lotusui.Card(th, lotusui.CardProps{}, lotusui.VStack(th.Space.MD,
		func(gtx C) D {
			l := lotusui.LabelBody(th, "Change your password here. After saving, you'll be logged out.")
			l.Color = th.Palette.FgMuted
			return l.Layout(gtx)
		},
		lotusui.Field(th, lotusui.FieldProps{Label: "Current password"}, func(gtx C) D {
			tabAcct.current.Editor.Mask = '•'
			return tabAcct.current.LayoutField(th, gtx, "")
		}),
		lotusui.Field(th, lotusui.FieldProps{Label: "New password"}, func(gtx C) D {
			tabAcct.newPw.Editor.Mask = '•'
			return tabAcct.newPw.LayoutField(th, gtx, "")
		}),
		func(gtx C) D {
			return lotusui.Button(th, &tabAcct.savePw, "Save password", lotusui.ButtonProps{})(gtx)
		},
	))
	return card(th, gtx,
		section(th, "Tabs — the muted well, active tab raised", func(gtx C) D {
			gtx.Constraints.Max.X = min(gtx.Constraints.Max.X, gtx.Dp(400))
			body := accountBody
			if tabs.Value() == "Password" {
				body = passwordBody
			}
			return lotusui.VStack(lotusui.Space.MD,
				func(gtx C) D { return tabs.Layout(th, gtx) },
				body,
			)(gtx)
		}),
		section(th, "Line — the classic underline strip", func(gtx C) D {
			lineTabs.Update(gtx)
			return lineTabs.Layout(th, gtx)
		}),
		section(th, "Subtle — pill-styled labels", func(gtx C) D {
			subtleTabs.Update(gtx)
			return subtleTabs.Layout(th, gtx)
		}),
		section(th, "Vertical — the strip as a column", func(gtx C) D {
			vertTabs.Update(gtx)
			return vertTabs.Layout(th, gtx)
		}),
		section(th, "Icons — each option's own glyph", func(gtx C) D {
			iconTabs.Update(gtx)
			return iconTabs.Layout(th, gtx)
		}),
		section(th, "Disabled tab — dimmed, unclickable, skipped", func(gtx C) D {
			disTabs.Update(gtx)
			return disTabs.Layout(th, gtx)
		}),
		section(th, "Wrapping strip — resize; whole tabs reflow", func(gtx C) D {
			wrapTabs.Update(gtx)
			return wrapTabs.Layout(th, gtx)
		}),
	)
}

var tabAcct struct {
	name, username lotusui.Input
	current, newPw lotusui.Input
	save, savePw   widget.Clickable
}

// ---- modal ----

var dlg struct {
	m              lotusui.Dialog
	open           widget.Clickable
	openNoClose    widget.Clickable
	openResponsive widget.Clickable
	sizeBtns       [7]widget.Clickable
	close          widget.Clickable
	name, username lotusui.Input
	isOpen         bool
	mode           int // 0 profile, 1 scrollable, 2 sticky footer
	openScroll     widget.Clickable
	openSticky     widget.Clickable
	scroll         widget.List
}

func dialogState(state string) {
	was := dlg.isOpen
	dlg.isOpen = strings.HasPrefix(state, "open")
	dlg.m.Size = lotusui.SizeMD
	dlg.m.Sizes = lotusui.ResponsiveSize{}
	dlg.m.Widths = lotusui.ResponsiveDp{}
	dlg.m.Width = 0
	switch state {
	case "open-2xs":
		dlg.m.Size = lotusui.Size2XS
	case "open-xs":
		dlg.m.Size = lotusui.SizeXS
	case "open-sm":
		dlg.m.Size = lotusui.SizeSM
	case "open-lg":
		dlg.m.Size = lotusui.SizeLG
	case "open-xl":
		dlg.m.Size = lotusui.SizeXL
	case "open-2xl":
		dlg.m.Size = lotusui.Size2XL
	case "open-responsive":
		dlg.m.Sizes = lotusui.Sizes(lotusui.SizeSM).At("md", lotusui.SizeLG).At("xl", lotusui.Size2XL)
	}
	if dlg.isOpen && !was {
		dlg.m.Appear()
	}
}

func dialogDemo(th *lotusui.Theme, gtx C) D {
	// Clicks live INSIDE each section closure so OverlayOwner records
	// the region that was actually clicked (the demo body runs once
	// per visible region in the strip).
	openAt := func(mode int, hideClose bool) {
		dlg.m.Size = lotusui.SizeMD
		dlg.m.Sizes = lotusui.ResponsiveSize{}
		dlg.m.Widths = lotusui.ResponsiveDp{}
		dlg.m.Width = 0
		dlg.m.HideClose = hideClose
		dlg.mode = mode
		dlg.isOpen = true
		OwnOverlay()
		dlg.m.Appear()
	}
	sizes := []struct {
		label string
		size  lotusui.Size
	}{
		{"2XS", lotusui.Size2XS}, {"XS", lotusui.SizeXS}, {"SM", lotusui.SizeSM},
		{"MD", lotusui.SizeMD}, {"LG", lotusui.SizeLG}, {"XL", lotusui.SizeXL},
		{"2XL", lotusui.Size2XL},
	}
	return card(th, gtx,
		section(th, "Dialog — scrim over the entire window", func(gtx C) D {
			if dlg.open.Clicked(gtx) && !dlg.isOpen {
				openAt(0, false)
			}
			return lotusui.Button(th, &dlg.open, "Open Dialog", lotusui.ButtonProps{})(gtx)
		}),
		section(th, "Sizes — each opens at that size", func(gtx C) D {
			var sizeRow []layout.Widget
			for i, sz := range sizes {
				i, sz := i, sz
				if dlg.sizeBtns[i].Clicked(gtx) {
					dlg.m.Size = sz.size
					dlg.m.Sizes = lotusui.ResponsiveSize{}
					dlg.m.Widths = lotusui.ResponsiveDp{}
					dlg.m.Width = 0
					if !dlg.isOpen {
						dlg.mode = 0
						dlg.isOpen = true
						OwnOverlay()
						dlg.m.Appear()
					}
				}
				sizeRow = append(sizeRow, lotusui.Button(th, &dlg.sizeBtns[i], sz.label,
					lotusui.ButtonProps{Variant: lotusui.ButtonOutline, Size: lotusui.SizeSM}))
			}
			return lotusui.HStack(lotusui.Space.SM, sizeRow...)(gtx)
		}),
		section(th, "No close button — the footer is the only way out", func(gtx C) D {
			if dlg.openNoClose.Clicked(gtx) && !dlg.isOpen {
				openAt(0, true)
			}
			return lotusui.Button(th, &dlg.openNoClose, "Open Dialog", lotusui.ButtonProps{Variant: lotusui.ButtonOutline})(gtx)
		}),
		section(th, "Scrollable content — the body scrolls, the frame holds", func(gtx C) D {
			if dlg.openScroll.Clicked(gtx) && !dlg.isOpen {
				openAt(1, false)
			}
			return lotusui.Button(th, &dlg.openScroll, "Scrollable Content", lotusui.ButtonProps{Variant: lotusui.ButtonOutline})(gtx)
		}),
		section(th, "Sticky footer — actions stay while content scrolls", func(gtx C) D {
			if dlg.openSticky.Clicked(gtx) && !dlg.isOpen {
				openAt(2, false)
			}
			return lotusui.Button(th, &dlg.openSticky, "Sticky Footer", lotusui.ButtonProps{Variant: lotusui.ButtonOutline})(gtx)
		}),
		section(th, "Responsive width — SM→LG→2XL at md/xl (resize)", func(gtx C) D {
			if dlg.openResponsive.Clicked(gtx) && !dlg.isOpen {
				dlg.m.Sizes = lotusui.Sizes(lotusui.SizeSM).At("md", lotusui.SizeLG).At("xl", lotusui.Size2XL)
				dlg.m.Size = lotusui.SizeMD
				dlg.m.Width = 0
				dlg.m.Widths = lotusui.ResponsiveDp{}
				dlg.m.HideClose = false
				dlg.mode = 0
				dlg.isOpen = true
				OwnOverlay()
				dlg.m.Appear()
			}
			return lotusui.VStack(th.Space.SM,
				lotusui.LabelCaption(th, "breakpoint: "+th.BreakpointName(gtx)).Layout,
				lotusui.Button(th, &dlg.openResponsive, "Open responsive Dialog", lotusui.ButtonProps{Variant: lotusui.ButtonOutline}),
			)(gtx)
		}),
	)
}

func dialogOverlay(th *lotusui.Theme, gtx C) D {
	if !dlg.isOpen {
		return D{}
	}
	if dlg.close.Clicked(gtx) {
		dlg.isOpen = false
	}
	dismiss := func() { dlg.isOpen = false }
	if dlg.mode != 0 {
		title, desc := "Scrollable Content", "This is a dialog with scrollable content."
		if dlg.mode == 2 {
			title, desc = "Sticky Footer", "This dialog has a sticky footer that stays visible while the content scrolls."
		}
		return dlg.m.Layout(th, gtx, dismiss, func(gtx C) D {
			rows := []layout.Widget{
				lotusui.LabelTitle(th, title).Layout,
				func(gtx C) D {
					l := lotusui.LabelBody(th, desc)
					l.Color = th.Palette.FgMuted
					return l.Layout(gtx)
				},
				func(gtx C) D {
					gtx.Constraints.Max.Y = gtx.Dp(300)
					gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
					dlg.scroll.Axis = layout.Vertical
					return lotusui.ListView(th, &dlg.scroll, gtx, 10, func(gtx C, i int) D {
						return layout.Inset{Bottom: lotusui.Space.MD}.Layout(gtx, func(gtx C) D {
							l := lotusui.LabelBody(th, dialogLorem)
							l.Color = th.Palette.FgMuted
							return l.Layout(gtx)
						})
					})
				},
			}
			if dlg.mode == 2 {
				rows = append(rows, lotusui.RightAligned(
					lotusui.Button(th, &dlg.close, "Close", lotusui.ButtonProps{Variant: lotusui.ButtonOutline})))
			}
			return lotusui.VStack(lotusui.Space.MD, rows...)(gtx)
		})
	}
	return dlg.m.Layout(th, gtx, dismiss, func(gtx C) D {
		return lotusui.VStack(lotusui.Space.MD,
			lotusui.LabelTitle(th, "Edit profile").Layout,
			func(gtx C) D {
				l := lotusui.LabelBody(th, "Make changes to your profile here. Click save when you're done.")
				l.Color = th.Palette.FgMuted
				return l.Layout(gtx)
			},
			lotusui.Field(th, lotusui.FieldProps{Label: "Name"}, func(gtx C) D {
				return dlg.name.LayoutField(th, gtx, "Pedro Duarte")
			}),
			lotusui.Field(th, lotusui.FieldProps{Label: "Username"}, func(gtx C) D {
				return dlg.username.LayoutField(th, gtx, "@peduarte")
			}),
			lotusui.RightAligned(lotusui.Button(th, &dlg.close, "Save changes", lotusui.ButtonProps{})),
		)(gtx)
	})
}

// ---- list (virtualized) ----

var listState struct {
	list widget.List
	rows [10000]widget.Clickable
	sel  int
}

// listEdges reports the list demo's scroll position for wheel
// chaining: at the top edge upward wheels belong to the page, at the
// bottom edge downward ones do.
func listEdges() (bool, bool) {
	pos := listState.list.Position
	return pos.First == 0 && pos.Offset <= 0, !pos.BeforeEnd
}

func listDemo(th *lotusui.Theme, gtx C) D {
	return lotusui.SurfaceCard(th, gtx, func(gtx C) D {
		gtx.Constraints.Max.Y = gtx.Dp(400)
		gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
		return lotusui.ListView(th, &listState.list, gtx, len(listState.rows), func(gtx C, i int) D {
			if listState.rows[i].Clicked(gtx) {
				listState.sel = i
			}
			return lotusui.HoverRow(th, &listState.rows[i], listState.sel == i, func(gtx C) D {
				return lotusui.HStack(lotusui.Space.SM,
					lotusui.LabelBody(th, fmt.Sprintf("Row %d of 10,000", i+1)).Layout,
					lotusui.LabelCaption(th, "only visible rows are laid out").Layout,
				)(gtx)
			})(gtx)
		})
	})
}

// ---- pill ----

func badgeDemo(th *lotusui.Theme, gtx C) D {
	p := th.Palette
	return card(th, gtx,
		section(th, "Badge", lotusui.HStack(lotusui.Space.SM,
			lotusui.Badge(th, "Badge", lotusui.BadgeProps{}),
		)),
		section(th, "Secondary", lotusui.HStack(lotusui.Space.SM,
			lotusui.Badge(th, "Secondary", lotusui.BadgeProps{Variant: lotusui.BadgeSecondary}),
		)),
		section(th, "Destructive", lotusui.HStack(lotusui.Space.SM,
			lotusui.Badge(th, "Destructive", lotusui.BadgeProps{Variant: lotusui.BadgeDestructive}),
		)),
		section(th, "Outline", lotusui.HStack(lotusui.Space.SM,
			lotusui.Badge(th, "Outline", lotusui.BadgeProps{Variant: lotusui.BadgeOutline}),
		)),
		section(th, "Ghost", lotusui.HStack(lotusui.Space.SM,
			lotusui.Badge(th, "Ghost", lotusui.BadgeProps{Variant: lotusui.BadgeGhost}),
		)),
		section(th, "With icon", lotusui.HStack(lotusui.Space.SM,
			lotusui.Badge(th, "Verified", lotusui.BadgeProps{Icon: lotusui.IconAccept}),
			lotusui.Badge(th, "Changes", lotusui.BadgeProps{Variant: lotusui.BadgeSecondary, Icon: lotusui.IconChanges}),
		)),
		section(th, "Spinner — working states", lotusui.HStack(lotusui.Space.SM,
			lotusui.Badge(th, "Deleting", lotusui.BadgeProps{Variant: lotusui.BadgeDestructive,
				Start: lotusui.SpinnerTint(th, 12, p.Danger)}),
			lotusui.Badge(th, "Generating", lotusui.BadgeProps{Variant: lotusui.BadgeSecondary,
				End: lotusui.Spinner(th, 12)}),
		)),
		section(th, "Color — the pastel way, from any scale", lotusui.HStack(lotusui.Space.SM,
			lotusui.Badge(th, "Teal", lotusui.BadgeProps{Color: lotusui.Teal}),
			lotusui.Badge(th, "Purple", lotusui.BadgeProps{Color: lotusui.Purple}),
			lotusui.Badge(th, "Pink", lotusui.BadgeProps{Color: lotusui.Pink}),
		)),
		section(th, "Status — the raw token pairs", lotusui.HStack(lotusui.Space.SM,
			lotusui.Badge(th, "Success", lotusui.BadgeProps{Bg: p.SuccessBg, Fg: p.Success}),
			lotusui.Badge(th, "Warning", lotusui.BadgeProps{Bg: p.WarningBg, Fg: p.Warning}),
			lotusui.Badge(th, "Error", lotusui.BadgeProps{Bg: p.DangerBg, Fg: p.Danger}),
			lotusui.Badge(th, "Info", lotusui.BadgeProps{Bg: p.InfoBg, Fg: p.Info}),
		)),
		section(th, "Sizes — 2XS to 2XL", lotusui.HStack(lotusui.Space.SM,
			lotusui.Badge(th, "2XS", lotusui.BadgeProps{Size: lotusui.Size2XS}),
			lotusui.Badge(th, "XS", lotusui.BadgeProps{Size: lotusui.SizeXS}),
			lotusui.Badge(th, "SM", lotusui.BadgeProps{Size: lotusui.SizeSM}),
			lotusui.Badge(th, "MD", lotusui.BadgeProps{}),
			lotusui.Badge(th, "LG", lotusui.BadgeProps{Size: lotusui.SizeLG}),
			lotusui.Badge(th, "XL", lotusui.BadgeProps{Size: lotusui.SizeXL}),
			lotusui.Badge(th, "2XL", lotusui.BadgeProps{Size: lotusui.Size2XL}),
		)),
	)
}

// ---- card ----

func mediaBlock(th *lotusui.Theme, h unit.Dp) layout.Widget {
	return func(gtx C) D {
		sz := image.Pt(gtx.Constraints.Max.X, gtx.Dp(h))
		gtx.Constraints.Min, gtx.Constraints.Max = sz, sz
		return lotusui.Fill(gtx, th.Palette.BgMuted)
	}
}

var cardBtns struct{ a, b, c widget.Clickable }
var cardTerms widget.List

func termsPara(th *lotusui.Theme, text string) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		l := lotusui.LabelCaption(th, text)
		l.Color = th.Palette.FgMuted
		return l.Layout(gtx)
	}
}

var cardLogin struct {
	email, pw             lotusui.Input
	login, google, signup widget.Clickable
}

func cardDemo(th *lotusui.Theme, gtx C) D {
	basic := func(v lotusui.CardVariant, title string) layout.Widget {
		return lotusui.Card(th, lotusui.CardProps{Variant: v}, lotusui.VStack(th.Space.XS,
			lotusui.LabelCardTitle(th, title).Layout,
			lotusui.LabelMeta(th, "A short description of the thing.").Layout,
		))
	}
	return card(th, gtx,
		section(th, "Card — the login card", func(gtx C) D {
			gtx.Constraints.Max.X = min(gtx.Constraints.Max.X, gtx.Dp(380))
			return lotusui.Card(th, lotusui.CardProps{}, lotusui.VStack(th.Space.MD,
				// Header: title + description, the Sign Up action at the end.
				func(gtx C) D {
					return layout.Flex{}.Layout(gtx,
						layout.Flexed(1, lotusui.VStack(4,
							lotusui.LabelCardTitle(th, "Login to your account").Layout,
							func(gtx C) D {
								l := lotusui.LabelCaption(th, "Enter your email below to login to your account")
								l.Color = th.Palette.FgMuted
								return l.Layout(gtx)
							},
						)),
						layout.Rigid(lotusui.Button(th, &cardLogin.signup, "Sign Up", lotusui.ButtonProps{Variant: lotusui.ButtonLink, Size: lotusui.SizeSM})),
					)
				},
				// Content: the form fields.
				lotusui.Field(th, lotusui.FieldProps{Label: "Email"}, func(gtx C) D {
					return cardLogin.email.LayoutField(th, gtx, "m@example.com")
				}),
				lotusui.Field(th, lotusui.FieldProps{Label: "Password"}, func(gtx C) D {
					cardLogin.pw.Editor.Mask = '•'
					return cardLogin.pw.LayoutField(th, gtx, "")
				}),
				// Footer: full-width actions.
				lotusui.FullWidth(lotusui.Button(th, &cardLogin.login, "Login", lotusui.ButtonProps{})),
				lotusui.FullWidth(lotusui.Button(th, &cardLogin.google, "Login with Google", lotusui.ButtonProps{Variant: lotusui.ButtonOutline})),
			))(gtx)
		}),
		section(th, "Variants — outline (default), elevated, subtle", func(gtx C) D {
			return lotusui.SimpleGrid(th, gtx, []lotusui.CardVariant{lotusui.CardOutline, lotusui.CardElevated, lotusui.CardSubtle},
				lotusui.SimpleGridProps{MinChildWidth: 150, MaxCols: 3, Gap: th.Space.MD},
				func(gtx C, v lotusui.CardVariant) D {
					names := map[lotusui.CardVariant]string{lotusui.CardOutline: "Outline", lotusui.CardElevated: "Elevated", lotusui.CardSubtle: "Subtle"}
					return basic(v, names[v])(gtx)
				})
		}),
		section(th, "Sizes", lotusui.VStack(th.Space.SM,
			lotusui.Card(th, lotusui.CardProps{Size: lotusui.Size2XS}, lotusui.LabelMeta(th, "2XS padding").Layout),
			lotusui.Card(th, lotusui.CardProps{Size: lotusui.SizeXS}, lotusui.LabelMeta(th, "XS padding").Layout),
			lotusui.Card(th, lotusui.CardProps{Size: lotusui.SizeSM}, lotusui.LabelMeta(th, "SM padding").Layout),
			lotusui.Card(th, lotusui.CardProps{}, lotusui.LabelMeta(th, "MD padding").Layout),
			lotusui.Card(th, lotusui.CardProps{Size: lotusui.SizeLG}, lotusui.LabelMeta(th, "LG padding").Layout),
			lotusui.Card(th, lotusui.CardProps{Size: lotusui.SizeXL}, lotusui.LabelMeta(th, "XL padding").Layout),
			lotusui.Card(th, lotusui.CardProps{Size: lotusui.Size2XL}, lotusui.LabelMeta(th, "2XL padding").Layout),
		)),
		section(th, "Image — the event card", func(gtx C) D {
			gtx.Constraints.Max.X = min(gtx.Constraints.Max.X, gtx.Dp(380))
			return lotusui.Card(th, lotusui.CardProps{}, lotusui.VStack(th.Space.MD,
				// The cover: a 16:9 media area (an image in your app),
				// with the Featured badge overlaid at its corner.
				func(gtx C) D {
					w := gtx.Constraints.Max.X
					h := w * 9 / 16
					sz := image.Pt(w, h)
					gtx.Constraints.Min, gtx.Constraints.Max = sz, sz
					lotusui.Fill(gtx, th.Palette.BgInverted)
					off := op.Offset(image.Pt(w-gtx.Dp(86), gtx.Dp(10))).Push(gtx.Ops)
					lotusui.Badge(th, "Featured", lotusui.BadgeProps{Variant: lotusui.BadgeSecondary})(gtx)
					off.Pop()
					return D{Size: sz}
				},
				lotusui.LabelCardTitle(th, "Design systems meetup").Layout,
				func(gtx C) D {
					l := lotusui.LabelCaption(th, "A practical talk on component APIs, accessibility, and shipping faster.")
					l.Color = th.Palette.FgMuted
					return l.Layout(gtx)
				},
				lotusui.FullWidth(lotusui.Button(th, &cardBtns.b, "View Event", lotusui.ButtonProps{})),
			))(gtx)
		}),
		section(th, "Edge to edge — a scrolling well inside the card", func(gtx C) D {
			gtx.Constraints.Max.X = min(gtx.Constraints.Max.X, gtx.Dp(380))
			return lotusui.Card(th, lotusui.CardProps{}, lotusui.VStack(th.Space.SM,
				lotusui.LabelCardTitle(th, "Terms of Service").Layout,
				func(gtx C) D {
					l := lotusui.LabelCaption(th, "Review the terms before accepting the agreement.")
					l.Color = th.Palette.FgMuted
					return l.Layout(gtx)
				},
				func(gtx C) D {
					gtx.Constraints.Max.Y = gtx.Dp(150)
					gtx.Constraints.Min = gtx.Constraints.Max
					lotusui.Fill(gtx, th.Palette.BgSubtle)
					return lotusui.Scrollable(th, &cardTerms, gtx, lotusui.VStack(th.Space.SM,
						termsPara(th, "These terms govern your use of the workspace, including access to shared documents, project files, and collaboration tools."),
						termsPara(th, "You are responsible for the content you upload and for ensuring that your team has the appropriate permissions to view or edit it."),
						termsPara(th, "Workspaces may be suspended if usage violates the acceptable use policy or exceeds the plan limits."),
						termsPara(th, "These terms may change; continued use after an update constitutes acceptance of the new terms."),
					))
				},
				lotusui.FullWidth(lotusui.Button(th, &cardBtns.c, "Accept", lotusui.ButtonProps{})),
			))(gtx)
		}),
		section(th, "Horizontal", func(gtx C) D {
			return lotusui.Card(th, lotusui.CardProps{}, lotusui.HStack(th.Space.MD,
				func(gtx C) D {
					sz := image.Pt(gtx.Dp(72), gtx.Dp(72))
					gtx.Constraints.Min, gtx.Constraints.Max = sz, sz
					return lotusui.Fill(gtx, th.Palette.BgMuted)
				},
				lotusui.VStack(th.Space.XS,
					lotusui.LabelCardTitle(th, "Side by side").Layout,
					lotusui.LabelMeta(th, "Media left, content right — one HStack.").Layout,
				),
			))(gtx)
		}),
	)
}

// ---- grid ----

func gridTile(th *lotusui.Theme, label string) layout.Widget {
	return func(gtx C) D {
		// Intrinsic 48dp height; stretches when the grid assigns taller
		// rows (Min.Y) — never balloons in an unbounded measure pass.
		h := gtx.Dp(48)
		if gtx.Constraints.Min.Y > h {
			h = gtx.Constraints.Min.Y
		}
		sz := gtx.Constraints.Constrain(image.Pt(gtx.Constraints.Max.X, h))
		gtx.Constraints.Min, gtx.Constraints.Max = sz, sz
		lotusui.Fill(gtx, th.Palette.BrandSubtle)
		return layout.Center.Layout(gtx, lotusui.LabelCaption(th, label).Layout)
	}
}

func gridDemo(th *lotusui.Theme, gtx C) D {
	tile := func(l string) layout.Widget { return gridTile(th, l) }
	tall := tile("1")
	return card(th, gtx,
		section(th, "Col span — 4 tracks, one item spanning 2", func(gtx C) D {
			return lotusui.Grid{Columns: 4, Gap: th.Space.SM}.Layout(th, gtx,
				lotusui.Span(2, tall), lotusui.Cell(tile("2")), lotusui.Cell(tile("3")),
				lotusui.Cell(tile("4")), lotusui.Cell(tile("5")), lotusui.Cell(tile("6")), lotusui.Cell(tile("7")),
			)
		}),
		section(th, "Spanning rows and columns", func(gtx C) D {
			return lotusui.Grid{Columns: 4, Gap: th.Space.SM}.Layout(th, gtx,
				lotusui.GridItem{RowSpan: 2, W: tile("rowSpan 2")},
				lotusui.GridItem{ColSpan: 2, W: tall},
				lotusui.Cell(tile("3")),
				lotusui.Cell(tile("4")), lotusui.Cell(tile("5")), lotusui.Cell(tile("6")),
				lotusui.GridItem{ColSpan: 4, W: tile("colSpan 4")},
			)
		}),
		section(th, "Responsive columns — resize; steps at md / lg", func(gtx C) D {
			return lotusui.VStack(th.Space.SM,
				lotusui.LabelCaption(th, "breakpoint: "+th.BreakpointName(gtx)).Layout,
				func(gtx C) D {
					return lotusui.Grid{
						Cols: lotusui.Cols(1).At("md", 2).At("lg", 4),
						Gap:  th.Space.SM,
					}.Layout(th, gtx,
						lotusui.Cell(tile("a")), lotusui.Cell(tile("b")),
						lotusui.Cell(tile("c")), lotusui.Cell(tile("d")),
					)
				},
			)(gtx)
		}),
	)
}

// ---- simplegrid ----

func simpleGridDemo(th *lotusui.Theme, gtx C) D {
	tile := func(l string) layout.Widget { return gridTile(th, l) }
	items := []string{"1", "2", "3", "4", "5", "6"}
	return card(th, gtx,
		section(th, "SimpleGrid — columns derive from the available width", func(gtx C) D {
			return lotusui.SimpleGrid(th, gtx, items, lotusui.SimpleGridProps{
				MinChildWidth: 140, MaxCols: 4, Gap: th.Space.SM,
			}, func(gtx C, l string) D {
				return tile(l)(gtx)
			})
		}),
		section(th, "Equal-height rows — the tallest cell sets its row", func(gtx C) D {
			type entry struct {
				label string
				lines int
			}
			entries := []entry{{"short", 1}, {"taller\ncell", 2}, {"short", 1}}
			return lotusui.SimpleGrid(th, gtx, entries, lotusui.SimpleGridProps{
				MinChildWidth: 140, MaxCols: 3, Gap: th.Space.SM,
			}, func(gtx C, e entry) D {
				h := gtx.Dp(unit.Dp(32 * e.lines))
				if gtx.Constraints.Min.Y > h {
					h = gtx.Constraints.Min.Y
				}
				sz := gtx.Constraints.Constrain(image.Pt(gtx.Constraints.Max.X, h))
				gtx.Constraints.Min, gtx.Constraints.Max = sz, sz
				lotusui.Fill(gtx, th.Palette.BrandSubtle)
				return layout.Center.Layout(gtx, lotusui.LabelCaption(th, e.label).Layout)
			})
		}),
		section(th, "Stepped columns — Cols at sm / lg (resize)", func(gtx C) D {
			return lotusui.VStack(th.Space.SM,
				lotusui.LabelCaption(th, "breakpoint: "+th.BreakpointName(gtx)).Layout,
				func(gtx C) D {
					return lotusui.SimpleGrid(th, gtx, items, lotusui.SimpleGridProps{
						Columns: lotusui.Cols(1).At("sm", 2).At("lg", 4),
						Gap:     th.Space.SM,
					}, func(gtx C, l string) D { return tile(l)(gtx) })
				},
			)(gtx)
		}),
	)
}

// ---- split ----

var split struct {
	s            lotusui.Split
	depth        int
	open, back   widget.Clickable
	deeper, home widget.Clickable
	col, pane    widget.List
	fillFooter   widget.Clickable
}

var vsl struct {
	v                 lotusui.VSlide
	open              bool
	openBtn, closeBtn widget.Clickable
}

func splitState(state string) {
	switch state {
	case "depth1":
		split.depth = 1
	case "depth2":
		split.depth = 2
	default:
		split.depth = 0
	}
}

func splitDemo(th *lotusui.Theme, gtx C) D {
	if split.open.Clicked(gtx) {
		split.depth = 1
	}
	if split.deeper.Clicked(gtx) {
		split.depth = 2
	}
	if split.back.Clicked(gtx) {
		split.depth--
	}
	if split.home.Clicked(gtx) {
		split.depth = 0
	}
	if vsl.openBtn.Clicked(gtx) {
		vsl.open = true
	}
	if vsl.closeBtn.Clicked(gtx) {
		vsl.open = false
	}
	box := func(title, body string, buttons ...layout.Widget) layout.Widget {
		children := []layout.Widget{
			lotusui.LabelCardTitle(th, title).Layout,
			lotusui.LabelMeta(th, body).Layout,
		}
		children = append(children, buttons...)
		return lotusui.SplitBox(th, func(gtx C) D {
			return layout.UniformInset(lotusui.Space.LG).Layout(gtx,
				lotusui.VStack(lotusui.Space.MD, children...))
		})
	}
	return card(th, gtx,
		section(th, "Split — the carousel of boxes", func(gtx C) D {
			return split.s.Layout(gtx, lotusui.Space.MD, split.depth,
				box("Inbox", "Depth 0: this pane owns the full width.",
					lotusui.Button(th, &split.open, "Open a conversation", lotusui.ButtonProps{})),
				box("Conversation", "Depth 1: panes 0 and 1 share the width.",
					lotusui.HStack(lotusui.Space.SM,
						lotusui.Button(th, &split.back, "Back", lotusui.ButtonProps{Variant: lotusui.ButtonSecondary}),
						lotusui.Button(th, &split.deeper, "Open details", lotusui.ButtonProps{}),
					)),
				box("Details", "Depth 2: the strip slid left; pane 0 is off-screen.",
					lotusui.HStack(lotusui.Space.SM,
						lotusui.Button(th, &split.home, "All the way back", lotusui.ButtonProps{Variant: lotusui.ButtonSecondary}),
					)),
			)
		}),
		section(th, "VSlide — the vertical full-screen pivot", func(gtx C) D {
			h := gtx.Dp(240)
			sz := image.Pt(gtx.Constraints.Max.X, h)
			gtx.Constraints.Min, gtx.Constraints.Max = sz, sz
			return vsl.v.Layout(gtx, th, vsl.open,
				box("The list", "The base screen. Expanding an item slides a full screen up over it.",
					lotusui.Button(th, &vsl.openBtn, "Expand an item", lotusui.ButtonProps{})),
				box("The item", "The over-screen: verbatim pixels slid up — nothing reflowed.",
					lotusui.Button(th, &vsl.closeBtn, "Collapse", lotusui.ButtonProps{Variant: lotusui.ButtonSecondary})),
			)
		}),
		section(th, "Column scroll — stacked natural cards", func(gtx C) D {
			h := gtx.Dp(280)
			return lotusui.SplitColumnScroll(th, &split.col, h, lotusui.VStack(lotusui.Space.MD,
				lotusui.SplitBox(th, lotusui.VStack(lotusui.Space.SM,
					lotusui.LabelCardTitle(th, "Card A").Layout,
					lotusui.LabelMeta(th, "Natural height — the column scrolls.").Layout,
				)),
				lotusui.SplitBox(th, lotusui.VStack(lotusui.Space.SM,
					lotusui.LabelCardTitle(th, "Card B").Layout,
					lotusui.LabelMeta(th, "Another natural card in the stack.").Layout,
				)),
				lotusui.SplitBox(th, lotusui.VStack(lotusui.Space.SM,
					lotusui.LabelCardTitle(th, "Card C").Layout,
					lotusui.LabelMeta(th, "Wheel here — the column moves, not each card.").Layout,
				)),
				lotusui.SplitBox(th, lotusui.VStack(lotusui.Space.SM,
					lotusui.LabelCardTitle(th, "Card D").Layout,
					lotusui.LabelMeta(th, "Still more content below the fold.").Layout,
				)),
			))(gtx)
		}),
		section(th, "Pane scroll — hug then scroll inside the card", func(gtx C) D {
			h := gtx.Dp(220)
			lines := make([]layout.Widget, 0, 12)
			for i := 0; i < 12; i++ {
				n := i + 1
				lines = append(lines, func(gtx C) D {
					return lotusui.LabelBody(th, "Line "+strconv.Itoa(n)+" — body scrolls inside one card.").Layout(gtx)
				})
			}
			return lotusui.SplitBoxScroll(th, &split.pane, h, lotusui.VStack(lotusui.Space.SM, lines...))(gtx)
		}),
		section(th, "Fill + pinned footer — Flexed body, Rigid actions", func(gtx C) D {
			h := gtx.Dp(220)
			return lotusui.SplitBoxFillScroll(th, &split.pane, h, func(gtx C) D {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Flexed(1, func(gtx C) D {
						return layout.N.Layout(gtx, lotusui.LabelMeta(th, "Flexed body fills the card — footer stays pinned.").Layout)
					}),
					layout.Rigid(lotusui.FullWidth(lotusui.Button(th, &split.fillFooter, "Pinned action", lotusui.ButtonProps{}))),
				)
			})(gtx)
		}),
	)
}

// ---- icons ----

func iconsDemo(th *lotusui.Theme, gtx C) D {
	names := []string{lotusui.IconAdd, lotusui.IconRemove, lotusui.IconSettings, lotusui.IconFile, lotusui.IconExpand, lotusui.IconChanges}
	mono := []string{lotusui.IconEye, lotusui.IconEyeOff, lotusui.IconEdit, lotusui.IconAccept, lotusui.IconRefuse}
	tile := func(name string, icon layout.Widget) layout.Widget {
		return lotusui.VStack(lotusui.Space.XS, icon, lotusui.LabelCaption(th, name).Layout)
	}
	return card(th, gtx,
		section(th, "Full-color Fluent icons — embedded, no network at build", func(gtx C) D {
			tiles := make([]tileItem, len(names))
			for i, n := range names {
				tiles[i] = tileItem{n, tile(n, lotusui.SVGIcon(n, 32, color.NRGBA{}))}
			}
			return tileGrid(th, gtx, tiles)
		}),
		section(th, "Mono icons — tinted via currentColor", func(gtx C) D {
			tiles := make([]tileItem, len(mono))
			for i, n := range mono {
				tiles[i] = tileItem{n, tile(n, lotusui.SVGIcon(n, 24, th.Palette.FgSubtle))}
			}
			return tileGrid(th, gtx, tiles)
		}),
	)
}

type tileItem struct {
	name string
	w    layout.Widget
}

func tileGrid(th *lotusui.Theme, gtx C, items []tileItem) D {
	return lotusui.SimpleGrid(th, gtx, items, lotusui.SimpleGridProps{
		MinChildWidth: 90, MaxCols: 6, Gap: lotusui.Space.SM,
	}, func(gtx C, it tileItem) D {
		return it.w(gtx)
	})
}

var _ = unit.Dp(0)

// dialogLorem fills the scrollable dialog demos.
const dialogLorem = "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur."
