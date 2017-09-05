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
$(HUGO):
	go get github.com/kardianos/govendor
	govendor get github.com/spf13/hugo

$(GODOCDOWN):
	go get github.com/robertkrimen/godocdown/godocdown

$(MINIFY):
	go get github.com/tdewolff/minify/cmd/minify

$(THEME):
	mkdir -p $(THEME)
	git clone https://github.com/itkSource/hugorha.git doc/themes/hugorha

.PHONY: godoc
godoc: $(GODOCDOWN)
	@echo "write godoc"
	@echo -e "---\ntitle: \"GoDoc\"\nmenu: \n    main:\n        weight: 21\nsubnav: \"true\"\n---\n" > doc/content/godoc.md
	@for p in `go list ./src/...`; do \
		$(GODOCDOWN) $$p >> doc/content/godoc.md ; \
	done

define addFile
    cp $(1) $(2)
    sed -i -E "s/doc\/static\/images\/([^\)]+)/\/images\/\1/g" $(2)
    echo -e "---\ntitle: \"$(3)\"\nmenu: \n    main:\n        weight: $(4)\nsubnav: \"$(5)\"\n---\n`cat $(2)`" > $(2)
endef

.PHONY: cpDocFiles
cpDocFiles:
	$(call addFile,README.md,doc/content/README.md,"Lorhammer",1,"true")
	$(call addFile,CHANGELOG.md,doc/content/CHANGELOG.md,"Changelog",10,"true")
	$(call addFile,CONTRIBUTING.md,doc/content/CONTRIBUTING.md,"Contributing",20,"true")
	$(call addFile,LICENCE.md,doc/content/LICENCE.md,"Licence",30,"false")

.PHONY: doc
doc: $(HUGO) $(MINIFY) $(THEME) godoc cpDocFiles
	hugo -s doc
	$(MINIFY) --recursive --output ./doc/public_min/ ./doc/public
	cp -u -r ./doc/public/images/. ./doc/public_min/images

.PHONY: docDev
docDev: $(HUGO) $(MINIFY) $(THEME) godoc cpDocFiles
	export HUGO_BASEURL="http://127.0.0.1:1313/"
	-$(HUGO) server -wDs doc
	!!!
	rm doc/content/README.md
	rm doc/content/CHANGELOG.md
	rm doc/content/CONTRIBUTING.md
	rm doc/content/LICENCE.md
	rm doc/content/godoc.md

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