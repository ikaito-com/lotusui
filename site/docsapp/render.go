package main

import (
	"fmt"
	"image"
	"image/color"
	"strings"

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
)

type sectionUI struct {
	box     lotusui.Example
	copyBtn widget.Clickable
}

func (ui *docsUI) ensureSections(n int) {
	for len(ui.secUI) < n {
		ui.secUI = append(ui.secUI, sectionUI{})
	}
}

// ch66 approximates CSS max-width: 66ch at 15px body.
func constrainCh66(gtx C) C {
	max := gtx.Dp(unit.Dp(560))
	if gtx.Constraints.Max.X > max {
		gtx.Constraints.Max.X = max
	}
	return gtx
}

func (ui *docsUI) renderDocPage(th *lotusui.Theme, p *docspages.Page, group string, prev, next *docspages.Page) layout.Widget {
	ui.ensureSections(len(p.Sections))
	ui.tocHeadings = ui.tocHeadings[:0]
	for _, s := range p.Sections {
		ui.tocHeadings = append(ui.tocHeadings, s.Heading)
	}
	if len(p.Props) > 0 {
		ui.tocHeadings = append(ui.tocHeadings, "API")
	}
	ui.tocYs = make([]int, len(ui.tocHeadings))

	type part struct {
		w   layout.Widget
		toc int // TOC index, or -1
	}

	return func(gtx C) D {
		var parts []part

		parts = append(parts, part{toc: -1, w: func(gtx C) D {
			return layout.Inset{Bottom: unit.Dp(10)}.Layout(gtx, func(gtx C) D {
				return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						l := material.Label(th.Material, unit.Sp(13), group)
						l.Color = th.Palette.FgDisabled
						return l.Layout(gtx)
					}),
					layout.Rigid(func(gtx C) D {
						return layout.Inset{Left: unit.Dp(4), Right: unit.Dp(4)}.Layout(gtx, func(gtx C) D {
							l := material.Label(th.Material, unit.Sp(13), "/")
							l.Color = th.Palette.FgDisabled
							return l.Layout(gtx)
						})
					}),
					layout.Rigid(func(gtx C) D {
						l := material.Label(th.Material, unit.Sp(13), p.Title)
						l.Color = th.Palette.FgDisabled
						return l.Layout(gtx)
					}),
				)
			})
		}})

		parts = append(parts, part{toc: -1, w: func(gtx C) D {
			return layout.Inset{Bottom: unit.Dp(6)}.Layout(gtx, func(gtx C) D {
				l := material.Label(th.Material, unit.Sp(30), p.Title)
				l.Font.Weight = 700
				l.Color = th.Palette.Fg
				return l.Layout(gtx)
			})
		}})

		if p.Kicker != "" {
			parts = append(parts, part{toc: -1, w: func(gtx C) D {
				return layout.Inset{Bottom: unit.Dp(18)}.Layout(gtx, func(gtx C) D {
					l := material.Label(th.Material, unit.Sp(16), p.Kicker)
					l.Color = th.Palette.FgSubtle
					return l.Layout(gtx)
				})
			}})
		}

		if len(p.Platforms) > 0 {
			plats := p.Platforms
			parts = append(parts, part{toc: -1, w: func(gtx C) D {
				return layout.Inset{Bottom: unit.Dp(14)}.Layout(gtx, platformChips(th, plats))
			}})
		}

		for _, b := range htmlBlocks(p.Intro) {
			b := b
			parts = append(parts, part{toc: -1, w: renderHTMLBlock(th, b)})
		}

		for i, sec := range p.Sections {
			i, sec := i, sec
			parts = append(parts, part{toc: i, w: ui.renderSection(th, i, sec)})
		}
		if len(p.Props) > 0 {
			tocAPI := len(p.Sections)
			parts = append(parts, part{toc: tocAPI, w: ui.renderProps(th, p.Props)})
		}
		parts = append(parts, part{toc: -1, w: layout.Spacer{Height: unit.Dp(52)}.Layout})
		parts = append(parts, part{toc: -1, w: ui.renderPrevNext(th, prev, next)})

		// Lay out manually so TOC can scroll to section Y offsets.
		y := 0
		maxW := gtx.Constraints.Max.X
		for _, it := range parts {
			if it.toc >= 0 && it.toc < len(ui.tocYs) {
				ui.tocYs[it.toc] = y
			}
			trans := op.Offset(image.Pt(0, y)).Push(gtx.Ops)
			cgtx := gtx
			cgtx.Constraints.Min = image.Point{}
			d := it.w(cgtx)
			trans.Pop()
			y += d.Size.Y
			if d.Size.X > maxW {
				maxW = d.Size.X
			}
		}
		return D{Size: image.Pt(gtx.Constraints.Max.X, y)}
	}
}

