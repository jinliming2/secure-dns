export CGO_ENABLED=0

PACKAGE=github.com/jinliming2/secure-dns
VERSION=`git describe --tags --abbrev=0`
HASH=`git rev-parse --short HEAD`
DATE=`date +%Y%m%d%H%M%S`
LDFLAGS="-X '${PACKAGE}/versions.VERSION=${VERSION} (${DATE})' -X '${PACKAGE}/versions.BUILDHASH=${HASH}' -s -w"

.PHONY: all build clean

all: clean build

build: linux_amd64 linux_arm64 darwin_amd64 darwin_arm64 windows_386 windows_amd64

linux_amd64:
	GOOS=linux \
	GOARCH=amd64 \
	go build -v -ldflags ${LDFLAGS} -o build/secure-dns-linux-amd64

linux_arm64:
	GOOS=linux \
	GOARCH=arm64 \
	go build -v -ldflags ${LDFLAGS} -o build/secure-dns-linux-arm64

darwin_amd64:
	GOOS=darwin \
	GOARCH=amd64 \
	go build -v -ldflags ${LDFLAGS} -o build/secure-dns-darwin-amd64

darwin_arm64:
	GOOS=darwin \
	GOARCH=arm64 \
	go build -v -ldflags ${LDFLAGS} -o build/secure-dns-darwin-arm64

windows_386:
	GOOS=windows \
	GOARCH=386 \
	go build -v -ldflags ${LDFLAGS} -o build/secure-dns-windows-386.exe

windows_amd64:
	GOOS=windows \
	GOARCH=amd64 \
	go build -v -ldflags ${LDFLAGS} -o build/secure-dns-windows-amd64.exe

clean:
	rm build/secure-dns-* || true
