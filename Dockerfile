# Build stage
FROM golang:1.23.3-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build both binaries
RUN CGO_ENABLED=0 GOOS=linux go build -o /glazeld ./cmd/glazeld
RUN CGO_ENABLED=0 GOOS=linux go build -o /glazel-agent ./cmd/glazel-agent

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy both binaries from builder
COPY --from=builder /glazeld .
COPY --from=builder /glazel-agent .

# Default to running glazeld (orchestrator)
CMD ["./glazeld"]
