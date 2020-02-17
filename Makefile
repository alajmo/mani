NAME    := mani
PACKAGE := github.com/samiralajmovic/$(NAME)
GIT     := $(shell git rev-parse --short HEAD)
DATE    := $(shell date +%FT%T%Z)
VERSION := v0.2.1

default: help

format:
	gofmt -w -s .

lint:
	go vet ./...
	golint ./...

test:      ## Run all tests
	go clean --testcache && go test ./...

build:     ## Builds the CLI
	go build \
	-ldflags "-w -X ${PACKAGE}/cmd.version=${VERSION} -X ${PACKAGE}/cmd.commit=${GIT} -X ${PACKAGE}/cmd.date=${DATE}" \
	-a -tags netgo -o execs/${NAME} main.go

build-and-link:     ## Builds the CLI and Adds autocompletion
	go build \
	-ldflags "-w -X ${PACKAGE}/cmd.version=${VERSION} -X ${PACKAGE}/cmd.commit=${GIT} -X ${PACKAGE}/cmd.date=${DATE}" \
	-a -tags netgo -o execs/${NAME} main.go
	cp execs/mani ~/.local/bin/mani
	./execs/mani completion > ~/workstation/scripts/completions/mani-completion.sh

help:
	echo "Available commands: lint, test, build"
