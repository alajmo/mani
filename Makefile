NAME    := mani
PACKAGE := github.com/alajmo/$(NAME)
GIT     := $(shell git rev-parse --short HEAD)
DATE    := $(shell date +%FT%T%Z)
VERSION := v0.2.1

SRC_DIR = .
SOURCES = $(shell find $(SRC_DIR) -type f -name '*.go')
TEST_PATTERN?=.
TEST_FILES?="./..."
TEST_OPTIONS?=

.PHONY: lint test test-debug test-update test-update-debug build build-and-link build-test

lint:
	gofmt -w -s .
	go mod tidy
	goimports ./...

test: $(SOURCES)
	go vet ./...
	staticcheck ./...
	./test/test --run -verbose ./...

test-debug: $(SOURCES)
	go test $(TEST_OPTIONS) -run $(TEST_PATTERN) ./test/integration/main_test.go $(TEST_FILES) -dirty -verbose

test-update: $(SOURCES)
	go test $(TEST_OPTIONS) -run $(TEST_PATTERN) ./test/integration/main_test.go ./test/integration/info_test.go -update

test-update-debug: $(SOURCES)
	go test $(TEST_OPTIONS) -run $(TEST_PATTERN) ./test/integration/main_test.go $(TEST_FILES) -update -verbose

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

build-test:
	go build

build-docker-images:
	./test/build.sh
