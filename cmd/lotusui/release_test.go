package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRotateDocVersionsPatchKeepsArchives(t *testing.T) {
	dir := t.TempDir()
	site := filepath.Join(dir, "site")
	if err := os.MkdirAll(site, 0o755); err != nil {
		t.Fatal(err)
	}
	prev, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(prev) })

	initial := `[{"version":"v0.5.0","path":"/"},
 {"version":"v0.4.0","path":"/v0.4.0/"}]
`
	if err := os.WriteFile("site/versions.json", []byte(initial), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := rotateDocVersions("0.5.1", "0.5.0"); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile("site/versions.json")
	// Patch: latest is root; previous root is archived (patches appear
	// in the switcher, not only minor/major lines).
	want := `[{"version":"v0.5.1","path":"/"},
 {"version":"v0.5.0","path":"/v0.5.0/"},
 {"version":"v0.4.0","path":"/v0.4.0/"}]
`
	if string(got) != want {
		t.Fatalf("patch rotate:\n got %s\nwant %s", got, want)
	}
}

func TestRotateDocVersionsMinorArchivesPrevious(t *testing.T) {
	dir := t.TempDir()
	site := filepath.Join(dir, "site")
	if err := os.MkdirAll(site, 0o755); err != nil {
		t.Fatal(err)
	}
	prev, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(prev) })

	initial := `[{"version":"v0.5.1","path":"/"},
 {"version":"v0.4.0","path":"/v0.4.0/"}]
`
	if err := os.WriteFile("site/versions.json", []byte(initial), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := rotateDocVersions("0.6.0", "0.5.1"); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile("site/versions.json")
	want := `[{"version":"v0.6.0","path":"/"},
 {"version":"v0.5.1","path":"/v0.5.1/"},
 {"version":"v0.4.0","path":"/v0.4.0/"}]
`
	if string(got) != want {
		t.Fatalf("minor rotate:\n got %s\nwant %s", got, want)
	}
}
