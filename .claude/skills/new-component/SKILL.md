---
name: new-component
description: Build a new lotusui component the best-of-both-worlds way — shadcn defines existence/naming/page anatomy, chakra adds breadth, our color engine underneath — with the full same-commit checklist.
---

# New lotusui component

Follow this ritual whenever a component is added (or substantially
reworked). The doctrine lives in CLAUDE.md "Grounding"; this is the
procedure.

## 1. Ground it (before writing code)

- Ensure `_ref/shadcn-ui` exists (see `_ref/README.md` / CLAUDE.md).
  Read the real example sources:
  `_ref/shadcn-ui/apps/v4/registry/**/examples/<name>-*.tsx` and
  `…/ui/<name>.tsx` — the marketing page alone is not enough.
- WebFetch the shadcn page first: `https://ui.shadcn.com/docs/components/base/<name>`
  (fall back to `/docs/components/<name>`). shadcn is the PRIMARY
  reference: it defines the component's existence, NAME, variants and
  the docs page's section list. No judgment calls against it.
- WebFetch the chakra page (`https://chakra-ui.com/docs/components/<name>`)
  for capability BREADTH shadcn is silent on (sizes, invalid states,
  slots). Chakra never decides structure.
- Components in neither catalog are lotusui EXTENSIONS: first-class,
  documented as such, same page anatomy.
- Web-only concepts (asChild/composition, aria, forms, RTL) become
  explicit "from the web" notes — that counts as parity because it is
  auditable. Everything else lands or gets a ⏳ row in site/PARITY.md.
- DEMO FIDELITY (hard rule): demos REPRODUCE shadcn's example
  compositions and content (their login card, their copy, their
  layout) — never invent alternative example content. The component
  must look and behave like shadcn; color is the only divergence.

## 2. Build it (the mechanisms)

- One FILE per family (`<name>.go`), named exports carrying the family
  prefix (`FooItem` lives with `Foo`, never bolted elsewhere).
- Enrichment decision rules: paints differently → Props field; adds
  structure → wrapper / nil-able widget slot; adds behavior → caller
  state + func fields; floating layer → the `Floating` portal
  primitive; web-only → documented note.
- Color: variants resolve role schemes from the theme; per-instance
  color is `Color ColorScale` (never a raw lone color — the ladder
  must derive); `Scheme *Scheme` wins over both.
- Sizes: the shared `Size` enum, Size2XS…Size2XL, MD default — add a
  row to every size switch the component grows.
- Interaction state is read INSIDE the Clickable.Layout closure
  (events drain at the top of Layout; reading before it renders one
  event late). Interaction colors walk the scale steps, never
  arithmetic. Keyboard focus renders the FocusRing token.
- Everything resolves at NewTheme — nothing per frame; hot paths stay
  zero-alloc (pin with a benchmark when in doubt).

## 3. Ship it (all in the SAME commit)

- [ ] `<name>.go` in the library root.
- [ ] Registry entry in `cmd/lotusui/registry.go` components table,
      then `go run ./cmd/lotusui registry`.
- [ ] Gallery demo (`site/gallery/main.go`): one `section(...)` per
      capability, indexes matching the docs page's `Demo: "<slug>/N"`.
      One state struct per on-screen instance — identity is the
      struct. Floating demos: the docs box must be tall enough to
      CONTAIN the open panel (DemoH).
- [ ] Docs page (`site/gen/pages.go`): installSection first, then
      Usage, then shadcn's sections in shadcn's order, then our
      extras; nav entry in alphabetical Components order.
- [ ] Smoke-test entry in `layout_test.go`.
- [ ] site/PARITY.md row (✅/⏳/➖/🧩 per capability).
- [ ] CHANGELOG.md [Unreleased]: every new/changed symbol, exact.
- [ ] `go run ./cmd/lotusui api -o api.txt` to accept the baseline.
- [ ] `make check` green; `cd site && make build` green.
- [ ] Consumers (vaultalia-simple, app-mandarin) still build.
