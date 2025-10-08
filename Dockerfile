# Stage 1: Build
FROM golang:1.23-alpine AS builder

RUN apk update && apk add --no-cache git

WORKDIR /app

# Copy go.mod and download dependencies
COPY go.mod ./
RUN go mod download

# Copy source code
COPY . .

# Build the Go app (statically linked binary)
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/main.go

# Stage 2: Run
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/main .

# Expose port
EXPOSE 8085

# Set environment variable for Gin mode
ENV GIN_MODE=release

# Run the binary
CMD ["./main"]