BUILDTAGS=interlock
CGO_ENABLED=0
GOOS=linux
GOARCH=amd64
COMMIT=`git rev-parse --short HEAD`
APP=interlock
REPO?=ehazlett/$(APP)
TAG?=latest
SHELL=/bin/bash

PACKAGES=$(shell go list ./... | grep -v /vendor/)

export GO15VENDOREXPERIMENT=1
export GOPATH:=$(PWD)/vendor:$(GOPATH)

all: image

deps:
	@rm -rf Godeps vendor
	@godep save ./...

build: build-static

build-app:
	@cd cmd/$(APP) && go build -ldflags "-w -X github.com/$(REPO)/version.GitCommit=$(COMMIT)" .

build-static:
	@cd cmd/$(APP) && go build -a -tags "netgo static_build" -installsuffix netgo -ldflags "-w -X github.com/$(REPO)/version.GitCommit=$(COMMIT)" .

test:
	@go test -v -cover -race ${PACKAGES}

image:
	@docker build -t $(REPO):$(TAG) .
	@echo "Image created: $(REPO):$(TAG)"

clean:
	@rm cmd/$(APP)/$(APP)

.PHONY: add-deps build build-static build-app image clean test

