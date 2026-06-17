SHELL := /usr/bin/env bash

LOCALBIN := $(CURDIR)/bin
export PATH := $(LOCALBIN):$(PATH)
export GOBIN := $(LOCALBIN)

GO ?= go
SERVER_BIN := $(LOCALBIN)/server

GQLGEN_VERSION ?= v0.17.91
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
	@if [ -f gqlgen.yml ]; then $(GO) tool github.com/99designs/gqlgen generate; else echo "gqlgen.yml not present yet — skipping gqlgen"; fi

.PHONY: build
build: $(LOCALBIN) ## Build the server binary into ./bin
	$(GO) build -o $(SERVER_BIN) ./cmd/server

.PHONY: run
run: ## Run the server locally
	$(GO) run ./cmd/server

.PHONY: tls-certs
tls-certs: ## Generate a local self-signed TLS cert into ./certs (dev only)
	$(GO) run ./cmd/gencert

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

K8S_NS ?= petstore
K8S_IMAGE ?= petstore-api:dev
WEB_IMAGE ?= petstore-web:dev
K8S_DIR := deploy/k8s
# Must match demoStoreID in cmd/seed; the seeder pins the demo store to this id.
DEMO_STORE_ID := 11111111-1111-1111-1111-111111111111

.PHONY: k8s-up
k8s-up: ## Bring up the full stack on Minikube (single command)
	@set -euo pipefail; \
	command -v minikube >/dev/null || { echo "minikube not found"; exit 1; }; \
	command -v kubectl >/dev/null || { echo "kubectl not found"; exit 1; }; \
	minikube status >/dev/null 2>&1 || minikube start; \
	echo ">> enabling ingress addon"; \
	minikube addons enable ingress >/dev/null; \
	echo ">> building API image into minikube"; \
	minikube image build -t $(K8S_IMAGE) -f deploy/docker/Dockerfile .; \
	echo ">> building web (frontend) image into minikube"; \
	minikube image build -t $(WEB_IMAGE) frontend; \
	echo ">> applying namespace + config"; \
	kubectl apply -f $(K8S_DIR)/namespace.yaml -f $(K8S_DIR)/config.yaml; \
	echo ">> creating purpose-scoped secrets (values never committed) and migrations configmap"; \
	kubectl -n $(K8S_NS) get secret petstore-api-secret >/dev/null 2>&1 || \
	  kubectl -n $(K8S_NS) create secret generic petstore-api-secret \
	    --from-literal=DATABASE_URL='postgres://petstore:petstore@postgres:5432/petstore?sslmode=disable' \
	    --from-literal=PII_ENCRYPTION_KEY="$$(openssl rand -base64 32)"; \
	kubectl -n $(K8S_NS) get secret petstore-postgres-secret >/dev/null 2>&1 || \
	  kubectl -n $(K8S_NS) create secret generic petstore-postgres-secret \
	    --from-literal=POSTGRES_USER=petstore \
	    --from-literal=POSTGRES_PASSWORD=petstore \
	    --from-literal=POSTGRES_DB=petstore; \
	kubectl -n $(K8S_NS) get secret petstore-minio-secret >/dev/null 2>&1 || \
	  kubectl -n $(K8S_NS) create secret generic petstore-minio-secret \
	    --from-literal=MINIO_ACCESS_KEY=minioadmin \
	    --from-literal=MINIO_SECRET_KEY=minioadmin; \
	kubectl -n $(K8S_NS) get secret petstore-web-secret >/dev/null 2>&1 || \
	  kubectl -n $(K8S_NS) create secret generic petstore-web-secret \
	    --from-literal=AMBIENT_AUTH="Basic $$(printf 'customer@petstore.local:demo-password' | base64)"; \
	kubectl -n $(K8S_NS) get secret petstore-tls >/dev/null 2>&1 || { \
	  echo ">> generating TLS cert (localhost + petstore.local)"; \
	  $(GO) run ./cmd/gencert; \
	  kubectl -n $(K8S_NS) create secret tls petstore-tls --cert=certs/cert.pem --key=certs/key.pem; }; \
	kubectl -n $(K8S_NS) delete configmap petstore-migrations --ignore-not-found; \
	kubectl -n $(K8S_NS) create configmap petstore-migrations --from-file=db/migrations; \
	echo ">> starting Postgres, Redis, MinIO"; \
	kubectl apply -f $(K8S_DIR)/postgres.yaml -f $(K8S_DIR)/redis.yaml -f $(K8S_DIR)/minio.yaml; \
	kubectl -n $(K8S_NS) rollout status deploy/postgres --timeout=180s; \
	kubectl -n $(K8S_NS) rollout status deploy/minio --timeout=180s; \
	echo ">> deploying API (runs migrations first)"; \
	kubectl apply -f $(K8S_DIR)/api.yaml; \
	kubectl -n $(K8S_NS) rollout status deploy/petstore-api --timeout=180s; \
	echo ">> seeding demo accounts and catalog"; \
	kubectl -n $(K8S_NS) delete job petstore-seed --ignore-not-found; \
	kubectl apply -f $(K8S_DIR)/seed-job.yaml; \
	kubectl -n $(K8S_NS) wait --for=condition=complete job/petstore-seed --timeout=120s; \
	echo ">> deploying web storefront + ingress"; \
	kubectl apply -f $(K8S_DIR)/web.yaml -f $(K8S_DIR)/ingress.yaml; \
	kubectl -n $(K8S_NS) rollout status deploy/petstore-web --timeout=180s; \
	echo ""; \
	echo "Stack is up. Open the customer storefront:"; \
	echo "  kubectl port-forward -n $(K8S_NS) svc/petstore-web 8080:80   # leave running"; \
	echo "  then open http://localhost:8080/store/$(DEMO_STORE_ID)"; \
	echo "  (or, via ingress: add 'petstore.local' to /etc/hosts -> minikube ip, browse https://petstore.local/store/$(DEMO_STORE_ID))"; \
	echo ""; \
	echo "Direct API access (merchant ops / curl) over TLS:"; \
	echo "  kubectl port-forward -n $(K8S_NS) svc/petstore-api 8443:8443"; \
	echo "  curl -k https://localhost:8443/healthz"; \
	echo "  Demo store id: $(DEMO_STORE_ID)"

.PHONY: k8s-down
k8s-down: ## Tear down the stack (keeps the Minikube VM)
	kubectl delete namespace $(K8S_NS) --ignore-not-found

.PHONY: logs
logs: ## Tail the API logs
	kubectl -n $(K8S_NS) logs -f deploy/petstore-api -c api
