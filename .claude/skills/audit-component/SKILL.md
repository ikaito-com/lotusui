---
name: audit-component
description: Audit an existing lotusui component against its shadcn reference page — fetch the real page, diff section lists, implement gaps, note web-only, update the ledger.
---

# Audit a component against shadcn

Never audit from memory — read the local checkout, then the page.

0. Ensure `_ref/shadcn-ui` is present (CLAUDE.md / `_ref/README.md`).
   Open `_ref/shadcn-ui/apps/v4/registry/**/examples/<component>-*.tsx`
   and `…/ui/<component>.tsx` first — that is how each example
   actually behaves (e.g. ellipsis → DropdownMenu), not a guess from
   a screenshot.

1. WebFetch `https://ui.shadcn.com/docs/components/base/<name>` (or
   `/docs/components/<name>`), asking for the ordered section list and
   what each example demonstrates.
2. Diff against our page in `site/docspages/pages.go` (or pages2.go):
   - Missing FEATURE (a prop/behavior we lack) → implement it in the
     library, per the new-component decision rules.
   - Missing SECTION for an existing capability → add the section +
     a gallery demo state for it.
   - Composition example (their form/grid/group demos) → add a
     section composing our existing API; no new props for what
     composition already does.
   - Web-only (asChild, aria, RTL, form libraries) → "from the web"
     note in the page intro or a prose-only section.
   - Genuinely deferred → ⏳ in site/PARITY.md with a one-line reason.
3. Match shadcn's SECTION ORDER for the shared sections; our extra
   capabilities (Sizes 2XS–2XL, Color, Scheme…) follow after.
4. Visual behavior follows shadcn too (metrics, weights, chrome,
   micro-interactions like the 1dp press nudge); COLOR stays ours —
   everything paints from the palette/scales.
   DEMO FIDELITY: reproduce shadcn's example COMPOSITIONS and example
   CONTENT verbatim-in-spirit (their login card, their copy, their
   layout) — never invent alternative example content. Color is the
   only divergence.
5. Finish with the same-commit checklist from the new-component skill
   (demo indexes, PARITY, changelog, api.txt, registry, make check,
   site build, consumers).
