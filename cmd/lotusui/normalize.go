package main

import (
	"regexp"
	"strings"
)

// SVG normalization happens HERE, at fetch time, so the library ships
// none of it: committed icons are already flat and parseable, and the
// runtime pipeline is read → tint → rasterize, with no text surgery.

// gradientRe matches one <linearGradient>/<radialGradient> block;
// stopRe finds its first stop color. Fluent Color icons paint with
// gradient fills, which the runtime rasterizer can't draw — and which
// the design language doesn't want (flat colors per the theme).
// Every gradient reference is mapped to its first stop color so each
// shape renders as one solid, on-brand-with-the-set flat color.
var (
	gradientRe = regexp.MustCompile(`(?s)<(?:linear|radial)Gradient[^>]*\bid="([^"]+)".*?</(?:linear|radial)Gradient>`)
	stopRe     = regexp.MustCompile(`stop-color="([^"]+)"`)
	// Iconify emits width="1em" height="1em", which the rasterizer
	// can't parse — the icon's scale collapses to zero and it renders
	// nothing. The viewBox alone is sufficient, so em-sized dimensions
	// are dropped.
	emSizeRe = regexp.MustCompile(`\s(?:width|height)="[0-9.]+em"`)
)

func normalizeSVG(src []byte) []byte {
	s := string(src)
	for _, m := range gradientRe.FindAllStringSubmatch(s, -1) {
		id := m[1]
		colorHex := "#888888"
		if sm := stopRe.FindStringSubmatch(m[0]); sm != nil {
			colorHex = sm[1]
		}
		s = strings.ReplaceAll(s, `url(#`+id+`)`, colorHex)
	}
	s = emSizeRe.ReplaceAllString(s, "")
	return []byte(s)
}
