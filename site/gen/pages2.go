package main

// The second wave of component pages — same anatomy as the first:
// install → usage → one section per capability, demos index-aligned
// with site/gallery/demos2.go.

func accordionPage() *page {
	return &page{
		Slug:   "accordion",
		Title:  "Accordion",
		Kicker: "Stacked disclosures: one panel at a time, animated on the shared clock.",
		Intro: `<p>Titles that expand one panel of content at a time (<code>Multiple</code> for
independent panels). The reveal grows on the shared clock — content is measured at natural
height and clipped, never reflowed mid-flight.</p>`,
		Sections: []section{
			installSection("accordion"),
			{
				Heading: "Usage",
				Snippet: `var acc lotusui.Accordion // zero value: all closed, single-open

acc.Layout(th, gtx,
	lotusui.AccordionItem{Title: "Is it accessible?", Content: answer1},
	lotusui.AccordionItem{Title: "Is it styled?", Content: answer2},
)`,
				Demo:  "accordion/0",
				DemoH: 260,
			},
			{
				Heading: "Multiple",
				Prose:   `<p><code>Multiple</code> lets panels expand independently.</p>`,
				Snippet: `acc := lotusui.Accordion{Multiple: true}`,
				Demo:    "accordion/1",
				DemoH:   280,
			},
			{
				Heading: "Disabled item",
				Snippet: `lotusui.AccordionItem{Title: "Premium feature information", Disabled: true, Content: c}`,
				Demo:    "accordion/2",
				DemoH:   260,
			},
			{
				Heading: "Borders",
				Prose:   `<p><code>Bordered</code> wraps the accordion in a rounded outline with side padding.</p>`,
				Snippet: `acc := lotusui.Accordion{Bordered: true}`,
				Demo:    "accordion/3",
				DemoH:   220,
			},
			{
				Heading: "Card",
				Prose:   `<p>Inside a padded surface — composition with <code>SurfaceCard</code>.</p>`,
				Snippet: `lotusui.SurfaceCard(th, gtx, acc.Layout(…))`,
				Demo:    "accordion/4",
				DemoH:   240,
			},
		},
		Props: []prop{
			{"Open", "int", "The expanded item in single mode; -1 = all closed."},
			{"Multiple", "bool", "Independent panels, tracked in Expanded."},
			{"Expanded", "[]bool", "Per-item state in Multiple mode."},
		},
	}
}

func alertPage() *page {
	return &page{
		Slug:   "alert",
		Title:  "Alert",
		Kicker: "The static callout: icon, title, description — it informs, never acts.",
		Intro: `<p>A bordered, tinted box with an icon, a title and an optional description. For
interruptions that require action, use <a href="../alert-dialog/">AlertDialog</a>.</p>`,
		Sections: []section{
			installSection("alert"),
			{
				Heading: "Usage",
				Snippet: `lotusui.Alert(th, lotusui.AlertProps{
	Title:       "Heads up!",
	Description: "You can vendor any component into your app with the CLI.",
})`,
				Demo:  "alert/0",
				DemoH: 130,
			},
			{
				Heading: "Destructive",
				Prose:   `<p>Danger ink on the danger tint; the icon defaults to the warning glyph.</p>`,
				Snippet: `lotusui.Alert(th, lotusui.AlertProps{
	Variant: lotusui.AlertDestructive,
	Title:   "Your session has expired",
})`,
				Demo:  "alert/1",
				DemoH: 130,
			},
			{
				Heading: "With action",
				Prose:   `<p><code>Action</code> renders at the alert's end — typically a small button.</p>`,
				Snippet: `lotusui.Alert(th, lotusui.AlertProps{
	Title:       "Dark mode is now available",
	Description: "Enable it under your profile settings to get started.",
	Action:      lotusui.Button(th, &enable, "Enable", lotusui.ButtonProps{Size: lotusui.SizeXS}),
})`,
				Demo:  "alert/2",
				DemoH: 130,
			},
			{
				Heading: "Colors",
				Prose: `<p><code>Color</code> re-tints the whole alert from any <code>ColorScale</code>, the
pastel way — tinted well, deep same-hue ink, readability preserved by construction.</p>`,
				Snippet: `lotusui.Alert(th, lotusui.AlertProps{
	Color: lotusui.Orange,
	Title: "Your subscription will expire in 3 days.",
})`,
				Demo:  "alert/3",
				DemoH: 150,
			},
		},
		Props: []prop{
			{"Variant", "AlertVariant", "AlertDefault, AlertDestructive."},
			{"Icon", "string", "Overrides the variant's default (info / warning)."},
			{"Title / Description", "string", "Description optional."},
			{"Action", "layout.Widget", "Rendered at the end — typically a small Button."},
			{"Color", "ColorScale", "Re-tints the alert the pastel way from any scale."},
		},
	}
}

func alertDialogPage() *page {
	return &page{
		Slug:   "alert-dialog",
		Title:  "AlertDialog",
		Kicker: "The interruption that requires a decision — no backdrop dismissal.",
		Intro: `<p>A Dialog that absorbs every outside click and Escape and offers exactly Cancel
and Action. Poll <code>Confirmed</code>/<code>Cancelled</code> each frame while open; lay it
out at window constraints like Dialog.</p>`,
		Sections: []section{
			installSection("alert-dialog"),
			{
				Heading: "Usage",
				Snippet: `var confirm lotusui.AlertDialog
var open bool

if deleteBtn.Clicked(gtx) { open = true; confirm.Appear() }
if open {
	if confirm.Confirmed(gtx) { open = false; doDelete() }
	if confirm.Cancelled(gtx) { open = false }
	confirm.Layout(th, gtx, lotusui.AlertDialogProps{
		Title:       "Are you absolutely sure?",
		Description: "This action cannot be undone. This will permanently delete your account.",
	})
}`,
				Demo:  "alert-dialog/0",
				DemoH: 420,
			},
			{
				Heading: "Destructive",
				Prose: `<p><code>Destructive: true</code> turns the action red; <code>Media</code> renders a
medallion above the title; <code>Size</code> picks the width preset.</p>`,
				Snippet: `confirm.Layout(th, gtx, lotusui.AlertDialogProps{
	Size: lotusui.SizeSM, Title: "Delete chat?",
	Description: "This will permanently delete this chat conversation.",
	Action: "Delete", Destructive: true, Media: dangerMedallion,
})`,
				Demo:  "alert-dialog/1",
				DemoH: 460,
			},
			{
				Heading: "Small",
				Snippet: `confirm.Layout(th, gtx, lotusui.AlertDialogProps{
	Size: lotusui.SizeSM, Title: "Allow accessory to connect?",
	Description: "Do you want to allow the USB accessory to connect to this device?",
	Cancel: "Don't allow", Action: "Allow",
})`,
				Demo:  "alert-dialog/2",
				DemoH: 400,
			},
		},
		Props: []prop{
			{"Appear()", "method", "Restart the entrance animation on the closed→open transition."},
			{"Confirmed(gtx) / Cancelled(gtx)", "bool", "Poll each frame while open."},
			{"Title / Description", "string", "The decision's copy."},
			{"Cancel / Action", "string", "Button labels; empty = Cancel / Continue."},
			{"Size", "Size", "Dialog width preset (Size2XS…Size2XL)."},
			{"Sizes", "ResponsiveSize", "Stepped Size; when Set(), overrides Size."},
			{"Width / Widths", "unit.Dp / ResponsiveDp", "Free-form or stepped width; Widths wins when Set()."},
			{"Media", "layout.Widget", "Renders above the title — an icon medallion, an image."},
			{"Destructive", "bool", "Renders the action destructively."},
		},
	}
}

