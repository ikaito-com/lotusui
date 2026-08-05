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
