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
| **Badge** | demo ✓ · variants ✓ · icon ✓ · colors ✓ · spinner ✓ · link ➖ · rtl ➖ |
| **Breadcrumb** | demo ✓ · basic ✓ · link ✓ · separator ✓ · ellipsis ✓ · dropdown ✓ · responsive ✓(BreadcrumbNav) · rtl ➖ |
| **Button** | demo ✓ · default ✓ · secondary ✓ · destructive ✓ · outline ✓ · ghost ✓ · link ✓ · icon ✓ · with-icon ✓ · rounded ✓ · spinner ✓ · size ✓ · render ➖ · rtl ➖ |
| **ButtonGroup** | demo ✓ · orientation ✓ · size ✓ · separator ✓ · split ✓ · input ✓ · select ✓(ButtonGroupSelect) · nested 🧩(flex Widget slot) · input-group ⏳ · popover ⏳ · dropdown 🧩(trigger beside group) |
| **Card** | demo ✓ · small ✓ · spacing ✓ · image ✓ · edge-to-edge ✓ · rtl ➖ |
| **Checkbox** | demo ✓ · basic ✓ · description ✓ · disabled ✓ · group ✓ · invalid ✓ · table ✓ · rtl ➖ |
| **Dialog** | demo ✓ · close-button ✓ · no-close-button ✓ · scrollable-content ✓ · sticky-footer ✓ · **responsive width** ✓ · rtl ➖ |
| **DropdownMenu** | demo ✓ · basic ✓ · icons ✓ · shortcuts ✓ · checkboxes ✓ · radio-group ✓ · destructive ✓ · complex ✓ · checkboxes-icons ✓ · radio-icons ✓ · submenu ✓ · avatar ⏳ · rtl ➖ |
| **Field** | demo ✓ · input ✓ · select ✓ · textarea ✓ · group 🧩 · checkbox ⏳ · radio ⏳ · switch ⏳ · slider ⏳ · fieldset ⏳ · choice-card ⏳ · responsive ⏳ · rtl ➖ |
| **Input** | demo ✓ · basic ✓ · disabled ✓ · field ✓ · fieldgroup ✓ · grid ✓ · badge ✓ · inline ✓ · invalid ✓ · required ✓ · input-group 🧩(Start/End) · button-group ✓(ButtonGroupInput + Attached) · file ➖ · form ➖ · rtl ➖ |
| **Item** | demo ✓ · variant ✓ · size ✓ · icon ✓ · avatar ✓ · image ✓ · group ✓ · header ✓ · link ✓ · dropdown 🧩(DropdownMenu) · rtl ➖ |
| **InputGroup** | icon ✓ · text ✓ · kbd ✓ · spinner ✓ · button ✓ · block-start ✓ · block-end ✓ · basic ✓(Input frame) · textarea ⏳ · dropdown ⏳ · tooltip ⏳ · custom 🧩 — mapped onto Input's Start/End/Top/Bottom slots + Kbd |
| **InputOTP** | demo ✓ · separator ✓ · four-digits ✓ · pattern ✓ · disabled ✓ · invalid ✓ · controlled ✓(Value/SetValue) · form 🧩 · rtl ➖ |
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
| **Toggle** | demo ✓ · outline ✓ · text ✓ · sizes ✓ · disabled ✓ · rtl ➖ |
| **ToggleGroup** | demo ✓ · outline ✓ · sizes ✓ · disabled ✓ · vertical ✓ · spacing ✓ · font-weight-selector ✓ · rtl ➖ |
| **Tooltip** | demo ✓ · sides ✓ · keyboard ⏳ · disabled ⏳ · rtl ➖ |
| **Scroll Area** | demo ✓ · composition (Scrollbar) ✓ · horizontal ✓ · **always** ✓ · **sizes** ✓ · **thumb color** ✓ · **show track** ✓ · floating-inside ✓ · rtl ➖ |
| **Extensions** (no counterpart) | Grid (spans, auto-flow, row contract, **Cols responsive**) ✓ · SimpleGrid (continuous minChildWidth + **stepped Columns**) ✓ · Theme breakpoints (ParseBreakpointsJSON / WithBreakpoints) ✓ · Show / ResponsiveInt|Dp|Bool ✓ · Stack (VStack / HStack) ✓ · Wrap ✓ · Split + VSlide ✓ · Split scroll helpers (ColumnScroll / BoxScroll / FillScroll) ✓ · TitleWithIcons ✓ · AnnotatedText / SplitGlossary ✓ · ListView (virtualized) ✓ · Floating (portal) ✓ · **CodeBlock** / CodeSpan (highlighter stays in the app) ✓ · **Example** (Preview\|Code chrome) ✓ · **ScrollArea** / **Scrollbar** (macOS overlay; Floating-safe; List-backed Scrollable remains) ✓ |
| **Missing families** | Empty, Combobox, Command, Menubar, Navigation Menu, Sheet/Drawer, Calendar, Context Menu ⏳ · Item ✓ |

## Working order

1. Close the demo-pending ⏳s where the API already exists (Accordion
   multiple, AlertDialog destructive, RadioGroup invalid, Textarea
   disabled, Toggle sizes/disabled) — pure demo work.
2. ToggleGroup and Tooltip example sets; the small per-component ⏳s
   (Pagination simple, Progress label, Skeleton shapes, Table footer).
3. The ButtonGroup family (attached, shared-border groups) — it also
   closes Input button-group and four Button-page examples.
4. InputGroup family, then Slider range/vertical, then submenus.
5. New families in order of ubiquity: Sheet, Combobox, InputOTP,
   Empty.
