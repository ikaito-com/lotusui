package main

import "testing"

func TestHTMLBlocksPerformanceShapes(t *testing.T) {
	html := `
<p>Intro paragraph.</p>
<div class="statgrid">
  <div class="stat"><div class="statnum">0<span>frames/s</span></div>
    <div class="statlabel">When idle</div>
    <div class="statnote">Settled UI</div></div>
  <div class="stat"><div class="statnum">400×</div>
    <div class="statlabel">Long list</div>
    <div class="statnote">ListView</div></div>
</div>
<div class="perf-takeaway"><strong>Use ListView</strong> for long lists.</div>
<div class="proptable-wrap"><table class="proptable">
<thead><tr><th>Control</th><th>Time</th></tr></thead>
<tbody>
<tr><td>Button</td><td>4 µs</td></tr>
<tr><td>Card</td><td>3 µs</td></tr>
</tbody></table></div>
<h3 class="perfsub">Desktop</h3>
<ol class="perflist"><li>Resolve theme once.</li><li>Cache icons.</li></ol>
`
	blocks := htmlBlocks(html)
	kinds := map[string]int{}
	for _, b := range blocks {
		kinds[b.Kind]++
	}
	if kinds["p"] < 1 || kinds["stats"] != 1 || kinds["takeaway"] != 1 || kinds["table"] != 1 || kinds["h3"] != 1 || kinds["list"] != 1 {
		t.Fatalf("unexpected blocks %#v from %d items", kinds, len(blocks))
	}
	var stats htmlBlock
	for _, b := range blocks {
		if b.Kind == "stats" {
			stats = b
		}
	}
	if len(stats.Stats) != 2 {
		t.Fatalf("stats=%d want 2", len(stats.Stats))
	}
	var table htmlBlock
	for _, b := range blocks {
		if b.Kind == "table" {
			table = b
		}
	}
	if len(table.Headers) != 2 || len(table.Rows) != 2 {
		t.Fatalf("table headers=%v rows=%v", table.Headers, table.Rows)
	}
}
