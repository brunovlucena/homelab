# Production stage - nginx for static site
FROM nginx:alpine

# Copy static files to nginx html directory
COPY index.html styles.css script.js /usr/share/nginx/html/

# Copy nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Expose port 80
EXPOSE 80

# Start nginx
CMD ["nginx", "-g", "daemon off;"]

