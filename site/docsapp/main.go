// Command docsapp is the lotusui documentation site as a Gio app —
// chrome and Previews are real widgets (not an HTML shell).
package main

import (
	"image"
	"image/color"
	"log"
	"os"
	"strings"

	"gioui.org/app"
	"gioui.org/font"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	lotusui "github.com/ikaito-com/lotusui"
	"github.com/ikaito-com/lotusui/site/docspages"
	"github.com/ikaito-com/lotusui/site/live"
	"github.com/ikaito-com/lotusui/site/looks"
	"github.com/ikaito-com/lotusui/site/palettes"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

const (
	githubRepo = "https://github.com/ikaito-com/lotusui"
	siteTag    = "A Go design system for desktop and mobile — one codebase, native apps. Web when you want it."
)

func main() {
	go func() {
		w := new(app.Window)
		w.Option(app.Title("lotusui"), app.Size(1280, 860))
		if err := loop(w); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

var th = live.NewTheme() // lotus look × lavender — same as the static site

func loop(w *app.Window) error {
	th.UpgradeShaperAsync(w.Invalidate)
	onRouteChange(w.Invalidate)

	live.OnThemeChange = func() {
		nt := live.NewTheme()
		nt.UpgradeShaperAsync(w.Invalidate)
		th = nt
	}
	live.Embed = true
	initPalette(func(slug string) {
		live.SetPalette(slug)
		w.Invalidate()
	})
	initLook(func(slug string) {
		live.SetLook(slug)
		w.Invalidate()
	})

	ui := newDocsUI()
	ui.invalidate = w.Invalidate
	var ops op.Ops
	for {
		switch e := w.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			th.ApplyPendingShaper()
			ui.syncRoute()
			ui.Layout(th, gtx)
			e.Frame(gtx.Ops)
		}
	}
}

type docsUI struct {
	groups    []docspages.Group
	navBtns   []*widget.Clickable
	navKeys   []string
	page      string
	body      lotusui.ScrollArea
	navScroll lotusui.ScrollArea

	wordmarkBtn                    widget.Clickable
	palBtn, lookBtn, verBtn, ghBtn widget.Clickable
	palOpen, lookOpen, verOpen     bool
	palBtns, lookBtns, verBtns     []*widget.Clickable
	versions                       []versionEntry
	version                        string

	secUI            []sectionUI
	prevBtn, nextBtn widget.Clickable
	ctaBtn           widget.Clickable
	homeCards        map[string]*widget.Clickable
	heroOps          []paint.ImageOp

	tocHeadings []string
	tocYs       []int // content Y of each TOC target (scroll coords)
	tocBtns     []*widget.Clickable

	invalidate func()
}

type versionEntry struct {
	Version string `json:"version"`
	Path    string `json:"path"`
}

func newDocsUI() *docsUI {
	docsGroups = loadDocsGroups()
	ui := &docsUI{
		groups:    docsGroups,
		homeCards: map[string]*widget.Clickable{},
		// heroOps loaded lazily on first home paint — Catmull/decode
		// of three showcase PNGs was blocking the window open.
	}
	// Sidebar has no Home link — wordmark navigates home (static site).
	for _, g := range ui.groups {
		for _, p := range g.Pages {
			ui.navBtns = append(ui.navBtns, new(widget.Clickable))
			ui.navKeys = append(ui.navKeys, p.Slug)
		}
	}
	for range palettes.Presets {
		ui.palBtns = append(ui.palBtns, new(widget.Clickable))
	}
	for range looks.Presets {
		ui.lookBtns = append(ui.lookBtns, new(widget.Clickable))
	}
	ui.versions = loadVersions()
	if len(ui.versions) > 0 {
		ui.version = ui.versions[0].Version
	}
	for range ui.versions {
		ui.verBtns = append(ui.verBtns, new(widget.Clickable))
	}
	ui.page = normalizeRoute(currentRoute())
	return ui
}

func normalizeRoute(r string) string {
	for len(r) > 0 && (r[0] == '#' || r[0] == '/') {
		r = r[1:]
	}
	if i := indexByte(r, '/'); i >= 0 {
		r = r[:i]
	}
	if i := indexByte(r, '&'); i >= 0 {
		r = r[:i]
	}
	if r == "home" {
		return ""
	}
	return r
}

func indexByte(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

func (ui *docsUI) syncRoute() {
	r := normalizeRoute(currentRoute())
	if r != ui.page {
		ui.page = r
		ui.body.Reset()
		ui.secUI = nil
		live.ResetEmbed()
	}
}

func (ui *docsUI) Layout(th *lotusui.Theme, gtx C) D {
	// Paint the window Max — Fill uses Min, which is zero on some
	// first frames and reads as a blank app.
	if sz := gtx.Constraints.Max; sz.X > 0 && sz.Y > 0 {
		cl := clip.Rect{Max: sz}.Push(gtx.Ops)
		paint.Fill(gtx.Ops, th.Palette.Bg)
		cl.Pop()
	}

	for i, b := range ui.navBtns {
		if b.Clicked(gtx) {
			ui.navigate(ui.navKeys[i])
		}
	}
	if ui.palBtn.Clicked(gtx) {
		ui.palOpen = !ui.palOpen
		ui.lookOpen, ui.verOpen = false, false
	}
	if ui.lookBtn.Clicked(gtx) {
		ui.lookOpen = !ui.lookOpen
		ui.palOpen, ui.verOpen = false, false
	}
	if ui.verBtn.Clicked(gtx) {
		ui.verOpen = !ui.verOpen
		ui.palOpen, ui.lookOpen = false, false
	}
	if ui.ghBtn.Clicked(gtx) {
		_ = lotusui.OpenURL(githubRepo)
	}
	if ui.wordmarkBtn.Clicked(gtx) {
		ui.navigate("")
	}
	for i, b := range ui.palBtns {
		if b.Clicked(gtx) {
			live.SetPalette(palettes.Presets[i].Slug)
			ui.palOpen = false
		}
	}
	for i, b := range ui.lookBtns {
		if b.Clicked(gtx) {
			live.SetLook(looks.Presets[i].Slug)
			ui.lookOpen = false
		}
	}
	for i, b := range ui.verBtns {
		if b.Clicked(gtx) {
			ui.version = ui.versions[i].Version
			ui.verOpen = false
		}
	}

	dims := layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx C) D { return ui.topbar(th, gtx) }),
		layout.Flexed(1, func(gtx C) D { return ui.shell(th, gtx) }),
	)
	live.PaintOverlay(th, gtx)
	return dims
}

