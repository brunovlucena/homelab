# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod ./

# Download dependencies (if any)
RUN go mod download || true

# Copy source code
COPY . .

# Build the metric-pusher binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o metric-pusher .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S metricpusher && \
    adduser -u 1001 -S metricpusher -G metricpusher

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/metric-pusher .

# Change ownership to non-root user
RUN chown -R metricpusher:metricpusher /app

# Switch to non-root user
USER metricpusher

# Set entrypoint
ENTRYPOINT ["./metric-pusher"] 