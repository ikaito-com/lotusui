package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// cmdAdd vendors a component's source INTO the app — shadcn's
// ownership model, in Go. The default way to use lotusui stays the
// module import (idiomatic, updated by `go get`); `lotusui add` is
// for the moment a component must diverge: the app takes ownership of
// a copy, edits it freely, and `lotusui update` can still merge
// upstream improvements later because every vendored file carries a
// version + pristine-hash stamp.
//
//	go run github.com/ikaito-com/lotusui/cmd/lotusui add button
//	go run github.com/ikaito-com/lotusui/cmd/lotusui add dialog -dir internal/ui -pkg ui
//
// The copy still imports the lotusui CORE (theme, scales, icons) —
// exported identifiers are qualified automatically, unexported
// helpers are gathered once into a CLI-owned companion file — so a
// vendored Button keeps taking *lotusui.Theme and drops into existing
// call sites. The vendored DIRECTORY is the unit of coherence: adding
// a component re-vendors the set, so components reference each
// other's local copies; your edits to already-vendored files are
// preserved by the same three-way merge `lotusui update` uses.
func cmdAdd(args []string) error {
	fs := flagSet("add")
	dir := fs.String("dir", "ui", "directory to vendor into (created if missing)")
	pkgName := fs.String("pkg", "", "target package name (default: the directory's basename)")
	force := fs.Bool("force", false, "overwrite files that already exist")
	fs.Parse(reorderFlags(args, map[string]bool{"force": true}))
	names := fs.Args()
	if len(names) == 0 {
		return fmt.Errorf("usage: lotusui add <component|block> [...] — see registry.json for the catalog")
	}
	if *pkgName == "" {
		*pkgName = filepath.Base(*dir)
	}

	moduleDir, err := lotusuiModuleDir(".")
	if err != nil {
		return err
	}
	version, err := moduleVersion(moduleDir)
	if err != nil {
		return err
	}
	pkg, err := loadRootPackage(moduleDir)
	if err != nil {
		return err
	}

	// The set already vendored here, from the stamps on disk.
	existing, _ := scanStamps(*dir) // missing dir = empty set
	var oldComps []componentSpec
	existingByName := map[string]stampInfo{}
	for _, st := range existing {
		existingByName[st.component] = st
		if c, ok := findComponent(st.component); ok {
			if st.version != version {
				return fmt.Errorf("%s is vendored at v%s but the module is v%s — run `lotusui update` first", st.component, st.version, version)
			}
			oldComps = append(oldComps, c)
		}
	}
	if len(existing) > 0 {
		*pkgName = filePackage(existing[0].files[0].path)
	}

	// Split the request into components (set semantics) and blocks
	// (self-contained copies).
	newComps := append([]componentSpec(nil), oldComps...)
	inNew := map[string]bool{}
	for _, c := range oldComps {
		inNew[c.Name] = true
	}
	var addedComps, addedBlocks []componentSpec
	for _, name := range names {
		if comp, ok := findComponent(name); ok {
			if !inNew[name] {
				inNew[name] = true
				newComps = append(newComps, comp)
				addedComps = append(addedComps, comp)
			}
			continue
		}
		if blk, ok := findBlock(name); ok {
			addedBlocks = append(addedBlocks, blk)
			continue
		}
		return fmt.Errorf("unknown component %q — see registry.json for the catalog", name)
	}

	if err := os.MkdirAll(*dir, 0o755); err != nil {
		return err
	}

	if len(addedComps) > 0 || len(oldComps) > 0 {
		newOut, err := vendorSet(pkg, newComps, *pkgName, version)
		if err != nil {
			return err
		}
		var oldOut *vendored
		if len(oldComps) > 0 && len(addedComps) > 0 {
			// The merge base for already-vendored files: the same
			// version under the OLD set.
			oldOut, err = vendorSet(pkg, oldComps, *pkgName, version)
			if err != nil {
				return err
			}
		}
		// Newly added components: plain writes.
		for _, c := range addedComps {
			for _, f := range newOut.own[c.Name] {
				path := filepath.Join(*dir, f)
				if _, err := os.Stat(path); err == nil && !*force {
					return fmt.Errorf("%s exists — pass -force to overwrite, or `lotusui update` to merge", path)
				}
				if err := os.WriteFile(path, newOut.files[f], 0o644); err != nil {
					return err
				}
				fmt.Printf("  + %s\n", path)
			}
		}
		// Already-vendored components: re-vendor under the grown set —
		// clean replace when untouched, three-way merge when edited.
		for _, c := range oldComps {
			st := existingByName[c.Name]
			if !localModified(st) {
				for _, f := range newOut.own[c.Name] {
					if err := os.WriteFile(filepath.Join(*dir, f), newOut.files[f], 0o644); err != nil {
						return err
					}
				}
				continue
			}
			for _, f := range newOut.own[c.Name] {
				var local []byte
				for _, sf := range st.files {
					if filepath.Base(sf.path) == f {
						local = stripStamp(sf.src)
					}
				}
				if local == nil || oldOut == nil {
					continue
				}
				merged, n, err := merge3(local, stripStamp(oldOut.files[f]), stripStamp(newOut.files[f]))
				if err != nil {
					return err
				}
				content := newOut.files[f]
				stampEnd := len(content) - len(stripStamp(content))
				if err := os.WriteFile(filepath.Join(*dir, f), append(append([]byte(nil), content[:stampEnd]...), merged...), 0o644); err != nil {
					return err
				}
				if n > 0 {
					fmt.Printf("  ⚠ %s: %d conflict(s) marked — your edits vs the re-vendored set\n", f, n)
				}
			}
		}
		// The companion is CLI-owned: always regenerated, never merged.
		if c, ok := newOut.files[carriedFile]; ok {
			if err := os.WriteFile(filepath.Join(*dir, carriedFile), c, 0o644); err != nil {
				return err
			}
		} else {
			os.Remove(filepath.Join(*dir, carriedFile))
		}
	}

	for _, blk := range addedBlocks {
		files, err := vendorBlock(moduleDir, blk, *pkgName, version)
		if err != nil {
			return fmt.Errorf("%s: %w", blk.Name, err)
		}
		for base, content := range files {
			path := filepath.Join(*dir, base)
			if _, err := os.Stat(path); err == nil && !*force {
				return fmt.Errorf("%s exists — pass -force to overwrite, or `lotusui update` to merge", path)
			}
			if err := os.WriteFile(path, content, 0o644); err != nil {
				return err
			}
			fmt.Printf("  + %s\n", path)
		}
	}

	fmt.Printf("  vendored at v%s. The copies are yours to edit; `lotusui update` merges upstream changes.\n", version)
	return nil
}

