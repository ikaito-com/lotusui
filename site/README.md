# site/ — the lotusui documentation website

A nested Go module (excluded from the parent module's zip, so library
consumers download nothing of it). The docs are a **Gio/lotusui app**
(`docsapp/`) compiled once to WebAssembly for GitHub Pages — chrome,
nav, Previews, and Code tabs are real widgets, not an HTML shell with
gallery iframes.

```
docsapp/     the Gio docs app (WASM deploy + native: make run-docsapp)
docspages/   page model + prose/snippets (shared content)
live/        addressable component demos (docsapp Previews)
gallery/     screenshot / golden harness (still used by make media)
cmd/serve/   tiny static server for dist/
dist/        built output (gitignored — CI builds and deploys)
```

## Working on it

```sh
make serve        # build docsapp WASM → dist/, serve on :3030
make build        # same as docsapp (Pages artifact)
make run-docsapp  # native window (PAGE=button by default)
make check        # fmt + vet + test + native/wasm builds
```

Adding a component page: add the demo to `live/demos.go` (addressable
slug), the page to `docspages/pages.go` (or pages2.go), nav entry in
`docspages/nav.go`, in the SAME commit as the component change.

Deploy: `make build` writes a thin `index.html` + `docsapp.wasm` +
`wasm_exec.js` + `media/` heroes at the site root. CI
(`.github/workflows/site.yml`) publishes `site/dist` to GitHub Pages.
WASM is never committed — sources only.
