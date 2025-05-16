# Go Sentinel Frontend JavaScript

This directory contains the JavaScript code for the Go Sentinel frontend, which provides a lightweight, progressive enhancement layer on top of server-rendered HTML. The JavaScript is primarily used for testing and enhancing the user interface with interactive features.

## 🚀 Development Setup

### Prerequisites
- Node.js 18.0.0+ (as specified in package.json)
- pnpm 8.6.0+

### Installation

```bash
# Install dependencies
pnpm install
```

## 🧪 Testing

### Running Tests

```bash
# Run tests once
pnpm test

# Run tests in watch mode
pnpm test:watch

# Run tests with coverage
pnpm test:coverage

# Run tests with UI
pnpm test:ui
```

### Test Structure

Tests are located in the `test/` directory:
- `test/setup.js` - Test setup and global configurations
- `test/example.test.js` - Example test file

## 🛠 Code Quality

### Linting
```bash
# Lint code
pnpm lint

# Auto-fix linting issues
pnpm lint:fix
```

### Formatting
```bash
# Format code
pnpm format
```

## 📁 Project Structure

```
web/static/js/
├── test/                   # Test files
│   ├── example.test.js     # Example test file
│   └── setup.js            # Test setup and configurations
├── main.js                 # Main JavaScript file with core functionality
├── coverage.js             # Coverage report handling
├── settings.js             # Settings page functionality
├── toast.js                # Toast notification system
├── .eslintrc.cjs           # ESLint configuration
├── .prettierrc             # Prettier configuration
├── package.json            # Project manifest
└── vitest.config.js        # Vitest test runner configuration
```

## 🧩 Key Features

### Core Functionality
- Mobile-responsive navigation
- Test selection and management
- WebSocket integration for real-time updates
- Toast notifications

### Implementation Details
- Vanilla JavaScript with modern ES6+ features
- HTMX for progressive enhancement
- Vitest for testing
- ESLint and Prettier for code quality

## 📝 Writing Tests

### Example Test
```javascript
import { describe, it, expect, beforeEach } from 'vitest';
import { setupMobileMenu } from '../main.js';

describe('Mobile Menu', () => {
  beforeEach(() => {
    document.body.innerHTML = `
      <button class="mobile-menu-toggle">☰</button>
      <nav class="main-nav">
        <a href="/">Home</a>
      </nav>
    `;
  });

  it('should toggle mobile menu when button is clicked', () => {
    setupMobileMenu();
    const toggleBtn = document.querySelector('.mobile-menu-toggle');
    const nav = document.querySelector('.main-nav');
    
    // Initial state
    expect(nav.classList.contains('active')).toBe(false);
    
    // After click
    toggleBtn.click();
    expect(nav.classList.contains('active')).toBe(true);
    
    // After second click
    toggleBtn.click();
    expect(nav.classList.contains('active')).toBe(false);
  });
});
```

## 🔧 Configuration

### Environment Variables
No environment-specific configuration is required for local development.

### HTMX Integration
The frontend leverages HTMX for dynamic content loading and WebSocket integration. The main JavaScript (`main.js`) enhances the server-rendered HTML with additional interactivity.

## 📚 Documentation

- [HTMX Documentation](https://htmx.org/docs/)
- [Vitest Documentation](https://vitest.dev/guide/)
- [JavaScript Testing Best Practices](https://github.com/goldbergyoni/javascript-testing-best-practices)

## 🤝 Contributing

1. Ensure all tests pass before submitting changes
2. Follow the existing code style and patterns
3. Add tests for new functionality
4. Update documentation as needed

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](../../../LICENSE) file for details.

Example test:

```javascript
import { describe, it, expect } from 'vitest';

describe('MyComponent', () => {
  it('should do something', () => {
    const result = 1 + 1;
    expect(result).toBe(2);
  });
});
```

## Code Style

- Follow the [Airbnb JavaScript Style Guide](https://github.com/airbnb/javascript)
- Use ESLint and Prettier for code quality and formatting
- Write tests for all new features and bug fixes
