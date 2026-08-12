//go:build darwin

package main

import (
	"gioui.org/app"
	"gioui.org/io/event"
	"gioui.org/unit"

	lotusui "github.com/ikaito-com/lotusui"
)

// The seamless macOS window, the same pairing the consumer apps use
// (see macchrome_darwin.go): app.Decorated(false) makes Gio hide the
// titlebar natively AND skip its fallback decorations, and
// MakeSeamlessWindow adds what that leaves out — visible traffic
// lights and the no-flash titlebar hide. Only the native macOS build
// gets this; wasm has no window chrome and Windows/Linux keep their
// standard decorations.

func chromeOptions() []app.Option {
	return []app.Option{app.Decorated(false)}
}

// applyChrome runs on every window event: the AppKit view can
// re-attach, and MakeSeamlessWindow is idempotent.
func applyChrome(e event.Event) {
	if v, ok := e.(app.AppKitViewEvent); ok {
		lotusui.MakeSeamlessWindow(v.View)
	}
}

// topbarLeftInset clears the traffic lights, which overlay the
// seamless window's top-left corner where the wordmark sits.
func topbarLeftInset() unit.Dp { return 78 }
