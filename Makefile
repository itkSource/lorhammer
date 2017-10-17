SHELL := /bin/bash

VERSION=`git describe --exact-match --tags HEAD 2> /dev/null`
COMMIT=`git rev-parse HEAD`
DATE_BUILD=`date +%Y-%m-%d\_%H:%M`

BIN_DIR = $(GOPATH)/bin
DEP = $(BIN_DIR)/dep
HUGO = $(BIN_DIR)/hugo
THEME = doc/themes/hugorha
GODOCDOWN = $(BIN_DIR)/godocdown
MINIFY = $(BIN_DIR)/minify

.PHONY: first
first: build

####################
## DEP
#####
$(DEP):
	go get -u github.com/golang/dep/cmd/dep

vendor: $(DEP)
	dep ensure

####################
## LINT
#####
.PHONY: lint
lint: vendor
	go tool vet -composites=false -shadow=true src/**/*.go

####################
## TEST
#####
.PHONY: test
test: vendor
	go test -race ./src/...

####################
## BUILD
#####
.PHONY: build
build: vendor
	rm -rf build
	go build -race -ldflags "-extldflags '-static' -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE_BUILD}" -o "build/lorhammer" src/lorhammer/main.go
	go build -race -ldflags "-extldflags '-static' -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE_BUILD}" -o "build/orchestrator" src/orchestrator/main.go

####################
## DOC
#####
.PHONY: doc
doc:
	./resources/scripts/makeDoc.sh
	rm -rf doc/public
	mv doc/public_min public

####################
## CLEAN
#####
.PHONY: clean
clean:
	rm -rf vendor
	rm -rf build
	rm -rf doc/public
	rm -rf doc/public_min
	rm -rf doc/themes