func annotatedTextPage() *page {
	return &page{
		Slug:   "annotated-text",
		Title:  "AnnotatedText",
		Kicker: "In-text glossary terms with HoverCard tips — BrandFg ink, stay-open panels.",
		Intro: `<p>A lotusui extension: highlight terms in body copy and preview a tip on hover.
<code>SplitGlossary</code> segments the string (longest match wins);
<code>AnnotatedText</code> paints terms in <code>BrandFg</code> and opens a
<a href="../hover-card/">HoverCard</a> when a card identity is provided.
Segments flow with <a href="../wrap/">Wrap</a> (gap 0) so a narrow Max.X never
squeezes labels into one-character columns. <code>Tooltip</code> is the wrong tool —
inverted string tips do not stay open on the panel.
HoverCard defaults stay max-320dp / 700ms (card hugs content); compact UIs set <code>Width</code>, <code>OpenDelay</code>,
and <code>Side</code> on each card.</p>`,
		Sections: []section{
			installSection("annotated-text"),
			{
				Heading: "Usage",
				Prose:   `<p>Neutral glossary terms (API, SLA) — one <code>HoverCard</code> per term identity; multi-site Layout covers repeats.</p>`,
				Snippet: `terms := []lotusui.GlossaryTerm{
	{Term: "API", Tip: "Application Programming Interface"},
	{Term: "SLA", Tip: "Service Level Agreement"},
}
var api, sla lotusui.HoverCard
api.Width, sla.Width = 200, 200
api.OpenDelay, sla.OpenDelay = 400*time.Millisecond, 400*time.Millisecond
api.Side, sla.Side = lotusui.HoverCardTop, lotusui.HoverCardTop

lotusui.AnnotatedText(th,
	"The API must meet the SLA under load.",
	terms, []*lotusui.HoverCard{&api, &sla})`,
				Demo:  "annotated-text/0",
				DemoH: 200,
			},
			{
				Heading: "SplitGlossary",
				Prose:   `<p>Pure segmentation — longest literal match wins when spans collide.</p>`,
				Snippet: `lotusui.SplitGlossary(text, terms)`,
			},
		},
		Props: []prop{
			{"GlossaryTerm", "struct", "Term (literal) + Tip (Caption in the card)."},
			{"SplitGlossary", "func", "[]GlossarySeg — plain vs term runs."},
			{"AnnotatedText", "widget", "BrandFg terms; cards[i] for terms[i] (nil / short slice = ink only)."},
		},
	}
}

func avatarPage() *page {
	return &page{
		Slug:   "avatar",
		Title:  "Avatar",
		Kicker: "The circular identity mark: initials, a fallback glyph, any scale's tint.",
		Intro: `<p>Initials centered on a tinted circle; empty initials fall back to the person
glyph. <code>Color</code> re-tints the pastel way from any scale.</p>`,
		Sections: []section{
			installSection("avatar"),
			{
				Heading: "Usage",
				Snippet: `lotusui.Avatar(th, lotusui.AvatarProps{Initials: "AL"})
lotusui.Avatar(th, lotusui.AvatarProps{Initials: "KB", Color: lotusui.Teal})`,
				Demo:  "avatar/0",
				DemoH: 110,
			},
			{
				Heading: "Sizes",
				Snippet: `lotusui.Avatar(th, lotusui.AvatarProps{Initials: "A", Size: lotusui.Size2XL})`,
				Demo:    "avatar/1",
				DemoH:   120,
			},
			{
				Heading: "Group",
				Prose:   `<p><code>AvatarGroup</code> overlaps right-over-left, each avatar ringed in the panel background.</p>`,
				Snippet: `lotusui.AvatarGroup(th, lotusui.AvatarGroupProps{},
	lotusui.AvatarProps{Initials: "CN"},
	lotusui.AvatarProps{Initials: "LR", Color: lotusui.Teal},
	lotusui.AvatarProps{Initials: "ER", Color: lotusui.Pink},
)`,
				Demo:  "avatar/2",
				DemoH: 110,
			},
			{
				Heading: "Badge",
				Prose:   `<p><code>Badge</code> pins a status dot to the bottom-right rim — a color anchor, or an icon inside.</p>`,
				Snippet: `lotusui.Avatar(th, lotusui.AvatarProps{Initials: "CN",
	Badge: &lotusui.AvatarBadge{Color: lotusui.Green}})
lotusui.Avatar(th, lotusui.AvatarProps{Initials: "PP",
	Badge: &lotusui.AvatarBadge{Icon: lotusui.IconPlus}})`,
				Demo:  "avatar/3",
				DemoH: 110,
			},
			{
				Heading: "Group with count",
				Prose:   `<p><code>Count</code> (or <code>CountIcon</code>) closes the group with a "+N" bubble.</p>`,
				Snippet: `lotusui.AvatarGroup(th, lotusui.AvatarGroupProps{Count: "+3"}, a1, a2, a3)`,
				Demo:    "avatar/4",
				DemoH:   110,
			},
		},
		Props: []prop{
			{"Initials", "string", "Centered in the circle; empty falls back to the person glyph."},
			{"Size", "Size", "The shared size presets (Size2XS–Size2XL)."},
			{"Color", "ColorScale", "Pastel re-tint from any scale."},
			{"Scheme", "*Scheme", "Wins over Color: full manual slot control."},
			{"Badge", "*AvatarBadge", "Status dot on the rim: Color anchors the fill, Icon renders inside."},
			{"AvatarGroup(th, opts, avatars…)", "widget", "Ringed overlap; Count/CountIcon add the trailing bubble."},
		},
	}
}