func (ui *docsUI) renderSection(th *lotusui.Theme, i int, sec docspages.Section) layout.Widget {
	return func(gtx C) D {
		var parts []layout.Widget
		// h2 — 20px, margin: 44px 0 10px
		parts = append(parts, func(gtx C) D {
			return layout.Inset{Top: unit.Dp(44), Bottom: unit.Dp(10)}.Layout(gtx, func(gtx C) D {
				title := func(gtx C) D {
					l := material.Label(th.Material, unit.Sp(20), sec.Heading)
					l.Font.Weight = 700
					l.Color = th.Palette.Fg
					return l.Layout(gtx)
				}
				if len(sec.Platforms) == 0 {
					return title(gtx)
				}
				chips := make([]layout.Widget, 0, 1+len(sec.Platforms))
				chips = append(chips, title)
				for _, p := range sec.Platforms {
					p := p
					chips = append(chips, platformChip(th, p))
				}
				return lotusui.HStack(unit.Dp(4), chips...)(gtx)
			})
		})
		for _, b := range htmlBlocks(sec.Prose) {
			b := b
			parts = append(parts, renderHTMLBlock(th, b))
		}
		switch {
		case sec.Demo != "" && sec.Snippet != "":
			parts = append(parts, func(gtx C) D {
				return layout.Inset{Top: unit.Dp(14)}.Layout(gtx, func(gtx C) D {
					return ui.secUI[i].box.Layout(th, gtx, lotusui.ExampleProps{
						Preview: previewPanel(th, sec.Demo),
						Code:    codeBlock(th, &ui.secUI[i].copyBtn, sec.Snippet, sec.Lang, true),
					})
				})
			})
		case sec.Demo != "":
			parts = append(parts, func(gtx C) D {
				return layout.Inset{Top: unit.Dp(14)}.Layout(gtx, func(gtx C) D {
					return ui.secUI[i].box.Layout(th, gtx, lotusui.ExampleProps{
						Preview: previewPanel(th, sec.Demo),
					})
				})
			})
		case sec.Snippet != "":
			parts = append(parts, func(gtx C) D {
				return layout.Inset{Top: unit.Dp(14)}.Layout(gtx,
					codeBlock(th, &ui.secUI[i].copyBtn, sec.Snippet, sec.Lang, false),
				)
			})
		}
		return lotusui.VStack(unit.Dp(0), parts...)(gtx)
	}
}

func previewPanel(th *lotusui.Theme, demo string) layout.Widget {
	return func(gtx C) D {
		// Gallery iframe content uses Space.LG (=24) inset.
		return layout.UniformInset(lotusui.Space.LG).Layout(gtx, func(gtx C) D {
			return live.RenderSection(th, gtx, demo)
		})
	}
}

func codeBlock(th *lotusui.Theme, copyBtn *widget.Clickable, snippet, lang string, nested bool) layout.Widget {
	// Defer chroma until this widget actually lays out (Code tab).
	// Building ExampleProps used to tokenize every snippet every frame
	// even while Preview was showing — freezing heavy pages at open.
	return func(gtx C) D {
		return lotusui.CodeBlock(th, lotusui.CodeBlockProps{
			Lang: lang, Plain: snippet, Lines: highlightLines(th, snippet, lang),
			Copy: copyBtn, Nested: nested,
		})(gtx)
	}
}

