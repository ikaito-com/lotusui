package main

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

// cmdBenchDoc refreshes site/bench.json from go test -bench output
// (medians of -count runs) and optionally records the docs gallery
// WASM size. Commit the JSON with Performance page updates.
func cmdBenchDoc(args []string) error {
	fs := flagSet("bench-doc")
	out := fs.String("o", "site/bench.json", "output JSON path")
	count := fs.Int("count", 5, "go test -count (median is taken)")
	benchTime := fs.String("benchtime", "300ms", "go test -benchtime")
	wasm := fs.String("wasm", "", "optional path to a built gallery.wasm to record size")
	in := fs.String("in", "", "parse an existing go test -bench text file instead of running")
	fs.Parse(args)

	var text string
	if *in != "" {
		b, err := os.ReadFile(*in)
		if err != nil {
			return err
		}
		text = string(b)
	} else {
		pat := `Benchmark(SVGIconCacheHit|ButtonFrame|LabelFrame|BadgeFrame|CardFrame|InputFrame|CheckboxFrame|SwitchFrame|TabsFrame|SimpleGridFrame|ListView10k|Scrollable10k|NewTheme)$`
		cmd := exec.Command("go", "test",
			"-bench="+pat,
			"-benchmem",
			fmt.Sprintf("-count=%d", *count),
			"-benchtime="+*benchTime,
			".",
		)
		cmd.Stderr = os.Stderr
		outb, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("go test -bench: %w\n%s", err, outb)
		}
		text = string(outb)
		fmt.Print(text)
	}

	report, err := parseBenchOutput(text)
	if err != nil {
		return err
	}
	report.Generated = time.Now().Format(time.RFC3339)
	report.Note = fmt.Sprintf("layout/ops microbenchmarks — ops generation cost per frame, not painted GPU FPS. Medians of %d×%s runs.", *count, *benchTime)

	// Always preserve prior size artifacts, then overwrite only what this run measured.
	if prev, err := os.ReadFile(*out); err == nil {
		var old struct {
			WasmBytes            int64      `json:"wasmBytes"`
			WasmGzipBytes        int64      `json:"wasmGzipBytes"`
			WasmNote             string     `json:"wasmNote"`
			AppBytes             int64      `json:"appBytes"`
			AppNote              string     `json:"appNote"`
			HelloWasmBytes       int64      `json:"helloWasmBytes"`
			HelloWasmNote        string     `json:"helloWasmNote"`
			MinimalWasmBytes     int64      `json:"minimalWasmBytes"`
			MinimalWasmGzipBytes int64      `json:"minimalWasmGzipBytes"`
			MinimalWasmNote      string     `json:"minimalWasmNote"`
			WasmExecGzipBytes    int64      `json:"wasmExecGzipBytes"`
			WasmExecNote         string     `json:"wasmExecNote"`
			Peers                []peerSize `json:"peers"`
		}
		if json.Unmarshal(prev, &old) == nil {
			if old.WasmBytes > 0 {
				report.WasmBytes = old.WasmBytes
				report.WasmNote = old.WasmNote
			}
			if old.WasmGzipBytes > 0 {
				report.WasmGzipBytes = old.WasmGzipBytes
			}
			if old.AppBytes > 0 {
				report.AppBytes = old.AppBytes
				report.AppNote = old.AppNote
			}
			if old.MinimalWasmBytes > 0 {
				report.MinimalWasmBytes = old.MinimalWasmBytes
				report.MinimalWasmNote = old.MinimalWasmNote
			} else if old.HelloWasmBytes > 0 {
				report.MinimalWasmBytes = old.HelloWasmBytes
				report.MinimalWasmNote = old.HelloWasmNote
			}
			if old.MinimalWasmGzipBytes > 0 {
				report.MinimalWasmGzipBytes = old.MinimalWasmGzipBytes
			}
			if old.WasmExecGzipBytes > 0 {
				report.WasmExecGzipBytes = old.WasmExecGzipBytes
				report.WasmExecNote = old.WasmExecNote
			}
			if len(old.Peers) > 0 {
				report.Peers = old.Peers
			}
		}
	}
	if *wasm != "" {
		st, err := os.Stat(*wasm)
		if err != nil {
			return err
		}
		report.WasmBytes = st.Size()
		report.WasmNote = "docs gallery (web · GOOS=js GOARCH=wasm, -ldflags=-s -w); every component demo + fonts · gzip = Content-Encoding transfer size"
		if gz, err := gzipFileSize(*wasm); err == nil {
			report.WasmGzipBytes = gz
		}
	}

	// Default human notes when sizes exist but notes are empty/outdated.
	if report.AppBytes > 0 && report.AppNote == "" {
		report.AppNote = "minimal one-button desktop app (Theme + Button, darwin/amd64, -ldflags=-s -w)"
	}
	if report.MinimalWasmBytes > 0 && report.MinimalWasmNote == "" {
		report.MinimalWasmNote = "same minimal one-button app compiled to WASM (web)"
	}

	if err := os.MkdirAll(filepath.Dir(*out), 0o755); err != nil {
		return err
	}
	enc, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	enc = append(enc, '\n')
	if err := os.WriteFile(*out, enc, 0o644); err != nil {
		return err
	}
	fmt.Printf("  wrote %s (%d benches)\n", *out, len(report.Benches))
	return nil
}

