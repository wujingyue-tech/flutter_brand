package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if isFlutterApp(dir) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("run from the Flutter app root (pubspec.yaml + android/ + ios/)")
		}
		dir = parent
	}
}

func isFlutterApp(dir string) bool {
	return fileExists(filepath.Join(dir, "pubspec.yaml")) &&
		fileExists(filepath.Join(dir, "android", "app", "build.gradle.kts")) &&
		fileExists(filepath.Join(dir, "ios", "Runner.xcodeproj", "project.pbxproj"))
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func resolvePath(root, p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(root, p)
}
