//go:build !js

package main

import "os"

// currentRoute on native builds comes from the environment, so any
// state is reachable without click choreography:
//
//	LOTUSUI_GALLERY=modal/open go run ./gallery
func currentRoute() string {
	return os.Getenv("LOTUSUI_GALLERY")
}

func onRouteChange(func()) {}

func reportHeight(int) {}

// initPalette on native builds reads LOTUSUI_PALETTE once — the same
// no-choreography principle as LOTUSUI_GALLERY.
func initPalette(apply func(string)) {
	if v := os.Getenv("LOTUSUI_PALETTE"); v != "" {
		apply(v)
	}
}

func initOverlay(func()) {}

func sendMeasured([]int) {}

// initLook on native builds reads LOTUSUI_LOOK once.
func initLook(apply func(string)) {
	if v := os.Getenv("LOTUSUI_LOOK"); v != "" {
		apply(v)
	}
}

func setRegionScroll(int, int8) {}

func initWheelRouter() {}

func notifyPainted() {}