func platformChips(th *lotusui.Theme, names []string) layout.Widget {
	ws := make([]layout.Widget, len(names))
	for i, n := range names {
		n := n
		ws[i] = platformChip(th, n)
	}
	return lotusui.HStack(unit.Dp(6), ws...)
}

func platformChip(th *lotusui.Theme, name string) layout.Widget {
	return func(gtx C) D {
		dot := chipDotColor(th, name)
		return layout.Background{}.Layout(gtx,
			func(gtx C) D {
				rr := clip.UniformRRect(image.Rectangle{Max: gtx.Constraints.Min}, gtx.Constraints.Min.Y/2).Push(gtx.Ops)
				paint.Fill(gtx.Ops, th.Palette.BgSubtle)
				rr.Pop()
				widget.Border{
					Color: th.Palette.BorderSubtle, Width: unit.Dp(1),
					CornerRadius: unit.Dp(999),
				}.Layout(gtx, func(gtx C) D { return D{Size: gtx.Constraints.Min} })
				return D{Size: gtx.Constraints.Min}
			},
			func(gtx C) D {
				return layout.Inset{
					Top: unit.Dp(2), Bottom: unit.Dp(2),
					Left: unit.Dp(9), Right: unit.Dp(9),
				}.Layout(gtx, func(gtx C) D {
					return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
						layout.Rigid(func(gtx C) D {
							sz := gtx.Dp(unit.Dp(6))
							rr := clip.UniformRRect(image.Rectangle{Max: image.Pt(sz, sz)}, sz/2).Push(gtx.Ops)
							paint.Fill(gtx.Ops, dot)
							rr.Pop()
							return D{Size: image.Pt(sz, sz)}
						}),
						layout.Rigid(layout.Spacer{Width: unit.Dp(5)}.Layout),
						layout.Rigid(func(gtx C) D {
							l := material.Label(th.Material, unit.Sp(11), strings.ToUpper(name))
							l.Font.Weight = 700
							l.Color = th.Palette.FgSubtle
							return l.Layout(gtx)
						}),
					)
				})
			},
		)
	}
}

func chipDotColor(th *lotusui.Theme, name string) color.NRGBA {
	n := strings.ToLower(name)
	switch {
	case strings.Contains(n, "desktop"), strings.Contains(n, "macos"),
		strings.Contains(n, "windows"), strings.Contains(n, "linux"):
		return th.Palette.BrandFg
	case strings.Contains(n, "mobile"), strings.Contains(n, "android"),
		strings.Contains(n, "ios"):
		return th.Palette.Success
	case strings.Contains(n, "web"), strings.Contains(n, "wasm"):
		return color.NRGBA{R: 0x2A, G: 0x62, B: 0xB5, A: 0xFF}
	default:
		return th.Palette.FgDisabled
	}
}

func (ui *docsUI) renderProps(th *lotusui.Theme, props []docspages.Prop) layout.Widget {
	return func(gtx C) D {
		var rows []layout.Widget
		rows = append(rows, func(gtx C) D {
			return layout.Inset{Top: unit.Dp(44), Bottom: unit.Dp(10)}.Layout(gtx, func(gtx C) D {
				l := material.Label(th.Material, unit.Sp(20), "API")
				l.Font.Weight = 700
				l.Color = th.Palette.Fg
				return l.Layout(gtx)
			})
		})
		rows = append(rows, func(gtx C) D {
			return lotusui.Card(th, lotusui.CardProps{}, func(gtx C) D {
				var inner []layout.Widget
				inner = append(inner, propRow(th, "Option", "Type", "Description", true))
				for _, p := range props {
					p := p
					inner = append(inner, propRow(th, p.Name, p.Type, p.Desc, false))
				}
				return lotusui.VStack(unit.Dp(0), inner...)(gtx)
			})(gtx)
		})
		return lotusui.VStack(unit.Dp(0), rows...)(gtx)
	}
}

