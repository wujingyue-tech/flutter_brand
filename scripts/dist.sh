#!/bin/sh
# Cross-compile flutter_brand for macOS, Linux, and Windows.
# Usage: scripts/dist.sh [version]
set -eu
cd "$(dirname "$0")/.."

VERSION="${1:-dev}"
BIN=flutter_brand
OUT=dist
rm -rf "$OUT"
mkdir -p "$OUT"

build() {
	goos=$1
	goarch=$2
	ext=""
	if [ "$goos" = windows ]; then
		ext=.exe
	fi
	name="${BIN}_${VERSION}_${goos}_${goarch}"
	dir="$OUT/$name"
	mkdir -p "$dir"
	echo "building $name"
	CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" go build -trimpath \
		-ldflags "-s -w -X main.version=${VERSION}" \
		-o "$dir/${BIN}${ext}" .
	if [ "$goos" = windows ]; then
		(cd "$OUT" && zip -q -r "${name}.zip" "$name")
	else
		tar -C "$OUT" -czf "$OUT/${name}.tar.gz" "$name"
	fi
	rm -rf "$dir"
}

build darwin arm64
build darwin amd64
build linux amd64
build linux arm64
build windows amd64
build windows arm64

(
	cd "$OUT"
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum * >checksums.txt
	else
		shasum -a 256 * >checksums.txt
	fi
)

echo "wrote $OUT"
ls -l "$OUT"
