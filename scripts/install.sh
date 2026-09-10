#!/bin/sh
# Install flutter_brand from GitHub Releases. No clone, no Go toolchain.
#
#   curl -fsSL https://raw.githubusercontent.com/wujingyue-tech/flutter_brand/main/scripts/install.sh | sh
#
# Private repo: export GITHUB_TOKEN=... first (or use an authenticated gh).
# Optional: VERSION=v1.2.3 PREFIX=$HOME/.local/bin
set -eu

REPO="${FLUTTER_BRAND_REPO:-wujingyue-tech/flutter_brand}"
BIN=flutter_brand

die() {
	echo "flutter_brand install: $*" >&2
	exit 1
}

need() {
	command -v "$1" >/dev/null 2>&1 || die "need $1 on PATH"
}

need curl

os=$(uname -s | tr '[:upper:]' '[:lower:]')
arch=$(uname -m)
case "$os" in
darwin) os=darwin ;;
linux) os=linux ;;
msys*|mingw*|cygwin*) os=windows ;;
*) die "unsupported OS $(uname -s); download a release zip instead" ;;
esac
case "$arch" in
x86_64|amd64) arch=amd64 ;;
arm64|aarch64) arch=arm64 ;;
*) die "unsupported arch $arch" ;;
esac

ext=tar.gz
if [ "$os" = windows ]; then
	ext=zip
	need unzip
else
	need tar
fi

github_curl() {
	if [ -n "${GITHUB_TOKEN:-}" ]; then
		curl -fsSL -H "Authorization: Bearer ${GITHUB_TOKEN}" "$@"
	else
		curl -fsSL "$@"
	fi
}

resolve_tag() {
	if [ -n "${VERSION:-}" ]; then
		echo "$VERSION"
		return
	fi
	if command -v gh >/dev/null 2>&1 && gh auth status >/dev/null 2>&1; then
		gh release view -R "$REPO" --json tagName -q .tagName
		return
	fi
	api="https://api.github.com/repos/${REPO}/releases/latest"
	json=$(github_curl -H "Accept: application/vnd.github+json" "$api") ||
		die "could not read latest release (private repo? set GITHUB_TOKEN or install gh)"
	tag=$(printf '%s' "$json" | tr ',' '\n' | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' | head -1)
	[ -n "$tag" ] || die "latest release has no tag_name"
	echo "$tag"
}

tag=$(resolve_tag)
asset="${BIN}_${tag}_${os}_${arch}.${ext}"
url="https://github.com/${REPO}/releases/download/${tag}/${asset}"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

echo "downloading $asset"
if command -v gh >/dev/null 2>&1 && gh auth status >/dev/null 2>&1; then
	gh release download -R "$REPO" -t "$tag" -p "$asset" -D "$tmp"
else
	github_curl -L -o "$tmp/$asset" "$url" || die "download failed: $url"
fi

mkdir -p "$tmp/out"
if [ "$ext" = zip ]; then
	unzip -q "$tmp/$asset" -d "$tmp/out"
else
	tar -C "$tmp/out" -xzf "$tmp/$asset"
fi

found=$(find "$tmp/out" -type f \( -name "$BIN" -o -name "${BIN}.exe" \) | head -1)
[ -n "$found" ] || die "archive did not contain $BIN"

if [ -n "${PREFIX:-}" ]; then
	dest="$PREFIX"
elif [ -w /usr/local/bin ] 2>/dev/null; then
	dest=/usr/local/bin
else
	dest="${HOME}/.local/bin"
fi
mkdir -p "$dest"
name=$BIN
if [ "$os" = windows ]; then
	name="${BIN}.exe"
fi
cp "$found" "$dest/$name"
chmod +x "$dest/$name"

echo "installed $dest/$name ($tag)"
case ":$PATH:" in
*:"$dest":*) ;;
*) echo "add $dest to PATH, then run: $BIN init" ;;
esac
