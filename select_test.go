package lotusui

import (
	"testing"

	"gioui.org/layout"
)

// A Select speaks VALUES, never indexes: reordering or rewording the
// options must not change what a chosen option means. The index is
// unexported precisely so it cannot leak into app state or storage.
func TestSelectValueSemantics(t *testing.T) {
	scope := Select{Options: []SelectOption{
		{Label: "One per environment (recommended)", Value: "per-env"},
		{Label: "One for all environments", Value: "shared"},
	}}
	// Zero value selects the first option, like a <select> with no
	// `selected` attribute.
	if got := scope.Value(); got != "per-env" {
		t.Errorf("zero value should select the first option, got %q", got)
	}
	scope.SetValue("shared")
	if got := scope.Value(); got != "shared" {
		t.Errorf("SetValue/Value round trip broken: %q", got)
	}
	// REORDERED and REWORDED — the stored value still resolves to the
	// same choice, which is the whole point.
	reordered := Select{Options: []SelectOption{
		{Label: "Shared across environments", Value: "shared"},
		{Label: "Per environment", Value: "per-env"},
	}}
	reordered.SetValue(scope.Value())
	if got := reordered.Value(); got != "shared" {
		t.Errorf("value must survive a reordered, reworded list, got %q", got)
	}
	// An unknown value clears rather than silently meaning option 0.
	scope.SetValue("gone")
	if scope.Chosen() || scope.Value() != "" {
		t.Errorf("unknown value must clear the choice, got %q", scope.Value())
	}
	scope.Clear()
	if scope.Chosen() {
		t.Error("Clear must leave nothing chosen")
	}
	// Label-only options: the label IS the value (HTML's rule).
	sizes := Select{Options: SelectOpts("10", "25", "50")}
	sizes.SetValue("25")
	if got := sizes.Value(); got != "25" {
		t.Errorf("label-only option value = %q, want 25", got)
	}
	// Groups flatten in order and carry values the same way.
	grouped := Select{Groups: []SelectGroup{
		{Label: "Fruit", Options: []SelectOption{{Label: "Apple", Value: "apple"}}},
		{Label: "Veg", Options: []SelectOption{{Label: "Leek", Value: "leek"}}},
	}}
	grouped.SetValue("leek")
	if got := grouped.Value(); got != "leek" {
		t.Errorf("grouped value = %q, want leek", got)
	}
	// Meta is display-only — it must not affect Value().
	withMeta := Select{Options: []SelectOption{
		{Label: "Demo", Value: "p1", Meta: "3"},
		{Label: "Other", Value: "p2", Meta: "0"},
	}}
	withMeta.SetValue("p1")
	if got := withMeta.Value(); got != "p1" {
		t.Errorf("Meta must not affect Value, got %q", got)
	}
	if withMeta.Options[0].Meta != "3" {
		t.Errorf("Meta not preserved on option")
	}
	// Build-time composition constructors.
	built := Select{Options: SelectItems(
		SelectItem("Apple"),
		SelectItemValue("ban", "Banana"),
	)}
	built.SetValue("ban")
	if got := built.Value(); got != "ban" {
		t.Errorf("SelectItems/SelectItemValue = %q, want ban", got)
	}
	gbuilt := Select{Groups: SelectGroups(
		SelectGrouped("Fruit", SelectItemValue("a", "Apple")),
		SelectGrouped("Veg", SelectItemValue("l", "Leek")),
	)}
	gbuilt.SetValue("l")
	if got := gbuilt.Value(); got != "l" {
		t.Errorf("SelectGroups/SelectGrouped = %q, want l", got)
	}
	// Content is display-only — it must not affect Value().
	withContent := Select{Options: SelectItems(SelectOption{
		Label: "Pro", Value: "pro",
		Content: func(gtx layout.Context) layout.Dimensions { return layout.Dimensions{} },
	})}
	withContent.SetValue("pro")
	if got := withContent.Value(); got != "pro" {
		t.Errorf("Content must not affect Value, got %q", got)
	}
}
