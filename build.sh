#!/bin/sh
# Builds static binaries in the official Go container (no local Go needed).
#   ./build.sh            linux/amd64 (server)
#   ./build.sh all        linux/amd64, linux/arm64, darwin/arm64
# During development the library image-watermark is expected next to this
# checkout (replace directive in go.mod) and mounted into the container.
set -e
cd "$(dirname "$0")"
VERSION=$(sed -n 's/^var version = "\(.*\)"/\1/p' main.go)
TARGETS="linux/amd64"
[ "$1" = "all" ] && TARGETS="linux/amd64 linux/arm64 darwin/arm64"
LIB=""
[ -d ../image-watermark ] && LIB="-v $(cd ../image-watermark && pwd):/src/image-watermark"
docker run --rm -v "$PWD":/src/camindo-watermark $LIB -w /src/camindo-watermark -v camindo-gomod:/go/pkg/mod golang:1.24-bookworm sh -c "
	set -e
	test -z \"\$(gofmt -l .)\" || { gofmt -l .; echo 'run gofmt'; exit 1; }
	go vet ./...
	go test ./...
	for t in $TARGETS; do
		os=\${t%/*}; arch=\${t#*/}
		CGO_ENABLED=0 GOOS=\$os GOARCH=\$arch go build -trimpath -ldflags='-s -w' \
			-o dist/camindo-watermark-\$os-\$arch .
		echo built dist/camindo-watermark-\$os-\$arch
	done
"
(cd dist && shasum -a 256 camindo-watermark-* > SHA256SUMS)
echo "version $VERSION"
