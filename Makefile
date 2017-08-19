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
	@echo "getting dependency tool"
	@go get -u github.com/golang/dep/cmd/dep
	@echo "update dependencies"
	@dep ensure

lint: dep
	@echo "verify src with go vet"
	@go tool vet -composites=false -shadow=true src/**/*.go

test: lint
	@echo "let's doing some tests"
	@go test -race ./src/...

.PHONY: test dep build