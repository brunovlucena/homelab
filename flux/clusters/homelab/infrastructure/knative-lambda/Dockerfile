# Dockerfile for Knative Lambda Builder Service

# ---- Builder Stage ----
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
# It's recommended to have a .dockerignore file to exclude unnecessary files
COPY . .

# Build static binary
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o service ./cmd/service

# ---- Final Stage ----
# Use a specific version for reproducibility
FROM alpine:3.19

RUN apk --no-cache add ca-certificates

# Run as non-root user for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

WORKDIR /home/appuser

# Copy artifacts and set ownership
COPY --from=builder --chown=appuser:appgroup /app/service .

CMD ["./service"] 