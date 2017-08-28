SHELL := /bin/bash

VERSION=`git describe --exact-match --tags HEAD 2> /dev/null`
COMMIT=`git rev-parse HEAD`
DATE_BUILD=`date +%Y-%m-%d\_%H:%M`

BIN_DIR := $(GOPATH)/bin
DEP := $(BIN_DIR)/dep

.PHONY: first
first: build

$(DEP):
	go get -u github.com/golang/dep/cmd/dep

vendor: $(DEP)
	dep ensure

.PHONY: lint
lint: vendor
	go tool vet -composites=false -shadow=true src/**/*.go

.PHONY: test
test: vendor
	go test -race ./src/...

.PHONY: build
build: vendor
	rm -rf build
	go build -race -ldflags "-extldflags '-static' -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE_BUILD}" -o "build/lorhammer" src/lorhammer/main.go
	go build -race -ldflags "-extldflags '-static' -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE_BUILD}" -o "build/orchestrator" src/orchestrator/main.go

.PHONY: clean
clean:
	rm -rf vendor
	rm -rf build
	rm -rf doc/public
	rm -rf doc/public_min
	rm -rf doc/themes