import { expect, vi, afterEach, beforeAll, afterAll } from 'vitest';
import { cleanup } from '@testing-library/react';
import * as matchers from '@testing-library/jest-dom/matchers';

// Extend Vitest's expect with jest-dom matchers
expect.extend(matchers);

// Mock WebSocket
class MockWebSocket extends EventTarget {
  static readonly CONNECTING = 0;
  static readonly OPEN = 1;
  static readonly CLOSING = 2;
  static readonly CLOSED = 3;

  url: string;
  readyState: number;
  binaryType: BinaryType = 'arraybuffer';
  bufferedAmount = 0;
  extensions = '';
  protocol = '';
  onopen: ((this: WebSocket, ev: Event) => any) | null = null;
  onclose: ((this: WebSocket, ev: CloseEvent) => any) | null = null;
  onmessage: ((this: WebSocket, ev: MessageEvent) => any) | null = null;
  onerror: ((this: WebSocket, ev: Event) => any) | null = null;

  constructor(url: string | URL, _protocols?: string | string[]) {
    super();
    this.url = typeof url === 'string' ? url : url.toString();
    this.readyState = MockWebSocket.CONNECTING;
  }

  send(_data: string | ArrayBufferLike | Blob | ArrayBufferView): void {
    // Mock implementation
  }

  close(code?: number, reason?: string): void {
    this.readyState = MockWebSocket.CLOSED;
    if (this.onclose) {
      const event = new CloseEvent('close', { code: code || 1000, reason: reason || '', wasClean: true });
      this.onclose.call(this as unknown as WebSocket, event);
    }
  }

  // Test helper methods
  simulateOpen() {
    this.readyState = MockWebSocket.OPEN;
    if (this.onopen) {
      this.onopen.call(this as unknown as WebSocket, new Event('open'));
    }
  }

  simulateMessage(data: any) {
    const event = new MessageEvent('message', { data });
    
    // Call the onmessage handler if it exists
    if (this.onmessage) {
      this.onmessage.call(this as unknown as WebSocket, event);
    }
    
    // Dispatch the event to all listeners
    this.dispatchEvent(event);
  }

  simulateError() {
    if (this.onerror) {
      this.onerror.call(this as unknown as WebSocket, new Event('error'));
    }
  }
}

// Add to global scope
global.WebSocket = MockWebSocket as unknown as typeof WebSocket;

// Mock global objects
const localStorageMock = (() => {
  let store: Record<string, string> = {};
  return {
    getItem: (key: string) => store[key] || null,
    setItem: (key: string, value: string) => {
      store[key] = value.toString();
    },
    removeItem: (key: string) => {
      delete store[key];
    },
    clear: () => {
      store = {};
    },
  };
})();

// Set up global mocks
global.localStorage = localStorageMock as Storage;
global.sessionStorage = {
  ...localStorageMock,
  clear: () => {
    // Keep session storage behavior different if needed
  },
} as Storage;

// Mock matchMedia
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation((query) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
});

// Mock requestAnimationFrame and cancelAnimationFrame
global.requestAnimationFrame = ((callback: FrameRequestCallback) => {
  return setTimeout(callback, 0) as unknown as number;
}) as any;

global.cancelAnimationFrame = ((id: number) => {
  clearTimeout(id);
}) as any;

// Clean up after each test
afterEach(() => {
  cleanup();
  vi.clearAllMocks();
  localStorage.clear();
  sessionStorage.clear();
});

// Global test setup
beforeAll(() => {
  // Add any global test setup here
});

afterAll(() => {
  // Clean up any resources
  vi.restoreAllMocks();
});

// Mock HTMX
const htmx = {
  on: vi.fn(),
  off: vi.fn(),
  process: vi.fn(),
  trigger: vi.fn(),
  find: vi.fn(),
  findAll: vi.fn(),
  ajax: vi.fn(),
};

// Mock window.htmx
Object.defineProperty(window, 'htmx', {
  value: htmx,
  writable: true,
});

// Mock console methods
const consoleError = console.error;
const consoleWarn = console.warn;
const consoleLog = console.log;

beforeAll(() => {
  // Suppress console output during tests
  console.error = vi.fn();
  console.warn = vi.fn();
  console.log = vi.fn();
});

afterAll(() => {
  // Restore console methods
  console.error = consoleError;
  console.warn = consoleWarn;
  console.log = consoleLog;
});

// Mock the WebSocket client
const mockWebSocketClient = {
  connect: vi.fn(),
  disconnect: vi.fn(),
  send: vi.fn(),
  on: vi.fn(),
  off: vi.fn(),
};

Object.defineProperty(window, 'goSentinelWebSocket', {
  value: mockWebSocketClient,
  writable: true,
});

// Mock the toast notification system
const mockToast = {
  success: vi.fn(),
  error: vi.fn(),
  info: vi.fn(),
  warning: vi.fn(),
};

Object.defineProperty(window, 'toast', {
  value: mockToast,
  writable: true,
});
