# Multi-arch Linkerd CLI image
FROM alpine:latest

ARG LINKERD_VERSION=edge-25.9.1
ARG TARGETARCH

RUN apk add --no-cache bash curl ca-certificates kubectl && \
    curl -fsSL "https://github.com/linkerd/linkerd2/releases/download/${LINKERD_VERSION}/linkerd2-cli-${LINKERD_VERSION}-linux-${TARGETARCH}" \
        -o /usr/local/bin/linkerd && \
    chmod +x /usr/local/bin/linkerd

ENTRYPOINT ["/bin/bash"]

