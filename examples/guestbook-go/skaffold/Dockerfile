FROM golang:1.13.5 AS builder
# Use ADD in case of a URL only.
COPY ./cmd/app /cmd/app
COPY ./cmd/utils /cmd/utils
WORKDIR /cmd/app
# CGO is disabled as it is not compatible with the alpine image. 
# To use CGO write normal Go code that imports a pseudo-package "C". 
# Use go -gcflags to adds more debug information for the debugger.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o main .

# final stage
FROM alpine:3.11
COPY ./cmd/app/config.yaml /app/config.yaml
COPY --from=builder /cmd/app/main /app/main
RUN chmod +x /app/main
ENTRYPOINT ["/app/main"]
EXPOSE 8000
