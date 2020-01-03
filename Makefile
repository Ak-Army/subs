GO_EXECUTABLE ?= go
BUILD_VERSION ?= $(shell git describe --tags)
GOPATH ?= ~/go
BUILD_TIME = $(shell date +%FT%T%z)
BUILD_NAME = subs
MAIN_FILE = main.go
LIST_OF_FILES = $(shell ${GO_EXECUTABLE} list ./... | grep -v /vendor/ | grep -v /src/ |grep -v /proto/)

build:
	${GO_EXECUTABLE} build \
		-o build/${BUILD_NAME} \
		-ldflags="-X main.Version=${BUILD_VERSION} -X main.BuildTime=${BUILD_TIME}" \
		.

test:
	${GO_EXECUTABLE} vet -tests ${LIST_OF_FILES}
	${GO_EXECUTABLE} test -race -cover -bench . ${LIST_OF_FILES}

build-all:
	${GO_EXECUTABLE} install github.com/Ak-Army/gox
	${GOPATH}/bin/gox -verbose \
		-ldflags="-X main.Version=${BUILD_VERSION} -X main.BuildTime=${BUILD_TIME}" \
		-output="build/${BUILD_NAME}-{{.OS}}-{{.Arch}}" .

.PHONY: build test build-all