func breadcrumbPage() *page {
	return &page{
		Slug:   "breadcrumb",
		Title:  "Breadcrumb",
		Kicker: "The path to the current page: muted ancestors, chevrons, current in ink.",
		Intro: `<p>Ancestor labels in muted ink separated by chevrons; the last label is the current
page and never interactive. Pass clickables to make ancestors navigate. The shadcn demo
collapses middle segments behind an ellipsis <code>DropdownMenuTrigger</code> (ghost icon-only,
<code>PopoverStart</code>). For automatic collapse use <code>BreadcrumbNav</code> — same
<code>ITEMS_TO_DISPLAY</code> rule as shadcn's responsive example.</p>`,
		Sections: []section{
			installSection("breadcrumb"),
			{
				Heading: "Usage",
				Prose: `<p>The primary demo mirrors shadcn's: Home, an ellipsis menu
(Documentation / Themes / GitHub), Components, Breadcrumb. Align the menu with
<code>PopoverStart</code> so it opens under the dots, not centered.</p>`,
				Snippet: `var menu = lotusui.DropdownMenuTrigger{
	Variant: lotusui.ButtonGhost, Icon: lotusui.IconMoreHorizontal,
	Size: lotusui.SizeSM, Width: 160, Align: lotusui.PopoverStart,
}
menu.Layout(th, gtx, "",
	lotusui.DropdownMenuItem(th, &docs, "Documentation", false),
	lotusui.DropdownMenuItem(th, &themes, "Themes", false),
	lotusui.DropdownMenuItem(th, &github, "GitHub", false),
)`,
				Demo:  "breadcrumb/0",
				DemoH: 280,
			},
			{
				Heading: "Custom separator",
				Prose: `<p>The composable pieces — <code>BreadcrumbLink</code>, <code>BreadcrumbPage</code>,
<code>BreadcrumbSep</code> — build custom trails; pass any icon as the separator.</p>`,
				Snippet: `layout.Flex{Alignment: layout.Middle}.Layout(gtx,
	layout.Rigid(lotusui.BreadcrumbLink(th, &home, "Home")),
	layout.Rigid(lotusui.BreadcrumbSep(th, lotusui.IconDot)),
	layout.Rigid(lotusui.BreadcrumbPage(th, "Breadcrumb")),
)`,
				Demo:  "breadcrumb/1",
				DemoH: 100,
			},
			{
				Heading: "Dropdown",
				Prose:   `<p>A trail label can open a menu — ghost <code>DropdownMenuTrigger</code> in the row (shadcn's "Dropdown" example).</p>`,
				Snippet: `var menu = lotusui.DropdownMenuTrigger{Variant: lotusui.ButtonGhost, Width: 160, Align: lotusui.PopoverStart}
menu.Layout(th, gtx, "Components",
	lotusui.DropdownMenuItem(th, &docs, "Documentation", false),
	lotusui.DropdownMenuItem(th, &themes, "Themes", false),
)`,
				Demo:  "breadcrumb/2",
				DemoH: 280,
			},
			{
				Heading: "Collapsed",
				Prose:   `<p><code>BreadcrumbEllipsis</code> is the bare collapsed-depth mark (shadcn's "Collapsed"). Wire it to a menu as in Usage when the dots should reveal the hidden trail — or use <code>BreadcrumbNav</code>.</p>`,
				Snippet: `lotusui.BreadcrumbEllipsis(th)`,
				Demo:    "breadcrumb/3",
				DemoH:   100,
			},
			{
				Heading: "Responsive",
				Prose: `<p><code>BreadcrumbNav</code> mirrors shadcn's responsive example: when the path is longer
than <code>ItemsToDisplay</code> (default 3), the middle collapses into an ellipsis menu
(<code>items.slice(1, -2)</code>). Below the theme <code>md</code> breakpoint the trail shows
first+last only and labels truncate (<code>max-w-20</code>). Mobile Drawer is "from the web";
the ellipsis still opens a dropdown at every width.</p>`,
				Snippet: `var nav lotusui.BreadcrumbNav
nav.Layout(th, gtx,
	lotusui.BreadcrumbSegLink(&home, "Home"),
	lotusui.BreadcrumbSegLink(&docs, "Documentation"),
	lotusui.BreadcrumbSegLink(&build, "Building Your Application"),
	lotusui.BreadcrumbSegLink(&fetch, "Data Fetching"),
	lotusui.BreadcrumbSegOf("Caching and Revalidating"),
)`,
				Demo:  "breadcrumb/4",
				DemoH: 320,
			},
		},
		Props: []prop{
			{"btns", "[]*widget.Clickable", "Index-aligned; makes ancestor labels clickable, nil = static."},
			{"labels", "...string", "The path; the LAST label is the current page, never interactive."},
			{"BreadcrumbNav", "struct", "Responsive trail: ItemsToDisplay (default 3), auto ellipsis menu, md truncate."},
			{"BreadcrumbSeg / SegLink / SegOf / Segs", "ctors", "Build-time segments for BreadcrumbNav."},
			{"BreadcrumbLink / BreadcrumbPage", "widget", "The composable pieces: clickable ancestor; current page."},
			{"BreadcrumbSep(th, icon)", "widget", "Between-items glyph; empty icon = the chevron."},
			{"BreadcrumbEllipsis(th)", "widget", "Bare collapsed-depth mark — prefer BreadcrumbNav or DropdownMenuTrigger{Icon, Align: PopoverStart} for interactive dots."},
		},
	}
}

func buttonGroupPage() *page {
	return &page{
		Slug:   "button-group",
		Title:  "ButtonGroup",
		Kicker: "Attached controls: one shared border, square inner corners.",
		Intro: `<p>Children sit edge-to-edge — neighbors overlap by 1dp so their borders collapse
to a single line, every child stretches to one shared height, inner corners render square, and
only the group's outer corners keep the radius. Slots hold buttons, a separator seam, an
<code>Input</code>/<code>Select</code> (auto-attached), or any widget with flex weight.</p>`,
		Sections: []section{
			installSection("button-group"),
			{
				Heading: "Usage",
				Snippet: `lotusui.ButtonGroup(th, lotusui.ButtonGroupProps{},
	lotusui.ButtonGroupItem{Btn: &archive, Label: "Archive", Props: lotusui.ButtonProps{Variant: lotusui.ButtonOutline}},
	lotusui.ButtonGroupItem{Btn: &report, Label: "Report", Props: lotusui.ButtonProps{Variant: lotusui.ButtonOutline}},
	lotusui.ButtonGroupItem{Btn: &snooze, Label: "Snooze", Props: lotusui.ButtonProps{Variant: lotusui.ButtonOutline}},
)`,
				Demo:  "button-group/0",
				DemoH: 280,
			},
			{
				Heading: "Orientation",
				Prose:   `<p><code>Vertical</code> stacks the group top-to-bottom.</p>`,
				Snippet: `lotusui.ButtonGroup(th, lotusui.ButtonGroupProps{Vertical: true}, plus, minus)`,
				Demo:    "button-group/1",
				DemoH:   140,
			},
			{
				Heading: "Sizes",
				Prose:   `<p>Size the items — the group follows.</p>`,
				Snippet: `lotusui.ButtonGroupItem{Btn: &b, Label: "Small",
	Props: lotusui.ButtonProps{Variant: lotusui.ButtonOutline, Size: lotusui.SizeSM}}`,
				Demo:  "button-group/2",
				DemoH: 220,
			},
			{
				Heading: "Separator",
				Prose:   `<p><code>ButtonGroupSeparator()</code> renders the hairline seam between attached neighbors.</p>`,
				Snippet: `lotusui.ButtonGroup(th, lotusui.ButtonGroupProps{},
	copyItem, lotusui.ButtonGroupSeparator(), pasteItem)`,
				Demo:  "button-group/3",
				DemoH: 110,
			},
			{
				Heading: "Split",
				Prose:   `<p>The split button: a main action and an attached icon action across a seam.</p>`,
				Snippet: `lotusui.ButtonGroup(th, lotusui.ButtonGroupProps{},
	mainItem, lotusui.ButtonGroupSeparator(), iconItem)`,
				Demo:  "button-group/4",
				DemoH: 110,
			},
			{
				Heading: "With input",
				Prose:   `<p>A flexed <code>Input</code> slot fuses with neighboring buttons — shared height, square inner corners, one collapsed border (shadcn ButtonGroup + Input).</p>`,
				Snippet: `lotusui.ButtonGroup(th, lotusui.ButtonGroupProps{},
	lotusui.ButtonGroupInput(&search, "Search...", 1),
	lotusui.ButtonGroupItem{Btn: &go_, Props: lotusui.ButtonProps{Variant: lotusui.ButtonOutline, IconStart: lotusui.IconSearch}},
)`,
				Demo:  "button-group/5",
				DemoH: 120,
			},
		},
		Props: []prop{
			{"ButtonGroupProps.Vertical", "bool", "Stack the group top-to-bottom."},
			{"ButtonGroupItem.Btn / Label / Props", "…", "A button slot; the group sets Props.Attached."},
			{"ButtonGroupItem.Input / Hint / Flex", "*Input / string / float32", "A field slot; the group sets Input.Attached and stretches height."},
			{"ButtonGroupItem.Separator", "bool", "The hairline seam (ButtonGroupSeparator())."},
			{"ButtonGroupItem.Widget / Flex", "layout.Widget / float32", "Arbitrary content (no auto-attach); prefer Input for fields."},
			{"ButtonProps.Attached / Input.Attached", "AttachedEdges", "Squares corners on attached sides — set by the group."},
		},
	}
}

