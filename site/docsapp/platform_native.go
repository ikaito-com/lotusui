//go:build !js

package main

// visitorOS is the wasm-only hero-download hint. Natively you ARE
// running the desktop app already, so no OS is reported and the
// download button never renders.
func visitorOS() string { return "" }
