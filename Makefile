NAME    := mani
PACKAGE := github.com/alajmo/$(NAME)
DATE    := $(shell date +"%Y %B %d")
GIT     := $(shell [ -d .git ] && git rev-parse --short HEAD)
VERSION := v0.30.0

default: build

tidy:
	go get -u && go mod tidy

gofmt:
	go fmt ./cmd/***.go
	go fmt ./core/***.go
	go fmt ./core/dao/***.go
	go fmt ./core/exec/***.go
	go fmt ./core/print/***.go
	go fmt ./core/tui/***.go
	go fmt ./test/integration/***.go

lint:
	golangci-lint run ./cmd/... ./core/...
	deadcode .

test:
	# Unit tests
	go test -v ./core/dao/***

	# Integration tests
	./test/scripts/test --build --count 5 --clean

test-unit:
	go test -v ./core/dao/***

test-integration:
	./test/scripts/test --count 5 --build --clean

update-golden-files:
	./test/scripts/test --build --update

build:
	CGO_ENABLED=0 go build \
	-ldflags "-w -X '${PACKAGE}/cmd.version=${VERSION}' -X '${PACKAGE}/core/tui.version=${VERSION}' -X '${PACKAGE}/cmd.commit=${GIT}' -X '${PACKAGE}/cmd.date=${DATE}'" \
	-a -tags netgo -o dist/${NAME} main.go

build-all:
	goreleaser release --skip-publish --rm-dist --snapshot

build-test:
	CGO_ENABLED=0 go build \
	-ldflags "-X '${PACKAGE}/core/dao.build_mode=TEST'" \
	-a -tags netgo -o dist/${NAME} main.go

gen-man:
	go run -ldflags="-X 'github.com/alajmo/mani/cmd.buildMode=man' -X '${PACKAGE}/cmd.version=${VERSION}' -X '${PACKAGE}/cmd.commit=${GIT}' -X '${PACKAGE}/cmd.date=${DATE}'" ./main.go gen-docs

release:
	git tag ${VERSION} && git push origin ${VERSION}

clean:
	$(RM) -r dist target

.PHONY: tidy gofmt lint test test-unit test-integration update-golden-files build build-all build-test gen-man release clean
