# Homepage Frontend

A modern, responsive React frontend for the Bruno portfolio website built with TypeScript, Vite, and Tailwind CSS.

## 🚀 Features

- **React 19** with TypeScript for type safety
- **Vite** for fast development and building
- **Tailwind CSS** for utility-first styling
- **React Query** for server state management
- **React Router** for client-side routing
- **Framer Motion** for smooth animations
- **Lucide React** for beautiful icons
- **React Hot Toast** for notifications
- **Responsive design** for all devices
- **Dark mode support**
- **Accessibility features**
- **Performance optimized**

## 📁 Project Structure

```
frontend/
├── src/
│   ├── components/     # Reusable UI components
│   │   ├── Header.tsx
│   │   └── Chatbot.tsx
│   ├── pages/         # Page components
│   │   ├── Home.tsx
│   │   └── Resume.tsx
│   ├── hooks/         # Custom React hooks
│   │   └── useApi.ts
│   ├── services/      # API services
│   │   └── api.ts
│   ├── utils/         # Utility functions
│   │   └── index.ts
│   ├── types/         # TypeScript type definitions
│   ├── App.tsx        # Main application component
│   ├── main.tsx       # Application entry point
│   └── index.css      # Global styles
├── public/            # Static assets
├── dist/              # Build output
├── package.json       # Dependencies and scripts
├── vite.config.ts     # Vite configuration
├── tailwind.config.js # Tailwind CSS configuration
├── tsconfig.json      # TypeScript configuration
└── README.md          # This file
```

## 🛠️ Technology Stack

- **Framework**: React 19 with TypeScript
- **Build Tool**: Vite
- **Styling**: Tailwind CSS
- **State Management**: React Query (TanStack Query)
- **Routing**: React Router DOM
- **Icons**: Lucide React
- **Animations**: Framer Motion
- **Notifications**: React Hot Toast
- **Testing**: Vitest + Playwright
- **Linting**: ESLint
- **Formatting**: Prettier

## 🚀 Quick Start

### Prerequisites

- Node.js 18+ 
- npm or yarn

### Installation

```bash
# Install dependencies
npm install

# Start development server
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview
```

### Environment Variables

Create a `.env` file in the frontend directory:

```env
# API Configuration
VITE_API_URL=http://localhost:8080/api

# Google Analytics Configuration
# Get your Measurement ID from Google Analytics:
# 1. Go to https://analytics.google.com
# 2. Click Admin (bottom left)
# 3. Select your property
# 4. Click "Data Streams" under Property column
# 5. Click on your web stream
# 6. Copy the "Measurement ID" (starts with G-)
VITE_GA_MEASUREMENT_ID=G-XXXXXXXXXX

# Google Cloud CDN Configuration (Optional)
# Set this to your Google Cloud Storage bucket URL with CDN enabled
# Example: https://storage.googleapis.com/your-bucket-name
# Or your custom domain if using Cloud CDN with custom domain
# Example: https://cdn.lucena.cloud
# Leave empty to use relative paths (fallback to local assets)
VITE_CDN_BASE_URL=

# Feature Flags
VITE_ENABLE_ANALYTICS=true
VITE_ENABLE_CHATBOT=true
```

#### Google Cloud CDN Setup

This project supports serving static assets via Google Cloud CDN using the free tier. To set it up:

1. **Run the Makefile target** (requires gcloud CLI):
   ```bash
   export GCP_PROJECT_ID=your-project-id
   export GCP_BUCKET_NAME=lucena-cloud-assets  # Optional, defaults to this
   make setup-gcp-cdn
   ```

2. **Or manually configure**:
   - Create a Google Cloud Storage bucket
   - Make it publicly readable
   - Upload assets from `public/assets/` to the bucket
   - (Optional) Set up Cloud CDN with a load balancer for better performance

3. **Set the CDN URL** in your `.env` file:
   ```env
   VITE_CDN_BASE_URL=https://storage.googleapis.com/your-bucket-name
   ```

