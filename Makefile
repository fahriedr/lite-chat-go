build:
	@go build -o ./tmp/main ./cmd/main.go

test:
	@go test -v ./...

test-coverage:
	@go test -v -cover ./...

test-coverage-html:
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-race:
	@go test -race -v ./...

test-short:
	@go test -short -v ./...

test-clean:
	@go clean -testcache
	@rm -f coverage.out coverage.html

benchmark:
	@go test -bench=. -benchmem ./...

test-user:
	@go test -v ./service/user/...

test-conversation:
	@go test -v ./service/conversation/...

test-message:
	@go test -v ./service/message/...

test-contact:
	@go test -v ./service/contact/...

test-api:
	@go test -v ./cmd/api/...

test-utils:
	@go test -v ./utils/...

test-all: test-clean test

run: build
	@./tmp/main

run-docker:
	@docker compose up -d --build

stop-docker:
	@docker compose down

logs-docker:
	@docker compose logs -f api

.PHONY: build test test-coverage test-coverage-html test-race test-short test-clean benchmark test-user test-conversation test-message test-contact test-api test-utils test-all run run-docker stop-docker logs-docker