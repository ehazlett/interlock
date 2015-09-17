CGO_ENABLED=0
GOOS=linux
GOARCH=amd64
TAG=${TAG:-latest}
COMMIT=`git rev-parse --short HEAD`

all: build

clean:
	@rm -rf interlock

build:
	@godep go build -a -tags 'netgo' -ldflags "-w -X github.com/ehazlett/interlock/version.GITCOMMIT=$(COMMIT) -linkmode external -extldflags -static" .

image: build
	@echo Building Interlock image $(TAG)
	@docker build -t ehazlett/interlock:$(TAG) .

release: deps build image
	@docker push ehazlett/interlock:$(TAG)

test: clean 
	@godep go test -v ./...

.PHONY: all deps build clean image test release
