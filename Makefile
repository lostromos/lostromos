# Copyright 2017 the lostromos Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

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

coverage: | vendor
	@echo Generating coverage report...
	@./test-scripts/coverage.sh

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

travis: lint test build coverage

vendor:
	@echo Vendoring...
	@dep ensure

pull-linters:
	docker pull ${MARKDOWN_LINTER}

clean:
	@echo Cleaning...
	@rm -rf ./vendor/
