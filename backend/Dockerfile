# Build stage
FROM golang:1.23.3 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /wallet-tracker ./cmd/main.go

# Runtime stage
FROM alpine:latest
RUN apk add --no-cache make ca-certificates go

WORKDIR /app

# Copy the compiled binary and all source files
COPY --from=builder /wallet-tracker /wallet-tracker
COPY --from=builder /app /app

# Set environment variables for Go
ENV GOPATH=/app
ENV GO111MODULE=on

# Pre-download dependencies
RUN go mod download

EXPOSE 8080

CMD ["/bin/sh", "-c", "make migrate-up && /wallet-tracker"]