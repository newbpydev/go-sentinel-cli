import { expect, vi, beforeEach, afterEach } from 'vitest';

// Global mocks or setup can go here
beforeEach(() => {
  // Reset any mocks or setup before each test
  vi.clearAllMocks();
  // Reset the DOM before each test
  document.body.innerHTML = '';
});

// Make expect available globally for convenience
global.expect = expect;
