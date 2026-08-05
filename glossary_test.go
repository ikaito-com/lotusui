package lotusui

import (
	"image"
	"reflect"
	"testing"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
)

func TestSplitGlossaryLongestFirst(t *testing.T) {
	terms := []GlossaryTerm{
		{Term: "API", Tip: "Application Programming Interface"},
		{Term: "SLA", Tip: "Service Level Agreement"},
		{Term: "AP", Tip: "should lose to API"},
	}
	got := SplitGlossary("The API and SLA matter; also AP alone.", terms)
	want := []GlossarySeg{
		{Text: "The "},
		{Text: "API", Term: true},
		{Text: " and "},
		{Text: "SLA", Term: true},
		{Text: " matter; also "},
		{Text: "AP", Term: true},
		{Text: " alone."},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v\nwant %#v", got, want)
	}
}

func TestSplitGlossaryEmpty(t *testing.T) {
	if SplitGlossary("", nil) != nil {
		t.Fatal("empty text")
	}
	got := SplitGlossary("plain", nil)
	if len(got) != 1 || got[0].Term || got[0].Text != "plain" {
		t.Fatalf("no terms: %#v", got)
	}
}

func TestAnnotatedTextLenMismatchNoPanic(t *testing.T) {
	th := NewTheme()
	terms := []GlossaryTerm{{Term: "API", Tip: "tip"}, {Term: "SLA", Tip: "tip2"}}
	var ops op.Ops
	var r input.Router
	gtx := testCtx(&ops, &r, layout.Constraints{Max: image.Pt(400, 40)})
	// Fewer cards than terms — must not panic.
	AnnotatedText(th, "API then SLA", terms, []*HoverCard{new(HoverCard)})(gtx)
	// Nil cards slice.
	AnnotatedText(th, "API then SLA", terms, nil)(gtx)
}

// AnnotatedText must Wrap segments — Flex+Rigid squeezed labels under
// a narrow Max.X (the bug Wrap was added to fix).
func TestAnnotatedTextWrapsUnderNarrowMax(t *testing.T) {
	th := NewTheme()
	terms := []GlossaryTerm{
		{Term: "XXXXXXXX", Tip: "a"},
		{Term: "YYYYYYYY", Tip: "b"},
	}
	var ops op.Ops
	var r input.Router
	gtx := testCtx(&ops, &r, layout.Constraints{Max: image.Pt(50, 1<<20)})
	dims := AnnotatedText(th, "XXXXXXXX YYYYYYYY", terms, nil)(gtx)
	if dims.Size.Y < 28 {
		t.Fatalf("AnnotatedText height = %d under Max.X=50; want ≥ 28 (wrapped)", dims.Size.Y)
	}
}
