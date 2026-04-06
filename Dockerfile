# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Download dependencies first (layer caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /pipedrive-mcp-server ./cmd/server

# Runtime stage (minimal scratch image)
FROM scratch

# Copy CA certificates for HTTPS calls to Pipedrive API
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy binary
COPY --from=builder /pipedrive-mcp-server /pipedrive-mcp-server

# Run as non-root user (nobody = UID 65534)
USER 65534

ENTRYPOINT ["/pipedrive-mcp-server"]
