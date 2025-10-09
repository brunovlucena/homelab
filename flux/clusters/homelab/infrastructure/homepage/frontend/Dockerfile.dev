# 🚀 Homepage Frontend Dockerfile
# Multi-stage build supporting both development and production

# Build stage
FROM node:24 AS builder

# Set working directory
WORKDIR /app

# Copy package files first for better caching
COPY package*.json ./

# Install all dependencies
RUN npm ci --legacy-peer-deps && npm cache clean --force

# Copy source code
COPY . .

# Build the application
RUN npm run build

# Development stage
FROM node:22-alpine AS development

# Set working directory
WORKDIR /app

# Copy package files first for better caching
COPY package*.json ./

# Install all dependencies with better error handling
RUN npm ci --legacy-peer-deps --no-audit --no-fund || \
    (npm cache clean --force && npm ci --legacy-peer-deps --no-audit --no-fund) || \
    npm install --legacy-peer-deps --no-audit --no-fund

# Copy source code
COPY . .

# Expose port 8080
EXPOSE 8080

# Start development server
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0", "--port", "8080"]

# Production stage - nginx with debug tools
FROM nginx:stable AS production

# Install debug tools
RUN apt-get update && apt-get install -y \
    curl \
    wget \
    net-tools \
    procps \
    vim \
    less \
    htop \
    tcpdump \
    netcat-openbsd \
    && rm -rf /var/lib/apt/lists/*

# Create nginx cache directories with proper permissions
RUN mkdir -p /tmp/cache/nginx/{client_temp,proxy_temp,fastcgi_temp,uwsgi_temp,scgi_temp} && \
    chmod -R 755 /tmp/cache/nginx

# Copy built assets from builder stage
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy custom nginx config
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Expose port 8080
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/ || exit 1

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
