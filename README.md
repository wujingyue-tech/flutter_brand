# flutter_brand

Go CLI that brands a Flutter iOS/Android template: Dart package name, home-screen display name, bundle id, splash, launcher icon.

It is not a Flutter feature. Users do **not** need this repo on disk.

## Install (no clone)

macOS / Linux — downloads the matching binary from [Releases](https://github.com/wujingyue-tech/flutter_brand/releases):

```text
curl -fsSL https://raw.githubusercontent.com/wujingyue-tech/flutter_brand/main/scripts/install.sh | sh
flutter_brand init
```

If Go is already installed:

```text
go install github.com/wujingyue-tech/flutter_brand@latest
```

Windows: download `flutter_brand_v*_windows_amd64.zip` (or `arm64`) from the same Releases page, unzip, and put `flutter_brand.exe` on `PATH`.

This repository is private. One-line install and `go install` need either a public repo, or credentials:

```text
GITHUB_TOKEN=ghp_... curl -fsSL https://raw.githubusercontent.com/wujingyue-tech/flutter_brand/main/scripts/install.sh | sh
GOPRIVATE=github.com/wujingyue-tech/flutter_brand go install github.com/wujingyue-tech/flutter_brand@latest
```

Push a tag to publish binaries for darwin / linux / windows (`amd64` + `arm64`):

```text
git tag v0.1.0
git push origin v0.1.0
```

From a local checkout (maintainers):

```text
go install .
flutter_brand
flutter_brand init
flutter_brand name --pub my_app
flutter_brand display --title "My App" --title-zh "我的应用"
flutter_brand package --id com.example.myapp
flutter_brand splash --image path.png --image-dark path_dark.png
flutter_brand icon --image path.png
flutter_brand bump
```

`init` copies [github_module](https://github.com/wujingyue-tech/github_module) into a new folder and brands it. It can run from anywhere. It asks for project name, app name, and bundle id; splash is optional (Enter keeps the template images). A sibling `github_module` checkout is used when present; otherwise the repo is cloned. Non-interactive:

```text
flutter_brand init --pub my_app --title "My App" --title-zh "我的应用" --id com.example.myapp --yes
```

`go run ../flutter_brand` from the Flutter git repo does not work (that repo has no `go.mod`). `$(go env GOPATH)/bin` must be on `PATH`.
