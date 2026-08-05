//go:build js

package main

import "syscall/js"

// jsRoute is the live docs hash for WASM. location.hash alone is not
// enough: a failed or deferred hash write lets syncRoute snap the UI
// back to the previous page on the next frame (reads as “browser back”).
// Keep an in-memory source of truth, mirror it to the URL for sharing /
// back-forward, and accept hashchange from the browser.
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
