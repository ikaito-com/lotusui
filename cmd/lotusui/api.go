package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"sort"
	"strings"
)

// cmdAPI writes the exported-API inventory (api.txt): one line per
// exported symbol with its signature. The committed baseline is the
// versioning tripwire — `lotusui verify` fails when the live API no
// longer matches it, which forces every API change through
// CHANGELOG.md before `make check` can pass again.
func cmdAPI(args []string) error {
	fs := flagSet("api")
	dir := fs.String("dir", ".", "package directory to inventory")
	out := fs.String("o", "", "output file (default stdout)")
	fs.Parse(args)

	inv, err := apiInventory(*dir)
	if err != nil {
		return err
	}
	body := strings.Join(inv, "\n") + "\n"
	if *out == "" {
		fmt.Print(body)
		return nil
	}
	fmt.Printf("  api -> %s (%d symbols)\n", *out, len(inv))
	return os.WriteFile(*out, []byte(body), 0o644)
}

// apiInventory renders the package's exported surface, sorted and
// deterministic: funcs and methods with signatures, types with their
// exported fields, exported consts and vars with types.
func apiInventory(dir string) ([]string, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, 0)
	if err != nil {
		return nil, err
	}
	var lines []string
	add := func(format string, node any) {
		var b bytes.Buffer
		printer.Fprint(&b, fset, node)
		// One line per symbol keeps diffs readable.
		sig := strings.Join(strings.Fields(b.String()), " ")
		lines = append(lines, fmt.Sprintf(format, sig))
	}
	for name, pkg := range pkgs {
		if name == "main" {
			continue
		}
		for _, f := range pkg.Files {
			for _, decl := range f.Decls {
				switch d := decl.(type) {
				case *ast.FuncDecl:
					if !d.Name.IsExported() || (d.Recv != nil && !exportedRecv(d.Recv)) {
						continue
					}
					d.Body = nil
					d.Doc = nil
					add("%s", d)
				case *ast.GenDecl:
					d.Doc = nil
					for _, spec := range d.Specs {
						switch sp := spec.(type) {
						case *ast.TypeSpec:
							if !sp.Name.IsExported() {
								continue
							}
							stripUnexportedFields(sp)
							sp.Doc, sp.Comment = nil, nil
							add("type %s", sp)
						case *ast.ValueSpec:
							sp.Doc, sp.Comment = nil, nil
							for _, n := range sp.Names {
								if !n.IsExported() {
									continue
								}
								kind := "var"
								if d.Tok == token.CONST {
									kind = "const"
								}
								typ := ""
								if sp.Type != nil {
									var b bytes.Buffer
									printer.Fprint(&b, fset, sp.Type)
									typ = " " + strings.Join(strings.Fields(b.String()), " ")
								}
								lines = append(lines, kind+" "+n.Name+typ)
							}
						}
					}
				}
			}
		}
	}
	sort.Strings(lines)
	return lines, nil
}

func exportedRecv(recv *ast.FieldList) bool {
	if len(recv.List) == 0 {
		return false
	}
	t := recv.List[0].Type
	if star, ok := t.(*ast.StarExpr); ok {
		t = star.X
	}
	if id, ok := t.(*ast.Ident); ok {
		return id.IsExported()
	}
	return false
}

// stripUnexportedFields removes unexported struct fields from the
// inventory — internal layout changes are not API changes.
func stripUnexportedFields(sp *ast.TypeSpec) {
	st, ok := sp.Type.(*ast.StructType)
	if !ok || st.Fields == nil {
		return
	}
	var kept []*ast.Field
	for _, f := range st.Fields.List {
		exported := len(f.Names) == 0 // embedded
		for _, n := range f.Names {
			if n.IsExported() {
				exported = true
			}
		}
		if exported {
			f.Doc, f.Comment = nil, nil
			kept = append(kept, f)
		}
	}
	st.Fields.List = kept
}
