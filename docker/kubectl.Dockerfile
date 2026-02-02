# 🐰 kubectl image with krew and RabbitMQ plugin
FROM alpine:3.22.2

ARG KUBECTL_VERSION=v1.34.0
ARG KREW_VERSION=v0.4.4
ARG TARGETARCH
# Set Linkerd version
ARG LINKERD_VERSION=edge-25.11.1

# Set PATH for krew
ENV PATH="${PATH}:/root/.krew/bin"


RUN apk add --no-cache \
    curl \
    kubectl \
    openssl \
    bash \
    ca-certificates \
    git \
    jq

# Install kubectl and linkerd-cli
RUN echo "📦 Installing Linkerd CLI ${LINKERD_VERSION} for ${TARGETARCH}..." && \
    curl -fsSL "https://github.com/linkerd/linkerd2/releases/download/${LINKERD_VERSION}/linkerd2-cli-${LINKERD_VERSION}-linux-${TARGETARCH}" \
    -o /usr/local/bin/linkerd && \
    chmod +x /usr/local/bin/linkerd && \
    /usr/local/bin/linkerd version --client && \
    echo "✅ Linkerd CLI installed successfully"

# Install krew
RUN set -x; cd "$(mktemp -d)" && \
    OS="$(uname | tr '[:upper:]' '[:lower:]')" && \
    ARCH="${TARGETARCH}" && \
    KREW="krew-${OS}_${ARCH}" && \
    curl -fsSLO "https://github.com/kubernetes-sigs/krew/releases/download/${KREW_VERSION}/${KREW}.tar.gz" && \
    tar zxvf "${KREW}.tar.gz" && \
    ./"${KREW}" install krew

# Install RabbitMQ kubectl plugin
RUN kubectl krew install rabbitmq

# Install Inspektor Gadget kubectl plugin
RUN kubectl krew install gadget

# Install step-cli
ARG STEP_VERSION=0.28.0
RUN echo "📦 Installing step-cli ${STEP_VERSION} for ${TARGETARCH}..." && \
    case "${TARGETARCH}" in \
      amd64) STEP_ARCH="amd64" ;; \
      arm64) STEP_ARCH="arm64" ;; \
      *) echo "❌ Unsupported architecture: ${TARGETARCH}" >&2; exit 1 ;; \
    esac && \
    cd "$(mktemp -d)" && \
    curl -fsSL "https://github.com/smallstep/cli/releases/download/v${STEP_VERSION}/step_linux_${STEP_VERSION}_${STEP_ARCH}.tar.gz" \
      | tar -xz && \
    mv step_${STEP_VERSION}/bin/step /usr/local/bin/step && \
    chmod +x /usr/local/bin/step && \
    /usr/local/bin/step version && \
    echo "✅ step-cli installed successfully"

ENTRYPOINT ["/bin/bash"]

