package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

type app struct {
	root string
}

func newRoot() *cobra.Command {
	a := &app{}
	cmd := &cobra.Command{
		Use:   "flutter_brand",
		Short: "Template branding for a Flutter iOS/Android app",
		Long: `Run from the Flutter app root (uses fvm dart when .fvmrc exists).

  go install .
  flutter_brand name --pub <snake_case>
  flutter_brand display --title "App Name" [--title-zh "应用名"]
  flutter_brand package --id <reverse.domain.id>
  flutter_brand splash --image <png> [--image-dark <png>] [--color #FFFFFF] [--color-dark #121212]
  flutter_brand icon --image <png>
  flutter_brand bump`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			if !cmd.HasParent() || cmd.Name() == "help" || cmd.Name() == "completion" {
				return nil
			}
			root, err := findRepoRoot()
			if err != nil {
				return err
			}
			a.root = root
			return nil
		},
	}
	cmd.CompletionOptions.DisableDefaultCmd = true
	cmd.AddCommand(
		a.nameCmd(),
		a.displayCmd(),
		a.packageCmd(),
		a.splashCmd(),
		a.iconCmd(),
		a.bumpCmd(),
	)
	return cmd
}

func (a *app) bumpCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "bump",
		Short: "Increment pubspec +buildNumber only (iOS launch-screen cache)",
		RunE: func(*cobra.Command, []string) error {
			n, err := bumpBuild(a.root)
			if err != nil {
				return err
			}
			fmt.Printf("buildNumber is now %d\n", n)
			return nil
		},
	}
}
