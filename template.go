package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const templateRepo = "https://github.com/wujingyue-tech/github_module.git"
const templateSeed = "github_module"

var skipCopyDirs = map[string]bool{
	".git": true, ".dart_tool": true, "build": true, "Pods": true,
	"ephemeral": true, ".gradle": true, "coverage": true, ".idea": true,
	".symlinks": true, "DerivedData": true, ".fvm": true,
	".plugin_symlinks": true,
}

var skipCopyFiles = map[string]bool{
	".dart_defines.json":            true,
	".DS_Store":                     true,
	".flutter-plugins-dependencies": true,
}

func findLocalTemplate(from, cwd string) (string, error) {
	if from != "" {
		abs, err := filepath.Abs(from)
		if err != nil {
			return "", err
		}
		if !isFlutterApp(abs) {
			return "", fmt.Errorf("init: --from is not a Flutter app root: %s", abs)
		}
		return abs, nil
	}
	if env := strings.TrimSpace(os.Getenv("FLUTTER_BRAND_TEMPLATE")); env != "" {
		abs, err := filepath.Abs(env)
		if err != nil {
			return "", err
		}
		if !isFlutterApp(abs) {
			return "", fmt.Errorf("init: FLUTTER_BRAND_TEMPLATE is not a Flutter app root: %s", abs)
		}
		return abs, nil
	}
	for _, cand := range []string{
		filepath.Join(cwd, templateSeed),
		filepath.Join(filepath.Dir(cwd), templateSeed),
	} {
		if isFlutterApp(cand) {
			return cand, nil
		}
	}
	if isFlutterApp(cwd) {
		if name, err := readPubspecName(cwd); err == nil && name == templateSeed {
			return cwd, nil
		}
	}
	return "", nil
}

func cloneDefaultTemplate() (string, func(), error) {
	dir, err := os.MkdirTemp("", "flutter_brand-template-*")
	if err != nil {
		return "", nil, err
	}
	cmd := exec.Command("git", "clone", "--depth", "1", "--single-branch", templateRepo, dir)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		_ = os.RemoveAll(dir)
		return "", nil, fmt.Errorf("init: clone %s: %w", templateRepo, err)
	}
	return dir, func() { _ = os.RemoveAll(dir) }, nil
}

func defaultDest(cwd, template, pub string) string {
	base := cwd
	if samePath(cwd, template) {
		base = filepath.Dir(cwd)
	}
	return filepath.Join(base, pub)
}

func samePath(a, b string) bool {
	aa, err := filepath.Abs(a)
	if err != nil {
		return false
	}
	bb, err := filepath.Abs(b)
	if err != nil {
		return false
	}
	return filepath.Clean(aa) == filepath.Clean(bb)
}

func destInsideSrc(src, dest string) bool {
	rel, err := filepath.Rel(src, dest)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)))
}

func copyTemplate(src, dest string) (int, error) {
	if destInsideSrc(src, dest) {
		return 0, fmt.Errorf("init: destination %s is inside the template", dest)
	}
	if fileExists(dest) {
		entries, err := os.ReadDir(dest)
		if err != nil {
			return 0, err
		}
		if len(entries) > 0 {
			return 0, fmt.Errorf("init: %s already exists and is not empty", dest)
		}
	}
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return 0, err
	}
	n := 0
	err := filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		if d.IsDir() {
			if skipCopyDirs[d.Name()] {
				return filepath.SkipDir
			}
			return os.MkdirAll(filepath.Join(dest, rel), 0o755)
		}
		if skipCopyFiles[d.Name()] || strings.HasSuffix(d.Name(), ".iml") {
			return nil
		}
		if err := copyFile(path, filepath.Join(dest, rel)); err != nil {
			return err
		}
		n++
		return nil
	})
	return n, err
}

func copyFile(from, to string) error {
	if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
		return err
	}
	src, err := os.Open(from)
	if err != nil {
		return err
	}
	defer src.Close()
	info, err := src.Stat()
	if err != nil {
		return err
	}
	dst, err := os.OpenFile(to, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(dst, src)
	closeErr := dst.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func expandHome(p string) string {
	if p == "~" || strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return p
		}
		if p == "~" {
			return home
		}
		return filepath.Join(home, p[2:])
	}
	return p
}

func resolveUserPath(p string) (string, error) {
	p = expandHome(strings.TrimSpace(p))
	if p == "" {
		return "", nil
	}
	return filepath.Abs(p)
}

func validateSplashPath(p string) error {
	abs, err := resolveUserPath(p)
	if err != nil {
		return err
	}
	b, err := os.ReadFile(abs)
	if err != nil {
		return fmt.Errorf("找不到文件 %s", abs)
	}
	if len(b) < 8 || string(b[:8]) != "\x89PNG\r\n\x1a\n" {
		return fmt.Errorf("%s 不是 PNG", abs)
	}
	return nil
}
