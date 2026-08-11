//go:build js

package main

import (
	"strings"
	"syscall/js"
)

// visitorOS reports the OS of the browser viewing the wasm docs —
// "macos", "windows" or "linux" — or "" when unknown. GOOS is js
// here, so the host OS is unknowable at BUILD time (the ShortcutHint
// rule); at RUNTIME the browser knows it. userAgentData first, then
// the legacy navigator strings. Labels the hero download button only;
// the Desktop app page lists every platform without detection.
func visitorOS() string {
	nav := js.Global().Get("navigator")
	if !nav.Truthy() {
		return ""
	}
	probe := ""
	if uad := nav.Get("userAgentData"); uad.Truthy() {
		if p := uad.Get("platform"); p.Type() == js.TypeString {
			probe = p.String()
		}
	}
	if probe == "" {
		if p := nav.Get("platform"); p.Type() == js.TypeString {
			probe = p.String()
		}
	}
	if probe == "" {
		if ua := nav.Get("userAgent"); ua.Type() == js.TypeString {
			probe = ua.String()
		}
	}
	probe = strings.ToLower(probe)
	switch {
	case strings.Contains(probe, "mac"):
		return "macos"
	case strings.Contains(probe, "win"):
		return "windows"
	case strings.Contains(probe, "linux"), strings.Contains(probe, "x11"):
		return "linux"
	}
	return ""
}
