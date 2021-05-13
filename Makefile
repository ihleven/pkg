OUT := pkg
PKG := github.com/ihleven/pkg
VERSION := $(shell git describe --always --long --dirty)
PKG_LIST:= $(shell go list ${PKG}/... | grep -v /vendor/)
GO_FILES:= $(shell find . -name '*.go' | grep -v /vendor/)
CLIENT_ID:=$(CLIENT_ID)
CLIENT_SECRET:=$(CLIENT_SECRET)

all: run

kunst:
	cd cmd
	go build -v -o kunst -ldflags="-X main.version=${VERSION} -X main.ClientID=${CLIENT_ID} -X main.ClientSecret=${CLIENT_SECRET}" 


generate:
	go generate

hidrive:
	go build  -o hdrv -ldflags="-X main.version=${VERSION} -X main.ClientID=${CLIENT_ID} -X main.ClientSecret=${CLIENT_SECRET}" ./cmd/hidrive

build: 
	go build -v -o ${OUT} -ldflags="-X main.version=${VERSION} -X main.ClientID=${CLIENT_ID} -X main.ClientSecret=${CLIENT_SECRET}" 

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

opal: 
	GOOS=linux GOARCH=amd64 go build -v -o ${OUT} -ldflags="-X main.version=${VERSION} -X main.ClientID=${CLIENT_ID} -X main.ClientSecret=${CLIENT_SECRET}" 
	scp ./icloud ihle@opal6.opalstack.com:~/bin/icloud

.PHONY: run server static vet lint hidrive