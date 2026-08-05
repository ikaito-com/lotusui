package docspages

import (
	"encoding/json"
	"fmt"
	"html"
	"math"
	"os"
	"strings"
	"time"
)

// BenchReport is the committed snapshot behind the Performance docs
// page (site/bench.json). Refresh with `make bench-doc` from the repo
// root so hero numbers cannot drift from the suite by hand.
type BenchReport struct {
	Generated            string     `json:"generated"`
	Go                   string     `json:"go"`
	GOOS                 string     `json:"goos"`
	GOARCH               string     `json:"goarch"`
	CPU                  string     `json:"cpu"`
	Note                 string     `json:"note"`
	WasmBytes            int64      `json:"wasmBytes,omitempty"`
	WasmGzipBytes        int64      `json:"wasmGzipBytes,omitempty"`
	WasmNote             string     `json:"wasmNote,omitempty"`
	AppBytes             int64      `json:"appBytes,omitempty"`
	AppNote              string     `json:"appNote,omitempty"`
	MinimalWasmBytes     int64      `json:"minimalWasmBytes,omitempty"`
	MinimalWasmGzipBytes int64      `json:"minimalWasmGzipBytes,omitempty"`
	MinimalWasmNote      string     `json:"minimalWasmNote,omitempty"`
	WasmExecGzipBytes    int64      `json:"wasmExecGzipBytes,omitempty"`
	WasmExecNote         string     `json:"wasmExecNote,omitempty"`
	Peers                []PeerSize `json:"peers,omitempty"`
	Benches              []BenchRow `json:"benches"`
}

// PeerSize is a same-machine measurement for a peer toolkit.
// Desktop: stripped native binary (Bytes = on-disk). Web: prefer
// Encoding "gzip" so Bytes is transfer size; BytesRaw is optional.
type PeerSize struct {
	Name     string `json:"name"`
	Surface  string `json:"surface"` // desktop | web
	Bytes    int64  `json:"bytes"`
	BytesRaw int64  `json:"bytesRaw,omitempty"`
	Encoding string `json:"encoding,omitempty"` // gzip when Bytes is transfer size
	Note     string `json:"note"`
}

// BenchRow is one go test -benchmem line, reduced to a median.
type BenchRow struct {
	Name        string  `json:"name"`
	NsPerOp     float64 `json:"nsPerOp"`
	BytesPerOp  int64   `json:"bytesPerOp"`
	AllocsPerOp int64   `json:"allocsPerOp"`
}

func LoadBenchReport(path string) (BenchReport, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return BenchReport{}, err
	}
	return ParseBenchReport(b)
}

// ParseBenchReport unmarshals a bench.json blob (also used when the
// report is embedded in the docsapp WASM).
func ParseBenchReport(b []byte) (BenchReport, error) {
	var r BenchReport
	if err := json.Unmarshal(b, &r); err != nil {
		return BenchReport{}, err
	}
	return r, nil
}

func (r BenchReport) byName(name string) (BenchRow, bool) {
	for _, b := range r.Benches {
		if b.Name == name {
			return b, true
		}
	}
	return BenchRow{}, false
}

func (r BenchReport) peer(name string) (PeerSize, bool) {
	for _, p := range r.Peers {
		if p.Name == name {
			return p, true
		}
	}
	return PeerSize{}, false
}

func formatNs(ns float64) string {
	switch {
	case ns < 1000:
		if ns == math.Trunc(ns) {
			return fmt.Sprintf("%.0f ns", ns)
		}
		return fmt.Sprintf("%.1f ns", ns)
	case ns < 1e6:
		return fmt.Sprintf("%.1f µs", ns/1e3)
	default:
		return fmt.Sprintf("%.1f ms", ns/1e6)
	}
}

