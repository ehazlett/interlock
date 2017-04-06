BUILDTAGS=interlock
CGO_ENABLED=0
GOOS=linux
GOARCH=amd64
COMMIT=`git rev-parse --short HEAD`
APP=interlock
REPO?=ehazlett/$(APP)
TAG?=latest
BUILD?=-dev

all: clean-image image

deps:
	@glide i
# this causes an import conflict with tests.  remove this vendor. le sigh
	@rm -rf vendor/github.com/docker/docker/vendor/github.com/docker/go-connections

build: build-static

build-app:
	@cd cmd/$(APP) && go build -v -ldflags "-w -X github.com/$(REPO)/version.GitCommit=$(COMMIT) -X github.com/$(REPO)/version.Build=$(BUILD)" .
	@echo "Built $$(./cmd/$(APP)/$(APP) -v)"

build-static:
	@cd cmd/$(APP) && go build -v -a -tags "netgo static_build" -installsuffix netgo -ldflags "-w -X github.com/$(REPO)/version.GitCommit=$(COMMIT) -X github.com/$(REPO)/version.Build=$(BUILD)" .
	@echo "Built $$(./cmd/$(APP)/$(APP) -v)"

build-image: clean-image
	@echo "Building image with $$(./cmd/$(APP)/$(APP) -v)"
	@docker build $(BUILD_ARGS) -t $(REPO):$(TAG) .
	@echo "Image created: $(REPO):$(TAG)"

container:
	@docker build $(BUILD_ARGS) -t interlock-build -f Dockerfile.build .

build-in-container: container
	@docker run -it -e BUILD -e TAG --name interlock-build -ti interlock-build make build
	@docker cp interlock-build:/go/src/github.com/$(REPO)/cmd/$(APP)/$(APP) ./cmd/$(APP)/$(APP)

build-container: build-in-container clean-image

integration: container

test-integration:
	@go test -v -cover -race -tags integration $$(glide novendor)

test:
	@go test -v -cover -race $(TEST_ARGS) $$(glide novendor)

image: build-container build-image

clean-image:
	@docker rm -fv interlock-build interlock-test >/dev/null 2>&1 || exit 0

clean:
	@rm cmd/$(APP)/$(APP)

.PHONY: deps build build-static build-app build-image image clean test
