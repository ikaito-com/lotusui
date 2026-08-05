package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"os"
	"strings"

	"github.com/ikaito-com/lotusui/site/docspages"
)

// Bundled for WASM / GitHub Pages. Populated by `make docsapp-data`
// before the wasm build. Keep this set small (JSON + markdown) —
// hero PNGs ship next to the wasm under media/ so the binary stays lean.
//
//go:embed data/*
var embeddedData embed.FS

func readData(name string) ([]byte, bool) {
	b, err := embeddedData.ReadFile("data/" + name)
	if err != nil || len(b) == 0 || (len(b) < 32 && bytes.Contains(b, []byte("placeholder"))) {
		return nil, false
	}
	return b, true
}

func loadVersions() []versionEntry {
	var raw []versionEntry
	if b, ok := readData("versions.json"); ok {
		_ = json.Unmarshal(b, &raw)
	}
	if len(raw) == 0 {
		for _, c := range []string{"versions.json", "../versions.json"} {
			if b, err := os.ReadFile(c); err == nil {
				if json.Unmarshal(b, &raw) == nil {
					break
				}
			}
		}
	}
	// Switcher shows one entry per major (latest of that line). While
	// we are still on v0.x there is a single option — the chip stays
	// painted but is not a menu.
	return majorVersionMenu(raw)
}

// majorVersionMenu keeps the newest release of each SemVer major
// (v1.2.0 and v0.9.0 both appear; v0.2.0 wins over v0.1.0). Input is
// usually newest-first from site/versions.json.
func majorVersionMenu(all []versionEntry) []versionEntry {
	if len(all) == 0 {
		return nil
	}
	best := map[int]versionEntry{}
	var majors []int
	for _, e := range all {
		maj, ok := semverMajor(e.Version)
		if !ok {
			continue
		}
		cur, seen := best[maj]
		if !seen {
			best[maj] = e
			majors = append(majors, maj)
			continue
		}
		if versionNewer(e.Version, cur.Version) {
			best[maj] = e
		}
	}
	// Newest major first (v2, v1, v0).
	for i := 0; i < len(majors); i++ {
		for j := i + 1; j < len(majors); j++ {
			if majors[j] > majors[i] {
				majors[i], majors[j] = majors[j], majors[i]
			}
		}
	}
	out := make([]versionEntry, 0, len(majors))
	for _, m := range majors {
		out = append(out, best[m])
	}
	return out
}

func semverMajor(v string) (int, bool) {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	if v == "" {
		return 0, false
	}
	i := strings.IndexByte(v, '.')
	if i < 0 {
		i = len(v)
	}
	n := 0
	for _, c := range v[:i] {
		if c < '0' || c > '9' {
			return 0, false
		}
		n = n*10 + int(c-'0')
	}
	return n, true
}

// versionNewer reports whether a is a higher SemVer than b (v-prefix ok).
func versionNewer(a, b string) bool {
	ap := semverParts(a)
	bp := semverParts(b)
	for i := 0; i < 3; i++ {
		if ap[i] != bp[i] {
			return ap[i] > bp[i]
		}
	}
	return false
}

func semverParts(v string) [3]int {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	var p [3]int
	for i, part := range strings.Split(v, ".") {
		if i >= 3 {
			break
		}
		n := 0
		for _, c := range part {
			if c < '0' || c > '9' {
				break
			}
			n = n*10 + int(c-'0')
		}
		p[i] = n
	}
	return p
}

func loadChangelogMarkdown() string {
	if b, ok := readData("CHANGELOG.md"); ok {
		return string(b)
	}
	for _, c := range []string{"../CHANGELOG.md", "../../CHANGELOG.md", "CHANGELOG.md"} {
		if b, err := os.ReadFile(c); err == nil {
			return string(b)
		}
	}
	return ""
}

func loadPerformancePage() *docspages.Page {
	if b, ok := readData("bench.json"); ok {
		if r, err := docspages.ParseBenchReport(b); err == nil {
			return docspages.PerformancePage(r)
		}
	}
	for _, c := range []string{"bench.json", "../bench.json"} {
		if r, err := docspages.LoadBenchReport(c); err == nil {
			return docspages.PerformancePage(r)
		}
	}
	return docspages.PerformancePage(docspages.BenchReport{Note: "bench.json missing — run make bench-doc"})
}
