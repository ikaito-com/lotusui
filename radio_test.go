package lotusui

import "testing"

// TestRadioValueSemantics is TestSelectValueSemantics for RadioGroup:
// the choice contract is shared as behaviour, so it is asserted for
// every component that offers a choice.
func TestRadioValueSemantics(t *testing.T) {
	scope := RadioGroup{Options: []RadioOption{
		{Label: "One per environment (recommended)", Value: "per-env", Description: "Isolates every stage."},
		{Label: "One for all environments", Value: "shared"},
	}}
	// Zero value selects the first option.
	if got := scope.Value(); got != "per-env" {
		t.Errorf("zero value should select the first option, got %q", got)
	}
	scope.SetValue("shared")
	if got := scope.Value(); got != "shared" {
		t.Errorf("SetValue/Value round trip broken: %q", got)
	}
	// REORDERED and REWORDED — the stored value still resolves to the
	// same choice, which is the whole point.
	reordered := RadioGroup{Options: []RadioOption{
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
	// Label-only options: the label IS the value.
	density := RadioGroup{Options: RadioOpts("Default", "Comfortable", "Compact")}
	density.SetValue("Compact")
	if got := density.Value(); got != "Compact" {
		t.Errorf("label-only option value = %q, want Compact", got)
	}
	// Per-option Disabled replaced the index-aligned []bool: a
	// disabled option is still addressable by value, it just cannot be
	// clicked (asserted through layout, not here).
	plans := RadioGroup{Options: []RadioOption{
		{Label: "Starter", Value: "starter"},
		{Label: "Pro", Value: "pro", Disabled: true},
	}}
	plans.SetValue("pro")
	if got := plans.Value(); got != "pro" {
		t.Errorf("disabled options stay addressable by value, got %q", got)
	}
}
