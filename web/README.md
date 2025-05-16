# Go Sentinel Web Interface

This directory contains the web-based user interface for Go Sentinel, providing a rich, interactive experience for monitoring and controlling test execution.

## 📁 Directory Structure

```
web/
├── static/           # Static assets
│   ├── css/          # Stylesheets (Sass/CSS)
│   ├── js/           # Frontend JavaScript
│   └── images/       # Image assets
└── templates/        # Server-side templates
    ├── layouts/      # Base templates
    ├── pages/        # Page templates
    └── partials/     # Reusable components
```

## 🎨 Design System

### Template Hierarchy

1. **Layouts**: Base templates that define the overall page structure
   - `base.tmpl`: Main layout with HTML structure, head, and body
   - `auth.tmpl`: Authentication-specific layout

2. **Pages**: Individual page templates that extend layouts
   - `dashboard.tmpl`: Main test execution dashboard
   - `history.tmpl`: Test history and results
   - `settings.tmpl`: User and application settings

3. **Partials**: Reusable components
   - `_header.tmpl`: Navigation header
   - `_footer.tmpl`: Page footer
   - `_test_card.tmpl`: Test result card component
   - `_toast.tmpl`: Notification toasts

### Styling

- **CSS Methodology**: BEM (Block Element Modifier)
- **Preprocessor**: Sass (SCSS syntax)
- **Responsive**: Mobile-first approach
- **Theming**: Support for light/dark modes

## 🚀 Getting Started

### Prerequisites

- Node.js 16+ and npm/yarn/pnpm
- Go 1.17+

### Development

1. Install frontend dependencies:
   ```bash
   cd web/static/js
   pnpm install
   ```

2. Start the development server:
   ```bash
   pnpm dev
   ```

3. Build for production:
   ```bash
   pnpm build
   ```

## 🛠 Frontend Architecture

### Key Technologies

- **HTMX** for dynamic content loading and interactions
- **Alpine.js** for client-side reactivity
- **Tailwind CSS** for utility-first styling
- **Vite** for modern frontend tooling
- **Vitest** for unit testing
- **ESLint** and **Prettier** for code quality

### State Management

- **Server State**: Managed through HTMX and HTML fragments
- **Client State**: Lightweight state management with Alpine.js
- **URL State**: Used for navigation and sharing test results

### Real-time Updates

- **WebSocket** for real-time test execution updates
- **Server-Sent Events (SSE)** for one-way server-to-client updates
- **Polling Fallback** for environments without WebSocket support

## 🧪 Testing

### Unit Tests

Run unit tests with:
```bash
cd web/static/js
pnpm test
```

### E2E Tests

End-to-end tests use Playwright:
```bash
pnpm test:e2e
```

## 🎛 Configuration

Frontend configuration is managed through environment variables:

```env
VITE_API_BASE_URL=/api/v1
VITE_WS_URL=ws://localhost:8080/ws
VITE_APP_ENV=development
```

## 🌐 Browser Support

- Chrome (latest)
- Firefox (latest)
- Safari 14+
- Edge (latest)

## 📱 Responsive Design

The web interface is fully responsive and works on:
- Desktop (≥1024px)
- Tablet (≥768px)
- Mobile (<768px)

## 🔌 Plugins & Extensions

### HTMX Extensions
- `htmx-ws.js`: WebSocket integration
- `htmx-sse.js`: Server-Sent Events integration
- `htmx-debug.js`: Debugging helpers

### Alpine.js Plugins
- `focus`: For managing focus states
- `persist`: For persisting UI state
- `morph`: For DOM diffing and patching

## 🚀 Performance

### Optimizations

- **Code Splitting**: Automatic code splitting with Vite
- **Lazy Loading**: Components loaded on demand
- **Asset Optimization**: Minification and compression
- **Caching**: Proper cache headers for static assets

### Monitoring

- **Web Vitals**: Core Web Vitals monitoring
- **Error Tracking**: Frontend error logging
- **Analytics**: Usage statistics and performance metrics

## 🔒 Security

### Best Practices

- **CSP**: Content Security Policy headers
- **CSRF**: Built-in CSRF protection
- **XSS**: Automatic escaping in templates
- **CORS**: Proper CORS headers for API requests

## 📚 Documentation

### Additional Resources

- [HTMX Documentation](https://htmx.org/docs/)
- [Alpine.js Documentation](https://alpinejs.dev/)
- [Tailwind CSS Documentation](https://tailwindcss.com/docs)
- [Vite Documentation](https://vitejs.dev/guide/)

## 🤝 Contributing

Please see the main [CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines on contributing to the project.

## 📄 License

This project is licensed under the [MIT License](../LICENSE).
