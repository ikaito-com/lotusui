package lotusui

import "testing"

// TestToggleGroupValueSemantics covers both modes of the choice
// contract: single-select speaks Value/SetValue, multi-select speaks
// Values/SetValues — never a []bool aligned to a list's order.
func TestToggleGroupValueSemantics(t *testing.T) {
	view := ToggleGroup{Options: ToggleOpts("All", "Missed")}
	if got := view.Value(); got != "All" {
		t.Errorf("zero value should select the first option, got %q", got)
	}
	view.SetValue("Missed")
	if got := view.Value(); got != "Missed" {
		t.Errorf("SetValue/Value round trip broken: %q", got)
	}
	view.SetValue("gone")
	if view.Chosen() || view.Value() != "" {
		t.Errorf("unknown value must clear the selection, got %q", view.Value())
	}

	// Multi-select: icon-only options MUST carry explicit values,
	// since an empty Label has no value to store.
	marks := ToggleGroup{Multiple: true, Options: []ToggleOption{
		{Value: "bold", Icon: IconTextBold},
		{Value: "italic", Icon: IconTextItalic},
		{Value: "underline", Icon: IconTextUnderline},
	}}
	marks.SetValues([]string{"underline", "bold", "gone"})
	got := marks.Values()
	// Values come back in OPTIONS order, not the caller's order, and
	// unknown values are ignored.
	if len(got) != 2 || got[0] != "bold" || got[1] != "underline" {
		t.Errorf("Values = %v, want [bold underline]", got)
	}
	if !marks.Chosen() {
		t.Error("Chosen must report a multi-select selection")
	}
	// A reordered, reworded list keeps the same stored marks.
	reordered := ToggleGroup{Multiple: true, Options: []ToggleOption{
		{Value: "underline", Icon: IconTextUnderline},
		{Value: "bold", Icon: IconTextBold},
	}}
	reordered.SetValues(got)
	if v := reordered.Values(); len(v) != 2 || v[0] != "underline" || v[1] != "bold" {
		t.Errorf("marks must survive a reordered list, got %v", v)
	}
	marks.Clear()
	if len(marks.Values()) != 0 || marks.Chosen() {
		t.Error("Clear must leave nothing selected")
	}
	// Single-select mode reads through Value, and Values reports at
	// most one entry.
	if v := view.Values(); len(v) != 0 {
		t.Errorf("cleared single-select must report no values, got %v", v)
	}
}
