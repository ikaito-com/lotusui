package lotusui

import (
	"image"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
)

func TestBreadcrumbNavCollapse(t *testing.T) {
	th := NewTheme()
	var home, docs, build, fetch widget.Clickable
	nav := &BreadcrumbNav{}
	segs := BreadcrumbSegs(
		BreadcrumbSegLink(&home, "Home"),
		BreadcrumbSegLink(&docs, "Documentation"),
		BreadcrumbSegLink(&build, "Building Your Application"),
		BreadcrumbSegLink(&fetch, "Data Fetching"),
		BreadcrumbSegOf("Caching and Revalidating"),
	)
	ops := new(op.Ops)
	gtx := layout.Context{
		Ops: ops,
		Constraints: layout.Constraints{
			Max: image.Pt(900, 48),
		},
		Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1},
	}
	dims := nav.Layout(th, gtx, segs...)
	if dims.Size.X <= 0 || dims.Size.Y <= 0 {
		t.Fatalf("expected non-empty trail, got %+v", dims)
	}
	if nav.menu.Align != PopoverStart {
		t.Errorf("ellipsis menu Align = %v, want PopoverStart (shadcn align=start)", nav.menu.Align)
	}
	if nav.menu.Variant != ButtonGhost {
		t.Errorf("ellipsis Variant = %v, want ButtonGhost", nav.menu.Variant)
	}
	if nav.menu.Icon != IconMoreHorizontal {
		t.Errorf("ellipsis Icon = %q, want IconMoreHorizontal", nav.menu.Icon)
	}
}