func formatBytes(n int64) string {
	switch {
	case n < 1024:
		return fmt.Sprintf("%d B", n)
	case n < 1024*1024:
		return fmt.Sprintf("%.1f KB", float64(n)/1024)
	default:
		return fmt.Sprintf("%.1f MB", float64(n)/(1024*1024))
	}
}

func formatRatio(a, b float64) string {
	if a <= 0 {
		return "—"
	}
	return fmt.Sprintf("%.0f×", b/a)
}

func framePct(ns float64) string {
	if ns <= 0 {
		return "—"
	}
	pct := 100 * ns / 16.67e6
	switch {
	case pct < 0.01:
		return "<0.01%"
	case pct < 1:
		return fmt.Sprintf("%.2f%%", pct)
	case pct < 10:
		return fmt.Sprintf("%.1f%%", pct)
	default:
		return fmt.Sprintf("%.0f%%", pct)
	}
}

func barWidth(part, total float64) string {
	if total <= 0 {
		return "1%"
	}
	pct := 100 * part / total
	if pct < 1.2 {
		pct = 1.2
	}
	if pct > 100 {
		pct = 100
	}
	return fmt.Sprintf("%.1f%%", pct)
}

func benchLabel(name string) string {
	switch name {
	case "SVGIconCacheHit":
		return "Warm icon cache hit"
	case "LabelFrame":
		return "One body label"
	case "BadgeFrame":
		return "One badge"
	case "ButtonFrame":
		return "One settled button"
	case "CardFrame":
		return "One card + body text"
	case "CheckboxFrame":
		return "One checkbox + label"
	case "SwitchFrame":
		return "One switch"
	case "InputFrame":
		return "One labeled input"
	case "TabsFrame":
		return "Three tabs"
	case "SimpleGridFrame":
		return "SimpleGrid · 30 cards"
	case "ListView10k":
		return "ListView · 10,000 rows"
	case "Scrollable10k":
		return "Scrollable · 10,000 rows"
	case "NewTheme":
		return "NewTheme (startup, once)"
	default:
		return name
	}
}