func propRow(th *lotusui.Theme, name, typ, desc string, header bool) layout.Widget {
	return func(gtx C) D {
		return layout.Background{}.Layout(gtx,
			func(gtx C) D {
				if header {
					return lotusui.Fill(gtx, th.Palette.BgSubtle)
				}
				return D{Size: gtx.Constraints.Min}
			},
			func(gtx C) D {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						return layout.Inset{
							Top: unit.Dp(9), Bottom: unit.Dp(9),
							Left: unit.Dp(14), Right: unit.Dp(14),
						}.Layout(gtx, func(gtx C) D {
							return layout.Flex{}.Layout(gtx,
								layout.Rigid(func(gtx C) D {
									gtx.Constraints.Min.X = gtx.Dp(unit.Dp(140))
									gtx.Constraints.Max.X = gtx.Dp(unit.Dp(160))
									l := material.Label(th.Material, unit.Sp(14), name)
									if header {
										l = material.Label(th.Material, unit.Sp(12), strings.ToUpper(name))
										l.Font.Weight = 700
										l.Color = th.Palette.FgSubtle
									} else {
										l.Font.Weight = 600
										l.Color = th.Palette.Fg
									}
									return l.Layout(gtx)
								}),
								layout.Rigid(func(gtx C) D {
									gtx.Constraints.Min.X = gtx.Dp(unit.Dp(100))
									gtx.Constraints.Max.X = gtx.Dp(unit.Dp(140))
									l := material.Label(th.Material, unit.Sp(14), typ)
									l.Color = th.Palette.BrandFg
									if header {
										l = material.Label(th.Material, unit.Sp(12), strings.ToUpper(typ))
										l.Font.Weight = 700
										l.Color = th.Palette.FgSubtle
									}
									return l.Layout(gtx)
								}),
								layout.Flexed(1, func(gtx C) D {
									l := material.Label(th.Material, unit.Sp(14), desc)
									l.Color = th.Palette.FgMuted
									if header {
										l = material.Label(th.Material, unit.Sp(12), strings.ToUpper(desc))
										l.Font.Weight = 700
										l.Color = th.Palette.FgSubtle
									}
									return l.Layout(gtx)
								}),
							)
						})
					}),
					layout.Rigid(func(gtx C) D {
						if header {
							return D{}
						}
						sz := image.Pt(gtx.Constraints.Max.X, gtx.Dp(1))
						cl := clip.Rect{Max: sz}.Push(gtx.Ops)
						paint.Fill(gtx.Ops, th.Palette.BorderSubtle)
						cl.Pop()
						return D{Size: sz}
					}),
				)
			},
		)
	}
}

func (ui *docsUI) renderPrevNext(th *lotusui.Theme, prev, next *docspages.Page) layout.Widget {
	return func(gtx C) D {
		left := func(gtx C) D { return D{} }
		right := func(gtx C) D { return D{} }
		if prev != nil {
			if ui.prevBtn.Clicked(gtx) {
				ui.navigate(prev.Slug)
			}
			left = pnCard(th, &ui.prevBtn, "Previous", prev.Title, false)
		}
		if next != nil {
			if ui.nextBtn.Clicked(gtx) {
				ui.navigate(next.Slug)
			}
			right = pnCard(th, &ui.nextBtn, "Next", next.Title, true)
		}
		return layout.Flex{Spacing: layout.SpaceBetween}.Layout(gtx,
			layout.Flexed(1, func(gtx C) D {
				max := gtx.Constraints.Max.X * 48 / 100
				if gtx.Constraints.Max.X > max && max > 0 {
					// keep card ≤48% like .pn
				}
				_ = max
				return left(gtx)
			}),
			layout.Rigid(layout.Spacer{Width: unit.Dp(14)}.Layout),
			layout.Flexed(1, func(gtx C) D {
				return layout.E.Layout(gtx, right)
			}),
		)
	}
}

