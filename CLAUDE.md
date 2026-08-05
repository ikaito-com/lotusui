# CLAUDE.md — lotusui

lotusui is the ikaito design language for Gio (gioui.org) desktop
apps — components, theme, motion, the box-grammar primitives —
extracted from vltl (`~/_ikaito/vaultalia-simple`), which consumes it
together with rtl via a `replace` directive while both sides are in
active development. `make check` (fmt + vet + test) must pass before
committing.

## Git authorship — NON-NEGOTIABLE (never violate)

Human maintainers own every git identity on this repository. AI
systems (Claude, Cursor Agent, Cursor Bot, Copilot, ChatGPT, or any
other automated assistant) MUST NEVER appear in git history or on
GitHub as an author, committer, tagger, releaser, co-author, or
contributor.

When an agent is instructed to run git actions (commit, tag, release,
push, deploy, amend, merge, cherry-pick, or anything that writes git
objects or GitHub release metadata):

1. **Author / committer / tagger** MUST be the human maintainer's
   configured identity only (`git config user.name` /
   `user.email` — today: aquiseb). Do NOT set or override identity to
   an AI name, bot email, or `noreply` AI account. Do NOT invent a
   second identity.
2. **Forbidden trailers and metadata:** never add `Co-authored-by:`,
   `Signed-off-by:`, or similar lines naming Claude, Cursor, an
   agent, a model, or any AI product. Never use GitHub's
   "committed via" / bot association for these actions.
3. **Commits, tags, releases, deploys:** every commit message, annotated
   tag, GitHub Release, and deploy trigger must attribute solely to
   the human. If a tool would inject AI attribution, refuse that path
   and use a plain human-authored commit/tag instead.
4. **Verify before push:** `git log -1 --format='%an <%ae>%n%cn <%ce>'`
   and the full message (`git log -1 --format=%B`) MUST show only the
   human — no `Co-authored-by` (or similar) naming Cursor, Claude, or
   any AI. For tags: `git show -s --format=fuller <tag>`. If anything
   AI-shaped appears, do not push — fix or redo the object first.
5. **Cursor injection:** the IDE may append
   `Co-authored-by: Cursor <cursoragent@cursor.com>` even when not
   requested. Strip it by rebuilding with `git commit-tree` (or
   equivalent) before push.
6. **This rule outranks convenience.** Prefer failing the git step
   over shipping history that lists an AI. History rewrites after the
   fact are worse than stopping.

Also mirrored in `.cursor/rules/git-authorship.mdc` (alwaysApply).

## The boundary (do not drift)

lotusui NEVER imports an app package: no Backend, no store, no tf —
nothing that knows about records, projects or AWS. Anything generic
enough for two apps belongs here; anything app-shaped stays in the
consumer. The UI doctrine itself (box grammar, row grammar, fixed
words) lives with the app that applies it (`vaultalia-simple/
.claude/ui.md`); this repo owns the MECHANISMS.

## Grounding: shadcn/ui × chakra-ui — best of both worlds (lessons paid for, do not relearn)

lotusui is PUBLISHED. The reference model is a deliberate HYBRID —
with a clear PRECEDENCE: shadcn/ui is the PRIMARY reference. If a
component exists in shadcn's catalog (including its base/ pages, e.g.
Field), shadcn defines its existence, name, variants and docs-page
anatomy — no judgment call, no "chakra has a page for this" framing.
chakra is consulted to ADD capabilities shadcn is silent on, never to
decide structure. Check both references when touching any component
and take each world's strength:

LOCAL SOURCE CHECKOUT (mandatory for agents): do NOT audit or build
from the marketing site alone. Keep a shallow clone at
`_ref/shadcn-ui` (gitignored — see `_ref/README.md`). Read the real
example files under
`_ref/shadcn-ui/apps/v4/registry/**/examples/<component>-*.tsx` and
the ui source under `…/ui/<component>.tsx` so behavior (e.g. a
Breadcrumb ellipsis that opens a DropdownMenu) is never guessed.
Refresh with `git -C _ref/shadcn-ui pull --ff-only` (or re-clone).
Optional: `_ref/chakra-ui` for breadth.

FROM SHADCN (ui.shadcn.com) — vocabulary, structure, distribution:
- COMPONENT NAMES and VARIANTS: Button variants default/secondary/
  destructive/outline/ghost/link (`ButtonDefault`…`ButtonLink`), Badge
  default/secondary/destructive/outline, DropdownMenu (not Menu),
  Dialog (not Modal), Input, Select, Tabs, Card, Checkbox, Switch.
  Destructive is a VARIANT, not a color prop.
- DOCS STRUCTURE: per-component pages shaped like shadcn's (install →
  usage → one section PER capability, each with its own live demo) —
  but we KEEP our colorful identity (palette picker, look picker,
  Midnight dark theme). Structure from shadcn, skin ours.