func inputOTPPage() *page {
	return &page{
		Slug:   "input-otp",
		Title:  "InputOTP",
		Kicker: "The one-time-code input: attached character slots, one hidden editor.",
		Intro: `<p>Click anywhere on the row, type or paste — the code fills left to right, and the
active slot carries the focus ring. One editor drives every slot, so paste and backspace
behave exactly like a text field.</p>`,
		Sections: []section{
			installSection("input-otp"),
			{
				Heading: "Usage",
				Snippet: `var code lotusui.InputOTP // six slots by default

code.Layout(th, gtx)
value := code.Value()`,
				Demo:  "input-otp/0",
				DemoH: 110,
			},
			{
				Heading: "Separator",
				Prose:   `<p><code>Groups</code> splits the slots with a dash between groups.</p>`,
				Snippet: `code := lotusui.InputOTP{Groups: []int{2, 2, 2}}`,
				Demo:    "input-otp/1",
				DemoH:   110,
			},
			{
				Heading: "Length",
				Snippet: `code := lotusui.InputOTP{Length: 4}`,
				Demo:    "input-otp/2",
				DemoH:   110,
			},
			{
				Heading: "Pattern",
				Prose:   `<p><code>Filter</code> is the allow-list — digits only for numeric codes.</p>`,
				Snippet: `code := lotusui.InputOTP{Filter: "0123456789"}`,
				Demo:    "input-otp/3",
				DemoH:   140,
			},
			{
				Heading: "Disabled",
				Snippet: `code.Disabled = true`,
				Demo:    "input-otp/4",
				DemoH:   110,
			},
			{
				Heading: "Invalid",
				Prose:   `<p><code>Invalid</code> renders danger borders; pair it with a Field error.</p>`,
				Snippet: `code.Invalid = true`,
				Demo:    "input-otp/5",
				DemoH:   140,
			},
		},
		Props: []prop{
			{"Length", "int", "Slot count; zero means 6."},
			{"Groups", "[]int", "Slots per group, a dash between groups."},
			{"Filter", "string", "Allow-list; empty accepts everything."},
			{"Disabled / Invalid", "bool", "Dimmed and inert; danger borders."},
			{"Value() / SetValue(s)", "method", "The code typed so far."},
		},
	}
}

func paginationPage() *page {
	return &page{
		Slug:   "pagination",
		Title:  "Pagination",
		Kicker: "Previous, next, numbered pages — long ranges elide around the current one.",
		Intro: `<p><code>Page</code> is 1-based and lives on the struct; clicks are processed at the
top of Layout, so the click's own frame renders the new page.</p>`,
		Sections: []section{
			installSection("pagination"),
			{
				Heading: "Usage",
				Snippet: `var pg = lotusui.Pagination{Page: 1}
pg.Layout(th, gtx, 12) // 12 pages total
current := pg.Page`,
				Demo:  "pagination/0",
				DemoH: 110,
			},
			{
				Heading: "Simple",
				Prose:   `<p><code>Simple</code> renders numbered links only — no previous/next, no elision.</p>`,
				Snippet: `pg := lotusui.Pagination{Simple: true}`,
				Demo:    "pagination/1",
				DemoH:   110,
			},
			{
				Heading: "Rows per page",
				Prose:   `<p>The table toolbar: a Select for page size beside the pager — composition.</p>`,
				Snippet: `rows := lotusui.Select{Options: lotusui.SelectOpts("10", "25", "50", "100")}`,
				Demo:    "pagination/2",
				DemoH:   320,
			},
		},
		Props: []prop{
			{"Page", "int", "1-based current page — yours; clicks processed at the top of Layout."},
			{"total", "int", "Layout parameter — total page count; long ranges elide."},
		},
	}
}

func hoverCardPage() *page {
	return &page{
		Slug:   "hover-card",
		Title:  "HoverCard",
		Kicker: "Sighted preview of what's behind a link — a floating card on hover.",
		Intro: `<p>Rest the pointer on a trigger and a content card floats in on the portal
layer. Unlike <a href="../tooltip/">Tooltip</a> (inverted chrome, a string label), HoverCard is a
<code>BgPanel</code> card with arbitrary content and stays open while the pointer is over the
card itself. Web composition (<code>HoverCardTrigger</code>/<code>HoverCardContent</code>) is
"from the web": Go's <code>Layout(th, gtx, content, trigger)</code> wraps both. RTL layout is
not yet supported by the underlying toolkit.</p>`,
		Sections: []section{
			installSection("hover-card"),
			{
				Heading: "Usage",
				Snippet: `var tip lotusui.HoverCard
tip.Layout(th, gtx, cardBody,
	lotusui.Button(th, &btn, "Hover", lotusui.ButtonProps{Variant: lotusui.ButtonLink}))`,
				Demo:  "hover-card/0",
				DemoH: 360,
			},
			{
				Heading: "Composition",
				Prose: `<p>On the web the family splits into <code>HoverCard</code> /
<code>HoverCardTrigger</code> / <code>HoverCardContent</code>. In Go one struct owns the state;
<code>Layout</code> takes the card body and the trigger as widgets.</p>`,
			},
			{
				Heading: "Trigger delays",
				Prose: `<p><code>OpenDelay</code> and <code>CloseDelay</code> control when the card
opens and closes. Zero keeps the Radix defaults (700ms open, 300ms close) — long enough not to
flicker, short enough that the pointer can travel the gap onto the card.</p>`,
				Snippet: `tip := lotusui.HoverCard{
	OpenDelay:  100 * time.Millisecond,
	CloseDelay: 200 * time.Millisecond,
}`,
				Demo:  "hover-card/1",
				DemoH: 360,
			},
			{
				Heading: "Positioning",
				Prose: `<p><code>Side</code> floats the card above, below, or beside the trigger;
<code>Align</code> defaults to <code>PopoverCenter</code> (shadcn). Use
<code>PopoverStart</code> / <code>End</code> to slide it along that side.</p>`,
				Snippet: `tip := lotusui.HoverCard{
	Side:  lotusui.HoverCardTop,
	Align: lotusui.PopoverStart,
}`,
				Demo:  "hover-card/2",
				DemoH: 380,
			},
			{
				Heading: "Basic",
				Prose:   `<p>The shadcn profile preview: a link trigger and a card with handle, blurb, and join date.</p>`,
				Snippet: `tip.Layout(th, gtx,
	lotusui.VStack(th.Space.XS,
		lotusui.LabelBody(th, "@nextjs").Layout,
		lotusui.LabelMeta(th, "The React Framework – created and maintained by @vercel.").Layout,
		lotusui.LabelCaption(th, "Joined December 2021").Layout,
	),
	lotusui.Button(th, &btn, "@nextjs", lotusui.ButtonProps{Variant: lotusui.ButtonLink}),
)`,
				Demo:  "hover-card/3",
				DemoH: 400,
			},
			{
				Heading: "Sides",
				Snippet: `tip.Side = lotusui.HoverCardLeft // Top / Bottom / Right`,
				Demo:    "hover-card/4",
				DemoH:   360,
			},
			{
				Heading: "RTL",
				Prose:   `<p>RTL layout is "from the web" — not yet supported by the underlying toolkit.</p>`,
			},
			{
				Heading: "Glossary / compact",
				Prose: `<p><strong>Multi-site:</strong> one <code>HoverCard</code> value may call
<code>Layout</code> at many sites in a frame (every occurrence of a jargon term); only the
hovered site paints. <strong>Compact tips:</strong> set <code>Width</code>,
<code>OpenDelay</code>, and <code>Side</code> on the struct — defaults stay 320dp / 700ms
(do not change library defaults for density). For in-text glossary runs see
<a href="../annotated-text/">AnnotatedText</a>. <code>Tooltip</code> is the wrong tool for
stay-open glossary panels (inverted string tip, no card hover).</p>`,
			},
		},
		Props: []prop{
			{"Layout(th, gtx, content, trigger)", "method", "Wrap any trigger; content is the card body. Opens after OpenDelay; stays up while the pointer is over trigger or card."},
			{"Side", "HoverCardSide", "HoverCardBottom (default), HoverCardTop, HoverCardLeft, HoverCardRight."},
			{"Align", "PopoverAlign", "PopoverCenter (default, shadcn), PopoverStart, PopoverEnd along the side."},
			{"OpenDelay / CloseDelay", "time.Duration", "Zero uses 700ms / 300ms (Radix defaults)."},
			{"Width", "unit.Dp", "Max card width; zero means 320dp (shadcn w-80). Card hugs content up to that max."},
			{"Disabled", "bool", "Prevents opening (chakra)."},
		},
	}
}

