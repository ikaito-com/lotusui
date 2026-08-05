package docspages

func registryPage() *Page {
	return &Page{
		Slug:   "registry",
		Title:  "Registry",
		Kicker: "Own the code when you need to: vendor any component, keep updating it with real merges.",
		Intro: `<p>The default way to use lotusui is the module import — idiomatic Go, updated by
<code>go get</code>. The registry is for the moment a component must diverge: <code>lotusui
add</code> copies a component's source into your app, where it is yours to edit — and unlike a
snapshot, it stays updatable. Every capability here is BUILD-time: <code>registry.json</code>
is generated from the source, consumed by the CLI and by AI agents, and never read by app code
at runtime.</p>`,
		Sections: []Section{
			{
				Heading: "Add a component",
				Prose: `<p>The copy still imports the lotusui core (theme, scales, icons): exported
identifiers are qualified automatically by an AST pass, and unexported helpers are gathered
once into a CLI-owned companion file. A vendored Button keeps taking <code>*lotusui.Theme</code>
and drops into existing call sites. Components vendored side by side form a coherent SET —
they reference each other's local copies, never a mix of local and qualified.</p>`,
				Snippet: `go run github.com/ikaito-com/lotusui/cmd/lotusui add button
go run github.com/ikaito-com/lotusui/cmd/lotusui add dialog -dir internal/ui`,
			},
			{
				Heading: "Update — a real three-way merge",
				Prose: `<p>Every vendored file carries a stamp: component, version, and the hash of its
pristine form. On update, files you never touched are replaced cleanly. Files you customized go
through diff3 — the merge base is reconstructed EXACTLY, because the vendoring transform is a
pure function of (version, component set) and the Go module cache stores every version's
pristine source. Your edits, and only your edits, surface as diffs; conflicts are standard
markers, printed alongside the migration changelog sections between the two versions.</p>`,
				Snippet: `go run github.com/ikaito-com/lotusui/cmd/lotusui update            # everything under ./ui
go run github.com/ikaito-com/lotusui/cmd/lotusui update button -dry  # preview one component`,
			},
			{
				Heading: "Blocks",
				Prose: `<p>Blocks are ready-made compositions built only on the exported API — vendor
them as starting points and shape them to your app.</p>`,
				Snippet: `go run github.com/ikaito-com/lotusui/cmd/lotusui add login-form`,
			},
			{
				Heading: "Skills — teach your agent lotusui",
				Prose: `<p>lotusui ships agent skills in the module itself: the registry catalog, the
theming system, the add/update workflow and the changelog contract, as files the agent's
harness discovers. Install them into the app and re-run after upgrades so the skill
always matches the version you build against.</p>`,
				Snippet: `go run github.com/ikaito-com/lotusui/cmd/lotusui skills   # → .claude/skills/lotusui/`,
			},
			{
				Heading: "registry.json — the machine-readable catalog",
				Prose: `<p>Generated at build time from the source (<code>lotusui registry</code>, guarded
by <code>lotusui verify -registry</code>): every component and block with its files, computed
dependencies, carried helpers and pristine hash. Agents read it to know what exists and what
vendoring involves — app code never does.</p>`,
				Snippet: `{
  "name": "lotusui",
  "components": [
    {"name": "button", "files": ["button.go"], "hash": "…"},
    {"name": "card", "files": ["card.go"], "carried": ["cardShadow"], "hash": "…"}
  ]
}`,
			},
		},
	}
}

// ChangelogPage wraps pre-rendered changelog HTML (GFM).
func ChangelogPage(bodyHTML string) *Page {
	if bodyHTML == "" {
		bodyHTML = "<p>CHANGELOG.md could not be loaded.</p>"
	}
	return &Page{
		Slug:   "changelog",
		Title:  "Changelog",
		Kicker: "Every release, every API change — written so a developer can migrate from it alone.",
		Intro:  `<div class="mdpage">` + bodyHTML + `</div>`,
	}
}

func installationPage() *Page {
	return &Page{
		Slug:   "installation",
		Title:  "Installation",
		Kicker: "Add lotusui to a Go module and start a native desktop or mobile app.",
		Intro: `<p>lotusui is a themeable Go UI library for native desktop and mobile apps — a
semantic token palette, an embedded font, and the components that make an app read as one
product. Built on <a href="https://gioui.org">Gio</a>. Every demo on this site is the real
component running in your browser.</p>`,
		Sections: []Section{
			{
				Heading: "Install",
				Prose: `<p>lotusui is a normal Go module. The runtime stack brings its own platform
requirements (see the <a href="https://gioui.org/doc/install">Gio install docs</a>); on macOS,
Windows and js/wasm it works out of the box, Linux needs a few dev packages. When a component
must diverge from stock, you can also own its source — <code>lotusui add</code> vendors it into
your app and keeps it mergeable (see <a href="../registry/">Registry</a>).</p>`,
				Snippet: `go get github.com/ikaito-com/lotusui`,
				Lang:    "sh",
			},
			{
				Heading: "Scaffold an app",
				Prose: `<p>The <code>lotusui</code> CLI is the library's build-time companion: it scaffolds,
fetches icons, and generates themes — always via <code>go run</code>, so it needs no install and
its version is locked to the lotusui in your <code>go.mod</code>. <code>init</code> writes a
minimal themed app with the <code>go:generate</code> workflow already wired.</p>`,
				Snippet: `go mod init myapp
go get github.com/ikaito-com/lotusui
go run github.com/ikaito-com/lotusui/cmd/lotusui init
go mod tidy && go run .`,
				Lang: "sh",
			},
			{
				Heading: "Keep generated code honest",
				Prose: `<p>Everything the CLI generates (icon constants, themes) is guarded by two
standard mechanisms, so "someone forgot to regenerate" can't ship. First, typed constants make
most drift a <em>compile error</em>: an unfetched icon has no constant to call. Second,
<code>lotusui verify</code> is an offline, sub-second check — manifest ↔ SVGs ↔ constants, and
a hash stamp proving your generated theme matches its <code>theme.json</code> — safe in every
<code>make check</code> and in CI, no network. The scaffolded <code>AGENTS.md</code> teaches
AI coding agents the same rules, so assistants working in your repo regenerate instead of
hand-editing.</p>`,
				Snippet: `# Makefile
generate:   ## all build-time codegen
	go generate ./...

verify:     ## offline drift check — wire into make check and CI
	go run github.com/ikaito-com/lotusui/cmd/lotusui verify \
	    -manifest icons/manifest.txt -out icons -gen icons_gen.go \
	    -theme-config theme.json -theme-gen theme_gen.go`,
				Lang: "sh",
			},
			{
				Heading: "Versioning and upgrades",
				Prose: `<p>lotusui follows SemVer through Go modules: it lives at <code>v0.x</code> while the
component catalog is still filling out — breaking changes may land in minor versions — and
becomes <code>v1.0.0</code> when the parity ledger is clean, after which breaking changes mean
a <code>/v2</code> module path. Every API change is recorded in
<a href="https://github.com/ikaito-com/lotusui/blob/main/CHANGELOG.md">CHANGELOG.md</a> with
exact symbols and old→new forms — written specifically so an AI agent (or you) can migrate a
consuming app from the changelog alone. An exported-API baseline (<code>api.txt</code>) is
checked in CI, so no API change can ship undocumented.</p>`,
				Snippet: `go get github.com/ikaito-com/lotusui@latest
# read CHANGELOG.md for the versions you crossed, apply it, then:
go build ./...   # the compiler catches anything missed`,
				Lang: "sh",
			},
			{
				Heading: "Runtime versions move deliberately",
				Prose: `<p>Go's minimal version selection takes the <em>maximum</em> gioui.org requirement
across lotusui and your app, so a casual <code>go mod tidy</code> can silently upgrade the
UI runtime. Treat those bumps as deliberate changes, verified in lotusui and its consumers
together.</p>`,
			},
		},
	}
}

func quickstartPage() *Page {
	return &Page{
		Slug:   "quickstart",
		Title:  "Quickstart",
		Kicker: "A themed native window with a button, in one Go file.",
		Intro: `<p><code>NewTheme</code> builds the one Theme instance — palette plus the embedded
DM Sans font. Pass it to every component. The event loop is a standard Go app loop (Gio):
paint the canvas, lay out content in the readable column, hand the frame back.</p>`,
		Sections: []Section{
			{
				Heading: "A minimal app",
				Snippet: `package main

import (
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"

	lotusui "github.com/ikaito-com/lotusui"
)

func main() {
	go func() {
		w := new(app.Window)
		w.Option(app.Title("hello"))
		if err := loop(w); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func loop(w *app.Window) error {
	th := lotusui.NewTheme()
	var btn widget.Clickable
	var ops op.Ops
	for {
		switch e := w.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			lotusui.Fill(gtx, th.Palette.Bg)
			lotusui.LayoutPage(th, gtx, func(gtx layout.Context) layout.Dimensions {
				return lotusui.Button(th, &btn, "Hello", lotusui.ButtonProps{})(gtx)
			})
			e.Frame(gtx.Ops)
		}
	}
}`,
			},
		},
	}
}

func platformsPage() *Page {
	return &Page{
		Slug:      "platforms",
		Title:     "Platforms",
		Kicker:    "Desktop and mobile first — same Go module; the web is a compile target too.",
		Platforms: []string{"Desktop", "Mobile", "Web"},
		Intro: `<p>One Go codebase, one design language: lotusui embeds its font and paints a
consistent interface on every target, down to the typography. It sits on
<a href="https://gioui.org">Gio</a> — so desktop, mobile, and WebAssembly are compile
targets of the same module, not ports. The platform decides input (pointer or touch) and
packaging, nothing else. The few places where a platform genuinely differs are called out
across this site with badges like the ones above.</p>`,
		Sections: []Section{
			{
				Heading:   "Desktop",
				Platforms: []string{"macOS", "Windows", "Linux"},
				Prose: `<p>A first-class home for lotusui apps. Hover affordances (row pills, button shades,
pointer cursors) do their best work here, and on macOS the
<a href="../seamless-window/">seamless window</a> gives native edge-to-edge chrome — no title
bar, traffic lights kept. Windows and Linux run with standard decorations.</p>`,
			},
			{
				Heading:   "Mobile",
				Platforms: []string{"Android", "iOS"},
				Prose: `<p>Touch-ready by construction: the theme's tap targets default to a 44dp
finger size, scrolling (Scrollable, ListView) is native touch scrolling, and text input goes
through the platform IME. Hover states simply never fire — and that's safe by design, because
in lotusui hover is always an <em>affordance</em>, never the only signal: selection and
active states are explicit props (<code>active</code>, <code>Value</code>, <code>Sel</code>),
so nothing becomes unreachable without a pointer.</p>`,
			},
			{
				Heading:   "Web",
				Platforms: []string{"WASM"},
				Prose: `<p>The same Go module reaches the browser via WebAssembly — Gio’s first-class
<code>GOOS=js GOARCH=wasm</code> target, not a separate port. This documentation site
<em>is</em> a lotusui WASM app (<code>site/docsapp</code>): chrome, nav, Previews and
Code tabs are real widgets in one binary. Build once with
<code>-ldflags=&quot;-s -w&quot;</code>; fonts are embedded (several MB raw, less over the wire
compressed). Hero screenshots ship as static <code>media/</code> siblings so the wasm
stays leaner.</p>`,
				Snippet: `GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o app.wasm .
# serve with wasm_exec.js from $(go env GOROOT)/lib/wasm/
# and a <div id="giowindow"> for the canvas host`,
				Lang: "sh",
			},
			{
				Heading: "What's platform-specific in the API",
				Prose: `<p>Almost nothing. The one platform-gated symbol is
<code>MakeSeamlessWindow</code>, which exists only on macOS (behind a build tag) — call it from
a small <code>main_darwin.go</code> and every other target compiles cleanly without it.
Everything else in the library builds on every target.</p>`,
			},
		},
	}
}

func principlesPage() *Page {
	return &Page{
		Slug:   "design-principles",
		Title:  "Design principles",
		Kicker: "The rules that make an app built on lotusui read as one product.",
		Intro: `<p>lotusui owns the <em>mechanisms</em> of a design language. The rules are few and
fixed — components enforce them so screens don't have to remember them. The look itself is
yours: every rule below operates on theme tokens, so a custom palette keeps the discipline
while changing the skin (see <a href="../theme/">Theme &amp; palette</a>).</p>`,
		Sections: []Section{
			{
				Heading: "One accent, quiet grays",
				Prose: `<p>Neutral grays by default: panels floating on a tinted canvas, hierarchy from
size and color rather than bold shouting. There is exactly one accent — yours —
and red belongs to Delete-class actions exclusively. Status colors always render as a pastel
background with deep ink text, never saturated fill with white text.</p>`,
			},
			{
				Heading: "State lives with the caller",
				Prose: `<p>The core contract: components wrap arbitrary content and know nothing
about what's inside. Visibility, values, and selection belong to the caller —
<code>widget.Clickable</code> for buttons, <code>isOpen/onClose</code> for the dialog,
<code>Value</code> on checkbox and switch. Input rules reject wrong characters at the input
layer instead of flagging them after the fact.</p>`,
			},
			{
				Heading: "Reveal, never reflow",
				Prose: `<p>Motion is a viewport trick: panes live at natural width on a strip and are only
ever revealed or hidden by its edges — content never re-flows mid-animation. One shared
animation clock drives every transition, so everything in an app moves identically.</p>`,
			},
		},
	}
}

