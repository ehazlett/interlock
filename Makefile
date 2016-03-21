BUILDTAGS=interlock
CGO_ENABLED=0
GOOS=linux
GOARCH=amd64
COMMIT=`git rev-parse --short HEAD`
APP=interlock
REPO?=ehazlett/$(APP)
TAG?=latest
SHELL=/bin/bash
BUILD?=-dev

PACKAGES=$(shell go list ./... | grep -v /vendor/)

export GO15VENDOREXPERIMENT=1
export GOPATH:=$(PWD)/vendor:$(GOPATH)

all: image

deps:
	@rm -rf Godeps vendor
	@godep save ./...

build: build-static

build-app:
	@cd cmd/$(APP) && go build -ldflags "-w -X github.com/$(REPO)/version.GitCommit=$(COMMIT) -X github.com/$(REPO)/version.Build=$(BUILD)" .
	@echo "Built $$(./cmd/$(APP)/$(APP) -v)"

build-static:
	@cd cmd/$(APP) && go build -a -tags "netgo static_build" -installsuffix netgo -ldflags "-w -X github.com/$(REPO)/version.GitCommit=$(COMMIT) -X github.com/$(REPO)/version.Build=$(BUILD)" .
	@echo "Built $$(./cmd/$(APP)/$(APP) -v)"

build-image:
	@echo "Building image with $$(./cmd/$(APP)/$(APP) -v)"
	@docker build -t $(REPO):$(TAG) .
	@echo "Image created: $(REPO):$(TAG)"

test:
	@go test -v -cover -race ${PACKAGES}

image: build build-image

clean:
	@rm cmd/$(APP)/$(APP)

.PHONY: add-deps build build-static build-app build-image image clean test