**Free Tier Benefits:**
- 5 GB storage
- 1 GB egress per month
- Fast global content delivery

For more details, run `make help` to see all available targets.

## 📚 Available Scripts

- `npm run dev` - Start development server
- `npm run build` - Build for production
- `npm run preview` - Preview production build
- `npm run lint` - Run ESLint
- `npm run test` - Run unit tests
- `npm run test:ui` - Run tests with UI
- `npm run test:coverage` - Run tests with coverage
- `npm run test:e2e` - Run end-to-end tests
- `npm run test:e2e:ui` - Run e2e tests with UI

## 🎨 Styling

The project uses Tailwind CSS for styling with a custom configuration:

### Color Palette

```javascript
// tailwind.config.js
colors: {
  primary: {
    50: '#eff6ff',
    500: '#3b82f6',
    900: '#1e3a8a',
  },
  gray: {
    50: '#f9fafb',
    900: '#111827',
  }
}
```

### Dark Mode

Dark mode is supported and can be toggled using the `dark` class on the HTML element.

### Responsive Design

The application is fully responsive with breakpoints:
- `sm`: 640px
- `md`: 768px
- `lg`: 1024px
- `xl`: 1280px
- `2xl`: 1536px

## 🔧 API Integration

The frontend integrates with the backend API through a centralized service layer:

### API Client

```typescript
import { apiClient } from './services/api'

// Get all projects
const projects = await apiClient.getProjects()

// Create a new project
const newProject = await apiClient.createProject({
  title: 'My Project',
  description: 'Project description',
  // ... other fields
})
```

### React Query Hooks

```typescript
import { useProjects, useCreateProject } from './hooks/useApi'

function ProjectsList() {
  const { data: projects, isLoading, error } = useProjects()
  const createProject = useCreateProject()

  // Use the data and mutations
}
```

## 🧪 Testing

### Unit Tests

Tests are written with Vitest and React Testing Library:

```bash
# Run all tests
npm run test

# Run tests in watch mode
npm run test:watch

# Run tests with coverage
npm run test:coverage
```

### End-to-End Tests

E2E tests use Playwright:

```bash
# Run e2e tests
npm run test:e2e

# Run e2e tests with UI
npm run test:e2e:ui
```

## 📦 Build & Deployment

### Development Build

```bash
npm run dev
```

### Production Build

```bash
npm run build
```

The build output is optimized for production with:
- Code splitting
- Tree shaking
- Minification
- Asset optimization

### Docker Deployment

```dockerfile
FROM node:18-alpine as builder
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/nginx.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

## 🔍 Performance

### Optimization Features

- **Code Splitting**: Automatic route-based code splitting
- **Lazy Loading**: Components loaded on demand
- **Image Optimization**: Automatic image optimization
- **Caching**: React Query caching for API responses
- **Bundle Analysis**: Built-in bundle analyzer

### Performance Monitoring

```bash
# Analyze bundle size
npm run build -- --analyze
```

## 🛡️ Security

### Security Features

- **Content Security Policy**: Configured in index.html
- **HTTPS Only**: Production builds enforce HTTPS
- **XSS Protection**: React's built-in XSS protection
- **Input Validation**: Client-side validation with TypeScript

## 🌐 Browser Support

- Chrome 90+
- Firefox 88+
- Safari 14+
- Edge 90+

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

### Code Style

The project uses ESLint and Prettier for code formatting:

```bash
# Format code
npm run format

# Check code style
npm run lint
```

## 📄 License

This project is part of the Bruno portfolio website.

## 🆘 Troubleshooting

### Common Issues

1. **Port already in use**: Change the port in `vite.config.ts`
2. **API connection issues**: Check the `VITE_API_URL` environment variable
3. **Build failures**: Clear node_modules and reinstall dependencies

### Getting Help

- Check the [Vite documentation](https://vitejs.dev/)
- Review the [React Query documentation](https://tanstack.com/query)
- Consult the [Tailwind CSS documentation](https://tailwindcss.com/) 