func themePage() *Page {
	return &Page{
		Slug:   "theme",
		Title:  "Theming",
		Kicker: "Paired role tokens, color scales, schemes — resolved once at NewTheme.",
		Intro: `<p><code>Palette</code> is the single source of truth for color: named semantic tokens
instead of raw hex scattered through views. Tokens come in <em>pairs</em> — a fill and its
readable ink (<code>BrandSolid</code>+<code>BrandContrast</code>,
<code>Danger</code>+<code>DangerContrast</code>, each status ink with its pastel background) —
so a component that paints a surface always knows what stays readable on it. Dark mode is just
another palette: see <a href="../dark-mode/">Dark mode</a>.</p>`,
		Sections: []Section{
			{
				Heading: "Tokens",
				Prose: `<p>The standard semantic vocabulary: background ladder (<code>Bg</code>,
<code>BgSubtle</code>, <code>BgMuted</code>, <code>BgEmphasized</code>, <code>BgPanel</code>,
<code>BgInverted</code>), borders (<code>BorderSubtle</code> → <code>BorderEmphasized</code>),
a foreground ladder (<code>Fg</code> → <code>FgDisabled</code>, plus <code>FgInverted</code>),
the brand slots (<code>BrandSolid</code>, <code>BrandSubtle</code>, <code>BrandFg</code>,
<code>BrandEmphasized</code>, <code>BrandContrast</code>), four status pairs, the dialog
<code>Overlay</code> and the keyboard <code>FocusRing</code>. <code>Space</code> is the 8pt
spacing scale; <code>Radius</code> the corner scale; <code>Duration</code> the motion
ladder (Fast/Normal/Slow).</p>`,
				Snippet: `th := lotusui.NewTheme()

// Semantic tokens, never raw hex in views:
bg := th.Palette.Bg   // the canvas
ink := th.Palette.Fg  // primary text

// Scales:
gap := th.Space.MD        // 16dp
r := th.Radius.MD         // 10dp
hover := th.Duration.Fast // 150ms`,
				Demo:  "palette",
				DemoH: 560,
			},
			{
				Heading: "Duration — motion timing",
				Prose: `<p>Chakra-style duration tokens on the theme — not a per-component prop.
<code>Fast</code> (150ms) eases hover/color on Button and HoverRow;
<code>Normal</code> (200ms) drives switch, accordion, and Split;
<code>Slow</code> (300ms) is dialog / VSlide entrance. Override with
<code>WithDuration</code>; a zero step snaps (reduced motion).</p>`,
				Snippet: `th := lotusui.NewTheme(lotusui.WithDuration(lotusui.DurationScale{
	Fast:   120 * time.Millisecond,
	Normal: 180 * time.Millisecond,
	Slow:   280 * time.Millisecond,
}))`,
			},
			{
				Heading: "If you come from the web",
				Prose: `<p>The tokens are a direct transposition of the semantic-token vocabulary web
component libraries use — dotted paths become CamelCase, because in Go a field can't be both a
color and a namespace (<code>fg</code> can't simultaneously be <code>Fg</code> and hold
<code>Fg.Muted</code>; flattening keeps the most-used token the shortest).</p>
<div class="proptable-wrap"><table class="proptable">
<thead><tr><th>Web token</th><th>lotusui</th></tr></thead>
<tbody>
<tr><td><code>bg</code>, <code>bg.subtle</code>, <code>bg.muted</code>, <code>bg.emphasized</code>, <code>bg.panel</code>, <code>bg.inverted</code></td><td><code>Bg</code>, <code>BgSubtle</code>, <code>BgMuted</code>, <code>BgEmphasized</code>, <code>BgPanel</code>, <code>BgInverted</code></td></tr>
<tr><td><code>fg</code>, <code>fg.muted</code>, <code>fg.subtle</code>, <code>fg.inverted</code></td><td><code>Fg</code>, <code>FgMuted</code>, <code>FgSubtle</code>, <code>FgInverted</code> (+ <code>FgDisabled</code> for placeholders)</td></tr>
<tr><td><code>border</code>, <code>border.muted</code>, <code>border.subtle</code>, <code>border.emphasized</code></td><td><code>Border</code>, <code>BorderMuted</code>, <code>BorderSubtle</code>, <code>BorderEmphasized</code></td></tr>
<tr><td><code>colorPalette.solid / .subtle / .fg / .emphasized / .contrast</code></td><td><code>BrandSolid</code>, <code>BrandSubtle</code>, <code>BrandFg</code>, <code>BrandEmphasized</code>, <code>BrandContrast</code></td></tr>
<tr><td><code>fg.success / bg.success</code> (and warning, info, error)</td><td><code>Success</code>/<code>SuccessBg</code>, <code>Warning</code>/<code>WarningBg</code>, <code>Info</code>/<code>InfoBg</code>, <code>Danger</code>/<code>DangerBg</code>/<code>DangerContrast</code></td></tr>
<tr><td>focus ring / dialog overlay</td><td><code>FocusRing</code> / <code>Overlay</code></td></tr>
</tbody></table></div>`,
			},
			{
				Heading: "Schemes — variants carry their role",
				Prose: `<p>A <code>Scheme</code> is a semantic color scheme: how one color role renders
across a component's variants — including its interaction steps (base, hover, active). The
palette exposes three — <code>Accent()</code> (the brand), <code>Neutral()</code> (the greys),
and <code>DangerScheme()</code> (destructive actions only) — and each VARIANT resolves its own:
Default→Accent, Secondary/Outline/Ghost→Neutral, Destructive→Danger. Every scheme field is a
palette token or scale step — nothing hard-coded — so a custom palette propagates to every
variant and every interaction state.</p>`,
				Snippet: `accent := th.Palette.Accent()
// accent.Solid / SolidHover / SolidActive, OnSolid,
// accent.Subtle / SubtleHover / SubtleActive, OnSubtle, Outline`,
			},
			{
				Heading: "Per-instance color — the precedence ladder",
				Prose: `<p>Three layers, most-derived wins. Default: the variant's role scheme from the
theme. Per instance: <code>Color</code> (any <code>ColorScale</code>) re-colors the variant —
one anchor in, the whole .500/.600/.700 interaction ladder out, so a custom color can never
break its own hover and pressed states. Full control: <code>Scheme</code> overrides every slot
by hand.</p>`,
				Snippet: `lotusui.Button(th, &b, "Save", lotusui.ButtonProps{})                      // theme role
lotusui.Button(th, &b, "Go", lotusui.ButtonProps{Color: lotusui.Teal})     // scale — ladder derived
soft := lotusui.Teal.SoftScheme()
lotusui.Button(th, &b, "Go", lotusui.ButtonProps{Scheme: &soft})           // manual slots`,
			},
			{
				Heading: "Color scales",
				Prose: `<p>A <code>ColorScale</code> is one hue graded 50…900 — 50 the faintest tint, 500
the anchor, 900 the darkest ink. Ten stock scales ship (<code>Gray</code> through
<code>Pink</code>), and <code>ScaleFrom</code> grades a complete scale from any single color —
one brand color in, ten usable steps out. A scale becomes a button scheme via
<code>.Scheme()</code> (saturated, .500/.600/.700 interaction steps) or
<code>.SoftScheme()</code> (pastel).</p>
<p>Your brand's scale needs no setup at all: <code>th.BrandScale</code> is graded automatically
from the palette's <code>BrandFg</code> when the theme is built — define one custom brand color
and the whole 50…900 ladder is already there.</p>`,
				Snippet: `// Automatic — derived from your palette at NewTheme:
tint := th.BrandScale.C50
ink := th.BrandScale.C700
scheme := th.BrandScale.Scheme()

// Or grade any color on the spot:
mint := lotusui.ScaleFrom(color.NRGBA{R: 0x31, G: 0x97, B: 0x95, A: 0xFF})`,
				Demo:  "scales",
				DemoH: 560,
			},
			{
				Heading: "Multiple themes",
				Prose: `<p>A theme is plain data — a Palette is ~200 bytes — so an app can compile in as
many looks as it wants and let its users switch: build the themes once at startup, swap a
pointer on selection. No parsing, no assets, no measurable weight. The palette picker in this
site's top bar is exactly that pattern live: eleven presets compiled into both the pages and
the wasm demos; picking one restyles everything on the fly.</p>`,
				Snippet: `var themes = map[string]*lotusui.Theme{
	"lavender": lotusui.NewTheme(),
	"teal":     lotusui.NewTheme(lotusui.WithPalette(tealPalette)),
	"rose":     lotusui.NewTheme(lotusui.WithPalette(rosePalette)),
}

th := themes[userChoice] // switching themes is a pointer swap`,
			},
			{
				Heading: "Theming",
				Prose: `<p>Customization flows through <code>NewTheme</code> options and resolves once, at
construction — a fully custom look has zero per-frame cost. Components read every color from
<code>th.Palette</code>, every corner from <code>th.Radius</code>, every gap from
<code>th.Space</code>, every motion step from <code>th.Duration</code>; swap any of
them and the whole library follows consistently.
<code>WithFaces</code> replaces the embedded DM Sans with your brand font.</p>
<p>The fastest start: generate a palette from your brand color —
<code>go run github.com/ikaito-com/lotusui/cmd/lotusui theme -anchor '#319795' -o theme_gen.go</code>
grades the anchor into a scale, maps it onto the brand tokens, and emits plain Go you own.</p>
<p>For the full declarative workflow, keep a <code>theme.json</code>
(anchor, token overrides, radius/space/duration scales, text size) and regenerate with
<code>-config theme.json</code>: designers edit JSON, the build turns it into typed literals,
and the runtime never parses anything. Because this runs at build time it also
<em>validates</em>: every text-on-surface pair is contrast-checked (WCAG 4.5:1), and
<code>-strict</code> makes an unreadable palette a build failure instead of a shipped bug.</p>`,
				Snippet: `brand := lotusui.DefaultPalette
brand.BrandSolid = color.NRGBA{R: 0xD9, G: 0xF3, B: 0xE4, A: 0xFF}
brand.BrandFg = color.NRGBA{R: 0x11, G: 0x7A, B: 0x3D, A: 0xFF}

th := lotusui.NewTheme(
	lotusui.WithPalette(brand),
	lotusui.WithRadius(lotusui.RadiusScale{SM: 4, MD: 8, LG: 12}),
	lotusui.WithDuration(lotusui.DurationScale{
		Fast: 120 * time.Millisecond, Normal: 180 * time.Millisecond, Slow: 280 * time.Millisecond,
	}),
	lotusui.WithTextSize(15),
)`,
			},
		},
	}
}

func darkModePage() *Page {
	return &Page{
		Slug:   "dark-mode",
		Title:  "Dark mode",
		Kicker: "Not a mode — a palette. Build both themes at startup, swap a pointer.",
		Intro: `<p>lotusui has no dark <em>mode</em>: dark is a <code>Palette</code> like any other,
applied with <code>WithPalette</code>. <code>DefaultDarkPalette</code> ships in the box — the
same lavender brand on a deep cool-gray canvas, every ladder keeping the light palette's ORDER
(faint → prominent), so components never know which world they're in. Try it live: pick
<strong>Midnight</strong> in this site's palette picker — every demo on every page flips.</p>`,
		Sections: []Section{
			{
				Heading: "Usage",
				Prose: `<p>Construct both themes once at startup; following the system appearance (or a
user setting) is a pointer swap — no rebuild, no reload, no per-frame cost.</p>`,
				Snippet: `light := lotusui.NewTheme()
dark := lotusui.NewTheme(lotusui.WithPalette(lotusui.DefaultDarkPalette))

th := light
if prefersDark {
	th = dark
}`,
			},
			{
				Heading: "Custom dark palettes",
				Prose: `<p>A dark palette follows two rules. Panels sit slightly LIGHTER than the canvas —
elevation still reads as "closer to the light", exactly as white cards do on the tinted light
page. And status pairs invert their construction: light ink on a deep tinted well, instead of
deep ink on a pastel. Start from <code>DefaultDarkPalette</code> and override tokens the same
way you would the light one.</p>`,
				Snippet: `night := lotusui.DefaultDarkPalette
night.BrandSolid = myBrand      // pastels pop on dark even better
night.BrandFg = myBrandLight    // the readable ink flips to a light sibling

dark := lotusui.NewTheme(lotusui.WithPalette(night))`,
			},
			{
				Heading: "Contrast stays checked",
				Prose: `<p>The build-time theme generator validates dark palettes with the same WCAG
contrast checks as light ones — <code>lotusui theme -config theme.json -strict</code> makes an
unreadable pairing a build failure, whichever world it is in.</p>`,
			},
		},
	}
}

func typographyPage() *Page {
	return &Page{
		Slug:   "typography",
		Title:  "Typography",
		Kicker: "A fixed label scale in DM Sans — hierarchy from size and color, not bold shouting.",
		Intro: `<p>Every piece of text uses one of seven named styles. The scale is deliberately
small: one hero per screen, semibold only for titles, and quiet grays doing the hierarchy
work.</p>`,
		Sections: []Section{
			{
				Heading: "The scale",
				Prose: `<p>Each helper returns a <code>material.LabelStyle</code>, so call <code>.Layout(gtx)</code>
— or pass <code>.Layout</code> where a widget is expected. <code>Sp(th, ratio)</code> derives any further
size from the one base instead of scattering magic numbers.</p>`,
				Snippet: `lotusui.LabelHero(th, "Projects")            // 20sp — the screen's H1
lotusui.LabelTitle(th, "Add an item")        // 16sp semibold
lotusui.LabelCardTitle(th, "A card")         // 14sp semibold
lotusui.LabelBody(th, "Primary row text")    // 14sp
lotusui.LabelMeta(th, "Secondary text")      // 13sp
lotusui.LabelCaption(th, "2 minutes ago")    // 12sp
lotusui.SectionLabel(th, "GROUP")            // quiet group caption`,
				Demo:  "typography",
				DemoH: 320,
			},
		},
	}
}

func layoutPage() *Page {
	return &Page{
		Slug:   "layout",
		Title:  "Layout",
		Kicker: "The readable page column, scrolling, titles, and virtualized lists.",
		Intro: `<p>Screen scaffolding: <code>LayoutPage</code> caps content at a readable width and
centers it; <code>Scrollable</code> makes mixed content reachable; <code>ListView</code> is the
virtualized list for real data. How those pieces (plus Wrap, SimpleGrid, Tabs) reflow when the
window changes size is the <a href="../responsive/">Responsive</a> guide.</p>`,
		Sections: []Section{
			{
				Heading: "The page column",
				Prose: `<p><code>LayoutPage</code> is the readable column: content capped at 920dp,
centered, padded — the difference between a window and a designed page.</p>`,
				Snippet: `lotusui.LayoutPage(th, gtx, func(gtx C) D {
	return content(gtx)
})`,
			},
			{
				Heading: "TitleWithIcons",
				Prose: `<p>A lotusui extension: an in-content / section title with trailing action
icons. <code>TopBar</code> is screen chrome (centered title, optional <em>leading</em>);
<code>TitleWithIcons</code> is content chrome (start-aligned <code>LabelTitle</code>, trailing
icons). Empty icons → title alone.</p>`,
				Snippet: `lotusui.TitleWithIcons(th, "Documents",
	lotusui.SVGIconButton(th, &add, lotusui.IconAdd, 20, false),
)`,
				Demo:  "layout/0",
				DemoH: 100,
			},
			{
				Heading: "Two trailing icons",
				Snippet: `lotusui.TitleWithIcons(th, "Documents",
	lotusui.SVGIconButton(th, &add, lotusui.IconAdd, 20, false),
	lotusui.SVGIconButton(th, &more, lotusui.IconSettings, 20, false),
)`,
				Demo:  "layout/1",
				DemoH: 100,
			},
			{
				Heading: "Long lists",
				Prose: `<p>For a screen's mixed content, <code>Scrollable</code>; the moment a collection
can grow with data, the virtualized <a href="../listview/">ListView</a> component.</p>`,
			},
		},
		Props: []Prop{
			{"TitleWithIcons(th, title, icons…)", "widget", "LabelTitle + trailing icons (Space.XS between icons only)."},
			{"TopBar(th, title, leading)", "widget", "Screen chrome — centered title, optional leading."},
			{"LayoutPage(th, gtx, content)", "widget", "Readable column; cap th.PageMax (default 920) or PageMaxAt."},
		},
	}
}

func stackPage() *Page {
	return &Page{
		Slug:   "stack",
		Title:  "Stack",
		Kicker: "VStack and HStack: children size themselves, the gap is the whole job.",
		Intro: `<p>A lotusui extension: the stacking primitives that stop screens hand-weaving spacer
children between rows. <code>VStack</code> stacks vertically, <code>HStack</code> horizontally
with children centered on the cross axis. For flowing chips that wrap to the next line, see
<a href="../wrap/">Wrap</a>. For 1dp rules, see <a href="../separator/">Separator</a>.</p>`,
		Sections: []Section{
			InstallSection("stack"),
			{
				Heading: "VStack",
				Snippet: `lotusui.VStack(th.Space.SM, first, second, third)`,
				Demo:    "stack/0",
				DemoH:   140,
			},
			{
				Heading: "HStack",
				Prose:   `<p>Children center on the cross axis, so a switch and its label sit on one visual line. Never wraps — under a narrow Max.X, Rigid children are squeezed; use <a href="../wrap/">Wrap</a> instead.</p>`,
				Snippet: `lotusui.HStack(th.Space.SM, one, two, three)`,
				Demo:    "stack/1",
				DemoH:   110,
			},
			{
				Heading: "Spacers",
				Prose: `<p><code>Spacer</code>/<code>HSpacer</code> are fixed gaps for ad-hoc composition when you
are not using a Stack.</p>`,
				Snippet: `lotusui.Spacer(th.Space.MD)   // vertical
lotusui.HSpacer(th.Space.SM)  // horizontal`,
			},
		},
		Props: []Prop{
			{"VStack(gap, children...)", "unit.Dp", "Vertical stack; the gap is the whole job."},
			{"HStack(gap, children...)", "unit.Dp", "Horizontal stack, children centered on the cross axis. Never wraps."},
			{"Spacer(h) / HSpacer(w)", "unit.Dp", "Fixed gaps for ad-hoc composition."},
		},
	}
}

func wrapPage() *Page {
	return &Page{
		Slug:   "wrap",
		Title:  "Wrap",
		Kicker: "Flex-wrap row: chips and labels flow left-to-right, then the next line.",
		Intro: `<p>A lotusui extension (Chakra Wrap / CSS flex-wrap). Lives beside
<a href="../stack/">Stack</a> in source (<code>lotusui add stack</code>).
Measures each child at its intrinsic width so labels never squeeze into one-character
columns the way <code>HStack</code> can under a narrow <code>Max.X</code>. Prefer
<a href="../simplegrid/">SimpleGrid</a> when you want equal cells and a fixed column count.
See the <a href="../responsive/">Responsive</a> guide for how Wrap fits the full toolkit.</p>`,
		Sections: []Section{
			InstallSection("stack"),
			{
				Heading: "Usage",
				Prose: `<p>Same gap between items and between lines. Cross-axis <code>align</code> is per line
(<code>layout.Middle</code> for chips and icons). A child wider than Max.X alone is laid out
at Max.X. The live demo uses the box width — <strong>narrow or widen the browser</strong>
and the chips reflow onto more or fewer lines.</p>`,
				Snippet: `lotusui.Wrap(th.Space.SM, layout.Middle,
	chip("Design"), chip("Engineering"), chip("Product"),
	chip("Marketing"), chip("Sales"), chip("Support"),
	// …)`,
				Demo:  "wrap/0",
				DemoH: 160,
			},
			{
				Heading: "Versus HStack",
				Prose: `<p><code>HStack</code> never wraps — under a narrow Max.X, Rigid children are
squeezed. <code>Wrap</code> measures intrinsic widths and starts a new line.</p>`,
				Snippet: `// Prefer Wrap for flowing badges/chips:
lotusui.Wrap(th.Space.SM, layout.Middle, chips...)
// Prefer HStack for a single non-wrapping row (switch + label):
lotusui.HStack(th.Space.SM, switchW, label)`,
			},
		},
		Props: []Prop{
			{"Wrap(gap, align, children...)", "unit.Dp, Alignment", "Flex-wrap row; intrinsic widths; same gap for items and lines."},
		},
	}
}

