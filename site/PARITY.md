# Example parity ledger

The contract: every example the reference design system documents for
a component exists in lotusui's docs as a working live demo — same
composition, same content register, transposed to Go/Gio. This ledger
tracks that FILE BY FILE against the reference's example sources; a
component is DONE when its row has no ⏳. Components with no
counterpart are lotusui EXTENSIONS, documented with the examples the
reference would have shown.

Legend: ✓ implemented + documented + demoed · 🧩 delivered by
composition (documented) · ⏳ pending · ➖ not applicable to a native
canvas UI (RTL layout, web form libraries, asChild/render props,
native file pickers) — noted on the page, which counts as parity
because it is auditable.

| Component | Examples (reference file → status) |
|---|---|
| **Accordion** | demo ✓ · basic ✓ · multiple ✓ · disabled ✓ · borders ✓ · card ✓ · rtl ➖ |
| **Alert** | demo ✓ · basic ✓ · destructive ✓ · action ✓ · colors ✓ · rtl ➖ |
| **AlertDialog** | demo ✓ · basic ✓ · destructive ✓ · small ✓ · media ✓(in destructive) · small-media 🧩 · rtl ➖ |
| **Avatar** | demo ✓ · basic ✓ · size ✓ · group ✓ · badge ✓ · badge-icon ✓ · group-count ✓ · group-count-icon ✓ · dropdown ⏳ · rtl ➖ |
| **Badge** | demo ✓ · variants ✓(split one-per-section, shadcn's page now shows a single Variants row ⏳) · icon ✓(second badge is `Changes`; shadcn's is a trailing-icon `Bookmark` ⏳) · colors ✓(3 hues; shadcn shows 5) · spinner ✓ · link ⏳(shadcn's `asChild` anchor — a clickable badge is expressible) · rtl ➖ — the badge is a PILL (`ClampCorner(h/2)`), matching shadcn's `rounded-full` |
| **Breadcrumb** | demo ✓ · basic ✓ · link ✓ · separator ✓ · ellipsis ✓ · dropdown ✓ · responsive ✓(BreadcrumbNav) · rtl ➖ |
| **Button** | demo ✓ · default ✓ · secondary ✓ · destructive ✓ · outline ✓ · ghost ✓ · link ✓ · icon ✓ · with-icon ✓(ours is `Add item`/`Continue`; shadcn's is one sm `New Branch` ⏳) · rounded ✓ · spinner ✓(ours `Loading`/`Please wait`; shadcn's `Generating`/`Downloading`, one with a TRAILING spinner ⏳) · size ✓(ours 2XS–2XL labels; shadcn pairs each size with a matching icon button ⏳) · as-child ➖(a Link-variant button IS the link here; noted in the page intro) · button-group ✓(own family + page) · render ➖ · rtl ➖ |
| **ButtonGroup** | demo ✓ · orientation ✓ · size ✓ · separator ✓ · split ✓ · input ✓ · select ✓(ButtonGroupSelect) · nested 🧩(flex Widget slot) · input-group ⏳ · popover ⏳ · dropdown 🧩(trigger beside group) |
| **Card** | demo ✓ · small ✓ · spacing ✓ · image ✓ · edge-to-edge ✓ · rtl ➖ |
| **Checkbox** | demo ✓ · basic ✓ · description ✓ · disabled ✓ · group ✓ · invalid ✓ · table ✓ · rtl ➖ |
| **Context Menu** | demo ✓ · basic ✓ · submenu ✓ · shortcuts ✓ · groups ✓ · icons ✓ · checkboxes ✓ · radio ✓ · destructive ✓ · disabled-rows ⏳ · touch long-press ⏳ · sides ⏳(a real section on shadcn's base and aria pages; the panel already flips at a known edge) · rtl ➖ — opens at the pointer via `ContextMenuPress` (secondary press everywhere + Ctrl+primary on macOS); rows are the shared menu grammar under family names |
| **Dialog** | demo ✓ · close-button ✓ · no-close-button ✓ · scrollable-content ✓ · sticky-footer ✓ · **responsive width** ✓ · rtl ➖ |
| **DropdownMenu** | demo ✓ · basic ✓ · icons ✓ · shortcuts ✓ · checkboxes ✓ · radio-group ✓ · destructive ✓ · complex ✓ · checkboxes-icons ✓ · radio-icons ✓ · submenu ✓ · avatar ⏳(Avatar + a ghost round trigger both exist) · composition ⏳(shadcn's part-hierarchy tree; we ship a Props table only) · disabled-rows ⏳(no menu-row constructor takes a disabled flag — shadcn's `Forward`/`API` rows) · rtl ➖ — several sections carry OUR example copy, not shadcn's (basic, icons, shortcuts, destructive, complex) ⏳ |
| **Field** | demo ✓ · input ✓ · select ✓ · textarea ✓ · group 🧩 · checkbox ⏳ · radio ⏳ · switch ⏳ · slider ⏳ · fieldset ⏳ · choice-card ⏳ · responsive ⏳ · rtl ➖ |
| **Input** | demo ✓ · basic ✓ · disabled ✓ · field ✓ · fieldgroup ✓ · grid ✓ · badge ✓ · inline ✓ · invalid ✓ · required ✓ · input-group 🧩(Start/End) · button-group ✓(ButtonGroupInput + Attached) · file ➖ · form ➖ · rtl ➖ |
| **Item** | demo ✓ · variant ✓ · size ✓ · icon ✓ · avatar ✓ · image ✓ · group ✓ · header ✓ · link ✓ · dropdown 🧩(DropdownMenu) · rtl ➖ |
| **InputGroup** | icon ✓ · text ✓ · kbd ✓ · spinner ✓ · button ✓ · block-start ✓ · block-end ✓ · basic ✓(Input frame) · textarea ⏳ · dropdown ⏳ · tooltip ⏳ · custom 🧩 — mapped onto Input's Start/End/Top/Bottom slots + Kbd |
| **InputOTP** | demo ✓ · separator ✓ · four-digits ✓ · pattern ✓ · disabled ✓ · invalid ✓ · controlled ✓(Value/SetValue; no section echoing the value like shadcn's ⏳) · alphanumeric ⏳(`Filter` already expresses it) · form 🧩(Card + Field + Button compose it; shadcn ships a verify-login card ⏳) · rtl ➖ |
| **Pagination** | demo ✓ · simple ✓ · rows-per-page ✓ · icons-only ⏳ · rtl ➖ |
| **HoverCard** | demo ✓ · usage ✓ · composition ➖(Go Layout) · trigger-delays ✓ · positioning ✓ · basic ✓ · sides ✓ · rtl ➖ · disabled ✓(prop) · glossary/compact ✓ · controlled ⏳ · arrow ➖ |
| **Popover** | demo ✓ · basic ✓ · form ✓ · alignments ✓ · rtl ➖ |
| **Progress** | demo ✓ · controlled ✓ · label ✓ · rtl ➖ |
| **RadioGroup** | demo ✓ · disabled ✓ · invalid ✓ · description ✓ · fields ⏳ · fieldset ⏳ · choice-card ⏳ · rtl ➖ |
| **Select** | demo ✓ · align-item ✓ · groups ✓ · scrollable ✓ · meta ✓ · icons ✓ · subscription-plan ✓(Content) · disabled ✓ · invalid ✓ · sizes ✓ · rtl ➖ |
| **Separator** | demo ✓ · vertical ✓ · menu ✓ · list ✓ · rtl ➖ |
| **Skeleton** | demo ✓ · avatar ✓ · text ✓ · card ⏳ · form ⏳ · table ⏳ · rtl ➖ |
| **Slider** | demo ✓ · controlled ✓ · disabled ✓ · range ✓ · multiple ✓ · vertical ✓ · rtl ➖ |
| **Spinner** | demo ✓ · size ✓ · custom ✓ · button ✓ · badge ✓ · empty ⏳ · input-group ⏳ · rtl ➖ |
| **Switch** | demo ✓ · description ✓ · disabled ✓ · invalid ✓ · sizes ✓ · choice-card ✓ · rtl ➖ |
| **Table** | demo ✓ · footer ✓ · actions ✓ · rtl ➖ |
| **Tabs** | demo ✓ · line ✓ · vertical ✓ · icons ✓ · disabled ✓ · wrap ✓ · rtl ➖ |
| **Textarea** | demo ✓ · invalid ✓ · field ✓ · disabled ✓ · button ✓ · rtl ➖ |
| **Toast** | demo ✓ · types ✓ · promise ✓ |
| **Toggle** | demo ✓ · outline ✓(shown inside Usage/Sizes; no standalone section ⏳) · text ✓(inside Usage; no standalone section ⏳) · sizes ✓ · disabled ✓ · rtl ➖ — our page interleaves Toggle and ToggleGroup sections, so it reads out of shadcn's order ⏳ |
| **ToggleGroup** | demo ✓ · outline ✓ · sizes ✓ · disabled ✓ · vertical ✓ · spacing ✓ · font-weight-selector ✓ · rtl ➖ |
| **Tooltip** | demo ✓ · sides ✓ · keyboard ⏳(BLOCKED: `Tooltip.Layout` takes a `string`, so a `Kbd` cannot go in the label — needs a rich-content slot) · disabled ⏳(unblocked: `ButtonProps.Disabled` + the pass-through hover area already work) · rtl ➖(no "from the web" note on the page yet ⏳) |
| **Scroll Area** | demo ✓ · composition (Scrollbar) ✓ · horizontal ✓ · **always** ✓ · **sizes** ✓ · **thumb color** ✓ · **show track** ✓ · floating-inside ✓ · rtl ➖ |
| **Extensions** (no counterpart) | Grid (spans, auto-flow, row contract, **Cols responsive**) ✓ · SimpleGrid (continuous minChildWidth + **stepped Columns**) ✓ · Theme breakpoints (ParseBreakpointsJSON / WithBreakpoints) ✓ · Show / ResponsiveInt|Dp|Bool ✓ · Stack (VStack / HStack) ✓ · Wrap ✓ · Split + VSlide ✓ · Split scroll helpers (ColumnScroll / BoxScroll / FillScroll) ✓ · TitleWithIcons ✓ · AnnotatedText / SplitGlossary ✓ · ListView (virtualized) ✓ · Floating (portal) ✓ · **CodeBlock** / CodeSpan (highlighter stays in the app) ✓ · **Example** (Preview\|Code chrome) ✓ · **ScrollArea** / **Scrollbar** (macOS overlay; Floating-safe; List-backed Scrollable remains) ✓ · **FloatingPanel** (the shell's sidebar surface — documented on the Layout page; components.go is not vendorable, so no `lotusui add`) ✓ · **SVGIconButtonTint** (mono icon button carrying its own ink — documented on the Icons page) ✓ |
| **Missing families** | Empty, Combobox, Command, Menubar, Navigation Menu, Sheet/Drawer, Calendar ⏳ · Context Menu ✓ · Item ✓ |

## Working order

1. Close the demo-pending ⏳s where the API already exists (Accordion
   multiple, AlertDialog destructive, RadioGroup invalid, Textarea
   disabled, Toggle sizes/disabled) — pure demo work.
2. ToggleGroup and Tooltip example sets; the small per-component ⏳s
   (Pagination simple, Progress label, Skeleton shapes, Table footer).
2b. The 2026-08-12 audit against the LOCAL `_ref/shadcn-ui` checkout
   added ⏳s of two kinds, and they are not equal work: **missing
   sections** (tooltip disabled, context-menu sides, dropdown avatar,
   badge link, input-otp alphanumeric/form) are demo work against an
   API that already exists, while **content divergence** (our example
   copy where shadcn has its own — dropdown basic/icons/shortcuts/
   destructive/complex, checkbox group/disabled/table, button
   with-icon/spinner/size) is a transposition pass, page by page, with
   `_ref` open. One capability is genuinely missing and blocks four
   examples: no menu-row constructor takes a `disabled` flag.
   NOTE the live-demo indexes are now guarded — a docs `Demo: "x/N"`
   with no section N fails the docsapp smoke test instead of silently
   rendering the whole demo (that shipped once, on the Button page).
3. The ButtonGroup family (attached, shared-border groups) — it also
   closes Input button-group and four Button-page examples.
4. InputGroup family, then Slider range/vertical, then submenus.
5. New families in order of ubiquity: Sheet, Combobox, InputOTP,
   Empty.
