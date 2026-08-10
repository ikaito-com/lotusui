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

// testCtx builds a real layout context: live input router, 1:1 metric,
// the given constraints. Enough for any component to lay out exactly as
// it would inside a window.
func testCtx(ops *op.Ops, r *input.Router, cs layout.Constraints) layout.Context {
	return layout.Context{
		Ops:         ops,
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Constraints: cs,
		Now:         time.Unix(1, 0),
		Source:      r.Source(),
	}
}

// TestComponentsSurviveHostileConstraints lays out every component at
// the constraint shapes that break naive layout code — zero area, a
// narrow sliver, a huge fixed area — and requires only that nothing
// panics and nothing reports a size beyond Max. This is the Gio bug
// class (division by column count, negative insets, unbounded Min)
// that unit logic can't catch and screenshots catch too late.
func TestComponentsSurviveHostileConstraints(t *testing.T) {
	th := NewTheme()

	var (
		btn      widget.Clickable
		list     = widget.List{List: layout.List{Axis: layout.Vertical}}
		vlist    = widget.List{List: layout.List{Axis: layout.Vertical}}
		field    Input
		nameF    = Input{Filter: "abcdefghijklmnopqrstuvwxyz-", Error: "taken"}
		check    Checkbox
		sw       Switch
		swOff    = Switch{Disabled: true, Invalid: true}
		encT     = Tabs{Variant: TabsDefault, Options: []TabOption{{Label: "A"}, {Label: "B", Disabled: true}}}
		drop     = Select{Options: SelectOpts("a", "b", "c")}
		tabs     = Tabs{Options: TabOpts("A", "B")}
		modal    Dialog
		split    Split
		vslide   VSlide
		back     widget.Clickable
		iconBtn  widget.Clickable
		hoverBtn widget.Clickable
		cancelB  widget.Clickable
		confirmB widget.Clickable
		tip      Tooltip
		hc       HoverCard
		cmenu    ContextMenu
	)

	components := map[string]func(gtx layout.Context) layout.Dimensions{
		"Fill":             func(gtx layout.Context) layout.Dimensions { return Fill(gtx, th.Palette.Bg) },
		"Button":           Button(th, &btn, "Save", ButtonProps{}),
		"ButtonAll":        Button(th, &btn, "X", ButtonProps{Variant: ButtonDestructive, Size: SizeXL, Loading: true}),
		"DropdownMenuItem": DropdownMenuItem(th, &btn, "Delete", true),
		"Badge":            Badge(th, "3 items", BadgeProps{Variant: BadgeOutline}),
		"BadgeRaw":         Badge(th, "ok", BadgeProps{Bg: th.Palette.SuccessBg, Fg: th.Palette.Success}),
		"Hairline":         Hairline(th),
		"Alert":            Alert(th, AlertProps{Title: "T", Description: "D"}),
		"AlertDestructive": Alert(th, AlertProps{Variant: AlertDestructive, Title: "T"}),
		"Avatar":           Avatar(th, AvatarProps{Initials: "AL"}),
		"AvatarFallback":   Avatar(th, AvatarProps{Size: Size2XL}),
		"Breadcrumb":       Breadcrumb(th, nil, "Home", "Here"),
		"Separator":        Separator(th),
		"Skeleton":         Skeleton(th, 40, 12),
		"Progress":         Progress(th, 0.5),
		"ProgressIndet":    Progress(th, -1),
		"Spinner":          Spinner(th, 20),
		"TableText": TableText(th, TableProps{Caption: "c"},
			[]string{"A", "B"}, [][]string{{"1", "2"}}),
		"VHairline": VerticalHairline(th),
		"Spacer":    Spacer(Space.MD),
		"Section":   SectionLabel(th, "GROUP"),
		"Labels": VStack(Space.XS,
			LabelHero(th, "h").Layout, LabelTitle(th, "t").Layout, LabelCardTitle(th, "c").Layout,
			LabelBody(th, "b").Layout, LabelMeta(th, "m").Layout, LabelCaption(th, "cap").Layout),
		"VStack": VStack(Space.SM, Spacer(Space.SM), Spacer(Space.SM)),
		"HStack": HStack(Space.SM, HSpacer(Space.SM), HSpacer(Space.SM)),
		"Wrap":   Wrap(Space.SM, layout.Middle, HSpacer(Space.SM), HSpacer(Space.SM), HSpacer(Space.SM)),
		"TopBar": TopBar(th, "Screen", SVGIconButton(th, &back, IconExpand, 20, false)),
		"TitleWithIcons": TitleWithIcons(th, "Section",
			SVGIconButton(th, &iconBtn, IconAdd, 20, false)),
		"FullW":    FullWidth(Spacer(Space.SM)),
		"RightAl":  RightAligned(Spacer(Space.SM)),
		"SVGIcon":  SVGIcon(IconSettings, 24, th.Palette.FgSubtle),
		"SVGIconB": SVGIconButton(th, &iconBtn, "settings", 24, true),
		"HoverRow": HoverRow(th, &hoverBtn, true, LabelBody(th, "row").Layout),
		"Item": Item(th, ItemProps{
			Variant: ItemOutline,
			Media:   ItemMedia(th, ItemMediaIcon, SVGIcon(IconMail, 18, th.Palette.Fg)),
			Content: ItemContent(th, ItemTitle(th, "Title"), ItemDescription(th, "Description")),
			Actions: ItemActions(th, Button(th, &btn, "Go", ButtonProps{Size: SizeSM})),
		}),
		"ItemGroup": ItemGroup(th,
			Item(th, ItemProps{Content: ItemTitle(th, "A")}),
			ItemSeparator(th),
			Item(th, ItemProps{Content: ItemTitle(th, "B")}),
		),
		"DropdownMenu": DropdownMenu(th,
			DropdownMenuItem(th, &cancelB, "Duplicate", false),
			DropdownMenuItem(th, &confirmB, "Delete", true)),
		"ContextMenu": func(gtx layout.Context) layout.Dimensions {
			return cmenu.Layout(th, gtx, LabelBody(th, "area").Layout,
				ContextMenuItem(th, &cancelB, "Back", false),
				ContextMenuShortcutItem(th, &confirmB, "Reload", "⌘R", false))
		},
		"Tooltip": func(gtx layout.Context) layout.Dimensions {
			return tip.Layout(th, gtx, "hint", Button(th, &btn, "Go", ButtonProps{}))
		},
		"HoverCard": func(gtx layout.Context) layout.Dimensions {
			return hc.Layout(th, gtx, LabelBody(th, "body").Layout, Button(th, &btn, "Hover", ButtonProps{Variant: ButtonLink}))
		},
		"AnnotatedText": AnnotatedText(th, "API and SLA",
			[]GlossaryTerm{{Term: "API", Tip: "t"}, {Term: "SLA", Tip: "t"}},
			[]*HoverCard{&hc, nil}),

		"LayoutPage": func(gtx layout.Context) layout.Dimensions {
			return LayoutPage(th, gtx, Spacer(Space.SM))
		},
		"Scrollable": func(gtx layout.Context) layout.Dimensions {
			return Scrollable(th, &list, gtx, Spacer(Space.SM))
		},
		"ScrollArea": func(gtx layout.Context) layout.Dimensions {
			var sa ScrollArea
			return sa.Layout(th, gtx, Spacer(Space.SM))
		},
		"Scrollbar": func(gtx layout.Context) layout.Dimensions {
			var bar Scrollbar
			d, _ := bar.Layout(th, gtx, ScrollbarProps{Variant: ScrollbarAlways}, 0.1, 0.4)
			return d
		},
		"CodeBlock": CodeBlock(th, CodeBlockProps{Lang: "go", Plain: "x := 1"}),
		"Example": func(gtx layout.Context) layout.Dimensions {
			var ex Example
			return ex.Layout(th, gtx, ExampleProps{
				Preview: LabelBody(th, "preview").Layout,
				Code:    CodeBlock(th, CodeBlockProps{Nested: true, Plain: "x"}),
			})
		},
		"ListView": func(gtx layout.Context) layout.Dimensions {
			return ListView(th, &vlist, gtx, 1000, func(gtx layout.Context, i int) layout.Dimensions {
				return LabelBody(th, "row").Layout(gtx)
			})
		},
		"SurfaceCard": func(gtx layout.Context) layout.Dimensions {
			return SurfaceCard(th, gtx, LabelBody(th, "content").Layout)
		},
		"SplitColumnScroll": func(gtx layout.Context) layout.Dimensions {
			return SplitColumnScroll(th, &list, 120, SplitBox(th, LabelBody(th, "col").Layout))(gtx)
		},
		"SplitBoxScroll": func(gtx layout.Context) layout.Dimensions {
			return SplitBoxScroll(th, &vlist, 120, LabelBody(th, "pane").Layout)(gtx)
		},
		"SplitBoxFillScroll": func(gtx layout.Context) layout.Dimensions {
			return SplitBoxFillScroll(th, &list, 120, LabelBody(th, "fill").Layout)(gtx)
		},
		"FloatingPanel": func(gtx layout.Context) layout.Dimensions {
			return FloatingPanel(th, gtx, Spacer(Space.SM))
		},
		"Grid": func(gtx layout.Context) layout.Dimensions {
			return Grid{Columns: 4, Gap: Space.SM}.Layout(th, gtx,
				GridItem{RowSpan: 2, W: Spacer(Space.MD)},
				GridItem{ColSpan: 2, W: Spacer(Space.SM)},
				Cell(Spacer(Space.SM)), Cell(Spacer(Space.SM)),
				GridItem{ColSpan: 4, W: Spacer(Space.SM)})
		},
		"CardVariants": VStack(Space.SM,
			Card(th, CardProps{Variant: CardOutline}, Spacer(Space.SM)),
			Card(th, CardProps{Variant: CardElevated, Size: SizeSM}, Spacer(Space.SM)),
			Card(th, CardProps{Variant: CardSubtle, Size: SizeLG}, Spacer(Space.SM))),
		"SimpleGrid": func(gtx layout.Context) layout.Dimensions {
			return SimpleGrid(th, gtx, []int{1, 2, 3, 4, 5}, SimpleGridProps{
				MinChildWidth: 120, MaxCols: 3, Gap: Space.MD,
			}, func(gtx layout.Context, _ int) layout.Dimensions {
				return SurfaceCard(th, gtx, LabelBody(th, "card").Layout)
			})
		},
		"TextField": func(gtx layout.Context) layout.Dimensions {
			return field.Layout(th, gtx, "Label", "hint")
		},
		"TextFieldSuffix": func(gtx layout.Context) layout.Dimensions {
			return nameF.LayoutSuffix(th, gtx, "Name", "hint", "-dev")
		},
		"Checkbox": func(gtx layout.Context) layout.Dimensions { return check.Layout(th, gtx, "On") },
		"Switch":   func(gtx layout.Context) layout.Dimensions { return sw.Layout(th, gtx) },
		"SwitchStates": func(gtx layout.Context) layout.Dimensions {
			return swOff.Layout(th, gtx)
		},
		"TabsDefault": func(gtx layout.Context) layout.Dimensions {
			encT.Update(gtx)
			return encT.Layout(th, gtx)
		},
		"Select": func(gtx layout.Context) layout.Dimensions { return drop.Layout(th, gtx, "Env") },
		"Tabs": func(gtx layout.Context) layout.Dimensions {
			tabs.Update(gtx)
			return tabs.Layout(th, gtx)
		},
		"Dialog": func(gtx layout.Context) layout.Dimensions {
			return modal.Layout(th, gtx, nil, LabelBody(th, "body").Layout)
		},
		"Split": func(gtx layout.Context) layout.Dimensions {
			return split.Layout(gtx, Space.MD, 1,
				SplitBox(th, LabelBody(th, "a").Layout),
				SplitBox(th, LabelBody(th, "b").Layout))
		},
		"VSlide": func(gtx layout.Context) layout.Dimensions {
			return vslide.Layout(gtx, th, true, LabelBody(th, "base").Layout, LabelBody(th, "over").Layout)
		},
		"ButtonGroup": func(gtx layout.Context) layout.Dimensions {
			return ButtonGroup(th, ButtonGroupProps{},
				ButtonGroupItem{Btn: &back, Label: "Copy", Props: ButtonProps{Variant: ButtonSecondary}},
				ButtonGroupSeparator(),
				ButtonGroupItem{Btn: &iconBtn, Label: "Paste", Props: ButtonProps{Variant: ButtonSecondary}},
				ButtonGroupInput(&field, "Search...", 1),
			)(gtx)
		},
		"InputOTP": func(gtx layout.Context) layout.Dimensions {
			var code InputOTP
			code.Groups = []int{2, 2, 2}
			return code.Layout(th, gtx)
		},
		"ButtonGroupVertical": func(gtx layout.Context) layout.Dimensions {
			return ButtonGroup(th, ButtonGroupProps{Vertical: true},
				ButtonGroupItem{Btn: &hoverBtn, Label: "Up", Props: ButtonProps{Variant: ButtonOutline}},
				ButtonGroupItem{Btn: &cancelB, Label: "Down", Props: ButtonProps{Variant: ButtonOutline}},
			)(gtx)
		},
	}

	shapes := map[string]layout.Constraints{
		"zero":   {},
		"sliver": {Max: image.Pt(3, 2000)},
		"phone":  {Max: image.Pt(360, 640)},
		"window": {Min: image.Pt(1200, 800), Max: image.Pt(1200, 800)},
		"huge":   {Max: image.Pt(100000, 100000)},
	}

	for cname, w := range components {
		for sname, cs := range shapes {
			t.Run(cname+"/"+sname, func(t *testing.T) {
				defer func() {
					if r := recover(); r != nil {
						t.Fatalf("panicked: %v", r)
					}
				}()
				r := new(input.Router)
				var ops op.Ops
				// Two frames: the second exercises state carried over
				// (click handling, animation clocks, cached buttons).
				for i := 0; i < 2; i++ {
					ops.Reset()
					gtx := testCtx(&ops, r, cs)
					d := w(gtx)
					if d.Size.X > cs.Max.X || d.Size.Y > cs.Max.Y {
						t.Fatalf("reported %v beyond max %v", d.Size, cs.Max)
					}
					r.Frame(&ops)
				}
			})
		}
	}
}