func pnCard(th *lotusui.Theme, btn *widget.Clickable, label, title string, alignEnd bool) layout.Widget {
	return func(gtx C) D {
		return btn.Layout(gtx, func(gtx C) D {
			pointer.CursorPointer.Add(gtx.Ops)
			maxHalf := gtx.Constraints.Max.X
			if maxHalf > gtx.Constraints.Max.X*48/100 && gtx.Constraints.Max.X > 0 {
				// parent already flexed; constrain visually via content
			}
			_ = maxHalf
			border := th.Palette.Border
			if btn.Hovered() {
				border = th.Palette.BrandFg
			}
			return layout.Background{}.Layout(gtx,
				func(gtx C) D {
					rr := clip.UniformRRect(image.Rectangle{Max: gtx.Constraints.Min}, gtx.Dp(th.Radius.MD)).Push(gtx.Ops)
					paint.Fill(gtx.Ops, th.Palette.BgPanel)
					rr.Pop()
					widget.Border{
						Color: border, Width: unit.Dp(1), CornerRadius: th.Radius.MD,
					}.Layout(gtx, func(gtx C) D { return D{Size: gtx.Constraints.Min} })
					return D{Size: gtx.Constraints.Min}
				},
				func(gtx C) D {
					gtx.Constraints.Min.X = gtx.Constraints.Max.X
					align := layout.W
					if alignEnd {
						align = layout.E
					}
					return layout.Inset{
						Top: unit.Dp(12), Bottom: unit.Dp(12),
						Left: unit.Dp(16), Right: unit.Dp(16),
					}.Layout(gtx, func(gtx C) D {
						return align.Layout(gtx, lotusui.VStack(unit.Dp(2),
							func(gtx C) D {
								l := material.Label(th.Material, unit.Sp(12), label)
								l.Color = th.Palette.FgDisabled
								return l.Layout(gtx)
							},
							func(gtx C) D {
								l := material.Label(th.Material, unit.Sp(15), title)
								l.Font.Weight = 700
								l.Color = th.Palette.BrandFg
								return l.Layout(gtx)
							},
						))
					})
				},
			)
		})
	}
}

