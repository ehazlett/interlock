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
	@glide i
# this causes an import conflict with tests.  remove this vendor. le sigh
	@rm -rf vendor/github.com/docker/docker/vendor/github.com/docker/go-connections

build: build-static

build-app:
	@echo " -> Building $(TAG)$(BUILD)"
	@cd cmd/$(APP) && go build -v -ldflags "-w -X github.com/$(REPO)/version.GitCommit=$(COMMIT) -X github.com/$(REPO)/version.Build=$(BUILD)" .
	@echo "Built $$(./cmd/$(APP)/$(APP) -v)"

build-static:
	@echo " -> Building $(TAG)$(BUILD)"
	@cd cmd/$(APP) && go build -v -a -tags "netgo static_build" -installsuffix netgo -ldflags "-w -X github.com/$(REPO)/version.GitCommit=$(COMMIT) -X github.com/$(REPO)/version.Build=$(BUILD)" .
	@echo "Built $$(./cmd/$(APP)/$(APP) -v)"

image:
	@docker build --build-arg TAG=$(TAG) --build-arg BUILD=$(BUILD) -t $(REPO):$(TAG) .
	@echo "Image created: $(REPO):$(TAG)"

integration: image
	# TODO

test-integration:
	@go test -v $(TEST_ARGS) ./test/integration/...

test:
	@go test -v -cover -race $(TEST_ARGS) $$(glide novendor | grep -v ./test)

clean:
	@rm cmd/$(APP)/$(APP)

.PHONY: deps build build-static build-app build-image image clean test
