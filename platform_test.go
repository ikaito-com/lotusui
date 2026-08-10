package lotusui

import (
	"testing"

	"gioui.org/io/key"
	"gioui.org/io/pointer"
)

// TestContextMenuPressConventions exercises every OS's context-menu
// gesture from one build machine via the platformOS seam.
func TestContextMenuPressConventions(t *testing.T) {
	defer func(os string) { platformOS = os }(platformOS)

	press := func(btns pointer.Buttons, mods key.Modifiers) pointer.Event {
		return pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Buttons: btns, Modifiers: mods}
	}
	cases := []struct {
		name string
		os   string
		ev   pointer.Event
		want bool
	}{
		{"secondary/linux", "linux", press(pointer.ButtonSecondary, 0), true},
		{"secondary/windows", "windows", press(pointer.ButtonSecondary, 0), true},
		{"secondary/darwin", "darwin", press(pointer.ButtonSecondary, 0), true},
		{"secondary/js", "js", press(pointer.ButtonSecondary, 0), true},
		{"ctrl-primary/darwin", "darwin", press(pointer.ButtonPrimary, key.ModCtrl), true},
		{"ctrl-primary/linux", "linux", press(pointer.ButtonPrimary, key.ModCtrl), false},
		{"ctrl-primary/windows", "windows", press(pointer.ButtonPrimary, key.ModCtrl), false},
		{"plain-primary/darwin", "darwin", press(pointer.ButtonPrimary, 0), false},
		{"release-not-press/darwin", "darwin",
			pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, Buttons: pointer.ButtonSecondary}, false},
		{"move-not-press/linux", "linux",
			pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Buttons: pointer.ButtonSecondary}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			platformOS = c.os
			if got := ContextMenuPress(c.ev); got != c.want {
				t.Errorf("ContextMenuPress on %s = %v, want %v", c.os, got, c.want)
			}
		})
	}
}

func TestShortcutHint(t *testing.T) {
	defer func(os string) { platformOS = os }(platformOS)

	for os, want := range map[string]string{
		"darwin":  "⌘C",
		"linux":   "Ctrl+C",
		"windows": "Ctrl+C",
		"js":      "Ctrl+C",
	} {
		platformOS = os
		if got := ShortcutHint("C"); got != want {
			t.Errorf("ShortcutHint(%q) on %s = %q, want %q", "C", os, got, want)
		}
	}
}