func renderHTMLBlock(th *lotusui.Theme, b htmlBlock) layout.Widget {
	switch b.Kind {
	case "h3":
		return func(gtx C) D {
			return layout.Inset{Top: unit.Dp(18), Bottom: unit.Dp(8)}.Layout(gtx, func(gtx C) D {
				l := material.Label(th.Material, unit.Sp(16), b.Text)
				l.Font.Weight = 700
				l.Color = th.Palette.Fg
				return l.Layout(gtx)
			})
		}
	case "takeaway":
		return func(gtx C) D {
			return layout.Inset{Bottom: unit.Dp(12)}.Layout(gtx, func(gtx C) D {
				return lotusui.Card(th, lotusui.CardProps{Variant: lotusui.CardSubtle, Size: lotusui.SizeSM},
					func(gtx C) D {
						l := material.Label(th.Material, unit.Sp(14), b.Text)
						l.Color = th.Palette.Fg
						return l.Layout(gtx)
					},
				)(gtx)
			})
		}
	case "note":
		return func(gtx C) D {
			gtx = constrainCh66(gtx)
			l := material.Label(th.Material, unit.Sp(13), b.Text)
			l.Color = th.Palette.FgSubtle
			return layout.Inset{Bottom: unit.Dp(8)}.Layout(gtx, l.Layout)
		}
	case "table":
		return func(gtx C) D {
			return layout.Inset{Top: unit.Dp(8), Bottom: unit.Dp(14)}.Layout(gtx, func(gtx C) D {
				headers := b.Headers
				rows := b.Rows
				if len(headers) == 0 && len(rows) > 0 {
					// synthesize blank headers matching column count
					headers = make([]string, len(rows[0]))
				}
				// White panel + rounded corners — matches the old
				// .proptable card chrome (bg --card, radius-md).
				return lotusui.Card(th, lotusui.CardProps{
					Variant: lotusui.CardOutline,
					Size:    lotusui.Size2XS,
				}, lotusui.TableText(th, lotusui.TableProps{}, headers, rows))(gtx)
			})
		}
	case "list":
		return func(gtx C) D {
			items := make([]layout.Widget, len(b.Items))
			for i, it := range b.Items {
				it := it
				n := i + 1
				items[i] = func(gtx C) D {
					gtx = constrainCh66(gtx)
					l := material.Label(th.Material, unit.Sp(15), fmt.Sprintf("%d. %s", n, it))
					l.Color = th.Palette.FgMuted
					return layout.Inset{Bottom: unit.Dp(6)}.Layout(gtx, l.Layout)
				}
			}
			return lotusui.VStack(unit.Dp(0), items...)(gtx)
		}
	case "stats":
		return func(gtx C) D {
			cells := make([]lotusui.GridItem, len(b.Stats))
			for i, st := range b.Stats {
				st := st
				cells[i] = lotusui.Cell(func(gtx C) D {
					// White bordered tile — matches old .stat (card fill + border).
					return lotusui.Card(th, lotusui.CardProps{Variant: lotusui.CardOutline, Size: lotusui.SizeSM},
						lotusui.VStack(unit.Dp(4),
							func(gtx C) D {
								l := material.Label(th.Material, unit.Sp(22), st.Num)
								l.Font.Weight = 700
								l.Color = th.Palette.BrandFg
								return l.Layout(gtx)
							},
							func(gtx C) D {
								l := material.Label(th.Material, unit.Sp(13), st.Label)
								l.Font.Weight = 600
								l.Color = th.Palette.Fg
								return l.Layout(gtx)
							},
							func(gtx C) D {
								l := material.Label(th.Material, unit.Sp(12), st.Note)
								l.Color = th.Palette.FgSubtle
								return l.Layout(gtx)
							},
						),
					)(gtx)
				})
			}
			return layout.Inset{Top: unit.Dp(8), Bottom: unit.Dp(16)}.Layout(gtx, func(gtx C) D {
				return lotusui.Grid{Columns: 2, Gap: unit.Dp(10)}.Layout(th, gtx, cells...)
			})
		}
	default: // p
		return func(gtx C) D {
			gtx = constrainCh66(gtx)
			l := material.Label(th.Material, unit.Sp(15), b.Text)
			l.Color = th.Palette.FgMuted
			return layout.Inset{Bottom: unit.Dp(10)}.Layout(gtx, l.Layout)
		}
	}
}

