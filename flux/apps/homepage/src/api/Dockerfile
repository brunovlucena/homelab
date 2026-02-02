# Build stage
FROM golang:1.25-alpine AS builder

# Install git and ca-certificates for dependencies
RUN apk update && apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations
ARG TARGETOS=linux
ARG TARGETARCH=amd64
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build \
    -ldflags="-w -s -extldflags=-static" \
    -a -installsuffix cgo \
    -o main .

# Final stage - use scratch for minimal size
FROM scratch

# Copy ca-certificates from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary
COPY --from=builder /app/main /app/main

# Expose port
EXPOSE 8080

# Run the application
CMD ["/app/main"] 