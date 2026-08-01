BINDIR      := $(CURDIR)/bin
BINNAME     := sb
SRC         := $(shell find . -type f -name '*.go' -not -path './vendor/*')
ASSETS      := $(shell find internal/rootfs/assets -type f 2>/dev/null)

GOFLAGS     :=
TAGS        :=
LDFLAGS     := -w -s
CGO_ENABLED ?= 0

ALPINE_IMAGE ?= alpine:3.23.5
ALPINE_DIR   ?= $(CURDIR)/alpinefs
ALPINE_ASSET ?= $(CURDIR)/internal/rootfs/assets
ALPINE_TAR   ?= $(ALPINE_ASSET)/$(subst :,-,$(ALPINE_IMAGE)).tar.gz

.PHONY: all
all: build ## Build the binary (default)

.PHONY: build
build: fmt vet lint tidy $(BINDIR)/$(BINNAME) ## Build the binary

$(BINDIR)/$(BINNAME): $(SRC) $(ASSETS) go.mod go.sum
	@mkdir -p '$(BINDIR)'
	CGO_ENABLED=$(CGO_ENABLED) go build $(GOFLAGS) -trimpath -tags '$(TAGS)' -ldflags '$(LDFLAGS)' -o '$@' .

.PHONY: fmt
fmt: ## Format source
	go fmt ./...

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: lint
lint: ## Run golangci-lint
	golangci-lint run ./...

.PHONY: tidy
tidy: ## Tidy go.mod/go.sum
	go mod tidy

.PHONY: test
test: ## Run tests
	go test ./...

.PHONY: clean
clean: ## Remove build artifacts
	@rm -rf '$(BINDIR)'

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: get-alpinefs
get-alpinefs: clean-alpinefs extract-alpinefs archive-alpinefs

.PHONY: clean-alpinefs
clean-alpinefs:
	@test -n '$(ALPINE_DIR)' && test '$(abspath $(ALPINE_DIR))' != '/'
	@rm -rf '$(ALPINE_DIR)'

.PHONY: extract-alpinefs
extract-alpinefs:
	@mkdir -p '$(ALPINE_DIR)'
	@container_id="$$(docker create '$(ALPINE_IMAGE)')"; \
	trap 'docker rm -f "$$container_id" >/dev/null 2>&1 || true' EXIT HUP INT TERM; \
	docker export "$$container_id" | \
		tar --extract --preserve-permissions --no-same-owner \
			--file=- --directory='$(ALPINE_DIR)'; \
	docker rm "$$container_id" >/dev/null; \
	trap - EXIT HUP INT TERM

.PHONY: archive-alpinefs
archive-alpinefs:
	@mkdir -p '$(dir $(ALPINE_TAR))'
	@archive_tmp='$(ALPINE_TAR).tmp'; \
	trap 'rm -f "$$archive_tmp"' EXIT HUP INT TERM; \
	tar --sort=name \
		--mtime='UTC 2026-01-01' \
		--owner=0 \
		--group=0 \
		--numeric-owner \
		--format=posix \
		--create \
		--file=- \
		--directory='$(ALPINE_DIR)' . | \
	gzip -n -9 > "$$archive_tmp"; \
	mv "$$archive_tmp" '$(ALPINE_TAR)'; \
	trap - EXIT HUP INT TERM
