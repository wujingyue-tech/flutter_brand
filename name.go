package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var dartPackageName = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

func (a *app) nameCmd() *cobra.Command {
	var pub string
	cmd := &cobra.Command{
		Use:   "name",
		Short: "Dart package (pubspec, package: imports, README, ARCHITECTURE seed, VS Code launch name)",
		RunE: func(*cobra.Command, []string) error {
			return renamePub(a.root, pub)
		},
	}
	cmd.Flags().StringVar(&pub, "pub", "", "new Dart package name (snake_case)")
	_ = cmd.MarkFlagRequired("pub")
	return cmd
}

func renamePub(root, pub string) error {
	if !dartPackageName.MatchString(pub) {
		return fmt.Errorf("name: --pub must be a Dart package name (snake_case), got %q", pub)
	}
	old, err := readPubspecName(root)
	if err != nil {
		return err
	}
	if old == pub {
		fmt.Println("name: nothing to change")
		return nil
	}
	n, err := replaceInRepo(root, "package:"+old+"/", "package:"+pub+"/")
	if err != nil {
		return err
	}
	if err := replacePubspecName(root, old, pub); err != nil {
		return err
	}
	if err := rewritePubNameAnchors(root, old, pub); err != nil {
		return err
	}
	fmt.Printf("name: %s → %s (%d files with package: imports)\n", old, pub, n)
	return nil
}

func rewritePubNameAnchors(root, old, next string) error {
	if err := replaceExact(root, "README.md", "# "+old+"\n", "# "+next+"\n"); err != nil {
		return err
	}
	if err := replaceExact(root, "ARCHITECTURE.md",
		"one seed name (`"+old+"`)", "one seed name (`"+next+"`)"); err != nil {
		return err
	}
	return replaceExact(root, ".vscode/launch.json",
		`"name": "`+old+`"`, `"name": "`+next+`"`)
}

func replacePubspecName(root, old, next string) error {
	path := filepath.Join(root, "pubspec.yaml")
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	out := strings.Replace(string(b), "name: "+old+"\n", "name: "+next+"\n", 1)
	return os.WriteFile(path, []byte(out), 0o644)
}

func readPubspecName(root string) (string, error) {
	b, err := os.ReadFile(filepath.Join(root, "pubspec.yaml"))
	if err != nil {
		return "", err
	}
	re := regexp.MustCompile(`(?m)^name:\s*([A-Za-z_][A-Za-z0-9_]*)\s*$`)
	m := re.FindSubmatch(b)
	if m == nil {
		return "", fmt.Errorf("pubspec.yaml: no name:")
	}
	return string(m[1]), nil
}
