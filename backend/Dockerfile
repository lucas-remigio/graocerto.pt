# Build stage
FROM golang:1.26.1-alpine AS builder

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

RUN apk add --no-cache ca-certificates bash

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy the compiled binaries from builder stage
COPY --from=builder /app/wallet-tracker ./
COPY --from=builder /app/migrate-tool ./

# Copy migration files from the correct path in builder stage
COPY --from=builder /app/cmd/migrate/migrations ./migrations/

# Copy prompts folder
COPY --from=builder /app/prompts ./prompts/

# Copy the SSL certificate
COPY --from=builder /app/db/prod-ca-2021.crt ./db/prod-ca-2021.crt

# Set environment variable for migrations path
ENV MIGRATIONS_PATH=file://migrations

# Secure permissions
RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080

# Use a shell script to run migrations then the app
CMD ["sh", "-c", "./migrate-tool up && ./wallet-tracker"]
