package main

import (
	"image"
	"testing"

	"gioui.org/font/gofont"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"

	"github.com/ikaito-com/lotusui/site/live"
)

func TestLayoutAllPages(t *testing.T) {
	live.Embed = true
	th := live.NewTheme()
	th.Material.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))
	ui := newDocsUI()
	var ops op.Ops
	pages := []string{""}
	for _, g := range ui.groups {
		for _, p := range g.Pages {
			pages = append(pages, p.Slug)
		}
	}
	for _, slug := range pages {
		slug := slug
		name := slug
		if name == "" {
			name = "home"
		}
		t.Run(name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("panic: %v", r)
				}
			}()
			ui.page = slug
			ui.secUI = nil
			ui.body.Reset()
			live.ResetEmbed()
			ops.Reset()
			gtx := layout.Context{
				Ops:         &ops,
				Constraints: layout.Exact(image.Pt(1280, 860)),
				Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
			}
			ui.Layout(th, gtx)
		})
	}
}
