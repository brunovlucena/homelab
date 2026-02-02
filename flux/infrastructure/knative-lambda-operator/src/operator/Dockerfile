# Dockerfile for Knative Lambda Operator
# Build context: src/operator/ directory

# ---- Builder Stage ----
FROM golang:1.24.3-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . ./

# Build static binary (multi-arch support via TARGETARCH)
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH:-amd64} go build -ldflags="-s -w" -o /app/manager ./cmd

# ---- Final Stage ----
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

# Run as non-root user for security (use numeric IDs for k8s runAsNonRoot)
RUN addgroup -g 65532 -S nonroot && adduser -u 65532 -S nonroot -G nonroot
USER 65532:65532

WORKDIR /home/nonroot

# Copy binary
COPY --from=builder --chown=65532:65532 /app/manager .

# Expose metrics, health, and webhook ports
EXPOSE 8080 8081 8082

ENTRYPOINT ["./manager"]