func popoverPage() *page {
	return &page{
		Slug:   "popover",
		Title:  "Popover",
		Kicker: "Arbitrary content on the floating layer, anchored to a trigger.",
		Intro: `<p>The general form of Select's dropdown, on the same portal primitive: the caller
owns <code>Open</code>, the panel floats 4dp below the anchor, and pressing anywhere else or
Escape closes it.</p>`,
		Sections: []section{
			installSection("popover"),
			{
				Heading: "Usage",
				Snippet: `var pop lotusui.Popover

if trigger.Clicked(gtx) { pop.Open = !pop.Open }
dims := lotusui.Button(th, &trigger, "Open", lotusui.ButtonProps{})(gtx)
pop.Layout(th, gtx, dims.Size, panelContent)`,
				Demo:  "popover/0",
				DemoH: 360,
			},
			{
				Heading: "Alignments",
				Prose: `<p><code>Align</code> positions the panel's edge against the anchor:
<code>PopoverCenter</code> (default, shadcn), <code>PopoverStart</code>, <code>PopoverEnd</code>.</p>`,
				Snippet: `pop.Width = 160
pop.Align = lotusui.PopoverEnd`,
				Demo:  "popover/1",
				DemoH: 280,
			},
		},
		Props: []prop{
			{"Open", "bool", "Caller-owned; toggled from the trigger's click."},
			{"Width", "unit.Dp", "Max panel width; zero matches the anchor. Non-zero: hug content up to Width."},
			{"Align", "PopoverAlign", "PopoverCenter (default, shadcn), PopoverStart, PopoverEnd against the anchor."},
		},
	}
}

func progressPage() *page {
	return &page{
		Slug:   "progress",
		Title:  "Progress",
		Kicker: "A determinate bar in the brand solid; negative value sweeps indeterminate.",
		Intro: `<p><code>value</code> is clamped to [0, 1]; pass a negative value for the
indeterminate sweep (self-invalidating — mount only while in flight).</p>`,
		Sections: []section{
			installSection("progress"),
			{
				Heading: "Usage",
				Snippet: `lotusui.Progress(th, 0.66)`,
				Demo:    "progress/0",
				DemoH:   150,
			},
			{
				Heading: "Indeterminate",
				Snippet: `lotusui.Progress(th, -1)`,
				Demo:    "progress/1",
				DemoH:   100,
			},
			{
				Heading: "Label",
				Prose:   `<p>A label row above the bar is composition.</p>`,
				Snippet: `lotusui.VStack(th.Space.SM, labelAndValueRow, lotusui.Progress(th, 0.66))`,
				Demo:    "progress/2",
				DemoH:   130,
			},
		},
		Props: []prop{
			{"value", "float32", "Clamped to [0, 1]; negative = the indeterminate sweep."},
		},
	}
}

func radioGroupPage() *page {
	return &page{
		Slug:   "radio-group",
		Title:  "RadioGroup",
		Kicker: "The exclusive choice: labeled circles, exactly one selected.",
		Intro: `<p>Options carry their own <code>Description</code> and <code>Disabled</code> flag;
clicks are processed at the top of Layout, so the click's own frame renders the new selection.</p>
<p class="note">Like every choice component, the selection is a VALUE: the cursor is
unexported, <code>Value()</code>/<code>SetValue()</code> read and write it, the zero value picks
the first option, and an unknown value clears rather than falling back to option 0 — see
<a href="../select/">Select</a> for the full contract.</p>`,
		Sections: []section{
			installSection("radio-group"),
			{
				Heading: "Usage",
				Snippet: `density := lotusui.RadioGroup{
	Options: lotusui.RadioOpts("Default", "Comfortable", "Compact"),
}
density.SetValue(stored) // an unknown value clears, never picks option 0

density.Layout(th, gtx)
stored = density.Value()`,
				Demo:  "radio-group/0",
				DemoH: 150,
			},
			{
				Heading: "Disabled option",
				Snippet: `plans := lotusui.RadioGroup{Options: []lotusui.RadioOption{
	{Label: "Starter", Value: "starter"},
	{Label: "Pro (contact sales)", Value: "pro", Disabled: true},
	{Label: "Enterprise", Value: "enterprise"},
}}`,
				Demo:  "radio-group/1",
				DemoH: 150,
			},
			{
				Heading: "Description",
				Prose:   `<p>Each option's <code>Description</code> renders muted under its label.</p>`,
				Snippet: `density := lotusui.RadioGroup{Options: []lotusui.RadioOption{
	{Label: "Default", Value: "default", Description: "Standard spacing for most use cases."},
	{Label: "Comfortable", Value: "comfortable", Description: "More space between elements."},
	{Label: "Compact", Value: "compact", Description: "Dense layout for scanning lots of data."},
}}`,
				Demo:  "radio-group/2",
				DemoH: 220,
			},
			{
				Heading: "Invalid",
				Prose:   `<p><code>Invalid</code> renders danger rings; pair it with a Field error.</p>`,
				Snippet: `prefs := lotusui.RadioGroup{Invalid: true}`,
				Demo:    "radio-group/3",
				DemoH:   240,
			},
		},
		Props: []prop{
			{"Options", "[]RadioOption", "The choices: Label, Value, Description (muted sub-label) and Disabled. RadioOpts(\"a\",\"b\") builds label-only lists."},
			{"Value() / SetValue(v)", "string", "The chosen option's value; SetValue with an unknown value clears."},
			{"Clear() / Chosen()", "method", "Drop the choice; whether anything is chosen."},
			{"Size", "Size", "The shared size presets scale the circles."},
			{"Invalid", "bool", "Danger chrome on the circles."},
		},
	}
}

func separatorPage() *page {
	return &page{
		Slug:   "separator",
		Title:  "Separator",
		Kicker: "The semantic divider: a 1dp hairline, horizontal or vertical.",
		Intro: `<p>Visually separates content — shadcn Separator. <code>Separator</code> /
<code>SeparatorVertical</code> are the component form of <code>Hairline</code> /
<code>VerticalHairline</code> (same paint; prefer the Separator names in app code).</p>`,
		Sections: []section{
			installSection("separator"),
			{
				Heading: "Usage",
				Snippet: `lotusui.Separator(th)          // horizontal
lotusui.SeparatorVertical(th)  // between row siblings`,
				Demo:  "separator/0",
				DemoH: 130,
			},
			{
				Heading: "List",
				Prose:   `<p>Inline items divided by vertical rules.</p>`,
				Snippet: `// Blog | Docs | Source — SeparatorVertical between Rigid children`,
				Demo:    "separator/1",
				DemoH:   100,
			},
			{
				Heading: "Hairline aliases",
				Prose: `<p><code>Hairline</code> / <code>VerticalHairline</code> are the low-level
primitives (<code>components.go</code>). <code>Separator</code> sets
<code>Min.X = Max.X</code> then calls <code>Hairline</code>;
<code>SeparatorVertical</code> is an alias of <code>VerticalHairline</code>.
Use Separator in screens; Hairline remains for internal/composed chrome
(tables, menus).</p>`,
				Snippet: `lotusui.Hairline(th)         // ≡ Separator without forcing Min.X
lotusui.VerticalHairline(th) // ≡ SeparatorVertical`,
			},
		},
		Props: []prop{
			{"Separator(th)", "widget", "The horizontal 1dp rule (forces full width)."},
			{"SeparatorVertical(th)", "widget", "Divides horizontal siblings."},
			{"Hairline(th) / VerticalHairline(th)", "widget", "Low-level aliases; prefer Separator in app code."},
		},
	}
}

