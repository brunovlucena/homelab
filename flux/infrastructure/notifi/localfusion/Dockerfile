# --- Build stage ---
FROM golang:1.23.2 AS builder
WORKDIR /app
COPY . .
RUN go mod download && go build -o main .

# --- Minimal runtime stage ---
FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /app/main /app/main
# If you need config files, add COPY lines here
CMD ["/app/main"]
