package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

type initFlags struct {
	Pub        string
	Title      string
	TitleZh    string
	ID         string
	Splash     string
	SplashDark string
	From       string
	Dir        string
	Yes        bool
}

type initAnswers struct {
	Pub        string
	Title      string
	TitleZh    string
	BundleID   string
	Splash     string
	SplashDark string
	Dir        string
}

func (a *app) initCmd() *cobra.Command {
	var f initFlags
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Copy github_module and brand a new app (interactive)",
		Long: `Copy the github_module template into a new folder, then set the Dart
package name, home-screen title, and bundle id. Splash is optional.

Run from anywhere (does not need an existing Flutter app). Missing
required values are asked one at a time. Enter on splash keeps the
template images.

  flutter_brand init
  flutter_brand init --pub my_app --title "My App" --id com.example.myapp --yes`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runInit(f, cmd.InOrStdin(), cmd.OutOrStdout())
		},
	}
	cmd.Flags().StringVar(&f.Pub, "pub", "", "Dart package name (snake_case)")
	cmd.Flags().StringVar(&f.Title, "title", "", "English home-screen name")
	cmd.Flags().StringVar(&f.TitleZh, "title-zh", "", "Chinese home-screen name (defaults to --title)")
	cmd.Flags().StringVar(&f.ID, "id", "", "reverse-domain application / bundle id")
	cmd.Flags().StringVar(&f.Splash, "splash", "", "light splash PNG (omit to keep the template)")
	cmd.Flags().StringVar(&f.SplashDark, "splash-dark", "", "dark splash PNG (defaults to --splash)")
	cmd.Flags().StringVar(&f.From, "from", "", "template path (default: sibling github_module, else git clone)")
	cmd.Flags().StringVar(&f.Dir, "dir", "", "destination folder (default: ./<pub>)")
	cmd.Flags().BoolVar(&f.Yes, "yes", false, "skip the confirmation prompt")
	return cmd
}

func runInit(f initFlags, in io.Reader, out io.Writer) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	t := newTerm(in, out)
	t.banner()

	local, err := findLocalTemplate(f.From, cwd)
	if err != nil {
		return err
	}

	answers, err := collectAnswers(f, cwd, local, t)
	if err != nil {
		return err
	}
	if !f.Yes {
		ok, err := t.confirm("即将创建", answers)
		if err != nil {
			return fmt.Errorf("init: %w", err)
		}
		if !ok {
			t.printf("  %s已取消%s\n", t.muted, t.reset)
			return nil
		}
	}

	src := local
	if src == "" {
		t.stepWait("正在克隆 github_module…")
		cloned, cleanup, err := cloneDefaultTemplate()
		if err != nil {
			return err
		}
		defer cleanup()
		src = cloned
	}

	t.stepWait("正在复制模板…")
	n, err := copyTemplate(src, answers.Dir)
	if err != nil {
		return err
	}
	t.stepOK(fmt.Sprintf("已复制模板（%d 个文件）", n))

	if err := brandNewApp(answers.Dir, answers); err != nil {
		return fmt.Errorf("%w\n  不完整的项目在 %s，删掉后重试", err, answers.Dir)
	}
	t.stepOK("项目名  " + answers.Pub)
	t.stepOK("应用名  " + answers.Title)
	t.stepOK("包名    " + answers.BundleID)

	usedFVM := fileExists(filepath.Join(answers.Dir, ".fvmrc"))
	if err := runPubGet(answers.Dir); err != nil {
		t.stepSkip("pub get 未执行（之后在项目里自己跑）")
	} else {
		t.stepOK("依赖已安装")
	}

	if answers.Splash != "" {
		if err := writeSplash(answers.Dir, answers.Splash, answers.SplashDark, "", ""); err != nil {
			return fmt.Errorf("%w\n  不完整的项目在 %s，删掉后重试", err, answers.Dir)
		}
		t.stepOK("闪屏已替换")
	} else {
		t.stepSkip("闪屏保持模板默认")
	}
	if err := gitInit(answers.Dir); err != nil {
		t.stepSkip("git init 已跳过")
	} else {
		t.stepOK("已 git init")
	}

	t.done(answers.Dir, usedFVM)
	return nil
}

