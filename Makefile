NAME    := mani
PACKAGE := github.com/alajmo/$(NAME)
DATE    := $(shell date +"%Y %B %d")
GIT     := $(shell [ -d .git ] && git rev-parse --short HEAD)
VERSION := v0.20.0

default: build

tidy:
	go get -u && go mod tidy

gofmt:
	go fmt ./cmd/***.go
	go fmt ./core/***.go
	go fmt ./core/dao/***.go
	go fmt ./core/exec/***.go
	go fmt ./core/print/***.go

lint:
	golangci-lint run ./cmd/... ./core/...

test:
	golangci-lint run
	./test/scripts/test --build --count 1 --clean

update-golden-files:
	./test/scripts/test --build --clean --update

build:
	CGO_ENABLED=0 go build \
	-ldflags "-w -X '${PACKAGE}/cmd.version=${VERSION}' -X '${PACKAGE}/cmd.commit=${GIT}' -X '${PACKAGE}/cmd.date=${DATE}'" \
	-a -tags netgo -o dist/${NAME} main.go

build-all:
	goreleaser release --skip-publish --rm-dist --snapshot

build-test:
	CGO_ENABLED=0 go build \
	-ldflags "-X '${PACKAGE}/core/dao.build_mode=TEST'" \
	-a -tags netgo -o dist/${NAME} main.go

gen-man:
	go run -ldflags="-X 'github.com/alajmo/mani/cmd.buildMode=man'" ./main.go gen-docs

release:
	git tag ${VERSION} && git push origin ${VERSION}

clean:
	$(RM) -r dist target

.PHONY: tidy gofmt lint test update-golden-files build build-all build-test gen-man release clean
