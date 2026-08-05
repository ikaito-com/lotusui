package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// cmdSkills installs lotusui's agent skills into the consuming app —
// the same idea as shadcn's skills: the library ships the knowledge a
// coding agent needs to use it well (the registry catalog, the
// theming system, the add/update workflow, the changelog contract),
// as files the agent's harness discovers. Skills live in the module
// itself, so the installed copy always matches the version the app
// actually builds against.
func cmdSkills(args []string) error {
	fs := flagSet("skills")
	dir := fs.String("dir", ".claude/skills", "skills directory of the app")
	fs.Parse(args)

	moduleDir, err := lotusuiModuleDir(".")
	if err != nil {
		return err
	}
	src := filepath.Join(moduleDir, "skills", "lotusui")
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("this lotusui version ships no skills: %w", err)
	}
	dst := filepath.Join(*dir, "lotusui")
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		b, err := os.ReadFile(filepath.Join(src, e.Name()))
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dst, e.Name()), b, 0o644); err != nil {
			return err
		}
		fmt.Printf("  + %s\n", filepath.Join(dst, e.Name()))
	}
	fmt.Println("  re-run after upgrading lotusui so the skill matches the built version.")
	return nil
}
