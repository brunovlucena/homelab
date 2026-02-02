FROM rust:1.83 AS builder

WORKDIR /app

# Copy workspace files
COPY Cargo.toml ./
COPY shared ./shared
COPY messaging-service ./messaging-service
COPY user-service ./user-service
COPY agent-gateway ./agent-gateway
COPY message-processor ./message-processor
COPY message-storage-service ./message-storage-service

# Copy Cargo.lock if it exists (optional for workspace)
COPY Cargo.lock* ./

# Build the service
RUN cargo build --release --package message-processor

# Runtime stage
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /app/target/release/message-processor /app/message-processor

EXPOSE 8080

CMD ["/app/message-processor"]
