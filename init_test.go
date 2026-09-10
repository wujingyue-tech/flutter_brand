package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVersionFlag(t *testing.T) {
	cmd := newRoot()
	cmd.SetArgs([]string{"--version"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), version) {
		t.Fatalf("got %q", out.String())
	}
}

func TestTitleFromPub(t *testing.T) {
	if got := titleFromPub("my_shop"); got != "My Shop" {
		t.Fatalf("got %q", got)
	}
	if got := suggestBundleID("my_shop"); got != "com.example.myshop" {
		t.Fatalf("got %q", got)
	}
}

func TestDisplayWidthCJK(t *testing.T) {
	if displayWidth("init") != 4 {
		t.Fatal(displayWidth("init"))
	}
	if displayWidth("应用") != 4 {
		t.Fatal(displayWidth("应用"))
	}
}

func TestDefaultDestLeavesTemplate(t *testing.T) {
	cwd := "/tmp/github_module"
	got := defaultDest(cwd, cwd, "my_app")
	if got != "/tmp/my_app" {
		t.Fatalf("got %s", got)
	}
	got = defaultDest("/tmp/work", "/tmp/github_module", "my_app")
	if got != "/tmp/work/my_app" {
		t.Fatalf("got %s", got)
	}
}

func TestDestInsideSrc(t *testing.T) {
	if !destInsideSrc("/tmp/tpl", "/tmp/tpl/my_app") {
		t.Fatal("child should be inside")
	}
	if destInsideSrc("/tmp/tpl", "/tmp/my_app") {
		t.Fatal("sibling should not be inside")
	}
}

func TestCollectAnswersFlags(t *testing.T) {
	cwd := t.TempDir()
	tui := newTerm(strings.NewReader(""), &bytes.Buffer{})
	a, err := collectAnswers(initFlags{
		Pub:   "my_shop",
		Title: "My Shop",
		ID:    "com.acme.myshop",
		Dir:   filepath.Join(cwd, "out"),
	}, cwd, "", tui)
	if err != nil {
		t.Fatal(err)
	}
	if a.Pub != "my_shop" || a.Title != "My Shop" || a.TitleZh != "My Shop" || a.BundleID != "com.acme.myshop" {
		t.Fatalf("%+v", a)
	}
	if a.Splash != "" {
		t.Fatal("splash should stay skipped")
	}
	if a.Dir != filepath.Join(cwd, "out") {
		t.Fatalf("dir %s", a.Dir)
	}
}

func TestCollectAnswersInteractiveSkipSplash(t *testing.T) {
	cwd := t.TempDir()
	in := strings.NewReader("my_shop\n\n我的店\n\n\n")
	tui := newTerm(in, &bytes.Buffer{})
	a, err := collectAnswers(initFlags{}, cwd, "", tui)
	if err != nil {
		t.Fatal(err)
	}
	if a.Pub != "my_shop" || a.Title != "My Shop" || a.TitleZh != "我的店" || a.BundleID != "com.example.myshop" {
		t.Fatalf("%+v", a)
	}
	if a.Splash != "" {
		t.Fatal("expected skipped splash")
	}
	if a.Dir != filepath.Join(cwd, "my_shop") {
		t.Fatalf("dir %s", a.Dir)
	}
}

func TestCollectAnswersRejectsBadPub(t *testing.T) {
	cwd := t.TempDir()
	in := strings.NewReader("My-App\nmy_app\nApp\n\ncom.foo.bar\n\n")
	var out bytes.Buffer
	tui := newTerm(in, &out)
	a, err := collectAnswers(initFlags{}, cwd, "", tui)
	if err != nil {
		t.Fatal(err)
	}
	if a.Pub != "my_app" {
		t.Fatalf("%+v", a)
	}
	if !strings.Contains(out.String(), "snake_case") {
		t.Fatalf("want validation hint, got:\n%s", out.String())
	}
}

