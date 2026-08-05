// Command gen builds the static lotusui docs site into -out: one HTML
// page per component (prose + live demo + highlighted Go snippet), an
// index, and the gallery shell that loads the single wasm bundle. The
// prose is plain HTML — that's what carries SEO; the canvas only
// carries the demo; each Preview is a gallery iframe addressed by hash
// nears the viewport. Syntax highlighting happens HERE, at build time
// (chroma) — the pages ship zero highlighting JavaScript.
package main

import (
	"bytes"
	"embed"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"encoding/json"

	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/yuin/goldmark"
	gmext "github.com/yuin/goldmark/extension"

	"github.com/ikaito-com/lotusui/site/looks"
)

//go:embed templates/*.html
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS

type site struct {
	Name string
	// BaseURL is the canonical public root (with trailing slash) —
	// canonical links, OG images, the sitemap all derive from it.
	BaseURL  string
	Tag      string
	Repo     string
	Groups   []group
	Palettes []palettePreset
	Looks    []lookPreset
	// Version is the release these docs document (first entry of
	// versions.json); Versions is the full switcher list.
	Version  string
	Versions []versionEntry
}

type versionEntry struct {
	Version string `json:"version"`
	Path    string `json:"path"`
}

// loadVersions reads the committed versions manifest — newest first,
// the root deployment always serving the newest.
func loadVersions() ([]versionEntry, error) {
	b, err := os.ReadFile("versions.json")
	if err != nil {
		return nil, err
	}
	var v []versionEntry
	if err := json.Unmarshal(b, &v); err != nil {
		return nil, err
	}
	return v, nil
}

// renderMarkdown converts a markdown file (the changelog) to HTML for
// embedding in a docs page — GFM so the migration tables render.
func renderMarkdown(path string) (template.HTML, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	md := goldmark.New(goldmark.WithExtensions(gmext.GFM))
	var buf bytes.Buffer
	if err := md.Convert(b, &buf); err != nil {
		return "", err
	}
	return template.HTML(buf.String()), nil
}

// lookPreset is one look-and-feel picker entry, resolved to CSS-ready
// metadata (the wasm side reads the same presets from site/looks).
type lookPreset struct {
	Slug, Name, Hint    string
	CSSFamily, FontFile string
	RadiusMD, RadiusLG  int
}

// palettePreset is one picker entry, resolved to CSS-ready colors.
type palettePreset struct {
	Slug, Name                         string
	BrandSolid, BrandFg, BrandContrast string
	BrandSubtle                        string
	Bg                                 string // page canvas tint
}

type group struct {
	Title string
	Pages []*page
}

type page struct {
	Slug      string // directory under components/
	Title     string
	Kicker    string // one line under the title, and on index cards
	Platforms []string
	Intro     template.HTML
	Sections  []section
	Props     []prop // optional API table, rendered as the last section
}

type section struct {
	Heading   string
	Platforms []string // platform chips next to the heading; empty = universal (no chips)
	Prose     template.HTML
	Snippet   string // source, shown highlighted below the demo
	Lang      string // chroma lexer name; "" = go
	Demo      string // gallery hash ("button", "modal/open"); "" = no demo
	DemoH     int    // Preview demobox height in px; 0 = 340
}

// A prop row documents one field or parameter, docs props-table.
type prop struct {
	Name string
	Type string
	Desc string
}

type pageData struct {
	Site       *site
	Canonical  string
	Page       *page
	Group      string // the group the page belongs to (breadcrumb)
	Prev, Next *page
	Root       string // relative prefix back to the site root
}

func main() {
	out := flag.String("out", "dist", "output directory")
	serve := flag.String("serve", "", "serve the output directory on this address (e.g. :3030) after building")
	flag.Parse()

	s := theSite()
	if vs, err := loadVersions(); err == nil && len(vs) > 0 {
		s.Versions = vs
		s.Version = vs[0].Version
	}
	if err := build(s, *out); err != nil {
		log.Fatal(err)
	}
	fmt.Println("site generated in", *out)

	if *serve != "" {
		url := "http://" + *serve
		if strings.HasPrefix(*serve, ":") {
			url = "http://localhost" + *serve
		}
		fmt.Println("serving on", url)
		log.Fatal(http.ListenAndServe(*serve, http.FileServer(http.Dir(*out))))
	}
}

