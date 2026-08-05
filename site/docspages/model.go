// Package docspages is the shared documentation content model — one
// source for the Gio docs app (site/docsapp).
package docspages

// Group is one sidebar group (Get started, Components, …).
type Group struct {
	Title string
	Pages []*Page
}

// Page is one docs page: install → usage → one section per capability.
type Page struct {
	Slug      string
	Title     string
	Kicker    string
	Platforms []string
	Intro     string // HTML fragment
	Sections  []Section
	Props     []Prop
}

// Section is one capability block (heading, prose, snippet, live demo).
type Section struct {
	Heading   string
	Platforms []string
	Prose     string // HTML fragment
	Snippet   string
	Lang      string // chroma lexer; "" = go
	Demo      string // gallery/live hash ("button/2"); "" = no demo
	DemoH     int    // preview height hint (px); 0 = default
}

// Prop is one API table row.
type Prop struct {
	Name string
	Type string
	Desc string
}

// InstallSection is the per-component install block.
func InstallSection(component string) Section {
	return Section{
		Heading: "Installation",
		Prose: `<p>Use the module import (the default), or own the source: <code>lotusui add</code>
vendors the component into your app and <code>lotusui update</code> keeps the copy mergeable —
see <a href="../registry/">Registry</a>.</p>`,
		Snippet: `go get github.com/ikaito-com/lotusui

# or vendor the source into your app:
go run github.com/ikaito-com/lotusui/cmd/lotusui add ` + component,
	}
}

// Flat returns pages in sidebar order.
func Flat(groups []Group) []*Page {
	var out []*Page
	for _, g := range groups {
		out = append(out, g.Pages...)
	}
	return out
}

// Lookup finds a page by slug and returns it, its group title, and prev/next.
func Lookup(groups []Group, slug string) (p *Page, group string, prev, next *Page) {
	flat := Flat(groups)
	for i, pg := range flat {
		if pg.Slug != slug {
			continue
		}
		p = pg
		if i > 0 {
			prev = flat[i-1]
		}
		if i+1 < len(flat) {
			next = flat[i+1]
		}
		for _, g := range groups {
			for _, x := range g.Pages {
				if x.Slug == slug {
					group = g.Title
					return
				}
			}
		}
		return
	}
	return
}
