package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// The registry: lotusui's answer to shadcn's ownership model. A
// component is DATA — a named set of source files plus everything the
// CLI needs to vendor it into an app (see cmdAdd) and update it later
// with a real three-way merge (see cmdUpdate). registry.json is the
// machine-readable manifest of that data, generated at BUILD time by
// `lotusui registry` and committed; it is consumed by the CLI and by
// AI agents — never fetched or parsed by app code at runtime.

// componentSpec declares one vendorable component: which root-package
// files make it up. Everything else in the manifest (description,
// hashes, carried helpers, dependencies) is COMPUTED from the source
// by AST scan, so the table can never drift from the code.
type componentSpec struct {
	Name  string   // registry name — kebab-case, shadcn vocabulary
	Files []string // root-package files that ARE the component
}

// components is the vendorable catalog. Files not listed here are
// CORE — the runtime every vendored component still imports (theme,
// scales, icons, layout utilities). Order is the docs order.
var components = []componentSpec{
	{Name: "button", Files: []string{"button.go"}},
	{Name: "button-group", Files: []string{"buttongroup.go"}},
	{Name: "badge", Files: []string{"badge.go"}},
	{Name: "card", Files: []string{"card.go"}},
	{Name: "checkbox", Files: []string{"checkbox.go"}},
	{Name: "switch", Files: []string{"switch.go"}},
	{Name: "input", Files: []string{"inputs.go"}},
	{Name: "item", Files: []string{"item.go"}},
	{Name: "kbd", Files: []string{"kbd.go"}},
	{Name: "input-otp", Files: []string{"inputotp.go"}},
	{Name: "field", Files: []string{"field.go"}},
	{Name: "select", Files: []string{"select.go"}},
	{Name: "tabs", Files: []string{"tabs.go"}},
	{Name: "dialog", Files: []string{"dialog.go"}},
	{Name: "dropdown-menu", Files: []string{"menu.go"}},
	{Name: "accordion", Files: []string{"accordion.go"}},
	{Name: "alert", Files: []string{"alert.go"}},
	{Name: "alert-dialog", Files: []string{"alertdialog.go"}},
	{Name: "avatar", Files: []string{"avatar.go"}},
	{Name: "breadcrumb", Files: []string{"breadcrumb.go"}},
	{Name: "pagination", Files: []string{"pagination.go"}},
	{Name: "popover", Files: []string{"popover.go"}},
	{Name: "hover-card", Files: []string{"hovercard.go"}},
	{Name: "annotated-text", Files: []string{"glossary.go"}},
	{Name: "progress", Files: []string{"progress.go"}},
	{Name: "radio-group", Files: []string{"radio.go"}},
	{Name: "separator", Files: []string{"separator.go"}},
	{Name: "skeleton", Files: []string{"skeleton.go"}},
	{Name: "slider", Files: []string{"slider.go"}},
	{Name: "spinner", Files: []string{"spinner.go"}},
	{Name: "table", Files: []string{"table.go"}},
	{Name: "textarea", Files: []string{"textarea.go"}},
	{Name: "toast", Files: []string{"toast.go"}},
	{Name: "toggle", Files: []string{"toggle.go"}},
	{Name: "tooltip", Files: []string{"tooltip.go"}},
	// lotusui extensions — components beyond the shadcn catalog.
	{Name: "grid", Files: []string{"grid.go"}},
	{Name: "listview", Files: []string{"listview.go"}},
	{Name: "split", Files: []string{"split.go", "split_scroll.go"}},
	{Name: "stack", Files: []string{"stack.go"}},
}

// blocks are ready-made compositions (shadcn's "blocks"): app-level
// source that uses only lotusui's exported API, vendored by copy.
var blocks = []componentSpec{
	{Name: "login-form", Files: []string{"registry/blocks/login/login.go"}},
}

// registryEntry is one component's manifest record.
type registryEntry struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"` // "component" | "block"
	Description string   `json:"description"`
	Files       []string `json:"files"`
	// Carried lists the unexported core helpers the vendored copy
	// carries along (they can't be import-qualified).
	Carried []string `json:"carried,omitempty"`
	// DependsOn lists other registry components this one references.
	DependsOn []string `json:"dependsOn,omitempty"`
	// Hash fingerprints the component's pristine vendored form at this
	// version — `lotusui update` compares against it to detect local
	// customization.
	Hash string `json:"hash"`
}

