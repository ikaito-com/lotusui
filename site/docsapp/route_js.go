//go:build js

package main

import "syscall/js"

// jsRoute is the live docs route for WASM, and the SOURCE OF TRUTH —
// the URL only mirrors it.
//
// Why the URL must not be navigated: Gio owns "popstate" on the
// window (os_js.go). It turns every history pop into a key.NameBack
// event and, when the app does not consume that key, answers with
// history.back(). Writing location.hash pushes a history entry, so
// the browser lands in Gio's back handler and navigates BACK — the
// click's target page appears for an instant and is then replaced by
// the previous one. Natively there is no history, which is why this
// only ever broke in the browser.
//
// replaceState updates the address bar in place: no entry pushed, no
// popstate, no hashchange — links behave like links, and deep links
// and reloads still work because boot reads location.hash.
var jsRoute string

func currentRoute() string {
	return jsRoute
}

func onRouteChange(invalidate func()) {
	jsRoute = js.Global().Get("location").Get("hash").String()
	js.Global().Get("window").Call("addEventListener", "hashchange",
		js.FuncOf(func(this js.Value, args []js.Value) any {
			jsRoute = js.Global().Get("location").Get("hash").String()
			invalidate()
			return nil
		}))
}

func setRoute(slug string) {
	h := "#" + slug
	if slug == "" {
		h = "#"
	}
	if jsRoute == h || (slug == "" && (jsRoute == "" || jsRoute == "#")) {
		jsRoute = h
		return
	}
	jsRoute = h
	// Mirror to the address bar WITHOUT navigating (see above): a
	// pushed entry would come back as Gio's key.NameBack → history.back().
	hist := js.Global().Get("history")
	if hist.Truthy() && hist.Get("replaceState").Truthy() {
		hist.Call("replaceState", js.Null(), "", h)
		return
	}
	js.Global().Get("location").Set("hash", h)
}

func initPalette(apply func(string)) {
	if v := js.Global().Get("localStorage").Call("getItem", "lotusui-palette"); v.Truthy() {
		apply(v.String())
	}
	js.Global().Get("window").Call("addEventListener", "message",
		js.FuncOf(func(this js.Value, args []js.Value) any {
			if len(args) == 0 {
				return nil
			}
			if p := args[0].Get("data").Get("lotusuiPalette"); p.Truthy() {
				apply(p.String())
			}
			return nil
		}))
}

func initLook(apply func(string)) {
	if v := js.Global().Get("localStorage").Call("getItem", "lotusui-look"); v.Truthy() {
		apply(v.String())
	}
	js.Global().Get("window").Call("addEventListener", "message",
		js.FuncOf(func(this js.Value, args []js.Value) any {
			if len(args) == 0 {
				return nil
			}
			if l := args[0].Get("data").Get("lotusuiLook"); l.Truthy() {
				apply(l.String())
			}
			return nil
		}))
}
