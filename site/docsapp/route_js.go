//go:build js

package main

import "syscall/js"

func currentRoute() string {
	return js.Global().Get("location").Get("hash").String()
}

func onRouteChange(invalidate func()) {
	js.Global().Get("window").Call("addEventListener", "hashchange",
		js.FuncOf(func(this js.Value, args []js.Value) any {
			invalidate()
			return nil
		}))
}

func setRoute(slug string) {
	h := "#" + slug
	if slug == "" {
		h = "#"
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
