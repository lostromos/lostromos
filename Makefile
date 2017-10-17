APP           := lostromos
PKG           := github.com/wpengine/lostromos

BUILD_NUMBER  := dev
VERSION       := $(shell git describe --tags --always --dirty)
GIT_HASH      := $(shell git rev-parse --short HEAD)
BUILD_TIME    := $(shell date -u '+%Y-%m-%d_%I:%M:%S%p')
LD_FLAGS      := -s -w -X $(PKG)/version.Version=$(VERSION) -X $(PKG)/version.GitHash=$(GIT_HASH) -X $(PKG)/version.BuildTime=$(BUILD_TIME)

all: lint test build

build:
	@echo Building...
	@CGO_ENABLED=0 go build -ldflags "$(LD_FLAGS)" -o lostromos main.go

test: | vendor
	@echo Testing...
	@go test ./... -cover

lint: | vendor
	@echo Linting...
	@gometalinter --enable=gofmt --vendor -D gotype

vendor:
	@echo Vendoring...
	@dep ensure

clean:
	@echo Cleaning...
	@rm -rf ./vendor/
