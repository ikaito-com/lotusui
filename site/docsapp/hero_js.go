//go:build js

package main

import "gioui.org/op/paint"

// Heroes stay on media/ for Pages; WASM does not sync-XHR them (that
// froze the tab). Optional async fetch can return later via invalidate.
func loadHeroImages() []paint.ImageOp { return nil }

func heroOpsFor(int) []paint.ImageOp { return nil }

func startHeroFetch(func()) {}
