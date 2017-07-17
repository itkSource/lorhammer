#!/bin/bash

#VERSION=`git describe --exact-match --tags HEAD`
#if [ -z "$VERSION" ]; then
#    VERSION="0.0.0"
#fi
#DATE_BUILD=`date +%Y-%m-%d\_%H:%M`
#
#echo "delete build directory"
#rm -rf build
#
#echo "use version ${VERSION} and build time ${DATE_BUILD}"
#
#echo "compile linux"
#go build -race -ldflags "-s -w -extldflags '-static' -X main.VERSION=${VERSION} -X main.DATE_BUILD=${DATE_BUILD}" -o "build/lorhammer_${VERSION}" src/lorhammer/main.go
#echo "compile linux orchestrator"
#go build -race -ldflags "-s -w -extldflags '-static' -X main.VERSION=${VERSION} -X main.DATE_BUILD=${DATE_BUILD}" -o "build/orchestrator_${VERSION}" src/orchestrator/main.go
#echo "compile windows 386 lorhammer"
#GOOS=windows GOARCH=386 go build -ldflags "-s -w -extldflags '-static' -X main.VERSION=${VERSION} -X main.DATE_BUILD=${DATE_BUILD}" -o "build/lorhammer_${VERSION}-386.exe" src/lorhammer/main.go
#echo "compile windows amd64 lorhammer"
#CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -ldflags "-s -w -extldflags '-static' -X main.VERSION=${VERSION} -X main.DATE_BUILD=${DATE_BUILD}" -o "build/lorhammer_${VERSION}-amd64.exe" src/lorhammer/main.go
#
#echo "compile darwin 386 lorhammer"
#GOOS=darwin GOARCH=386 go build -ldflags "-s -w -extldflags '-static' -X main.VERSION=${VERSION} -X main.DATE_BUILD=${DATE_BUILD}" -o "build/lorhammer_${VERSION}-darwin-386" src/lorhammer/main.go

docker run --rm -v "$(pwd)":/go/src/lorhammer/ golang:alpine \
  /bin/sh -c 'apk update --no-cache \
           && apk add --no-cache ca-certificates git wget ruby ruby-dev build-base libffi-dev tar rpm \
           && update-ca-certificates &>/dev/null \
           && wget -q -O /tmp/goreleaser.tar.gz \
              https://github.com/goreleaser/goreleaser/releases/download/v0.27.4/goreleaser_Linux_x86_64.tar.gz \
           && mkdir /tmp/goreleaser && tar xf /tmp/goreleaser.tar.gz -C /tmp/goreleaser \
           && gem install --no-ri --no-rdoc fpm \
           && cd /go/src/lorhammer/ && /tmp/goreleaser/goreleaser --rm-dist --skip-publish'