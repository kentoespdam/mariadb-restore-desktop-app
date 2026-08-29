.PHONY: lint format test build dev install-hooks

# one-time: install golangci-lint + lefthook, then activate the pre-commit hook
install-hooks:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	go install github.com/evilmartians/lefthook@latest
	lefthook install

lint:
	golangci-lint run ./...
	cd src/frontend && npx biome check src

format:
	cd src/frontend && npx biome check --write src
	golangci-lint fmt

test:
	go test ./...
	cd src/frontend && npm test

build:
	wails build -tags webkit2_41

dev:
	wails dev -tags webkit2_41