type registryDoc struct {
	Schema     string          `json:"$schema,omitempty"`
	Name       string          `json:"name"`
	Version    string          `json:"version"`
	GoModule   string          `json:"goModule"`
	Components []registryEntry `json:"components"`
	Blocks     []registryEntry `json:"blocks"`
	Skills     []string        `json:"skills"`
}

// cmdRegistry generates registry.json from the source of truth — the
// code itself. Run via go:generate; `lotusui verify -registry` guards
// the committed copy against drift.
func cmdRegistry(args []string) error {
	fs := flagSet("registry")
	dir := fs.String("dir", ".", "lotusui module root")
	out := fs.String("o", "registry.json", "output manifest path")
	fs.Parse(args)

	data, err := buildRegistry(*dir)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(*dir, *out), data, 0o644); err != nil {
		return err
	}
	fmt.Printf("  wrote %s (%d components, %d blocks)\n", *out, len(components), len(blocks))
	return nil
}

func buildRegistry(moduleDir string) ([]byte, error) {
	pkg, err := loadRootPackage(moduleDir)
	if err != nil {
		return nil, err
	}
	version, err := moduleVersion(moduleDir)
	if err != nil {
		return nil, err
	}

	doc := registryDoc{
		Name:     "lotusui",
		Version:  version,
		GoModule: "github.com/ikaito-com/lotusui",
		Skills:   []string{"skills/lotusui/SKILL.md"},
	}
	for _, c := range components {
		v, err := vendorComponent(pkg, c, "lotusui", version)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", c.Name, err)
		}
		doc.Components = append(doc.Components, registryEntry{
			Name:        c.Name,
			Type:        "component",
			Description: componentDoc(pkg, c),
			Files:       c.Files,
			Carried:     v.carried,
			DependsOn:   componentDeps(pkg, c),
			Hash:        v.hashOf(c.Name),
		})
	}
	for _, b := range blocks {
		src, err := os.ReadFile(filepath.Join(moduleDir, b.Files[0]))
		if err != nil {
			return nil, fmt.Errorf("block %s: %w", b.Name, err)
		}
		doc.Blocks = append(doc.Blocks, registryEntry{
			Name:        b.Name,
			Type:        "block",
			Description: firstDocSentence(string(src)),
			Files:       b.Files,
			Hash:        hash16(src),
		})
	}
	return json.MarshalIndent(doc, "", "  ")
}

// verifyRegistry reports drift between registry.json and the source.
func verifyRegistry(moduleDir string) []string {
	want, err := buildRegistry(moduleDir)
	if err != nil {
		return []string{fmt.Sprintf("registry: cannot rebuild manifest: %v", err)}
	}
	have, err := os.ReadFile(filepath.Join(moduleDir, "registry.json"))
	if err != nil {
		return []string{"registry.json missing — run `go run ./cmd/lotusui registry`"}
	}
	if !bytes.Equal(bytes.TrimSpace(want), bytes.TrimSpace(have)) {
		return []string{"registry.json drifted from the source — run `go run ./cmd/lotusui registry`"}
	}
	return nil
}

// ── Package model ────────────────────────────────────────────────────

// rootPackage is the parsed lotusui root package: every file's AST +
// source bytes, and the package-level definition index.
type rootPackage struct {
	dir     string
	fset    *token.FileSet
	files   map[string]*ast.File  // basename → AST
	src     map[string][]byte     // basename → source
	defs    map[string]string     // package-level name → defining basename
	methods map[string][]methodOf // receiver type name → its methods
}

type methodOf struct {
	file string
	decl *ast.FuncDecl
}

// recvTypeName unwraps a method receiver to its base type name.
func recvTypeName(recv *ast.FieldList) string {
	if recv == nil || len(recv.List) == 0 {
		return ""
	}
	t := recv.List[0].Type
	for {
		switch tt := t.(type) {
		case *ast.StarExpr:
			t = tt.X
		case *ast.IndexExpr: // generic receiver
			t = tt.X
		case *ast.Ident:
			return tt.Name
		default:
			return ""
		}
	}
}