// slugify turns a section heading into its anchor id.
func slugify(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z' || r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == ' ' || r == '-':
			b.WriteByte('-')
		}
	}
	return strings.Trim(b.String(), "-")
}

// highlight renders source through chroma with CSS classes — the token
// colors live in style.css, mapped to the palette.
func highlight(src, lang string) (template.HTML, error) {
	if lang == "" {
		lang = "go"
	}
	lexer := lexers.Get(lang)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	it, err := lexer.Tokenise(nil, src)
	if err != nil {
		return "", err
	}
	f := html.New(html.WithClasses(true))
	var buf bytes.Buffer
	// The style argument is required but irrelevant with classes on.
	if err := f.Format(&buf, styles.Get("github"), it); err != nil {
		return "", err
	}
	return template.HTML(buf.String()), nil
}

func build(s *site, out string) error {
	funcs := template.FuncMap{
		"demoH": func(h int) int {
			if h == 0 {
				return 340
			}
			// The example boxes keep a shared minimum height so pages
			// read as an even rhythm; content centers inside.
			if h < 140 {
				return 140
			}
			return h
		},
		"slug":      slugify,
		"lower":     strings.ToLower,
		"highlight": highlight,
		// demoURL builds the gallery link as a trusted URL: the demo
		// hash contains "/" (slug/section), which the auto-escaper
		// would otherwise mangle into %2f — breaking the wasm router.
		"relpath": func(p string) string {
			// versions.json paths are site-absolute ("/", "/v0.1.0/");
			// the version switcher uses "/" + relpath so links work from
			// archived /vX.Y.Z/ trees (Root-relative would nest wrong).
			return strings.TrimPrefix(p, "/")
		},
		"demoURL": func(root, demo string) template.URL {
			return template.URL(root + "gallery/#" + demo)
		},
		"langLabel": func(l string) string {
			if l == "" {
				return "Go"
			}
			return l
		},
	}
	tpl, err := template.New("").Funcs(funcs).ParseFS(templateFS, "templates/*.html")
	if err != nil {
		return err
	}

	writePage := func(rel, tplName string, data any) error {
		path := filepath.Join(out, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer f.Close()
		return tpl.ExecuteTemplate(f, tplName, data)
	}

	// Index + one page per component, threaded with prev/next in
	// reading order (the sidebar order, flattened).
	if err := writePage("index.html", "index.html", pageData{Site: s, Root: "", Canonical: s.BaseURL}); err != nil {
		return err
	}
	type flat struct {
		p     *page
		group string
	}
	var order []flat
	for _, g := range s.Groups {
		for _, p := range g.Pages {
			order = append(order, flat{p, g.Title})
		}
	}
	for i, f := range order {
		data := pageData{Site: s, Page: f.p, Group: f.group, Root: "../../",
			Canonical: s.BaseURL + "components/" + f.p.Slug + "/"}
		if i > 0 {
			data.Prev = order[i-1].p
		}
		if i < len(order)-1 {
			data.Next = order[i+1].p
		}
		rel := filepath.Join("components", f.p.Slug, "index.html")
		if err := writePage(rel, "page.html", data); err != nil {
			return err
		}
	}
	// The gallery shell: loads wasm_exec.js + gallery.wasm, which the
	// Makefile builds into the same directory.
	if err := writePage(filepath.Join("gallery", "index.html"), "gallery.html", pageData{Site: s, Root: "../"}); err != nil {
		return err
	}

	// The switcher manifest ships with the site so archived versions
	// can be listed from one place.
	if vs, err := os.ReadFile("versions.json"); err == nil {
		if err := os.WriteFile(filepath.Join(out, "versions.json"), vs, 0o644); err != nil {
			return err
		}
	}

	// Discoverability: sitemap + robots derive from the page list.
	var sm strings.Builder
	sm.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	sm.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")
	sm.WriteString("  <url><loc>" + s.BaseURL + "</loc></url>\n")
	for _, g := range s.Groups {
		for _, p := range g.Pages {
			sm.WriteString("  <url><loc>" + s.BaseURL + "components/" + p.Slug + "/</loc></url>\n")
		}
	}
	sm.WriteString("</urlset>\n")
	if err := os.WriteFile(filepath.Join(out, "sitemap.xml"), []byte(sm.String()), 0o644); err != nil {
		return err
	}
	robots := "User-agent: *\nAllow: /\nSitemap: " + s.BaseURL + "sitemap.xml\n"
	// The custom domain rides in the artifact so a Pages deploy can
	// never drop it.
	if err := os.WriteFile(filepath.Join(out, "CNAME"), []byte("lotusui.com\n"), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(out, "robots.txt"), []byte(robots), 0o644); err != nil {
		return err
	}

	// Showcase media (README + homepage screenshots) ships with the
	// site — regenerated by `make media` from the real components.
	if entries, err := os.ReadDir("media"); err == nil {
		if err := os.MkdirAll(filepath.Join(out, "media"), 0o755); err != nil {
			return err
		}
		for _, e := range entries {
			b, err := os.ReadFile(filepath.Join("media", e.Name()))
			if err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(out, "media", e.Name()), b, 0o644); err != nil {
				return err
			}
		}
	}

	// The registry manifest ships too — AI agents pointed at the docs
	// can read the component catalog from /registry.json. Generated at
	// build time from the library source; never read by app code.
	if rj, err := os.ReadFile("../registry.json"); err == nil {
		if err := os.WriteFile(filepath.Join(out, "registry.json"), rj, 0o644); err != nil {
			return err
		}
	}

	// Performance snapshot ships with the site so agents/readers can
	// re-check the numbers without re-running benches.
	if bj, err := os.ReadFile("bench.json"); err == nil {
		if err := os.WriteFile(filepath.Join(out, "bench.json"), bj, 0o644); err != nil {
			return err
		}
	}

	// The favicon ships from the site itself — fetched once from
	// Iconify (twemoji:lotus) and committed; pages never hotlink.
	if fav, err := staticFS.ReadFile("static/favicon.svg"); err == nil {
		if err := os.WriteFile(filepath.Join(out, "favicon.svg"), fav, 0o644); err != nil {
			return err
		}
	}
	// Topbar GitHub mark — same Iconify SVG the library embeds
	// (simple-icons:github via assets/icons/). Copied at gen time so
	// the static site and SVGIcon stay one source of truth.
	if gh, err := os.ReadFile("../assets/icons/github.svg"); err == nil {
		if err := os.WriteFile(filepath.Join(out, "github.svg"), gh, 0o644); err != nil {
			return err
		}
	}
	// Optional in-page bootstrap (unused by Preview iframes; kept for
	// experiments / standalone tooling).
	if emb, err := staticFS.ReadFile("static/gallery-embed.js"); err == nil {
		if err := os.WriteFile(filepath.Join(out, "gallery-embed.js"), emb, 0o644); err != nil {
			return err
		}
	}

	// Static assets, plus the GENERATED palette overrides: each preset
	// becomes a data-palette CSS block — pre-built at compile time, the
	// same way an app ships multiple lotusui themes.
	css, err := staticFS.ReadFile("static/style.css")
	if err != nil {
		return err
	}
	var pcss strings.Builder
	pcss.WriteString("\n/* generated palette presets */\n")
	for _, p := range s.Palettes {
		fmt.Fprintf(&pcss, ":root[data-palette=%q] { --brand: %s; --brand-subtle: %s; --brand-ink: %s; --on-brand: %s; --page: %s; }\n",
			p.Slug, p.BrandSolid, p.BrandSubtle, p.BrandFg, p.BrandContrast, p.Bg)
	}
	// Look presets: @font-face for each committed font, then one CSS
	// block per look — the same axis the wasm applies via WithFaces/
	// WithRadius, worn by the static pages.
	pcss.WriteString("\n/* generated look presets */\n")
	if err := os.MkdirAll(filepath.Join(out, "fonts"), 0o755); err != nil {
		return err
	}
	for _, l := range s.Looks {
		if l.FontFile != "" {
			b := looks.ByName(l.Slug).FontBytes()
			if err := os.WriteFile(filepath.Join(out, "fonts", l.FontFile), b, 0o644); err != nil {
				return err
			}
			fmt.Fprintf(&pcss, "@font-face { font-family: %q; src: url(\"fonts/%s\") format(\"truetype\"); font-weight: 100 900; font-display: swap; }\n",
				l.CSSFamily, l.FontFile)
		}
		fmt.Fprintf(&pcss, ":root[data-style=%q] { --font-body: %q; --radius-md: %dpx; --radius-lg: %dpx; }\n",
			l.Slug, l.CSSFamily, l.RadiusMD, l.RadiusLG)
	}
	return os.WriteFile(filepath.Join(out, "style.css"), append(css, []byte(pcss.String())...), 0o644)
}
