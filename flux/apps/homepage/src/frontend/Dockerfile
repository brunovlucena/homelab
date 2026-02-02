# Build stage
FROM node:22-alpine AS builder

# Set working directory
WORKDIR /app

# Copy package files first for better caching
COPY package*.json ./

# Install all dependencies (including dev dependencies for build)
RUN npm ci --legacy-peer-deps && npm cache clean --force

# Copy source code
COPY . .

# Build the application using the npm script (which runs tsc && vite build)
# But skip tsc by using vite build directly
RUN node node_modules/vite/bin/vite.js build

# Production stage - optimized nginx
FROM nginx:alpine

# Create nginx cache directories with proper permissions
RUN mkdir -p /tmp/cache/nginx/{client_temp,proxy_temp,fastcgi_temp,uwsgi_temp,scgi_temp} && \
    chmod -R 755 /tmp/cache/nginx

# Copy built assets from builder stage
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy optimized nginx configurations
COPY nginx-main.conf /etc/nginx/nginx.conf
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Expose port 8080
EXPOSE 8080

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