// vendorBlock copies a block (app-level composition using only the
// exported lotusui API): package rename + stamp, no qualification.
func vendorBlock(moduleDir string, blk componentSpec, targetPkg, version string) (map[string][]byte, error) {
	out := map[string][]byte{}
	var all []byte
	basenames := make([]string, 0, len(blk.Files))
	for _, f := range blk.Files {
		src, err := os.ReadFile(filepath.Join(moduleDir, f))
		if err != nil {
			return nil, err
		}
		src = blockPkgRe.ReplaceAll(src, []byte("package "+targetPkg))
		base := filepath.Base(f)
		out[base] = src
		basenames = append(basenames, base)
		all = append(all, src...)
	}
	hash := hash16(all)
	for _, base := range basenames {
		stamp := fmt.Sprintf("// Code vendored by `lotusui add %s`. Yours to edit.\n// lotusui:vendored %s v%s sha256:%s\n\n", blk.Name, blk.Name, version, hash)
		out[base] = append([]byte(stamp), out[base]...)
	}
	return out, nil
}

var blockPkgRe = regexp.MustCompile(`(?m)^package \w+$`)

func findComponent(name string) (componentSpec, bool) {
	for _, c := range components {
		if c.Name == name {
			return c, true
		}
	}
	return componentSpec{}, false
}

func findBlock(name string) (componentSpec, bool) {
	for _, b := range blocks {
		if b.Name == name {
			return b, true
		}
	}
	return componentSpec{}, false
}

// lotusuiModuleDir resolves where lotusui's source lives for the
// CURRENT build — the module cache, or a replace-directive checkout.
// This is the trick that makes the whole registry work offline: the
// Go module system already stores every version's pristine source.
func lotusuiModuleDir(appDir string) (string, error) {
	out, err := goCmd(appDir, "list", "-m", "-f", "{{.Dir}}", "github.com/ikaito-com/lotusui")
	if err != nil {
		return "", fmt.Errorf("resolving lotusui module (is it a dependency?): %w", err)
	}
	dir := strings.TrimSpace(out)
	if dir == "" {
		return "", fmt.Errorf("lotusui module has no local source — run `go mod download github.com/ikaito-com/lotusui`")
	}
	return dir, nil
}

// lotusuiVersionDir fetches the pristine source of a SPECIFIC version
// from the module cache (downloading it if needed).
func lotusuiVersionDir(appDir, version string) (string, error) {
	out, err := goCmd(appDir, "mod", "download", "-json", "github.com/ikaito-com/lotusui@v"+version)
	if err != nil {
		return "", fmt.Errorf("fetching lotusui v%s: %w", version, err)
	}
	var info struct{ Dir string }
	if err := json.Unmarshal([]byte(out), &info); err != nil || info.Dir == "" {
		return "", fmt.Errorf("fetching lotusui v%s: no module dir (local-only version?)", version)
	}
	return info.Dir, nil
}

func goCmd(dir string, args ...string) (string, error) {
	cmd := exec.Command("go", args...)
	cmd.Dir = dir
	var buf, errBuf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("go %s: %v: %s", strings.Join(args, " "), err, errBuf.String())
	}
	return buf.String(), nil
}
