package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// cmdUpdate merges upstream changes into vendored components. The
// engineering that makes this safe: the vendoring transform is a pure
// function of (source version, component set), and the Go module
// cache stores every version's pristine source forever. So for each
// stamped file we can reconstruct the EXACT base (the pristine
// vendored form at the stamped version) and remote (the same at the
// new version), making the app's local file the third side of a true
// three-way merge — the app's own edits, and only those, survive as
// diffs.
//
//	go run github.com/ikaito-com/lotusui/cmd/lotusui update            # everything stamped under ./ui
//	go run github.com/ikaito-com/lotusui/cmd/lotusui update button -dir internal/ui
//
// Files never modified since `add` (hash matches the stamp) fast-path
// to a clean replace. Modified files go through diff3; conflicts are
// left as standard markers, and the agent-facing CHANGELOG sections
// between the two versions are printed so an AI agent — or you — can
// resolve them with full context.
func cmdUpdate(args []string) error {
	fs := flagSet("update")
	dir := fs.String("dir", "ui", "directory holding vendored components")
	dry := fs.Bool("dry", false, "report what would change without writing")
	fs.Parse(reorderFlags(args, map[string]bool{"dry": true}))
	only := map[string]bool{}
	for _, n := range fs.Args() {
		only[n] = true
	}

	newDir, err := lotusuiModuleDir(".")
	if err != nil {
		return err
	}
	newVersion, err := moduleVersion(newDir)
	if err != nil {
		return err
	}
	newPkg, err := loadRootPackage(newDir)
	if err != nil {
		return err
	}

	stamped, err := scanStamps(*dir)
	if err != nil {
		return err
	}
	if len(stamped) == 0 {
		return fmt.Errorf("no vendored components under %s (missing lotusui:vendored stamps)", *dir)
	}
	targetPkg := filePackage(stamped[0].files[0].path)

	// The component SET is the transform's unit of coherence — the
	// whole set moves to the new version together (selecting single
	// components with args only limits which FILES get written).
	var setComps []componentSpec
	stampByName := map[string]stampInfo{}
	var blocksStamped []stampInfo
	for _, st := range stamped {
		stampByName[st.component] = st
		if st.component == carriedComponent {
			continue
		}
		if c, ok := findComponent(st.component); ok {
			setComps = append(setComps, c)
		} else if _, ok := findBlock(st.component); ok {
			blocksStamped = append(blocksStamped, st)
		} else {
			fmt.Printf("  ! %s: no longer in the registry — check the CHANGELOG\n", st.component)
		}
	}

	remote, err := vendorSet(newPkg, setComps, targetPkg, newVersion)
	if err != nil {
		return err
	}

	// Merge bases, one per distinct stamped version, reconstructed
	// from the module cache — only for components with local edits.
	baseByVersion := map[string]*vendored{}
	baseFor := func(version string) (*vendored, error) {
		if b, ok := baseByVersion[version]; ok {
			return b, nil
		}
		oldDir, err := lotusuiVersionDir(".", version)
		if err != nil {
			return nil, fmt.Errorf("%w (a replace-directive dev version can't be merged 3-way; resolve by hand)", err)
		}
		oldPkg, err := loadRootPackage(oldDir)
		if err != nil {
			return nil, err
		}
		// The old version may not have every component of the set.
		var avail []componentSpec
		for _, c := range setComps {
			ok := true
			for _, f := range c.Files {
				if oldPkg.files[f] == nil {
					ok = false
				}
			}
			if ok {
				avail = append(avail, c)
			}
		}
		b, err := vendorSet(oldPkg, avail, targetPkg, version)
		if err != nil {
			return nil, err
		}
		baseByVersion[version] = b
		return b, nil
	}

	updated, conflicted := 0, 0
	for _, c := range setComps {
		st := stampByName[c.Name]
		if len(only) > 0 && !only[c.Name] {
			continue
		}
		if st.version == newVersion && !localModified(st) {
			fmt.Printf("  = %s already at v%s\n", c.Name, newVersion)
			continue
		}
		conflicts := 0
		if !localModified(st) {
			for _, f := range remote.own[c.Name] {
				if err := writeOrDry(filepath.Join(*dir, f), remote.files[f], *dry); err != nil {
					return err
				}
			}
		} else {
			base, err := baseFor(st.version)
			if err != nil {
				return fmt.Errorf("%s: %w", c.Name, err)
			}
			for _, f := range remote.own[c.Name] {
				var local []byte
				for _, sf := range st.files {
					if filepath.Base(sf.path) == f {
						local = stripStamp(sf.src)
					}
				}
				content := remote.files[f]
				if local != nil && base.files[f] != nil {
					merged, n, err := merge3(local, stripStamp(base.files[f]), stripStamp(content))
					if err != nil {
						return fmt.Errorf("%s: %w", c.Name, err)
					}
					conflicts += n
					stampEnd := len(content) - len(stripStamp(content))
					content = append(append([]byte(nil), content[:stampEnd]...), merged...)
				}
				if err := writeOrDry(filepath.Join(*dir, f), content, *dry); err != nil {
					return err
				}
			}
		}
		updated++
		conflicted += conflicts
		mark, mode := "↑", "clean"
		if localModified(st) {
			mode = "3-way merge"
		}
		if conflicts > 0 {
			mark = "⚠"
		}
		fmt.Printf("  %s %s v%s → v%s (%s", mark, c.Name, st.version, newVersion, mode)
		if conflicts > 0 {
			fmt.Printf(", %d conflict(s) marked", conflicts)
		}
		fmt.Println(")")
	}

	// The companion is CLI-owned: regenerated wholesale, never merged.
	if updated > 0 {
		if c, ok := remote.files[carriedFile]; ok {
			if err := writeOrDry(filepath.Join(*dir, carriedFile), c, *dry); err != nil {
				return err
			}
		} else if _, err := os.Stat(filepath.Join(*dir, carriedFile)); err == nil && !*dry {
			os.Remove(filepath.Join(*dir, carriedFile))
		}
	}

	// Blocks: self-contained copies with the same clean/merge split.
	for _, st := range blocksStamped {
		if len(only) > 0 && !only[st.component] {
			continue
		}
		if st.version == newVersion {
			fmt.Printf("  = %s already at v%s\n", st.component, newVersion)
			continue
		}
		blk, _ := findBlock(st.component)
		remoteB, err := vendorBlock(newDir, blk, targetPkg, newVersion)
		if err != nil {
			return fmt.Errorf("%s: %w", st.component, err)
		}
		modified := localModified(st)
		conflicts := 0
		for name, content := range remoteB {
			if modified {
				oldDir, err := lotusuiVersionDir(".", st.version)
				if err != nil {
					return fmt.Errorf("%s: %w", st.component, err)
				}
				baseB, err := vendorBlock(oldDir, blk, targetPkg, st.version)
				if err != nil {
					return err
				}
				var local []byte
				for _, sf := range st.files {
					if filepath.Base(sf.path) == name {
						local = stripStamp(sf.src)
					}
				}
				if local != nil && baseB[name] != nil {
					merged, n, err := merge3(local, stripStamp(baseB[name]), stripStamp(content))
					if err != nil {
						return err
					}
					conflicts += n
					stampEnd := len(content) - len(stripStamp(content))
					content = append(append([]byte(nil), content[:stampEnd]...), merged...)
				}
			}
			if err := writeOrDry(filepath.Join(*dir, name), content, *dry); err != nil {
				return err
			}
		}
		updated++
		conflicted += conflicts
		fmt.Printf("  ↑ %s v%s → v%s\n", st.component, st.version, newVersion)
	}

	if updated > 0 {
		printChangelogDelta(newDir, stamped, newVersion)
		if conflicted > 0 {
			fmt.Println("\n  resolve the <<<<<<< markers, then build; the changelog above describes every upstream change.")
		}
	}
	return nil
}

