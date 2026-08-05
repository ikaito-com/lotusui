package main

import (
	"gioui.org/layout"
	"gioui.org/unit"

	lotusui "github.com/ikaito-com/lotusui"
	"github.com/ikaito-com/lotusui/site/docspages"
)

func (ui *docsUI) pageContent(th *lotusui.Theme) layout.Widget {
	if ui.page == "" || ui.page == "home" {
		return homePage(th, ui)
	}
	p, group, prev, next := docspages.Lookup(docsGroups, ui.page)
	if p == nil {
		ui.tocHeadings = nil
		ui.tocYs = nil
		return lotusui.VStack(unit.Dp(16),
			lotusui.LabelHero(th, pageTitle(ui.page)).Layout,
			lotusui.LabelBody(th, "This page is not in docspages.").Layout,
		)
	}
	return ui.renderDocPage(th, p, group, prev, next)
}
