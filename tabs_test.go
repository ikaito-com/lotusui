package lotusui

import (
	"image"
	"testing"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
)

// TestTabsValueSemantics is TestSelectValueSemantics for Tabs — a tab
// strip's order is presentation, so what an app stores or routes on is
// the VALUE, never the position.
func TestTabsValueSemantics(t *testing.T) {
	tabs := Tabs{Options: []TabOption{
		{Label: "Account", Value: "account"},
		{Label: "Password", Value: "password"},
	}}
	if got := tabs.Value(); got != "account" {
		t.Errorf("zero value should select the first tab, got %q", got)
	}
	tabs.SetValue("password")
	if got := tabs.Value(); got != "password" {
		t.Errorf("SetValue/Value round trip broken: %q", got)
	}
	// Reordered and reworded: the stored value still resolves.
	reordered := Tabs{Options: []TabOption{
		{Label: "Security", Value: "password"},
		{Label: "Your account", Value: "account"},
	}}
	reordered.SetValue(tabs.Value())
	if got := reordered.Value(); got != "password" {
		t.Errorf("value must survive a reordered, reworded strip, got %q", got)
	}
	tabs.SetValue("gone")
	if tabs.Chosen() || tabs.Value() != "" {
		t.Errorf("unknown value must clear the selection, got %q", tabs.Value())
	}
	tabs.Clear()
	if tabs.Chosen() {
		t.Error("Clear must leave nothing selected")
	}
	labelled := Tabs{Options: TabOpts("Overview", "Activity")}
	labelled.SetValue("Activity")
	if got := labelled.Value(); got != "Activity" {
		t.Errorf("label-only tab value = %q, want Activity", got)
	}
}

// TestTabsLayoutDoesNotSelect pins the load-bearing ordering contract:
// Layout must NEVER process clicks — Update does, and it must run
// before anything reads the selection in a frame. A consumer that
// forgets Update gets dead tabs (obvious in the first manual test)
// instead of a one-frame lag that survives review.
func TestTabsLayoutDoesNotSelect(t *testing.T) {
	th := NewTheme()
	tabs := Tabs{Options: TabOpts("Account", "Password")}
	var ops op.Ops
	var r input.Router
	gtx := testCtx(&ops, &r, layout.Constraints{Max: image.Pt(400, 200)})
	tabs.SetValue("Password")
	tabs.Layout(th, gtx)
	if got := tabs.Value(); got != "Password" {
		t.Errorf("Layout must not touch the selection, got %q", got)
	}
	tabs.Update(gtx) // no events queued: the selection stands
	if got := tabs.Value(); got != "Password" {
		t.Errorf("Update with no events must not change the selection, got %q", got)
	}
}

// Horizontal Tabs must Wrap under a narrow Max.X — Flex+Rigid used to
// squeeze the last label into a 1-character column (Split half-pane).
func TestTabsHorizontalWrapsUnderNarrowMax(t *testing.T) {
	th := NewTheme()
	tabs := Tabs{Options: TabOpts("Changes", "Staging", "Production")}
	var ops op.Ops
	var r input.Router
	gtx := testCtx(&ops, &r, layout.Constraints{Max: image.Pt(100, 1<<20)})
	tabs.Update(gtx)
	dims := tabs.Layout(th, gtx)
	// One enclosed tab is ~30–40 tall; two Wrap lines ≥ ~70 with gap.
	if dims.Size.Y < 60 {
		t.Fatalf("Tabs height = %d under Max.X=100; want ≥ 60 (wrapped lines)", dims.Size.Y)
	}
	if dims.Size.X > 100 {
		t.Fatalf("Tabs width = %d exceeds Max.X 100", dims.Size.X)
	}
}
