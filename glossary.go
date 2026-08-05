package lotusui

import (
	"image"
	"strings"

	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op/clip"
)

// GlossaryTerm is one in-text term with a HoverCard tip. Term is matched
// as a literal substring; longest Term wins when spans overlap.
type GlossaryTerm struct {
	Term string
	Tip  string // Caption inside the HoverCard
}

// GlossarySeg is one run from SplitGlossary: either plain text or a term.
type GlossarySeg struct {
	Text string
	Term bool
}

// SplitGlossary splits text into plain and term segments. Matching is
// literal and longest-first (same span: longer Term wins). Empty text
// or no terms yields a single plain segment when text is non-empty.
func SplitGlossary(text string, terms []GlossaryTerm) []GlossarySeg {
	if text == "" {
		return nil
	}
	if len(terms) == 0 {
		return []GlossarySeg{{Text: text}}
	}
	var out []GlossarySeg
	rest := text
	for len(rest) > 0 {
		earliest := -1
		hit := ""
		for _, t := range terms {
			if t.Term == "" {
				continue
			}
			i := strings.Index(rest, t.Term)
			if i < 0 {
				continue
			}
			if earliest < 0 || i < earliest || (i == earliest && len(t.Term) > len(hit)) {
				earliest, hit = i, t.Term
			}
		}
		if earliest < 0 {
			out = append(out, GlossarySeg{Text: rest})
			break
		}
		if earliest > 0 {
			out = append(out, GlossarySeg{Text: rest[:earliest]})
		}
		out = append(out, GlossarySeg{Text: hit, Term: true})
		rest = rest[earliest+len(hit):]
	}
	return out
}

// AnnotatedText lays out text with glossary terms in BrandFg. When
// cards[i] is non-nil for terms[i], that term gets a HoverCard tip
// (Tip as Caption). A nil card or missing index paints BrandFg only —
// len(cards) may differ from len(terms); never panics. Segments flow
// with Wrap (gap 0) so a narrow Max.X never squeezes labels into
// 1-character columns. Does not change HoverCard defaults (Width max
// 320, OpenDelay 700ms); compact UIs set Width / OpenDelay / Side on
// each card. Width is a max — the card hugs content so short tips
// stay on the trigger.
func AnnotatedText(th *Theme, text string, terms []GlossaryTerm, cards []*HoverCard) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		segs := SplitGlossary(text, terms)
		if len(segs) == 1 && !segs[0].Term {
			return LabelBody(th, segs[0].Text).Layout(gtx)
		}
		tipFor := func(term string) string {
			for _, t := range terms {
				if t.Term == term {
					return t.Tip
				}
			}
			return ""
		}
		cardFor := func(term string) *HoverCard {
			for i, t := range terms {
				if t.Term != term {
					continue
				}
				if i < len(cards) {
					return cards[i]
				}
				return nil
			}
			return nil
		}
		var children []layout.Widget
		for _, seg := range segs {
			seg := seg
			if !seg.Term {
				t := seg.Text
				children = append(children, func(gtx layout.Context) layout.Dimensions {
					return LabelBody(th, t).Layout(gtx)
				})
				continue
			}
			hc := cardFor(seg.Text)
			tip := tipFor(seg.Text)
			term := seg.Text
			if hc == nil || tip == "" {
				children = append(children, func(gtx layout.Context) layout.Dimensions {
					l := LabelBody(th, term)
					l.Color = th.Palette.BrandFg
					return l.Layout(gtx)
				})
				continue
			}
			children = append(children, func(gtx layout.Context) layout.Dimensions {
				return hc.Layout(th, gtx,
					func(gtx layout.Context) layout.Dimensions {
						// HoverCard already insets MD — Caption only, no extra pad.
						l := LabelCaption(th, tip)
						l.Color = th.Palette.Fg
						return l.Layout(gtx)
					},
					func(gtx layout.Context) layout.Dimensions {
						l := LabelBody(th, term)
						l.Color = th.Palette.BrandFg
						dims := l.Layout(gtx)
						defer clip.Rect(image.Rectangle{Max: dims.Size}).Push(gtx.Ops).Pop()
						pointer.CursorPointer.Add(gtx.Ops)
						return dims
					},
				)
			})
		}
		// Wrap (gap 0, Middle) — same cross-axis as the old Flex, but
		// segments flow to the next line at intrinsic width instead of
		// squeezing into 1-character columns under a narrow Max.X.
		return Wrap(0, layout.Middle, children...)(gtx)
	}
}