func loadRootPackage(dir string) (*rootPackage, error) {
	pkg := &rootPackage{
		dir:     dir,
		fset:    token.NewFileSet(),
		files:   map[string]*ast.File{},
		src:     map[string][]byte{},
		defs:    map[string]string{},
		methods: map[string][]methodOf{},
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		src, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		f, err := parser.ParseFile(pkg.fset, name, src, parser.ParseComments)
		if err != nil {
			return nil, err
		}
		pkg.files[name] = f
		pkg.src[name] = src
		for _, d := range f.Decls {
			switch d := d.(type) {
			case *ast.FuncDecl:
				if d.Recv == nil {
					pkg.defs[d.Name.Name] = name
				} else if r := recvTypeName(d.Recv); r != "" {
					// Methods travel with their receiver type: carrying
					// an unexported type carries its whole method set.
					pkg.methods[r] = append(pkg.methods[r], methodOf{file: name, decl: d})
				}
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					switch s := spec.(type) {
					case *ast.TypeSpec:
						pkg.defs[s.Name.Name] = name
					case *ast.ValueSpec:
						for _, n := range s.Names {
							pkg.defs[n.Name] = name
						}
					}
				}
			}
		}
	}
	return pkg, nil
}

// ── The vendoring transform ──────────────────────────────────────────
//
// vendorSet is a PURE function of (source tree, component SET, target
// package): it rewrites the set's files to live inside an app —
// package clause renamed, exported core identifiers qualified
// `lotusui.`, unexported core helpers gathered into ONE CLI-owned
// companion file shared by the whole set. The SET is the unit of
// coherence: components vendored side by side reference each other's
// local types (never a mix of local and qualified), and shared
// helpers exist exactly once. Determinism is the load-bearing
// property: `lotusui update` re-runs this transform on two pristine
// versions to reconstruct the exact base and remote of a three-way
// merge, so the app's own edits — and only those — surface as diffs.

// carriedFile is the shared companion's basename; its stamp uses the
// pseudo-component name below. The file is REGENERATED, never merged
// — it belongs to the CLI, not the app.
const (
	carriedFile      = "lotusui_carried.go"
	carriedComponent = "lotusui-carried"
)

type vendored struct {
	files   map[string][]byte   // every stamped output file, incl. the companion
	own     map[string][]string // component → the basenames it owns
	hashes  map[string]string   // component → pristine hash of its own files
	carried []string            // carried helper names, sorted
}

// forComp views one component's slice of a set result — used by the
// registry manifest, which documents components individually.
func (v *vendored) hashOf(comp string) string { return v.hashes[comp] }

