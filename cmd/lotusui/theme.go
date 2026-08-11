package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"math"
	"os"
	"strings"

	lotusui "github.com/ikaito-com/lotusui"
)

// cmdTheme turns a brand color — or a full theme.json — into
// compilable Go: anchors are graded into 50…900 scales
// (lotusui.ScaleFrom), token overrides are applied, and the result is
// emitted as LITERALS the app owns. Because this runs at build time it
// can do what runtime theming can't: validate the palette's contrast
// and refuse to generate an unreadable one.
func cmdTheme(args []string) error {
	fs := flagSet("theme")
	anchor := fs.String("anchor", "", "brand color as #RRGGBB")
	config := fs.String("config", "", "theme.json with anchor/palette/radius/space/duration/textSize")
	pkg := fs.String("pkg", "main", "package name for the generated file")
	name := fs.String("name", "BrandPalette", "variable name for the generated palette")
	out := fs.String("o", "", "output file (default stdout)")
	strict := fs.Bool("strict", false, "fail (exit 1) on contrast warnings instead of just printing them")
	fs.Parse(args)

	cfg := themeConfig{}
	sourceStamp := ""
	if *config != "" {
		data, err := os.ReadFile(*config)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("%s: %v", *config, err)
		}
		// The stamp lets `lotusui verify` prove the generated file
		// matches this exact config — edits without regeneration fail CI.
		sourceStamp = shortHash(data)
	}
	if *anchor != "" {
		cfg.Anchor = *anchor
	}
	if cfg.Anchor == "" && len(cfg.Palette) == 0 {
		return fmt.Errorf("nothing to generate: pass -anchor '#RRGGBB' or -config theme.json")
	}

	p := lotusui.DefaultPalette
	set := map[string]color.NRGBA{} // emitted fields, in stable order below

	if cfg.Anchor != "" {
		c, err := parseHex(cfg.Anchor)
		if err != nil {
			return err
		}
		// BrandInk maps to C700, not C600: the ink must hold 4.5:1 on
		// the C50 subtle background, and C600 sits right at the line
		// for mid-lightness anchors.
		s := lotusui.ScaleFrom(c)
		for f, v := range map[string]color.NRGBA{
			"BrandSolid": s.C100, "BrandSubtle": s.C50, "BrandFg": s.C700,
			"BrandEmphasized": s.C200, "BrandContrast": s.C800,
		} {
			set[f] = v
		}
	}
	for tok, hex := range cfg.Palette {
		c, err := parseHex(hex)
		if err != nil {
			return fmt.Errorf("palette.%s: %v", tok, err)
		}
		set[tok] = c
	}
	for f, v := range set {
		if err := setToken(&p, f, v); err != nil {
			return err
		}
	}

	// Build-time design validation: a palette that fails contrast is a
	// bug, and build time is the only place it can be a build error.
	warnings := contrastWarnings(p)
	for _, w := range warnings {
		fmt.Fprintln(os.Stderr, "  contrast:", w)
	}
	if *strict && len(warnings) > 0 {
		return fmt.Errorf("%d contrast warning(s) with -strict", len(warnings))
	}

	src := generateTheme(*pkg, *name, cfg, set, sourceStamp)
	if *out == "" {
		fmt.Print(src)
		return nil
	}
	fmt.Printf("  theme -> %s\n", *out)
	return os.WriteFile(*out, []byte(src), 0o644)
}

// themeConfig is the theme.json schema: declarative input, Go output.
type themeConfig struct {
	Anchor   string            `json:"anchor"`
	Palette  map[string]string `json:"palette"`
	Radius   map[string]int    `json:"radius"`
	Space    map[string]int    `json:"space"`
	Duration map[string]int    `json:"duration"` // ms: fast/normal/slow
	TextSize int               `json:"textSize"`
}