// performancePage builds the docs Performance page from site/bench.json.
func PerformancePage(r BenchReport) *Page {
	icon, _ := r.byName("SVGIconCacheHit")
	btn, _ := r.byName("ButtonFrame")
	card, _ := r.byName("CardFrame")
	in, _ := r.byName("InputFrame")
	tabs, _ := r.byName("TabsFrame")
	grid, _ := r.byName("SimpleGridFrame")
	list, _ := r.byName("ListView10k")
	scroll, _ := r.byName("Scrollable10k")
	theme, _ := r.byName("NewTheme")
	ratio := formatRatio(list.NsPerOp, scroll.NsPerOp)

	fyne, hasFyne := r.peer("fyne")
	egui, hasEgui := r.peer("egui")
	machine := strings.TrimSpace(strings.Join([]string{r.Go, r.GOOS + "/" + r.GOARCH, r.CPU}, " · "))
	if machine == " ·  · " || machine == "" {
		machine = "run make bench-doc to capture a machine snapshot"
	}
	genDay := r.Generated
	if t, err := time.Parse(time.RFC3339, r.Generated); err == nil {
		genDay = t.Format("2 Jan 2006")
	}

	appSize := "—"
	if r.AppBytes > 0 {
		appSize = formatBytes(r.AppBytes)
	}
	minWasmGzip := "—"
	minWasmRaw := "—"
	if r.MinimalWasmGzipBytes > 0 {
		minWasmGzip = formatBytes(r.MinimalWasmGzipBytes)
	}
	if r.MinimalWasmBytes > 0 {
		minWasmRaw = formatBytes(r.MinimalWasmBytes)
	}
	galGzip := "—"
	galRaw := "—"
	if r.WasmGzipBytes > 0 {
		galGzip = formatBytes(r.WasmGzipBytes)
	}
	if r.WasmBytes > 0 {
		galRaw = formatBytes(r.WasmBytes)
	}
	execGzip := "—"
	if r.WasmExecGzipBytes > 0 {
		execGzip = formatBytes(r.WasmExecGzipBytes)
	}

	// --- Intro: three promises a human can remember ---
	intro := fmt.Sprintf(`<p>lotusui is a design system on
<a href="https://gioui.org">Gio</a> — Go’s immediate-mode UI toolkit.
We keep windows quiet when nothing is happening, and cheap when something is.
This page is the evidence — not a bakeoff against every UI stack on earth.</p>
<p class="perfnote">Numbers below are from <strong>%s</strong> · %s.
Re-run with <code>make bench-doc</code>.</p>
<div class="statgrid">
  <div class="stat"><div class="statnum">0<span>frames/s</span></div>
    <div class="statlabel">When idle</div>
    <div class="statnote">Settled UI asks for no redraws</div></div>
  <div class="stat"><div class="statnum">%s</div>
    <div class="statlabel">Long list, done right</div>
    <div class="statnote">ListView vs laying out every row</div></div>
  <div class="stat"><div class="statnum">%s/op</div>
    <div class="statlabel">One button</div>
    <div class="statnote">%s of a 16.7&nbsp;ms frame · %d allocs</div></div>
  <div class="stat"><div class="statnum">0<span>allocs</span></div>
    <div class="statlabel">Warm icon hit</div>
    <div class="statnote">Pinned by test · %s/op</div></div>
</div>`,
		html.EscapeString(genDay),
		html.EscapeString(machine),
		html.EscapeString(ratio),
		formatNs(btn.NsPerOp),
		framePct(btn.NsPerOp),
		btn.AllocsPerOp,
		formatNs(icon.NsPerOp),
	)

	// --- How we measure (short, one metric) ---
	measure := `<div class="perf-takeaway">
  <strong>What “time” means on this page.</strong>
  Almost every µs/ms figure is <em>layout→ops</em>: how long it takes Go to lay
  out a widget and record <a href="https://gioui.org">Gio</a> draw commands for
  one frame (fixed 800×600). That is <em>not</em> GPU paint time and not FPS.
  The column <strong>% of 16.7&nbsp;ms</strong> is that cost as a share of one
  60&nbsp;Hz frame budget — the same budget Flutter and Compose talk about.
</div>
<p>We follow the practice serious UI stacks use: pin <strong>contracts</strong>
with tests, check costs against a <strong>frame budget</strong>, and only
compare peers when the <strong>artifact matches</strong> (same kind of file,
same machine, named build flags). We do not invent a single “faster than
Flutter” score.</p>`

	// --- The list cliff ---
	lists := fmt.Sprintf(`<p>If you put 10,000 rows in a plain scroll view, every frame rebuilds all of
them. That is the classic trap — the same one Flutter and Compose warn about
when you skip the lazy list API.</p>
<p><code>ListView</code> only lays out what is on screen. Same data, same row,
same window size:</p>
<div class="perfbars">
  <div class="perfrow"><span class="perfname">ListView</span>
    <span class="perfbar" style="width:%s"></span>
    <span class="perfval">%s/op · %s of frame</span></div>
  <div class="perfrow"><span class="perfname">Scrollable</span>
    <span class="perfbar warn" style="width:%s"></span>
    <span class="perfval">%s/op · %s of frame</span></div>
</div>
<div class="proptable-wrap"><table class="proptable">
<thead><tr><th></th><th>layout→ops</th><th>%% of 16.7&nbsp;ms</th><th>allocs/op</th></tr></thead>
<tbody>
<tr><td><code>ListView</code> · virtualized</td><td><strong>%s/op</strong></td><td>%s</td><td>%d</td></tr>
<tr class="warnrow"><td><code>Scrollable</code> · all rows</td><td><strong>%s/op</strong></td><td>%s</td><td>%d</td></tr>
<tr><td>Difference</td><td colspan="3"><strong>%s</strong> cheaper with virtualization</td></tr>
</tbody></table></div>
<div class="perf-takeaway">
  <strong>Use <code>ListView</code> for long lists.</strong>
  Use <code>Scrollable</code> for short, mixed screens that fit in memory.
</div>`,
		barWidth(list.NsPerOp, scroll.NsPerOp),
		formatNs(list.NsPerOp), framePct(list.NsPerOp),
		barWidth(scroll.NsPerOp, scroll.NsPerOp),
		formatNs(scroll.NsPerOp), framePct(scroll.NsPerOp),
		formatNs(list.NsPerOp), framePct(list.NsPerOp), list.AllocsPerOp,
		formatNs(scroll.NsPerOp), framePct(scroll.NsPerOp), scroll.AllocsPerOp,
		html.EscapeString(ratio),
	)

	// --- Everyday controls ---
	controls := fmt.Sprintf(`<p>A screen is mostly ordinary widgets. Here is what one of each costs for a
single layout→ops pass — and how much of a 60&nbsp;Hz frame that is.</p>
<div class="proptable-wrap"><table class="proptable">
<thead><tr><th>Control</th><th>layout→ops</th><th>%% of 16.7&nbsp;ms</th><th>allocs/op</th></tr></thead>
<tbody>
<tr><td>Warm icon lookup</td><td><strong>%s/op</strong></td><td>%s</td><td>0</td></tr>
<tr><td>Button</td><td><strong>%s/op</strong></td><td>%s</td><td>%d</td></tr>
<tr><td>Card + label</td><td><strong>%s/op</strong></td><td>%s</td><td>%d</td></tr>
<tr><td>Labeled input</td><td><strong>%s/op</strong></td><td>%s</td><td>%d</td></tr>
<tr><td>Three tabs</td><td><strong>%s/op</strong></td><td>%s</td><td>%d</td></tr>
<tr><td>SimpleGrid · 30 cards</td><td><strong>%s/op</strong></td><td>%s</td><td>%d</td></tr>
</tbody></table></div>
<p class="perfnote"><code>NewTheme</code> is paid once at startup (%s/op) — not every frame.</p>`,
		formatNs(icon.NsPerOp), framePct(icon.NsPerOp),
		formatNs(btn.NsPerOp), framePct(btn.NsPerOp), btn.AllocsPerOp,
		formatNs(card.NsPerOp), framePct(card.NsPerOp), card.AllocsPerOp,
		formatNs(in.NsPerOp), framePct(in.NsPerOp), in.AllocsPerOp,
		formatNs(tabs.NsPerOp), framePct(tabs.NsPerOp), tabs.AllocsPerOp,
		formatNs(grid.NsPerOp), framePct(grid.NsPerOp), grid.AllocsPerOp,
		formatNs(theme.NsPerOp),
	)

	// --- Idle ---
	idle := `<p>When nothing is animating or receiving input, a quiet UI should not keep
the CPU warm. lotusui’s contract: settled animations issue
<strong>no <code>InvalidateCmd</code></strong> — Gio schedules
<strong>0 frames/s</strong> until something changes.</p>
<p>That matches how Fyne (event-driven) and Flutter / Compose / SwiftUI
(skip work when nothing is dirty) behave when idle. Many game-style ImGui
hosts redraw every vsync unless you add idle waits — a different default,
not a score we invent here.</p>`

	// --- Ship size: desktop binaries vs web downloads ---
	deskRows := strings.Builder{}
	fmt.Fprintf(&deskRows,
		`<tr><td><strong>lotusui</strong></td><td><strong>%s</strong></td><td>Theme + one Button · Go + Gio + fonts + lotusui · <code>-ldflags=-s -w</code></td></tr>`,
		appSize)
	if hasFyne && fyne.Bytes > 0 {
		fmt.Fprintf(&deskRows,
			`<tr><td><strong>Fyne</strong></td><td><strong>%s</strong></td><td>%s</td></tr>`,
			formatBytes(fyne.Bytes), html.EscapeString(fyne.Note))
	}
	if hasEgui && egui.Bytes > 0 {
		fmt.Fprintf(&deskRows,
			`<tr><td><strong>egui</strong></td><td><strong>%s</strong></td><td>%s</td></tr>`,
			formatBytes(egui.Bytes), html.EscapeString(egui.Note))
	}

	lotusWebRows := strings.Builder{}
	fmt.Fprintf(&lotusWebRows,
		`<tr><td><strong>lotusui</strong> · minimal WASM</td><td><strong>%s</strong> gzip</td><td>%s raw</td><td>Theme + Button · Go + Gio + fonts + lotusui in the download</td></tr>`,
		minWasmGzip, minWasmRaw)
	fmt.Fprintf(&lotusWebRows,
		`<tr><td>lotusui docs gallery</td><td><strong>%s</strong> gzip</td><td>%s raw</td><td>Every demo + fonts — not a minimal app</td></tr>`,
		galGzip, galRaw)
	fmt.Fprintf(&lotusWebRows,
		`<tr><td><code>wasm_exec.js</code></td><td>%s gzip</td><td>—</td><td>Go’s JS glue, usually served beside the .wasm</td></tr>`,
		execGzip)

	ship := fmt.Sprintf(`<p>Compare only when the <strong>artifact matches</strong>. Desktop
binaries are like-for-like (each ships its toolkit runtime). On the web we
report lotusui’s own transfer sizes — not a DOM toolkit row. Inventing a
“browser engine weight” for React/MUI would be fiction; omitting that peer
keeps the table honest.</p>

<h3 class="perfsub">Desktop — stripped native binary</h3>
<p>Label + button (or Theme + Button), same machine. Each number includes that
toolkit’s runtime (Go+Gio, Fyne, or egui) — the OS does not ship those for you.</p>
<div class="proptable-wrap"><table class="proptable">
<thead><tr><th>Toolkit</th><th>File size</th><th>What we built</th></tr></thead>
<tbody>%s</tbody>
</table></div>
<p class="perfnote">OS/arch %s/%s. Go apps use <code>-ldflags=-s -w</code>; egui uses
cargo <code>strip=true</code> (no LTO in this snapshot). Still not identical
languages or font embedding — same <em>kind</em> of artifact.</p>

<h3 class="perfsub">Web — lotusui download (gzip transfer)</h3>
<p>lotusui on the web is a <a href="https://gioui.org">Gio</a> program compiled to
WebAssembly. The .wasm includes <strong>Go + Gio + embedded fonts + lotusui + your
UI</strong>. Gzip is the usual <code>Content-Encoding</code> transfer size
(best compression); raw is the on-disk .wasm. Hosts like GitHub Pages usually
gzip on the wire automatically — the gzip column is what users typically
transfer.</p>
<div class="proptable-wrap"><table class="proptable">
<thead><tr><th>Artifact</th><th>Transfer</th><th>On disk</th><th>What’s inside</th></tr></thead>
<tbody>%s</tbody>
</table></div>`,
		deskRows.String(),
		html.EscapeString(r.GOOS),
		html.EscapeString(r.GOARCH),
		lotusWebRows.String(),
	)

	// --- What we don't claim ---
	honest := `<p>Flutter and Compose publish device-lab scroll <em>jank %</em> and full-frame
timelines. That is excellent engineering — and a different instrument from our
layout→ops microbenchmarks (which sit on Gio’s frame loop). Until lotusui has a
matching device suite, we
<strong>cite</strong> their public figures when useful (e.g. Compose 1.9+
Pokedex scroll at 0.21% jank matching Views) and we <strong>do not</strong>
put them in a column next to our µs/op as if they were the same measurement.</p>
<p>No FPS leaderboard. No “we beat SwiftUI” claim. No web bakeoff against
DOM libraries whose “runtime” is already in the browser. The credible story is:
contracts you can re-run, a frame budget you can reason about, and sizes
with named units when the artifact matches.</p>`

	// --- Full snapshot (engineers) ---
	tableRows := strings.Builder{}
	for _, b := range r.Benches {
		warn := ""
		if b.Name == "Scrollable10k" {
			warn = ` class="warnrow"`
		}
		pct := framePct(b.NsPerOp)
		if b.Name == "NewTheme" {
			pct = "—"
		}
		fmt.Fprintf(&tableRows,
			`<tr%s><td>%s</td><td><code>%s</code></td><td>%s/op</td><td>%s</td><td>%d/op</td><td>%s/op</td></tr>`,
			warn,
			html.EscapeString(benchLabel(b.Name)),
			html.EscapeString(b.Name),
			formatNs(b.NsPerOp),
			pct,
			b.AllocsPerOp,
			formatBytes(b.BytesPerOp),
		)
	}
	numbers := fmt.Sprintf(`<p>Full median snapshot for engineers who want every line.</p>
<div class="proptable-wrap"><table class="proptable">
<thead><tr><th>What</th><th>Benchmark</th><th>Time</th><th>%% of 16.7&nbsp;ms</th><th>Allocs</th><th>Bytes</th></tr></thead>
<tbody>%s</tbody>
</table></div>
<p class="perfnote">%s<br>%s · %s</p>`,
		tableRows.String(),
		html.EscapeString(r.Note),
		html.EscapeString(genDay),
		html.EscapeString(machine),
	)

	howto := fmt.Sprintf(`<ol class="perflist">
<li><strong>Resolve theme once.</strong> <code>NewTheme</code> is a startup cost
(%s/op) — colors are not re-derived every frame.</li>
<li><strong>Cache expensive work.</strong> Warm SVG icons are %s/op and
<strong>0 allocs/op</strong> (test-pinned).</li>
<li><strong>Replay motion.</strong> Animations record ops once and translate them.</li>
<li><strong>Stop when settled.</strong> Idle redraw rate is <strong>0 frames/s</strong>.</li>
</ol>`,
		formatNs(theme.NsPerOp),
		formatNs(icon.NsPerOp),
	)

	reproduceSnippet := fmt.Sprintf(`make bench-doc

# Or:
go test -bench='Benchmark(SVGIconCacheHit|ButtonFrame|CardFrame|InputFrame|TabsFrame|SimpleGridFrame|ListView10k|Scrollable10k|NewTheme)$' \
  -benchmem -count=5 -benchtime=300ms .
go run ./cmd/lotusui bench-doc -o site/bench.json

go test github.com/ikaito-com/lotusui -run TestSVGIconCacheZeroAlloc

# Snapshot (%s): ListView %s/op · Scrollable %s/op · Button %s/op`,
		genDay,
		formatNs(list.NsPerOp),
		formatNs(scroll.NsPerOp),
		formatNs(btn.NsPerOp),
	)

	return &Page{
		Slug:   "performance",
		Title:  "Performance",
		Kicker: "Quiet when idle. Cheap on long lists. Measured — not marketed.",
		Intro:  intro,
		Sections: []Section{
			{Heading: "How we measure", Prose: measure},
			{Heading: "Long lists", Prose: lists},
			{Heading: "Everyday controls", Prose: controls},
			{Heading: "Idle windows", Prose: idle},
			{Heading: "How big is a tiny app?", Prose: ship},
			{Heading: "What we don’t compare", Prose: honest},
			{Heading: "Full snapshot", Prose: numbers},
			{Heading: "How we keep it cheap", Prose: howto},
			{
				Heading: "Reproduce",
				Prose:   `<p>Absolute ns/op move with CPU and Go version. Contracts and units do not.</p>`,
				Snippet: reproduceSnippet,
				Lang:    "sh",
			},
		},
	}
}