func vendorSet(pkg *rootPackage, comps []componentSpec, targetPkg, version string) (*vendored, error) {
	inSet := map[string]bool{}
	ownerOf := map[string]string{} // basename → component name
	for _, c := range comps {
		for _, f := range c.Files {
			if pkg.files[f] == nil {
				return nil, fmt.Errorf("%s: file %s not found in package", c.Name, f)
			}
			inSet[f] = true
			ownerOf[f] = c.Name
		}
	}

	// Closure over unexported helpers: scan the set's files, carry
	// what they need, then scan what was carried, until stable.
	carried := map[string]bool{}
	var scanQueue []scanUnit
	for _, c := range comps {
		for _, f := range c.Files {
			scanQueue = append(scanQueue, scanUnit{file: f, node: pkg.files[f]})
		}
	}
	type qual struct {
		file   string
		offset int
	}
	var quals []qual
	needsImport := map[string]bool{}
	for len(scanQueue) > 0 {
		u := scanQueue[0]
		scanQueue = scanQueue[1:]
		refs := scanRefs(pkg, u)
		for _, r := range refs {
			def := pkg.defs[r.name]
			if inSet[def] {
				continue // defined inside the set — stays as-is
			}
			if ast.IsExported(r.name) {
				quals = append(quals, qual{file: u.file, offset: r.offset})
				needsImport[u.file] = true
				continue
			}
			if !carried[r.name] {
				carried[r.name] = true
				decl, declFile, err := findDecl(pkg, r.name)
				if err != nil {
					return nil, err
				}
				scanQueue = append(scanQueue, scanUnit{file: declFile, node: decl, carriedName: r.name})
				// A carried type brings its methods; their bodies are
				// scanned too so THEIR dependencies join the closure.
				for _, m := range pkg.methods[r.name] {
					scanQueue = append(scanQueue, scanUnit{file: m.file, node: m.decl, carriedName: r.name})
				}
			}
		}
	}

	// Group qualification splices per file, apply back-to-front.
	byFile := map[string][]int{}
	for _, q := range quals {
		byFile[q.file] = append(byFile[q.file], q.offset)
	}

	out := map[string][]byte{}
	for _, c := range comps {
		for _, f := range c.Files {
			src := append([]byte(nil), pkg.src[f]...)
			src = spliceQuals(src, byFile[f])
			src = renamePackage(src, targetPkg)
			if needsImport[f] {
				src = addLotusuiImport(src)
			}
			out[f] = src
		}
	}

	carriedNames := make([]string, 0, len(carried))
	for n := range carried {
		carriedNames = append(carriedNames, n)
	}
	sort.Strings(carriedNames)
	var companion []byte
	if len(carriedNames) > 0 {
		var err error
		companion, err = buildCompanion(pkg, carriedNames, byFile, targetPkg)
		if err != nil {
			return nil, err
		}
	}

	// Fingerprint each component's own files (stamps excluded so the
	// hash is reproducible from content alone), then stamp.
	v := &vendored{files: map[string][]byte{}, own: map[string][]string{}, hashes: map[string]string{}, carried: carriedNames}
	for _, c := range comps {
		files := append([]string(nil), c.Files...)
		sort.Strings(files)
		h := sha256.New()
		for _, f := range files {
			h.Write([]byte(f))
			h.Write(out[f])
		}
		hash := hex.EncodeToString(h.Sum(nil))[:16]
		v.own[c.Name] = files
		v.hashes[c.Name] = hash
		for _, f := range files {
			stamp := fmt.Sprintf("// Code vendored by `lotusui add %s`. Yours to edit.\n// lotusui:vendored %s v%s sha256:%s\n\n", c.Name, c.Name, version, hash)
			v.files[f] = append([]byte(stamp), out[f]...)
		}
	}
	if companion != nil {
		stamp := fmt.Sprintf("// Code generated by `lotusui add` — core helpers shared by the vendored components. DO NOT EDIT.\n// lotusui:vendored %s v%s sha256:%s\n\n", carriedComponent, version, hash16(companion))
		v.files[carriedFile] = append([]byte(stamp), companion...)
	}
	return v, nil
}

// vendorComponent is the singleton-set view, used by the manifest.
func vendorComponent(pkg *rootPackage, comp componentSpec, targetPkg, version string) (*vendored, error) {
	return vendorSet(pkg, []componentSpec{comp}, targetPkg, version)
}

// scanUnit is one AST region to scan: a whole file, or one carried
// declaration inside a core file.
type scanUnit struct {
	file        string
	node        ast.Node
	carriedName string
}

type ref struct {
	name   string
	offset int
}

// scanRefs finds identifiers in the unit that resolve to package-level
// definitions, skipping every position where an identifier is a NAME
// being declared rather than a reference (field names, struct-literal
// keys, := targets, labels, selector .Sel).
func scanRefs(pkg *rootPackage, u scanUnit) []ref {
	skip := map[*ast.Ident]bool{}
	ast.Inspect(u.node, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.SelectorExpr:
			skip[n.Sel] = true
		case *ast.Field:
			for _, id := range n.Names {
				skip[id] = true
			}
		case *ast.FuncDecl:
			skip[n.Name] = true
		case *ast.TypeSpec:
			skip[n.Name] = true
		case *ast.ValueSpec:
			for _, id := range n.Names {
				skip[id] = true
			}
		case *ast.ImportSpec:
			if n.Name != nil {
				skip[n.Name] = true
			}
		case *ast.LabeledStmt:
			skip[n.Label] = true
		case *ast.BranchStmt:
			if n.Label != nil {
				skip[n.Label] = true
			}
		case *ast.KeyValueExpr:
			if id, ok := n.Key.(*ast.Ident); ok {
				skip[id] = true
			}
		case *ast.AssignStmt:
			if n.Tok == token.DEFINE {
				for _, lhs := range n.Lhs {
					if id, ok := lhs.(*ast.Ident); ok {
						skip[id] = true
					}
				}
			}
		case *ast.RangeStmt:
			if n.Tok == token.DEFINE {
				if id, ok := n.Key.(*ast.Ident); ok {
					skip[id] = true
				}
				if id, ok := n.Value.(*ast.Ident); ok {
					skip[id] = true
				}
			}
		}
		return true
	})
	var out []ref
	ast.Inspect(u.node, func(n ast.Node) bool {
		id, ok := n.(*ast.Ident)
		if !ok || skip[id] || id.Name == "_" {
			return true
		}
		if _, defined := pkg.defs[id.Name]; !defined {
			return true
		}
		out = append(out, ref{name: id.Name, offset: pkg.fset.Position(id.Pos()).Offset})
		return true
	})
	return out
}