func skeletonPage() *page {
	return &page{
		Slug:   "skeleton",
		Title:  "Skeleton",
		Kicker: "The loading placeholder: rounded blocks and circles pulsing on the shared clock.",
		Intro: `<p>Give it the shape of the content it stands in for. Zero width fills the available
width; <code>SkeletonCircle</code> is the round form. Self-invalidating — mount only while
loading.</p>`,
		Sections: []section{
			installSection("skeleton"),
			{
				Heading: "Usage",
				Snippet: `layout.Flex{Alignment: layout.Middle}.Layout(gtx,
	layout.Rigid(lotusui.SkeletonCircle(th, 48)),
	layout.Rigid(lotusui.HSpacer(th.Space.MD)),
	layout.Rigid(lotusui.VStack(th.Space.SM,
		lotusui.Skeleton(th, 250, 16),
		lotusui.Skeleton(th, 200, 16),
	)),
)`,
				Demo:  "skeleton/0",
				DemoH: 120,
			},
			{
				Heading: "Avatar",
				Snippet: `lotusui.SkeletonCircle(th, 40) // beside two text lines`,
				Demo:    "skeleton/1",
				DemoH:   110,
			},
			{
				Heading: "Card",
				Prose:   `<p>Two header lines and a 16:9 media block inside a Card.</p>`,
				Snippet: `lotusui.Card(th, lotusui.CardProps{}, lotusui.VStack(th.Space.SM,
	lotusui.Skeleton(th, 180, 16),
	lotusui.Skeleton(th, 140, 16),
	mediaBlockSkeleton, // Skeleton(th, 0, h) at 16:9
))`,
				Demo:  "skeleton/2",
				DemoH: 220,
			},
			{
				Heading: "Form",
				Snippet: `lotusui.VStack(th.Space.LG,
	lotusui.VStack(th.Space.SM, lotusui.Skeleton(th, 80, 16), lotusui.Skeleton(th, 0, 32)),
	lotusui.VStack(th.Space.SM, lotusui.Skeleton(th, 96, 16), lotusui.Skeleton(th, 0, 32)),
	lotusui.Skeleton(th, 96, 32),
)`,
				Demo:  "skeleton/3",
				DemoH: 240,
			},
			{
				Heading: "Table",
				Snippet: `// five rows of [flexible, 96dp, 80dp] lines`,
				Demo:    "skeleton/4",
				DemoH:   180,
			},
			{
				Heading: "Text",
				Snippet: `lotusui.VStack(th.Space.SM,
	lotusui.Skeleton(th, 0, 16),
	lotusui.Skeleton(th, 0, 16),
	threeQuarterLine,
)`,
				Demo:  "skeleton/5",
				DemoH: 130,
			},
		},
		Props: []prop{
			{"Skeleton(th, w, h)", "unit.Dp", "A pulsing rounded block; w = 0 fills the available width."},
			{"SkeletonCircle(th, d)", "unit.Dp", "The round form — avatars, icon slots."},
		},
	}
}

func sliderPage() *page {
	return &page{
		Slug:   "slider",
		Title:  "Slider",
		Kicker: "The draggable value: track, brand fill, a ringed thumb.",
		Intro: `<p><code>Value</code> is a fraction in [0, 1] — map it to your domain at the call
site. Press or drag anywhere on the control; <code>Step</code> snaps to multiples.</p>`,
		Sections: []section{
			installSection("slider"),
			{
				Heading: "Usage",
				Snippet: `var volume = lotusui.Slider{Value: 0.4}
volume.Layout(th, gtx)
percent := volume.Value * 100`,
				Demo:  "slider/0",
				DemoH: 120,
			},
			{
				Heading: "Step",
				Snippet: `quarters := lotusui.Slider{Step: 0.25}`,
				Demo:    "slider/1",
				DemoH:   90,
			},
			{
				Heading: "Disabled",
				Snippet: `s.Disabled = true`,
				Demo:    "slider/2",
				DemoH:   90,
			},
			{
				Heading: "Range",
				Prose:   `<p><code>Values</code> with two entries — two thumbs, the fill between them; each stays between its neighbors.</p>`,
				Snippet: `s := lotusui.Slider{Values: []float32{0.25, 0.5}, Step: 0.05}`,
				Demo:    "slider/3",
				DemoH:   90,
			},
			{
				Heading: "Multiple",
				Prose:   `<p>Any number of entries — one thumb each.</p>`,
				Snippet: `s := lotusui.Slider{Values: []float32{0.1, 0.2, 0.7}, Step: 0.1}`,
				Demo:    "slider/4",
				DemoH:   90,
			},
			{
				Heading: "Vertical",
				Prose:   `<p><code>Vertical</code> rotates the axis; cap the height at the call site — the value grows upward.</p>`,
				Snippet: `s := lotusui.Slider{Value: 0.5, Vertical: true}
gtx.Constraints.Max.Y = gtx.Dp(160)
s.Layout(th, gtx)`,
				Demo:  "slider/5",
				DemoH: 220,
			},
		},
		Props: []prop{
			{"Value", "float32", "Fraction in [0, 1] — map to your domain at the call site."},
			{"Values", "[]float32", "Multi-thumb mode: one thumb per entry, kept ordered; two = range."},
			{"Step", "float32", "Snaps to multiples; zero = continuous."},
			{"Vertical", "bool", "Rotates the axis; the value grows upward."},
			{"Disabled", "bool", "Dimmed, inert."},
		},
	}
}

func spinnerPage() *page {
	return &page{
		Slug:   "spinner",
		Title:  "Spinner",
		Kicker: "The standalone loading arc — Button's spinner, mountable anywhere.",
		Intro: `<p>Self-invalidating; mount it only while loading. <code>SpinnerTint</code> takes
any ink.</p>`,
		Sections: []section{
			installSection("spinner"),
			{
				Heading: "Usage",
				Snippet: `lotusui.Spinner(th, 24)
lotusui.SpinnerTint(th, 24, th.Palette.BrandFg)`,
				Demo:  "spinner/0",
				DemoH: 100,
			},
			{
				Heading: "Badge",
				Prose:   `<p>Inside a Badge via its <code>Start</code> slot — the syncing states.</p>`,
				Snippet: `lotusui.Badge(th, "Updating", lotusui.BadgeProps{Variant: lotusui.BadgeSecondary,
	Start: lotusui.Spinner(th, 12)})`,
				Demo:  "spinner/1",
				DemoH: 100,
			},
		},
		Props: []prop{
			{"Spinner(th, size)", "unit.Dp", "The arc in muted ink; self-invalidating."},
			{"SpinnerTint(th, size, color)", "color.NRGBA", "The arc in any ink."},
		},
	}
}

func tablePage() *page {
	return &page{
		Slug:   "table",
		Title:  "Table",
		Kicker: "Bounded tabular data: muted header, hairline rows, widget cells.",
		Intro: `<p>For the comparison-shaped data a screen actually shows. Cells are arbitrary
widgets (<code>Table</code>) or plain strings (<code>TableText</code>); column weights via
<code>Widths</code>. Long collections belong in <a href="../listview/">ListView</a>.</p>`,
		Sections: []section{
			installSection("table"),
			{
				Heading: "Usage",
				Snippet: `lotusui.TableText(th, lotusui.TableProps{
	Caption: "A list of recent invoices.",
	Footer:  []string{"Total", "", "", "$750.00"},
},
	[]string{"Invoice", "Status", "Method", "Amount"},
	[][]string{
		{"INV-001", "Paid", "Credit card", "$250.00"},
		{"INV-002", "Pending", "Transfer", "$150.00"},
	})`,
				Demo:  "table/0",
				DemoH: 260,
			},
			{
				Heading: "Actions",
				Prose:   `<p>A quiet menu per row — a ghost trigger aligned to the row's end, its panel end-aligned so it stays on screen.</p>`,
				Snippet: `menu.Variant = lotusui.ButtonGhost
menu.Align = lotusui.PopoverEnd
menu.Layout(th, gtx, "⋯",
	lotusui.DropdownMenuItem(th, &edit, "Edit", false),
	lotusui.DropdownMenuItem(th, &dup, "Duplicate", false),
	lotusui.DropdownMenuSeparator(th),
	lotusui.DropdownMenuItem(th, &del, "Delete", true),
)`,
				Demo:  "table/1",
				DemoH: 380,
			},
		},
		Props: []prop{
			{"Widths", "[]float32", "Per-column flex weights; nil = equal."},
			{"Footer", "[]string", "A final emphasized row (totals) above the caption."},
			{"Caption", "string", "Muted caption under the table."},
		},
	}
}

