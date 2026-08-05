package lotusui

import (
	"bytes"
	"embed"
	"fmt"
	"image"
	"image/color"
	"io/fs"
	"strings"
	"sync"
	"sync/atomic"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"

	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

// The UI's built-in SVG icons — fetched AND NORMALIZED at build time
// by `lotusui icons` per assets/icons/manifest.txt (see
// assets/icons/ICONS.md), then embedded here so the binary needs no
// network for its own artwork. Normalization (gradient flattening,
// em-size stripping) happens in the CLI, so the runtime pipeline is
// just read → tint → rasterize; icons added by hand instead of via
// the CLI must already be gradient-free.
//
//go:generate go run ./cmd/lotusui icons -manifest assets/icons/manifest.txt -out assets/icons -gen icons_gen.go -genpkg lotusui
//go:embed assets/icons/*.svg
var iconFS embed.FS

// iconSource is one registered icon filesystem: fsys rooted at dir.
type iconSource struct {
	fsys fs.FS
	dir  string
}

// extraIcons holds app-registered icon sources — read on cache misses
// only, so a plain mutex-guarded copy-on-write slice is plenty.
var (
	extraIconsMu sync.Mutex
	extraIcons   atomic.Pointer[[]iconSource]
)

// RegisterIconFS adds an app's own embedded icons to the icon
// namespace: fsys/dir/<name>.svg becomes SVGIcon(name, …). Registered
// sources are searched before the built-in set, so an app can also
// override a built-in icon by reusing its name. Call it from init or
// main — registration is cheap and safe from any goroutine, and the
// library stays exactly as heavy as the icons YOU embed.
//
//	//go:embed icons/*.svg
//	var appIcons embed.FS
//
//	func init() { lotusui.RegisterIconFS(appIcons, "icons") }
func RegisterIconFS(fsys fs.FS, dir string) {
	extraIconsMu.Lock()
	defer extraIconsMu.Unlock()
	var cur []iconSource
	if p := extraIcons.Load(); p != nil {
		cur = *p
	}
	next := make([]iconSource, 0, len(cur)+1)
	next = append(next, iconSource{fsys: fsys, dir: dir})
	next = append(next, cur...)
	extraIcons.Store(&next)
}

// readIconFile resolves an icon name: registered sources first (newest
// registration wins), then the built-in set.
func readIconFile(name string) ([]byte, error) {
	if p := extraIcons.Load(); p != nil {
		for _, src := range *p {
			if raw, err := fs.ReadFile(src.fsys, src.dir+"/"+name+".svg"); err == nil {
				return raw, nil
			}
		}
	}
	return iconFS.ReadFile("assets/icons/" + name + ".svg")
}

// iconKey identifies one rasterization: name+size+tint. A comparable
// struct, not a formatted string — the cache is consulted on every
// frame for every visible icon, and a fmt.Sprintf key would put an
// allocation on that hot path.
type iconKey struct {
	name string
	px   int
	col  color.NRGBA
}

// The icon cache is read-mostly to an extreme: every entry is written
// exactly once (the first frame that icon appears) and read on every
// frame after. So reads go through an atomic pointer to an immutable
// map — no lock, no contention, zero allocations — and the rare write
// path copies the map under a mutex (copy-on-write).
// TestSVGIconCacheZeroAlloc pins the hit path at 0 allocs/op.
var (
	svgIconCache atomic.Pointer[map[iconKey]paint.ImageOp]
	svgIconMu    sync.Mutex // serializes writers only
)

func init() {
	empty := map[iconKey]paint.ImageOp{}
	svgIconCache.Store(&empty)
}

// renderSVGIcon rasterizes <name>.svg at px×px, tinted (monochrome
// sets stroke with currentColor — substituted before parsing;
// full-color sets keep their own palette). The tint substitution is
// the ONLY runtime text work: everything else (gradient flattening,
// em-size stripping) happened at fetch time in the CLI. Returns nil
// if the icon isn't embedded or fails to parse — callers fall back to
// text-only rendering.
func renderSVGIcon(name string, px int, col color.NRGBA) *image.RGBA {
	raw, err := readIconFile(name)
	if err != nil {
		return nil
	}
	hex := fmt.Sprintf("#%02x%02x%02x", col.R, col.G, col.B)
	src := strings.ReplaceAll(string(raw), "currentColor", hex)
	icon, err := oksvg.ReadIconStream(bytes.NewReader([]byte(src)))
	if err != nil {
		return nil
	}
	icon.SetTarget(0, 0, float64(px), float64(px))
	img := image.NewRGBA(image.Rect(0, 0, px, px))
	scanner := rasterx.NewScannerGV(px, px, img, img.Bounds())
	icon.Draw(rasterx.NewDasher(px, px, scanner), 1)
	return img
}

func svgIconOp(name string, px int, col color.NRGBA) (paint.ImageOp, bool) {
	key := iconKey{name: name, px: px, col: col}
	if op, ok := (*svgIconCache.Load())[key]; ok {
		return op, op.Size() != image.Point{}
	}

	svgIconMu.Lock()
	defer svgIconMu.Unlock()
	cur := *svgIconCache.Load()
	if op, ok := cur[key]; ok { // another writer beat us to it
		return op, op.Size() != image.Point{}
	}
	var op paint.ImageOp
	if img := renderSVGIcon(name, px, col); img != nil {
		op = paint.NewImageOp(img)
	}
	// A failed render caches the zero ImageOp too — a missing icon must
	// not re-run the SVG pipeline every frame.
	next := make(map[iconKey]paint.ImageOp, len(cur)+1)
	for k, v := range cur {
		next[k] = v
	}
	next[key] = op
	svgIconCache.Store(&next)
	return op, op.Size() != image.Point{}
}

// SVGIcon lays out an embedded icon at the given dp size and tint.
// Renders nothing (zero dims beyond the reserved square) if the icon
// file isn't embedded, so screens degrade to their text.
func SVGIcon(name string, size unit.Dp, col color.NRGBA) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		px := gtx.Dp(size)
		op, ok := svgIconOp(name, px, col)
		sz := gtx.Constraints.Constrain(image.Pt(px, px))
		if ok {
			defer clip.Rect(image.Rectangle{Max: sz}).Push(gtx.Ops).Pop()
			op.Add(gtx.Ops)
			paint.PaintOp{}.Add(gtx.Ops)
		}
		return layout.Dimensions{Size: sz}
	}
}
