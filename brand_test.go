package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDartPackageName(t *testing.T) {
	ok := []string{"learn_flutter", "my_app", "a"}
	bad := []string{"LearnFlutter", "my-app", "1app", ""}
	for _, s := range ok {
		if !dartPackageName.MatchString(s) {
			t.Errorf("want valid %q", s)
		}
	}
	for _, s := range bad {
		if dartPackageName.MatchString(s) {
			t.Errorf("want invalid %q", s)
		}
	}
}

func TestRewritePubNameAnchors(t *testing.T) {
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
	write("README.md", "# old_app\n\nkeep `name --pub my_app`\n")
	write("ARCHITECTURE.md", "the template keeps one seed name (`old_app`) until you run `name`.\n")
	write(".vscode/launch.json", "{\n      \"name\": \"old_app\",\n      \"type\": \"dart\"\n}\n")
	if err := rewritePubNameAnchors(root, "old_app", "github_module"); err != nil {
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
	mustContain("README.md", "# github_module\n")
	mustContain("README.md", "name --pub my_app")
	mustContain("ARCHITECTURE.md", "one seed name (`github_module`)")
	mustContain(".vscode/launch.json", `"name": "github_module"`)
}

func TestRewritePubNameAnchorsRequiresFiles(t *testing.T) {
	err := rewritePubNameAnchors(t.TempDir(), "old_app", "new_app")
	if err == nil {
		t.Fatal("want error when README/ARCHITECTURE/launch.json are missing")
	}
}

func TestSetNativeAppName(t *testing.T) {
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
	write("ios/Runner/Info.plist", `<?xml version="1.0" encoding="UTF-8"?>
<plist version="1.0">
	<dict>
		<key>CFBundleDisplayName</key>
		<string>old</string>
		<key>CFBundleName</key>
		<string>old_pkg</string>
	</dict>
</plist>
`)
	write("ios/Runner/en.lproj/InfoPlist.strings", "CFBundleDisplayName = \"old\";\n")
	write("ios/Runner/zh.lproj/InfoPlist.strings", "CFBundleDisplayName = \"old\";\n")
	write("android/app/src/main/AndroidManifest.xml", `<manifest>
    <application android:label="@string/app_name" android:icon="@mipmap/ic_launcher"/>
</manifest>
`)
	write("android/app/src/main/res/values/strings.xml", `<?xml version="1.0" encoding="utf-8"?>
<resources>
    <string name="app_name">old</string>
</resources>
`)
	write("android/app/src/main/res/values-zh/strings.xml", `<?xml version="1.0" encoding="utf-8"?>
<resources>
    <string name="app_name">旧</string>
</resources>
`)
	if err := setNativeAppName(root, "  GitHub Module  ", "GitHub 模块"); err != nil {
		t.Fatal(err)
	}
	plist, err := os.ReadFile(filepath.Join(root, "ios/Runner/Info.plist"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(plist)
	if !strings.Contains(s, "GitHub Module") || strings.Contains(s, ">old<") || strings.Contains(s, "old_pkg") {
		t.Fatalf("plist:\n%s", s)
	}
	manifest, err := os.ReadFile(filepath.Join(root, "android/app/src/main/AndroidManifest.xml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(manifest), `android:label="@string/app_name"`) {
		t.Fatalf("manifest should reference @string/app_name:\n%s", manifest)
	}
	if strings.Contains(string(manifest), `android:label="GitHub Module"`) {
		t.Fatalf("manifest still has a literal label:\n%s", manifest)
	}
	en, err := os.ReadFile(filepath.Join(root, "android/app/src/main/res/values/strings.xml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(en), "GitHub Module") {
		t.Fatalf("en strings:\n%s", en)
	}
	zh, err := os.ReadFile(filepath.Join(root, "android/app/src/main/res/values-zh/strings.xml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(zh), "GitHub 模块") {
		t.Fatalf("zh strings:\n%s", zh)
	}
	enIOS, err := os.ReadFile(filepath.Join(root, "ios/Runner/en.lproj/InfoPlist.strings"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(enIOS), "GitHub Module") {
		t.Fatalf("en InfoPlist.strings:\n%s", enIOS)
	}
	zhIOS, err := os.ReadFile(filepath.Join(root, "ios/Runner/zh.lproj/InfoPlist.strings"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(zhIOS), "GitHub 模块") {
		t.Fatalf("zh InfoPlist.strings:\n%s", zhIOS)
	}
}

func TestSetNativeAppNameDefaultsTitleZh(t *testing.T) {
	en, zh, same, err := resolveDisplayTitles("GitHub Module", "")
	if err != nil {
		t.Fatal(err)
	}
	if en != "GitHub Module" || zh != "GitHub Module" || !same {
		t.Fatalf("en=%q zh=%q same=%v", en, zh, same)
	}
}

func TestSetNativeAppNameOnTemplateFiles(t *testing.T) {
	repo, err := filepath.Abs(filepath.Join("..", "learn_flutter"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(repo, "pubspec.yaml")); err != nil {
		t.Skip("sibling learn_flutter not present")
	}
	root := t.TempDir()
	copyFile := func(rel string) {
		t.Helper()
		src := filepath.Join(repo, rel)
		dst := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			t.Fatal(err)
		}
		b, err := os.ReadFile(src)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(dst, b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	copyFile("ios/Runner/Info.plist")
	copyFile("ios/Runner/en.lproj/InfoPlist.strings")
	copyFile("ios/Runner/zh.lproj/InfoPlist.strings")
	copyFile("android/app/src/main/AndroidManifest.xml")
	copyFile("android/app/src/main/res/values/strings.xml")
	copyFile("android/app/src/main/res/values-zh/strings.xml")
	if err := setNativeAppName(root, "GitHub Module", "GitHub 模块"); err != nil {
		t.Fatal(err)
	}
	plist, _ := os.ReadFile(filepath.Join(root, "ios/Runner/Info.plist"))
	for _, want := range []string{
		"CFBundleDisplayName", "GitHub Module",
		"com.posthog.posthog.AUTO_INIT",
		"$(PRODUCT_BUNDLE_IDENTIFIER)",
	} {
		if !strings.Contains(string(plist), want) {
			t.Fatalf("plist missing %q\n%s", want, plist)
		}
	}
	manifest, _ := os.ReadFile(filepath.Join(root, "android/app/src/main/AndroidManifest.xml"))
	for _, want := range []string{
		`android:label="@string/app_name"`,
		"android.permission.INTERNET",
		"flutterEmbedding",
		"Don't delete the meta-data below",
	} {
		if !strings.Contains(string(manifest), want) {
			t.Fatalf("manifest missing %q\n%s", want, manifest)
		}
	}
	if strings.Contains(string(manifest), `android:label="GitHub Module"`) {
		t.Fatalf("manifest still has a literal label:\n%s", manifest)
	}
}

func TestBundleID(t *testing.T) {
	if !bundleID.MatchString("com.example.myapp") {
		t.Fatal("com.example.myapp")
	}
	if bundleID.MatchString("com") || bundleID.MatchString("com.example.my-app") {
		t.Fatal("rejected ids should fail")
	}
}
