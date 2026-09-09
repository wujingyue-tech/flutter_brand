package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var skipDirNames = map[string]bool{
	".git": true, ".dart_tool": true, "build": true, "Pods": true,
	"ephemeral": true, ".gradle": true, "coverage": true, ".idea": true,
}

func replaceInRepo(root, old, next string) (int, error) {
	changed := 0
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if skipDirNames[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		switch filepath.Ext(path) {
		case ".dart", ".yaml", ".yml", ".md", ".plist", ".xml", ".kts", ".gradle", ".kt":
		default:
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		s := string(b)
		if !strings.Contains(s, old) {
			return nil
		}
		out := strings.ReplaceAll(s, old, next)
		if err := os.WriteFile(path, []byte(out), 0o644); err != nil {
			return err
		}
		changed++
		return nil
	})
	return changed, err
}

func replaceExact(root, rel, old, next string) error {
	path := filepath.Join(root, rel)
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	s := string(b)
	if !strings.Contains(s, old) {
		return fmt.Errorf("%s: did not find %q", rel, old)
	}
	return os.WriteFile(path, []byte(strings.ReplaceAll(s, old, next)), 0o644)
}
