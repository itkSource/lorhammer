#!/bin/bash

VERSION=`git describe --exact-match --tags HEAD`
if [ -z "$VERSION" ]; then
    VERSION="0.0.0"
fi
DATE_BUILD=`date +%Y-%m-%d\_%H:%M`

rm -rf build
go build -race -ldflags "-extldflags '-static' -X main.VERSION=${VERSION} -X main.DATE_BUILD=${DATE_BUILD}" -o "build/lorhammer" src/lorhammer/main.go
go build -race -ldflags "-extldflags '-static' -X main.VERSION=${VERSION} -X main.DATE_BUILD=${DATE_BUILD}" -o "build/orchestrator" src/orchestrator/main.go
