package lotusui

import (
	"encoding/json"
	"fmt"
	"sort"

	"gioui.org/layout"
	"gioui.org/unit"
)

// Breakpoints are named min-widths (dp), mobile-first — Chakra's
// theme.breakpoints. Compiled once at NewTheme; per-frame resolve is
// BreakpointIndex (a short ascending walk, zero alloc).
//
// Apps override with WithBreakpoints or ParseBreakpointsJSON — never
// via registry.json.
type Breakpoints struct {
	names []string
	mins  []unit.Dp // parallel, ascending
}

// DefaultBreakpoints mirrors Chakra's common scale (dp).
var DefaultBreakpoints = MustBreakpoints([]Breakpoint{
	{Name: "base", Min: 0},
	{Name: "sm", Min: 480},
	{Name: "md", Min: 768},
	{Name: "lg", Min: 992},
	{Name: "xl", Min: 1280},
	{Name: "2xl", Min: 1536},
})

// Breakpoint is one named step for BreakpointsFrom / MustBreakpoints.
type Breakpoint struct {
	Name string
	Min  unit.Dp
}

// BreakpointsFrom builds a Breakpoints from unordered steps. Names must
// be unique; at least one step is required. Sorted ascending by Min.
func BreakpointsFrom(steps []Breakpoint) (Breakpoints, error) {
	if len(steps) == 0 {
		return Breakpoints{}, fmt.Errorf("lotusui: breakpoints: empty")
	}
	cp := append([]Breakpoint(nil), steps...)
	sort.SliceStable(cp, func(i, j int) bool {
		if cp[i].Min != cp[j].Min {
			return cp[i].Min < cp[j].Min
		}
		return cp[i].Name < cp[j].Name
	})
	seen := make(map[string]struct{}, len(cp))
	out := Breakpoints{
		names: make([]string, len(cp)),
		mins:  make([]unit.Dp, len(cp)),
	}
	for i, s := range cp {
		if s.Name == "" {
			return Breakpoints{}, fmt.Errorf("lotusui: breakpoints: empty name")
		}
		if _, ok := seen[s.Name]; ok {
			return Breakpoints{}, fmt.Errorf("lotusui: breakpoints: duplicate %q", s.Name)
		}
		seen[s.Name] = struct{}{}
		out.names[i] = s.Name
		out.mins[i] = s.Min
	}
	return out, nil
}

// MustBreakpoints is BreakpointsFrom, panicking on error — for package
// defaults and tests.
func MustBreakpoints(steps []Breakpoint) Breakpoints {
	bp, err := BreakpointsFrom(steps)
	if err != nil {
		panic(err)
	}
	return bp
}

// Len is the number of named steps.
func (b Breakpoints) Len() int { return len(b.mins) }

// Name returns the step name at i, or "" if out of range.
func (b Breakpoints) Name(i int) string {
	if i < 0 || i >= len(b.names) {
		return ""
	}
	return b.names[i]
}

// Min returns the min width (dp) at i.
func (b Breakpoints) Min(i int) unit.Dp {
	if i < 0 || i >= len(b.mins) {
		return 0
	}
	return b.mins[i]
}

// IndexOf returns the step index for name, or -1.
func (b Breakpoints) IndexOf(name string) int {
	for i, n := range b.names {
		if n == name {
			return i
		}
	}
	return -1
}

// ParseBreakpointsJSON reads `{"base":0,"sm":480,"md":768,…}` — values
// are dp. Apps ship this file; lotusui does not load it at runtime.
func ParseBreakpointsJSON(data []byte) (Breakpoints, error) {
	var raw map[string]float64
	if err := json.Unmarshal(data, &raw); err != nil {
		return Breakpoints{}, fmt.Errorf("lotusui: breakpoints json: %w", err)
	}
	if len(raw) == 0 {
		return Breakpoints{}, fmt.Errorf("lotusui: breakpoints json: empty object")
	}
	steps := make([]Breakpoint, 0, len(raw))
	for name, v := range raw {
		if v < 0 {
			return Breakpoints{}, fmt.Errorf("lotusui: breakpoints json: %q min < 0", name)
		}
		steps = append(steps, Breakpoint{Name: name, Min: unit.Dp(v)})
	}
	return BreakpointsFrom(steps)
}

// WithBreakpoints replaces Theme.Breakpoints (default: DefaultBreakpoints).
func WithBreakpoints(bp Breakpoints) ThemeOption {
	return func(t *Theme) {
		if bp.Len() == 0 {
			t.Breakpoints = DefaultBreakpoints
			return
		}
		t.Breakpoints = bp
	}
}

