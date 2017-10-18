APP           := lostromos
PKG           := github.com/wpengine/lostromos

BUILD_NUMBER  := dev
VERSION       := $(shell git describe --tags --always --dirty)
GIT_HASH      := $(shell git rev-parse --short HEAD)
BUILD_TIME    := $(shell date -u '+%Y-%m-%d_%I:%M:%S%p')
LD_FLAGS      := -s -w -X $(PKG)/version.Version=$(VERSION) -X $(PKG)/version.GitHash=$(GIT_HASH) -X $(PKG)/version.BuildTime=$(BUILD_TIME)

MARKDOWN_LINTER := wpengine/mdl

all: lint test build

build:
	@echo Building...
	@CGO_ENABLED=0 go build -ldflags "$(LD_FLAGS)" -o lostromos main.go

test: | vendor
	@echo Testing...
	@go test ./... -cover

lint: golint lint-markdown

golint: | vendor
	@echo Linting...
	@gometalinter --enable=gofmt --vendor -D gotype

lint-markdown:
	@find . -path ./vendor -prune -o -name "*.md" -exec docker run --rm -v `pwd`/{}:/workspace/{} ${MARKDOWN_LINTER} /workspace/{} \;

install-deps:
	go get -u github.com/golang/dep/cmd/dep
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install

vendor:
	@echo Vendoring...
	@dep ensure

pull-linters:
	docker pull ${MARKDOWN_LINTER}

clean:
	@echo Cleaning...
	@rm -rf ./vendor/