.PHONY: install-tools lint test test-verbose format help	

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

install-tools: ## Install linting tools
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install mvdan.cc/gofumpt@latest
	go install github.com/segmentio/golines@latest

lint: ## Run golangci linter
	golangci-lint run -c ./golangci.yml ./...

format: ## Format code
	gofumpt -l -w -extra .
	golines . -w

test: ## Run tests
	go test -race -count=10 -shuffle on -cover ./...

test-verbose: ## Run tests with verbose output
	go test -race -count=10 -shuffle on -v -cover ./...

test-trimpath: ## Run tests with -trimpath
	GOFLAGS=-trimpath go test -race -count=10 -shuffle on -v -cover ./examples
