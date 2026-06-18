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

# Load test (k6) knobs — defaults encode the challenge target: 1k concurrent users.
LOAD_VUS ?= 1000
LOAD_DURATION ?= 30s

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

.PHONY: k8s-redeploy-api
k8s-redeploy-api: ## Rebuild only the API image and restart its deployment (fast update, no full k8s-up)
	minikube image build -t $(K8S_IMAGE) -f deploy/docker/Dockerfile .
	kubectl -n $(K8S_NS) rollout restart deploy/petstore-api
	kubectl -n $(K8S_NS) rollout status deploy/petstore-api --timeout=180s

.PHONY: k8s-redeploy-web
k8s-redeploy-web: ## Rebuild only the web (frontend) image and restart its deployment (fast update, no full k8s-up)
	minikube image build -t $(WEB_IMAGE) frontend
	kubectl -n $(K8S_NS) rollout restart deploy/petstore-web
	kubectl -n $(K8S_NS) rollout status deploy/petstore-web --timeout=180s

.PHONY: web-forward
web-forward: ## Port-forward the storefront to http://localhost:8080, auto-reconnecting across pod restarts (leave running)
	@echo ">> Storefront: http://localhost:8080/store/$(DEMO_STORE_ID)  (Ctrl-C to stop)"; \
	while true; do \
	  kubectl -n $(K8S_NS) port-forward svc/petstore-web 8080:80 || true; \
	  echo ">> web port-forward dropped (pod restart?) — reconnecting in 2s…"; sleep 2; \
	done

.PHONY: api-forward
api-forward: ## Port-forward the API to https://localhost:8443, auto-reconnecting across pod restarts (leave running)
	@echo ">> API: https://localhost:8443/playground  (Ctrl-C to stop)"; \
	while true; do \
	  kubectl -n $(K8S_NS) port-forward svc/petstore-api 8443:8443 || true; \
	  echo ">> api port-forward dropped (pod restart?) — reconnecting in 2s…"; sleep 2; \
	done

.PHONY: db-forward
db-forward: ## Port-forward Postgres to localhost:5440, auto-reconnecting across pod restarts (leave running)
	@echo ">> Postgres: localhost:5440 (petstore/petstore/petstore)  (Ctrl-C to stop)"; \
	while true; do \
	  kubectl -n $(K8S_NS) port-forward svc/postgres 5440:5432 || true; \
	  echo ">> db port-forward dropped (pod restart?) — reconnecting in 2s…"; sleep 2; \
	done

.PHONY: k8s-down
k8s-down: ## Tear down the stack (keeps the Minikube VM)
	kubectl delete namespace $(K8S_NS) --ignore-not-found

.PHONY: logs
logs: ## Tail the API logs
	kubectl -n $(K8S_NS) logs -f deploy/petstore-api -c api

.PHONY: load-test
load-test: ## Prove the <2s / 1k-user target (frontend + backend): run the k6 storefront load test in-cluster (needs make k8s-up)
	@set -euo pipefail; \
	command -v kubectl >/dev/null || { echo "kubectl not found"; exit 1; }; \
	kubectl -n $(K8S_NS) get svc petstore-web >/dev/null 2>&1 || { echo "stack not up — run 'make k8s-up' first"; exit 1; }; \
	echo ">> (re)creating k6 script configmap from loadtest/storefront.js"; \
	kubectl -n $(K8S_NS) delete configmap petstore-loadtest-script --ignore-not-found >/dev/null; \
	kubectl -n $(K8S_NS) create configmap petstore-loadtest-script --from-file=storefront.js=loadtest/storefront.js >/dev/null; \
	echo ">> launching k6 job (VUS=$(LOAD_VUS), DURATION=$(LOAD_DURATION)) against the storefront gateway"; \
	kubectl -n $(K8S_NS) delete job petstore-loadtest --ignore-not-found >/dev/null; \
	sed -e 's/__VUS__/$(LOAD_VUS)/' -e 's/__DURATION__/$(LOAD_DURATION)/' $(K8S_DIR)/loadtest-job.yaml | kubectl apply -f - >/dev/null; \
	echo ">> waiting for the k6 pod to start…"; \
	for i in $$(seq 1 60); do \
	  phase=$$(kubectl -n $(K8S_NS) get pods -l app=petstore-loadtest -o jsonpath='{.items[0].status.phase}' 2>/dev/null || true); \
	  case "$$phase" in Running|Succeeded|Failed) break;; esac; \
	  sleep 2; \
	done; \
	echo ">> streaming k6 output:"; echo; \
	kubectl -n $(K8S_NS) logs -f job/petstore-loadtest; \
	echo; \
	if kubectl -n $(K8S_NS) wait --for=condition=complete job/petstore-loadtest --timeout=10s >/dev/null 2>&1; then \
	  echo "PASS — all thresholds met (p95 < 2s, errors < 1%)."; \
	else \
	  echo "FAIL — k6 exited non-zero (a threshold was breached or the run errored). See output above."; exit 1; \
	fi