- DEMO FIDELITY (hard rule — violated repeatedly before it was
  written down; never again): a docs example REPRODUCES the shadcn
  page's example — the same COMPOSITION, the same example CONTENT and
  copy (the login card, "Accept terms and conditions", the invoices
  table, "m@example.com"), the same layout and proportions. Do NOT
  invent alternative example content. Components must also LOOK and
  BEHAVE like shadcn (metrics, chrome, micro-interactions); the ONLY
  sanctioned divergence is color, which comes from our palette. When
  writing or auditing a page, open the shadcn page and transpose its
  examples one by one — EVERY example on the shadcn page must exist
  here. For lotusui EXTENSIONS (no shadcn counterpart), design the
  examples shadcn WOULD have shown: same composition style, same
  realistic content register, same restraint.
- SCREENSHOTS: the README and docs-homepage images are captured from
  REAL components — the showcase scenes in site/gallery/showcase.go,
  rendered by `make -C site media` into site/media/. After ANY visual
  change, re-run media (and goldens) so every marketing image stays
  truthful; never hand-edit or mock these images.
- THE OWNERSHIP MODEL (the registry): default consumption is the
  module import; `lotusui add <component>` vendors source into the app
  (AST-qualified against the core, unexported helpers carried into a
  CLI-owned companion), `lotusui update` does a TRUE three-way merge —
  base and remote reconstructed by re-running the vendoring transform
  on pristine sources from the Go module cache. INVARIANTS: the
  transform (`vendorSet`) is a PURE function of (source version,
  component SET, target package) — never add nondeterminism to it; the
  vendored dir is coherent as a SET (components reference each other's
  local copies); the companion file is CLI-owned, regenerated, never
  merged. registry.json is generated (`lotusui registry`, go:generate)
  and drift-checked (`verify -registry`); the CLI and AI agents
  consume it — app code NEVER reads it at runtime. Skills ship in
  `skills/lotusui/` and install via `lotusui skills`; blocks live in
  `registry/blocks/` and use only the exported API.

FROM CHAKRA (chakra-ui.com) — breadth and the color system:
- FEATURE BREADTH: where shadcn is silent (loading buttons, sizes,
  invalid states, input Start/End slots), chakra's component page
  defines the capability set; site/PARITY.md is the ledger.
- SIZES are RICHER than shadcn: Size2XS…Size2XL (seven), MD default,
  shared enum across all components.
- THE COLOR ENGINE (ours, keep it): variants carry role schemes from
  the theme's paired tokens (shadcn's --primary/--primary-foreground
  discipline). Per-instance color = `Color ColorScale` (chakra's
  colorPalette power): ONE anchor in, the WHOLE interaction ladder out
  (.500 base → .600 hover → .700 pressed) — a specified color can
  never break hover/pressed. `Scheme *Scheme` wins over Color: full
  manual slot control. Never accept a raw lone color prop that would
  bypass the ladder. Badge Color uses SoftScheme (pastel + deep ink —
  the status doctrine holds for any color).
- Dark mode is A PALETTE, not a mode: DefaultDarkPalette, applied via
  WithPalette; both themes built at startup, pointer swap.

Component enrichment decision rules (unchanged):
- paints differently → Props field; adds structure → wrapper/nil-able
  widget slot; adds behavior → caller state + func fields; floating
  layer → the shared portal primitive (once built); web-only →
  documented "from the web" note.

NAMING, one root per concept (do not blur them again):
- `XxxProps` is a component's per-call prop bag — variant, size,
  color, icons, behavior flags. The zero value is always valid.
- `XxxOption` is ONE CHOICE in a list, `.Options` the field holding
  them, `XxxOpts(labels…)` the label-only constructor (SelectOpts,
  RadioOpts, TabOpts, ToggleOpts). Choices are addressed by VALUE;
  the cursor stays unexported.

The permanent rules, each learned the hard way:
- Dotted web tokens flatten to CamelCase (`fg.muted` → `FgMuted`) —
  a Go field can't be both a color and a namespace.
- FAMILIES: a piece belongs to its component family — DropdownMenuItem
  lives in menu.go with DropdownMenu and its docs page, never bolted
  onto Button. One file + one docs page per family.
- NOTHING APP-SPECIFIC, ever: no domain rules, no app composites, no
  app vocabulary in docs examples. Generic mechanisms here; domain
  policy in the consumer. When in doubt, it goes to the app.
- INTERACTION STATES walk the color scale via Scheme's Hover/Active
  fields — never ad-hoc arithmetic (shade() is only the fallback for
  hand-built schemes). Keyboard focus renders the FocusRing token.
- BUILD TIME over runtime: assets fetched+normalized and code
  generated by `cmd/lotusui` (go:generate + `lotusui verify` guard);
  theming resolves at NewTheme, never per frame; hot paths stay
  zero-alloc with benchmarks pinning it (`bench_test.go`).
- EVERY component change ships IN THE SAME COMMIT with: its docs page
  (site/gen/pages.go), its gallery demo (site/gallery), a smoke-test
  entry (layout_test.go), registry.json regenerated when files moved,
  and consumer updates (vaultalia-simple, app-mandarin) when the API
  moved.