// BreakpointIndex is the largest step whose min width ≤ Max.X
// (mobile-first). Zero alloc; O(number of steps).
func (th *Theme) BreakpointIndex(gtx layout.Context) int {
	bp := th.Breakpoints
	if bp.Len() == 0 {
		bp = DefaultBreakpoints
	}
	w := gtx.Constraints.Max.X
	idx := 0
	for i := 1; i < len(bp.mins); i++ {
		if w >= gtx.Dp(bp.mins[i]) {
			idx = i
		}
	}
	return idx
}

// BreakpointName is the active step name (for demos / debugging).
func (th *Theme) BreakpointName(gtx layout.Context) string {
	bp := th.Breakpoints
	if bp.Len() == 0 {
		bp = DefaultBreakpoints
	}
	return bp.Name(th.BreakpointIndex(gtx))
}

// ── Responsive values (layout structure) ───────────────────────────────

type riStep struct {
	name string
	val  int
}

// ResponsiveInt is a mobile-first int (columns, spans). Build with
// Cols(base).At("md", 2). Resolve is zero-alloc.
type ResponsiveInt struct {
	base  int
	set   bool
	n     uint8
	steps [8]riStep
}

// Cols starts a responsive column/span value at base (always applies).
func Cols(base int) ResponsiveInt {
	return ResponsiveInt{base: base, set: true}
}

// At adds an override at a named breakpoint (Chakra object syntax).
// Unknown names are ignored at resolve time. Max 8 overrides.
func (r ResponsiveInt) At(name string, v int) ResponsiveInt {
	r.set = true
	if r.n < 8 {
		r.steps[r.n] = riStep{name: name, val: v}
		r.n++
	}
	return r
}

// Set reports whether any Cols/At configured this value.
func (r ResponsiveInt) Set() bool { return r.set }

// Resolve returns the value for the active breakpoint (mobile-first).
func (r ResponsiveInt) Resolve(th *Theme, gtx layout.Context) int {
	return r.ResolveAt(th, th.BreakpointIndex(gtx))
}

// ResolveAt is Resolve with a precomputed BreakpointIndex — use when
// resolving several responsive props in one frame.
func (r ResponsiveInt) ResolveAt(th *Theme, idx int) int {
	if !r.set {
		return 0
	}
	v := r.base
	if r.n == 0 {
		return v
	}
	bp := th.Breakpoints
	if bp.Len() == 0 {
		bp = DefaultBreakpoints
	}
	if idx < 0 {
		idx = 0
	}
	if idx >= bp.Len() {
		idx = bp.Len() - 1
	}
	for i := 0; i <= idx; i++ {
		name := bp.names[i]
		for j := uint8(0); j < r.n; j++ {
			if r.steps[j].name == name {
				v = r.steps[j].val
				break
			}
		}
	}
	return v
}

// Compile densifies against bp for callers that want a one-shot slice
// (startup). Per-frame code should prefer Resolve.
func (r ResponsiveInt) Compile(bp Breakpoints) []int {
	if bp.Len() == 0 {
		bp = DefaultBreakpoints
	}
	vals := make([]int, bp.Len())
	cur := r.base
	for i := 0; i < bp.Len(); i++ {
		name := bp.names[i]
		for j := uint8(0); j < r.n; j++ {
			if r.steps[j].name == name {
				cur = r.steps[j].val
				break
			}
		}
		vals[i] = cur
	}
	return vals
}

type rdStep struct {
	name string
	val  unit.Dp
}

// ResponsiveDp is a mobile-first dp (gaps, page max).
type ResponsiveDp struct {
	base  unit.Dp
	set   bool
	n     uint8
	steps [8]rdStep
}

// Dps starts a responsive dp at base.
func Dps(base unit.Dp) ResponsiveDp {
	return ResponsiveDp{base: base, set: true}
}

// At adds a dp override at a named breakpoint.
func (r ResponsiveDp) At(name string, v unit.Dp) ResponsiveDp {
	r.set = true
	if r.n < 8 {
		r.steps[r.n] = rdStep{name: name, val: v}
		r.n++
	}
	return r
}

// Set reports whether Dps/At configured this value.
func (r ResponsiveDp) Set() bool { return r.set }

// Resolve returns the dp for the active breakpoint.
func (r ResponsiveDp) Resolve(th *Theme, gtx layout.Context) unit.Dp {
	return r.ResolveAt(th, th.BreakpointIndex(gtx))
}

