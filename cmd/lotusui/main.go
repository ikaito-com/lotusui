// Command lotusui is the library's developer CLI. It exists for the
// work that should happen at DEVELOPMENT time so the library and the
// apps built on it stay light at runtime: assets are fetched once and
// committed, themes are generated once and compiled. Nothing an app
// ships ever depends on this command or on the network.
//
//	lotusui icons -manifest assets/icons/manifest.txt -out assets/icons
//	lotusui theme -anchor '#319795' -pkg ui -o theme_gen.go
//
// Run it from an app with `go run github.com/ikaito-com/lotusui/cmd/lotusui`.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "icons":
		err = cmdIcons(os.Args[2:])
	case "theme":
		err = cmdTheme(os.Args[2:])
	case "init":
		err = cmdInit(os.Args[2:])
	case "verify":
		err = cmdVerify(os.Args[2:])
	case "api":
		err = cmdAPI(os.Args[2:])
	case "release":
		err = cmdRelease(os.Args[2:])
	case "gen-scales":
		err = cmdGenScales(os.Args[2:])
	case "registry":
		err = cmdRegistry(os.Args[2:])
	case "add":
		err = cmdAdd(os.Args[2:])
	case "update":
		err = cmdUpdate(os.Args[2:])
	case "skills":
		err = cmdSkills(os.Args[2:])
	case "-h", "--help", "help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "lotusui: unknown command %q\n\n", os.Args[1])
		usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "lotusui:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `lotusui — developer CLI for the lotusui design system

Commands:
  init        Scaffold a minimal themed app (main.go) with the
              go:generate lines for the build-time workflow wired in.
  icons       Fetch the SVG icons a manifest lists (Iconify API) into
              a directory you embed and register with RegisterIconFS.
              SVGs are normalized at fetch time; with -gen, typed
              icon-name constants are generated so a typo'd icon is a
              COMPILE error. Network runs HERE, once — builds never
              need it.
  theme       Generate a Go palette from a brand color (-anchor) or a
              declarative theme.json (-config): scales graded, tokens
              emitted as literals, contrast validated at build time
              (-strict makes warnings fatal).
  release     Cut a version: validates the changelog and drift checks,
              retitles [Unreleased], bumps version.go, rotates the
              docs version manifest, refreshes api.txt, and prints the
              git commands to finish. -dry previews; -bump overrides
              the changelog-inferred SemVer bump.
  api         Write the exported-API inventory (api.txt) — the
              committed baseline that verify compares against, so no
              API change can pass make check without a CHANGELOG entry.
  verify      OFFLINE drift check: manifest ↔ SVGs ↔ constants, and
              theme.json ↔ generated theme (via a source hash stamp).
              Fast enough for every make check; safe in CI (no network).
  gen-scales  Regenerate the library's stock scales as literals
              (used by lotusui's own go:generate).
  add         Vendor a component's (or block's) source into your app —
              the ownership model. The copy is yours to edit; it still
              imports the lotusui core, with references qualified
              automatically. Stamped with version + pristine hash.
  update      Merge upstream changes into vendored components: clean
              replace when untouched, a true three-way merge (base
              reconstructed from the Go module cache) when customized.
              Prints the agent-facing CHANGELOG delta.
  registry    Regenerate registry.json — the machine-readable catalog
              of components, blocks and skills. Build-time only: the
              CLI and AI agents consume it; app code never does.
  skills      Install lotusui's agent skills into the app
              (.claude/skills/lotusui) so coding agents know the
              registry, the theming system and the changelog contract.

Use "lotusui <command> -h" for the command's flags.
`)
}

func flagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ExitOnError)
	return fs
}

// reorderFlags lets flags appear after positionals (`add button -dir
// ui`) — Go's flag package stops at the first non-flag, which turns a
// trailing -dir into a component name. boolFlags marks flags that take
// no value.
func reorderFlags(args []string, boolFlags map[string]bool) []string {
	var flags, positional []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		if len(a) > 1 && a[0] == '-' {
			flags = append(flags, a)
			name := strings.TrimLeft(a, "-")
			if !strings.Contains(a, "=") && !boolFlags[name] && i+1 < len(args) {
				i++
				flags = append(flags, args[i])
			}
			continue
		}
		positional = append(positional, a)
	}
	return append(flags, positional...)
}