func homePage(th *lotusui.Theme, ui *docsUI) layout.Widget {
	tag := siteTag
	body := "Neutral grays, white cards on a tinted canvas, one accent you choose — a complete UI kit for Go apps that ship to desktop and mobile. One module, one design language; macOS, Windows, Linux, iOS, and Android are compile targets, not ports. The same code reaches the web via WebAssembly when you want it — which is also how every demo on this site is the real component in your browser. Built on Gio."
	ctaTitle := "You're already running it"
	ctaBody := "Every control and demo on this page is lotusui itself, compiled to WebAssembly — not screenshots, not a sandbox. The same Go code builds native apps for macOS, Windows, Linux, iOS and Android."

	ui.tocHeadings = ui.tocHeadings[:0]
	for _, g := range ui.groups {
		ui.tocHeadings = append(ui.tocHeadings, g.Title)
	}
	ui.tocYs = make([]int, len(ui.tocHeadings))

	type part struct {
		w   layout.Widget
		toc int
	}

	return func(gtx C) D {
		var parts []part
		parts = append(parts, part{toc: -1, w: func(gtx C) D {
			return layout.Inset{Bottom: unit.Dp(2)}.Layout(gtx, func(gtx C) D {
				l := material.Label(th.Material, unit.Sp(34), "lotusui")
				l.Font.Weight = 700
				l.Color = th.Palette.Fg
				return l.Layout(gtx)
			})
		}})
		parts = append(parts, part{toc: -1, w: func(gtx C) D {
			l := material.Label(th.Material, unit.Sp(16), tag)
			l.Color = th.Palette.FgSubtle
			return layout.Inset{Bottom: unit.Dp(12)}.Layout(gtx, l.Layout)
		}})
		parts = append(parts, part{toc: -1, w: func(gtx C) D {
			gtx = constrainCh66(gtx)
			l := material.Label(th.Material, unit.Sp(15), body)
			l.Color = th.Palette.FgMuted
			return layout.Inset{Bottom: unit.Dp(10)}.Layout(gtx, l.Layout)
		}})
		// The CTA: title, copy and the button in ONE brand-tinted card.
		// Every ink and fill is a palette token, so switching palette
		// re-colors the card with everything else.
		parts = append(parts, part{toc: -1, w: func(gtx C) D {
			if ui.ctaBtn.Clicked(gtx) {
				ui.navigate("quickstart")
			}
			gtx = constrainCh66(gtx)
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
			return layout.Inset{Top: unit.Dp(6), Bottom: unit.Dp(22)}.Layout(gtx, func(gtx C) D {
				return layout.Background{}.Layout(gtx,
					func(gtx C) D {
						sz := gtx.Constraints.Min
						r := gtx.Dp(th.Radius.LG)
						rr := clip.UniformRRect(image.Rectangle{Max: sz}, r).Push(gtx.Ops)
						paint.Fill(gtx.Ops, th.Palette.BrandSubtle)
						rr.Pop()
						return D{Size: sz}
					},
					func(gtx C) D {
						return layout.UniformInset(unit.Dp(20)).Layout(gtx, func(gtx C) D {
							return lotusui.VStack(unit.Dp(8),
								func(gtx C) D {
									l := material.Label(th.Material, unit.Sp(17), ctaTitle)
									l.Font.Weight = 600
									l.Color = th.Palette.Fg
									return l.Layout(gtx)
								},
								func(gtx C) D {
									l := material.Label(th.Material, unit.Sp(14), ctaBody)
									l.Color = th.Palette.FgMuted
									return layout.Inset{Bottom: unit.Dp(4)}.Layout(gtx, l.Layout)
								},
								func(gtx C) D {
									// Left-aligned: the button hugs its label
									// instead of stretching to the card width.
									gtx.Constraints.Min.X = 0
									return lotusui.Button(th, &ui.ctaBtn, "Get started →", lotusui.ButtonProps{})(gtx)
								},
							)(gtx)
						})
					},
				)
			})
		}})
		// .heroshots — CPU-scaled to Max.X (never op.Affine under
		// ScrollArea: that collapsed the macOS window to 0×0).
		startHeroFetch(ui.invalidate)
		parts = append(parts, part{toc: -1, w: func(gtx C) D {
			ops := heroOpsFor(gtx.Constraints.Max.X)
			if len(ops) == 0 {
				return D{}
			}
			ui.heroOps = ops
			return layout.Inset{Top: unit.Dp(8), Bottom: unit.Dp(28)}.Layout(gtx, func(gtx C) D {
				var shots []layout.Widget
				for i := range ops {
					i := i
					shots = append(shots, heroShot(th, ops[i]))
				}
				return lotusui.VStack(unit.Dp(18), shots...)(gtx)
			})
		}})
		for gi, g := range ui.groups {
			gi, g := gi, g
			parts = append(parts, part{toc: gi, w: func(gtx C) D {
				return layout.Inset{Top: unit.Dp(44), Bottom: unit.Dp(12)}.Layout(gtx, func(gtx C) D {
					l := material.Label(th.Material, unit.Sp(20), g.Title)
					l.Font.Weight = 700
					l.Color = th.Palette.Fg
					return l.Layout(gtx)
				})
			}})
			parts = append(parts, part{toc: -1, w: homeCardsGrid(th, ui, g.Pages)})
		}

		var widgets []layout.Widget
		y := 0
		for _, it := range parts {
			it := it
			widgets = append(widgets, func(gtx C) D {
				if it.toc >= 0 && it.toc < len(ui.tocYs) {
					ui.tocYs[it.toc] = y
				}
				d := it.w(gtx)
				y += d.Size.Y
				return d
			})
		}
		return lotusui.VStack(unit.Dp(0), widgets...)(gtx)
	}
}