func responsivePage() *Page {
	return &Page{
		Slug:   "responsive",
		Title:  "Responsive",
		Kicker: "Continuous Max.X plus Theme breakpoints — reflow by default, stepped structure when you need it.",
		Intro: `<p>Every widget lays out from the <code>Max.X</code> it is given this frame — that is
the continuous layer (<a href="../wrap/">Wrap</a>, auto-fit
<a href="../simplegrid/">SimpleGrid</a>). For Chakra-style stepped structure
(<code>columns={{ base: 1, md: 2, lg: 4 }}</code>), Theme owns named breakpoints and layout
props resolve against them with a zero-alloc walk.</p>
<p><strong>Resize the browser</strong> on the demos below. Apps override breakpoint sizes with
JSON loaded at <code>NewTheme</code> — not <code>registry.json</code>.</p>`,
		Sections: []Section{
			{
				Heading: "The doctrine",
				Prose: `<p><strong>Pass the width you have; never invent a fake one.</strong> Gio hands every
layout a <code>layout.Constraints</code>. lotusui components honor it. The anti-pattern is
clamping <code>gtx.Constraints.Max.X</code> to a constant “for the demo.”</p>
<p>Two layers: <em>continuous</em> reflow (Wrap, minChildWidth SimpleGrid) and <em>stepped</em>
structure (Theme breakpoints + <code>Cols</code> / <code>Dps</code> / <code>Bools</code> /
<code>Show</code>). Prefer continuous when equal tiles should grow smoothly; use stepped when
column count or visibility must jump at named widths.</p>`,
			},
			{
				Heading: "Theme breakpoints",
				Prose: `<p>Defaults (dp): <code>base=0</code>, <code>sm=480</code>, <code>md=768</code>,
<code>lg=992</code>, <code>xl=1280</code>, <code>2xl=1536</code>. Mobile-first: each step whose
min ≤ Max.X overrides. Resolve is <code>Theme.BreakpointIndex</code> — O(steps), zero alloc.
Layout structure only (columns, spans, gaps, page max, show/hide) — not every visual prop.</p>`,
				Snippet: `// App-owned JSON (dp mins) — load at startup:
bp, err := lotusui.ParseBreakpointsJSON(jsonBytes)
th := lotusui.NewTheme(lotusui.WithBreakpoints(bp))

// Stepped columns (Chakra object syntax):
lotusui.Grid{
  Cols: lotusui.Cols(1).At("md", 2).At("lg", 4),
  Gap:  th.Space.SM,
}.Layout(th, gtx, items...)`,
				Demo:  "grid/2",
				DemoH: 200,
			},
			{
				Heading: "Readable page column",
				Prose: `<p><a href="../layout/"><code>LayoutPage(th, gtx, …)</code></a> caps content at
<code>th.PageMax</code> (default 920dp) or <code>th.PageMaxAt</code> when set. Inside that
column, children still see the column width.</p>`,
				Snippet: `lotusui.LayoutPage(th, gtx, func(gtx C) D {
	return lotusui.VStack(th.Space.MD,
		header,
		lotusui.SimpleGrid(th, gtx, cards, lotusui.SimpleGridProps{
			MinChildWidth: 180, MaxCols: 4, Gap: th.Space.SM,
		}, cardCell),
		lotusui.Wrap(th.Space.SM, layout.Middle, chips...),
	)(gtx)
})`,
			},
			{
				Heading: "Wrap — flowing chips and labels",
				Prose: `<p><a href="../wrap/"><code>Wrap</code></a> is CSS flex-wrap for Gio: each child
is measured at intrinsic width, then packed left-to-right; when the next child would exceed
Max.X, a new line starts. Prefer it for badges and filter chips.
<a href="../stack/"><code>HStack</code></a> never wraps.</p>`,
				Snippet: `lotusui.Wrap(th.Space.SM, layout.Middle,
	chip("Design"), chip("Engineering"), chip("Product"),
	chip("Marketing"), chip("Sales"), chip("Support"),
	// …)`,
				Demo:  "wrap/0",
				DemoH: 160,
			},
			{
				Heading: "SimpleGrid — continuous or stepped",
				Prose: `<p><a href="../simplegrid/"><code>SimpleGrid</code></a> continuous mode:
<code>columns = min(MaxCols, floor(Max.X / MinChildWidth))</code>. Stepped mode: set
<code>Columns: Cols(1).At("sm", 2).At("lg", 4)</code> and ignore minChildWidth.</p>`,
				Snippet: `lotusui.SimpleGrid(th, gtx, items, lotusui.SimpleGridProps{
	MinChildWidth: 140, MaxCols: 4, Gap: th.Space.SM,
}, cell)
// or stepped:
lotusui.SimpleGrid(th, gtx, items, lotusui.SimpleGridProps{
	Columns: lotusui.Cols(1).At("sm", 2).At("lg", 4), Gap: th.Space.SM,
}, cell)`,
				Demo:  "simplegrid/0",
				DemoH: 180,
			},
			{
				Heading: "Tabs and AnnotatedText",
				Prose: `<p>Horizontal <a href="../tabs/">Tabs</a> and
<a href="../annotated-text/">AnnotatedText</a> use Wrap so labels never squeeze to one
character under a narrow Max.X.</p>`,
				Snippet: `tabs := lotusui.Tabs{Options: lotusui.TabOpts(
	"Changes", "Staging", "Production", "Reviews", "Approvals", "History",
)}`,
				Demo:  "tabs/6",
				DemoH: 180,
			},
			{
				Heading: "Show / hide and Dialog width",
				Prose: `<p><code>Show(th, gtx, Bools(false).At("lg", true), w)</code> hides below a step.
Dialog / AlertDialog take <code>Sizes</code> / <code>Widths</code> for stepped card width —
not Button density Size.</p>`,
				Snippet: `lotusui.Show(th, gtx, lotusui.Bools(false).At("lg", true), sidebar)

d.Sizes = lotusui.Sizes(lotusui.SizeSM).At("md", lotusui.SizeLG).At("xl", lotusui.Size2XL)`,
				Demo:  "dialog/5",
				DemoH: 420,
			},
			{
				Heading: "Panes and scrolling",
				Prose: `<p><a href="../split/">Split</a> panes each get their own Max.X — put Wrap /
SimpleGrid / Tabs inside. Scroll helpers are vertical overflow, not a substitute for
horizontal reflow.</p>`,
				Snippet: `s.Layout(gtx, th.Space.MD, depth,
	lotusui.SplitBox(th, listPane),
	lotusui.SplitBox(th, func(gtx C) D {
		return lotusui.Wrap(th.Space.SM, layout.Middle, filters...)(gtx)
	}),
)`,
			},
			{
				Heading: "Choosing the right primitive",
				Prose: `<div class="proptable-wrap"><table class="proptable">
<thead><tr><th>Need</th><th>Use</th><th>Avoid</th></tr></thead>
<tbody>
<tr><td>Readable app column</td><td><code>LayoutPage(th, …)</code></td><td>Full-bleed Flexed edge-to-edge</td></tr>
<tr><td>Flowing chips / tags</td><td><code>Wrap</code></td><td><code>HStack</code> (squeezes)</td></tr>
<tr><td>Equal tiles, smooth N</td><td><code>SimpleGrid</code> minChildWidth</td><td>Hard-coded column Flex</td></tr>
<tr><td>Equal tiles, stepped N</td><td><code>SimpleGrid</code> <code>Columns</code> / <code>Grid.Cols</code></td><td>Hand <code>if Max.X</code> trees</td></tr>
<tr><td>Explicit spans</td><td><code>Grid</code></td><td>Nested HStacks</td></tr>
<tr><td>Hide below a step</td><td><code>Show</code> + <code>Bools</code></td><td>Clamping Max.X to 0</td></tr>
<tr><td>Custom step mins</td><td><code>ParseBreakpointsJSON</code> + <code>WithBreakpoints</code></td><td>Editing registry.json</td></tr>
</tbody></table></div>`,
			},
		},
	}
}

func iconsPage() *Page {
	return &Page{
		Slug:   "icons",
		Title:  "Icons",
		Kicker: "Fluent SVGs, embedded in the binary — builds never need the network.",
		Intro: `<p>Icons are SVGs fetched once by the CLI (normalized at fetch time) and committed. A
build-failing test asserts every icon rasterizes. There is exactly one icon path — never add a
second.</p>`,
		Sections: []Section{
			{
				Heading: "SVGIcon and SVGIconButton",
				Prose: `<p>Full-color icons ignore the tint; mono icons take it via
<code>currentColor</code>. Icon names are <em>generated constants</em>
(<code>lotusui.IconSettings</code>, and your own via the CLI's <code>-gen</code>), so a typo'd
or removed icon is a compile error — at runtime a missing icon renders nothing beyond its
reserved square, so screens degrade to their text.</p>`,
				Snippet: `lotusui.SVGIcon(lotusui.IconAdd, 32, color.NRGBA{})              // full-color
lotusui.SVGIcon(lotusui.IconEdit, 24, th.Palette.FgSubtle)       // mono, tinted
lotusui.SVGIconButton(th, &btn, lotusui.IconSettings, 24, active) // clickable`,
				Demo:  "icons",
				DemoH: 380,
			},
			{
				Heading: "Your own icons — the whole Iconify universe",
				Prose: `<p>Any icon on <a href="https://icon-sets.iconify.design">icon-sets.iconify.design</a>
(200,000+ across every major set) is one command away: paste its ref and the CLI appends it to
your manifest and fetches the SVG. You embed the result and register it — after that it renders
through the same cached pipeline as the built-ins, and your app carries exactly the icons it
uses. The network runs at <em>development</em> time only; builds and runtime never touch it.</p>`,
				Snippet: `# 1. Browse icon-sets.iconify.design, copy a ref, fetch it —
#    SVGs are normalized and typed constants generated in one step:
go run github.com/ikaito-com/lotusui/cmd/lotusui icons \
    -manifest icons/manifest.txt -out icons \
    -gen icons_gen.go -genpkg main tabler:rocket

# 2. Embed and register (once), and wire go:generate so the
#    workflow is one standard command from then on:
#    //go:embed icons/*.svg
#    var appIcons embed.FS
#    func init() { lotusui.RegisterIconFS(appIcons, "icons") }

# 3. Use them, compile-checked:
#    lotusui.SVGIcon(IconRocket, 24, th.Palette.FgSubtle)`,
				Lang: "sh",
			},
		},
	}
}

func seamlessPage() *Page {
	return &Page{
		Slug:      "seamless-window",
		Title:     "Seamless window",
		Kicker:    "Native edge-to-edge macOS chrome: no title bar, traffic lights kept, zero flash.",
		Platforms: []string{"macOS"},
		Intro: `<p>Content owns the whole window. The macOS title bar disappears as a bar, the three
traffic lights stay — comfortably inset, native, draggable strip and all — and the decorated
bar never appears, not even for one frame at launch.</p>`,
		Sections: []Section{
			{
				Heading: "The pairing",
				Prose: `<p>Two calls, both required. <code>app.Decorated(false)</code> makes Gio hide the
title bar natively — and, critically, stops Gio painting its own fallback decorations (on
Gio ≥ v0.8 it reads the window's full-size-content flag back and would otherwise draw a
client-side title bar over yours). <code>MakeSeamlessWindow</code> then adds what
<code>Decorated(false)</code> doesn't: it re-shows the traffic lights (Gio hides them), gives
them the comfortable unified-toolbar inset, and applies the titlebar transparency
<em>before the first frame is presented</em>, so there is no flash of decorated chrome.
Call it on every <code>AppKitViewEvent</code> — the symbol and the event both exist only on
macOS, so this lives in a <code>main_darwin.go</code>:</p>`,
				Snippet: `// main.go — create the window undecorated:
w := new(app.Window)
w.Option(app.Title("myapp"), app.Decorated(false))

// main_darwin.go — apply the seamless chrome:
case app.AppKitViewEvent:
	lotusui.MakeSeamlessWindow(e.View)`,
			},
			{
				Heading: "What it does under the hood",
				Prose: `<p>Synchronously at view-attach (before the first present): transparent title
bar, hidden title, full-size content — the no-flash guarantee. On the next runloop tick: an
empty unified toolbar (the sanctioned way to get inset traffic lights that stay put across
resizes), a one-point frame nudge to force AppKit's relayout, and a brief re-assertion loop
that wins the race against Gio's own configure pass, which would hide the buttons again.
Everything is idempotent — calling it on every <code>AppKitViewEvent</code> is the intended
usage.</p>
<p>One doctrine note for apps that use this: Gio version bumps must be verified with a real
launch on macOS — the pairing depends on how Gio's decoration handling reads the window's
style mask, which has changed between Gio versions before.</p>`,
			},
		},
		Props: []Prop{
			{"view", "uintptr", "The AppKit view from app.AppKitViewEvent — pass e.View; zero is ignored."},
		},
	}
}

