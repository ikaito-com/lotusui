package lotusui

import (
	"image"
	"image/color"
	"io"
	"strings"

	"gioui.org/io/clipboard"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// CodeSpan is one styled run inside a CodeBlock line — typically a
// highlighter token. The library does not embed a highlighter: apps
// (or the docs site) tokenize at build time or with chroma and pass
// spans here, keeping the module zip lean.
type CodeSpan struct {
	Text  string
	Color color.NRGBA
	Bold  bool
}

// CodeBlockProps are the code block's props. Prefer Lines (highlighted);
// Plain is the unstyled fallback when Lines is empty.
type CodeBlockProps struct {
	Lang  string       // caption (empty → "Go")
	Lines [][]CodeSpan // rows of token runs
	Plain string       // used when Lines is empty
	// Copy, when non-nil, shows a Copy button that writes Plain (or
	// joined Lines text) to the clipboard.
	Copy *widget.Clickable
	// Nested omits the outer border/radius — use inside a parent card
	// (e.g. docs Preview|Code chrome).
	Nested bool
}

// CodeBlock is a lotusui extension for fenced source: figcaption +
// pre-formatted body, matching the docs site's .codeblock chrome
// (surface2 caption bar, 13sp body, optional Copy).
func CodeBlock(th *Theme, o CodeBlockProps) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		label := "GO"
		if o.Lang != "" {
			label = strings.ToUpper(o.Lang)
		}
		copyText := o.Plain
		if copyText == "" && len(o.Lines) > 0 {
			var b strings.Builder
			for i, line := range o.Lines {
				if i > 0 {
					b.WriteByte('\n')
				}
				for _, s := range line {
					b.WriteString(s.Text)
				}
			}
			copyText = b.String()
		}
		if o.Copy != nil && o.Copy.Clicked(gtx) && copyText != "" {
			gtx.Execute(clipboard.WriteCmd{
				Type: "application/text",
				Data: io.NopCloser(strings.NewReader(copyText)),
			})
		}

		body := func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Background{}.Layout(gtx,
						func(gtx layout.Context) layout.Dimensions {
							return Fill(gtx, th.Palette.BgSubtle)
						},
						func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{
								Top: unit.Dp(7), Bottom: unit.Dp(7),
								Left: unit.Dp(14), Right: unit.Dp(14),
							}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								gtx.Constraints.Min.X = gtx.Constraints.Max.X
								return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
									layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
										l := material.Label(th.Material, unit.Sp(11), label)
										l.Font.Weight = 700
										l.Color = th.Palette.FgDisabled
										return l.Layout(gtx)
									}),
									layout.Rigid(codeCopyBtn(th, o.Copy)),
								)
							})
						},
					)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					sz := image.Pt(gtx.Constraints.Max.X, gtx.Dp(1))
					cl := clip.Rect{Max: sz}.Push(gtx.Ops)
					paint.Fill(gtx.Ops, th.Palette.BorderSubtle)
					cl.Pop()
					return layout.Dimensions{Size: sz}
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{
						Top: unit.Dp(14), Bottom: unit.Dp(14),
						Left: unit.Dp(18), Right: unit.Dp(18),
					}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						if len(o.Lines) > 0 {
							return layoutCodeLines(th, gtx, o.Lines)
						}
						l := material.Label(th.Material, unit.Sp(13), o.Plain)
						l.Color = th.Palette.Fg
						return l.Layout(gtx)
					})
				}),
			)
		}

		if o.Nested {
			return body(gtx)
		}
		return codeBlockFrame(th, body)(gtx)
	}
}

func codeCopyBtn(th *Theme, btn *widget.Clickable) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		if btn == nil {
			return layout.Dimensions{}
		}
		return btn.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			pointer.CursorPointer.Add(gtx.Ops)
			border := th.Palette.Border
			ink := th.Palette.FgSubtle
			if btn.Hovered() {
				border = th.Palette.BrandFg
				ink = th.Palette.BrandFg
			}
			return layout.Background{}.Layout(gtx,
				func(gtx layout.Context) layout.Dimensions {
					rr := clip.UniformRRect(image.Rectangle{Max: gtx.Constraints.Min}, gtx.Dp(6)).Push(gtx.Ops)
					paint.Fill(gtx.Ops, th.Palette.BgPanel)
					rr.Pop()
					widget.Border{
						Color: border, Width: unit.Dp(1), CornerRadius: unit.Dp(6),
					}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Dimensions{Size: gtx.Constraints.Min}
					})
					return layout.Dimensions{Size: gtx.Constraints.Min}
				},
				func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{
						Top: unit.Dp(4), Bottom: unit.Dp(4),
						Left: unit.Dp(10), Right: unit.Dp(10),
					}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						l := material.Label(th.Material, unit.Sp(12), "Copy")
						l.Font.Weight = 600
						l.Color = ink
						return l.Layout(gtx)
					})
				},
			)
		})
	}
}

func layoutCodeLines(th *Theme, gtx layout.Context, lines [][]CodeSpan) layout.Dimensions {
	lineH := gtx.Sp(unit.Sp(21)) // ~13sp × 1.6
	var rows []layout.Widget
	for _, line := range lines {
		line := line
		rows = append(rows, func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.Y = lineH
			if len(line) == 0 {
				return layout.Dimensions{Size: image.Pt(0, lineH)}
			}
			children := make([]layout.FlexChild, 0, len(line))
			for _, s := range line {
				s := s
				children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					l := material.Label(th.Material, unit.Sp(13), s.Text)
					l.Color = s.Color
					l.MaxLines = 1
					if s.Bold {
						l.Font.Weight = 600
					}
					return l.Layout(gtx)
				}))
			}
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx, children...)
		})
	}
	return VStack(unit.Dp(0), rows...)(gtx)
}

func codeBlockFrame(th *Theme, content layout.Widget) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		// Measure then paint border last (same Floating-safe pattern as Card).
		mdims := measureContent(gtx, content)
		sz := image.Pt(gtx.Constraints.Max.X, mdims.Size.Y)
		r := gtx.Dp(th.Radius.MD)
		cl := clip.UniformRRect(image.Rectangle{Max: sz}, r).Push(gtx.Ops)
		paint.Fill(gtx.Ops, th.Palette.BgPanel)
		content(gtx)
		cl.Pop()
		widget.Border{
			Color: th.Palette.Border, Width: unit.Dp(1), CornerRadius: th.Radius.MD,
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Dimensions{Size: sz}
		})
		return layout.Dimensions{Size: sz}
	}
}
