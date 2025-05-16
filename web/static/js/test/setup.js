import { expect, vi, beforeEach, afterEach } from 'vitest';

// Mock WebSocket implementation
class MockWebSocket {
  static CONNECTING = 0;
  static OPEN = 1;
  static CLOSING = 2;
  static CLOSED = 3;

  constructor(url) {
    this.url = url;
    this.readyState = MockWebSocket.CONNECTING;
    this.binaryType = 'arraybuffer';
    this.bufferedAmount = 0;
    this.extensions = '';
    this.protocol = '';
    
    // Event handlers
    this.onopen = null;
    this.onmessage = null;
    this.onclose = null;
    this.onerror = null;
    
    // Mock connection
    setTimeout(() => {
      this.readyState = MockWebSocket.OPEN;
      if (this.onopen) this.onopen(new Event('open'));
    }, 10);
  }
  
  send(data) {
    if (this.readyState !== MockWebSocket.OPEN) {
      throw new Error('WebSocket is not open');
    }
    // No-op for mock
  }
  
  close(code = 1000, reason) {
    if (this.readyState === MockWebSocket.CLOSED || this.readyState === MockWebSocket.CLOSING) {
      return;
    }
    
    this.readyState = MockWebSocket.CLOSING;
    
    setTimeout(() => {
      this.readyState = MockWebSocket.CLOSED;
      if (this.onclose) {
        this.onclose(new CloseEvent('close', { 
          code, 
          reason: reason || '',
          wasClean: true 
        }));
      }
    }, 10);
  }
  
  // Test helper methods
  _triggerMessage(data) {
    if (this.onmessage) {
      const messageEvent = new MessageEvent('message', { data });
      this.onmessage(messageEvent);
    }
  }
  
  _triggerError() {
    if (this.onerror) {
      this.onerror(new Event('error'));
    }
  }
}

// Global mocks and setup
beforeEach(() => {
  // Reset any mocks or setup before each test
  vi.clearAllMocks();
  
  // Reset the DOM before each test
  document.body.innerHTML = '';
  
  // Mock WebSocket
  global.WebSocket = vi.fn().mockImplementation((url) => new MockWebSocket(url));
  
  // Mock HTMX if needed
  global.htmx = {
    process: vi.fn(),
    on: vi.fn(),
    off: vi.fn(),
    find: vi.fn(() => document.body),
    findAll: vi.fn(() => [document.body]),
    remove: vi.fn(),
    addClass: vi.fn(),
    removeClass: vi.fn(),
    trigger: vi.fn(),
    ws: vi.fn()
  };
  
  // Mock console methods
  global.console = {
    ...console,
    error: vi.fn(),
    warn: vi.fn(),
    log: vi.fn(),
    info: vi.fn(),
    debug: vi.fn()
  };
});

afterEach(() => {
  // Cleanup after each test
  vi.restoreAllMocks();
  delete global.WebSocket;
  delete global.htmx;
});

// Make expect and other test utilities available globally
global.expect = expect;
global.vi = vi;

// Mock timers for testing timeouts
beforeEach(() => {
  vi.useFakeTimers();
});

afterEach(() => {
  vi.useRealTimers();
});

// Helper function to wait for promises to resolve
global.flushPromises = () => new Promise(setImmediate);

// Helper to wait for a condition to be true
global.waitFor = async (condition, timeout = 1000, interval = 10) => {
  const start = Date.now();
  while (Date.now() - start < timeout) {
    if (await condition()) return true;
    await new Promise(resolve => setTimeout(resolve, interval));
  }
  throw new Error(`Condition not met within ${timeout}ms`);
};
