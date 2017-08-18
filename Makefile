SHELL := /bin/bash
VERSION=`git describe --exact-match --tags HEAD 2> /dev/null`
COMMIT=`git rev-parse HEAD`
DATE_BUILD=`date +%Y-%m-%d\_%H:%M`

build: test
	@echo "it's time to build"
	@rm -rf build
	@go build -race -ldflags "-extldflags '-static' -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE_BUILD}" -o "build/lorhammer" src/lorhammer/main.go
	@go build -race -ldflags "-extldflags '-static' -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE_BUILD}" -o "build/orchestrator" src/orchestrator/main.go
	@echo "that's all."

dep:
	@go get -u github.com/golang/dep/cmd/dep
	@dep ensure

test: dep
	@echo "let's doing some tests"
	@go test -race ./src/...

.PHONY: test dep build