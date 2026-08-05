---
name: lotusui
description: How to build UI with lotusui, the Gio design system — components, theming with schemes and color scales, the registry (add/update vendoring), and how to consume its changelog when upgrading.
---

# Building UI with lotusui

lotusui (`github.com/ikaito-com/lotusui`) is a design system for Gio
(gioui.org) — desktop, mobile and web from one Go codebase. Immediate
mode: components are functions returning `layout.Widget`; interactive
state lives in YOUR structs (`widget.Clickable`, component structs) and
identity is the struct pointer — never share one state struct between
two on-screen instances.

## Components

Component vocabulary follows the web convention: Button, Badge, Card, Checkbox, Switch,
Input, Field, Select, Tabs, Dialog, DropdownMenu — plus lotusui
extensions beyond the usual web catalogs: Grid, SimpleGrid, Split, Stack,
ListView.

- Variants (Button): `ButtonDefault` (solid primary), `ButtonSecondary`,
  `ButtonDestructive`, `ButtonOutline`, `ButtonGhost`, `ButtonLink`.
  Badge: `BadgeDefault`, `BadgeSecondary`, `BadgeDestructive`,
  `BadgeOutline`.
- Sizes: `Size2XS SizeXS SizeSM SizeMD (default)
  SizeLG SizeXL Size2XL`, shared across components.
- Color: three layers, most-derived wins —
  1. default: the variant's role scheme from the theme,
  2. `Color: lotusui.Teal` (any `ColorScale`, incl. `ScaleFrom(hex)`) —
     one anchor in, the full interaction ladder out (.500 base, .600
     hover, .700 pressed); a custom color can never break hover states,
  3. `Scheme: &s` — full manual slot control (`Teal.SoftScheme()` etc.).

## Theming

Everything resolves at `NewTheme(opts...)` — never per frame. Tokens
live on `th.Palette` (Bg/Fg/Border ladders, Brand slots, status pairs),
radii on `th.Radius`, spacing on `th.Space`. Dark mode is only a
palette: `NewTheme(WithPalette(lotusui.DefaultDarkPalette))`; build both
themes at startup and swap the pointer. Generate an app palette from a
brand color at BUILD time: `go run github.com/ikaito-com/lotusui/cmd/lotusui
theme -anchor '#319795' -pkg ui -o theme_gen.go`.

## The registry (ownership model)

Default consumption is the module import. When a component must
diverge, take ownership of a copy:

    go run github.com/ikaito-com/lotusui/cmd/lotusui add button -dir ui

The copy still imports the lotusui core (references are qualified
automatically); it is stamped `// lotusui:vendored <name> v<ver>
sha256:<pristine-hash>`. NEVER remove or edit the stamp line. To pull
upstream improvements later:

    go run github.com/ikaito-com/lotusui/cmd/lotusui update -dir ui

Untouched files are replaced cleanly; customized files get a true
three-way merge (base reconstructed from the Go module cache) with
standard conflict markers, and the CHANGELOG sections between the two
versions are printed for you to resolve conflicts with full context.

`registry.json` at the module root is the machine-readable catalog
(components, blocks, skills, dependency and carried-helper info). It is
a build-time artifact for you and the CLI — app code must never read it
at runtime.

## Upgrading lotusui

CHANGELOG.md is written FOR agents: every release section lists ALL
created/renamed/changed/removed symbols in tables with exact old → new
names. On `go get` upgrades, read the sections between the old and new
version and apply the renames mechanically. `lotusui update` prints the
same delta for vendored components.

## Icons

Add Iconify refs to `assets/icons/manifest.txt` in the app, run the
icons command (network happens at DEV time only), commit manifest +
SVGs + generated constants. A typo'd icon name is a compile error.