// ResolveAt is Resolve with a precomputed BreakpointIndex.
func (r ResponsiveDp) ResolveAt(th *Theme, idx int) unit.Dp {
	if !r.set {
		return 0
	}
	v := r.base
	if r.n == 0 {
		return v
	}
	bp := th.Breakpoints
	if bp.Len() == 0 {
		bp = DefaultBreakpoints
	}
	if idx < 0 {
		idx = 0
	}
	if idx >= bp.Len() {
		idx = bp.Len() - 1
	}
	for i := 0; i <= idx; i++ {
		name := bp.names[i]
		for j := uint8(0); j < r.n; j++ {
			if r.steps[j].name == name {
				v = r.steps[j].val
				break
			}
		}
	}
	return v
}

type rbStep struct {
	name string
	val  bool
}

// ResponsiveBool is a mobile-first bool (Show / hide).
type ResponsiveBool struct {
	base  bool
	set   bool
	n     uint8
	steps [8]rbStep
}

// Bools starts a responsive bool at base.
func Bools(base bool) ResponsiveBool {
	return ResponsiveBool{base: base, set: true}
}

// At adds a bool override at a named breakpoint.
func (r ResponsiveBool) At(name string, v bool) ResponsiveBool {
	r.set = true
	if r.n < 8 {
		r.steps[r.n] = rbStep{name: name, val: v}
		r.n++
	}
	return r
}

// Set reports whether Bools/At configured this value.
func (r ResponsiveBool) Set() bool { return r.set }

// Resolve returns the bool for the active breakpoint.
func (r ResponsiveBool) Resolve(th *Theme, gtx layout.Context) bool {
	return r.ResolveAt(th, th.BreakpointIndex(gtx))
}

// ResolveAt is Resolve with a precomputed BreakpointIndex.
func (r ResponsiveBool) ResolveAt(th *Theme, idx int) bool {
	if !r.set {
		return false
	}
	v := r.base
	if r.n == 0 {
		return v
	}
	bp := th.Breakpoints
	if bp.Len() == 0 {
		bp = DefaultBreakpoints
	}
	if idx < 0 {
		idx = 0
	}
	if idx >= bp.Len() {
		idx = bp.Len() - 1
	}
	for i := 0; i <= idx; i++ {
		name := bp.names[i]
		for j := uint8(0); j < r.n; j++ {
			if r.steps[j].name == name {
				v = r.steps[j].val
				break
			}
		}
	}
	return v
}

type rsStep struct {
	name string
	val  Size
}

// ResponsiveSize is a mobile-first Size — for Dialog / AlertDialog
// width presets (density Size on Button stays a plain Size).
type ResponsiveSize struct {
	base  Size
	set   bool
	n     uint8
	steps [8]rsStep
}

// Sizes starts a responsive Size at base.
func Sizes(base Size) ResponsiveSize {
	return ResponsiveSize{base: base, set: true}
}

// At adds a Size override at a named breakpoint.
func (r ResponsiveSize) At(name string, v Size) ResponsiveSize {
	r.set = true
	if r.n < 8 {
		r.steps[r.n] = rsStep{name: name, val: v}
		r.n++
	}
	return r
}

// Set reports whether Sizes/At configured this value.
func (r ResponsiveSize) Set() bool { return r.set }

// Resolve returns the Size for the active breakpoint.
func (r ResponsiveSize) Resolve(th *Theme, gtx layout.Context) Size {
	return r.ResolveAt(th, th.BreakpointIndex(gtx))
}

// ResolveAt is Resolve with a precomputed BreakpointIndex.
func (r ResponsiveSize) ResolveAt(th *Theme, idx int) Size {
	if !r.set {
		return SizeMD
	}
	v := r.base
	if r.n == 0 {
		return v
	}
	bp := th.Breakpoints
	if bp.Len() == 0 {
		bp = DefaultBreakpoints
	}
	if idx < 0 {
		idx = 0
	}
	if idx >= bp.Len() {
		idx = bp.Len() - 1
	}
	for i := 0; i <= idx; i++ {
		name := bp.names[i]
		for j := uint8(0); j < r.n; j++ {
			if r.steps[j].name == name {
				v = r.steps[j].val
				break
			}
		}
	}
	return v
}

// Show lays out w only when when resolves true — Chakra display:none
// analogue for stepped structure. Zero size when hidden.
func Show(th *Theme, gtx layout.Context, when ResponsiveBool, w layout.Widget) layout.Dimensions {
	if !when.Set() || !when.Resolve(th, gtx) {
		return layout.Dimensions{}
	}
	return w(gtx)
}