func (ui *docsUI) navigate(slug string) {
	ui.page = slug
	ui.body.Reset()
	ui.secUI = nil
	live.ResetEmbed()
	setRoute(slug)
}

// shell: max-width ~1240 centered — sidebar | content | toc (static .shell).
func (ui *docsUI) shell(th *lotusui.Theme, gtx C) D {
	return layout.Center.Layout(gtx, func(gtx C) D {
		max := gtx.Dp(unit.Dp(1240))
		if gtx.Constraints.Max.X > max {
			gtx.Constraints.Max.X = max
		}
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
		return layout.Inset{
			Top: unit.Dp(28), Bottom: unit.Dp(64),
			Left: unit.Dp(24), Right: unit.Dp(24),
		}.Layout(gtx, func(gtx C) D {
			// .shell gap: 40px between all columns. TOC only on article pages.
			children := []layout.FlexChild{
				layout.Rigid(func(gtx C) D { return ui.sidebar(th, gtx) }),
				layout.Rigid(layout.Spacer{Width: unit.Dp(40)}.Layout),
				layout.Flexed(1, func(gtx C) D { return ui.bodyPane(th, gtx) }),
			}
			if ui.page != "" && ui.page != "home" {
				children = append(children,
					layout.Rigid(layout.Spacer{Width: unit.Dp(40)}.Layout),
					layout.Rigid(func(gtx C) D { return ui.toc(th, gtx) }),
				)
			}
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx, children...)
		})
	})
}