type benchDocReport struct {
	Generated            string        `json:"generated"`
	Go                   string        `json:"go"`
	GOOS                 string        `json:"goos"`
	GOARCH               string        `json:"goarch"`
	CPU                  string        `json:"cpu"`
	Note                 string        `json:"note"`
	WasmBytes            int64         `json:"wasmBytes,omitempty"`
	WasmGzipBytes        int64         `json:"wasmGzipBytes,omitempty"`
	WasmNote             string        `json:"wasmNote,omitempty"`
	AppBytes             int64         `json:"appBytes,omitempty"`
	AppNote              string        `json:"appNote,omitempty"`
	MinimalWasmBytes     int64         `json:"minimalWasmBytes,omitempty"`
	MinimalWasmGzipBytes int64         `json:"minimalWasmGzipBytes,omitempty"`
	MinimalWasmNote      string        `json:"minimalWasmNote,omitempty"`
	WasmExecGzipBytes    int64         `json:"wasmExecGzipBytes,omitempty"`
	WasmExecNote         string        `json:"wasmExecNote,omitempty"`
	Peers                []peerSize    `json:"peers,omitempty"`
	Benches              []benchDocRow `json:"benches"`
}

type peerSize struct {
	Name     string `json:"name"`
	Surface  string `json:"surface"`
	Bytes    int64  `json:"bytes"`
	BytesRaw int64  `json:"bytesRaw,omitempty"`
	Encoding string `json:"encoding,omitempty"`
	Note     string `json:"note"`
}

type benchDocRow struct {
	Name        string  `json:"name"`
	NsPerOp     float64 `json:"nsPerOp"`
	BytesPerOp  int64   `json:"bytesPerOp"`
	AllocsPerOp int64   `json:"allocsPerOp"`
}

var (
	benchLineRe = regexp.MustCompile(`^(Benchmark\w+)-\d+\s+\d+\s+([\d.]+) ns/op\s+(\d+) B/op\s+(\d+) allocs/op`)
	metaGoosRe  = regexp.MustCompile(`^goos:\s+(\S+)`)
	metaGoarch  = regexp.MustCompile(`^goarch:\s+(\S+)`)
	metaCPU     = regexp.MustCompile(`^cpu:\s+(.+)$`)
)

func parseBenchOutput(text string) (benchDocReport, error) {
	r := benchDocReport{
		Go:     runtime.Version(),
		GOOS:   runtime.GOOS,
		GOARCH: runtime.GOARCH,
	}
	samples := map[string][]benchDocRow{}
	sc := bufio.NewScanner(strings.NewReader(text))
	for sc.Scan() {
		line := sc.Text()
		if m := metaGoosRe.FindStringSubmatch(line); m != nil {
			r.GOOS = m[1]
			continue
		}
		if m := metaGoarch.FindStringSubmatch(line); m != nil {
			r.GOARCH = m[1]
			continue
		}
		if m := metaCPU.FindStringSubmatch(line); m != nil {
			r.CPU = strings.TrimSpace(m[1])
			continue
		}
		m := benchLineRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		name := strings.TrimPrefix(m[1], "Benchmark")
		ns, _ := strconv.ParseFloat(m[2], 64)
		bytes, _ := strconv.ParseInt(m[3], 10, 64)
		allocs, _ := strconv.ParseInt(m[4], 10, 64)
		samples[name] = append(samples[name], benchDocRow{
			Name: name, NsPerOp: ns, BytesPerOp: bytes, AllocsPerOp: allocs,
		})
	}
	order := []string{
		"SVGIconCacheHit",
		"LabelFrame",
		"BadgeFrame",
		"ButtonFrame",
		"CardFrame",
		"CheckboxFrame",
		"SwitchFrame",
		"InputFrame",
		"TabsFrame",
		"SimpleGridFrame",
		"ListView10k",
		"Scrollable10k",
		"NewTheme",
	}
	for _, name := range order {
		rows := samples[name]
		if len(rows) == 0 {
			continue
		}
		r.Benches = append(r.Benches, medianBench(rows))
	}
	if len(r.Benches) == 0 {
		return r, fmt.Errorf("no benchmark lines parsed")
	}
	return r, nil
}

func medianBench(rows []benchDocRow) benchDocRow {
	sort.Slice(rows, func(i, j int) bool { return rows[i].NsPerOp < rows[j].NsPerOp })
	mid := rows[len(rows)/2]
	// Median bytes/allocs from the same mid-time sample is fine; if even
	// count, Go's /2 picks the upper middle — good enough for docs.
	return mid
}

// gzipFileSize returns the gzip BestCompression size of path (transfer-size proxy).
func gzipFileSize(path string) (int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	var n int64
	zw, err := gzip.NewWriterLevel(&countWriter{n: &n}, gzip.BestCompression)
	if err != nil {
		return 0, err
	}
	if _, err := io.Copy(zw, f); err != nil {
		_ = zw.Close()
		return 0, err
	}
	if err := zw.Close(); err != nil {
		return 0, err
	}
	return n, nil
}

type countWriter struct{ n *int64 }

func (c *countWriter) Write(p []byte) (int, error) {
	*c.n += int64(len(p))
	return len(p), nil
}
