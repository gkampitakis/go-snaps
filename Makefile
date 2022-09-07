.PHONY: install-tools lint test test-verbose format

install-tools:
	# Install linting tools
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.49.0
	go install mvdan.cc/gofumpt@latest
	go install github.com/segmentio/golines@latest

lint:
	golangci-lint run -c ./golangci.yml ./...

format:
	gofumpt -l -w -extra .
	golines . -w

test:
	go test -race -count=1 -cover ./...

test-verbose:
	go test -race -count=1 -v -cover ./...
