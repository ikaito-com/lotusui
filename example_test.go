package lotusui_test

import (
	"gioui.org/layout"
	"gioui.org/widget"

	lotusui "github.com/ikaito-com/lotusui"
)

// These examples render on pkg.go.dev as runnable usage docs. They
// have no Output comments — they compile with the test suite (so the
// documented API can never drift from the real one) but need a window
// to actually draw.

func ExampleButton() {
	th := lotusui.NewTheme()
	var save widget.Clickable
	w := lotusui.Button(th, &save, "Save", lotusui.ButtonProps{Loading: false})
	_ = w // lay out inside a frame: w(gtx)
}

func ExampleButtonProps_scheme() {
	th := lotusui.NewTheme()
	var btn widget.Clickable
	teal := lotusui.Teal.Scheme() // any ColorScale becomes a button scheme
	w := lotusui.Button(th, &btn, "Go", lotusui.ButtonProps{Scheme: &teal})
	_ = w
}

func ExampleVStack() {
	th := lotusui.NewTheme()
	w := lotusui.VStack(th.Space.MD,
		lotusui.LabelTitle(th, "Section").Layout,
		lotusui.Hairline(th),
		lotusui.LabelBody(th, "Body under the divider").Layout,
	)
	_ = w
}

func ExampleNewTheme_customPalette() {
	p := lotusui.DefaultPalette
	p.BrandFg = lotusui.Teal.C600 // one brand color…
	th := lotusui.NewTheme(lotusui.WithPalette(p))
	_ = th.BrandScale // …and the graded 50…900 scale is derived automatically
}

func ExampleListView() {
	th := lotusui.NewTheme()
	list := widget.List{List: layout.List{Axis: layout.Vertical}}
	items := []string{"a", "b", "c"}
	render := func(gtx layout.Context) layout.Dimensions {
		return lotusui.ListView(th, &list, gtx, len(items), func(gtx layout.Context, i int) layout.Dimensions {
			return lotusui.LabelBody(th, items[i]).Layout(gtx)
		})
	}
	_ = render
}
