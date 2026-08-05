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

### Added

- `IconGithub` — Simple Icons GitHub mark (`simple-icons:github`), for
  mono tinting via `SVGIcon` like the other Fluent mono icons.
- `Input.Attached` / `Select.Attached` (`AttachedEdges`) — square
  neighboring corners and drop the seat shadow so a field/trigger can
  fuse into a `ButtonGroup` (shadcn ButtonGroup + Input / Select).
- `ButtonGroupItem.Input` / `Hint`, `ButtonGroupInput(in, hint, flex)` —
  field slot with auto-`Attached` (prefer over a bare `Widget` slot).
- `ButtonGroupItem.Select`, `ButtonGroupSelect(sel, flex)` — Select
  trigger slot with auto-`Attached`.

### Changed

- `ButtonGroup` stretches every child to one shared cross-axis size
  (shadcn `items-stretch`) so split/icon segments match the bar height
  instead of sitting as a shorter centered chip.
- `Button` expands to `Constraints.Min` and centers its label/icon —
  required for the group stretch to paint full-height chrome.

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