func (ui *docsUI) topbar(th *lotusui.Theme, gtx C) D {
	verLabel := "docs"
	if ui.version != "" {
		verLabel = ui.version
	}
	dims := layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return layout.Background{}.Layout(gtx,
				func(gtx C) D {
					// rgba(255,255,255,0.75) — frosted panel
					c := th.Palette.BgPanel
					c.A = 191
					return lotusui.Fill(gtx, c)
				},
				func(gtx C) D {
					return layout.Inset{
						Top: unit.Dp(12), Bottom: unit.Dp(12),
						Left: unit.Dp(20), Right: unit.Dp(20),
					}.Layout(gtx, func(gtx C) D {
						return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
							layout.Rigid(func(gtx C) D {
								return ui.wordmarkBtn.Layout(gtx, func(gtx C) D {
									pointer.CursorPointer.Add(gtx.Ops)
									l := material.Label(th.Material, unit.Sp(17), "lotusui")
									l.Font.Weight = 700
									l.Color = th.Palette.Fg
									if ui.wordmarkBtn.Hovered() {
										l.Color = th.Palette.BrandFg
									}
									return l.Layout(gtx)
								})
							}),
							layout.Rigid(layout.Spacer{Width: unit.Dp(12)}.Layout),
							layout.Flexed(1, func(gtx C) D {
								l := material.Label(th.Material, unit.Sp(13), siteTag)
								l.Color = th.Palette.FgSubtle // --text-sec
								l.MaxLines = 1
								return l.Layout(gtx)
							}),
							// .topbar-actions gap: 10
							layout.Rigid(func(gtx C) D { return ui.chipBtn(th, gtx, &ui.verBtn, verLabel, true) }),
							layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
							layout.Rigid(func(gtx C) D { return ui.chipBtn(th, gtx, &ui.lookBtn, "Aa", false) }),
							layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
							layout.Rigid(func(gtx C) D { return ui.paletteBtn(th, gtx) }),
							layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
							layout.Rigid(func(gtx C) D { return ui.githubBtn(th, gtx) }),
						)
					})
				},
			)
		}),
		layout.Rigid(func(gtx C) D {
			sz := image.Pt(gtx.Constraints.Max.X, gtx.Dp(1))
			cl := clip.Rect{Max: sz}.Push(gtx.Ops)
			paint.Fill(gtx.Ops, th.Palette.BorderSubtle)
			cl.Pop()
			return D{Size: sz}
		}),
	)

	if ui.palOpen || ui.lookOpen || ui.verOpen {
		menuW := gtx.Dp(unit.Dp(180))
		lotusui.Floating(gtx, func(gtx C) D {
			offX := gtx.Constraints.Max.X - menuW - gtx.Dp(20)
			if offX < 0 {
				offX = 0
			}
			defer op.Offset(image.Pt(offX, dims.Size.Y)).Push(gtx.Ops).Pop()
			gtx.Constraints.Min = image.Point{}
			gtx.Constraints.Max.X = menuW
			return ui.palMenu(th, gtx)
		})
	}
	return dims
}

// chipBtn paints .palbtn.verbtn / .lookbtn — 22px tall, card fill, r=6.
func (ui *docsUI) chipBtn(th *lotusui.Theme, gtx C, btn *widget.Clickable, label string, mono bool) D {
	return btn.Layout(gtx, func(gtx C) D {
		pointer.CursorPointer.Add(gtx.Ops)
		padX := unit.Dp(7)
		size := unit.Sp(11)
		weight := font.Weight(600)
		if !mono {
			padX = unit.Dp(5)
			size = unit.Sp(12)
			weight = 700
		}
		return layout.Background{}.Layout(gtx,
			func(gtx C) D {
				rr := clip.UniformRRect(image.Rectangle{Max: gtx.Constraints.Min}, gtx.Dp(6)).Push(gtx.Ops)
				paint.Fill(gtx.Ops, th.Palette.BgPanel)
				rr.Pop()
				widget.Border{
					Color: th.Palette.Border, Width: unit.Dp(1), CornerRadius: unit.Dp(6),
				}.Layout(gtx, func(gtx C) D { return D{Size: gtx.Constraints.Min} })
				return D{Size: gtx.Constraints.Min}
			},
			func(gtx C) D {
				h := gtx.Dp(unit.Dp(22))
				return layout.Inset{Left: padX, Right: padX}.Layout(gtx, func(gtx C) D {
					gtx.Constraints.Min.Y = h
					return layout.Center.Layout(gtx, func(gtx C) D {
						l := material.Label(th.Material, size, label)
						l.Font.Weight = weight
						l.Color = th.Palette.FgSubtle
						if btn.Hovered() {
							l.Color = th.Palette.BrandFg
						}
						return l.Layout(gtx)
					})
				})
			},
		)
	})
}