func buttonPage() *Page {
	return &Page{
		Slug:   "button",
		Title:  "Button",
		Kicker: "Six variants × seven sizes × any color, options-driven.",
		Intro: `<p>One button, configured by <code>ButtonProps</code>. State lives in the caller's
<code>widget.Clickable</code>; every visual state below walks the theme's tokens and scale
steps, so a custom theme restyles all of it. Pressing nudges the button 1dp down — the same
tactile response on every variant; the pointer cursor signals interactivity. Anchor-style
rendering (as-link) and RTL are "from the web": a Link-variant button IS the link here, and RTL
layout is not yet supported by the toolkit.</p>`,
		Sections: []Section{
			InstallSection("button"),
			{
				Heading: "Usage",
				Snippet: `var save widget.Clickable
if save.Clicked(gtx) { /* … */ }
lotusui.Button(th, &save, "Button", lotusui.ButtonProps{})`,
				Demo:  "button/0",
				DemoH: 110,
			},
			{
				Heading: "Secondary",
				Prose:   `<p>The muted solid for auxiliary actions — quiet next to a Default.</p>`,
				Snippet: `lotusui.Button(th, &b, "Secondary", lotusui.ButtonProps{Variant: lotusui.ButtonSecondary})`,
				Demo:    "button/1",
				DemoH:   110,
			},
			{
				Heading: "Destructive",
				Prose:   `<p>The danger solid — Delete-class actions exclusively.</p>`,
				Snippet: `lotusui.Button(th, &b, "Destructive", lotusui.ButtonProps{Variant: lotusui.ButtonDestructive})`,
				Demo:    "button/2",
				DemoH:   110,
			},
			{
				Heading: "Outline",
				Prose:   `<p>Bordered with quiet ink; fills faintly on hover.</p>`,
				Snippet: `lotusui.Button(th, &b, "Outline", lotusui.ButtonProps{Variant: lotusui.ButtonOutline})`,
				Demo:    "button/3",
				DemoH:   110,
			},
			{
				Heading: "Ghost",
				Prose:   `<p>No chrome until hovered — toolbars, dense rows.</p>`,
				Snippet: `lotusui.Button(th, &b, "Ghost", lotusui.ButtonProps{Variant: lotusui.ButtonGhost})`,
				Demo:    "button/4",
				DemoH:   110,
			},
			{
				Heading: "Link",
				Prose:   `<p>Primary ink only; underlines on hover, like the link it imitates.</p>`,
				Snippet: `lotusui.Button(th, &b, "Link", lotusui.ButtonProps{Variant: lotusui.ButtonLink})`,
				Demo:    "button/5",
				DemoH:   110,
			},
			{
				Heading: "Icon",
				Prose: `<p>An empty label with <code>IconStart</code> renders an icon-only, square-padded
button.</p>`,
				Snippet: `lotusui.Button(th, &b, "", lotusui.ButtonProps{
	Variant: lotusui.ButtonOutline, IconStart: lotusui.IconSettings})`,
				Demo:  "button/6",
				DemoH: 110,
			},
			{
				Heading: "With icon",
				Prose: `<p><code>IconStart</code>/<code>IconEnd</code> render beside the label, tinted with
the button's own ink and sized to its text.</p>`,
				Snippet: `lotusui.Button(th, &add, "Add item", lotusui.ButtonProps{IconStart: lotusui.IconAdd})
lotusui.Button(th, &next, "Continue", lotusui.ButtonProps{
	Variant: lotusui.ButtonOutline, IconEnd: lotusui.IconExpand})`,
				Demo:  "button/7",
				DemoH: 110,
			},
			{
				Heading: "Rounded",
				Prose:   `<p><code>Rounded</code> renders the pill form — full-round corners on any variant.</p>`,
				Snippet: `lotusui.Button(th, &b, "Rounded", lotusui.ButtonProps{Rounded: true})`,
				Demo:    "button/8",
				DemoH:   110,
			},
			{
				Heading: "Loading",
				Prose: `<p><code>Loading</code> disables input and shows a spinner over the label (width
preserved); <code>LoadingText</code> swaps the label instead, spinner at the start.</p>`,
				Snippet: `lotusui.Button(th, &save, "Save", lotusui.ButtonProps{Loading: saving})
lotusui.Button(th, &save, "Save", lotusui.ButtonProps{
	Loading: saving, LoadingText: "Saving…"})`,
				Demo:  "button/9",
				DemoH: 110,
			},
			{
				Heading: "Sizes",
				Prose: `<p>Seven presets on the shared <code>Size</code> scale — <code>2XS</code> through
<code>2XL</code>, text and padding moving together.</p>`,
				Snippet: `lotusui.Button(th, &btn, "2XS", lotusui.ButtonProps{Size: lotusui.Size2XS})
lotusui.Button(th, &btn, "2XL", lotusui.ButtonProps{Size: lotusui.Size2XL})`,
				Demo:  "button/10",
				DemoH: 130,
			},
			{
				Heading: "Color",
				Prose: `<p><code>Color</code> re-colors any variant from a <code>ColorScale</code> — one
anchor in, the whole interaction ladder out (.500 base, .600 hover, .700 pressed): a custom
color can never fall out of step with its hover and pressed states.</p>`,
				Snippet: `lotusui.Button(th, &go1, "Teal", lotusui.ButtonProps{Color: lotusui.Teal})
lotusui.Button(th, &go2, "Brand", lotusui.ButtonProps{Color: lotusui.ScaleFrom(brand)})`,
				Demo:  "button/11",
				DemoH: 110,
			},
			{
				Heading: "Scheme",
				Prose: `<p><code>Scheme</code> wins over <code>Color</code> and the variant's role scheme:
full manual control of every slot.</p>`,
				Snippet: `soft := lotusui.Teal.SoftScheme()
lotusui.Button(th, &go1, "Go", lotusui.ButtonProps{Scheme: &soft})`,
				Demo:  "button/12",
				DemoH: 110,
			},
			{
				Heading: "Disabled",
				Prose: `<p>Disabled walks a <em>brighter</em> step of the button's own color scale — same hue,
quieter chrome (Soft fills lift to <code>Subtle</code>; <code>Color</code> solids mute to
<code>.200</code> fill / <code>.600</code> ink; outline/ghost/link ink softens) — and drops the
pointer cursor, hover fill, and press nudge.</p>`,
				Snippet: `lotusui.Button(th, &btn, "Disabled", lotusui.ButtonProps{Disabled: true})`,
				Demo:    "button/13",
				DemoH:   110,
			},
			{
				Heading: "Cursor",
				Prose: `<p>Buttons render the pointer cursor whenever they are enabled — interactivity is
signalled by the cursor, the hover fill, and the press nudge together. Disabled buttons drop
all three.</p>`,
			},
			{
				Heading: "Group",
				Prose: `<p>There is no group component on purpose: <code>HStack</code> IS the group —
buttons are widgets, so grouping is ordinary composition. (An attached, shared-border group is
on the roadmap.)</p>`,
				Snippet: `lotusui.HStack(th.Space.SM,
	lotusui.Button(th, &cancel, "Cancel", lotusui.ButtonProps{Variant: lotusui.ButtonOutline}),
	lotusui.Button(th, &save, "Save", lotusui.ButtonProps{}),
)`,
				Demo:  "button/14",
				DemoH: 110,
			},
			{
				Heading: "Keyboard focus",
				Prose: `<p>A focused button renders a 2dp ring in the theme's <code>FocusRing</code> token —
focus is never signalled by color alone. <code>FullWidth</code> makes any button span the
column; <code>RightAligned</code> anchors a dialog's button row to the right edge.</p>`,
			},
		},
		Props: []Prop{
			{"Variant", "ButtonVariant", "ButtonDefault (default), ButtonSecondary, ButtonDestructive, ButtonOutline, ButtonGhost, ButtonLink."},
			{"Size", "Size", "Size2XS … Size2XL, SizeMD default — the shared size presets."},
			{"Color", "ColorScale", "Re-colors the variant from a scale; the interaction ladder derives from it."},
			{"Scheme", "*Scheme", "Wins over Color: full manual slot control — e.g. Teal.SoftScheme()."},
			{"IconStart / IconEnd", "string", "Embedded icon names beside the label; empty label + IconStart = icon-only."},
			{"Rounded", "bool", "The pill form — full-round corners."},
			{"Loading", "bool", "Spinner over the label (width kept), input off."},
			{"LoadingText", "string", "Replaces the label while Loading, spinner at the start."},
			{"Disabled", "bool", "Brighter same-hue step of the button's scale; swallows clicks and drops the pointer cursor."},
		},
	}
}

func fieldPage() *Page {
	return &Page{
		Slug:   "field",
		Title:  "Field",
		Kicker: "Label, helper, error and required — around any control.",
		Intro: `<p><code>Field</code> is the form-field wrapper: it adds the label above, helper or
error text below, and the required marker — around <em>any</em> control (Input, Select,
Checkbox, a custom widget). The control knows nothing about it; composition, not slots.</p>`,
		Sections: []Section{
			InstallSection("field"),
			{
				Heading: "Usage",
				Snippet: `lotusui.Field(th, lotusui.FieldProps{Label: "Email"}, func(gtx C) D {
	return email.LayoutField(th, gtx, "you@example.com")
})`,
				Demo:  "field/0",
				DemoH: 130,
			},
			{
				Heading: "Helper text",
				Prose:   `<p><code>Helper</code> renders muted guidance under the control.</p>`,
				Snippet: `lotusui.FieldProps{Label: "Email", Helper: "We'll never share it."}`,
				Demo:    "field/1",
				DemoH:   150,
			},
			{
				Heading: "Required",
				Prose:   `<p><code>Required</code> marks the label; validation stays yours.</p>`,
				Snippet: `lotusui.FieldProps{Label: "Email", Required: true}`,
				Demo:    "field/2",
				DemoH:   130,
			},
			{
				Heading: "Error",
				Prose: `<p><code>Error</code> replaces the helper with the message in danger ink — pair it
with the control's own invalid chrome.</p>`,
				Snippet: `lotusui.FieldProps{Label: "Workspace name", Error: msg}`,
				Demo:    "field/3",
				DemoH:   150,
			},
			{
				Heading: "Any control",
				Prose:   `<p>The wrapped control is any <code>layout.Widget</code> — here a Select.</p>`,
				Snippet: `lotusui.Field(th, lotusui.FieldProps{Label: "Role"}, func(gtx C) D {
	return role.Layout(th, gtx, "")
})`,
				Demo:  "field/4",
				DemoH: 140,
			},
		},
		Props: []Prop{
			{"Label", "string", "The label above the control."},
			{"Helper", "string", "Muted guidance below; recomputed each frame (counters)."},
			{"Error", "string", "Replaces Helper in danger ink."},
			{"Required", "bool", "Marks the label."},
		},
	}
}

func inputPage() *Page {
	return &Page{
		Slug:   "input",
		Title:  "Input",
		Kicker: "The single-line input: three variants, seven sizes, composable everything else.",
		Intro: `<p>A single-line editor in themed chrome. Structure (labels, buttons, adornments) is
COMPOSITION — nil-able <code>Start</code>/<code>End</code> slots and the <code>Field</code>
wrapper; behavior (filters, transforms) is caller-owned functions. A file input is a native
picker on the web — out of scope for a canvas UI, use your platform's file dialog ("from the
web").</p>`,
		Sections: []Section{
			InstallSection("input"),
			{
				Heading: "Usage",
				Snippet: `var name lotusui.Input
name.LayoutField(th, gtx, "Email")`,
				Demo:  "input/0",
				DemoH: 110,
			},
			{
				Heading: "Disabled",
				Prose:   `<p>Read-only, dimmed, no caret.</p>`,
				Snippet: `name.Disabled = true`,
				Demo:    "input/1",
				DemoH:   110,
			},
			{
				Heading: "With label",
				Prose:   `<p>Wrap any control in <code>Field</code> for the label.</p>`,
				Snippet: `lotusui.Field(th, lotusui.FieldProps{Label: "Email"}, func(gtx C) D {
	return email.LayoutField(th, gtx, "you@example.com")
})`,
				Demo:  "input/2",
				DemoH: 130,
			},
			{
				Heading: "With button",
				Prose:   `<p>An input beside a button is ordinary composition — a Flex row.</p>`,
				Snippet: `layout.Flex{Alignment: layout.Middle}.Layout(gtx,
	layout.Flexed(1, func(gtx C) D { return email.LayoutField(th, gtx, "you@example.com") }),
	layout.Rigid(lotusui.HSpacer(th.Space.SM)),
	layout.Rigid(lotusui.Button(th, &sub, "Subscribe", lotusui.ButtonProps{Variant: lotusui.ButtonOutline})),
)`,
				Demo:  "input/3",
				DemoH: 110,
			},
			{
				Heading: "Variants",
				Prose: `<p><code>InputOutline</code> (default), <code>InputSubtle</code> (filled, border on
focus), <code>InputFlushed</code> (underline only).</p>`,
				Snippet: `in.Variant = lotusui.InputSubtle
in.Variant = lotusui.InputFlushed`,
				Demo:  "input/4",
				DemoH: 200,
			},
			{
				Heading: "Sizes",
				Prose:   `<p>All seven shared <code>Size</code> presets — text and padding move together.</p>`,
				Snippet: `in.Size = lotusui.Size2XS // … through lotusui.Size2XL`,
				Demo:    "input/5",
				DemoH:   420,
			},
			{
				Heading: "Start and End elements",
				Prose: `<p>Widgets INSIDE the field, beside the editor — an icon, a clear ✕, a visibility
eye. nil costs nothing.</p>`,
				Snippet: `search.Start = lotusui.SVGIcon(lotusui.IconEye, 16, th.Palette.FgSubtle)
secret.End = lotusui.SVGIconButtonTint(th, &eyeBtn, lotusui.IconEyeOff, 16, false, th.Palette.FgSubtle)

// Masking is the editor's own: the eye toggles it.
secret.Editor.Mask = '•'
if reveal { secret.Editor.Mask = 0 }`,
				Demo:  "input/6",
				DemoH: 160,
			},
			{
				Heading: "Suffix addon",
				Prose:   `<p>A fixed suffix rendered as part of the field's chrome.</p>`,
				Snippet: `sub.LayoutSuffix(th, gtx, "Subdomain", "yourname", ".example.com")`,
				Demo:    "input/7",
				DemoH:   130,
			},
			{
				Heading: "Invalid",
				Prose:   `<p><code>Error</code> switches the chrome to danger and renders the message.</p>`,
				Snippet: `in.Error = "this name is already taken"`,
				Demo:    "input/8",
				DemoH:   150,
			},
			{
				Heading: "Field — label, helper, required",
				Prose: `<p>The full wrapper is its own component — see <a href="../field/">Field</a> for
helper text, required markers and error slots around any control.</p>`,
				Snippet: `lotusui.Field(th, lotusui.FieldProps{
	Label: "Email", Helper: "We'll never share it.", Required: true,
}, control)`,
				Demo:  "input/9",
				DemoH: 150,
			},
			{
				Heading: "Filter and Transform",
				Prose: `<p>Generic mechanisms for the app's rules: <code>Filter</code> is an allow-list,
<code>Transform</code> a same-frame rewrite (here: fold to lowercase).</p>`,
				Snippet: `in.Filter = "abcdefghijklmnopqrstuvwxyz0123456789-"
in.Transform = strings.ToLower`,
				Demo:  "input/10",
				DemoH: 150,
			},
			{
				Heading: "Clear button — composed",
				Prose:   `<p>A clear button is just an <code>End</code> widget; the core knows nothing about it.</p>`,
				Snippet: `if clearBtn.Clicked(gtx) { in.Editor.SetText("") }
in.End = lotusui.SVGIconButtonTint(th, &clearBtn, lotusui.IconRemove, 14, false, th.Palette.FgSubtle)`,
				Demo:  "input/11",
				DemoH: 110,
			},
			{
				Heading: "Character counter — composed",
				Prose:   `<p>Cap length with <code>Transform</code>; the counter is <code>Field.Helper</code>, recomputed each frame.</p>`,
				Snippet: `n := len(bio.Editor.Text())
lotusui.Field(th, lotusui.FieldProps{Label: "Bio", Helper: fmt.Sprintf("%d / 80", n)}, control)`,
				Demo:  "input/12",
				DemoH: 150,
			},
			{
				Heading: "Grid",
				Prose:   `<p>Inputs side by side are a Flex row — each field takes an equal share.</p>`,
				Snippet: `layout.Flex{}.Layout(gtx,
	layout.Flexed(1, firstNameField),
	layout.Rigid(lotusui.HSpacer(th.Space.MD)),
	layout.Flexed(1, lastNameField),
)`,
				Demo:  "input/13",
				DemoH: 150,
			},
			{
				Heading: "With badge",
				Prose:   `<p>The label row is yours to compose — here a Beta badge beside the label.</p>`,
				Snippet: `lotusui.HStack(th.Space.SM,
	lotusui.SectionLabel(th, "API token"),
	lotusui.Badge(th, "Beta", lotusui.BadgeProps{Variant: lotusui.BadgeSecondary, Size: lotusui.SizeXS}),
)`,
				Demo:  "input/14",
				DemoH: 140,
			},
			{
				Heading: "Field group",
				Prose:   `<p>Stacked Fields in a Card read as one form section.</p>`,
				Snippet: `lotusui.Card(th, lotusui.CardProps{}, lotusui.VStack(th.Space.MD,
	cityField, postalField,
))`,
				Demo:  "input/15",
				DemoH: 260,
			},
			{
				Heading: "Text addons",
				Prose:   `<p>Muted text segments in the frame — a currency pair, a URL scheme and TLD — are <code>Start</code>/<code>End</code> widgets.</p>`,
				Snippet: `amount.Start = addonText(th, "$")
amount.End = addonText(th, "USD")`,
				Demo:  "input/16",
				DemoH: 140,
			},
			{
				Heading: "Kbd",
				Prose:   `<p><code>Kbd</code> renders the keyboard-cap hint — compose it as the End slot.</p>`,
				Snippet: `search.Start = lotusui.SVGIcon(lotusui.IconSearch, 16, th.Palette.FgSubtle)
search.End = lotusui.Kbd(th, "⌘K")`,
				Demo:  "input/17",
				DemoH: 100,
			},
			{
				Heading: "Spinner",
				Snippet: `field.End = lotusui.Spinner(th, 14)`,
				Demo:    "input/18",
				DemoH:   100,
			},
			{
				Heading: "Inline button",
				Prose:   `<p>A small icon button inside the frame — the copy-URL field.</p>`,
				Snippet: `url.End = lotusui.SVGIconButtonTint(th, &copyBtn, lotusui.IconFile, 14, false, th.Palette.FgSubtle)`,
				Demo:    "input/19",
				DemoH:   100,
			},
			{
				Heading: "Block addons",
				Prose:   `<p><code>Top</code>/<code>Bottom</code> render full-width rows INSIDE the frame, above/below the editor line.</p>`,
				Snippet: `name.Top = headerRow // a caption, a status row with actions`,
				Demo:    "input/20",
				DemoH:   120,
			},
		},
		Props: []Prop{
			{"Variant", "InputVariant", "InputOutline (default), InputSubtle, InputFlushed."},
			{"Size", "Size", "The shared size presets (Size2XS–Size2XL)."},
			{"Filter", "string", "Allow-list of runes; empty accepts everything."},
			{"Transform", "func(string) string", "Same-frame rewrite — folds, caps, masks."},
			{"Error", "string", "Danger chrome + message below the field."},
			{"Disabled", "bool", "Read-only, dimmed, no caret."},
			{"Start / End", "layout.Widget", "Widgets inside the field, beside the editor; nil costs nothing."},
			{"Editor", "widget.Editor", "The underlying Gio editor — its own Mask (password dots), InputHint (mobile keyboard), MaxLen and selection API are yours to set directly."},
			{"Top / Bottom", "layout.Widget", "Full-width rows inside the frame — the block addons."},
			{"Kbd(th, key)", "widget", "The keyboard-cap hint, for End slots and prose."},
		},
	}
}

