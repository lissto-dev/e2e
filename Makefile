# E2E Testing Makefile

# Go parameters
GINKGO ?= go run github.com/onsi/ginkgo/v2/ginkgo
GO ?= go

# Cluster parameters
CLUSTER_NAME ?= lissto-e2e
K3D_VERSION ?= v5.7.5

# Version parameters (can be overridden)
CLI_REF ?= latest
API_TAG ?= main
CONTROLLER_TAG ?= main
HELM_REF ?= main
USE_HELM_VERSIONS ?= false

# Test parameters
TEST_TIMEOUT ?= 30m
JUNIT_REPORT ?= e2e-results.xml

.PHONY: help
help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

##@ Dependencies

.PHONY: deps
deps: ## Install Go dependencies
	$(GO) mod download
	$(GO) mod tidy

.PHONY: install-ginkgo
install-ginkgo: ## Install Ginkgo CLI
	$(GO) install github.com/onsi/ginkgo/v2/ginkgo@latest

.PHONY: install-k3d
install-k3d: ## Install k3d
	curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash

##@ Cluster Management

.PHONY: cluster-create
cluster-create: ## Create k3d cluster
	./scripts/setup-cluster.sh

.PHONY: cluster-delete
cluster-delete: ## Delete k3d cluster
	./scripts/teardown-cluster.sh

.PHONY: cluster-status
cluster-status: ## Check cluster status
	@k3d cluster list
	@echo ""
	@kubectl cluster-info 2>/dev/null || echo "Cluster not accessible"

##@ Deployment

.PHONY: deploy
deploy: ## Deploy Lissto via Helm
	API_TAG=$(API_TAG) CONTROLLER_TAG=$(CONTROLLER_TAG) HELM_REF=$(HELM_REF) USE_HELM_VERSIONS=$(USE_HELM_VERSIONS) ./scripts/deploy-lissto.sh

.PHONY: download-cli
download-cli: ## Download Lissto CLI
	./scripts/download-cli.sh $(CLI_REF)

.PHONY: setup-cli
setup-cli: ## Configure CLI contexts (admin + user)
	./scripts/setup-cli-contexts.sh

.PHONY: wait-ready
wait-ready: ## Wait for all components to be ready
	./scripts/wait-ready.sh

##@ Testing

.PHONY: test
test: ## Run all e2e tests
	cd tests && $(GINKGO) -v --timeout=$(TEST_TIMEOUT) --junit-report=../$(JUNIT_REPORT) ./...

.PHONY: test-focus
test-focus: ## Run tests matching FOCUS pattern (e.g., make test-focus FOCUS="Blueprint")
	cd tests && $(GINKGO) -v --timeout=$(TEST_TIMEOUT) --focus="$(FOCUS)" ./...

.PHONY: test-dry-run
test-dry-run: ## List all tests without running them
	cd tests && $(GINKGO) --dry-run -v ./...

##@ Full E2E Flow

.PHONY: e2e
e2e: cluster-create deploy download-cli setup-cli wait-ready test ## Run full e2e test suite (cluster + deploy + test)
	@echo "E2E tests completed!"

.PHONY: e2e-clean
e2e-clean: cluster-delete ## Clean up e2e environment
	@echo "E2E environment cleaned up"

##@ Development

.PHONY: lint
lint: ## Run linters
	$(GO) vet ./...
	@which golangci-lint > /dev/null && golangci-lint run ./... || echo "golangci-lint not installed, skipping"

.PHONY: fmt
fmt: ## Format code
	$(GO) fmt ./...

.PHONY: verify
verify: fmt lint ## Verify code (format + lint)

##@ CI Helpers

.PHONY: ci-setup
ci-setup: install-k3d deps ## Setup CI environment
	@echo "CI environment ready"

.PHONY: ci-test
ci-test: cluster-create deploy download-cli setup-cli wait-ready ## Setup and run tests for CI
	cd tests && $(GINKGO) -v --timeout=$(TEST_TIMEOUT) --junit-report=../$(JUNIT_REPORT) ./... || (make cluster-delete && exit 1)
	make cluster-delete
