# Go Sentinel Frontend

This directory contains the TypeScript/JavaScript code for the Go Sentinel frontend, which provides a lightweight, progressive enhancement layer on top of server-rendered HTML. The application is built with TypeScript for type safety and better developer experience.

## 🚀 Development Setup

### Prerequisites
- Node.js 18.0.0+ (as specified in package.json)
- pnpm 8.6.0+
- TypeScript 5.0.0+

### Installation

```bash
# Install dependencies
pnpm install

# Build the project
pnpm build

# Start development server
pnpm dev
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

# Run type checking
pnpm type-check
```

### Test Structure

Tests are located in the `test/` directory:
- `test/setup.ts` - Test setup and global configurations
- `test/**/*.test.ts` - Test files
- `test/utils/` - Test utilities and helpers

## 🛠 Code Quality

### Linting
```bash
# Lint code
pnpm lint

# Auto-fix linting issues
pnpm lint:fix
```

### Type Checking
```bash
# Check types
pnpm type-check
```

### Formatting
```bash
# Format code
pnpm format
```

## 📁 Project Structure

```
web/static/js/
├── src/                     # Source files
│   ├── types/              # TypeScript type definitions
│   │   ├── env.d.ts        # Environment variable types
│   │   └── global.d.ts     # Global type declarations
│   ├── utils/              # Utility functions
│   │   └── example.ts      # Example utility module
│   ├── main.ts             # Main application entry point
│   └── ...
├── test/                   # Test files
│   ├── setup.ts            # Test setup and configurations
│   ├── utils/              # Test utilities
│   └── **/*.test.ts        # Test files
├── .vscode/                # VSCode settings
│   ├── extensions.json     # Recommended extensions
│   └── settings.json       # Workspace settings
├── .eslintrc.cjs           # ESLint configuration
├── .prettierrc             # Prettier configuration
├── package.json            # Project manifest
├── tsconfig.json           # TypeScript configuration
├── tsconfig.test.json      # TypeScript test configuration
└── vitest.config.ts        # Vitest test runner configuration
```

## 🧩 Key Features

### Core Functionality
- Type-safe code with TypeScript
- Mobile-responsive navigation
- Test selection and management
- WebSocket integration for real-time updates
- Toast notifications

### Implementation Details
- TypeScript for type safety
- Vite for fast development and building
- HTMX for progressive enhancement
- Vitest for testing
- ESLint and Prettier for code quality

## 📝 Writing Tests

### Example Test
```typescript
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { greet } from '../../src/utils/example';

describe('greet', () => {
  it('should return a greeting with the provided name', () => {
    const name = 'John';
    const result = greet(name);
    expect(result).toBe(`Hello, ${name}! Welcome to Go Sentinel!`);
  });

  it('should throw an error if name is not provided', () => {
    expect(() => greet('')).toThrow('Name is required');
  });
});
```

## 🔧 Development Workflow

1. **Start the development server**:
   ```bash
   pnpm dev
   ```

2. **Run tests in watch mode**:
   ```bash
   pnpm test:watch
   ```

3. **Run type checking**:
   ```bash
   pnpm type-check
   ```

4. **Build for production**:
   ```bash
   pnpm build
   ```

## 🛡 Type Safety

This project uses TypeScript for type safety. All new code should be written in TypeScript with proper type annotations. Use the following guidelines:

- Always define types for function parameters and return values
- Use interfaces for object shapes
- Leverage TypeScript's utility types when appropriate
- Avoid using `any` - prefer `unknown` or proper type definitions
- Use type guards for runtime type checking

## 📚 Resources

- [TypeScript Documentation](https://www.typescriptlang.org/docs/)
- [Vite Documentation](https://vitejs.dev/guide/)
- [Vitest Documentation](https://vitest.dev/guide/)
- [HTMX Documentation](https://htmx.org/docs/)
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
