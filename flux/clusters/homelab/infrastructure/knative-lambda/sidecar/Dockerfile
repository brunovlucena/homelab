# Dockerfile for the sidecar container
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum from the build context root
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire knative-lambda project
COPY . .

# Build the sidecar binary from the correct path, output to bin directory to avoid naming conflict
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o bin/sidecar ./sidecar/cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates

# Create a non-root user
RUN addgroup -g 1000 sidecar && adduser -u 1000 -G sidecar -s /bin/sh -D sidecar

# Set working directory and change ownership
WORKDIR /app
RUN chown sidecar:sidecar /app

# Copy the binary from builder stage and set proper permissions
COPY --from=builder /app/bin/sidecar ./sidecar
RUN chmod +x ./sidecar && chown sidecar:sidecar ./sidecar

# Switch to non-root user
USER sidecar

CMD ["./sidecar"] 