package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	ansiReset  = "\033[0m"
	ansiBold   = "\033[1m"
	ansiDim    = "\033[2m"
	ansiRed    = "\033[31m"
	ansiGreen  = "\033[32m"
	ansiOrange = "\033[38;2;217;119;87m"
	ansiMuted  = "\033[38;5;245m"
	ansiBorder = "\033[38;5;240m"
)

type term struct {
	in     *bufio.Reader
	out    io.Writer
	color  bool
	reset  string
	bold   string
	dim    string
	red    string
	green  string
	orange string
	muted  string
	border string
}

func newTerm(in io.Reader, out io.Writer) *term {
	t := &term{in: bufio.NewReader(in), out: out}
	if os.Getenv("NO_COLOR") == "" && writerIsTTY(out) {
		t.color = true
		t.reset = ansiReset
		t.bold = ansiBold
		t.dim = ansiDim
		t.red = ansiRed
		t.green = ansiGreen
		t.orange = ansiOrange
		t.muted = ansiMuted
		t.border = ansiBorder
	}
	return t
}

func writerIsTTY(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	fi, err := f.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

func (t *term) printf(format string, args ...any) {
	_, _ = fmt.Fprintf(t.out, format, args...)
}

func (t *term) readLine() (string, error) {
	line, err := t.in.ReadString('\n')
	if err != nil && !(err == io.EOF && line != "") {
		if err == io.EOF {
			return "", fmt.Errorf("cancelled")
		}
		return "", err
	}
	return strings.TrimSpace(strings.TrimSuffix(line, "\r")), nil
}

func (t *term) banner() {
	const inner = 56
	rule := strings.Repeat("─", inner)
	t.printf("\n")
	t.printf("  %s╭%s╮%s\n", t.border, rule, t.reset)
	title := "  " + t.orange + "flutter_brand" + t.reset + " · init"
	t.boxRow(inner, title, 22)
	t.boxRow(inner, "", 0)
	t.boxRow(inner, "  从 github_module 模板创建 Flutter 应用", 2+displayWidth("从 github_module 模板创建 Flutter 应用"))
	t.boxRow(inner, "  项目名、应用名、包名必填；闪屏回车即跳过", 2+displayWidth("项目名、应用名、包名必填；闪屏回车即跳过"))
	t.printf("  %s╰%s╯%s\n\n", t.border, rule, t.reset)
}

func (t *term) boxRow(inner int, body string, visible int) {
	t.printf("  %s│%s%s%s│%s\n", t.border, body, pad(inner-visible), t.border, t.reset)
}

func pad(n int) string {
	if n < 0 {
		return ""
	}
	return strings.Repeat(" ", n)
}

func displayWidth(s string) int {
	w := 0
	for _, r := range s {
		w += runeWidth(r)
	}
	return w
}

func runeWidth(r rune) int {
	if r < 0x1100 {
		return 1
	}
	switch {
	case r >= 0x2E80 && r <= 0xA4CF,
		r >= 0xAC00 && r <= 0xD7A3,
		r >= 0xF900 && r <= 0xFAFF,
		r >= 0xFE10 && r <= 0xFE6F,
		r >= 0xFF00 && r <= 0xFF60,
		r >= 0xFFE0 && r <= 0xFFE6,
		r >= 0x1F300 && r <= 0x1FAFF:
		return 2
	default:
		return 1
	}
}

func (t *term) ask(label, hint, suggestion string, required bool, validate func(string) error) (string, error) {
	for {
		t.printf("  %s%s%s\n", t.bold, label, t.reset)
		if hint != "" {
			t.printf("  %s%s%s\n", t.muted, hint, t.reset)
		}
		t.printf("  %s❯%s ", t.orange, t.reset)
		line, err := t.readLine()
		if err != nil {
			return "", err
		}
		if line == "" && suggestion != "" {
			if validate != nil {
				if err := validate(suggestion); err != nil {
					t.printf("  %s✗ %s%s\n\n", t.red, err.Error(), t.reset)
					continue
				}
			}
			t.printf("  %s· %s%s\n\n", t.muted, suggestion, t.reset)
			return suggestion, nil
		}
		if line == "" {
			if required {
				t.printf("  %s✗ 这项是必填的%s\n\n", t.red, t.reset)
				continue
			}
			t.printf("  %s· 已跳过%s\n\n", t.muted, t.reset)
			return "", nil
		}
		if validate != nil {
			if err := validate(line); err != nil {
				t.printf("  %s✗ %s%s\n\n", t.red, err.Error(), t.reset)
				continue
			}
		}
		t.printf("\n")
		return line, nil
	}
}

func (t *term) confirm(title string, a initAnswers) (bool, error) {
	t.printf("  %s╭─%s %s%s%s\n", t.border, t.reset, t.bold, title, t.reset)
	t.printf("  %s│%s\n", t.border, t.reset)
	t.kv("目录", a.Dir)
	t.kv("项目名", a.Pub)
	app := a.Title
	if a.TitleZh != "" && a.TitleZh != a.Title {
		app = a.Title + "  ·  " + a.TitleZh
	}
	t.kv("应用名", app)
	t.kv("包名", a.BundleID)
	splash := "模板默认（跳过）"
	if a.Splash != "" {
		splash = a.Splash
		if a.SplashDark != "" && a.SplashDark != a.Splash {
			splash += "  ·  深色 " + a.SplashDark
		}
	}
	t.kv("闪屏", splash)
	t.printf("  %s╰─%s\n\n", t.border, t.reset)
	t.printf("  创建这个项目？ %sY/n%s\n", t.muted, t.reset)
	t.printf("  %s❯%s ", t.orange, t.reset)
	line, err := t.readLine()
	if err != nil {
		return false, err
	}
	switch strings.ToLower(line) {
	case "", "y", "yes", "是":
		t.printf("\n")
		return true, nil
	default:
		return false, nil
	}
}

func (t *term) kv(key, value string) {
	const keyCols = 6
	t.printf("  %s│%s  %s%s%s%s  %s\n", t.border, t.reset, t.muted, key, pad(keyCols-displayWidth(key)), t.reset, value)
}

func (t *term) stepOK(msg string) {
	t.printf("  %s✓%s %s\n", t.green, t.reset, msg)
}

func (t *term) stepSkip(msg string) {
	t.printf("  %s·%s %s%s%s\n", t.muted, t.reset, t.muted, msg, t.reset)
}

func (t *term) stepWait(msg string) {
	t.printf("  %s●%s %s\n", t.orange, t.reset, msg)
}

func (t *term) done(dir string, usedFVM bool) {
	run := "flutter"
	if usedFVM {
		run = "fvm flutter"
	}
	t.printf("\n  %s完成。%s接下来：\n\n", t.bold, t.reset)
	t.printf("    %scd %s%s\n", t.dim, dir, t.reset)
	t.printf("    %s%s run%s\n\n", t.dim, run, t.reset)
}

func titleFromPub(pub string) string {
	parts := strings.Split(pub, "_")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p == "" {
			continue
		}
		out = append(out, strings.ToUpper(p[:1])+p[1:])
	}
	return strings.Join(out, " ")
}

func suggestBundleID(pub string) string {
	return "com.example." + strings.ReplaceAll(pub, "_", "")
}
