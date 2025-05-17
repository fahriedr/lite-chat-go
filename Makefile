build:
	@go build -o ./tmp/main ./cmd/main.go

test:
	@go test -v ./...

run: build
	@./tmp/main