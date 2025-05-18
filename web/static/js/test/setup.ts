/**
 * Test setup file for Go Sentinel
 * 
 * This file sets up the test environment with necessary mocks and utilities
 * before running any tests.
 */

// Polyfill CloseEvent if not present (for jsdom)
if (typeof globalThis.CloseEvent === 'undefined') {
  class CloseEvent extends Event {
    code: number;
    reason: string;
    wasClean: boolean;
    constructor(type: string, eventInitDict: CloseEventInit = {}) {
      super(type, eventInitDict);
      this.code = eventInitDict.code || 1000;
      this.reason = eventInitDict.reason || '';
      this.wasClean = eventInitDict.wasClean || false;
    }
  }
  // @ts-ignore - We're polyfilling the global CloseEvent
  globalThis.CloseEvent = CloseEvent as typeof globalThis.CloseEvent;
}

// Diagnostic: Check if document is defined
if (typeof document === 'undefined') {
  // @ts-ignore - We're checking if document is defined
  globalThis.document = undefined;
  // eslint-disable-next-line no-console
  console.warn('[setup.ts] Warning: document is not defined at setup time. jsdom may not be initialized.');
}

import { expect, vi, afterEach, beforeAll, afterAll } from 'vitest';
import { cleanup } from '@testing-library/react';
import * as matchers from '@testing-library/jest-dom/matchers';

// Extend Vitest's expect with jest-dom matchers
expect.extend(matchers);

// WebSocket implementation for testing
class MockWebSocket extends EventTarget {
  // Static properties to match WebSocket interface
  static readonly CONNECTING: 0 = 0;
  static readonly OPEN: 1 = 1;
  static readonly CLOSING: 2 = 2;
  static readonly CLOSED: 3 = 3;

  // WebSocket properties
  binaryType: BinaryType = 'blob';
  bufferedAmount = 0;
  extensions = '';
  onclose: ((this: WebSocket, ev: CloseEvent) => any) | null = null;
  onerror: ((this: WebSocket, ev: Event) => any) | null = null;
  onmessage: ((this: WebSocket, ev: MessageEvent) => any) | null = null;
  onopen: ((this: WebSocket, ev: Event) => any) | null = null;
  protocol = '';
  readyState: number = MockWebSocket.CONNECTING;
  url: string;
  
  // Cast this to WebSocket for event handlers
  private get webSocketThis(): WebSocket {
    return this as unknown as WebSocket;
  }
  
  // Mock implementation
  private mockSend = vi.fn();
  private mockClose = vi.fn();
  private mockAddEventListener = vi.fn();
  private mockRemoveEventListener = vi.fn();
  private mockDispatchEvent = vi.fn();

  constructor(url: string | URL, _protocols?: string | string[]) {
    super();
    this.url = url.toString();
    
    // Auto-connect after a small delay
    setTimeout(() => {
      this.readyState = MockWebSocket.OPEN;
      const openEvent = new Event('open');
      this.dispatchEvent(openEvent);
      if (this.onopen) {
        this.onopen.call(this.webSocketThis, openEvent);
      }
    }, 0);
  }

  // WebSocket methods
  send(data: string | ArrayBufferLike | Blob | ArrayBufferView): void {
    this.mockSend(data);
    if (this.readyState !== MockWebSocket.OPEN) {
      throw new Error('WebSocket is not open');
    }
  }

  close(code?: number, reason?: string): void {
    this.mockClose(code, reason);
    this.readyState = MockWebSocket.CLOSED;
    if (this.onclose) {
      const closeEvent = new CloseEvent('close', { 
        code: code || 1000, 
        reason: reason || '',
        wasClean: true 
      });
      this.dispatchEvent(closeEvent);
      this.onclose.call(this.webSocketThis, closeEvent);
    }
  }

  // Test helpers
  simulateMessage(data: unknown): void {
    if (this.onmessage) {
      const messageData = typeof data === 'string' ? data : JSON.stringify(data);
      const messageEvent = new MessageEvent('message', { data: messageData });
      this.dispatchEvent(messageEvent);
      this.onmessage.call(this.webSocketThis, messageEvent);
    }
  }

  simulateError(error?: Error): void {
    if (this.onerror) {
      const errorObj = error || new Error('WebSocket error');
      const errorEvent = new ErrorEvent('error', {
        message: errorObj.message,
        error: errorObj,
        bubbles: false,
        cancelable: true
      });
      this.dispatchEvent(errorEvent);
      this.onerror.call(this.webSocketThis, errorEvent);
    }
  }

  simulateClose(code = 1000, reason = ''): void {
    this.readyState = MockWebSocket.CLOSED;
    if (this.onclose) {
      const closeEvent = new CloseEvent('close', { 
        code, 
        reason,
        wasClean: true 
      });
      this.dispatchEvent(closeEvent);
      this.onclose.call(this.webSocketThis, closeEvent);
    }
  }
  
  // Override EventTarget methods to track calls
  override addEventListener(
    type: string, 
    listener: EventListenerOrEventListenerObject | null, 
    options?: boolean | AddEventListenerOptions
  ): void {
    this.mockAddEventListener(type, listener, options);
    super.addEventListener(type, listener, options);
  }
  
  override removeEventListener(
    type: string, 
    listener: EventListenerOrEventListenerObject | null, 
    options?: boolean | EventListenerOptions
  ): void {
    this.mockRemoveEventListener(type, listener, options);
    super.removeEventListener(type, listener, options);
  }
  
  override dispatchEvent(event: Event): boolean {
    this.mockDispatchEvent(event);
    return super.dispatchEvent(event);
  }
}

// Extend the global WebSocket interface to include our test methods
declare global {
  interface WebSocket {
    simulateMessage: (data: unknown) => void;
    simulateError: (error?: Error) => void;
    simulateClose: (code?: number, reason?: string) => void;
  }
}

// Extend the global WebSocket interface with our test methods
declare global {
  interface WebSocket {
    simulateMessage: (data: unknown) => void;
    simulateError: (error?: Error) => void;
    simulateClose: (code?: number, reason?: string) => void;
  }
}

// Assign our mock WebSocket to the global scope
global.WebSocket = MockWebSocket as any as typeof WebSocket;

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
