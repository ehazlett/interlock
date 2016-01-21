CGO_ENABLED=0
GOOS=linux
GOARCH=amd64
COMMIT=`git rev-parse --short HEAD`
APP=interlock
REPO?=ehazlett/$(APP)
TAG?=latest
export GO15VENDOREXPERIMENT=1

all: image

add-deps:
	@godep save -t ./...

build: build-static

build-app:
	@GO15VENDOREXPERIMENT=1 cd cmd/$(APP) && go build -ldflags "-w -X github.com/$(REPO)/version.GitCommit=$(COMMIT)" .

build-static:
	@GO15VENDOREXPERIMENT=1 cd cmd/$(APP) && go build -a -tags "netgo static_build" -installsuffix netgo -ldflags "-w -X github.com/$(REPO)/version.GitCommit=$(COMMIT)" .

test:
	@# HACK to work around "vendor" dir and go test ./... (will test every vendor package as well)
	@GO15VENDOREXPERIMENT=1 find . -maxdepth 1 -type d -not -path ./Godeps -not -path ./.git -not -path ./vendor -not -path . -exec go test -v {}/... \;

image:
	@docker build -t $(REPO):$(TAG) .
	@echo "Image created: $(REPO):$(TAG)"

clean:
	@rm cmd/$(APP)/$(APP)

.PHONY: add-deps build build-static build-app image clean

