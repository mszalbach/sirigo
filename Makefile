.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help.
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##//'

.PHONY: build
build: ## Build the sirigo binary.
	@go build -v -o bin/sirigo ./cmd/client/...

.PHONY: test
test: ## Run the tests.
	@go test ./...

.PHONY: lint
lint: ## Run the linter.
	@golangci-lint run

.PHONY: fmt
fmt: ## Format the code.
	@golangci-lint fmt

.PHONY: clean
clean: ## Clean the build artifacts.
	rm -rf bin/sirigo