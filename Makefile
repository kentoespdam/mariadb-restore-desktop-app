.PHONY: lint format test build dev

lint:
	go vet ./...
	cd src/frontend && npx biome check src

format:
	cd src/frontend && npx biome format --write src

test:
	go test ./...

build:
	wails build -tags webkit2_41

dev:
	wails dev