- LOTUSUI EXTENSIONS: components beyond shadcn's catalog (Grid,
  SimpleGrid, Split, Stack, ListView, Field) are first-class and
  documented as extensions — never dropped for the sake of parity.
  For a component chakra HAS but shadcn LACKS, chakra's page defines
  the names, props and capability sections; shadcn still defines the
  DOCS PAGE STRUCTURE (install → usage → one section per capability)
  and the component is vendorable through the registry like any other.

## Gio version moves with the consumers (lesson paid for)

Go's MVS takes the MAX gioui.org requirement across lotusui and its
consumer apps, so a casual `go mod tidy` here can silently upgrade
the apps' gio. Bump gio DELIBERATELY, in this repo and the consumers
in the same change, and re-verify the seamless macOS window with a
real launch + screenshot: `macchrome_darwin.go` pairs with
`app.Decorated(false)` in the consumers' mains — on gio >= v0.8 that
pairing is what stops Gio painting fallback decorations over the
chrome, and it hides the traffic lights that our ObjC re-shows.

## Releases (versioning, changelog, migrations)

SemVer via Go modules: v0.x while the PARITY ledger has gaps —
breaking changes allowed in minors; v1.0.0 when it's clean, after
which breaking means /v2. Commits follow Conventional Commits
(feat:/fix:/feat!: + BREAKING CHANGE footers). CHANGELOG.md is
written FOR AI AGENTS in consuming apps — it must record EVERY
API-visible create/update/rename/remove with exact symbols, old→new
signatures and replacement guidance, precise enough that an agent
migrates an app from it alone (`go build ./...` is the safety net).
EVERY API change lands in the same commit with its CHANGELOG entry
and both consumer apps updated. There is deliberately NO migrate
command: agents read the changelog and do the work.

THE COMMIT RITUAL (api.txt is the tripwire — make check enforces it).
Authorship: see § "Git authorship — NON-NEGOTIABLE" above — human
identity only; no AI co-authors, committers, or taggers.
1. Make the change. `make check` now FAILS on API drift, listing the
   exact added/removed symbols.
2. Record every one of them in CHANGELOG.md [Unreleased] under the
   right section (Renamed table / Removed-with-replacement / Changed
   signatures / Added), exact symbols, old→new.
3. `go run ./cmd/lotusui api -o api.txt` to accept the new baseline.
4. Update both consumer apps if the change touched them.
5. Commit code + changelog + api.txt + consumers TOGETHER, message in
   Conventional Commits (`feat:`/`fix:`; breaking = `feat!:` plus a
   `BREAKING CHANGE:` footer naming the symbols).
6. Releasing: `make release` (optionally BUMP=minor|patch|major; the
   default bump is inferred from the changelog sections). It validates
   [Unreleased] is non-empty and drift checks pass, then retitles the
   changelog, bumps version.go, rotates site/versions.json (new
   version at "/" and previous root archived at "/vPREV/" for every
   release — including patches — so the docs switcher always lists the
   latest tag; CI builds archives from their tags), refreshes api.txt,
   and prints the exact git commit/tag/push commands to finish by hand.

## Icons

Fluent SVGs via Iconify: add a line to `assets/icons/manifest.txt`,
run `make icons` (network, Iconify API), commit both the manifest and
the fetched SVGs — builds never need the network. A build-failing
test asserts every icon rasterizes. Never add a second icon path.

## The documentation website — site/

The docs site lives IN THIS REPO as a NESTED MODULE, and that choice
carries the invariants:

- `site/` gets its OWN `go.mod` (module
  `github.com/ikaito-com/lotusui/site`, with
  `replace github.com/ikaito-com/lotusui => ..`). A nested module is
  EXCLUDED from the parent module's zip — consumers of the library
  download NOTHING of the site. This is the whole reason it may live
  here; breaking the nesting breaks every consumer's download weight.
- Everything heavy belongs under `site/`: prose, screenshots, media.
  The LIBRARY ROOT stays lean forever — its files ship in every
  consumer's module zip, immutably, per version.
- The site is Chakra-style: per-component pages with prose + Go
  snippet + a LIVE demo. The demos are the real components compiled
  to WebAssembly (`GOOS=js GOARCH=wasm` — a first-class Gio target):
  ONE gallery app routing to a component/state by URL hash, compiled
  ONCE — never a separate wasm *binary* per component (each bundle is
  several MB with fonts embedded). Each docs Preview is a gallery
  iframe for that hash (`loading="lazy"`); the prose carries SEO, the
  iframe carries the demo.
- The gallery app doubles as the component test bed — the browser
  twin of vltl's `VLTL_PREVIEW` state harness: every component gets
  addressable states, screenshot-able without click choreography.
- Built WASM bundles are NEVER committed: CI builds the site and
  deploys the static output (GitHub Pages / Cloudflare Pages). The
  repo holds sources only.
- The site documents the code it sits next to: a component change
  and its docs page update in the SAME commit.
