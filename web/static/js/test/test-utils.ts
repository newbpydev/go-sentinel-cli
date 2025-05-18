// Test utilities for Go Sentinel

/**
 * Creates a mock WebSocket instance for testing
 */
export function createMockWebSocket(): WebSocket {
  const mockWebSocket = {
    readyState: 0, // CONNECTING
    url: '',
    onopen: null as ((this: WebSocket, ev: Event) => any) | null,
    onclose: null as ((this: WebSocket, ev: CloseEvent) => any) | null,
    onerror: null as ((this: WebSocket, ev: Event) => any) | null,
    onmessage: null as ((this: WebSocket, ev: MessageEvent) => any) | null,
    close: vi.fn(),
    send: vi.fn(),
    addEventListener: vi.fn((event: string, callback: any) => {
      // Simplified event listener for testing
      if (event === 'open') {
        mockWebSocket.onopen = callback;
      } else if (event === 'close') {
        mockWebSocket.onclose = callback;
      } else if (event === 'error') {
        mockWebSocket.onerror = callback;
      } else if (event === 'message') {
        mockWebSocket.onmessage = callback;
      }
    }),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
    binaryType: 'blob' as BinaryType,
    bufferedAmount: 0,
    extensions: '',
    protocol: '',
    CONNECTING: 0,
    OPEN: 1,
    CLOSING: 2,
    CLOSED: 3,
  };

  return mockWebSocket as unknown as WebSocket;
}

/**
 * Waits for a condition to be true
 * @param condition Function that returns a boolean or a promise that resolves to a boolean
 * @param timeout Timeout in milliseconds (default: 1000ms)
 * @param interval Interval to check the condition (default: 50ms)
 */
export async function waitFor(
  condition: () => boolean | Promise<boolean>,
  timeout = 1000,
  interval = 50
): Promise<boolean> {
  const start = Date.now();
  
  while (Date.now() - start < timeout) {
    const result = await Promise.resolve(condition());
    if (result) return true;
    await new Promise(resolve => setTimeout(resolve, interval));
  }
  
  return false;
}

/**
 * Mocks the WebSocket global for testing
 * @returns Object with mock functions and utilities
 */
export function mockWebSocketGlobal() {
  const originalWebSocket = globalThis.WebSocket;
  const mockWebSocket = createMockWebSocket();
  
  // @ts-ignore - Mocking WebSocket
  globalThis.WebSocket = vi.fn().mockImplementation(() => mockWebSocket);
  
  // Add mockWebSocket to the global scope for test access
  (globalThis as any).mockWebSocket = mockWebSocket;
  
  return {
    mockWebSocket,
    restore() {
      globalThis.WebSocket = originalWebSocket;
      delete (globalThis as any).mockWebSocket;
    },
  };
}

/**
 * Creates a test event with the specified type and data
 * @param type The event type
 * @param data The event data
 * @returns A custom event
 */
export function createTestEvent<T = any>(type: string, data?: T): CustomEvent<T> {
  return new CustomEvent(type, { detail: data });
}

/**
 * Mocks the window.matchMedia function for testing
 */
export function mockMatchMedia() {
  Object.defineProperty(window, 'matchMedia', {
    writable: true,
    value: vi.fn().mockImplementation((query: string) => ({
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
}

/**
 * Mocks the window.localStorage for testing
 */
export function mockLocalStorage() {
  const localStorageMock = (() => {
    let store: Record<string, string> = {};
    return {
      getItem: vi.fn((key: string) => store[key] || null),
      setItem: vi.fn((key: string, value: string) => {
        store[key] = String(value);
      }),
      removeItem: vi.fn((key: string) => {
        delete store[key];
      }),
      clear: vi.fn(() => {
        store = {};
      }),
      key: vi.fn((index: number) => Object.keys(store)[index] || null),
      get length() {
        return Object.keys(store).length;
      },
    };
  })();

  Object.defineProperty(window, 'localStorage', {
    value: localStorageMock,
    writable: true,
  });

  return localStorageMock;
}

/**
 * Mocks the window.sessionStorage for testing
 */
export function mockSessionStorage() {
  const sessionStorageMock = (() => {
    let store: Record<string, string> = {};
    return {
      getItem: vi.fn((key: string) => store[key] || null),
      setItem: vi.fn((key: string, value: string) => {
        store[key] = String(value);
      }),
      removeItem: vi.fn((key: string) => {
        delete store[key];
      }),
      clear: vi.fn(() => {
        store = {};
      }),
      key: vi.fn((index: number) => Object.keys(store)[index] || null),
      get length() {
        return Object.keys(store).length;
      },
    };
  })();

  Object.defineProperty(window, 'sessionStorage', {
    value: sessionStorageMock,
    writable: true,
  });

  return sessionStorageMock;
}

/**
 * Mocks the window.requestAnimationFrame for testing
 */
export function mockRequestAnimationFrame() {
  const originalRequestAnimationFrame = window.requestAnimationFrame;
  const originalCancelAnimationFrame = window.cancelAnimationFrame;
  
  const mockRequestAnimationFrame = vi.fn((callback: FrameRequestCallback) => {
    const id = setTimeout(() => callback(performance.now()));
    return id as unknown as number;
  });
  
  const mockCancelAnimationFrame = vi.fn((id: number) => {
    clearTimeout(id);
  });
  
  window.requestAnimationFrame = mockRequestAnimationFrame;
  window.cancelAnimationFrame = mockCancelAnimationFrame;
  
  return {
    mockRequestAnimationFrame,
    mockCancelAnimationFrame,
    restore() {
      window.requestAnimationFrame = originalRequestAnimationFrame;
      window.cancelAnimationFrame = originalCancelAnimationFrame;
    },
  };
}
