# Lucena.app - Coming Soon Page

A static "coming soon" page for lucena.app with a cyberpunk/command center aesthetic matching the agent command centers in the homelab repo.

## Features

- **Cyberpunk Design**: Dark theme with purple/pink gradients
- **Animated Elements**: Floating particles, progress bar, and grid background
- **Responsive**: Works on all device sizes
- **Modern UI**: Glassmorphism effects and smooth animations
- **Kubernetes Ready**: Full K8s deployment with nginx, health checks, and Cloudflare Tunnel integration

## Files

- `index.html` - Main HTML structure
- `styles.css` - All styling with cyberpunk theme
- `script.js` - Interactive animations and effects
- `Dockerfile` - Container image definition
- `nginx.conf` - Nginx configuration
- `k8s/` - Kubernetes manifests
- `Makefile` - Build and deployment automation

## Local Development

Open `index.html` in a browser or use a simple HTTP server:

```bash
python3 -m http.server 8000
# or
npx serve .
```

## Kubernetes Deployment

### Prerequisites

- Docker registry accessible at `localhost:5001`
- Kubernetes cluster with Flux CD
- Cloudflare Tunnel Operator installed (for external access)

### Build and Deploy

1. **Build the Docker image:**
   ```bash
   make build
   ```

2. **Push to registry:**
   ```bash
   make push
   ```

3. **Deploy to Kubernetes:**
   ```bash
   make deploy
   # or
   kubectl apply -k k8s/kustomize/base/
   ```

4. **Or do everything at once:**
   ```bash
   make all
   ```

### Flux CD Integration

To manage via Flux, add to your cluster's kustomization:

```yaml
resources:
  - ../../../../apps/lucena-app/k8s/kustomize/base
```

### Access

The site will be available at `https://lucena.app` via Cloudflare Tunnel.

## Customization

### Colors
Edit the CSS variables in `styles.css`:
```css
:root {
  --cyber-purple: #8b5cf6;
  --cyber-pink: #ec4899;
  /* ... */
}
```

### Progress Percentage
Update the progress value in `index.html`:
```html
<div class="progress-fill" data-progress="87"></div>
```

And update the display:
```html
<span class="progress-percentage">87%</span>
```

### Text Content
All text content is in `index.html` and can be easily modified.

### Image Registry
Update the `Makefile` or deployment YAML to use your registry:
```yaml
image: your-registry.com/lucena-app:latest
```

## Kubernetes Resources

- **Namespace**: `lucena-app`
- **Deployment**: 2 replicas with rolling updates
- **Service**: ClusterIP on port 80
- **Ingress**: Cloudflare Tunnel at `lucena.app`
- **Resources**: 50m CPU / 64Mi memory (requests), 200m CPU / 128Mi (limits)

## Health Checks

- **Liveness**: `/health` endpoint
- **Readiness**: `/health` endpoint
- Both configured with appropriate delays

## Browser Support

- Chrome/Edge (latest)
- Firefox (latest)
- Safari (latest)
- Mobile browsers

## License

Part of the homelab project.

