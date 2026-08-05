# site/ — the lotusui documentation website

A nested Go module (excluded from the parent module's zip, so library
consumers download nothing of it). Static HTML carries the prose and
SEO; the live demos are the real components compiled ONCE to
WebAssembly — one gallery app, one bundle, addressed by URL hash;
each docs Preview is its own gallery iframe (`loading="lazy"`).

```
gen/      the static site generator: pages as data (pages.go),
          html/template layouts, the lotusui-palette stylesheet
gallery/  the ONE Gio demo app — every documented state is reachable
          by hash (#button, #modal/open, …); natively it doubles as
          the component test bed: LOTUSUI_GALLERY=modal/open go run ./gallery
dist/     built output (gitignored — CI builds and deploys)
```

## Working on it

```sh
make serve   # build everything, serve on :3030
make build   # dist/ only
make check   # fmt + vet + both build targets (native and js/wasm)
```

Adding a component page: add the demo to `gallery/main.go` (a `demo`
entry with an addressable slug), the page to `gen/pages.go`, in the
SAME commit as the component change it documents.

Deploys from `.github/workflows/site.yml` to GitHub Pages on push to
main. All links are relative, so it works from any subpath.