func homeCardsGrid(th *lotusui.Theme, ui *docsUI, pages []*docspages.Page) layout.Widget {
	return func(gtx C) D {
		gap := unit.Dp(14)
		minCol := gtx.Dp(unit.Dp(220))
		cols := 1
		if minCol > 0 {
			cols = (gtx.Constraints.Max.X + gtx.Dp(gap)) / (minCol + gtx.Dp(gap))
		}
		if cols < 1 {
			cols = 1
		}
		items := make([]lotusui.GridItem, len(pages))
		for i, p := range pages {
			p := p
			items[i] = lotusui.Cell(pageCard(th, ui, p))
		}
		return lotusui.Grid{Columns: cols, Gap: gap}.Layout(th, gtx, items...)
	}
}

func pageCard(th *lotusui.Theme, ui *docsUI, p *docspages.Page) layout.Widget {
	btn, ok := ui.homeCards[p.Slug]
	if !ok {
		btn = new(widget.Clickable)
		ui.homeCards[p.Slug] = btn
	}
	return func(gtx C) D {
		if btn.Clicked(gtx) {
			ui.navigate(p.Slug)
		}
		return btn.Layout(gtx, func(gtx C) D {
			pointer.CursorPointer.Add(gtx.Ops)
			border := th.Palette.Border
			if btn.Hovered() {
				border = th.Palette.BrandFg
			}
			return layout.Background{}.Layout(gtx,
				func(gtx C) D {
					rr := clip.UniformRRect(image.Rectangle{Max: gtx.Constraints.Min}, gtx.Dp(th.Radius.MD)).Push(gtx.Ops)
					paint.Fill(gtx.Ops, th.Palette.BgPanel)
					rr.Pop()
					widget.Border{
						Color: border, Width: unit.Dp(1), CornerRadius: th.Radius.MD,
					}.Layout(gtx, func(gtx C) D { return D{Size: gtx.Constraints.Min} })
					return D{Size: gtx.Constraints.Min}
				},
				func(gtx C) D {
					gtx.Constraints.Min.X = gtx.Constraints.Max.X
					return layout.Inset{
						Top: unit.Dp(16), Bottom: unit.Dp(16),
						Left: unit.Dp(18), Right: unit.Dp(18),
					}.Layout(gtx, lotusui.VStack(unit.Dp(4),
						func(gtx C) D {
							l := material.Label(th.Material, unit.Sp(14), p.Title)
							l.Font.Weight = 600
							l.Color = th.Palette.Fg
							return l.Layout(gtx)
						},
						func(gtx C) D {
							l := material.Label(th.Material, unit.Sp(13), p.Kicker)
							l.Color = th.Palette.FgSubtle
							return l.Layout(gtx)
						},
					))
				},
			)
		})
	}
}

// heroShot paints a CPU-pre-scaled ImageOp at 1:1 (no op.Affine).
func heroShot(th *lotusui.Theme, imgOp paint.ImageOp) layout.Widget {
	return func(gtx C) D {
		sz := imgOp.Size()
		if sz.X == 0 {
			return D{}
		}
		out := sz
		if maxW := gtx.Constraints.Max.X; maxW > 0 && out.X > maxW {
			out.X = maxW
		}
		r := gtx.Dp(unit.Dp(14))
		defer clip.UniformRRect(image.Rectangle{Max: out}, r).Push(gtx.Ops).Pop()
		paint.Fill(gtx.Ops, th.Palette.BgPanel)
		imgOp.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
		widget.Border{
			Color: th.Palette.Border, Width: unit.Dp(1), CornerRadius: unit.Dp(14),
		}.Layout(gtx, func(gtx C) D { return D{Size: out} })
		return D{Size: out}
	}
}