// tokenOrder fixes emission order so regenerated files diff cleanly.
var tokenOrder = []string{
	"Bg", "BgSubtle", "BgMuted", "BgEmphasized", "BgPanel", "BgInverted",
	"BorderSubtle", "BorderMuted", "Border", "BorderEmphasized",
	"Fg", "FgMuted", "FgSubtle", "FgDisabled", "FgInverted",
	"BrandSolid", "BrandSubtle", "BrandFg", "BrandEmphasized", "BrandContrast",
	"Success", "SuccessBg", "Warning", "WarningBg", "Info", "InfoBg",
	"Danger", "DangerContrast", "DangerBg", "Overlay",
}

func setToken(p *lotusui.Palette, name string, c color.NRGBA) error {
	switch name {
	case "Bg":
		p.Bg = c
	case "BgPanel":
		p.BgPanel = c
	case "BgSubtle":
		p.BgSubtle = c
	case "BgMuted":
		p.BgMuted = c
	case "BgEmphasized":
		p.BgEmphasized = c
	case "BgInverted":
		p.BgInverted = c
	case "BorderSubtle":
		p.BorderSubtle = c
	case "BorderMuted":
		p.BorderMuted = c
	case "BorderEmphasized":
		p.BorderEmphasized = c
	case "Border":
		p.Border = c
	case "Fg":
		p.Fg = c
	case "FgMuted":
		p.FgMuted = c
	case "FgSubtle":
		p.FgSubtle = c
	case "FgDisabled":
		p.FgDisabled = c
	case "FgInverted":
		p.FgInverted = c
	case "BrandSolid":
		p.BrandSolid = c
	case "BrandSubtle":
		p.BrandSubtle = c
	case "BrandFg":
		p.BrandFg = c
	case "BrandEmphasized":
		p.BrandEmphasized = c
	case "BrandContrast":
		p.BrandContrast = c
	case "Success":
		p.Success = c
	case "SuccessBg":
		p.SuccessBg = c
	case "Warning":
		p.Warning = c
	case "WarningBg":
		p.WarningBg = c
	case "Info":
		p.Info = c
	case "InfoBg":
		p.InfoBg = c
	case "Danger":
		p.Danger = c
	case "DangerContrast":
		p.DangerContrast = c
	case "DangerBg":
		p.DangerBg = c
	case "Overlay":
		p.Overlay = c
	default:
		return fmt.Errorf("unknown palette token %q", name)
	}
	return nil
}

func generateTheme(pkg, name string, cfg themeConfig, set map[string]color.NRGBA, sourceStamp string) string {
	var b strings.Builder
	b.WriteString("// Code generated by lotusui theme; edit theme.json and regenerate.\n")
	if sourceStamp != "" {
		fmt.Fprintf(&b, "// source sha256:%s (checked by `lotusui verify`)\n", sourceStamp)
	}
	fmt.Fprintf(&b, "package %s\n\nimport (\n\t\"image/color\"\n", pkg)
	if len(cfg.Duration) > 0 {
		b.WriteString("\t\"time\"\n\n")
	} else {
		b.WriteString("\n")
	}
	b.WriteString("\tlotusui \"github.com/ikaito-com/lotusui\"\n)\n\n")
	fmt.Fprintf(&b, "// %s is the app's palette: lotusui's neutral defaults with the\n// tokens below overridden. Build the theme with %sTheme().\n", name, name)
	fmt.Fprintf(&b, "var %s = func() lotusui.Palette {\n\tp := lotusui.DefaultPalette\n", name)
	for _, f := range tokenOrder {
		if c, ok := set[f]; ok {
			fmt.Fprintf(&b, "\tp.%s = color.NRGBA{R: 0x%02X, G: 0x%02X, B: 0x%02X, A: 0x%02X}\n", f, c.R, c.G, c.B, c.A)
		}
	}
	fmt.Fprintf(&b, "\treturn p\n}()\n\n")

	fmt.Fprintf(&b, "// %sTheme builds the app's Theme with every configured option.\nfunc %sTheme() *lotusui.Theme {\n\treturn lotusui.NewTheme(\n\t\tlotusui.WithPalette(%s),\n", name, name, name)
	if r := cfg.Radius; len(r) > 0 {
		fmt.Fprintf(&b, "\t\tlotusui.WithRadius(lotusui.RadiusScale{XS: %d, SM: %d, MD: %d, LG: %d}),\n",
			pick(r, "xs", 4), pick(r, "sm", 6), pick(r, "md", 10), pick(r, "lg", 16))
	}
	if s := cfg.Space; len(s) > 0 {
		fmt.Fprintf(&b, "\t\tlotusui.WithSpace(lotusui.SpaceScale{XS: %d, SM: %d, MD: %d, LG: %d, XL: %d}),\n",
			pick(s, "xs", 4), pick(s, "sm", 8), pick(s, "md", 16), pick(s, "lg", 24), pick(s, "xl", 32))
	}
	if d := cfg.Duration; len(d) > 0 {
		fmt.Fprintf(&b, "\t\tlotusui.WithDuration(lotusui.DurationScale{Fast: %d * time.Millisecond, Normal: %d * time.Millisecond, Slow: %d * time.Millisecond}),\n",
			pick(d, "fast", 150), pick(d, "normal", 200), pick(d, "slow", 300))
	}
	if cfg.TextSize > 0 {
		fmt.Fprintf(&b, "\t\tlotusui.WithTextSize(%d),\n", cfg.TextSize)
	}
	fmt.Fprintf(&b, "\t)\n}\n")
	return b.String()
}

