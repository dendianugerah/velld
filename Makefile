API_DIR := apps/api

.PHONY: help test test-api test-integration lint lint-api fmt vet build run

help:
	@echo "make test              run API unit tests"
	@echo "make test-integration  run API integration tests (needs docker + psql)"
	@echo "make lint              vet and check formatting"
	@echo "make fmt               format Go sources"
	@echo "make build             build the API and web app"
	@echo "make run               run the API locally"

test: test-api

test-api:
	cd $(API_DIR) && go test -race -cover ./...

test-integration:
	cd $(API_DIR) && go test -tags integration -race ./...

lint: lint-api

lint-api:
	cd $(API_DIR) && go vet ./...
	@unformatted=$$(cd $(API_DIR) && gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "unformatted files:"; echo "$$unformatted"; exit 1; \
	fi

fmt:
	cd $(API_DIR) && gofmt -w .

vet:
	cd $(API_DIR) && go vet ./...

build:
	cd $(API_DIR) && go build ./...
	cd apps/web && npm run build

run:
	cd $(API_DIR) && go run cmd/api-server/main.go
