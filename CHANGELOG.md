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

## [0.1.0] - 2026-08-05

Initial public release of lotusui — Gio design-system components,
theme (palette / radius / space / duration), registry CLI
(`add` / `update` / `skills` / `theme` / `icons` / `verify`), and the
nested docs site.

See `api.txt` for the exported API baseline and the docs site for
per-component usage.