func (ui *docsUI) githubBtn(th *lotusui.Theme, gtx C) D {
	return ui.ghBtn.Layout(gtx, func(gtx C) D {
		pointer.CursorPointer.Add(gtx.Ops)
		sz := gtx.Dp(unit.Dp(22))
		col := th.Palette.FgSubtle
		if ui.ghBtn.Hovered() {
			col = th.Palette.BrandFg
		}
		return layout.Stack{Alignment: layout.Center}.Layout(gtx,
			layout.Stacked(func(gtx C) D { return D{Size: image.Pt(sz, sz)} }),
			layout.Stacked(lotusui.SVGIcon(lotusui.IconGithub, unit.Dp(16), col)),
		)
	})
}

func (ui *docsUI) palMenu(th *lotusui.Theme, gtx C) D {
	var rows []layout.Widget
	switch {
	case ui.palOpen:
		for i, p := range palettes.Presets {
			i, p := i, p
			rows = append(rows, menuRow(th, ui.palBtns[i], p.Name, func(gtx C) D {
				return miniSwatch(th, gtx, palettes.Presets[i].Palette)
			}))
		}
	case ui.lookOpen:
		for i, l := range looks.Presets {
			i, l := i, l
			rows = append(rows, menuRow(th, ui.lookBtns[i], l.Name, func(gtx C) D {
				lb := material.Label(th.Material, unit.Sp(12), "Aa")
				lb.Font.Weight = 700
				lb.Color = th.Palette.BrandFg
				return lb.Layout(gtx)
			}))
		}
	case ui.verOpen:
		for i, v := range ui.versions {
			i, v := i, v
			rows = append(rows, menuRow(th, ui.verBtns[i], v.Version, nil))
		}
		if len(rows) == 0 {
			rows = append(rows, lotusui.LabelMeta(th, "no versions.json").Layout)
		}
	}
	return layout.Background{}.Layout(gtx,
		func(gtx C) D {
			rr := clip.UniformRRect(image.Rectangle{Max: gtx.Constraints.Min}, gtx.Dp(th.Radius.MD)).Push(gtx.Ops)
			paint.Fill(gtx.Ops, th.Palette.BgPanel)
			rr.Pop()
			widget.Border{
				Color: th.Palette.Border, Width: unit.Dp(1), CornerRadius: th.Radius.MD,
			}.Layout(gtx, func(gtx C) D { return D{Size: gtx.Constraints.Min} })
			return D{Size: gtx.Constraints.Min}
		},
		func(gtx C) D {
			return layout.UniformInset(unit.Dp(6)).Layout(gtx, lotusui.VStack(unit.Dp(0), rows...))
		},
	)
}

func menuRow(th *lotusui.Theme, btn *widget.Clickable, label string, leading layout.Widget) layout.Widget {
	return func(gtx C) D {
		return btn.Layout(gtx, func(gtx C) D {
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
			bg := th.Palette.BgPanel
			if btn.Hovered() {
				bg = th.Palette.BgSubtle
			}
			return layout.Background{}.Layout(gtx,
				func(gtx C) D {
					rr := clip.UniformRRect(image.Rectangle{Max: gtx.Constraints.Min}, gtx.Dp(8)).Push(gtx.Ops)
					paint.Fill(gtx.Ops, bg)
					rr.Pop()
					return D{Size: gtx.Constraints.Min}
				},
				func(gtx C) D {
					return layout.Inset{
						Top: unit.Dp(6), Bottom: unit.Dp(6),
						Left: unit.Dp(8), Right: unit.Dp(8),
					}.Layout(gtx, func(gtx C) D {
						return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
							layout.Rigid(func(gtx C) D {
								if leading == nil {
									return D{}
								}
								return layout.Inset{Right: unit.Dp(8)}.Layout(gtx, leading)
							}),
							layout.Flexed(1, func(gtx C) D {
								l := material.Label(th.Material, unit.Sp(13), label)
								l.Font.Weight = 500
								l.Color = th.Palette.FgMuted
								return l.Layout(gtx)
							}),
						)
					})
				},
			)
		})
	}
}

func miniSwatch(th *lotusui.Theme, gtx C, p lotusui.Palette) D {
	sz := gtx.Dp(unit.Dp(16))
	half := sz / 2
	gap := 1
	colors := []color.NRGBA{p.BrandSolid, p.BrandFg, p.BrandSubtle, p.BrandContrast}
	pts := []image.Rectangle{
		{Max: image.Pt(half-gap, half-gap)},
		{Min: image.Pt(half+gap, 0), Max: image.Pt(sz, half-gap)},
		{Min: image.Pt(0, half+gap), Max: image.Pt(half-gap, sz)},
		{Min: image.Pt(half+gap, half+gap), Max: image.Pt(sz, sz)},
	}
	for i, r := range pts {
		c := clip.UniformRRect(r, gtx.Dp(2)).Push(gtx.Ops)
		paint.Fill(gtx.Ops, colors[i])
		c.Pop()
	}
	_ = th
	return D{Size: image.Pt(sz, sz)}
}

