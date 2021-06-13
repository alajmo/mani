NAME    := mani
PACKAGE := github.com/alajmo/$(NAME)
GIT     := $(shell git rev-parse --short HEAD)
DATE    := $(shell date +%FT%T%Z)
VERSION := v0.3.0

.PHONY: lint test build build-and-link

default: build

lint:
	gofmt -w -s .
	go mod tidy
	goimports ./...

test:
	go vet ./...
	staticcheck ./...
	./test/test -verbose --count 10 --clean

# GOOS=linux GOARCH=amd64
build:
	CGO_ENABLED=0 go build \
	-ldflags "-w -X ${PACKAGE}/cmd.version=${VERSION} -X ${PACKAGE}/cmd.commit=${GIT} -X ${PACKAGE}/cmd.date=${DATE}" \
	-a -tags netgo -o execs/${NAME} main.go

build-and-link:
	go build \
	-ldflags "-w -X ${PACKAGE}/cmd.version=${VERSION} -X ${PACKAGE}/cmd.commit=${GIT} -X ${PACKAGE}/cmd.date=${DATE}" \
	-a -tags netgo -o execs/${NAME} main.go
	cp execs/mani ~/.local/bin/mani
	./execs/mani completion bash > ~/workstation/scripts/completions/mani-completion.sh

build-docker-images:
	./test/build.sh