func writeOrDry(path string, content []byte, dry bool) error {
	if dry {
		fmt.Printf("    would write %s\n", path)
		return nil
	}
	return os.WriteFile(path, content, 0o644)
}

// stampInfo is one vendored component found on disk.
type stampInfo struct {
	component string
	version   string
	hash      string
	files     []stampedFile
}

type stampedFile struct {
	path string
	src  []byte
}

var vendorStampRe = regexp.MustCompile(`// lotusui:vendored (\S+) v(\S+) sha256:(\S+)`)

func scanStamps(dir string) ([]stampInfo, error) {
	byComponent := map[string]*stampInfo{}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		src, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		m := vendorStampRe.FindSubmatch(src)
		if m == nil {
			continue
		}
		key := string(m[1])
		st := byComponent[key]
		if st == nil {
			st = &stampInfo{component: key, version: string(m[2]), hash: string(m[3])}
			byComponent[key] = st
		}
		st.files = append(st.files, stampedFile{path: path, src: src})
	}
	out := make([]stampInfo, 0, len(byComponent))
	for _, st := range byComponent {
		sort.Slice(st.files, func(i, j int) bool { return st.files[i].path < st.files[j].path })
		out = append(out, *st)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].component < out[j].component })
	return out, nil
}

// localModified reports whether the on-disk copy differs from the
// pristine vendored form it was stamped with — computed the same way
// vendorSet fingerprints a component's own files.
func localModified(st stampInfo) bool {
	h := sha256.New()
	for _, f := range st.files {
		h.Write([]byte(filepath.Base(f.path)))
		h.Write(stripStamp(f.src))
	}
	return hex.EncodeToString(h.Sum(nil))[:16] != st.hash
}

