package main

import (
	"image/color"
	"strings"
	"sync"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	lotusui "github.com/ikaito-com/lotusui"
	"github.com/ikaito-com/lotusui/site/live"
)

type highlightKey struct {
	src, lang, pal string
}

var highlightCache sync.Map // highlightKey → [][]lotusui.CodeSpan

// highlightLines tokenizes src with chroma into CodeBlock lines,
// matching the old static site's chroma → palette mapping.
func highlightLines(th *lotusui.Theme, src, lang string) [][]lotusui.CodeSpan {
	if lang == "" {
		lang = "go"
	}
	key := highlightKey{src: src, lang: lang, pal: live.CurPalette}
	if v, ok := highlightCache.Load(key); ok {
		return v.([][]lotusui.CodeSpan)
	}
	lexer := lexers.Get(lang)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)
	it, err := lexer.Tokenise(nil, src)
	if err != nil {
		lines := strings.Split(src, "\n")
		out := make([][]lotusui.CodeSpan, len(lines))
		for i, ln := range lines {
			out[i] = []lotusui.CodeSpan{{Text: ln, Color: th.Palette.Fg}}
		}
		highlightCache.Store(key, out)
		return out
	}
	var lines [][]lotusui.CodeSpan
	var cur []lotusui.CodeSpan
	flush := func() {
		lines = append(lines, cur)
		cur = nil
	}
	for _, tok := range it.Tokens() {
		col, bold := tokenStyle(th, tok.Type)
		parts := strings.Split(tok.Value, "\n")
		for i, p := range parts {
			if i > 0 {
				flush()
			}
			if p == "" {
				continue
			}
			cur = append(cur, lotusui.CodeSpan{Text: p, Color: col, Bold: bold})
		}
	}
	flush()
	if len(lines) == 0 {
		lines = [][]lotusui.CodeSpan{{}}
	}
	highlightCache.Store(key, lines)
	return lines
}

func tokenStyle(th *lotusui.Theme, t chroma.TokenType) (color.NRGBA, bool) {
	switch {
	case t.InCategory(chroma.Keyword):
		return th.Palette.BrandFg, true
	case t.InCategory(chroma.String):
		return th.Palette.Success, false
	case t.InCategory(chroma.Comment):
		return th.Palette.FgDisabled, false
	case t.InCategory(chroma.LiteralNumber):
		return th.Palette.Warning, false
	case t == chroma.NameFunction:
		return color.NRGBA{R: 0x82, G: 0x50, B: 0xDF, A: 0xFF}, false
	case t.InCategory(chroma.NameBuiltin), t == chroma.NameBuiltin:
		return th.Palette.Danger, false
	case t.InCategory(chroma.Operator), t.InCategory(chroma.Punctuation):
		return th.Palette.FgSubtle, false
	case t == chroma.KeywordType:
		return th.Palette.BrandFg, false
	default:
		return th.Palette.Fg, false
	}
}
