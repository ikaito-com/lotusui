package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// cmdRelease cuts a version: it validates the release is actually
// ready (non-empty changelog, clean drift checks), then performs the
// whole ritual atomically — retitle [Unreleased], bump version.go,
// rotate site/versions.json, refresh api.txt — and prints the git
// commands to finish. It never runs git itself: committing and
// tagging stay deliberate.
func cmdRelease(args []string) error {
	fs := flagSet("release")
	bump := fs.String("bump", "", "minor | patch | major (default: inferred from the changelog)")
	explicit := fs.String("version", "", "explicit x.y.z (overrides -bump)")
	dry := fs.Bool("dry", false, "validate and print the plan without writing")
	fs.Parse(args)

	// ── Preconditions ────────────────────────────────────────────────
	changelog, err := os.ReadFile("CHANGELOG.md")
	if err != nil {
		return err
	}
	unreleased, rest, err := splitUnreleased(string(changelog))
	if err != nil {
		return err
	}
	if body := strings.TrimSpace(unreleased); body == "" || body == "Nothing yet." {
		return fmt.Errorf("CHANGELOG.md [Unreleased] is empty — nothing to release")
	}
	if problems := verifyAPI("api.txt"); len(problems) > 0 {
		return fmt.Errorf("API drifted from api.txt — finish the commit ritual first:\n  %s",
			strings.Join(problems, "\n  "))
	}
	if problems := verifyIcons("assets/icons/manifest.txt", "assets/icons", "icons_gen.go"); len(problems) > 0 {
		return fmt.Errorf("generated code drifted:\n  %s", strings.Join(problems, "\n  "))
	}

	// ── Version arithmetic ───────────────────────────────────────────
	verSrc, err := os.ReadFile("version.go")
	if err != nil {
		return err
	}
	m := regexp.MustCompile(`Version = "(\d+)\.(\d+)\.(\d+)"`).FindStringSubmatch(string(verSrc))
	if m == nil {
		return fmt.Errorf("version.go: cannot find Version constant")
	}
	major, _ := strconv.Atoi(m[1])
	minor, _ := strconv.Atoi(m[2])
	patch, _ := strconv.Atoi(m[3])
	cur := fmt.Sprintf("%d.%d.%d", major, minor, patch)

	next := *explicit
	if next == "" {
		b := *bump
		if b == "" {
			// SemVer inference from the changelog itself: new or changed
			// API → minor (breaking is allowed in v0 minors); fixes-only
			// → patch.
			if regexp.MustCompile(`### (Renamed|Removed|Changed|Added)`).MatchString(unreleased) {
				b = "minor"
			} else {
				b = "patch"
			}
			fmt.Printf("  bump inferred from changelog: %s\n", b)
		}
		switch b {
		case "major":
			next = fmt.Sprintf("%d.0.0", major+1)
		case "minor":
			next = fmt.Sprintf("%d.%d.0", major, minor+1)
		case "patch":
			next = fmt.Sprintf("%d.%d.%d", major, minor, patch+1)
		default:
			return fmt.Errorf("unknown -bump %q (minor|patch|major)", b)
		}
	}
	if next == cur {
		return fmt.Errorf("next version equals current (%s)", cur)
	}
	if major >= 1 && strings.HasPrefix(next, fmt.Sprintf("%d.", major+1)) {
		fmt.Println("  NOTE: post-v1 major bump — Go requires a /v" + m[1] + " module path change; do that first")
	}

	date := time.Now().Format("2006-01-02")
	fmt.Printf("  releasing v%s (from v%s), dated %s\n", next, cur, date)
	if *dry {
		fmt.Println("  dry run: no files written")
		return nil
	}

	// ── Mutations, all-or-nothing in spirit ─────────────────────────
	// 1. CHANGELOG: fresh empty [Unreleased], the old body under the
	//    new version heading.
	newLog := strings.Replace(rest,
		"@@UNRELEASED@@",
		"## [Unreleased]\n\nNothing yet.\n\n## ["+next+"] - "+date+"\n"+unreleased,
		1)
	if err := os.WriteFile("CHANGELOG.md", []byte(newLog), 0o644); err != nil {
		return err
	}
	// 2. version.go
	nv := strings.Replace(string(verSrc), `Version = "`+cur+`"`, `Version = "`+next+`"`, 1)
	if err := os.WriteFile("version.go", []byte(nv), 0o644); err != nil {
		return err
	}
	// 3. site/versions.json — docs switcher: every release (including
	//    patches) becomes the new root label and archives the previous
	//    root at /vPREV/ for CI to build from that tag.
	if err := rotateDocVersions(next, cur); err != nil {
		return err
	}
	// 4. API baseline refresh (Version constant changed value only —
	//    inventory is type-level, but regenerate for good measure).
	if inv, err := apiInventory("."); err == nil {
		os.WriteFile("api.txt", []byte(strings.Join(inv, "\n")+"\n"), 0o644)
	}

	fmt.Printf(`  done. Finish with:

    make check
    git add -A
    git commit -m "chore(release): v%s"
    git tag v%s
    git push --follow-tags
`, next, next)
	return nil
}