func checkboxPage() *Page {
	return &Page{
		Slug:   "checkbox",
		Title:  "Checkbox",
		Kicker: "A themed box, a check, a label — plus indeterminate, invalid and sizes.",
		Intro: `<p>State lives in the caller's struct (<code>Value</code>); the component renders and
reports clicks. One struct per on-screen instance — identity in immediate mode is the
struct.</p>`,
		Sections: []Section{
			InstallSection("checkbox"),
			{
				Heading: "Usage",
				Snippet: `var accept lotusui.Checkbox
if accept.Clicked(gtx) { accept.Value = !accept.Value }
accept.Layout(th, gtx, "Accept terms and conditions")`,
				Demo:  "checkbox/0",
				DemoH: 100,
			},
			{
				Heading: "Checked state",
				Prose: `<p>The state IS the struct's <code>Value</code> — set it for an initial checked
box, flip it on <code>Clicked</code>. There is no controlled/uncontrolled split in immediate
mode; every checkbox is "controlled" by construction.</p>`,
				Snippet: `accept := lotusui.Checkbox{Value: true} // starts checked`,
			},
			{
				Heading: "With text",
				Prose:   `<p>A caption under the label is composition — a stack and an inset.</p>`,
				Snippet: `lotusui.VStack(th.Space.XS,
	func(gtx C) D { return terms.Layout(th, gtx, "Accept terms and conditions") },
	func(gtx C) D {
		return layout.Inset{Left: 26}.Layout(gtx,
			lotusui.LabelCaption(th, "By clicking this checkbox, you agree to the terms and conditions.").Layout)
	},
)`,
				Demo:  "checkbox/1",
				DemoH: 120,
			},
			{
				Heading: "Group",
				Prose:   `<p>A checkbox list is composition — one struct per row, values yours.</p>`,
				Snippet: `for _, c := range []*lotusui.Checkbox{&recents, &home, &apps} {
	if c.Clicked(gtx) { c.Value = !c.Value }
}`,
				Demo:  "checkbox/2",
				DemoH: 140,
			},
			{
				Heading: "Disabled",
				Snippet: `included.Value, included.Disabled = true, true`,
				Demo:    "checkbox/3",
				DemoH:   100,
			},
			{
				Heading: "Indeterminate and invalid",
				Prose: `<p><code>Indeterminate</code> renders the dash (a parent with some children
selected); <code>Invalid</code> switches the chrome to danger.</p>`,
				Snippet: `parent.Indeterminate = true
terms.Invalid = true`,
				Demo:  "checkbox/4",
				DemoH: 130,
			},
			{
				Heading: "In a table",
				Prose:   `<p>Row selection is composition: checkbox cells in a <a href="../table/">Table</a>.</p>`,
				Snippet: `lotusui.Table(th, lotusui.TableProps{Widths: []float32{0.5, 2, 1}},
	[]string{"", "Repository", "Visibility"}, rowsWithCheckboxCells)`,
				Demo:  "checkbox/5",
				DemoH: 240,
			},
			{
				Heading: "Sizes",
				Prose:   `<p>All seven shared <code>Size</code> presets.</p>`,
				Snippet: `box.Size = lotusui.Size2XS // … through lotusui.Size2XL`,
				Demo:    "checkbox/6",
				DemoH:   110,
			},
		},
		Props: []Prop{
			{"Value", "bool", "The state — yours."},
			{"Size", "Size", "The shared size presets (Size2XS–Size2XL)."},
			{"Indeterminate", "bool", "The parent-of-mixed-children dash."},
			{"Invalid", "bool", "Danger chrome."},
			{"Disabled", "bool", "Dimmed, unclickable."},
		},
	}
}

func codeBlockPage() *Page {
	return &Page{
		Slug:   "code-block",
		Title:  "Code Block",
		Kicker: "Fenced source chrome: caption, optional Copy, pre-formatted body.",
		Intro: `<p>A lotusui extension for showing source in apps and docs. The library
owns the chrome and paints token runs; it does <strong>not</strong> embed a highlighter —
pass <code>Lines [][]CodeSpan</code> from your tokenizer (chroma, tree-sitter, build-time)
so the module zip stays lean. <code>Plain</code> is the unstyled fallback.</p>`,
		Sections: []Section{
			InstallSection("code-block"),
			{
				Heading: "Usage",
				Prose: `<p>Caption language + body. Prefer highlighted <code>Lines</code>;
<code>Plain</code> alone is fine for short snippets.</p>`,
				Snippet: `lotusui.CodeBlock(th, lotusui.CodeBlockProps{
	Lang:  "go",
	Plain: "fmt.Println(\"hello\")",
})`,
				Demo:  "code-block/0",
				DemoH: 160,
			},
			{
				Heading: "Highlighted spans",
				Prose: `<p>Each line is a slice of <code>CodeSpan</code> — text, color, optional bold.
The docs site tokenizes with chroma and maps token kinds onto palette roles
(BrandFg keywords, Success strings, …).</p>`,
				Snippet: `lotusui.CodeBlock(th, lotusui.CodeBlockProps{
	Lang: "go",
	Lines: [][]lotusui.CodeSpan{
		{{Text: "package ", Color: muted}, {Text: "main", Color: fg, Bold: true}},
	},
})`,
				Demo:  "code-block/1",
				DemoH: 200,
			},
			{
				Heading: "Copy",
				Prose: `<p>Pass a durable <code>*widget.Clickable</code> on <code>Copy</code> — the
button writes <code>Plain</code> (or joined Lines text) to the clipboard.</p>`,
				Snippet: `var copy widget.Clickable
lotusui.CodeBlock(th, lotusui.CodeBlockProps{
	Lang: "go", Plain: src, Copy: &copy,
})`,
				Demo:  "code-block/2",
				DemoH: 180,
			},
			{
				Heading: "Nested",
				Prose: `<p><code>Nested: true</code> omits the outer border/radius — use inside
<a href="../example/">Example</a> Code chrome (or any parent card).</p>`,
				Snippet: `lotusui.CodeBlock(th, lotusui.CodeBlockProps{Nested: true, Plain: src})`,
				Demo:    "code-block/3",
				DemoH:   140,
			},
		},
		Props: []Prop{
			{"Lang", "string", "Caption label (empty → \"Go\")."},
			{"Lines", "[][]CodeSpan", "Highlighted rows of token runs (preferred)."},
			{"Plain", "string", "Unstyled body / clipboard source when Lines empty."},
			{"Copy", "*widget.Clickable", "Optional Copy button (clipboard write)."},
			{"Nested", "bool", "Omit outer frame — for parent card chrome."},
			{"CodeSpan", "struct", "Text + Color + Bold — one highlighted run."},
		},
	}
}

func examplePage() *Page {
	return &Page{
		Slug:   "example",
		Title:  "Example",
		Kicker: "Preview|Code chrome: one card, tab strip on top, live widget inside.",
		Intro: `<p>A lotusui extension for recipe surfaces (docs, design systems, in-app
guides). Not <code>Tabs</code> — the strip is the top of a bordered card, and the body is
either a live Preview or a Code panel. Pair Code with
<code>CodeBlock{Nested: true}</code> so the fence does not double-frame.</p>`,
		Sections: []Section{
			InstallSection("example"),
			{
				Heading: "Usage",
				Prose: `<p>Durable state on <code>*Example</code>. Pass Preview always; omit Code for
Preview-only.</p>`,
				Snippet: `var ex lotusui.Example
ex.Layout(th, gtx, lotusui.ExampleProps{
	Preview: liveWidget,
	Code:    lotusui.CodeBlock(th, lotusui.CodeBlockProps{Nested: true, Plain: src}),
})`,
				Demo:  "example/0",
				DemoH: 220,
			},
			{
				Heading: "Preview only",
				Prose:   `<p>When <code>Code</code> is nil, the Code tab is hidden and Preview stays active.</p>`,
				Snippet: `ex.Layout(th, gtx, lotusui.ExampleProps{Preview: liveWidget})`,
				Demo:    "example/1",
				DemoH:   160,
			},
			{
				Heading: "With CodeBlock",
				Prose: `<p>The docs site uses this for every capability section: Preview renders the
live demo; Code shows the snippet (chroma → <code>CodeSpan</code> in the site module).</p>`,
				Snippet: `ex.Layout(th, gtx, lotusui.ExampleProps{
	Preview: preview,
	Code: lotusui.CodeBlock(th, lotusui.CodeBlockProps{
		Nested: true, Lang: "go", Plain: src, Lines: spans,
	}),
})`,
				Demo:  "example/2",
				DemoH: 260,
			},
		},
		Props: []Prop{
			{"Example", "*Example", "Durable Preview|Code tab state; must outlive the frame."},
			{"ExampleProps.Preview", "layout.Widget", "Live body (required)."},
			{"ExampleProps.Code", "layout.Widget", "Code panel; nil → Preview only."},
		},
	}
}

func switchPage() *Page {
	return &Page{
		Slug:   "switch",
		Title:  "Switch",
		Kicker: "The toggle: sliding thumb, accent when on, animated on the shared clock.",
		Intro: `<p>A rounded track with a sliding thumb. Clicking flips <code>Value</code>; the thumb
animates on the shared animation clock so every toggle in an app moves identically.</p>`,
		Sections: []Section{
			InstallSection("switch"),
			{
				Heading: "Usage",
				Snippet: `var notify lotusui.Switch
notify.Layout(th, gtx)
if notify.Value { /* … */ }`,
				Demo:  "switch/0",
				DemoH: 130,
			},
			{
				Heading: "Sizes",
				Prose:   `<p>All seven shared <code>Size</code> presets.</p>`,
				Snippet: `sw.Size = lotusui.Size2XS // … through lotusui.Size2XL`,
				Demo:    "switch/1",
				DemoH:   110,
			},
			{
				Heading: "Disabled",
				Prose:   `<p>On and off, both facts — dimmed, unclickable.</p>`,
				Snippet: `included.Value, included.Disabled = true, true`,
				Demo:    "switch/2",
				DemoH:   110,
			},
			{
				Heading: "Invalid",
				Snippet: `mustEnable.Invalid = true`,
				Demo:    "switch/3",
				DemoH:   110,
			},
			{
				Heading: "Choice card",
				Prose: `<p>Settings rows — title, caption, switch at the end — inside a Card: pure
composition.</p>`,
				Snippet: `lotusui.Card(th, lotusui.CardProps{}, lotusui.VStack(th.Space.MD,
	settingRow(&marketing, "Marketing emails", "Receive emails about new products and more."),
	lotusui.Hairline(th),
	settingRow(&security, "Security emails", "Receive emails about your account security."),
))`,
				Demo:  "switch/4",
				DemoH: 200,
			},
		},
		Props: []Prop{
			{"Value", "bool", "The state; clicking flips it."},
			{"Size", "Size", "The shared size presets (Size2XS–Size2XL)."},
			{"Invalid", "bool", "Danger chrome."},
			{"Disabled", "bool", "Dimmed, unclickable."},
		},
	}
}

func selectPage() *Page {
	return &Page{
		Slug:   "select",
		Title:  "Select",
		Kicker: "Trigger + chevron, a floating panel of options, check-marked selection.",
		Intro: `<p>A bordered trigger showing the current choice; clicking it opens a <em>floating</em>
panel over the content beneath — built on the shared portal primitive (<code>Floating</code>),
so it escapes any parent clipping, paints above everything, and wins the pointer. Picking an
option, pressing anywhere else, or Escape closes it; the panel opens aligned to the current
selection. Web-specific composition (<code>SelectTrigger</code>/<code>SelectValue</code>
sub-components) and RTL are "from the web": Go structs replace composition, and RTL layout is
not yet supported by the underlying toolkit.</p>`,
		Sections: []Section{
			InstallSection("select"),
			{
				Heading: "Usage",
				Prose: `<p>Options carry a <code>Label</code> (what the user reads) and a
<code>Value</code> (what your app stores) — <code>SelectItem</code> / <code>SelectOpts</code> build lists. Call
<code>Clear()</code> for the placeholder state; read the choice with <code>Value()</code>,
never an index. Open the demo's panel: it floats over whatever is beneath.</p>`,
				Snippet: `fruit := lotusui.Select{
	Options:     lotusui.SelectOpts("Apple", "Banana", "Blueberry", "Grapes", "Pineapple"),
	Placeholder: "Select a fruit",
}
fruit.Clear() // start on the placeholder

fruit.Layout(th, gtx, "Fruit")
chosen := fruit.Value() // "Banana"`,
				Demo:  "select/0",
				DemoH: 340,
			},
			{
				Heading: "Align item with trigger",
				Prose: `<p><code>AlignItemWithTrigger</code> overlays the open panel so the SELECTED row
sits directly over the trigger — the native-select feel — instead of dropping below. Long
lists still scroll; the alignment accounts for scrolled-away rows.</p>`,
				Snippet: `font := lotusui.Select{Options: fonts, AlignItemWithTrigger: true}
font.SetValue("Roboto")`,
				Demo:  "select/1",
				DemoH: 380,
			},
			{
				Heading: "Groups",
				Prose: `<p><code>Groups</code> renders options under muted labels with separators between
groups. <code>Value()</code> reads the choice whichever group it came from — grouping is
presentation, not identity.</p>`,
				Snippet: `produce := lotusui.Select{
	Groups: lotusui.SelectGroups(
		lotusui.SelectGrouped("Fruits", lotusui.SelectOpts("Apple", "Banana", "Cherry")...),
		lotusui.SelectGrouped("Vegetables", lotusui.SelectOpts("Carrot", "Leek", "Spinach")...),
	),
	Placeholder: "Pick a produce…",
}
produce.Clear()`,
				Demo:  "select/2",
				DemoH: 460,
			},
			{
				Heading: "Scrollable",
				Prose: `<p>The panel caps at seven default rows; longer lists scroll inside it, and the panel
opens scrolled to the current selection — the align-with-trigger behavior, transposed.
Variable-height <code>Content</code> rows share the same height budget.</p>`,
				Snippet: `tz := lotusui.Select{Options: timezones}
tz.SetValue("Europe/Paris") // the panel opens scrolled to it`,
				Demo:  "select/3",
				DemoH: 440,
			},
			{
				Heading: "Meta",
				Prose: `<p><code>SelectOption.Meta</code> is optional secondary text on the far right of
an option row (and on the closed trigger) — a count, shortcut, … Empty omits it. The selected
check still sits after Meta.</p>`,
				Snippet: `lotusui.Select{Options: []lotusui.SelectOption{
	{Label: "roteland", Value: "r", Meta: "1"},
	{Label: "test", Value: "t", Meta: "2"},
}}`,
				Demo:  "select/4",
				DemoH: 280,
			},
			{
				Heading: "Icons",
				Prose:   `<p><code>SelectOption.Icon</code> paints a leading icon on the row and closed trigger (when <code>Content</code> is nil).</p>`,
				Snippet: `lotusui.SelectItems(
	{Label: "Line", Value: "line", Icon: lotusui.IconEdit},
	{Label: "Bar", Value: "bar", Icon: lotusui.IconSettings},
)`,
				Demo:  "select/5",
				DemoH: 340,
			},
			{
				Heading: "Subscription plan",
				Prose: `<p>Build-time <code>Content</code> replaces Icon+Label with arbitrary widgets — the shadcn
plan card (title + description). The same Content paints in the closed trigger. Keep
<code>Label</code>/<code>Value</code> for the choice contract; build Content when the options
list is built, not every frame in hot paths.</p>`,
				Snippet: `planRow := func(name, desc string) layout.Widget {
	return lotusui.VStack(2,
		lotusui.LabelBody(th, name).Layout,
		lotusui.LabelCaption(th, desc).Layout,
	)
}
sel.Options = lotusui.SelectItems(
	{Label: "Starter", Value: "starter", Content: planRow("Starter", "Perfect for individuals getting started.")},
	{Label: "Professional", Value: "pro", Content: planRow("Professional", "Ideal for growing teams and businesses.")},
)`,
				Demo:  "select/6",
				DemoH: 420,
			},
			{
				Heading: "Disabled",
				Prose:   `<p>Freezes the control on its current choice — identity fields that cannot change after creation.</p>`,
				Snippet: `engine.Disabled = true`,
				Demo:    "select/7",
				DemoH:   130,
			},
			{
				Heading: "Invalid",
				Prose:   `<p>Danger chrome on the trigger — pair it with a Field error message.</p>`,
				Snippet: `size.Invalid = true`,
				Demo:    "select/8",
				DemoH:   320,
			},
			{
				Heading: "Sizes",
				Prose:   `<p>All seven shared <code>Size</code> presets on the trigger frame.</p>`,
				Snippet: `sel.Size = lotusui.Size2XS // … through lotusui.Size2XL`,
				Demo:    "select/9",
				DemoH:   480,
			},
		},
		Props: []Prop{
			{"Options", "[]SelectOption", "The choices: Label is what the user reads, Value what your app stores (HTML's &lt;option value&gt;). Empty Value = the Label is the value; SelectItem / SelectOpts build lists."},
			{"SelectOption.Icon", "string", "Leading icon on the option row and closed trigger when Content is nil."},
			{"SelectOption.Content", "layout.Widget", "Build-time rich row body (multiline plan cards, …). Replaces Icon+Label in the panel and trigger; Label/Value still own identity."},
			{"SelectOption.Meta", "string", "Optional secondary text on the far right of the option row and closed trigger. Empty omits it; the selected check still sits after Meta. Hidden when Content is set."},
			{"SelectItem / SelectItemValue / SelectItems", "ctors", "Build-time composition: one option, valued option, pack into a slice."},
			{"SelectGrouped / SelectGroups", "ctors", "Build-time groups (the type remains SelectGroup — Go cannot also export func SelectGroup)."},
			{"Groups", "[]SelectGroup", "Wins over Options: options under muted labels, separators between groups, flattened in order."},
			{"Size", "Size", "The shared size presets for the trigger frame."},
			{"Value() / SetValue(v)", "string", "Read and write the CHOICE — never an index, so reordering or rewording the options cannot change what stored data means. An unknown SetValue clears the choice."},
			{"Clear() / Chosen()", "—", "Back to the placeholder state; whether anything is chosen. The zero value selects the first option, like a &lt;select&gt; with no selected attribute."},
			{"Placeholder", "string", "Shown in muted ink while nothing is chosen."},
			{"AlignItemWithTrigger", "bool", "Open the panel with the selected row over the trigger (uniform string rows; Content rows drop below)."},
			{"Invalid", "bool", "Danger chrome on the trigger."},
			{"Disabled", "bool", "Freezes the control: no pointer cursor, no opening, dimmed value."},
		},
	}
}

