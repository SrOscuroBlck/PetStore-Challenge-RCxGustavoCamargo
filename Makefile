SHELL := /usr/bin/env bash

LOCALBIN := $(CURDIR)/bin
export PATH := $(LOCALBIN):$(PATH)
export GOBIN := $(LOCALBIN)

GO ?= go
SERVER_BIN := $(LOCALBIN)/server

GQLGEN_VERSION ?= latest
SQLC_VERSION ?= latest
ATLAS_VERSION ?= latest
GOIMPORTS_VERSION ?= latest
GOLANGCI_VERSION ?= v2.12.2

# Load local environment overrides (e.g. DATABASE_URL) when a .env file is present.
ifneq (,$(wildcard .env))
include .env
export
endif

.PHONY: help
help: ## Show this help
	@grep -hE '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-14s\033[0m %s\n",$$1,$$2}'

$(LOCALBIN):
	@mkdir -p $(LOCALBIN)

.PHONY: tools
tools: $(LOCALBIN) ## Install developer tooling into ./bin
	$(GO) install github.com/99designs/gqlgen@$(GQLGEN_VERSION)
	$(GO) install github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION)
	@curl -fsSL "https://release.ariga.io/atlas/atlas-community-$$($(GO) env GOOS)-$$($(GO) env GOARCH)-$(ATLAS_VERSION)" -o $(LOCALBIN)/atlas && chmod +x $(LOCALBIN)/atlas
	$(GO) install golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION)
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(LOCALBIN) $(GOLANGCI_VERSION)

.PHONY: generate
generate: ## Run code generation (sqlc, gqlgen) once configured
	@if [ -f sqlc.yaml ]; then sqlc generate; else echo "sqlc.yaml not present yet — skipping sqlc"; fi
	@if [ -f gqlgen.yml ]; then gqlgen generate; else echo "gqlgen.yml not present yet — skipping gqlgen"; fi

.PHONY: build
build: $(LOCALBIN) ## Build the server binary into ./bin
	$(GO) build -o $(SERVER_BIN) ./cmd/server

.PHONY: run
run: ## Run the server locally
	$(GO) run ./cmd/server

.PHONY: fmt
fmt: ## Format code (gofmt + goimports)
	gofmt -w .
	@if command -v goimports >/dev/null 2>&1; then goimports -w .; fi

.PHONY: fmt-check
fmt-check: ## Verify all Go files are gofmt-clean
	@unformatted=$$(gofmt -l .); if [ -n "$$unformatted" ]; then echo "Not gofmt-clean:"; echo "$$unformatted"; exit 1; fi

.PHONY: vet
vet: ## Run go vet
	$(GO) vet ./...

.PHONY: lint
lint: ## Run golangci-lint
	golangci-lint run

.PHONY: test
test: ## Run tests
	$(GO) test ./...

.PHONY: test-race
test-race: ## Run tests with the race detector
	$(GO) test -race ./...

.PHONY: check
check: fmt-check vet lint build test-race ## Run the full quality gate

.PHONY: dev
dev: ## Start local dependencies (Postgres, Redis, MinIO)
	@if [ -f deploy/docker/docker-compose.yml ]; then docker compose -f deploy/docker/docker-compose.yml up -d; else echo "deploy/docker/docker-compose.yml not present yet — added in a later milestone"; fi

.PHONY: migrate-new
migrate-new: ## Generate a new versioned migration (name=...)
	@if [ -f atlas.hcl ]; then atlas migrate diff $(name) --env local --to "file://db/schema/schema.sql"; else echo "atlas.hcl not present yet — added in the data-layer milestone"; fi

.PHONY: migrate-up
migrate-up: ## Apply pending migrations
	@if [ -f atlas.hcl ]; then atlas migrate apply --env local; else echo "atlas.hcl not present yet — added in the data-layer milestone"; fi

.PHONY: migrate-lint
migrate-lint: ## Lint migrations for unsafe changes
	@if [ -f atlas.hcl ]; then atlas migrate lint --env local --latest 1; else echo "atlas.hcl not present yet — added in the data-layer milestone"; fi
