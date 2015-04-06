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

get-deps: .gobuild .gobuild/bin/go-bindata

.gobuild/bin/go-bindata:
	GOOS=$(shell go env GOHOSTOS) GOPATH=$(GOPATH) go get github.com/jteeuwen/go-bindata/...

deps:
	@${MAKE} -B -s .gobuild

.gobuild:
	@mkdir -p $(PROJECT_PATH)
	@rm -f $(PROJECT_PATH)/$(PROJECT) && cd "$(PROJECT_PATH)" && ln -s ../../../.. $(PROJECT)

	#
	# Fetch private packages first (so `go get` skips them later)
	@builder get dep -b definition-v2 git@github.com:giantswarm/user-config.git $(PROJECT_PATH)/user-config
	@builder get dep -b 0.1.0 git@github.com:giantswarm/docker-types-go.git $(PROJECT_PATH)/docker-types-go
	@builder get dep -b 0.26.2 git@github.com:giantswarm/app-service $(PROJECT_PATH)/app-service

	#
	# Fetch public dependencies via `go get`
	GOPATH=$(GOPATH) go get -d -v github.com/giantswarm/$(PROJECT)

	#
	# Pin versions of certain libs
	@builder get dep -b v0.8.3 git@github.com:coreos/fleet.git $(GOPATH)/src/github.com/coreos/fleet

	#
	# Build test packages (we only want those two, so we use `-d` in go get)
	GOPATH=$(GOPATH) go get -d -v github.com/onsi/gomega
	GOPATH=$(GOPATH) go get -d -v github.com/onsi/ginkgo

$(BIN): VERSION $(SOURCE) $(PROJECT_PATH)/app-service/service/unitfile-service/templates_bindata.go
	echo Building for $(GOOS)/$(GOARCH)
	docker run \
	    --rm \
	    -v $(shell pwd):/usr/code \
	    -e GOPATH=/usr/code/.gobuild \
	    -e GOOS=$(GOOS) \
	    -e GOARCH=$(GOARCH) \
	    -w /usr/code \
	    golang:1.3.1-cross \
	    go build -a -ldflags "-X main.projectVersion $(VERSION) -X main.projectBuild $(COMMIT)" -o $(BIN)

# Special rule, because this file is generated
$(PROJECT_PATH)/app-service/service/unitfile-service/templates_bindata.go: .gobuild/bin/go-bindata $(TEMPLATES)
	.gobuild/bin/go-bindata -pkg unitfileservice -o $(PROJECT_PATH)/app-service/service/unitfile-service/templates_bindata.go $(PROJECT_PATH)/app-service/service/unitfile-service/templates/

run-tests:
	GOPATH=$(GOPATH) go test ./...

fmt:
	gofmt -l -w .
