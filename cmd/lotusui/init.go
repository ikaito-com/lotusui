package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// cmdInit scaffolds a minimal themed app: the Gio window loop, a
// theme, a button — compiling in under a minute instead of after an
// afternoon of event-loop boilerplate. It also wires the go:generate
// lines for icons and theming so the build-time workflow is
// discoverable from day one.
func cmdInit(args []string) error {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	dir := fs.String("dir", ".", "directory to scaffold into")
	fs.Parse(args)

	path := filepath.Join(*dir, "main.go")
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%s already exists — refusing to overwrite", path)
	}
	if err := os.MkdirAll(*dir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(initTemplate), 0o644); err != nil {
		return err
	}
	if err := writeAgentsMD(*dir); err != nil {
		return err
	}
	fmt.Printf(`  %s written. Next:

    go mod init <your-module>   # if you haven't
    go mod tidy
    go run .

  Then make it yours:

    go run github.com/ikaito-com/lotusui/cmd/lotusui theme -anchor '#319795' -pkg main -o theme_gen.go
    go run github.com/ikaito-com/lotusui/cmd/lotusui icons -manifest icons/manifest.txt -out icons tabler:rocket
`, path)
	return nil
}

// writeAgentsMD writes (or appends to) AGENTS.md so any AI coding
// agent working in the consuming repo discovers the build-time
// workflow — the discoverability an MCP server would provide, with
// zero infrastructure. Existing files only gain the section if it
// isn't already there.
func writeAgentsMD(dir string) error {
	path := filepath.Join(dir, "AGENTS.md")
	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if strings.Contains(string(existing), "## lotusui") {
		return nil // already documented
	}
	body := string(existing)
	if body != "" && !strings.HasSuffix(body, "\n") {
		body += "\n"
	}
	if body != "" {
		body += "\n"
	}
	fmt.Printf("  %s written (agent instructions for the lotusui workflow)\n", path)
	return os.WriteFile(path, []byte(body+agentsSection), 0o644)
}

const agentsSection = `## lotusui

This app's UI is built on lotusui (github.com/ikaito-com/lotusui).
Its assets and theme are GENERATED at development time by the lotusui
CLI — run via ` + "`go run`" + `, so it needs no install and always matches
the lotusui version in go.mod.

Rules:

- Never edit ` + "`*_gen.go`" + ` files — edit the source (the icons manifest,
  theme.json) and run ` + "`go generate ./...`" + `.
- To add an icon: browse https://icon-sets.iconify.design, copy its
  ref, then ` + "`go run github.com/ikaito-com/lotusui/cmd/lotusui icons -manifest icons/manifest.txt -out icons -gen icons_gen.go -genpkg main <set:icon>`" + `.
  Commit the SVG, the manifest, and the regenerated constants together.
- To change the look: edit theme.json, run ` + "`go generate ./...`" + `.
  The generator validates WCAG contrast; take its warnings seriously.
- ` + "`go run github.com/ikaito-com/lotusui/cmd/lotusui verify …`" + ` is the
  offline drift check — run it (or ` + "`make check`" + `) before committing.
- Builds and runtime never need the network; if something seems to,
  that's a bug.
- When upgrading lotusui: read its CHANGELOG.md top-to-bottom for the
  versions you crossed — it records every rename, removal and
  signature change with exact symbols. Apply them, then
  ` + "`go build ./...`" + ` until clean.
`

const initTemplate = `// A minimal lotusui app: themed window, page background, one button.
//
// Build-time workflow (run with go generate ./...):
//
//go:generate go run github.com/ikaito-com/lotusui/cmd/lotusui icons -manifest icons/manifest.txt -out icons -gen icons_gen.go -genpkg main
package main

import (
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"

	lotusui "github.com/ikaito-com/lotusui"
)

func main() {
	go func() {
		w := new(app.Window)
		w.Option(app.Title("hello lotusui"))
		if err := loop(w); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func loop(w *app.Window) error {
	th := lotusui.NewTheme()
	th.UpgradeShaperAsync(w.Invalidate)
	var btn widget.Clickable
	var ops op.Ops
	for {
		switch e := w.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			th.ApplyPendingShaper()
			if btn.Clicked(gtx) {
				log.Println("clicked")
			}
			lotusui.Fill(gtx, th.Palette.Bg)
			lotusui.LayoutPage(th, gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.UniformInset(th.Space.LG).Layout(gtx,
					lotusui.VStack(th.Space.MD,
						lotusui.LabelHero(th, "Hello, lotusui").Layout,
						lotusui.LabelBody(th, "Edit main.go and go run . again.").Layout,
						lotusui.Button(th, &btn, "Click me", lotusui.ButtonProps{}),
					))
			})
			e.Frame(gtx.Ops)
		}
	}
}
`
