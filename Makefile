## gosqlite – top-level build helper
#
# This Makefile keeps the project completely self-contained and able to build
# in a hermetic, air-gapped environment (no network access).  All commands rely
# solely on vendored modules or the Go standard tool-chain already installed.
#
# Targets:
#   make build            – Compile all packages and the CLI binary.
#   make test             – Run unit + fuzz-style tests with the race detector.
#   make lint             – Run static analysis (`go vet`).
#   make vendor-refresh   – Re-vendor Go dependencies (should normally be a no-op
#                           because the project avoids non-stdlib imports).
#   make run ARGS="..."   – Build and run the CLI with optional arguments.
#
# The GOFLAGS include `-mod=vendor` to ensure the build never reaches out to the
# network even if future third-party imports are introduced by accident.

GO           ?= go
CMD_DIR       = lite/gosqlite
MAIN_PKG      = $(CMD_DIR)
BIN_NAME      = gosqlite
GOFLAGS      += -mod=vendor

.PHONY: build
build: vendor ## Compile all packages.
	$(GO) build $(GOFLAGS) ./...

.PHONY: test
test: vendor ## Run tests (+ race detector).
	$(GO) test $(GOFLAGS) -race ./...

.PHONY: lint
lint: ## Static analysis – vet.
	$(GO) vet ./...

.PHONY: vendor-refresh vendor
vendor-refresh vendor: ## Vendor Go dependencies so that builds never need network.
	$(GO) mod vendor

.PHONY: run
run: build ## Build (if needed) and run the CLI. Pass ARGS="..." to forward args.
	$(GO) run $(GOFLAGS) $(MAIN_PKG) $(ARGS)

.PHONY: help
help: ## Show this help.
	@grep -E '^[a-zA-Z_-]+:.*?##' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
