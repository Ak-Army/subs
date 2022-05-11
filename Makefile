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
	${GO_EXECUTABLE} get github.com/Ak-Army/gox
	${GOPATH}/bin/gox -verbose \
		-ldflags="-X main.Version=${BUILD_VERSION} -X main.BuildTime=${BUILD_TIME}" \
		-output="build/${BUILD_NAME}-{{.OS}}-{{.Arch}}" .

full-test: static-check test

static-check:
	${GO_EXECUTABLE} get github.com/jgautheron/goconst/cmd/goconst
	${GO_EXECUTABLE} get github.com/alecthomas/gocyclo
	#${GO_EXECUTABLE} get github.com/golangci/interfacer
	${GO_EXECUTABLE} get github.com/walle/lll/cmd/lll
	${GO_EXECUTABLE} get github.com/mdempsky/unconvert
	${GO_EXECUTABLE} get mvdan.cc/unparam
	${GO_EXECUTABLE} get honnef.co/go/tools/cmd/staticcheck
	${GOPATH}/bin/staticcheck ./...
	${GOPATH}/bin/unparam -tests=false ./...
	${GOPATH}/bin/unconvert ./...
	${GOPATH}/bin/lll -g -l 140
	#interfacer ${LIST_OF_FILES}
	@test "`/usr/local/go/bin/gofmt -l -s . |wc -l`" = "0" \
		|| { echo Check fmt for files:; /usr/local/go/bin/gofmt -l -s .; exit 1; }
	${GOPATH}/bin/gocyclo -over 10 -avg .
	${GOPATH}/bin/goconst -min-occurrences 3 -min-length 3 -ignore-tests .

.PHONY: build test build-all full-test
