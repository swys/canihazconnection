PKGS := $(shell go list ./... | grep -v /vendor)
GOOS := linux
GOARC := amd64
BUILD_IMG := golang:latest
BIN_NAME := canihazconnection
BIN_DIR := $(shell go env GOPATH)/bin
GOLANGCLILINT := $(BIN_DIR)/golangci-lint
TOOLS_DIR := $(CURDIR)/tools
DOCKER_REG := ###### ADD YOUR REGISTRY HERE ######
IMG_PATH := ## YOUR IMAGE PATH ##
IMG_NAME := canihazconnection
IMG_VERSION ?= $(shell git describe --tags --always --match=v* 2> /dev/null || echo v1.0.0)

# docker is required so check for it
ifeq (, $(shell which docker))
 $(error "docker was not found in your PATH, docker is required to run this build")
 endif

.PHONY: all
all: clean lint test build

.PHONY: clean
clean:
	rm -rf tools

tools/golangci-lint: tools/go.mod
	cd $(TOOLS_DIR) && GO111MODULE=on go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.17.1

tools/go.mod:
	@mkdir -p tools
	@rm -f $@
	cd $(TOOLS_DIR) && GO111MODULE=on go mod init local-tools

.PHONY: lint
lint: tools/golangci-lint
	$(GOLANGCLILINT) run -v

.PHONY: test
test:
	export TELNET_HOSTS=localhost:8888;export HTTP_REQUESTS="https://google.com";export LOG_LEVEL=DEBUG;go test -v
	@TELNET_HOSTS="";HTTP_REQUESTS="";LOG_LEVEL=""

build: lint test
	docker build -t $(DOCKER_REG)/$(IMG_PATH)/$(IMG_NAME):$(IMG_VERSION) --no-cache .

.PHONY: dockerpush
dockerpush:
	docker push $(DOCKER_REG)/$(IMG_PATH)/$(IMG_NAME):$(IMG_VERSION)