func tabsPage() *Page {
	return &Page{
		Slug:   "tabs",
		Title:  "Tabs",
		Kicker: "Three variants and an explicit Update contract.",
		Intro: `<p>Tab labels in a row; the active one styled by variant. <code>Update</code> must run
before anything reads the selection in the same frame — Layout deliberately does NOT process
clicks, so a consumer that forgets Update gets dead tabs (impossible to miss) instead of subtle
one-frame lag (which survives review).</p>
<p class="note">Like every choice component, the selection is a VALUE: the cursor is
unexported, <code>Value()</code>/<code>SetValue()</code> read and write it, the zero value picks
the first option, and an unknown value clears rather than falling back to option 0 — see
<a href="../select/">Select</a> for the full contract.</p>`,
		Sections: []Section{
			InstallSection("tabs"),
			{
				Heading: "Usage",
				Prose: `<p>The default look: the strip sits in a muted rounded well and the active tab is
a raised panel inside it.</p>`,
				Snippet: `tabs := lotusui.Tabs{Options: lotusui.TabOpts("Account", "Password")}

tabs.Update(gtx) // BEFORE anything reads the selection
body := accountBody
if tabs.Value() == "Password" {
	body = passwordBody
}
tabs.Layout(th, gtx)`,
				Demo:  "tabs/0",
				DemoH: 460,
			},
			{
				Heading: "Line variant",
				Prose: `<p><code>TabsLine</code> is the classic underline strip: quiet labels, a 2dp
brand-ink bar under the active tab, hover pre-shadowed.</p>`,
				Snippet: `tabs := lotusui.Tabs{Variant: lotusui.TabsLine}`,
				Demo:    "tabs/1",
				DemoH:   110,
			},
			{
				Heading: "Subtle variant",
				Prose:   `<p><code>TabsSubtle</code> renders pill-styled labels without the well — the quietest strip.</p>`,
				Snippet: `tabs := lotusui.Tabs{Variant: lotusui.TabsSubtle}`,
				Demo:    "tabs/2",
				DemoH:   110,
			},
			{
				Heading: "Vertical",
				Prose:   `<p><code>Vertical</code> stacks the strip as a column — pair it with your content beside it.</p>`,
				Snippet: `tabs := lotusui.Tabs{Vertical: true}`,
				Demo:    "tabs/3",
				DemoH:   180,
			},
			{
				Heading: "Icons",
				Prose:   `<p>Each option's <code>Icon</code> renders before its label, in the tab's own ink.</p>`,
				Snippet: `tabs := lotusui.Tabs{Options: []lotusui.TabOption{
	{Label: "Files", Icon: lotusui.IconFile},
	{Label: "Changes", Icon: lotusui.IconChanges},
	{Label: "Settings", Icon: lotusui.IconSettings},
}}`,
				Demo:  "tabs/4",
				DemoH: 110,
			},
			{
				Heading: "Disabled tab",
				Prose:   `<p>A disabled option dims, drops the pointer cursor, and its clicks are skipped by <code>Update</code>.</p>`,
				Snippet: `tabs := lotusui.Tabs{Options: []lotusui.TabOption{
	{Label: "Overview"},
	{Label: "Archived", Disabled: true},
	{Label: "Settings"},
}}`,
				Demo:  "tabs/5",
				DemoH: 110,
			},
			{
				Heading: "Wrapping strip",
				Prose: `<p>The horizontal strip uses <a href="../wrap/">Wrap</a> so whole tabs flow to the
next line at intrinsic width — never a squeezed 1-character label under a narrow
<code>Max.X</code> (Split pane, or a narrow window). The enclosed well grows with wrapped
height. Resize the demo box: the strip reflows. <code>Vertical</code> stays a column Flex.</p>`,
				Snippet: `tabs := lotusui.Tabs{Options: lotusui.TabOpts(
	"Changes", "Staging", "Production", "Reviews", "Approvals", "History",
)}
// under a narrow Max.X the strip wraps to two lines`,
				Demo:  "tabs/6",
				DemoH: 180,
			},
		},
		Props: []Prop{
			{"Options", "[]TabOption", "The tabs: Label, Value, Icon and Disabled. TabOpts(\"a\",\"b\") builds label-only strips."},
			{"Update(gtx)", "method", "Processes clicks. MUST run before anything reads the selection in a frame — Layout never processes clicks."},
			{"Value() / SetValue(v)", "string", "The selected tab's value; SetValue with an unknown value clears."},
			{"Clear() / Chosen()", "method", "Leave nothing selected; whether a tab is selected."},
			{"Variant", "TabsVariant", "TabsDefault (muted well, raised active tab), TabsLine (underline strip), TabsSubtle (pill-styled)."},
			{"Vertical", "bool", "Stacks the strip as a column (Flex). Horizontal uses Wrap."},
		},
	}
}

func dialogPage() *Page {
	return &Page{
		Slug:   "dialog",
		Title:  "Dialog",
		Kicker: "The one overlay primitive: dimmed scrim, width-capped card, caller-owned visibility.",
		Intro: `<p>A dimmed scrim over the <em>entire</em> window with a width-capped
<code>SurfaceCard</code> centered on it. It wraps arbitrary content and knows nothing about
what's inside; visibility stays with the caller — the isOpen/onClose contract. Don't call
<code>Layout</code> when closed.</p>`,
		Sections: []Section{
			InstallSection("dialog"),
			{
				Heading: "Usage",
				Prose: `<p><code>onClose</code> controls backdrop dismissal: pass your close func to let a
backdrop click dismiss, or <code>nil</code> to absorb outside clicks — for typed confirmations,
anywhere an accidental dismissal costs more than it saves. Lay the dialog out at
<em>window</em> constraints (your shell's top layer / portal): inside a content column it
inherits that column's constraints — the "scrim only covers part of the window" bug this type
exists to end.</p>
<p>Escape dismisses exactly like a backdrop click (both go through <code>onClose</code>, so
absorb-only dialogs ignore it too). The first open animates on the shared clock — the scrim
fades in while the card settles upward; call <code>Appear()</code> when your isOpen transitions
closed→open to play the entrance again.</p>`,
				Snippet: `var d lotusui.Dialog
var isOpen bool

if openBtn.Clicked(gtx) && !isOpen {
	isOpen = true
	d.Appear()
}
if isOpen {
	d.Layout(th, gtx, func() { isOpen = false }, dialogBody)
}`,
				Demo:  "dialog/0",
				DemoH: 520,
			},
			{
				Heading: "Sizes",
				Prose: `<p>Width presets on the shared <code>Size</code> scale — 2XS 280dp through 2XL 840dp;
<code>Width</code> stays as the free-form override. Each button below opens the dialog at that
size.</p>`,
				Snippet: `d.Size = lotusui.Size2XS // per open — set before Appear()
d.Size = lotusui.Size2XL`,
				Demo:  "dialog/1",
				DemoH: 420,
			},
			{
				Heading: "Close button",
				Prose: `<p>Dismissable dialogs (non-nil <code>onClose</code>) render the corner ✕
automatically; <code>HideClose</code> suppresses it when the footer carries the only exits —
the demo opens without the ✕, and Save is the way out. A CUSTOM close control is any button
calling your close func.</p>`,
				Snippet: `d.HideClose = true // footer buttons are the only way out
lotusui.Button(th, &done, "Done", lotusui.ButtonProps{}) // your custom close`,
				Demo:  "dialog/2",
				DemoH: 420,
			},
			{
				Heading: "Scrollable content",
				Prose: `<p>A tall dialog stops short of the window edges; longer content scrolls INSIDE —
a height-capped list in the body, title fixed above.</p>`,
				Snippet: `func(gtx C) D {
	gtx.Constraints.Max.Y = gtx.Dp(300)
	return lotusui.ListView(th, &scroll, gtx, len(paras), para)
}`,
				Demo:  "dialog/3",
				DemoH: 560,
			},
			{
				Heading: "Sticky footer",
				Prose: `<p>The footer stays while the content scrolls — fixed siblings around the
scrolling middle; the pattern is a VStack.</p>`,
				Snippet: `lotusui.VStack(th.Space.MD,
	lotusui.LabelTitle(th, "Sticky Footer").Layout,
	scrollingBody,
	lotusui.RightAligned(lotusui.Button(th, &closeBtn, "Close", lotusui.ButtonProps{Variant: lotusui.ButtonOutline})),
)`,
				Demo:  "dialog/4",
				DemoH: 560,
			},
			{
				Heading: "Responsive width",
				Prose: `<p>Stepped width against <a href="../responsive/">Theme breakpoints</a>:
<code>Sizes</code> overrides <code>Size</code>; <code>Widths</code> (dp) overrides both when
set. Priority: Widths → Width → Sizes → Size. Resize with the dialog open.</p>`,
				Snippet: `d.Sizes = lotusui.Sizes(lotusui.SizeSM).At("md", lotusui.SizeLG).At("xl", lotusui.Size2XL)
// or free-form dp:
d.Widths = lotusui.Dps(320).At("lg", 720)`,
				Demo:  "dialog/5",
				DemoH: 420,
			},
		},
		Props: []Prop{
			{"Appear()", "method", "Restart the entrance animation — call on the closed→open transition."},
			{"HideClose", "bool", "Suppress the corner ✕ on dismissable dialogs."},
			{"Size", "Size", "Width preset: Size2XS 280dp … Size2XL 840dp (SizeMD, the default, is 480dp)."},
			{"Sizes", "ResponsiveSize", "Stepped Size; when Set(), overrides Size."},
			{"Width", "unit.Dp", "Free-form width override; zero defers to Size/Sizes."},
			{"Widths", "ResponsiveDp", "Stepped dp width; when Set(), wins over Width/Sizes/Size."},
			{"onClose", "func()", "Layout parameter — backdrop-click and Escape dismissal; nil absorbs without dismissing."},
			{"content", "layout.Widget", "Layout parameter — arbitrary content; the dialog knows nothing about what's inside."},
		},
	}
}

func menuPage() *Page {
	return &Page{
		Slug:   "dropdown-menu",
		Title:  "DropdownMenu",
		Kicker: "Labels, items, separators, checkboxes, radio groups — the full menu grammar.",
		Intro: `<p><code>DropdownMenuTrigger</code> opens the floating panel from a trigger button — press
anywhere else, Escape, or pick an item to close it (checkbox and radio menus set
<code>KeepOpen</code>, because picking is not leaving). <code>DropdownMenu</code> remains the
raw panel for inline use. The row family covers plain items, icon items, shortcut hints,
toggleable checkbox items, exclusive radio items, group labels and separators; all row STATE
lives with the caller. Submenus are on the roadmap; web-specific composition and RTL are
"from the web".</p>`,
		Sections: []Section{
			InstallSection("dropdown-menu"),
			{
				Heading: "Usage",
				Prose:   `<p>The trigger opens the panel; labels head groups, separators divide them, items act.</p>`,
				Snippet: `var menu lotusui.DropdownMenuTrigger

menu.Layout(th, gtx, "Open",
	lotusui.DropdownMenuLabel(th, "My Account"),
	lotusui.DropdownMenuItem(th, &profile, "Profile", false),
	lotusui.DropdownMenuItem(th, &billing, "Billing", false),
	lotusui.DropdownMenuSeparator(th),
	lotusui.DropdownMenuItem(th, &team, "Team", false),
	lotusui.DropdownMenuItem(th, &sub, "Subscription", false),
)`,
				Demo:  "dropdown-menu/0",
				DemoH: 420,
			},
			{
				Heading: "Icons",
				Snippet: `lotusui.DropdownMenuItemIcon(th, &rename, lotusui.IconEdit, "Rename", false)`,
				Demo:    "dropdown-menu/1",
				DemoH:   320,
			},
			{
				Heading: "Shortcuts",
				Prose:   `<p>Right-aligned keyboard hints in muted ink — display only; binding the key is yours.</p>`,
				Snippet: `lotusui.DropdownMenuShortcutItem(th, &save, "Save", "⌘S", false)`,
				Demo:    "dropdown-menu/2",
				DemoH:   320,
			},
			{
				Heading: "Checkboxes",
				Prose:   `<p>Toggleable rows with a check-mark gutter; flip your bool on <code>Clicked</code>.</p>`,
				Snippet: `if chk.Clicked(gtx) { showStatus = !showStatus }
lotusui.DropdownMenuCheckboxItem(th, &chk, "Show status bar", showStatus)
menu.KeepOpen = true // picking is not leaving`,
				Demo:  "dropdown-menu/3",
				DemoH: 360,
			},
			{
				Heading: "Checkboxes with icons",
				Prose:   `<p>A leading icon between the gutter and the label.</p>`,
				Snippet: `lotusui.DropdownMenuCheckboxItemIcon(th, &mail, lotusui.IconMail, "Email notifications", notifMail)`,
				Demo:    "dropdown-menu/4",
				DemoH:   420,
			},
			{
				Heading: "Radio group with icons",
				Snippet: `lotusui.DropdownMenuRadioItemIcon(th, &card, lotusui.IconCreditCard, "Credit Card", payment == 0)`,
				Demo:    "dropdown-menu/5",
				DemoH:   420,
			},
			{
				Heading: "Radio group",
				Prose:   `<p>Exclusive rows with a dot gutter; the group is your int.</p>`,
				Snippet: `lotusui.DropdownMenuRadioItem(th, &top, "Top", position == 0)
lotusui.DropdownMenuRadioItem(th, &bottom, "Bottom", position == 1)`,
				Demo:  "dropdown-menu/6",
				DemoH: 400,
			},
			{
				Heading: "Destructive",
				Prose:   `<p><code>danger</code> renders danger ink on a danger-tinted hover — never a saturated fill.</p>`,
				Snippet: `lotusui.DropdownMenuItem(th, &del, "Delete workspace…", true)`,
				Demo:    "dropdown-menu/7",
				DemoH:   300,
			},
			{
				Heading: "Submenu",
				Prose:   `<p><code>DropdownMenuSub</code> nests a side panel that opens while the pointer rests on the row or the panel.</p>`,
				Snippet: `var sub lotusui.DropdownMenuSub

menu.Layout(th, gtx, "Open",
	lotusui.DropdownMenuItem(th, &newTab, "New Tab", false),
	sub.Item(th, "More Tools",
		lotusui.DropdownMenuItem(th, &save, "Save Page As…", false),
		lotusui.DropdownMenuItem(th, &dev, "Developer Tools", false),
	),
)`,
				Demo:  "dropdown-menu/8",
				DemoH: 440,
			},
			{
				Heading: "Complex",
				Prose:   `<p>Everything together: label, shortcut item, icon item, separators, a toggle, a destructive exit.</p>`,
				Snippet: `lotusui.DropdownMenu(th,
	lotusui.DropdownMenuLabel(th, "My Account"),
	lotusui.DropdownMenuShortcutItem(th, &profile, "Profile", "⇧⌘P", false),
	lotusui.DropdownMenuItemIcon(th, &settings, lotusui.IconSettings, "Settings", false),
	lotusui.DropdownMenuSeparator(th),
	lotusui.DropdownMenuCheckboxItem(th, &urls, "Show full URLs", showURLs),
	lotusui.DropdownMenuSeparator(th),
	lotusui.DropdownMenuItem(th, &del, "Delete account", true),
)`,
				Demo:  "dropdown-menu/9",
				DemoH: 480,
			},
		},
		Props: []Prop{
			{"DropdownMenuTrigger.Open", "bool", "The panel's state; the trigger toggles it."},
			{"DropdownMenuTrigger.KeepOpen", "bool", "Suppress close-on-selection — checkbox/radio menus."},
			{"DropdownMenuTrigger.Width", "unit.Dp", "Max panel width; zero means min 224dp and grow with content."},
			{"DropdownMenuTrigger.Variant", "ButtonVariant", "Trigger button style; the zero value renders outline."},
			{"DropdownMenuTrigger.Icon", "string", "Optional icon; empty label + Icon = icon-only (Breadcrumb …)."},
			{"DropdownMenuTrigger.Size", "Size", "Shared size preset for the trigger button."},
			{"DropdownMenuTrigger.Align", "PopoverAlign", "Panel edge against the trigger — zero is PopoverCenter; PopoverEnd keeps right-edge menus on screen."},
			{"DropdownMenuCheckboxItemIcon / RadioItemIcon", "widget", "Gutter rows with a leading icon."},
			{"DropdownMenuItem(th, btn, label, danger)", "widget", "One action row; danger = destructive ink."},
			{"DropdownMenuItemIcon / ShortcutItem", "widget", "Leading icon; right-aligned keyboard hint."},
			{"DropdownMenuCheckboxItem / RadioItem", "widget", "Check/dot gutter; state lives with the caller."},
			{"DropdownMenuLabel / Separator", "widget", "Group heading; hairline divider."},
			{"DropdownMenuSub.Item(th, label, items…)", "widget", "The nested submenu row; side panel on hover."},
		},
	}
}