func (ui *docsUI) paletteBtn(th *lotusui.Theme, gtx C) D {
	return ui.palBtn.Layout(gtx, func(gtx C) D {
		pointer.CursorPointer.Add(gtx.Ops)
		// .palbtn: 22×22, pad 2, gap 1, cell r=2, outer r=6
		sz := gtx.Dp(unit.Dp(22))
		pad := gtx.Dp(unit.Dp(2))
		gap := 1
		inner := sz - 2*pad
		half := (inner - gap) / 2
		colors := []color.NRGBA{
			th.Palette.BrandSolid, th.Palette.BrandFg,
			th.Palette.BrandSubtle, th.Palette.BrandContrast,
		}
		base := image.Pt(pad, pad)
		pts := []image.Rectangle{
			{Min: base, Max: base.Add(image.Pt(half, half))},
			{Min: base.Add(image.Pt(half+gap, 0)), Max: base.Add(image.Pt(inner, half))},
			{Min: base.Add(image.Pt(0, half+gap)), Max: base.Add(image.Pt(half, inner))},
			{Min: base.Add(image.Pt(half+gap, half+gap)), Max: base.Add(image.Pt(inner, inner))},
		}
		rr := clip.UniformRRect(image.Rectangle{Max: image.Pt(sz, sz)}, gtx.Dp(6)).Push(gtx.Ops)
		paint.Fill(gtx.Ops, th.Palette.BgPanel)
		for i, r := range pts {
			c := clip.UniformRRect(r, gtx.Dp(2)).Push(gtx.Ops)
			paint.Fill(gtx.Ops, colors[i])
			c.Pop()
		}
		rr.Pop()
		border := th.Palette.Border
		if ui.palBtn.Hovered() {
			border = th.Palette.BrandFg
		}
		widget.Border{
			Color: border, Width: unit.Dp(1), CornerRadius: unit.Dp(6),
		}.Layout(gtx, func(gtx C) D { return D{Size: image.Pt(sz, sz)} })
		return D{Size: image.Pt(sz, sz)}
	})
}

func (ui *docsUI) sidebar(th *lotusui.Theme, gtx C) D {
	w := gtx.Dp(unit.Dp(200))
	gtx.Constraints.Min.X = w
	gtx.Constraints.Max.X = w
	return layout.Inset{Top: unit.Dp(4), Left: unit.Dp(2), Right: unit.Dp(2)}.Layout(gtx, func(gtx C) D {
		var widgets []layout.Widget
		bi := 0
		for gi, g := range ui.groups {
			title := strings.ToUpper(g.Title)
			top := unit.Dp(0)
			if gi > 0 {
				top = unit.Dp(22)
			}
			widgets = append(widgets, func(gtx C) D {
				return layout.Inset{Top: top, Bottom: unit.Dp(6), Left: unit.Dp(10)}.Layout(gtx, func(gtx C) D {
					l := material.Label(th.Material, unit.Sp(11), title)
					l.Font.Weight = 700
					l.Color = th.Palette.Fg
					return l.Layout(gtx)
				})
			})
			for _, it := range g.Pages {
				btn := ui.navBtns[bi]
				active := ui.page == it.Slug
				label := it.Title
				bi++
				widgets = append(widgets, navLink(th, btn, label, active))
			}
		}
		return ui.navScroll.LayoutWith(th, gtx, lotusui.ScrollAreaProps{NoShadowRoom: true},
			lotusui.VStack(unit.Dp(1), widgets...))
	})
}

