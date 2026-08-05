# App icons — how they get here

The UI's icons are SVGs downloaded at development time from the
[Iconify API](https://icon-sets.iconify.design/) (one URL scheme over
every major icon set — we use the **Fluent Color** set for the
full-color look) and **embedded into the binary** via `go:embed`. They
are app assets, not runtime data: they ship inside the binary, never
live in S3, and need no network at runtime.

## Adding or changing an icon

1. Find the icon on https://icon-sets.iconify.design/ and note its
   `set:name` (e.g. `fluent-color:apps-32`).
2. Add a line to `manifest.txt`: `<set>:<name> <output-file>.svg`.
3. Run `make icons` from the repo root — it downloads every manifest
   entry into this directory (curl -f: a typo'd name fails the build
   step loudly instead of embedding a 404 page).
4. Commit both the manifest line and the downloaded SVG, so a checkout
   builds without the network.

## Rendering

`internal/ui/iconsvg.go` rasterizes these SVGs at the needed size and
tint (any `currentColor` stroke is replaced with a palette color before
rasterizing; full-color sets keep their own palette, with gradients
flattened to solid stops), cached per name+size+color. Use
`SVGIcon(name, size, tint)` — never add a second rendering path.

## License

Fluent UI System Color Icons are MIT (Microsoft). If you pull from
another set via Iconify, check that set's license before committing it.