type docVersion struct {
	Version string `json:"version"`
	Path    string `json:"path"`
}

// rotateDocVersions updates site/versions.json for the docs switcher.
// Every release (patch, minor, or major) puts v{next} at "/" and archives
// the previous root at /v{cur}/ so the switcher always lists the latest
// tag — including patch releases, not only minor/major lines. On a
// minor/major bump, older archives on the previous X.Y line are dropped
// (the frozen previous root is enough for that line).
func rotateDocVersions(next, cur string) error {
	b, err := os.ReadFile("site/versions.json")
	if err != nil {
		return err
	}
	var list []docVersion
	if err := json.Unmarshal(b, &list); err != nil {
		return fmt.Errorf("site/versions.json: %w", err)
	}

	out := []docVersion{
		{Version: "v" + next, Path: "/"},
		{Version: "v" + cur, Path: "/v" + cur + "/"},
	}
	for _, e := range list {
		if e.Path == "/" {
			continue
		}
		ver := strings.TrimPrefix(e.Version, "v")
		if ver == cur {
			continue // already archived as the previous root
		}
		if !sameMinorLine(next, cur) {
			// Minor/major: drop older archives on cur's X.Y (replaced by
			// the frozen previous root) and any premature next-line entry.
			if sameMinorLine(ver, cur) || sameMinorLine(ver, next) {
				continue
			}
		}
		out = append(out, e)
	}

	return os.WriteFile("site/versions.json", []byte(compactVersionsJSON(out)+"\n"), 0o644)
}

func sameMinorLine(a, b string) bool {
	pa := strings.Split(strings.TrimPrefix(a, "v"), ".")
	pb := strings.Split(strings.TrimPrefix(b, "v"), ".")
	if len(pa) < 2 || len(pb) < 2 {
		return false
	}
	return pa[0] == pb[0] && pa[1] == pb[1]
}

func compactVersionsJSON(list []docVersion) string {
	var b strings.Builder
	b.WriteByte('[')
	for i, e := range list {
		if i > 0 {
			b.WriteString(",\n ")
		}
		fmt.Fprintf(&b, `{"version":%q,"path":%q}`, e.Version, e.Path)
	}
	b.WriteByte(']')
	return b.String()
}

// splitUnreleased returns the [Unreleased] section body and the full
// document with that section replaced by the @@UNRELEASED@@ marker.
func splitUnreleased(doc string) (body, rest string, err error) {
	re := regexp.MustCompile(`(?s)## \[Unreleased\]\n(.*?)(\n## \[|\z)`)
	m := re.FindStringSubmatchIndex(doc)
	if m == nil {
		return "", "", fmt.Errorf("CHANGELOG.md has no [Unreleased] section")
	}
	body = doc[m[2]:m[3]]
	rest = doc[:m[0]] + "@@UNRELEASED@@" + doc[m[3]:]
	return body, rest, nil
}