func badgePage() *Page {
	return &Page{
		Slug:   "badge",
		Title:  "Badge",
		Kicker: "Small rounded labels: four variants × seven sizes × any color.",
		Intro: `<p>A small rounded label for counts, statuses and health. The status doctrine holds
whatever the colors: tinted pill + deep ink, never saturated fill + white text.</p>`,
		Sections: []Section{
			InstallSection("badge"),
			{
				Heading: "Usage",
				Snippet: `lotusui.Badge(th, "Badge", lotusui.BadgeProps{})`,
				Demo:    "badge/0",
				DemoH:   100,
			},
			{
				Heading: "Secondary",
				Prose:   `<p>Neutral tint + ink — metadata that shouldn't shout.</p>`,
				Snippet: `lotusui.Badge(th, "Secondary", lotusui.BadgeProps{Variant: lotusui.BadgeSecondary})`,
				Demo:    "badge/1",
				DemoH:   100,
			},
			{
				Heading: "Destructive",
				Prose:   `<p>Danger tint + danger ink.</p>`,
				Snippet: `lotusui.Badge(th, "Destructive", lotusui.BadgeProps{Variant: lotusui.BadgeDestructive})`,
				Demo:    "badge/2",
				DemoH:   100,
			},
			{
				Heading: "Outline",
				Prose:   `<p>Border + ink, no fill.</p>`,
				Snippet: `lotusui.Badge(th, "Outline", lotusui.BadgeProps{Variant: lotusui.BadgeOutline})`,
				Demo:    "badge/3",
				DemoH:   100,
			},
			{
				Heading: "Ghost",
				Prose:   `<p>Ink only — no fill, no border.</p>`,
				Snippet: `lotusui.Badge(th, "Ghost", lotusui.BadgeProps{Variant: lotusui.BadgeGhost})`,
				Demo:    "badge/4",
				DemoH:   100,
			},
			{
				Heading: "With icon",
				Prose:   `<p><code>Icon</code> renders an embedded icon before the text, in the badge's ink.</p>`,
				Snippet: `lotusui.Badge(th, "Verified", lotusui.BadgeProps{Icon: lotusui.IconAccept})`,
				Demo:    "badge/5",
				DemoH:   100,
			},
			{
				Heading: "Spinner",
				Prose:   `<p><code>Start</code>/<code>End</code> render arbitrary widgets around the text — a working state's spinner.</p>`,
				Snippet: `lotusui.Badge(th, "Deleting", lotusui.BadgeProps{Variant: lotusui.BadgeDestructive,
	Start: lotusui.SpinnerTint(th, 12, th.Palette.Danger)})`,
				Demo:  "badge/6",
				DemoH: 100,
			},
			{
				Heading: "Color",
				Prose: `<p><code>Color</code> re-colors the badge from any <code>ColorScale</code>, the
pastel way — tinted fill with deep same-hue ink.</p>`,
				Snippet: `lotusui.Badge(th, "Teal", lotusui.BadgeProps{Color: lotusui.Teal})
lotusui.Badge(th, "Pink", lotusui.BadgeProps{Color: lotusui.Pink})`,
				Demo:  "badge/7",
				DemoH: 100,
			},
			{
				Heading: "Status",
				Prose: `<p>The status pairs are tokens, not schemes — set <code>Bg</code>/<code>Fg</code>
directly.</p>`,
				Snippet: `p := th.Palette
lotusui.Badge(th, "Healthy", lotusui.BadgeProps{Bg: p.SuccessBg, Fg: p.Success})
lotusui.Badge(th, "Error", lotusui.BadgeProps{Bg: p.DangerBg, Fg: p.Danger})`,
				Demo:  "badge/8",
				DemoH: 100,
			},
			{
				Heading: "Sizes",
				Snippet: `lotusui.Badge(th, "New", lotusui.BadgeProps{Size: lotusui.Size2XS})
lotusui.Badge(th, "New", lotusui.BadgeProps{Size: lotusui.Size2XL})`,
				Demo:  "badge/9",
				DemoH: 100,
			},
		},
		Props: []Prop{
			{"Variant", "BadgeVariant", "BadgeDefault (default), BadgeSecondary, BadgeDestructive, BadgeOutline, BadgeGhost."},
			{"Size", "Size", "The shared size presets (Size2XS–Size2XL)."},
			{"Color", "ColorScale", "Re-colors the badge the pastel way (SoftScheme) from any scale."},
			{"Scheme", "*Scheme", "Wins over Color: full manual slot control."},
			{"Bg / Fg", "color.NRGBA", "Raw override for token pairs (statuses); both set wins over everything."},
			{"Icon", "string", "Embedded icon before the text, in the badge's ink."},
			{"Start / End", "layout.Widget", "Arbitrary leading/trailing widgets — a Spinner; the widget owns its color."},
		},
	}
}

func cardPage() *Page {
	return &Page{
		Slug:   "card",
		Title:  "Card",
		Kicker: "The grouping surface: three variants, size-scaled padding, everything else composed.",
		Intro: `<p><code>Card</code> is a rounded panel around arbitrary content. Header, body and
footer are composition — stack labels, media and buttons; the card only owns chrome and
padding. It honors <code>Constraints.Min.Y</code>, which is what makes equal-height rows work
inside grids.</p>
<p><strong>Hard rule — pad vs Max.Y:</strong> <code>Card</code> / <code>SurfaceCard</code> inset
content by <code>CardProps.Pad()</code> (from <code>Size</code>) but do <em>not</em> shrink content
<code>Constraints.Max.Y</code> by that pad. A child that fills <code>Max.Y</code> paints under the
chrome. Budget with <code>CardProps{Size: …}.Pad()</code> (or use Split pane helpers) — never
hardcode <code>20</code>. Content <code>Min.Y</code> is zeroed before the child layouts; re-assert
<code>Min.Y = Max.Y</code> on the content if you need fill layout inside (see
<a href="../split/">SplitBoxFillScroll</a>).</p>`,
		Sections: []Section{
			InstallSection("card"),
			{
				Heading: "Usage",
				Prose: `<p>Header, content and footer are COMPOSITION — the login card: title and
description with an action at the end, form fields, full-width footer buttons.</p>`,
				Snippet: `lotusui.Card(th, lotusui.CardProps{}, lotusui.VStack(th.Space.MD,
	header("Login to your account", "Enter your email below to login to your account",
		lotusui.Button(th, &signup, "Sign Up", lotusui.ButtonProps{Variant: lotusui.ButtonLink})),
	emailField, passwordField,
	lotusui.FullWidth(lotusui.Button(th, &login, "Login", lotusui.ButtonProps{})),
	lotusui.FullWidth(lotusui.Button(th, &google, "Login with Google", lotusui.ButtonProps{Variant: lotusui.ButtonOutline})),
))`,
				Demo:  "card/0",
				DemoH: 440,
			},
			{
				Heading: "Variants",
				Prose: `<p><code>CardOutline</code> (bordered panel, default), <code>CardElevated</code>
(soft shadow — the design language's standard surface, also available as
<code>SurfaceCard</code>), <code>CardSubtle</code> (tinted fill, no border).</p>`,
				Snippet: `lotusui.Card(th, lotusui.CardProps{Variant: lotusui.CardElevated}, content)
lotusui.Card(th, lotusui.CardProps{Variant: lotusui.CardSubtle}, content)`,
				Demo:  "card/1",
				DemoH: 170,
			},
			{
				Heading: "Sizes and spacing",
				Prose: `<p>The shared <code>Size</code> presets scale content padding via
<code>CardProps.Pad()</code> (MD default 20dp); app-wide spacing customization goes through the
theme's <code>WithSpace</code> scale — the token, not a per-card knob. Same Size ladder as
Button and every other component.</p>`,
				Snippet: `lotusui.Card(th, lotusui.CardProps{Size: lotusui.Size2XL}, content)
// Budget a fill child: maxY - 2*gtx.Dp(lotusui.CardProps{}.Pad())`,
				Demo:  "card/2",
				DemoH: 560,
			},
			{
				Heading: "Image",
				Prose: `<p>The event card: a 16:9 cover (an image in your app) with a badge overlaid,
title, description, and a full-width action.</p>`,
				Snippet: `lotusui.Card(th, lotusui.CardProps{}, lotusui.VStack(th.Space.MD,
	cover, // media with lotusui.Badge overlaid
	lotusui.LabelCardTitle(th, "Design systems meetup").Layout,
	description,
	lotusui.FullWidth(lotusui.Button(th, &view, "View Event", lotusui.ButtonProps{})),
))`,
				Demo:  "card/3",
				DemoH: 220,
			},
			{
				Heading: "Edge to edge",
				Prose: `<p>A muted, scrolling well inside the card — long content (terms, logs) scrolls
while the title and the action stay fixed.</p>`,
				Snippet: `lotusui.Card(th, lotusui.CardProps{}, lotusui.VStack(th.Space.SM,
	lotusui.LabelCardTitle(th, "Terms of Service").Layout,
	func(gtx C) D { return lotusui.Scrollable(th, &scroll, gtx, terms) },
	lotusui.FullWidth(lotusui.Button(th, &accept, "Accept", lotusui.ButtonProps{})),
))`,
				Demo:  "card/4",
				DemoH: 400,
			},
			{
				Heading: "Horizontal",
				Snippet: `lotusui.Card(th, lotusui.CardProps{}, lotusui.HStack(th.Space.MD,
	media, textColumn))`,
				Demo:  "card/5",
				DemoH: 160,
			},
		},
		Props: []Prop{
			{"Variant", "CardVariant", "CardOutline (default), CardElevated, CardSubtle."},
			{"Size", "Size", "Scales content padding via Pad() (Size2XS–Size2XL)."},
			{"Pad()", "unit.Dp", "Content inset for Size — use when budgeting Max.Y around fill/scroll children."},
			{"content", "layout.Widget", "Layout parameter — arbitrary content; header/body/footer are composition."},
		},
	}
}

func gridPage() *Page {
	return &Page{
		Slug:   "grid",
		Title:  "Grid",
		Kicker: "The 2D layout primitive: equal tracks, spans, and stepped columns.",
		Intro: `<p><code>Grid</code> lays items into equal-width tracks with row-major auto-flow —
each item takes the first slot wide and deep enough for its spans. Fixed
<code>Columns</code>, or stepped <code>Cols: Cols(1).At("md", 2).At("lg", 4)</code> against
<a href="../responsive/">Theme breakpoints</a>. Row heights derive from content; spanning
items stretch.</p>`,
		Sections: []Section{
			InstallSection("grid"),
			{
				Heading: "Col span",
				Snippet: `lotusui.Grid{Columns: 4, Gap: th.Space.SM}.Layout(th, gtx,
	lotusui.Span(2, a), lotusui.Cell(b), lotusui.Cell(c),
	lotusui.Cell(d), lotusui.Cell(e),
)`,
				Demo:  "grid/0",
				DemoH: 170,
			},
			{
				Heading: "Spanning rows and columns",
				Snippet: `lotusui.Grid{Columns: 4, Gap: th.Space.SM}.Layout(th, gtx,
	lotusui.GridItem{RowSpan: 2, W: sidebar},
	lotusui.GridItem{ColSpan: 2, W: hero},
	lotusui.Cell(c),
	lotusui.GridItem{ColSpan: 4, W: footer},
)`,
				Demo:  "grid/1",
				DemoH: 220,
			},
			{
				Heading: "Responsive columns",
				Prose: `<p>When <code>Cols</code> is set it wins over <code>Columns</code>. Resize —
track count follows Theme breakpoints (defaults: md=768, lg=992).</p>`,
				Snippet: `lotusui.Grid{
	Cols: lotusui.Cols(1).At("md", 2).At("lg", 4),
	Gap:  th.Space.SM,
}.Layout(th, gtx, items...)`,
				Demo:  "grid/2",
				DemoH: 200,
			},
		},
		Props: []Prop{
			{"Columns", "int", "Fixed track count when Cols unset (min 1)."},
			{"Cols", "ResponsiveInt", "Stepped columns; Cols(n).At(\"md\", 2). When Set(), overrides Columns."},
			{"Gap / RowGap / ColGap", "unit.Dp", "Spacing; Gap covers both axes unless overridden."},
			{"Gaps", "ResponsiveDp", "Stepped gap for both axes when Set()."},
			{"GridItem.ColSpan / .RowSpan", "int", "Tracks the item spans; zero means 1. Cell(w) and Span(cols, w) are shorthands."},
			{"GridItem.ColSpans / .RowSpans", "ResponsiveInt", "Stepped spans when Set()."},
			{"Layout(th, gtx, items…)", "method", "Needs Theme for breakpoint resolve."},
		},
	}
}

