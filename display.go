package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/beevik/etree"
	"github.com/spf13/cobra"
)

func (a *app) displayCmd() *cobra.Command {
	var title, titleZh string
	cmd := &cobra.Command{
		Use:   "display",
		Short: "Home-screen app name (Android app_name + iOS InfoPlist.strings, en/zh)",
		RunE: func(*cobra.Command, []string) error {
			en, zh, same, err := resolveDisplayTitles(title, titleZh)
			if err != nil {
				return err
			}
			if err := setNativeAppName(a.root, en, zh); err != nil {
				return err
			}
			if same {
				fmt.Printf("display name: %s\n", en)
				fmt.Println("en and zh are the same; pass --title-zh to set a Chinese name")
				return nil
			}
			fmt.Printf("display name: %s / zh %s\n", en, zh)
			return nil
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "English home-screen name")
	cmd.Flags().StringVar(&titleZh, "title-zh", "", "Chinese home-screen name (defaults to --title)")
	_ = cmd.MarkFlagRequired("title")
	return cmd
}

func resolveDisplayTitles(title, titleZh string) (en, zh string, same bool, err error) {
	en = strings.TrimSpace(title)
	zh = strings.TrimSpace(titleZh)
	if en == "" {
		return "", "", false, fmt.Errorf("display: --title is empty")
	}
	if zh == "" {
		zh = en
	}
	return en, zh, en == zh, nil
}

func setNativeAppName(root, title, titleZh string) error {
	if err := requireAndroidLabelResource(filepath.Join(root, "android/app/src/main/AndroidManifest.xml")); err != nil {
		return err
	}
	plist := filepath.Join(root, "ios/Runner/Info.plist")
	if err := setPlistStrings(plist, map[string]string{
		"CFBundleDisplayName": title,
		"CFBundleName":        title,
	}); err != nil {
		return err
	}
	if err := writeInfoPlistStrings(filepath.Join(root, "ios/Runner/en.lproj/InfoPlist.strings"), title); err != nil {
		return err
	}
	if err := writeInfoPlistStrings(filepath.Join(root, "ios/Runner/zh.lproj/InfoPlist.strings"), titleZh); err != nil {
		return err
	}
	if err := setAndroidString(filepath.Join(root, "android/app/src/main/res/values/strings.xml"), "app_name", title); err != nil {
		return err
	}
	return setAndroidString(filepath.Join(root, "android/app/src/main/res/values-zh/strings.xml"), "app_name", titleZh)
}

func requireAndroidLabelResource(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if !strings.Contains(string(b), `android:label="@string/app_name"`) {
		return fmt.Errorf("%s: android:label must be @string/app_name", path)
	}
	return nil
}

func setPlistStrings(path string, kv map[string]string) error {
	doc := etree.NewDocument()
	if err := doc.ReadFromFile(path); err != nil {
		return err
	}
	for key, value := range kv {
		found := false
		for _, el := range doc.FindElements("//key") {
			if el.Text() != key {
				continue
			}
			str := el.NextSibling()
			if str == nil || str.Tag != "string" {
				return fmt.Errorf("%s: %s is not a string", path, key)
			}
			str.SetText(value)
			found = true
			break
		}
		if !found {
			return fmt.Errorf("%s: missing %s", path, key)
		}
	}
	return doc.WriteToFile(path)
}

func writeInfoPlistStrings(path, title string) error {
	q := strconv.Quote(title)
	body := "CFBundleDisplayName = " + q + ";\nCFBundleName = " + q + ";\n"
	return os.WriteFile(path, []byte(body), 0o644)
}

func setAndroidString(path, name, value string) error {
	doc := etree.NewDocument()
	if err := doc.ReadFromFile(path); err != nil {
		return err
	}
	for _, el := range doc.FindElements("//string") {
		if el.SelectAttrValue("name", "") == name {
			el.SetText(value)
			doc.Indent(4)
			return doc.WriteToFile(path)
		}
	}
	return fmt.Errorf("%s: missing <string name=%q>", path, name)
}