func TestCopyTemplateSkipsJunk(t *testing.T) {
	src := t.TempDir()
	write := func(rel, body string) {
		t.Helper()
		path := filepath.Join(src, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("lib/main.dart", "void main() {}")
	write(".git/HEAD", "ref: refs/heads/main")
	write("build/out.txt", "no")
	write(".dart_defines.json", `{"SENTRY_DSN":"secret"}`)
	write("app.iml", "idea")
	dest := filepath.Join(t.TempDir(), "app")
	n, err := copyTemplate(src, dest)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("copied %d files", n)
	}
	if !fileExists(filepath.Join(dest, "lib/main.dart")) {
		t.Fatal("missing dart")
	}
	if fileExists(filepath.Join(dest, ".git/HEAD")) ||
		fileExists(filepath.Join(dest, "build/out.txt")) ||
		fileExists(filepath.Join(dest, ".dart_defines.json")) ||
		fileExists(filepath.Join(dest, "app.iml")) {
		t.Fatal("junk was copied")
	}
}

func TestCopyTemplateRefusesDestInsideSrc(t *testing.T) {
	src := t.TempDir()
	_, err := copyTemplate(src, filepath.Join(src, "nested"))
	if err == nil {
		t.Fatal("want error")
	}
}

func TestBrandNewApp(t *testing.T) {
	root := t.TempDir()
	write := func(rel, body string) {
		t.Helper()
		path := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("pubspec.yaml", "name: github_module\nversion: 2026.1.0+1\n")
	write("README.md", "# github_module\n")
	write("ARCHITECTURE.md", "one seed name (`github_module`)\n")
	write(".vscode/launch.json", "{\n      \"name\": \"github_module\"\n}\n")
	write("lib/main.dart", "import 'package:github_module/app.dart';\n")
	write("android/app/build.gradle.kts", `namespace = "com.github.learn"
applicationId = "com.github.learn"
`)
	write("android/app/src/main/kotlin/com/github/learn/MainActivity.kt", "package com.github.learn\n")
	write("android/app/src/main/AndroidManifest.xml", `<manifest>
    <application android:label="@string/app_name"/>
</manifest>
`)
	write("android/app/src/main/res/values/strings.xml", `<?xml version="1.0" encoding="utf-8"?>
<resources>
    <string name="app_name">GitHub Module</string>
</resources>
`)
	write("android/app/src/main/res/values-zh/strings.xml", `<?xml version="1.0" encoding="utf-8"?>
<resources>
    <string name="app_name">GitHub 模块</string>
</resources>
`)
	write("ios/Runner/Info.plist", `<?xml version="1.0" encoding="UTF-8"?>
<plist version="1.0">
	<dict>
		<key>CFBundleDisplayName</key>
		<string>old</string>
		<key>CFBundleName</key>
		<string>old</string>
	</dict>
</plist>
`)
	write("ios/Runner/en.lproj/InfoPlist.strings", "CFBundleDisplayName = \"old\";\n")
	write("ios/Runner/zh.lproj/InfoPlist.strings", "CFBundleDisplayName = \"old\";\n")
	write("ios/Runner.xcodeproj/project.pbxproj", "PRODUCT_BUNDLE_IDENTIFIER = com.github.learn;\n")
	if err := brandNewApp(root, initAnswers{
		Pub:      "my_shop",
		Title:    "My Shop",
		TitleZh:  "我的店",
		BundleID: "com.acme.myshop",
	}); err != nil {
		t.Fatal(err)
	}
	mustContain := func(rel, want string) {
		t.Helper()
		b, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(b), want) {
			t.Fatalf("%s: want %q in\n%s", rel, want, b)
		}
	}
	mustContain("pubspec.yaml", "name: my_shop\n")
	mustContain("lib/main.dart", "package:my_shop/")
	mustContain("android/app/src/main/res/values/strings.xml", "My Shop")
	mustContain("android/app/src/main/res/values-zh/strings.xml", "我的店")
	mustContain("android/app/build.gradle.kts", `applicationId = "com.acme.myshop"`)
	if !fileExists(filepath.Join(root, "android/app/src/main/kotlin/com/acme/myshop/MainActivity.kt")) {
		t.Fatal("kotlin package not moved")
	}
}

func TestInitHelpDoesNotNeedFlutterRoot(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	cmd := newRoot()
	cmd.SetArgs([]string{"init", "--help"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Splash is optional") {
		t.Fatalf("help:\n%s", out.String())
	}
}

func TestFindLocalTemplateFrom(t *testing.T) {
	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "pubspec.yaml"), []byte("name: github_module\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(src, "android/app"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "android/app/build.gradle.kts"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(src, "ios/Runner.xcodeproj"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "ios/Runner.xcodeproj/project.pbxproj"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := findLocalTemplate(src, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if got != src {
		t.Fatalf("got %s", got)
	}
}
