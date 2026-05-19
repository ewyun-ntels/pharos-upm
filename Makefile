ifdef GOOS
GOOS_SET='GOOS=$(GOOS)'
endif

ifdef GOARCH
GOARCH_SET='GOARCH=$(GOARCH)'
endif

ARCH := $(GOOS_SET) $(GOARCH_SET)

# SITE_MODE는 빌드 대상을 정의합니다.
# meta/site-config.json에 정의된 siteMode와 extension 매핑에 따라
# 해당되는 extension들이 빌드에 포함됩니다.
SITE_MODE ?= demo

# DOCKER INFO
CURRENT_TIME := $(shell date +%Y%m%d%H%M)
IMAGE_NAME := pharos/pharos-${SITE_MODE}-dev
TAG_NAME := local-dev-${CURRENT_TIME}-$(shell echo ${GIT_COMMIT} | cut -c 1-10)

# Versioning info for ldflags injection
GIT_COMMIT := $(shell git rev-parse --short=10 HEAD 2>/dev/null || echo unknown)
GIT_TAG := $(shell git describe --tags --abbrev=0 2>/dev/null || git rev-parse --abbrev-ref HEAD 2>/dev/null || echo develop)
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Inject version information into pkg/version variables
LDFLAGS += -X 'ntels.com/pharos/core/pkg/version.Version=$(GIT_TAG)'
LDFLAGS += -X 'ntels.com/pharos/core/pkg/version.Commit=$(GIT_COMMIT)'
LDFLAGS += -X 'ntels.com/pharos/core/pkg/version.BuildTime=$(BUILD_TIME)'

# Container registry info
# 192.168.61.145 registry는 사설 인증서를 사용하기 때문에 아래와 같이 docker config 변경 필요
# - config.json
#   {
#	  "auths": {
#		"192.168.61.145": {}
#	  }
#   }
# - daemon.json
#   {
#     "insecure-registries": [
#       "192.168.61.145"
#     ]
#   }
CONTAINER_REGISTRY_ADDR ?= 192.168.61.145

.PHONY: build
build: export BIN_PATH = ./bin
build: generate_pharos build_prepare build_pharos build_plugin

.PHONY: build_without_plugin
build_without_plugin: export BIN_PATH = ./bin
build_without_plugin: generate_pharos build_prepare build_pharos

.PHONY: docker_login
docker_login:
	docker login $(CONTAINER_REGISTRY_ADDR)

.PHONY: docker
docker: export GOOS ?= linux
docker: export GOARCH ?= amd64
docker: export ARCH = GOOS=$(GOOS) GOARCH=$(GOARCH)
docker: export DOCKER_ARCH=$(GOOS)/$(GOARCH)
docker: export BIN_PATH = ./bin
docker: docker_login
	@echo "Build $SITE_MODE image"
	docker build . -t $(CONTAINER_REGISTRY_ADDR)/$(IMAGE_NAME) --platform=$(DOCKER_ARCH) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg GIT_TAG=$(GIT_TAG) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg SITE_MODE=$(SITE_MODE)
	docker tag $(CONTAINER_REGISTRY_ADDR)/$(IMAGE_NAME):latest $(CONTAINER_REGISTRY_ADDR)/$(IMAGE_NAME):$(TAG_NAME)
	docker push $(CONTAINER_REGISTRY_ADDR)/$(IMAGE_NAME):$(TAG_NAME)
	docker push $(CONTAINER_REGISTRY_ADDR)/$(IMAGE_NAME):latest
	docker image rm $(CONTAINER_REGISTRY_ADDR)/$(IMAGE_NAME):$(TAG_NAME)
	docker image rm $(CONTAINER_REGISTRY_ADDR)/$(IMAGE_NAME):latest
	docker images -f "dangling=true" -q | xargs -r docker rmi -f

.PHONY: generate_pharos
generate_pharos:
	@echo "Generating the pharos frontend..."
	@echo "SITE_MODE: $(SITE_MODE)"
	@env SITE_MODE=$(SITE_MODE) go generate ./core/frontend
	@echo "Done"

