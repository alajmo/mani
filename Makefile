NAME    := mani
PACKAGE := github.com/alajmo/$(NAME)
DATE    := $(shell date +%FT%T%Z)
GIT     := $(shell [ -d .git ] && git rev-parse --short HEAD)
VERSION := v0.3.0

default: build

lint:
	golangci-lint run

test:
	golangci-lint run
	./test/scripts/test --build --count 5 --clean

interactive:
	./test/scripts/exec

build:
	CGO_ENABLED=0 go build \
	-ldflags "-w -X ${PACKAGE}/cmd.version=${VERSION} -X ${PACKAGE}/cmd.commit=${GIT} -X ${PACKAGE}/cmd.date=${DATE}" \
	-a -tags netgo -o dist/${NAME} main.go

build-all:
	goreleaser --rm-dist --snapshot

build-test:
	CGO_ENABLED=0 go build -a -tags netgo -o dist/${NAME} main.go

release:
	git tag ${VERSION} && git push origin ${VERSION}

clean:
	$(RM) -r dist target

.PHONY: lint test interactive build build-all build-test release clean
