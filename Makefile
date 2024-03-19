GORELEASER_RELEASE       ?= false
GORELEASER_DEBUG         ?= false
GORELEASER_IMAGE         := ghcr.io/goreleaser/goreleaser
GORELEASER_MOUNT_CONFIG  ?= false
GORELEASER_SNAPSHOT      ?= false
GORELEASER_SKIP_FLAGS    := $(GORELEASER_SKIP)
GORELEASER_SKIP          :=
GORELEASER_CONFIG        ?= .goreleaser.yaml

RELEASE_DOCKER_IMAGE     ?= ghcr.io/akash-network/e2e-test

GO_MOD_NAME              := $(shell go list -m 2>/dev/null)

null  :=
space := $(null) #
comma := ,

ifneq ($(GORELEASER_RELEASE),true)
	GITHUB_TOKEN=
	GORELEASER_SKIP_FLAGS += publish
endif

ifneq ($(GORELEASER_SKIP_FLAGS),)
	GORELEASER_SKIP := --skip=$(subst $(space),$(comma),$(strip $(GORELEASER_SKIP_FLAGS)))
endif

ifeq ($(GORELEASER_MOUNT_CONFIG),true)
	GORELEASER_IMAGE := -v $(HOME)/.docker/config.json:/root/.docker/config.json $(GORELEASER_IMAGE)
endif

.PHONY: release
release:
	docker run \
		--rm \
		-e MOD="$(GO_MOD)" \
		-e BUILD_TAGS="$(BUILD_TAGS)" \
		-e BUILD_VARS="$(GORELEASER_BUILD_VARS)" \
		-e STRIP_FLAGS="$(GORELEASER_STRIP_FLAGS)" \
		-e LINKMODE="$(GO_LINKMODE)" \
		-e GITHUB_TOKEN="$(GITHUB_TOKEN)" \
		-e GORELEASER_CURRENT_TAG="$(RELEASE_TAG)" \
		-e DOCKER_IMAGE=$(RELEASE_DOCKER_IMAGE) \
		-e GOTOOLCHAIN="$(GOTOOLCHAIN)" \
		-e GOPATH=/go \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v $(GOPATH):/go \
		-v $(shell pwd):/go/src/$(GO_MOD_NAME) \
		-w /go/src/$(GO_MOD_NAME)\
		$(GORELEASER_IMAGE) \
		-f "$(GORELEASER_CONFIG)" \
		release \
		$(GORELEASER_SKIP) \
		--debug=$(GORELEASER_DEBUG) \
		--snapshot=$(GORELEASER_SNAPSHOT) \
		--clean