.PHONY: build_prepare
build_prepare:
	@echo "=== Build Preparation ==="
	@echo "SITE_MODE: $(SITE_MODE)"
	@go work sync
	@go mod download
	@env SITE_MODE=$(SITE_MODE) go run meta/scripts/generate-metadata.go
	@go generate ./core/
	@echo "✓ Build preparation completed"

.PHONY: build_pharos
build_pharos: build_prepare
	@echo "Building the pharos ARCHITECTURE($(ARCH)) binary..."
	@env CGO_ENABLED=0 $(ARCH) go build -ldflags "$(LDFLAGS)" -o $(BIN_PATH)/pharos ./core/cmd/pharos
	@echo "Done"

.PHONY: build_plugin
build_plugin: export BIN_PATH ?= ./bin
build_plugin: build_plugin_altibase build_plugin_clickhouse build_plugin_http-receiver build_plugin_postgresql

.PHONY: build_plugin_altibase
build_plugin_altibase: export BIN_PATH ?= ./bin
build_plugin_altibase: export CGO_CFLAGS_ODBC=-I$(CURDIR)/core/config/shared-library/unixODBC/el9/2.3.9-4/include/
build_plugin_altibase: export CGO_LDFLAGS_ODBC=-L$(CURDIR)/core/config/shared-library/unixODBC/el9/2.3.9-4/lib64/
build_plugin_altibase: build_prepare
	@echo "Building the altibase ARCHITECTURE($(ARCH)) binary..."
	@env CGO_ENABLED=1 CGO_CFLAGS="$(CGO_CFLAGS_ODBC)" CGO_LDFLAGS="$(CGO_LDFLAGS_ODBC)" $(ARCH) go build \
		 -tags odbc \
		 -ldflags "$(LDFLAGS)" \
		 -o $(BIN_PATH)/plugins/altibase ./core/pkg/plugins/plugin/datasource/altibase/main/main.go
	@echo "Done"

.PHONY: build_plugin_clickhouse
build_plugin_clickhouse: export BIN_PATH ?= ./bin
build_plugin_clickhouse: build_prepare
	@echo "Building the clickhouse ARCHITECTURE($(ARCH)) binary..."
	@env CGO_ENABLED=0 $(ARCH) go build \
		 -ldflags "$(LDFLAGS)" \
		 -o $(BIN_PATH)/plugins/clickhouse ./core/pkg/plugins/plugin/datasource/clickhouse/main/main.go
	@echo "Done"

.PHONY: build_plugin_http-receiver
build_plugin_http-receiver: export BIN_PATH ?= ./bin
build_plugin_http-receiver: build_prepare
	@echo "Building the http-receiver ARCHITECTURE($(ARCH)) binary..."
	@env CGO_ENABLED=0 $(ARCH) go build \
		 -ldflags "$(LDFLAGS)" \
		 -o $(BIN_PATH)/plugins/http-receiver ./core/pkg/plugins/plugin/datasource/http-receiver/main/main.go
	@echo "Done"

.PHONY: build_plugin_postgresql
build_plugin_postgresql: export BIN_PATH ?= ./bin
build_plugin_postgresql: build_prepare
	@echo "Building the postgresql ARCHITECTURE($(ARCH)) binary..."
	@env CGO_ENABLED=0 $(ARCH) go build \
		 -ldflags "$(LDFLAGS)" \
		 -o $(BIN_PATH)/plugins/postgresql ./core/pkg/plugins/plugin/datasource/postgresql/main/main.go
	@echo "Done"

.PHONY: pack_build_pharos
pack_build_pharos:
	@echo "pharos packing..."
	@upx --best --lzma $(BIN_PATH)/pharos
	@echo "Done"

.PHONY: pack_build_plugin
pack_build_plugin:
	@echo "pharos plugin packing..."
	@upx --best --lzma $(BIN_PATH)/plugins/*
	@echo "Done"
