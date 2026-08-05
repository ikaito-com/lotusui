package docspages_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/ikaito-com/lotusui/site/docspages"
	"github.com/ikaito-com/lotusui/site/live"
)

func TestEveryDemoResolvesInLive(t *testing.T) {
	bench := docspages.BenchReport{Note: "test"}
	if r, err := docspages.LoadBenchReport(filepath.Join("..", "bench.json")); err == nil {
		bench = r
	} else if r, err := docspages.LoadBenchReport("bench.json"); err == nil {
		bench = r
	}
	groups := docspages.Groups("", docspages.PerformancePage(bench))
	var missing []string
	for _, g := range groups {
		for _, p := range g.Pages {
			for _, s := range p.Sections {
				if s.Demo == "" {
					continue
				}
				slug := s.Demo
				if i := strings.IndexByte(slug, '/'); i >= 0 {
					slug = slug[:i]
				}
				if i := strings.IndexByte(slug, '&'); i >= 0 {
					slug = slug[:i]
				}
				if !live.Has(slug) {
					missing = append(missing, p.Slug+": "+s.Demo)
				}
			}
		}
	}
	if len(missing) > 0 {
		t.Fatalf("docspages demos missing from site/live:\n  %s", strings.Join(missing, "\n  "))
	}
}

func TestNavSlugCount(t *testing.T) {
	groups := docspages.Groups("", docspages.PerformancePage(docspages.BenchReport{}))
	n := 0
	for _, g := range groups {
		n += len(g.Pages)
	}
	if n < 50 {
		t.Fatalf("expected ≥50 docs pages, got %d", n)
	}
}
