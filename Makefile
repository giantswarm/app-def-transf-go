PROJECT=app-def-transf-go
ifndef registry
registry=registry.giantswarm.io
endif

BUILD_PATH := $(shell pwd)/.gobuild

PROJECT_PATH := $(BUILD_PATH)/src/github.com/giantswarm

BIN := $(PROJECT)

.PHONY: clean deps get-deps fmt run-tests

GOPATH := $(BUILD_PATH)

ifndef GOOS
	GOOS := $(shell go env GOOS)
endif
ifndef GOARCH
	GOARCH := $(shell go env GOARCH)
endif

SOURCE=$(shell find . -name '*.go')

VERSION=$(shell cat VERSION)
COMMIT := $(shell git rev-parse --short HEAD)

all: get-deps $(BIN)

ci: clean all run-tests

clean:
	rm -rf $(BUILD_PATH) $(BIN)

get-deps: .gobuild

deps:
	@${MAKE} -B -s .gobuild

.gobuild:
	@mkdir -p $(PROJECT_PATH)
	@rm -f $(PROJECT_PATH)/$(PROJECT) && cd "$(PROJECT_PATH)" && ln -s ../../../.. $(PROJECT)

	#
	# Fetch private packages first (so `go get` skips them later)
	@builder get dep -b 0.15.0 git@github.com:giantswarm/user-config.git $(PROJECT_PATH)/user-config
	@builder get dep git@github.com:giantswarm/docker-types-go.git $(PROJECT_PATH)/docker-types-go

	#
	# Fetch public dependencies via `go get`
	GOPATH=$(GOPATH) go get -d -v github.com/giantswarm/$(PROJECT)

	#
	# Build test packages (we only want those two, so we use `-d` in go get)
	GOPATH=$(GOPATH) go get -d -v github.com/onsi/gomega
	GOPATH=$(GOPATH) go get -d -v github.com/onsi/ginkgo

$(BIN): VERSION $(SOURCE)
	echo Building for $(GOOS)/$(GOARCH)
	docker run \
	    --rm \
	    -v $(shell pwd):/usr/code \
	    -e GOPATH=/usr/code/.gobuild \
	    -e GOOS=$(GOOS) \
	    -e GOARCH=$(GOARCH) \
	    -w /usr/code \
	    golang:1.4.2-cross \
	    go build -a -ldflags "-X main.projectVersion $(VERSION) -X main.projectBuild $(COMMIT)" -o $(BIN)

run-tests:
	GOPATH=$(GOPATH) go test ./...

fmt:
	gofmt -l -w .
