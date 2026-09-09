package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var bundleID = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*(\.[a-zA-Z][a-zA-Z0-9_]*)+$`)

func (a *app) packageCmd() *cobra.Command {
	var id string
	cmd := &cobra.Command{
		Use:   "package",
		Short: "Android applicationId/namespace + iOS PRODUCT_BUNDLE_IDENTIFIER",
		RunE: func(*cobra.Command, []string) error {
			return setBundleID(a.root, id)
		},
	}
	cmd.Flags().StringVar(&id, "id", "", "reverse-domain application / bundle id")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func setBundleID(root, id string) error {
	if !bundleID.MatchString(id) {
		return fmt.Errorf("package: --id must look like com.example.app, got %q", id)
	}
	androidOld, err := readGradleID(root)
	if err != nil {
		return err
	}
	iosOld, err := readIOSBundleID(root)
	if err != nil {
		return err
	}
	if err := replaceExact(root, "android/app/build.gradle.kts",
		`namespace = "`+androidOld+`"`, `namespace = "`+id+`"`); err != nil {
		return err
	}
	if err := replaceExact(root, "android/app/build.gradle.kts",
		`applicationId = "`+androidOld+`"`, `applicationId = "`+id+`"`); err != nil {
		return err
	}
	if err := moveKotlinPackage(root, androidOld, id); err != nil {
		return err
	}
	if err := rewriteIOSBundle(root, iosOld, id); err != nil {
		return err
	}
	fmt.Printf("package: android %s → %s; ios %s → %s\n", androidOld, id, iosOld, id)
	return nil
}

func readGradleID(root string) (string, error) {
	b, err := os.ReadFile(filepath.Join(root, "android/app/build.gradle.kts"))
	if err != nil {
		return "", err
	}
	re := regexp.MustCompile(`applicationId\s*=\s*"([^"]+)"`)
	m := re.FindSubmatch(b)
	if m == nil {
		return "", fmt.Errorf("build.gradle.kts: no applicationId")
	}
	return string(m[1]), nil
}

func readIOSBundleID(root string) (string, error) {
	b, err := os.ReadFile(filepath.Join(root, "ios/Runner.xcodeproj/project.pbxproj"))
	if err != nil {
		return "", err
	}
	re := regexp.MustCompile(`PRODUCT_BUNDLE_IDENTIFIER = ([^;]+);`)
	m := re.FindSubmatch(b)
	if m == nil {
		return "", fmt.Errorf("project.pbxproj: no PRODUCT_BUNDLE_IDENTIFIER")
	}
	id := strings.TrimSpace(string(m[1]))
	id = strings.TrimSuffix(id, ".RunnerTests")
	return id, nil
}

func rewriteIOSBundle(root, old, next string) error {
	path := filepath.Join(root, "ios/Runner.xcodeproj/project.pbxproj")
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	s := string(b)
	s = strings.ReplaceAll(s, "PRODUCT_BUNDLE_IDENTIFIER = "+old+".RunnerTests;",
		"PRODUCT_BUNDLE_IDENTIFIER = "+next+".RunnerTests;")
	s = strings.ReplaceAll(s, "PRODUCT_BUNDLE_IDENTIFIER = "+old+";",
		"PRODUCT_BUNDLE_IDENTIFIER = "+next+";")
	return os.WriteFile(path, []byte(s), 0o644)
}

func moveKotlinPackage(root, oldID, newID string) error {
	ktRoot := filepath.Join(root, "android/app/src/main/kotlin")
	oldDir := filepath.Join(ktRoot, filepath.Join(strings.Split(oldID, ".")...))
	newDir := filepath.Join(ktRoot, filepath.Join(strings.Split(newID, ".")...))
	src := filepath.Join(oldDir, "MainActivity.kt")
	if !fileExists(src) {
		return fmt.Errorf("MainActivity.kt not at %s", src)
	}
	b, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	body := string(b)
	body = strings.Replace(body, "package "+oldID+"\n", "package "+newID+"\n", 1)
	if err := os.MkdirAll(newDir, 0o755); err != nil {
		return err
	}
	dst := filepath.Join(newDir, "MainActivity.kt")
	if err := os.WriteFile(dst, []byte(body), 0o644); err != nil {
		return err
	}
	if filepath.Clean(src) != filepath.Clean(dst) {
		if err := os.Remove(src); err != nil {
			return err
		}
		removeEmptyParents(oldDir, ktRoot)
	}
	return nil
}

func removeEmptyParents(dir, stop string) {
	stop = filepath.Clean(stop)
	for dir != stop && dir != filepath.Dir(dir) {
		entries, err := os.ReadDir(dir)
		if err != nil || len(entries) > 0 {
			return
		}
		_ = os.Remove(dir)
		dir = filepath.Dir(dir)
	}
}
