# Changelog

This changelog is written **for the AI agents that develop the apps
consuming lotusui**. From each release on, it records every API-visible
change — created, updated, renamed, removed — with exact symbols,
old→new forms, and replacement guidance, precise enough to migrate a
consuming app from this document alone. There is deliberately no
migrate tool: read the entries for the versions you crossed, apply
them, then run `go build ./...` until clean — the compiler is the
safety net. `lotusui update` prints the relevant sections when
upgrading vendored components.

Format: [Keep a Changelog](https://keepachangelog.com) sections,
[SemVer](https://semver.org) versions (Go modules enforce it). Until
v1.0.0 breaking changes may land in minor versions; each will be
recorded here in full.

## [Unreleased]

VERSION NUMBERING NOTE for the next release: it must be **v0.9.0**,
not v0.4.0. The abandoned pre-rewrite lineage published v0.4.0–v0.8.0
and the module proxy + checksum DB cache those version strings
FOREVER — a new tag reusing one would serve the ghost zip (or fail
sum verification), and anything inside the range is born retracted by
our own `retract [v0.4.0, v0.8.0]`. Jumping above v0.8.0 also finally
activates that retraction, so `go get @latest` resolves correctly
again from v0.9.0 on. Do not "correct" the jump.

### Added

| Symbol | Notes |
|---|---|
| `ContextMenu` | the shadcn Context Menu family: wraps any content in a PASS-THROUGH pointer area (the child keeps every primary-button interaction) and the platform's context gesture opens the menu panel AT THE POINTER on the floating layer. Fields: `KeepOpen bool` (suppress close-on-selection, for checkbox/radio menus), `Width unit.Dp` (panel max width; zero = min 224dp, grow with content). `Layout(th, gtx, content layout.Widget, items ...layout.Widget) layout.Dimensions`. Escape / outside press / selection closes; multi-`Layout` per frame is site-safe; a measure pass lays out the content alone. The panel opens down-right of the pointer and flips up/left at a known constraint edge (unbounded axes — scrollers — never flip) |
| `ContextMenuItem`, `ContextMenuItemIcon`, `ContextMenuShortcutItem`, `ContextMenuCheckboxItem`, `ContextMenuRadioItem`, `ContextMenuCheckboxItemIcon`, `ContextMenuRadioItemIcon`, `ContextMenuLabel`, `ContextMenuSeparator` | the menu row grammar under this family's names — same signatures and rendering as the `DropdownMenu*` rows of the same shape (they delegate); use either vocabulary, rows are interchangeable |
| `ContextMenuSub` | `= DropdownMenuSub` (type alias) — nested submenu; its `Item` goes in the items list |
| `ContextMenuPress(ev pointer.Event) bool` | THE platform answer for "is this a context-menu gesture": secondary-button press on every OS, plus Ctrl+primary on macOS (neither macOS nor Gio translates the one-button convention). Use it instead of testing buttons yourself |
| `ShortcutHint(k string) string` | platform spelling of the shortcut modifier for DISPLAY: `ShortcutHint("C")` → "⌘C" on macOS, "Ctrl+C" elsewhere (including wasm, where the host OS is unknowable at build time). Pair with Gio's `key.ModShortcut` when binding |
| `IconCopy`, `IconCut`, `IconClipboardPaste` | Fluent `copy` / `cut` / `clipboard-paste` (24 regular) — the standard edit-menu row icons |

## [0.3.4] - 2026-08-06

### Fixed

| Fix | Notes |
|---|---|
| **`ScrollArea` could not be scrolled by touch** | it consumed `pointer.Scroll` only — WHEEL events — so a finger drag moved nothing and the docs (and any app using it) were unscrollable on every phone and tablet. It now runs on `gesture.Scroll`, the same primitive `layout.List` uses, which reads wheel, drag and fling; the scroll range is still honoured, so leftover scroll chains to the parent, and a fling keeps frames coming until it settles |

Nothing yet.

## [0.3.3] - 2026-08-06

### Fixed

| Fix | Notes |
|---|---|
| **HoverCard / Tooltip / menu triggers panicked inside a `Card`** (regression in 0.3.2) | 0.3.2 made a throwaway pass report "no site" as `-1`, but these four index their per-site state BY that number (`trigs[idx]`, `over[idx]`, `rows[i]`), so measuring one inside a Card, Grid or ButtonGroup panicked with `index out of range [-1]` — which kills the Gio program, and in wasm shows as an endless "Go program has already exited" as dead callbacks keep firing on hover. A throwaway pass now measures the trigger alone and touches no site state. `Select` and `Popover` were never affected: they only compare the index |

Nothing yet.

## [0.3.2] - 2026-08-06

`go.mod` now carries `retract [v0.4.0, v0.8.0]`. Those versions are an
ABANDONED lineage — this history was squashed and re-tagged from
v0.1.0, but the proxy caches the pre-squash tags forever, and they are
a different codebase (no `item.go`, no `ScrollArea`). Resolving one
reads as a downgrade with missing API. **Always upgrade to an explicit
tag, never `@latest`:** cmd/go only honours retractions from the
module's highest version, so while the ghost v0.8.0 remains the
highest this directive is inert and `@latest` still lands on it.

### Added

| Symbol | Notes |
|---|---|
| `MeasurePass(gtx, w) Dimensions` | the supported way to size something before painting it: lays `w` out into a throwaway buffer, consumes no events, and claims no floating site — so a `Select`, `Popover` or menu inside `w` still opens in the live pass. `Card`, `CodeBlock`, `Example`, `Grid`, `SimpleGrid` and `ButtonGroup` all route their measuring through it |
| `IconNavigation` | mono `mdi:menu` ("navigation") — the hamburger, for folding a nav column into a drawer below a breakpoint |

### Fixed

| Fix | Notes |
|---|---|
| **Floating panels never opened inside a `Card`** | a `Select` (or `Popover`, or `DropdownMenuTrigger`) inside a `Card` toggled but painted nothing, while the same widget outside a card worked. Consumers paint only at floating "site 0" so one shared widget cannot stack N panels; `Card` lays its content out twice per frame (measure, then paint) with the same `gtx.Now`, and the DISCARDED pass was claiming site 0, leaving the live pass looking like a duplicate. Throwaway passes now announce themselves and claim no site. Same class fixed in `Grid`, `SimpleGrid` (whose scratch passes also drained the frame's events) and `ButtonGroup`'s superseded passes |
| `CodeBlock` painted tofu boxes for indentation | Go source is tab-indented and the embedded font has no glyph for U+0009, so every indent level rendered as ▯. Tabs now expand to four spaces FOR DISPLAY in both the highlighted (`Lines`) and `Plain` paths; the Copy button still writes the original tabs, which is what belongs in a .go file |

## [0.3.1] - 2026-08-05

### Fixed

| Fix | Notes |
|---|---|
| `Button` outline radius on pills | the fill was clipped to the computed radius (a capsule when `Rounded`) while the border always stroked 8dp corners, so an outline pill showed square corners inside a round fill. The border now follows the same clamped radius |

A note for consumers painting their own chrome: a rounded rect whose
CORNER RADIUS EXCEEDS HALF its shorter side is degenerate — Gio's
`widget.Border` will stroke that broken path across the whole window,
not just the widget. Use `ClampCorner(r, size)` (added in 0.2.0) for
clip radii, and never pass a "very large" constant like 999 as a
`CornerRadius` to get a pill: pass half the height.

## [0.3.0] - 2026-08-05

### Added

- `Scrollbar` / `ScrollbarProps` / `ScrollbarVariant` (`ScrollbarHover`,
  `ScrollbarAlways`) — macOS-style overlay thumb for the Scroll Area
  family (shadcn `ScrollBar` composition). Props: `Size` (shared enum,
  MD ≈ 6dp), `Color` / `Scheme`, `ShowTrack`, `Horizontal`.
  `Scrollbar.Layout` for a custom track box; returns content-fraction
  drag/click delta.
- `ScrollAreaProps.Scrollbar` / `NoScrollbar` — ScrollArea paints the
  overlay by default (hover fade); tune or hide per call.

### Changed

- `ScrollArea` shows a draggable overlay scrollbar (hover/scroll wake,
  fade-out) instead of a bare wheel-only viewport.

### Fixed

- `ScrollArea` no longer overscrolls: wheel/trackpad deltas are bounded
  to the remaining scrollable distance when consumed, instead of being
  added raw and clamped only after the content painted (one frame past
  the end, then a snap back — an elastic glitch). Leftover scroll at
  either end now chains to the parent scroller. No API change.

## [0.2.0] - 2026-08-05

### Changed

- Docs Performance page: driven by `site/bench.json` (layout→ops medians,
  frame-budget %, ship sizes with raw + gzip for WASM). Desktop peer
  binaries (Fyne / egui) only — no DOM-toolkit ship-size bakeoff.
- `ButtonGroup` stretches every child to one shared cross-axis size
  (shadcn `items-stretch`) so split/icon segments match the bar height
  instead of sitting as a shorter centered chip.
- `Button` expands to `Constraints.Min` and centers its label/icon —
  required for the group stretch to paint full-height chrome.
- `Card` lays content out without `op.Record` (scratch measure for chrome
  size, then live layout) so Floating inside a card escapes correctly —
  same footgun class as List-backed scroll.
- `Scrollable` docs: prefer `ScrollArea` when content hosts Floating;
  `Scrollable` remains the material.List scrollbar path.

### Added

- `lotusui bench-doc` / `make bench-doc` — refresh `site/bench.json` from
  `go test -bench` medians for the Performance docs page; optional
  `-wasm` records gallery size. Expanded component frame benches in
  `bench_test.go`.
- `Input.Attached` / `Select.Attached` (`AttachedEdges`) — square
  neighboring corners and drop the seat shadow so a field/trigger can
  fuse into a `ButtonGroup` (shadcn ButtonGroup + Input / Select).
- `ButtonGroupItem.Input` / `Hint`, `ButtonGroupInput(in, hint, flex)` —
  field slot with auto-`Attached` (prefer over a bare `Widget` slot).
- `ButtonGroupItem.Select`, `ButtonGroupSelect(sel, flex)` — Select
  trigger slot with auto-`Attached`.
- `ScrollArea` / `ScrollAreaProps` — shadcn Scroll Area: whole-content
  viewport scroll **without** `layout.List`'s `op.Record` macros, so
  `Floating` portals (Select, DropdownMenu, Popover, Tooltip, HoverCard)
  paint on the root ops stack instead of being trapped inside the
  scroller. Prefer over `Scrollable` when content hosts Floating; keep
  `ListView` for long virtualized collections. Methods: `Reset`,
  `ScrollTo`, `Layout`, `LayoutWith`. Field `Horizontal bool` (not
  `layout.Axis` — Axis's zero is Horizontal and would make bare
  `ScrollArea{}` sideways-scroll). `ScrollAreaProps.NoShadowRoom` for
  panes that already budget Card pad.
- `CodeBlock` / `CodeBlockProps` / `CodeSpan` — fenced source chrome
  (caption + Copy + pre body). Highlighting is **not** bundled: pass
  `Lines [][]CodeSpan` from your tokenizer (docsapp uses chroma) so the
  module zip stays lean; `Plain` is the unstyled fallback.
- `Example` / `ExampleProps` — Preview|Code chrome (bordered card with
  integrated tab strip). Pair Code with `CodeBlock{Nested: true}`. Not
  Tabs — recipe/docs surface for live + source side by side.

### Fixed

- Hover fade no longer flashes a darker mud mid-transition: outline /
  ghost / row fills use alpha-only fade (`fadeNRGBA`, Gio's
  `f32color.MulAlpha` model) instead of lerping RGB from transparent
  black. The hover clock also primes on the first frame (no invented
  16ms dt) so the ease starts from the resting color.

## [0.1.0] - 2026-08-05

Initial public release of lotusui — Gio design-system components,
theme (palette / radius / space / duration), registry CLI
(`add` / `update` / `skills` / `theme` / `icons` / `verify`), and the
nested docs site.

See `api.txt` for the exported API baseline and the docs site for
per-component usage.