func collectAnswers(f initFlags, cwd, template string, t *term) (initAnswers, error) {
	var a initAnswers
	var err error

	a.Pub, err = takeOrAsk(t, strings.TrimSpace(f.Pub), "项目名",
		"Dart 包名，snake_case，同时也是文件夹名", "", true, func(s string) error {
			if !dartPackageName.MatchString(s) {
				return fmt.Errorf("须为 snake_case，例如 my_app")
			}
			return nil
		})
	if err != nil {
		return a, err
	}

	titleHint := "主屏幕显示名称"
	titleSuggest := ""
	if strings.TrimSpace(f.Title) == "" {
		titleSuggest = titleFromPub(a.Pub)
		titleHint += " · 回车即用 " + titleSuggest
	}
	a.Title, err = takeOrAsk(t, strings.TrimSpace(f.Title), "应用名", titleHint, titleSuggest, true, func(s string) error {
		if strings.TrimSpace(s) == "" {
			return fmt.Errorf("不能为空")
		}
		return nil
	})
	if err != nil {
		return a, err
	}

	if zh := strings.TrimSpace(f.TitleZh); zh != "" {
		a.TitleZh = zh
	} else if flagWasEnough(f) {
		a.TitleZh = a.Title
	} else {
		a.TitleZh, err = takeOrAsk(t, "", "中文应用名", "可留空，默认与应用名相同", a.Title, false, nil)
		if err != nil {
			return a, err
		}
	}
	if a.TitleZh == "" {
		a.TitleZh = a.Title
	}

	idHint := "Android applicationId / iOS bundle id"
	idSuggest := ""
	if strings.TrimSpace(f.ID) == "" {
		idSuggest = suggestBundleID(a.Pub)
		idHint += " · 回车即用 " + idSuggest
	}
	a.BundleID, err = takeOrAsk(t, strings.TrimSpace(f.ID), "包名", idHint, idSuggest, true, func(s string) error {
		if !bundleID.MatchString(s) {
			return fmt.Errorf("须为反向域名，例如 com.example.myapp")
		}
		return nil
	})
	if err != nil {
		return a, err
	}

	splashFlag := strings.TrimSpace(f.Splash)
	if splashFlag != "" || flagWasEnough(f) {
		if splashFlag != "" {
			if err := validateSplashPath(splashFlag); err != nil {
				return a, fmt.Errorf("init: %w", err)
			}
			abs, err := resolveUserPath(splashFlag)
			if err != nil {
				return a, err
			}
			a.Splash = abs
			dark := strings.TrimSpace(f.SplashDark)
			if dark != "" {
				if err := validateSplashPath(dark); err != nil {
					return a, fmt.Errorf("init: %w", err)
				}
				abs, err := resolveUserPath(dark)
				if err != nil {
					return a, err
				}
				a.SplashDark = abs
			}
		}
	} else {
		a.Splash, err = takeOrAsk(t, "", "闪屏图",
			"浅色 PNG 路径 · 回车跳过，使用模板默认", "", false, validateSplashPath)
		if err != nil {
			return a, err
		}
		if a.Splash != "" {
			abs, err := resolveUserPath(a.Splash)
			if err != nil {
				return a, err
			}
			a.Splash = abs
			a.SplashDark, err = takeOrAsk(t, strings.TrimSpace(f.SplashDark), "深色闪屏图",
				"可留空，默认与浅色相同", a.Splash, false, validateSplashPath)
			if err != nil {
				return a, err
			}
			if a.SplashDark != "" {
				abs, err := resolveUserPath(a.SplashDark)
				if err != nil {
					return a, err
				}
				a.SplashDark = abs
			}
		}
	}

	if strings.TrimSpace(f.Dir) != "" {
		a.Dir, err = filepath.Abs(expandHome(strings.TrimSpace(f.Dir)))
	} else {
		a.Dir = defaultDest(cwd, template, a.Pub)
	}
	if err != nil {
		return a, err
	}
	if template != "" && destInsideSrc(template, a.Dir) {
		return a, fmt.Errorf("init: destination %s is inside the template; pass --dir", a.Dir)
	}
	return a, nil
}

func flagWasEnough(f initFlags) bool {
	return strings.TrimSpace(f.Pub) != "" &&
		strings.TrimSpace(f.Title) != "" &&
		strings.TrimSpace(f.ID) != ""
}

func takeOrAsk(t *term, flagged, label, hint, suggestion string, required bool, validate func(string) error) (string, error) {
	if flagged != "" {
		if validate != nil {
			if err := validate(flagged); err != nil {
				return "", fmt.Errorf("init: %s: %w", label, err)
			}
		}
		return flagged, nil
	}
	v, err := t.ask(label, hint, suggestion, required, validate)
	if err != nil {
		return "", fmt.Errorf("init: %s: %w", label, err)
	}
	return v, nil
}

func brandNewApp(root string, a initAnswers) error {
	if err := renamePub(root, a.Pub); err != nil {
		return err
	}
	if err := setNativeAppName(root, a.Title, a.TitleZh); err != nil {
		return err
	}
	return setBundleID(root, a.BundleID)
}

func runPubGet(root string) error {
	bin := "flutter"
	args := []string{"pub", "get"}
	if fileExists(filepath.Join(root, ".fvmrc")) {
		bin = "fvm"
		args = append([]string{"flutter"}, args...)
	}
	cmd := exec.Command(bin, args...)
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func gitInit(root string) error {
	if fileExists(filepath.Join(root, ".git")) {
		return nil
	}
	cmd := exec.Command("git", "init")
	cmd.Dir = root
	return cmd.Run()
}
