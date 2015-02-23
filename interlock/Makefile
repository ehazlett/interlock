CGO_ENABLED=0
GOOS=linux
GOARCH=amd64
NAME=interlock
TAG=${1:-latest}

all: deps build

deps:
	@go get -d ./...

build: deps
	@go build -a -tags 'netgo' -ldflags '-w -linkmode external -extldflags -static'

image: build
	@docker build -t ehazlett/$(NAME):$(TAG) .

.PHONY: build image deps
