package lotusui

import (
	"image"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
)

// The frame loop is this library's hot path, so the disciplines that
// matter are borrowed from low-latency code: no allocations on
// read-mostly paths, preallocated scratch buffers, and BENCHMARKS that
// pin those properties so they can't regress silently.

func benchCtx(ops *op.Ops, r *input.Router) layout.Context {
	return layout.Context{
		Ops:         ops,
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Constraints: layout.Constraints{Max: image.Pt(800, 600)},
		Now:         time.Unix(1, 0),
		Source:      r.Source(),
	}
}

// TestSVGIconCacheZeroAlloc pins the icon cache's hit path at zero
// allocations: after the first rasterization, looking an icon up must
// cost a map read through an atomic pointer — nothing else. This is
// the contract that makes SVGIcon safe to call for every icon on
// every frame.
func TestSVGIconCacheZeroAlloc(t *testing.T) {
	if _, ok := svgIconOp(IconSettings, 24, DefaultPalette.FgSubtle); !ok {
		t.Skip("icon assets not embedded")
	}
	allocs := testing.AllocsPerRun(1000, func() {
		svgIconOp(IconSettings, 24, DefaultPalette.FgSubtle)
	})
	if allocs != 0 {
		t.Fatalf("icon cache hit allocates %v times per lookup; want 0", allocs)
	}
}

func BenchmarkSVGIconCacheHit(b *testing.B) {
	svgIconOp(IconSettings, 24, DefaultPalette.FgSubtle)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		svgIconOp(IconSettings, 24, DefaultPalette.FgSubtle)
	}
}

// BenchmarkButtonFrame measures one settled button layout — the cost a
// screen pays per button per frame.
func BenchmarkButtonFrame(b *testing.B) {
	th := NewTheme()
	var btn widget.Clickable
	r := new(input.Router)
	var ops op.Ops
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ops.Reset()
		gtx := benchCtx(&ops, r)
		Button(th, &btn, "Save", ButtonProps{})(gtx)
	}
}

// BenchmarkSimpleGridFrame measures a 30-card grid's full frame cost —
// the measure pass reuses one scratch Ops, so this stays flat as the
// item count grows.
func BenchmarkSimpleGridFrame(b *testing.B) {
	th := NewTheme()
	items := make([]int, 30)
	r := new(input.Router)
	var ops op.Ops
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ops.Reset()
		gtx := benchCtx(&ops, r)
		SimpleGrid(th, gtx, items, SimpleGridProps{MinChildWidth: 120, MaxCols: 4, Gap: Space.MD},
			func(gtx layout.Context, _ int) layout.Dimensions {
				return SurfaceCard(th, gtx, LabelBody(th, "card").Layout)
			})
	}
}

// BenchmarkListView10k proves virtualization: laying out a 10,000-row
// list must cost a screenful of rows, not 10,000 — compare with
// BenchmarkScrollable10k, which pays for every row.
func BenchmarkListView10k(b *testing.B) {
	th := NewTheme()
	list := widget.List{List: layout.List{Axis: layout.Vertical}}
	r := new(input.Router)
	var ops op.Ops
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ops.Reset()
		gtx := benchCtx(&ops, r)
		ListView(th, &list, gtx, 10000, func(gtx layout.Context, i int) layout.Dimensions {
			return LabelBody(th, "row content").Layout(gtx)
		})
	}
}

func BenchmarkScrollable10k(b *testing.B) {
	th := NewTheme()
	list := widget.List{List: layout.List{Axis: layout.Vertical}}
	r := new(input.Router)
	var ops op.Ops
	rows := make([]layout.Widget, 10000)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ops.Reset()
		gtx := benchCtx(&ops, r)
		for j := range rows {
			rows[j] = LabelBody(th, "row content").Layout
		}
		Scrollable(th, &list, gtx, VStack(0, rows...))
	}
}
