# Go Sentinel Frontend JavaScript

This directory contains the JavaScript code for the Go Sentinel frontend.

## Development Setup

1. Install dependencies:
   ```bash
   pnpm install
   ```

2. Run tests:
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

3. Lint code:
   ```bash
   pnpm lint
   ```

4. Format code:
   ```bash
   pnpm format
   ```

## Project Structure

- `/test` - Test files
  - `setup.js` - Test setup and global configurations
  - `*.test.js` - Test files
- `main.js` - Main application entry point
- `vitest.config.js` - Vitest configuration
- `.eslintrc.cjs` - ESLint configuration
- `.prettierrc` - Prettier configuration

## Writing Tests

1. Create a new test file with the pattern `*.test.js` in the `test` directory.
2. Use the `describe` and `it` functions to structure your tests.
3. Use Vitest's `expect` for assertions.

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