func simpleGridPage() *Page {
	return &Page{
		Slug:   "simplegrid",
		Title:  "SimpleGrid",
		Kicker: "Equal cells: continuous minChildWidth or stepped Columns.",
		Intro: `<p>A lotusui extension. Continuous mode derives column count from available width /
<code>MinChildWidth</code> (never below 1, never above <code>MaxCols</code>). Stepped mode
sets <code>Columns: Cols(…).At(…)</code> against <a href="../responsive/">Theme
breakpoints</a>. Every cell in a row gets the tallest cell's height. Reach for
<code>Grid</code> when you need explicit tracks and spans.</p>`,
		Sections: []Section{
			InstallSection("grid"),
			{
				Heading: "Usage — continuous columns",
				Prose: `<p>Columns derive from <code>available width / MinChildWidth</code>. Narrow the
window and the grid reflows smoothly.</p>`,
				Snippet: `lotusui.SimpleGrid(th, gtx, items, lotusui.SimpleGridProps{
	MinChildWidth: 140, MaxCols: 4, Gap: th.Space.SM,
}, cell)`,
				Demo:  "simplegrid/0",
				DemoH: 180,
			},
			{
				Heading: "Equal-height rows",
				Prose: `<p>A measure pass finds the tallest cell per row; every sibling gets that height as
its minimum — cards in a row stay flush.</p>`,
				Snippet: `lotusui.SimpleGrid(th, gtx, entries, lotusui.SimpleGridProps{
	MinChildWidth: 140, MaxCols: 3, Gap: th.Space.SM,
}, cell)`,
				Demo:  "simplegrid/1",
				DemoH: 160,
			},
			{
				Heading: "Stepped columns",
				Prose: `<p>When <code>Columns</code> is set, minChildWidth is ignored — same Chakra
<code>columns={{…}}</code> idea as Grid.Cols.</p>`,
				Snippet: `lotusui.SimpleGrid(th, gtx, items, lotusui.SimpleGridProps{
	Columns: lotusui.Cols(1).At("sm", 2).At("lg", 4),
	Gap:     th.Space.SM,
}, cell)`,
				Demo:  "simplegrid/2",
				DemoH: 200,
			},
		},
		Props: []Prop{
			{"th", "*Theme", "Breakpoints for stepped Columns / Gaps."},
			{"items", "[]T", "One cell per item, any type."},
			{"SimpleGridProps.MinChildWidth", "unit.Dp", "Continuous mode: width / this → columns."},
			{"SimpleGridProps.MaxCols", "int", "Continuous mode upper bound."},
			{"SimpleGridProps.Columns", "ResponsiveInt", "Stepped mode when Set()."},
			{"SimpleGridProps.Gap / Gaps", "unit.Dp / ResponsiveDp", "Spacing between cells."},
			{"cell", "func(C, T) D", "Renders one item — a tile, a Card, anything."},
		},
	}
}

func listViewPage() *Page {
	return &Page{
		Slug:   "listview",
		Title:  "ListView",
		Kicker: "The virtualized list: 10,000 rows cost a screenful per frame.",
		Intro: `<p>A lotusui extension — the list family: <code>ListView</code> lays out only the rows
intersecting the viewport, <code>Scrollable</code> is its whole-content sibling for a screen's
mixed content (no portals), and <code>HoverRow</code> is the one way a list row reads as
interactive. For Floating hosts prefer <a href="../scroll-area/">ScrollArea</a>. Scroll
position for ListView/Scrollable lives in the caller's <code>widget.List</code>.</p>`,
		Sections: []Section{
			InstallSection("listview"),
			{
				Heading: "Usage — virtualization",
				Prose: `<p>Only visible rows are laid out. In the library's benchmarks a 10,000-row
<code>ListView</code> is hundreds of times cheaper per frame than laying out every row with
<code>Scrollable</code> — see <a href="../performance/">Performance</a> for the current
snapshot. The demo below scrolls 10,000 real rows — wheel over it to feel it.</p>`,
				Snippet: `var list widget.List

lotusui.ListView(th, &list, gtx, len(items), func(gtx C, i int) D {
	return row(items[i])(gtx)
})`,
				Demo:  "list",
				DemoH: 480,
			},
			{
				Heading: "Hover rows and selection",
				Prose: `<p><code>HoverRow</code> wraps a row's content: full-width click target, pointer
cursor, and the rounded hover pill — kept while <code>active</code>, so the selected row and
the hovered row speak the same language.</p>`,
				Snippet: `lotusui.HoverRow(th, &rowBtns[i], selected == i,
	lotusui.LabelBody(th, items[i].Name).Layout)`,
			},
			{
				Heading: "Scrollable — the non-virtualized sibling",
				Prose: `<p><code>Scrollable</code> lays its whole content out every frame — right for a
form or a settings screen, wrong past a few dozen rows. Reach for <code>ListView</code> the
moment a collection can grow with data. Prefer <a href="../scroll-area/">ScrollArea</a>
when content hosts Floating widgets. Keep <code>Scrollable</code> when you want
material.List's scrollbar and no floating layer.</p>
<p><strong>Hard rule — List vs Flexed:</strong> <code>material.List</code> (and thus
<code>Scrollable</code> / <code>ListView</code> items) measures with effectively unbounded
<code>Max.Y</code>. A <code>layout.Flexed</code> “fill remaining” <em>inside a list item cannot
work</em> (pinned footers break). For fill panes with Flexed body + Rigid footer, use bounded
<code>Min.Y = Max.Y</code> <strong>without</strong> an outer list — see
<a href="../split/">SplitBoxFillScroll</a>.</p>`,
				Snippet: `var scroll widget.List
lotusui.Scrollable(th, &scroll, gtx, screenContent)`,
			},
		},
		Props: []Prop{
			{"list", "*widget.List", "Scroll position; must outlive the frame. (ListView / Scrollable)"},
			{"count", "int", "Number of rows. (ListView)"},
			{"row", "func(C, int) D", "Lays out row i at the viewport's width. (ListView)"},
		},
	}
}

func scrollAreaPage() *Page {
	return &Page{
		Slug:   "scroll-area",
		Title:  "Scroll Area",
		Kicker: "Bounded viewport scroll that lets Floating portals escape.",
		Intro: `<p>shadcn Scroll Area — whole-content scrolling inside a clipped viewport
<strong>without</strong> <code>layout.List</code>'s <code>op.Record</code> macros, with a
macOS-style overlay <code>Scrollbar</code> (shadcn composition:
<code>ScrollArea</code> → <code>ScrollBar</code>). Floating (Select, Menu, Popover,
Tooltip, HoverCard) records into <code>op.Defer</code>; List-backed scroll traps those
ops so open panels look like they push the page. <code>ScrollArea</code> clips and
offsets on the root ops stack so portals paint correctly.</p>
<p>Prefer <code>ScrollArea</code> for screens that host Floating. Prefer
<a href="../listview/">ListView</a> for long virtualized collections.
<code>Scrollable</code> remains the material.List scrollbar path when there are no portals.</p>`,
		Sections: []Section{
			InstallSection("scroll-area"),
			{
				Heading: "Usage",
				Prose: `<p>State (<code>Offset</code>) must outlive the frame. Give the viewport a
bounded height (parent constraints or an explicit max). The overlay thumb
fades in on hover / scroll (macOS default) and is draggable.</p>`,
				Snippet: `var page lotusui.ScrollArea
page.Layout(th, gtx, screenContent)`,
				Demo:  "scroll-area/0",
				DemoH: 280,
			},
			{
				Heading: "Composition",
				Prose: `<p>shadcn builds <code>ScrollArea</code> around <code>ScrollBar</code>. In lotusui the
thumb is <code>Scrollbar</code> / <code>ScrollbarProps</code> — durable drag state lives on
<code>ScrollArea</code>; tune via <code>ScrollAreaProps.Scrollbar</code>, or lay
<code>Scrollbar.Layout</code> out yourself for a custom track box.</p>`,
				Snippet: `page.LayoutWith(th, gtx, lotusui.ScrollAreaProps{
	Scrollbar: lotusui.ScrollbarProps{Variant: lotusui.ScrollbarAlways},
}, content)`,
			},
			{
				Heading: "Horizontal",
				Prose:   `<p>Set <code>Horizontal: true</code> for sideways scroll (works strip, tag row). The zero value scrolls vertically — <code>layout.Axis</code> is not used (its zero is Horizontal). The overlay thumb follows the axis.</p>`,
				Snippet: `var strip lotusui.ScrollArea
strip.Horizontal = true
strip.Layout(th, gtx, wideRow)`,
				Demo:  "scroll-area/1",
				DemoH: 160,
			},
			{
				Heading: "Floating inside",
				Prose: `<p>Open a Select (or Menu) while scrolling — the panel paints above the
frame, not trapped inside the scroller. That is the whole reason ScrollArea exists
beside Scrollable.</p>`,
				Snippet: `var page lotusui.ScrollArea
page.Layout(th, gtx, lotusui.VStack(th.Space.MD,
	lotusui.LabelBody(th, "…").Layout,
	func(gtx C) D { return env.Layout(th, gtx, "Environment") },
))`,
				Demo:  "scroll-area/2",
				DemoH: 320,
			},
			{
				Heading: "Always visible",
				Prose: `<p>chakra <code>variant="always"</code> — keep the thumb painted while content
overflows (settings panes, dense tables). Default is <code>ScrollbarHover</code>
(macOS overlay fade).</p>`,
				Snippet: `page.LayoutWith(th, gtx, lotusui.ScrollAreaProps{
	Scrollbar: lotusui.ScrollbarProps{Variant: lotusui.ScrollbarAlways},
}, content)`,
				Demo:  "scroll-area/3",
				DemoH: 280,
			},
			{
				Heading: "Sizes",
				Prose: `<p>Thumb thickness uses the shared <code>Size</code> enum (MD ≈ 6dp — the macOS
overlay default). Reach for SM in dense chrome, LG when the pane is the
primary scroll surface.</p>`,
				Snippet: `lotusui.ScrollbarProps{Size: lotusui.SizeSM}
lotusui.ScrollbarProps{Size: lotusui.SizeLG}`,
				Demo:  "scroll-area/4",
				DemoH: 280,
			},
			{
				Heading: "Thumb color",
				Prose: `<p>Pass <code>Color</code> (a <code>ColorScale</code>) so the thumb walks the interaction
ladder, or <code>Scheme</code> for full slot control. Zero keeps the muted
FgSubtle overlay.</p>`,
				Snippet: `lotusui.ScrollbarProps{Color: lotusui.Teal}`,
				Demo:    "scroll-area/5",
				DemoH:   280,
			},
			{
				Heading: "Show track",
				Prose: `<p>macOS overlay has no track. Set <code>ShowTrack: true</code> when the pane needs
a clearer gutter (sidebars, code panels).</p>`,
				Snippet: `lotusui.ScrollbarProps{
	Variant:   lotusui.ScrollbarAlways,
	ShowTrack: true,
}`,
				Demo:  "scroll-area/6",
				DemoH: 280,
			},
			{
				Heading: "NoShadowRoom / NoScrollbar",
				Prose: `<p>Page-level scroll insets <code>shadowRoom</code> for card shadows.
Pane helpers that already budget Card pad pass
<code>NoShadowRoom: true</code> to avoid double inset.
<code>NoScrollbar: true</code> hides the overlay when a parent owns chrome.</p>`,
				Snippet: `page.LayoutWith(th, gtx, lotusui.ScrollAreaProps{
	NoShadowRoom: true,
	NoScrollbar:  true,
}, content)`,
			},
			{
				Heading: "ScrollTo / Reset",
				Prose:   `<p>Jump on route changes or TOC clicks; clamp happens on the next Layout.</p>`,
				Snippet: `page.Reset()
page.ScrollTo(y)`,
			},
			{
				Heading: "From the web",
				Prose: `<p>shadcn/Base UI also document RTL scrollbar placement and native
<code>overflow</code> styling. lotusui paints an overlay thumb in the logical
end of the viewport; RTL mirroring of chrome is not wired yet (➖).</p>`,
			},
		},
		Props: []Prop{
			{"Offset", "int", "Scroll position in px; must outlive the frame."},
			{"Horizontal", "bool", "Scroll on X when true; zero value scrolls vertically."},
			{"NoShadowRoom", "bool", "Skip shadowRoom inset (ScrollAreaProps)."},
			{"NoScrollbar", "bool", "Hide the overlay thumb (ScrollAreaProps)."},
			{"Scrollbar", "ScrollbarProps", "Overlay thumb: Variant, Size, Color, Scheme, ShowTrack, Horizontal."},
			{"ScrollbarHover", "ScrollbarVariant", "Fade in on hover/scroll (default, macOS)."},
			{"ScrollbarAlways", "ScrollbarVariant", "Stay visible while content overflows."},
			{"Reset()", "method", "Scroll back to the start."},
			{"ScrollTo(offset)", "method", "Jump to a content offset."},
			{"Scrollbar.Layout", "method", "Standalone track-box paint; returns content-fraction delta."},
		},
	}
}

func splitPage() *Page {
	return &Page{
		Slug:   "split",
		Title:  "Split",
		Kicker: "The carousel of panes — content is revealed and hidden, never reflowed.",
		Intro: `<p><code>Split</code> arranges a screen's main panes as a carousel. All panes live on
one horizontal strip behind an overflow-hidden viewport; the screen's depth decides what shows.
Two eased values drive every transition — pane 0's width fraction and the strip's X
translation. Panes are only ever <em>revealed</em> or <em>hidden</em> by the viewport's edges;
content never re-flows mid-animation.</p>
<p><strong>Scroll grammar:</strong> <code>SplitBox</code> wraps <code>SurfaceCard</code>.
Use <code>SplitColumnScroll</code> (column scrolls, stacked natural cards),
<code>SplitBoxScroll</code> (hug then scroll inside one card), or
<code>SplitBoxFillScroll</code> (flush height, Flexed body + Rigid footer — <em>no</em> outer list).
Pane helpers subtract <code>2×CardProps{}.Pad()</code> from content <code>Max.Y</code>; they do not
nest <code>Scrollable</code>.</p>`,
		Sections: []Section{
			InstallSection("split"),
			{
				Heading: "Usage",
				Prose: `<p>Pass every pane the screen can show — including currently hidden ones — so
mid-animation frames render content sliding, never popping. Build each pane with
<code>SplitBox</code> so appearing and disappearing panes look identical.
<code>LayoutSolo</code> is the full-width pivot: two panes, one visible at a time.</p>`,
				Snippet: `var s lotusui.Split

s.Layout(gtx, th.Space.MD, depth,
	lotusui.SplitBox(th, listPane),
	lotusui.SplitBox(th, detailPane),
	lotusui.SplitBox(th, editorPane),
)`,
				Demo:  "split/0",
				DemoH: 380,
			},
			{
				Heading: "VSlide",
				Prose: `<p><code>VSlide</code> is Split's vertical sibling for full-screen pivots: the
over-screen slides up from the bottom while the base screen is pushed up and away. Settled, only
the visible screen is laid out; mid-flight, recorded frames are merely translated — verbatim
pixels sliding, so moving content can never reflow, compress, or flicker.</p>`,
				Snippet: `var v lotusui.VSlide

v.Layout(gtx, th, expanded, baseScreen, overScreen)`,
				Demo:  "split/1",
				DemoH: 300,
			},
			{
				Heading: "Column scroll",
				Prose: `<p>Independent columns: stack natural-height cards inside
<code>SplitColumnScroll</code> — the column scrolls, not each card.</p>`,
				Snippet: `var col widget.List

lotusui.SplitColumnScroll(th, &col, 0, lotusui.VStack(th.Space.MD,
	lotusui.SplitBox(th, cardA),
	lotusui.SplitBox(th, cardB),
	lotusui.SplitBox(th, cardC),
))`,
				Demo:  "split/2",
				DemoH: 360,
			},
			{
				Heading: "Pane scroll",
				Prose: `<p><code>SplitBoxScroll</code> hugs while short; past <code>maxH</code> the body
scrolls inside the card. Content <code>Max.Y</code> is <code>maxH − 2×CardProps{}.Pad()</code>.</p>`,
				Snippet: `var pane widget.List

lotusui.SplitBoxScroll(th, &pane, 0, tallBody)`,
				Demo:  "split/3",
				DemoH: 320,
			},
			{
				Heading: "Fill + pinned footer",
				Prose: `<p><code>SplitBoxFillScroll</code> sets chrome height to <code>maxH</code> and
re-asserts <code>Min.Y = Max.Y</code> on content after Card zeros it — Flexed body + Rigid
footer work. Do not wrap that content in a list.</p>`,
				Snippet: `lotusui.SplitBoxFillScroll(th, &unused, 0, func(gtx C) D {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Flexed(1, scrollableBody),
		layout.Rigid(footerActions),
	)
})`,
				Demo:  "split/4",
				DemoH: 320,
			},
		},
		Props: []Prop{
			{"gap", "unit.Dp", "Layout parameter — spacing between the strip's panes."},
			{"depth", "int", "Which panes the viewport shows: 0 = pane 0 full width; n = panes n-1 and n side by side."},
			{"boxes", "...layout.Widget", "Every pane the screen can show, hidden ones included, so mid-animation frames slide instead of pop."},
			{"SplitBox", "widget", "SurfaceCard wrapper for a pane."},
			{"SplitColumnScroll", "widget", "Fixed-height column viewport; stack SplitBoxes inside."},
			{"SplitBoxScroll", "widget", "Hug while short; scroll inside the card past maxH (budgets CardProps{}.Pad())."},
			{"SplitBoxFillScroll", "widget", "Flush maxH card; Min.Y=Max.Y for Flexed body + Rigid footer (no list)."},
		},
	}
}
