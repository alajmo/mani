NAME    := mani
PACKAGE := github.com/alajmo/$(NAME)
GIT     := $(shell git rev-parse --short HEAD)
DATE    := $(shell date +%FT%T%Z)
VERSION := v0.2.1

SRC_DIR = .
SOURCES = $(shell find $(SRC_DIR) -type f -name '*.go')
TEST_PATTERN?=.
TEST_OPTIONS?=

default: build-dev

format:
	gofmt -w -s .

lint:
	go mod tidy
	golint ./...

test: $(SOURCES)
	# go vet ./...
	# golint ./...
	# goimports ./...
	go test $(TEST_OPTIONS) -run $(TEST_PATTERN) ./...

update-golden: $(SOURCES)
	go test $(TEST_OPTIONS) -run $(TEST_PATTERN) ./... -update

test-watch: $(SOURCES)
	ag -l | entr make test

build-dev:
	go build

build:
	go build \
	-ldflags "-w -X ${PACKAGE}/cmd.version=${VERSION} -X ${PACKAGE}/cmd.commit=${GIT} -X ${PACKAGE}/cmd.date=${DATE}" \
	-a -tags netgo -o execs/${NAME} main.go

build-and-link:
	go build \
	-ldflags "-w -X ${PACKAGE}/cmd.version=${VERSION} -X ${PACKAGE}/cmd.commit=${GIT} -X ${PACKAGE}/cmd.date=${DATE}" \
	-a -tags netgo -o execs/${NAME} main.go
	cp execs/mani ~/.local/bin/mani
	./execs/mani completion > ~/workstation/scripts/completions/mani-completion.sh

.PHONY: test
