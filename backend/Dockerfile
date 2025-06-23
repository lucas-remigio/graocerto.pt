# Build stage
FROM golang:1.23.3-alpine AS builder

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the main binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o wallet-tracker \
    ./cmd/main.go

# Build the migration binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o migrate-tool \
    ./cmd/migrate/main.go

# Runtime stage
FROM alpine:latest

# Install only necessary packages
RUN apk add --no-cache ca-certificates bash

WORKDIR /app

# Copy wait script and make executable
COPY wait-for-it.sh ./
RUN chmod +x wait-for-it.sh

# Copy the compiled binaries from builder stage
COPY --from=builder /app/wallet-tracker ./
COPY --from=builder /app/migrate-tool ./

# Copy migration files to a simple path
COPY --from=builder /app/cmd/migrate/migrations ./migrations/

# Set environment variable for migrations path
ENV MIGRATIONS_PATH=file://migrations

EXPOSE 8080

CMD ["sh", "-c", "./wait-for-it.sh mysql-container:3306 -- ./migrate-tool up && ./wallet-tracker"]