func pick(m map[string]int, k string, def int) int {
	if v, ok := m[k]; ok {
		return v
	}
	return def
}

// ── contrast (WCAG relative luminance) ──────────────────────────────────

func luminance(c color.NRGBA) float64 {
	lin := func(v uint8) float64 {
		f := float64(v) / 255
		if f <= 0.04045 {
			return f / 12.92
		}
		return math.Pow((f+0.055)/1.055, 2.4)
	}
	return 0.2126*lin(c.R) + 0.7152*lin(c.G) + 0.0722*lin(c.B)
}

func contrastRatio(a, b color.NRGBA) float64 {
	la, lb := luminance(a), luminance(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

// contrastWarnings checks the text-on-surface pairs the components
// actually paint. 4.5:1 is the WCAG AA threshold for normal text.
func contrastWarnings(p lotusui.Palette) []string {
	pairs := []struct {
		name string
		fg   color.NRGBA
		bg   color.NRGBA
	}{
		{"TextPrim on Page", p.Fg, p.Bg},
		{"TextPrim on Card", p.Fg, p.BgPanel},
		{"TextBody on Card", p.FgMuted, p.BgPanel},
		{"TextSec on Card", p.FgSubtle, p.BgPanel},
		{"OnBrand on Brand", p.BrandContrast, p.BrandSolid},
		{"BrandInk on Card", p.BrandFg, p.BgPanel},
		{"BrandInk on BrandSubtle", p.BrandFg, p.BrandSubtle},
		{"OnDanger on Danger", p.DangerContrast, p.Danger},
		{"Danger on DangerSurface", p.Danger, p.DangerBg},
		{"Healthy on HealthyBg", p.Success, p.SuccessBg},
		{"Attention on AttentionBg", p.Warning, p.WarningBg},
		{"Info on InfoBg", p.Info, p.InfoBg},
	}
	var out []string
	for _, pr := range pairs {
		if r := contrastRatio(pr.fg, pr.bg); r < 4.5 {
			out = append(out, fmt.Sprintf("%s is %.2f:1 (want ≥ 4.5:1)", pr.name, r))
		}
	}
	return out
}

func parseHex(s string) (color.NRGBA, error) {
	s = strings.TrimPrefix(s, "#")
	if len(s) != 6 {
		return color.NRGBA{}, fmt.Errorf("want '#RRGGBB', got %q", s)
	}
	var r, g, b uint8
	if _, err := fmt.Sscanf(s, "%02x%02x%02x", &r, &g, &b); err != nil {
		return color.NRGBA{}, fmt.Errorf("bad hex color %q: %v", s, err)
	}
	return color.NRGBA{R: r, G: g, B: b, A: 0xFF}, nil
}