func navLink(th *lotusui.Theme, btn *widget.Clickable, label string, active bool) layout.Widget {
	return func(gtx C) D {
		return btn.Layout(gtx, func(gtx C) D {
			pointer.CursorPointer.Add(gtx.Ops)
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
			return layout.Background{}.Layout(gtx,
				func(gtx C) D {
					if active {
						rr := clip.UniformRRect(image.Rectangle{Max: gtx.Constraints.Min}, gtx.Dp(8)).Push(gtx.Ops)
						paint.Fill(gtx.Ops, th.Palette.BrandSolid)
						rr.Pop()
					} else if btn.Hovered() {
						rr := clip.UniformRRect(image.Rectangle{Max: gtx.Constraints.Min}, gtx.Dp(8)).Push(gtx.Ops)
						paint.Fill(gtx.Ops, th.Palette.BgSubtle)
						rr.Pop()
					}
					return D{Size: gtx.Constraints.Min}
				},
				func(gtx C) D {
					return layout.Inset{
						Top: unit.Dp(4), Bottom: unit.Dp(4),
						Left: unit.Dp(10), Right: unit.Dp(10),
					}.Layout(gtx, func(gtx C) D {
						l := material.Label(th.Material, unit.Sp(14), label)
						if active {
							l.Font.Weight = 600
							l.Color = th.Palette.BrandContrast
						} else if btn.Hovered() {
							l.Color = th.Palette.Fg
						} else {
							l.Color = th.Palette.FgSubtle // --text-sec
						}
						return l.Layout(gtx)
					})
				},
			)
		})
	}
}

func (ui *docsUI) toc(th *lotusui.Theme, gtx C) D {
	w := gtx.Dp(unit.Dp(170))
	gtx.Constraints.Min.X = w
	gtx.Constraints.Max.X = w
	if len(ui.tocHeadings) == 0 {
		return D{Size: image.Pt(w, 0)}
	}
	for len(ui.tocBtns) < len(ui.tocHeadings) {
		ui.tocBtns = append(ui.tocBtns, new(widget.Clickable))
	}
	for i, b := range ui.tocBtns {
		if i >= len(ui.tocHeadings) {
			break
		}
		if b.Clicked(gtx) {
			y := 0
			if i < len(ui.tocYs) {
				y = ui.tocYs[i]
			}
			// Match static scroll-margin-top ~80px under sticky chrome.
			ui.body.ScrollTo(y - gtx.Dp(unit.Dp(12)))
		}
	}
	return layout.Inset{Top: unit.Dp(42)}.Layout(gtx, func(gtx C) D {
		var rows []layout.Widget
		rows = append(rows, func(gtx C) D {
			l := material.Label(th.Material, unit.Sp(11), "ON THIS PAGE")
			l.Font.Weight = 700
			l.Color = th.Palette.Fg
			return layout.Inset{Bottom: unit.Dp(8)}.Layout(gtx, l.Layout)
		})
		for i, h := range ui.tocHeadings {
			i, h := i, h
			rows = append(rows, func(gtx C) D {
				btn := ui.tocBtns[i]
				return btn.Layout(gtx, func(gtx C) D {
					pointer.CursorPointer.Add(gtx.Ops)
					return layout.Inset{Top: unit.Dp(3), Bottom: unit.Dp(3)}.Layout(gtx, func(gtx C) D {
						m := op.Record(gtx.Ops)
						d := layout.Inset{Left: unit.Dp(10)}.Layout(gtx, func(gtx C) D {
							l := material.Label(th.Material, unit.Sp(13), h)
							l.Color = th.Palette.FgSubtle
							if btn.Hovered() {
								l.Color = th.Palette.BrandFg
							}
							l.MaxLines = 2
							return l.Layout(gtx)
						})
						call := m.Stop()
						barCol := th.Palette.BorderSubtle
						if btn.Hovered() {
							barCol = th.Palette.BrandFg
						}
						bar := clip.Rect{Max: image.Pt(gtx.Dp(2), d.Size.Y)}.Push(gtx.Ops)
						paint.Fill(gtx.Ops, barCol)
						bar.Pop()
						call.Add(gtx.Ops)
						return d
					})
				})
			})
		}
		return lotusui.VStack(unit.Dp(0), rows...)(gtx)
	})
}

func (ui *docsUI) bodyPane(th *lotusui.Theme, gtx C) D {
	max := gtx.Dp(unit.Dp(720))
	if ui.page == "performance" {
		max = gtx.Dp(unit.Dp(820))
	}
	if gtx.Constraints.Max.X > max {
		gtx.Constraints.Max.X = max
	}
	if gtx.Constraints.Min.X > gtx.Constraints.Max.X {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
	}
	return ui.body.Layout(th, gtx, ui.pageContent(th))
}
