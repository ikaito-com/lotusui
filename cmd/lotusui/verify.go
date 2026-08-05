package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// cmdVerify is the "can't forget" guard: an OFFLINE, sub-second check
// that every generated artifact matches its source of truth — safe in
// CI and in every `make check`, no network, no side effects. It fails
// with exit 1 and says exactly what to run when something drifted.
//
// Checks (each only when its inputs exist / are passed):
//   - every manifest entry has its fetched SVG, and vice versa
//   - the generated constants file matches the manifest exactly
//   - theme_gen.go carries the sha256 of the theme.json it was
//     generated from (stamped by `lotusui theme -config`)
func cmdVerify(args []string) error {
	fs := flagSet("verify")
	manifest := fs.String("manifest", "", "icons manifest to verify (skip icon checks if empty)")
	out := fs.String("out", "", "icons directory (default: the manifest's directory)")
	gen := fs.String("gen", "", "generated icon-constants file to verify against the manifest")
	themeConfig := fs.String("theme-config", "", "theme.json to verify (skip theme check if empty)")
	themeGen := fs.String("theme-gen", "", "generated theme file that must match -theme-config")
	apiFile := fs.String("api", "", "API baseline to verify the live exported API against")
	registry := fs.Bool("registry", false, "verify registry.json against the source")
	fs.Parse(args)

	var problems []string

	if *apiFile != "" {
		problems = append(problems, verifyAPI(*apiFile)...)
	}
	if *registry {
		problems = append(problems, verifyRegistry(".")...)
	}

	if *manifest != "" {
		dir := *out
		if dir == "" {
			dir = filepath.Dir(*manifest)
		}
		problems = append(problems, verifyIcons(*manifest, dir, *gen)...)
	}
	if *themeConfig != "" {
		problems = append(problems, verifyTheme(*themeConfig, *themeGen)...)
	}

	if len(problems) > 0 {
		for _, p := range problems {
			fmt.Fprintln(os.Stderr, "  drift:", p)
		}
		return fmt.Errorf("%d problem(s) — see the messages above for the fix each needs", len(problems))
	}
	fmt.Println("  verify: generated code matches its sources")
	return nil
}

func verifyIcons(manifest, dir, gen string) []string {
	var problems []string
	data, err := os.ReadFile(manifest)
	if err != nil {
		return []string{err.Error()}
	}
	want := map[string]bool{} // icon name (file base) → listed
	for _, line := range strings.Split(string(data), "\n") {
		f := strings.Fields(line)
		if len(f) != 2 {
			continue
		}
		want[strings.TrimSuffix(f[1], ".svg")] = true
	}

	// Manifest ↔ SVG files, both directions.
	have := map[string]bool{}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return []string{err.Error()}
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".svg") {
			have[strings.TrimSuffix(e.Name(), ".svg")] = true
		}
	}
	for n := range want {
		if !have[n] {
			problems = append(problems, fmt.Sprintf("%s lists %s.svg but it is not in %s", manifest, n, dir))
		}
	}
	for n := range have {
		if !want[n] {
			problems = append(problems, fmt.Sprintf("%s/%s.svg exists but %s does not list it", dir, n, manifest))
		}
	}

	// Manifest ↔ generated constants, both directions.
	if gen != "" {
		src, err := os.ReadFile(gen)
		if err != nil {
			return append(problems, err.Error())
		}
		constRe := regexp.MustCompile(`Icon[A-Za-z0-9]+\s*=\s*"([a-z0-9-_]+)"`)
		generated := map[string]bool{}
		for _, m := range constRe.FindAllStringSubmatch(string(src), -1) {
			generated[m[1]] = true
		}
		for n := range want {
			if !generated[n] {
				problems = append(problems, fmt.Sprintf("%s has no constant for %q", gen, n))
			}
		}
		for n := range generated {
			if !want[n] {
				problems = append(problems, fmt.Sprintf("%s has a constant for %q, which the manifest no longer lists", gen, n))
			}
		}
	}
	return problems
}

// verifyAPI compares the live exported API against the committed
// baseline. A mismatch is the VERSIONING TRIPWIRE: the fix is to
// document every change in CHANGELOG.md's [Unreleased] section, then
// regenerate the baseline — in the same commit.
func verifyAPI(path string) []string {
	want, err := os.ReadFile(path)
	if err != nil {
		return []string{err.Error()}
	}
	inv, err := apiInventory(filepath.Dir(path))
	if err != nil {
		return []string{err.Error()}
	}
	wantSet := map[string]bool{}
	for _, l := range strings.Split(strings.TrimSpace(string(want)), "\n") {
		wantSet[l] = true
	}
	liveSet := map[string]bool{}
	var problems []string
	for _, l := range inv {
		liveSet[l] = true
		if !wantSet[l] {
			problems = append(problems, "API added/changed: "+l)
		}
	}
	for l := range wantSet {
		if !liveSet[l] {
			problems = append(problems, "API removed/changed: "+l)
		}
	}
	if len(problems) > 0 {
		problems = append(problems,
			"the exported API differs from api.txt — record EVERY change in CHANGELOG.md [Unreleased] (exact symbols, old→new), then `go run ./cmd/lotusui api -o api.txt`, and commit changelog + baseline + code together")
	}
	return problems
}

var stampRe = regexp.MustCompile(`source sha256:([0-9a-f]{12})`)

func verifyTheme(config, gen string) []string {
	if gen == "" {
		return []string{"-theme-config given without -theme-gen"}
	}
	cfg, err := os.ReadFile(config)
	if err != nil {
		return []string{err.Error()}
	}
	src, err := os.ReadFile(gen)
	if err != nil {
		return []string{err.Error()}
	}
	m := stampRe.FindSubmatch(src)
	if m == nil {
		return []string{fmt.Sprintf("%s has no source hash stamp — regenerate it with `lotusui theme -config %s`", gen, config)}
	}
	if got := shortHash(cfg); got != string(m[1]) {
		return []string{fmt.Sprintf("%s was edited after %s was generated (hash %s != stamped %s)", config, gen, got, m[1])}
	}
	return nil
}

func shortHash(b []byte) string {
	sum := sha256.Sum256(b)
	return fmt.Sprintf("%x", sum)[:12]
}
