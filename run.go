package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
)

func runDart(root string, args ...string) error {
	bin := "dart"
	full := args
	if fileExists(filepath.Join(root, ".fvmrc")) {
		bin = "fvm"
		full = append([]string{"dart"}, args...)
	}
	cmd := exec.Command(bin, full...)
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %v: %w", bin, full, err)
	}
	return nil
}

func bumpBuild(root string) (int, error) {
	path := filepath.Join(root, "pubspec.yaml")
	b, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	re := regexp.MustCompile(`(?m)^(version:\s*)([0-9]+\.[0-9]+\.[0-9]+)\+(\d+)\s*$`)
	m := re.FindSubmatch(b)
	if m == nil {
		return 0, fmt.Errorf("pubspec.yaml: expected version: MAJOR.MINOR.PATCH+build")
	}
	n, err := strconv.Atoi(string(m[3]))
	if err != nil {
		return 0, err
	}
	n++
	out := re.ReplaceAll(b, []byte(fmt.Sprintf("${1}${2}+%d", n)))
	if err := os.WriteFile(path, out, 0o644); err != nil {
		return 0, err
	}
	return n, nil
}