func textareaPage() *page {
	return &page{
		Slug:   "textarea",
		Title:  "Textarea",
		Kicker: "The multi-line Input: same chrome, wrapping editor, minimum rows.",
		Intro: `<p>Mechanisms mirror Input — <code>Error</code> for danger chrome,
<code>Disabled</code> for read-only, the shared sizes, Field for structure.</p>`,
		Sections: []section{
			installSection("textarea"),
			{
				Heading: "Usage",
				Snippet: `var msg lotusui.Textarea
msg.LayoutField(th, gtx, "Type your message here…")`,
				Demo:  "textarea/0",
				DemoH: 140,
			},
			{
				Heading: "Invalid — with a Field label",
				Snippet: `msg.Error = "your message is required"
msg.Layout(th, gtx, "Message", "Tell us what happened")`,
				Demo:  "textarea/1",
				DemoH: 180,
			},
			{
				Heading: "Disabled",
				Snippet: `msg.Disabled = true`,
				Demo:    "textarea/2",
				DemoH:   170,
			},
			{
				Heading: "With button",
				Snippet: `lotusui.VStack(th.Space.SM,
	messageField,
	lotusui.FullWidth(lotusui.Button(th, &send, "Send message", lotusui.ButtonProps{})),
)`,
				Demo:  "textarea/3",
				DemoH: 200,
			},
		},
		Props: []prop{
			{"Size", "Size", "The shared size presets (Size2XS–Size2XL)."},
			{"Variant", "InputVariant", "InputOutline (default), InputSubtle, InputFlushed."},
			{"Rows", "int", "Minimum visible lines; zero means 3."},
			{"Error", "string", "Danger chrome + message."},
			{"Disabled", "bool", "Read-only, dimmed."},
		},
	}
}

func toastPage() *page {
	return &page{
		Slug:   "toast",
		Title:  "Toast",
		Kicker: "Transient notifications: bottom-right stack, auto-dismissing.",
		Intro: `<p><code>Toaster</code> owns the queue: <code>Add</code> from anywhere, Layout ONCE
per frame at window constraints (the same portal rule as Dialog). Each toast dismisses after
its duration; destructive toasts carry danger ink.</p>`,
		Sections: []section{
			installSection("toast"),
			{
				Heading: "Usage",
				Snippet: `var toaster lotusui.Toaster // one per window

if saved { toaster.Add(lotusui.Toast{Title: "Event has been created"}) }

// at the END of your window's frame:
toaster.Layout(th, gtx)`,
				Demo:  "toast/0",
				DemoH: 340,
			},
			{
				Heading: "Types",
				Prose:   `<p>Success, info and warning color the title's ink; destructive colors everything.</p>`,
				Snippet: `toaster.Add(lotusui.Toast{Variant: lotusui.ToastSuccess, Title: "Event has been created"})
toaster.Add(lotusui.Toast{Variant: lotusui.ToastInfo, Title: "Arrive 10 minutes before the event"})`,
				Demo:  "toast/1",
				DemoH: 340,
			},
			{
				Heading: "Promise",
				Prose: `<p>The in-flight pattern: <code>Add</code> a <code>Loading</code> toast under an
<code>ID</code>, then <code>Update</code> the same ID with the outcome — it replaces in place
and restarts the clock.</p>`,
				Snippet: `toaster.Add(lotusui.Toast{ID: "create", Loading: true, Title: "Creating event…", Duration: time.Minute})
// …when the work completes:
toaster.Update("create", lotusui.Toast{Variant: lotusui.ToastSuccess, Title: "Event created."})`,
				Demo:  "toast/2",
				DemoH: 340,
			},
		},
		Props: []prop{
			{"Toaster.Add(Toast)", "method", "Enqueue from anywhere; safe during event handling."},
			{"Toaster.Update(id, Toast)", "method", "Replace the live toast with that ID in place — the promise pattern."},
			{"Toast.Title / Description", "string", "The notification's copy."},
			{"Toast.Variant", "ToastVariant", "ToastDefault, ToastDestructive, ToastSuccess, ToastInfo, ToastWarning."},
			{"Toast.ID", "string", "Names the toast for Update."},
			{"Toast.Loading", "bool", "A spinner before the title."},
			{"Toast.Duration", "time.Duration", "Auto-dismiss delay; zero means 4s."},
		},
	}
}

func togglePage() *page {
	return &page{
		Slug:   "toggle",
		Title:  "Toggle",
		Kicker: "The pressed-state button, and groups with single or multiple selection.",
		Intro: `<p><code>Toggle</code> stays filled while <code>On</code>; clicking flips it.
<code>ToggleGroup</code> coordinates a row — radio semantics by default
(<code>Sel</code>), independent bools with <code>Multiple</code>.</p>`,
		Sections: []section{
			installSection("toggle"),
			{
				Heading: "Usage",
				Snippet: `var bold lotusui.Toggle
bold.Layout(th, gtx, lotusui.ToggleProps{Icon: lotusui.IconTextBold})
if bold.On { /* … */ }`,
				Demo:  "toggle/0",
				DemoH: 110,
			},
			{
				Heading: "ToggleGroup — single",
				Snippet: `format := lotusui.ToggleGroup{Options: []lotusui.ToggleOption{
	// Icon-only options MUST carry a Value — an empty Label has none.
	{Value: "bold", Icon: lotusui.IconTextBold},
	{Value: "italic", Icon: lotusui.IconTextItalic},
	{Value: "underline", Icon: lotusui.IconTextUnderline},
}}
format.Layout(th, gtx, lotusui.SizeMD)
active := format.Value() // "italic"`,
				Demo:  "toggle/1",
				DemoH: 110,
			},
			{
				Heading: "ToggleGroup — multiple",
				Snippet: `marks := lotusui.ToggleGroup{Multiple: true, Options: []lotusui.ToggleOption{
	{Label: "Bold", Value: "bold", Icon: lotusui.IconTextBold},
	{Label: "Italic", Value: "italic", Icon: lotusui.IconTextItalic},
}}
marks.SetValues(stored)
stored = marks.Values() // in Options order`,
				Demo:  "toggle/2",
				DemoH: 110,
			},
			{
				Heading: "Sizes",
				Snippet: `small.Layout(th, gtx, lotusui.ToggleProps{Label: "Small", Outline: true, Size: lotusui.SizeSM})`,
				Demo:    "toggle/3",
				DemoH:   110,
			},
			{
				Heading: "Disabled",
				Snippet: `t.Layout(th, gtx, lotusui.ToggleProps{Label: "Disabled", Disabled: true})`,
				Demo:    "toggle/4",
				DemoH:   110,
			},
			{
				Heading: "Group outline",
				Snippet: `var g = lotusui.ToggleGroup{Sel: 0, Outline: true}`,
				Demo:    "toggle/5",
				DemoH:   110,
			},
			{
				Heading: "Group sizes",
				Prose:   `<p>The size is the Layout parameter — the whole group moves together.</p>`,
				Snippet: `g.Layout(th, gtx, lotusui.SizeSM, items...)`,
				Demo:    "toggle/6",
				DemoH:   150,
			},
			{
				Heading: "Group disabled",
				Snippet: `var g = lotusui.ToggleGroup{Multiple: true, Disabled: true}`,
				Demo:    "toggle/7",
				DemoH:   110,
			},
			{
				Heading: "Group vertical",
				Snippet: `var g = lotusui.ToggleGroup{Multiple: true, Vertical: true, Spacing: 1}`,
				Demo:    "toggle/8",
				DemoH:   190,
			},
			{
				Heading: "Group spacing",
				Prose:   `<p><code>Spacing</code> overrides the 2dp gap between items.</p>`,
				Snippet: `var g = lotusui.ToggleGroup{Sel: 0, Outline: true, Spacing: 8}`,
				Demo:    "toggle/9",
				DemoH:   110,
			},
			{
				Heading: "Font weight selector",
				Prose:   `<p><code>Content</code> replaces Icon/Label with arbitrary content inside the item's chrome — composed here with a Field.</p>`,
				Snippet: `lotusui.ToggleOption{Value: "bold", Content: weightCell("Aa", "Bold", font.Bold)}`,
				Demo:    "toggle/10",
				DemoH:   180,
			},
		},
		Props: []prop{
			{"On", "bool", "The pressed state — yours; clicking flips it."},
			{"Icon / Label", "string", "One or both."},
			{"Content", "layout.Widget", "Replaces Icon/Label with arbitrary content in the chrome."},
			{"Size", "Size", "The shared size presets."},
			{"Outline", "bool", "Bordered while off."},
			{"Disabled", "bool", "Dimmed, unclickable."},
			{"ToggleGroup.Options", "[]ToggleOption", "The choices: Label, Value, Icon, Content. ToggleOpts(\"a\",\"b\") builds label-only lists (the longer name: ToggleProps is this component's per-call options struct)."},
			{"ToggleGroup.Value() / SetValue(v)", "string", "Single-select choice; unknown values clear."},
			{"ToggleGroup.Values() / SetValues(vs)", "[]string", "Multi-select choices, in Options order; unknown values ignored."},
			{"ToggleGroup.Multiple", "bool", "Independent choices instead of radio semantics."},
			{"ToggleGroup.Outline / Disabled", "bool", "Group-wide chrome and state."},
			{"ToggleGroup.Vertical / Spacing", "bool / unit.Dp", "Stacked axis; gap override (default 2dp)."},
		},
	}
}

