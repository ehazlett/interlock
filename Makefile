BUILDTAGS=interlock
CGO_ENABLED=0
GOOS=linux
GOARCH=amd64
COMMIT=`git rev-parse --short HEAD`
APP=interlock
REPO?=ehazlett/$(APP)
TAG?=latest
BUILD?=-dev

all: image

deps:
	@glide install

build: build-static

build-app:
	@cd cmd/$(APP) && go build -ldflags "-w -X github.com/$(REPO)/version.GitCommit=$(COMMIT) -X github.com/$(REPO)/version.Build=$(BUILD)" .
	@echo "Built $$(./cmd/$(APP)/$(APP) -v)"

build-static:
	@cd cmd/$(APP) && go build -a -tags "netgo static_build" -installsuffix netgo -ldflags "-w -X github.com/$(REPO)/version.GitCommit=$(COMMIT) -X github.com/$(REPO)/version.Build=$(BUILD)" .
	@echo "Built $$(./cmd/$(APP)/$(APP) -v)"

build-image:
	@echo "Building image with $$(./cmd/$(APP)/$(APP) -v)"
	@docker build $(BUILD_ARGS) -t $(REPO):$(TAG) .
	@echo "Image created: $(REPO):$(TAG)"

build-container:
	@docker build $(BUILD_ARGS) -t interlock-build -f Dockerfile.build .
	@docker run -it -e BUILD -e TAG --name interlock-build -ti interlock-build make deps build
	@docker cp interlock-build:/go/src/github.com/$(REPO)/cmd/$(APP)/$(APP) ./cmd/$(APP)/$(APP)
	@docker rm -fv interlock-build

test:
	@go test -v -cover -race `go list ./... | grep -v /vendor/`

image: build-container build-image

clean:
	@rm cmd/$(APP)/$(APP)

.PHONY: deps build build-static build-app build-image image clean test
