export ARTIFACT:=city-suggestions
export SHELL:=/usr/bin/env bash -O extglob -c
export GO111MODULE:=on
export OS:=$(shell uname | tr '[:upper:]' '[:lower:]')

build: GOOS ?= ${OS}
build: GOARCH ?= amd64
build: clean
	GOOS=${GOOS} GOARCH=${GOARCH} CGO_ENABLED=0 go build -o ${ARTIFACT} .

test: build
	go test -v -vet=all -failfast

clean:
	rm -fv ${ARTIFACT}

release-linux:
	GOOS=linux $(MAKE) build
	tar Jcf ${ARTIFACT}-`git describe --abbrev=0 --tags`-linux-amd64.txz ${ARTIFACT}

release-darwin:
	GOOS=darwin $(MAKE) build
	tar Jcf ${ARTIFACT}-`git describe --abbrev=0 --tags`-darwin-amd64.txz ${ARTIFACT}

release: test clean release-linux release-darwin