// stripStamp removes the stamp header lines so hashes compare
// content, not metadata.
func stripStamp(src []byte) []byte {
	s := string(src)
	if i := strings.Index(s, "\n\n"); i >= 0 && strings.HasPrefix(s, "// Code ") {
		return []byte(s[i+2:])
	}
	return src
}

// merge3 runs diff3 through `git merge-file -p` — available anywhere
// git is, which for lotusui's audience is everywhere. A non-zero exit
// means conflicts, which we pass through as standard markers for the
// developer or agent to resolve.
func merge3(local, base, remote []byte) ([]byte, int, error) {
	tmp, err := os.MkdirTemp("", "lotusui-merge")
	if err != nil {
		return nil, 0, err
	}
	defer os.RemoveAll(tmp)
	paths := map[string][]byte{"local": local, "base": base, "remote": remote}
	for n, b := range paths {
		if err := os.WriteFile(filepath.Join(tmp, n), b, 0o644); err != nil {
			return nil, 0, err
		}
	}
	cmd := exec.Command("git", "merge-file", "-p",
		"-L", "app (your edits)", "-L", "base", "-L", "lotusui (upstream)",
		filepath.Join(tmp, "local"), filepath.Join(tmp, "base"), filepath.Join(tmp, "remote"))
	out, err := cmd.Output()
	if err != nil {
		if _, isExit := err.(*exec.ExitError); !isExit {
			return nil, 0, fmt.Errorf("git merge-file: %w", err)
		}
	}
	return out, strings.Count(string(out), "<<<<<<<"), nil
}

var filePkgRe = regexp.MustCompile(`(?m)^package (\w+)$`)

func filePackage(path string) string {
	src, err := os.ReadFile(path)
	if err != nil {
		return "ui"
	}
	if m := filePkgRe.FindSubmatch(src); m != nil {
		return string(m[1])
	}
	return "ui"
}

// printChangelogDelta prints every CHANGELOG section strictly newer
// than the oldest stamped version — the exact upstream story an agent
// needs to reconcile the app.
func printChangelogDelta(moduleDir string, stamped []stampInfo, newVersion string) {
	oldest := newVersion
	for _, st := range stamped {
		if semverLess(st.version, oldest) {
			oldest = st.version
		}
	}
	if oldest == newVersion {
		return
	}
	log, err := os.ReadFile(filepath.Join(moduleDir, "CHANGELOG.md"))
	if err != nil {
		return
	}
	sections := regexp.MustCompile(`(?m)^## \[`).Split(string(log), -1)
	fmt.Printf("\n  ── upstream changes v%s → v%s ──\n", oldest, newVersion)
	for _, s := range sections[1:] {
		end := strings.IndexAny(s, "]")
		if end < 0 {
			continue
		}
		ver := s[:end]
		if ver == "Unreleased" {
			continue
		}
		if semverLess(oldest, ver) && !semverLess(newVersion, ver) {
			fmt.Println("## [" + s)
		}
	}
}

func semverLess(a, b string) bool {
	pa, pb := semverParts(a), semverParts(b)
	for i := 0; i < 3; i++ {
		if pa[i] != pb[i] {
			return pa[i] < pb[i]
		}
	}
	return false
}

func semverParts(v string) [3]int {
	var out [3]int
	fmt.Sscanf(v, "%d.%d.%d", &out[0], &out[1], &out[2])
	return out
}
