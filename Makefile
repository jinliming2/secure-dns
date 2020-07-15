PACKAGE=github.com/jinliming2/secure-dns
VERSION=`git describe --tags --abbrev=0`
HASH=`git rev-parse --short HEAD`
DATE=`date +%Y%m%d%H%M%S`
LDFLAGS="-X '${PACKAGE}/versions.VERSION=${VERSION} (${DATE})' -X '${PACKAGE}/versions.BUILDHASH=${HASH}' -s -w"

.PHONY: all build clean

all: clean build

build:
	go build -v -ldflags ${LDFLAGS} -o build/secure-dns

clean:
	if [ -a build/secure-dns ] ; then rm build/secure-dns ; fi;
