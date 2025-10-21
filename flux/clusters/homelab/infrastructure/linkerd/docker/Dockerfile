# Dockerfile to build a custom Linkerd CLI image
# This solves registry authentication issues by hosting on our own GHCR
FROM curlimages/curl:latest as downloader

ARG LINKERD_VERSION=stable-2.16.2
ARG TARGETARCH

WORKDIR /tmp
RUN curl -sLO "https://github.com/linkerd/linkerd2/releases/download/${LINKERD_VERSION}/linkerd2-cli-${LINKERD_VERSION}-linux-${TARGETARCH}" && \
    chmod +x linkerd2-cli-${LINKERD_VERSION}-linux-${TARGETARCH}

FROM alpine:latest

ARG TARGETARCH

RUN apk add --no-cache bash curl ca-certificates

COPY --from=downloader /tmp/linkerd2-cli-* /linkerd

# Install kubectl
RUN curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/${TARGETARCH}/kubectl" && \
    chmod +x kubectl && \
    mv kubectl /usr/local/bin/

ENTRYPOINT ["/bin/bash"]

