.PHONY: lint format test build dev

lint:
	golangci-lint run ./...
	cd src/frontend && npx biome check src

format:
	cd src/frontend && npx biome check --write src
	golangci-lint fmt

test:
	go test ./...

build:
	wails build -tags webkit2_41

dev:
	wails dev
