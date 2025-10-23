# 🐰 kubectl image with krew and RabbitMQ plugin
FROM alpine:latest

ARG KUBECTL_VERSION=v1.34.0
ARG KREW_VERSION=v0.4.4
ARG TARGETARCH

# Install kubectl and dependencies
RUN apk add --no-cache bash curl ca-certificates git && \
    curl -fsSL "https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/${TARGETARCH}/kubectl" \
        -o /usr/local/bin/kubectl && \
    chmod +x /usr/local/bin/kubectl

# Install krew
RUN set -x; cd "$(mktemp -d)" && \
    OS="$(uname | tr '[:upper:]' '[:lower:]')" && \
    ARCH="${TARGETARCH}" && \
    KREW="krew-${OS}_${ARCH}" && \
    curl -fsSLO "https://github.com/kubernetes-sigs/krew/releases/download/${KREW_VERSION}/${KREW}.tar.gz" && \
    tar zxvf "${KREW}.tar.gz" && \
    ./"${KREW}" install krew

# Set PATH for krew
ENV PATH="${PATH}:/root/.krew/bin"

# Install RabbitMQ kubectl plugin
RUN kubectl krew install rabbitmq

# Verify installation
RUN kubectl rabbitmq help

ENTRYPOINT ["/bin/bash"]

