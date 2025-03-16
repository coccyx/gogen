GITHUB_OAUTH_CLIENT_ID = 39c483e563cd5cedf7c1
GITHUB_OAUTH_CLIENT_SECRET = 024b16270452504c35f541aca4bf78781cd06db9
FLAGS = -ldflags "-X github.com/coccyx/gogen/internal.gitHubClientID=$(GITHUB_OAUTH_CLIENT_ID) -X github.com/coccyx/gogen/internal.gitHubClientSecret=$(GITHUB_OAUTH_CLIENT_SECRET) -X main.Version=$(VERSION) -X main.GitSummary=$(SUMMARY) -X main.BuildDate=$(DATE)"
GOBIN ?= $(HOME)/go/bin
VERSION = $(shell cat $(CURDIR)/VERSION)
SUMMARY = $(shell git describe --tags --always --dirty)
DATE = $(shell date --rfc-3339=date)


.PHONY: all build deps install test docker splunkapp embed

ifeq ($(OS),Windows_NT)
	dockercmd := docker run -e TERM -e HOME=/go/src/github.com/coccyx/gogen --rm -it -v $(CURDIR):/go/src/github.com/coccyx/gogen -v $(HOME)/.ssh:/root/.ssh clintsharp/gogen bash
else
	cd := $(shell pwd)
	dockercmd := docker run --rm -it -v $(cd):/go/src/github.com/coccyx/gogen clintsharp/gogen bash
endif

all: install

build:
	GOOS=linux CGO_ENABLED=0 GOARCH=amd64 go build -tags netgo $(FLAGS) -o build/linux/gogen
	GOOS=darwin GOARCH=amd64 go build $(FLAGS) -o build/osx/gogen
	GOOS=windows GOARCH=amd64 go build $(FLAGS) -o build/windows/gogen.exe
	GOOS=wasip1 GOARCH=wasm go build $(FLAGS) -o build/wasm/gogen.wasm

deps:
	go install github.com/mattn/goveralls@latest

install:
	go install $(FLAGS)

test:
	go test -v ./...

docker:
	$(dockercmd)

