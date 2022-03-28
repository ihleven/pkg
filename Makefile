OUT := pkg
PKG := github.com/ihleven/pkg
VERSION := $(shell git describe --always --long --dirty)
CLIENT_ID:=$(CLIENT_ID)
CLIENT_SECRET:=$(CLIENT_SECRET)
LDFLAGS:= "-X main.version=${VERSION} -X main.ClientID=${CLIENT_ID} -X main.ClientSecret=${CLIENT_SECRET}" 


all: watch

watch:
	CompileDaemon -build="make build" -command="./${OUT}" -graceful-kill=true -color -log-prefix=false

build: 
	go build -o ${OUT} -ldflags=${LDFLAGS} 


.PHONY: all watch build