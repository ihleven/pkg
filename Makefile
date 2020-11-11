OUT := kunstpkg
PKG := github.com/ihleven/pkg
VERSION := $(shell git describe --always --long --dirty)
PKG_LIST:= $(shell go list ${PKG}/... | grep -v /vendor/)
GO_FILES:= $(shell find . -name '*.go' | grep -v /vendor/)

all: run

generate:
	go generate

build: 
	go build -v -o ${OUT} -ldflags="-X main.version=${VERSION}" 

server: generate build clean

test:
	@go test -short ${PKG_LIST}

vet:
	@go vet ${PKG_LIST}

lint:
	@for file in ${GO_FILES} ; do \
		golint $$file ; \
	done

static: vet lint
	go build -i -v -o ${OUT}-${VERSION} -tags netgo -ldflags="-extldflags \"-static\" -w -s -X main.version=${VERSION}" ${PKG}

run: build
	./${OUT}

clean:
	-@rm ${OUT} ${OUT}-${VERSION} pkged.go

.PHONY: run server static vet lint