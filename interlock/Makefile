CGO_ENABLED=0
GOOS=linux
GOARCH=amd64
TAG=${TAG:-latest}

all: deps build

deps:
	@godep restore

clean:
	@rm -rf Godeps/_workspace interlock

build:
	@godep go build -a -tags 'netgo' -ldflags '-w -linkmode external -extldflags -static' .

image: build
	@echo Building Interlock image $(TAG)
	@docker build -t ehazlett/interlock:$(TAG) .

release: deps build image
	@docker push ehazlett/interlock:$(TAG)

test: clean 
	@godep go test -v ./...

.PHONY: all deps build clean image test release
