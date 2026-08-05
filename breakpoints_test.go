package lotusui

import (
	"image"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

func TestParseBreakpointsJSON(t *testing.T) {
	bp, err := ParseBreakpointsJSON([]byte(`{"md":768,"base":0,"sm":480}`))
	if err != nil {
		t.Fatal(err)
	}
	if bp.Len() != 3 || bp.Name(0) != "base" || bp.Min(1) != 480 || bp.Name(2) != "md" {
		t.Fatalf("got names=%v mins=%v", bp.names, bp.mins)
	}
}

func TestParseBreakpointsJSONRejectsEmpty(t *testing.T) {
	if _, err := ParseBreakpointsJSON([]byte(`{}`)); err == nil {
		t.Fatal("want error")
	}
}

func TestResponsiveIntMobileFirst(t *testing.T) {
	th := NewTheme()
	r := Cols(1).At("md", 2).At("lg", 4)
	cases := []struct {
		w, want int
	}{
		{320, 1},
		{480, 1}, // sm — no override
		{768, 2},
		{992, 4},
		{1600, 4},
	}
	var ops op.Ops
	rt := new(input.Router)
	for _, c := range cases {
		gtx := layout.Context{
			Ops:         &ops,
			Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
			Constraints: layout.Constraints{Max: image.Pt(c.w, 400)},
			Now:         time.Unix(1, 0),
			Source:      rt.Source(),
		}
		if got := r.Resolve(th, gtx); got != c.want {
			t.Fatalf("w=%d: got %d want %d (bp=%s)", c.w, got, c.want, th.BreakpointName(gtx))
		}
	}
}

func TestResponsiveIntCustomBreakpoints(t *testing.T) {
	bp := MustBreakpoints([]Breakpoint{
		{Name: "base", Min: 0},
		{Name: "tablet", Min: 600},
		{Name: "desktop", Min: 1000},
	})
	th := NewTheme(WithBreakpoints(bp))
	r := Cols(1).At("tablet", 3).At("desktop", 5)
	var ops op.Ops
	rt := new(input.Router)
	gtx := layout.Context{
		Ops: &ops, Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Constraints: layout.Constraints{Max: image.Pt(700, 400)},
		Now:         time.Unix(1, 0), Source: rt.Source(),
	}
	if got := r.Resolve(th, gtx); got != 3 {
		t.Fatalf("got %d want 3", got)
	}
	if th.BreakpointName(gtx) != "tablet" {
		t.Fatalf("name %q", th.BreakpointName(gtx))
	}
}

func TestGridColsResponsive(t *testing.T) {
	th := NewTheme()
	var ops op.Ops
	rt := new(input.Router)
	tile := func(gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: image.Pt(gtx.Constraints.Max.X, 20)}
	}
	g := Grid{Cols: Cols(1).At("md", 2).At("lg", 4), Gap: Space.SM}
	// Narrow → 1 column → one row of 4 items stacked → taller
	narrow := layout.Context{
		Ops: &ops, Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Constraints: layout.Constraints{Max: image.Pt(400, 800)},
		Now:         time.Unix(1, 0), Source: rt.Source(),
	}
	dn := g.Layout(th, narrow, Cell(tile), Cell(tile), Cell(tile), Cell(tile))
	ops.Reset()
	wide := narrow
	wide.Constraints.Max.X = 1200
	dw := g.Layout(th, wide, Cell(tile), Cell(tile), Cell(tile), Cell(tile))
	if dn.Size.Y <= dw.Size.Y {
		t.Fatalf("narrow height %d should exceed wide %d", dn.Size.Y, dw.Size.Y)
	}
}

func TestShowHides(t *testing.T) {
	th := NewTheme()
	var ops op.Ops
	rt := new(input.Router)
	gtx := layout.Context{
		Ops: &ops, Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Constraints: layout.Constraints{Max: image.Pt(400, 400)},
		Now:         time.Unix(1, 0), Source: rt.Source(),
	}
	d := Show(th, gtx, Bools(false).At("lg", true), func(gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: image.Pt(50, 50)}
	})
	if d.Size != (image.Point{}) {
		t.Fatalf("want hidden, got %+v", d.Size)
	}
	gtx.Constraints.Max.X = 1000
	d = Show(th, gtx, Bools(false).At("lg", true), func(gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: image.Pt(50, 50)}
	})
	if d.Size.X != 50 {
		t.Fatalf("want shown, got %+v", d.Size)
	}
}

func TestResponsiveSizeDialogWidth(t *testing.T) {
	th := NewTheme()
	m := Dialog{Sizes: Sizes(SizeSM).At("lg", Size2XL)}
	var ops op.Ops
	rt := new(input.Router)
	narrow := layout.Context{
		Ops: &ops, Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Constraints: layout.Constraints{Max: image.Pt(500, 800)},
		Now:         time.Unix(1, 0), Source: rt.Source(),
	}
	if got := m.resolveWidth(th, narrow); got != 400 {
		t.Fatalf("narrow: got %v want 400", got)
	}
	wide := narrow
	wide.Constraints.Max.X = 1200
	if got := m.resolveWidth(th, wide); got != 840 {
		t.Fatalf("wide: got %v want 840", got)
	}
	m2 := Dialog{Widths: Dps(300).At("md", 700)}
	if got := m2.resolveWidth(th, wide); got != 700 {
		t.Fatalf("Widths: got %v want 700", got)
	}
}

func TestResponsiveSizeZeroAlloc(t *testing.T) {
	th := NewTheme()
	r := Sizes(SizeSM).At("md", SizeLG).At("xl", Size2XL)
	var ops op.Ops
	rt := new(input.Router)
	gtx := benchCtx(&ops, rt)
	allocs := testing.AllocsPerRun(1000, func() {
		_ = r.Resolve(th, gtx)
	})
	if allocs != 0 {
		t.Fatalf("ResponsiveSize.Resolve allocates %v; want 0", allocs)
	}
}

func TestBreakpointIndexZeroAlloc(t *testing.T) {
	th := NewTheme()
	var ops op.Ops
	rt := new(input.Router)
	gtx := benchCtx(&ops, rt)
	allocs := testing.AllocsPerRun(1000, func() {
		_ = th.BreakpointIndex(gtx)
	})
	if allocs != 0 {
		t.Fatalf("BreakpointIndex allocates %v; want 0", allocs)
	}
}

func TestResponsiveIntResolveZeroAlloc(t *testing.T) {
	th := NewTheme()
	r := Cols(1).At("md", 2).At("lg", 4)
	var ops op.Ops
	rt := new(input.Router)
	gtx := benchCtx(&ops, rt)
	allocs := testing.AllocsPerRun(1000, func() {
		_ = r.Resolve(th, gtx)
	})
	if allocs != 0 {
		t.Fatalf("Resolve allocates %v; want 0", allocs)
	}
}

func BenchmarkBreakpointIndex(b *testing.B) {
	th := NewTheme()
	var ops op.Ops
	rt := new(input.Router)
	gtx := benchCtx(&ops, rt)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = th.BreakpointIndex(gtx)
	}
}

func BenchmarkResponsiveIntAt(b *testing.B) {
	th := NewTheme()
	r := Cols(1).At("sm", 2).At("md", 3).At("lg", 4)
	var ops op.Ops
	rt := new(input.Router)
	gtx := benchCtx(&ops, rt)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = r.Resolve(th, gtx)
	}
}
