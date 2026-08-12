//go:build !darwin

package main

import (
	"gioui.org/app"
	"gioui.org/io/event"
	"gioui.org/unit"
)

// Non-macOS builds (wasm, Windows, Linux): standard window chrome,
// no traffic-light inset. See chrome_darwin.go for the seamless pair.

func chromeOptions() []app.Option { return nil }

func applyChrome(event.Event) {}

func topbarLeftInset() unit.Dp { return 0 }
