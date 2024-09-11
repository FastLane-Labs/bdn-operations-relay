GOLANGCI_LINT_VERSION=v1.59.1

.PHONY: lint
lint: golangci-lint
	@golangci-lint run --timeout 10m0s

.PHONY: golangci-lint
golangci-lint:
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "golangci-lint not found, installing..."; \
		$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}; \
	fi