// findDecl locates the top-level declaration of an unexported name.
func findDecl(pkg *rootPackage, name string) (ast.Node, string, error) {
	file, ok := pkg.defs[name]
	if !ok {
		return nil, "", fmt.Errorf("no package-level decl for %q", name)
	}
	for _, d := range pkg.files[file].Decls {
		switch d := d.(type) {
		case *ast.FuncDecl:
			if d.Recv == nil && d.Name.Name == name {
				return d, file, nil
			}
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					if s.Name.Name == name {
						return d, file, nil
					}
				case *ast.ValueSpec:
					for _, n := range s.Names {
						if n.Name == name {
							return d, file, nil
						}
					}
				}
			}
		}
	}
	return nil, "", fmt.Errorf("decl for %q not found in %s", name, file)
}

// buildCompanion assembles the carried-helpers file: package clause,
// the imports those helpers actually use, and each helper's source
// with core references qualified.
func buildCompanion(pkg *rootPackage, names []string, byFile map[string][]int, targetPkg string) ([]byte, error) {
	imports := map[string]string{} // path → local name ("" = default)
	needsLotusui := false
	var body bytes.Buffer
	// Each carried name emits its decl, then its methods (a carried
	// type travels with its whole method set).
	type emit struct {
		decl ast.Node
		file string
	}
	var emits []emit
	for _, name := range names {
		decl, file, err := findDecl(pkg, name)
		if err != nil {
			return nil, err
		}
		emits = append(emits, emit{decl, file})
		for _, m := range pkg.methods[name] {
			emits = append(emits, emit{m.decl, m.file})
		}
	}
	for _, e := range emits {
		decl, file := e.decl, e.file
		start := pkg.fset.Position(decl.Pos()).Offset
		end := pkg.fset.Position(decl.End()).Offset
		// Include the doc comment when it directly precedes the decl.
		if fd, ok := decl.(*ast.FuncDecl); ok && fd.Doc != nil {
			start = pkg.fset.Position(fd.Doc.Pos()).Offset
		} else if gd, ok := decl.(*ast.GenDecl); ok && gd.Doc != nil {
			start = pkg.fset.Position(gd.Doc.Pos()).Offset
		}
		src := pkg.src[file][start:end]
		// Qualify within this decl: reuse the file's recorded offsets
		// that fall inside [start, end).
		var local []int
		for _, off := range byFile[file] {
			if off >= start && off < end {
				local = append(local, off-start)
				needsLotusui = true
			}
		}
		src = spliceQuals(append([]byte(nil), src...), local)
		// Imports used by this decl, resolved from its home file.
		declImports(pkg, file, decl, imports)
		body.Write(src)
		body.WriteString("\n\n")
	}

	var out bytes.Buffer
	fmt.Fprintf(&out, "package %s\n\n", targetPkg)
	paths := make([]string, 0, len(imports))
	for p := range imports {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	out.WriteString("import (\n")
	if needsLotusui {
		out.WriteString("\tlotusui \"github.com/ikaito-com/lotusui\"\n")
	}
	for _, p := range paths {
		if n := imports[p]; n != "" {
			fmt.Fprintf(&out, "\t%s %q\n", n, p)
		} else {
			fmt.Fprintf(&out, "\t%q\n", p)
		}
	}
	out.WriteString(")\n\n")
	out.Write(body.Bytes())
	return out.Bytes(), nil
}

// declImports records which of the home file's imports a decl uses.
func declImports(pkg *rootPackage, file string, decl ast.Node, into map[string]string) {
	f := pkg.files[file]
	byName := map[string]*ast.ImportSpec{}
	for _, imp := range f.Imports {
		path, _ := strconv.Unquote(imp.Path.Value)
		name := ""
		if imp.Name != nil {
			name = imp.Name.Name
		} else {
			name = path[strings.LastIndex(path, "/")+1:]
		}
		byName[name] = imp
	}
	ast.Inspect(decl, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		id, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}
		if imp, hit := byName[id.Name]; hit {
			path, _ := strconv.Unquote(imp.Path.Value)
			local := ""
			if imp.Name != nil {
				local = imp.Name.Name
			}
			into[path] = local
		}
		return true
	})
}

