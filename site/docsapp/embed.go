package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"os"

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
	if b, ok := readData("versions.json"); ok {
		var v []versionEntry
		if json.Unmarshal(b, &v) == nil {
			return v
		}
	}
	for _, c := range []string{"versions.json", "../versions.json"} {
		if b, err := os.ReadFile(c); err == nil {
			var v []versionEntry
			if json.Unmarshal(b, &v) == nil {
				return v
			}
		}
	}
	return nil
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
