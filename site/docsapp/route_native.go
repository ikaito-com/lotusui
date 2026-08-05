//go:build !js

package main

import "os"

// nativeRoute is the live docs hash for desktop. LOTUSUI_DOCS is only
// the initial route — setRoute must mutate this or syncRoute snaps
// every click back to the env value.
var nativeRoute = os.Getenv("LOTUSUI_DOCS")

func currentRoute() string { return nativeRoute }

func onRouteChange(func()) {}

func setRoute(r string) { nativeRoute = r }

func initPalette(apply func(string)) {
	if v := os.Getenv("LOTUSUI_PALETTE"); v != "" {
		apply(v)
	}
}

func initLook(apply func(string)) {
	if v := os.Getenv("LOTUSUI_LOOK"); v != "" {
		apply(v)
	}
}