// spliceQuals inserts "lotusui." before each recorded offset,
// back-to-front so earlier offsets stay valid.
func spliceQuals(src []byte, offsets []int) []byte {
	sort.Sort(sort.Reverse(sort.IntSlice(offsets)))
	prev := -1
	for _, off := range offsets {
		if off == prev {
			continue // same ident recorded twice
		}
		prev = off
		src = append(src[:off], append([]byte("lotusui."), src[off:]...)...)
	}
	return src
}

var pkgClauseRe = regexp.MustCompile(`(?m)^package lotusui$`)

func renamePackage(src []byte, target string) []byte {
	return pkgClauseRe.ReplaceAll(src, []byte("package "+target))
}

// addLotusuiImport wires the core import into an existing import block.
func addLotusuiImport(src []byte) []byte {
	const line = "\n\tlotusui \"github.com/ikaito-com/lotusui\"\n"
	if i := bytes.Index(src, []byte("import (")); i >= 0 {
		at := i + len("import (")
		return append(src[:at], append([]byte(line), src[at:]...)...)
	}
	// No import block (rare): add one after the package clause.
	loc := pkgClauseRe.FindIndex(src)
	if loc == nil {
		return src
	}
	block := []byte("\n\nimport (" + line + ")")
	return append(src[:loc[1]], append(block, src[loc[1]:]...)...)
}

// ── Manifest metadata helpers ────────────────────────────────────────

// componentDoc extracts the first sentence of the component's file doc.
func componentDoc(pkg *rootPackage, comp componentSpec) string {
	return firstDocSentence(string(pkg.src[comp.Files[0]]))
}

var docLineRe = regexp.MustCompile(`(?m)^// (.+)$`)

func firstDocSentence(src string) string {
	// First comment block after the package clause's vicinity: collect
	// leading // lines and cut at the first period.
	var b strings.Builder
	for _, m := range docLineRe.FindAllStringSubmatch(src, 12) {
		line := m[1]
		if strings.HasPrefix(line, "Code vendored") || strings.HasPrefix(line, "lotusui:") || strings.HasPrefix(line, "go:") {
			continue
		}
		b.WriteString(line)
		b.WriteString(" ")
		if strings.Contains(line, ".") {
			break
		}
	}
	s := b.String()
	if i := strings.Index(s, ". "); i >= 0 {
		s = s[:i+1]
	}
	return strings.TrimSpace(s)
}

// componentDeps lists other registry components this one references.
func componentDeps(pkg *rootPackage, comp componentSpec) []string {
	owner := map[string]string{} // defining file → component name
	for _, c := range components {
		for _, f := range c.Files {
			owner[f] = c.Name
		}
	}
	inSet := map[string]bool{}
	for _, f := range comp.Files {
		inSet[f] = true
	}
	deps := map[string]bool{}
	for _, f := range comp.Files {
		for _, r := range scanRefs(pkg, scanUnit{file: f, node: pkg.files[f]}) {
			def := pkg.defs[r.name]
			if !inSet[def] && owner[def] != "" && owner[def] != comp.Name {
				deps[owner[def]] = true
			}
		}
	}
	out := make([]string, 0, len(deps))
	for d := range deps {
		out = append(out, d)
	}
	sort.Strings(out)
	return out
}

// moduleVersion reads the Version constant from version.go — works in
// a checkout, a replace-directive dir, and the module cache alike.
func moduleVersion(moduleDir string) (string, error) {
	src, err := os.ReadFile(filepath.Join(moduleDir, "version.go"))
	if err != nil {
		return "", fmt.Errorf("version.go: %w", err)
	}
	m := regexp.MustCompile(`Version = "([0-9.]+)"`).FindSubmatch(src)
	if m == nil {
		return "", fmt.Errorf("version.go: no Version constant")
	}
	return string(m[1]), nil
}

func hash16(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])[:16]
}
