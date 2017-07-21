#!/bin/bash

VERSION=`git describe --exact-match --tags HEAD 2> /dev/null`
if [ -z "$VERSION" ]; then
    VERSION="DEV"
fi
COMMIT=`git rev-parse HEAD`
DATE_BUILD=`date +%Y-%m-%d\_%H:%M`

rm -rf build
go build -race -ldflags "-extldflags '-static' -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE_BUILD}" -o "build/lorhammer" src/lorhammer/main.go
go build -race -ldflags "-extldflags '-static' -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE_BUILD}" -o "build/orchestrator" src/orchestrator/main.go
