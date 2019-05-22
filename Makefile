CURRENT_DIR=$(shell pwd)
PKG=github.com/dcos/dcos-core-cli
PKG_DIR=/go/src/$(PKG)
IMAGE_NAME=dcos/dcos-core-cli

PLATFORM?=$(shell uname | tr [A-Z] [a-z])
windows_EXE=.exe

export GOFLAGS := -mod=vendor

.PHONY: default
default: $(PLATFORM)

.PHONY: darwin linux windows
darwin linux windows: docker-image
	$(call inDocker,env GOOS=$(@) GO111MODULE=on go build \
		-tags '$(GO_BUILD_TAGS)' -mod=vendor \
		-o build/$(@)/dcos$($(@)_EXE) ./cmd/dcos)

.PHONY: install
install:
	@make $(PLATFORM)
	@make plugin
	dcos plugin add -u ./build/$(PLATFORM)/dcos-core-cli.zip

.PHONY: plugin
plugin: python
	@python3 scripts/plugin/package_plugin.py

.PHONY: python
python:
	@cd python/lib/dcoscli; \
		make binary

.PHONY: test
test: vet
	$(call inDocker,go test -race -cover ./...)

.PHONY: vet
vet: lint
	$(call inDocker,go vet ./...)

.PHONY: lint
lint: docker-image
	# Can be simplified once https://github.com/golang/lint/issues/320 is fixed.
	$(call inDocker,golint -set_exit_status ./cmd/... ./pkg/...)

.PHONY: generate
generate: docker-image
	$(call inDocker,go generate ./...)

.PHONY: vendor
vendor: docker-image
	$(call inDocker,go mod vendor)

.PHONY: docker-image
docker-image:
ifndef NO_DOCKER
	docker build -t $(IMAGE_NAME) .
endif

.PHONY: clean
clean:
	rm -rf build

ifdef NO_DOCKER
  define inDocker
    $1
  endef
else
  define inDocker
    docker run \
      -v $(CURRENT_DIR):$(PKG_DIR) \
      -w $(PKG_DIR) \
      --rm \
      $(IMAGE_NAME) \
    bash -c "$1"
  endef
endif
