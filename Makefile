SHELL := /bin/bash

VERSION=`git describe --exact-match --tags HEAD 2> /dev/null`
COMMIT=`git rev-parse HEAD`
DATE_BUILD=`date +%Y-%m-%d\_%H:%M`

BIN_DIR = $(GOPATH)/bin
GOLINT = $(BIN_DIR)/golint
DEP = $(BIN_DIR)/dep

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
$(GOLINT):
	go get -u github.com/golang/lint/golint

.PHONY: lint
lint: $(GOLINT)
	diff -u <(echo -n) <(gofmt -s -d ./src); [ $$? -eq 0 ]
	go tool vet -composites=false -shadow=true src/**/*.go
	diff -u <(echo -n) <(golint ./src/...); [ $$? -eq 0 ]


####################
## TEST
#####
.PHONY: test
test: vendor
	go test -race ./src/...

.PHONY: cover
cover: vendor
	./resources/scripts/cover.sh -terminal


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

.PHONY: doc-dev
doc-dev:
	./resources/scripts/makeDoc.sh -dev


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
	rm -rf public