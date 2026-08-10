package lotusui

import (
	"runtime"

	"gioui.org/io/key"
	"gioui.org/io/pointer"
)

// The platform layer: the one place OS interaction conventions are
// answered. Components consult these instead of runtime.GOOS so each
// convention lives — and is tested, from any build machine — in
// exactly one place.

// platformOS is runtime.GOOS behind a seam so tests exercise every
// OS's convention without cross-compiling.
var platformOS = runtime.GOOS

// ContextMenuPress reports whether ev is the platform's context-menu
// gesture: a secondary-button press on every OS, plus Ctrl+primary on
// macOS — the one-button convention, which neither macOS nor Gio
// translates into a secondary button for us.
func ContextMenuPress(ev pointer.Event) bool {
	if ev.Kind != pointer.Press {
		return false
	}
	if ev.Buttons.Contain(pointer.ButtonSecondary) {
		return true
	}
	return platformOS == "darwin" &&
		ev.Buttons.Contain(pointer.ButtonPrimary) &&
		ev.Modifiers.Contain(key.ModCtrl)
}

// ShortcutHint renders the platform spelling of the shortcut modifier
// plus key for display: ShortcutHint("C") is "⌘C" on macOS and
// "Ctrl+C" elsewhere — including the browser, where the build cannot
// know the host OS. Display only (menu rows, Kbd); binding the key is
// the caller's, e.g. via Gio's key.ModShortcut, which follows the
// same convention.
func ShortcutHint(k string) string {
	if platformOS == "darwin" {
		return "⌘" + k
	}
	return "Ctrl+" + k
}
