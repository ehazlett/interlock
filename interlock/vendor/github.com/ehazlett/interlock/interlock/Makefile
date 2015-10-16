CGO_ENABLED=0
GOOS=linux
GOARCH=amd64
TAG?=latest
COMMIT=`git rev-parse --short HEAD`
export GO15VENDOREXPERIMENT=1

all: build

clean:
	@rm -rf interlock

build:
	@godep go build -a -tags 'netgo' -ldflags "-w -X github.com/ehazlett/interlock/version.GITCOMMIT=$(COMMIT) -linkmode external -extldflags -static" .

image: build
	@echo Building Interlock image $(TAG)
	@docker build -t ehazlett/interlock:$(TAG) .

release: build image
	@docker push ehazlett/interlock:$(TAG)

test: clean 
	@godep go test -v ./...

.PHONY: all build clean image test release