func tooltipPage() *page {
	return &page{
		Slug:   "tooltip",
		Title:  "Tooltip",
		Kicker: "The hover label: inverted chrome, a rest delay, pass-through events.",
		Intro: `<p>Wrap any widget; hold the pointer for a beat and the label floats in beneath it
on the portal layer. The hover area passes events through — the child keeps its own
interactions. One Tooltip struct per wrapped instance.</p>`,
		Sections: []section{
			installSection("tooltip"),
			{
				Heading: "Usage",
				Snippet: `var tip lotusui.Tooltip
tip.Layout(th, gtx, "Add to library",
	lotusui.Button(th, &btn, "Hover me", lotusui.ButtonProps{Variant: lotusui.ButtonOutline}))`,
				Demo:  "tooltip/0",
				DemoH: 160,
			},
			{
				Heading: "Sides",
				Prose:   `<p><code>Side</code> floats the label above, below, or beside the child.</p>`,
				Snippet: `tip := lotusui.Tooltip{Side: lotusui.TooltipTop}`,
				Demo:    "tooltip/1",
				DemoH:   180,
			},
		},
		Props: []prop{
			{"Layout(th, gtx, text, child)", "method", "Wrap any widget; label floats in after a rest delay."},
		},
	}
}

func itemPage() *page {
	return &page{
		Slug:   "item",
		Title:  "Item",
		Kicker: "A versatile row for media, title, description, and actions.",
		Intro: `<p>A straightforward flex container for nearly any content — title,
description, and actions. Group with <code>ItemGroup</code> for lists.
Use <code>Field</code> when the row is a form control; use <code>Item</code>
when it is display content (including multiline <code>SelectOption.Content</code>).</p>`,
		Sections: []section{
			installSection("item"),
			{
				Heading: "Usage",
				Snippet: `lotusui.Item(th, lotusui.ItemProps{
	Variant: lotusui.ItemOutline,
	Content: lotusui.ItemContent(th,
		lotusui.ItemTitle(th, "Basic Item"),
		lotusui.ItemDescription(th, "A simple item with title and description."),
	),
	Actions: lotusui.ItemActions(th,
		lotusui.Button(th, &action, "Action", lotusui.ButtonProps{Variant: lotusui.ButtonOutline, Size: lotusui.SizeSM}),
	),
})`,
				Demo:  "item/0",
				DemoH: 220,
			},
			{
				Heading: "Variant",
				Prose:   `<p><code>Variant</code> selects default (transparent), outline, or muted chrome.</p>`,
				Snippet: `lotusui.Item(th, lotusui.ItemProps{Variant: lotusui.ItemOutline, …})
lotusui.Item(th, lotusui.ItemProps{Variant: lotusui.ItemMuted, …})`,
				Demo:  "item/1",
				DemoH: 320,
			},
			{
				Heading: "Size",
				Prose:   `<p>Shared <code>Size</code> enum — MD default, SM → sm, XS → xs (shadcn's three sizes).</p>`,
				Snippet: `lotusui.Item(th, lotusui.ItemProps{Size: lotusui.SizeSM, …})`,
				Demo:    "item/2",
				DemoH:   320,
			},
			{
				Heading: "Icon",
				Prose:   `<p><code>ItemMedia</code> with <code>ItemMediaIcon</code> wraps an icon in a muted well.</p>`,
				Snippet: `Media: lotusui.ItemMedia(th, lotusui.ItemMediaIcon, lotusui.SVGIcon(lotusui.IconLock, 18, ink))`,
				Demo:    "item/3",
				DemoH:   140,
			},
			{
				Heading: "Avatar",
				Prose:   `<p>Pass an <code>Avatar</code> as <code>ItemMediaDefault</code> content.</p>`,
				Snippet: `Media: lotusui.ItemMedia(th, lotusui.ItemMediaDefault, lotusui.Avatar(th, lotusui.AvatarProps{Initials: "ER"}))`,
				Demo:    "item/4",
				DemoH:   120,
			},
			{
				Heading: "Image",
				Prose:   `<p><code>ItemMediaImage</code> clips the slot to a rounded square.</p>`,
				Snippet: `Media: lotusui.ItemMedia(th, lotusui.ItemMediaImage, imageWidget)`,
				Demo:    "item/5",
				DemoH:   120,
			},
			{
				Heading: "Group",
				Prose:   `<p><code>ItemGroup</code> stacks related items; optional <code>ItemSeparator</code> between them.</p>`,
				Snippet: `lotusui.ItemGroup(th, item1, item2, item3)`,
				Demo:    "item/6",
				DemoH:   280,
			},
			{
				Heading: "Header",
				Prose:   `<p><code>ItemHeader</code> spans the full width above the main columns — model cards, media tiles.</p>`,
				Snippet: `Header: lotusui.ItemHeader(th, media, nil)`,
				Demo:    "item/7",
				DemoH:   220,
			},
			{
				Heading: "Link",
				Prose: `<p>Set <code>Btn</code> to make the whole row clickable (shadcn's <code>render</code>
prop). Hover fills like a HoverRow. Opening URLs is the caller's job — see
<code>OpenURL</code> when you need the system browser.</p>`,
				Snippet: `lotusui.Item(th, lotusui.ItemProps{Btn: &go, Content: …, Actions: …})`,
				Demo:    "item/8",
				DemoH:   200,
			},
			{
				Heading: "From the web",
				Prose: `<p>shadcn's <code>render</code>/<code>asChild</code> polymorphism and RTL
demo are web-document concerns — Gio uses an explicit <code>Btn</code> slot
and the platform's text direction. Dropdown-on-item is composition with
<code>DropdownMenu</code>.</p>`,
			},
		},
		Props: []prop{
			{"ItemProps.Variant", "ItemVariant", "ItemDefault, ItemOutline, ItemMuted."},
			{"ItemProps.Size", "Size", "Padding scale; MD default, SM/XS match shadcn sm/xs."},
			{"ItemProps.Btn", "*widget.Clickable", "Nil = static; set for whole-row click/hover."},
			{"ItemProps.Header/Media/Content/Actions/Footer", "layout.Widget", "Build-time slots; nil omitted."},
			{"ItemMedia(th, variant, content)", "func", "Leading well — Default / Icon / Image."},
			{"ItemContent / ItemTitle / ItemDescription", "func", "Text column helpers."},
			{"ItemActions / ItemHeader / ItemFooter", "func", "Trailing and full-width bands."},
			{"ItemGroup / ItemSeparator", "func", "Stacked list + hairline."},
		},
	}
}
