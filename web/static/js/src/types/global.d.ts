// Global TypeScript declarations for Go Sentinel

// Declare types for global objects
declare const __APP_VERSION__: string;
declare const __DEV__: boolean;
declare const __TEST__: boolean;
// Vitest globals
declare const describe: (name: string, fn: () => void) => void;
declare const it: (name: string, fn: (done?: () => void) => void | Promise<void>, timeout?: number) => void;
declare const test: typeof it;
declare const expect: any; // Using 'any' as the actual type is complex and provided by Vitest
declare const beforeAll: (fn: () => void | Promise<void>, timeout?: number) => void;
declare const afterAll: (fn: () => void | Promise<void>, timeout?: number) => void;
declare const beforeEach: (fn: () => void | Promise<void>, timeout?: number) => void;
declare const afterEach: (fn: () => void | Promise<void>, timeout?: number) => void;
declare const vi: any; // Vitest's mocking utilities

// Declare types for modules without type definitions
declare module '*.css' {
  const content: { [className: string]: string };
  export default content;
}

declare module '*.svg' {
  import * as React from 'react';
  export const ReactComponent: React.FunctionComponent<React.SVGProps<SVGSVGElement>>;
  const src: string;
  export default src;
}

// Extend the Window interface to include any global browser APIs or custom properties
interface Window extends Window {
  // Test environment globals
  __VITEST_BROWSER_PROVIDER__?: any;
  __VITEST_COVERAGE__?: any;
  __VITEST_MOCKS__?: Record<string, any>;
  __VITEST_RESULT__?: any;
  __VITEST_SEQUENCE__?: any;
  __VITEST_SNAPSHOT_CLIENT__?: any;
  __VITEST_WORKER__?: boolean;
  __VITEST_WORKER_IDS__?: number[];
  __VITEST_WORKER_PATH__?: string;
  __VITEST_WORKER_POOL__?: boolean;
  __VITEST_WORKER_INDEX__?: number;
  __VITEST_WORKER_COUNT__?: number;
  __VITEST_WORKER_FILE__?: string;
  __VITEST_WORKER_THREADS__?: boolean;
  __VITEST_WORKER_IPC__?: any;
  __VITEST_WORKER_RPC__?: any;
  __VITEST_WORKER_WS__?: any;
  __VITEST_WORKER_TIMER__?: any;
  __VITEST_WORKER_TIMERS__?: any;
  __VITEST_WORKER_EVENTS__?: any;
  __VITEST_WORKER_MODULE_CACHE__?: any;
  __VITEST_WORKER_MODULE_LOADER__?: any;
  __VITEST_WORKER_MODULE_MOCK__?: any;
  __VITEST_WORKER_MODULE_MOCK_FACTORY__?: any;
  __VITEST_WORKER_MODULE_MOCK_UTILS__?: any;
  __VITEST_WORKER_MODULE_UTILS__?: any;
  __VITEST_WORKER_RUNNER__?: any;
  __VITEST_WORKER_RUNNER_UTILS__?: any;
  __VITEST_WORKER_SNAPSHOT__?: any;
  __VITEST_WORKER_SNAPSHOT_CLIENT__?: any;
  __VITEST_WORKER_SNAPSHOT_SERVER__?: any;
  __VITEST_WORKER_SNAPSHOT_UTILS__?: any;
  __VITEST_WORKER_TEST_UTILS__?: any;
  __VITEST_WORKER_TYPES__?: any;
  __VITEST_WORKER_UTILS__?: any;
  __VITEST_WORKER_VITEST__?: any;
  __VITEST_WORKER_VITEST_UTILS__?: any;
  __VITEST_WORKER_WEBRTC__?: any;
  __VITEST_WORKER_WS__?: any;
  __VITEST_WORKER_WS_CLIENT__?: any;
  __VITEST_WORKER_WS_SERVER__?: any;
  __VITEST_WORKER_WS_UTILS__?: any;
  __VITEST_WORKER_WS_WEBSOCKET__?: any;
  __VITEST_WORKER_WS_WEBSOCKET_SERVER__?: any;
  __VITEST_WORKER_WS_WEBSOCKET_UTILS__?: any;
  __VITEST_WORKER_WS_WEBSOCKET_WEBSOCKET__?: any;
  __VITEST_WORKER_WS_WEBSOCKET_WEBSOCKET_SERVER__?: any;
  __VITEST_WORKER_WS_WEBSOCKET_WEBSOCKET_UTILS__?: any;
  __VITEST_WORKER_WS_WEBSOCKET_WEBSOCKET_WEBSOCKET__?: any;
  __VITEST_WORKER_WS_WEBSOCKET_WEBSOCKET_WEBSOCKET_SERVER__?: any;
  __VITEST_WORKER_WS_WEBSOCKET_WEBSOCKET_WEBSOCKET_UTILS__?: any;
  __VITEST_WORKER_WS_WEBSOCKET_WEBSOCKET_WEBSOCKET_WEBSOCKET__?: any;
  __VITEST_WORKER_WS_WEBSOCKET_WEBSOCKET_WEBSOCKET_WEBSOCKET_SERVER__?: any;
  __VITEST_WORKER_WS_WEBSOCKET_WEBSOCKET_WEBSOCKET_WEBSOCKET_UTILS__?: any;
  __VITEST_WORKER_WS_WEBSOCKET_WEBSOCKET_WEBSOCKET_WEBSOCKET_WEBSOCKET__?: any;
  __VITEST_WORKER_WS_WEBSOCKET_WEBSOCKET_WEBSOCKET_WEBSOCKET_WEBSOCKET_SERVER__?: any;
  __VITEST_WORKER_WS_WEBSOCKET_WEBSOCKET_WEBSOCKET_WEBSOCKET_WEBSOCKET_UTILS__?: any;
  
  // Test utilities
  mockClearAllMocks: () => void;
  mockResetAllMocks: () => void;
  mockRestoreAllMocks: () => void;
  mockFn: <T extends (...args: any[]) => any>(fn?: T) => jest.Mock<ReturnType<T>, Parameters<T>>;
  
  // Testing Library utilities
  getByTestId: (testId: string) => HTMLElement;
  getByText: (text: string | RegExp, selector?: string) => HTMLElement;
  getByLabelText: (text: string | RegExp, options?: any) => HTMLElement;
  getByPlaceholderText: (text: string | RegExp) => HTMLElement;
  getByRole: (role: string, options?: any) => HTMLElement;
  queryByTestId: (testId: string) => HTMLElement | null;
  queryByText: (text: string | RegExp, selector?: string) => HTMLElement | null;
  queryByLabelText: (text: string | RegExp, options?: any) => HTMLElement | null;
  queryByPlaceholderText: (text: string | RegExp) => HTMLElement | null;
  queryByRole: (role: string, options?: any) => HTMLElement | null;
  
  // WebSocket mocks for testing
  mockWebSocket: {
    clear: () => void;
    instance: () => WebSocket;
    instances: () => WebSocket[];
    mockClear: () => void;
    mockReset: () => void;
    mockRestore: () => void;
    mockImplementation: (impl: any) => void;
    mockImplementationOnce: (impl: any) => void;
    mockReturnThis: () => void;
    mockReturnValue: (value: any) => void;
    mockReturnValueOnce: (value: any) => void;
    mockResolvedValue: (value: any) => void;
    mockResolvedValueOnce: (value: any) => void;
    mockRejectedValue: (value: any) => void;
    mockRejectedValueOnce: (value: any) => void;
  };
  
  // Mock timers
  mockTimers: {
    useFakeTimers: () => void;
    useRealTimers: () => void;
    runAllTimers: () => void;
    runOnlyPendingTimers: () => void;
    runTimersToTime: (ms: number) => void;
    setSystemTime: (now?: number | Date) => void;
    getTimerCount: () => number;
  };
  // Add any global browser APIs or custom properties here
  htmx?: {
    defineExtension?: (name: string, extension: any) => void;
    process?: (element: HTMLElement) => void;
    on?: (event: string, handler: (event: CustomEvent) => void) => void;
    off?: (event: string, handler: (event: CustomEvent) => void) => void;
    trigger?: (element: HTMLElement, event: string, detail?: any) => void;
    find?: (selector: string) => HTMLElement | null;
    findAll?: (selector: string) => NodeListOf<HTMLElement>;
    ajax?: (method: string, url: string, config: any) => void;
  };
  
  // WebSocket client for Go Sentinel
  goSentinelWebSocket?: {
    connect: (url: string) => void;
    disconnect: () => void;
    send: (data: any) => void;
    on: (event: string, callback: (data: any) => void) => void;
    off: (event: string, callback: (data: any) => void) => void;
  };

  // Coverage visualization functions
  showFileDetails?: (filePath?: string) => void;
  getCoverageClass?: (percentage: number) => string;
  goToPage?: (page: number, filter?: string, search?: string) => void;
}

// Declare types for any global variables or functions
declare const toast: {
  success: (message: string, options?: any) => void;
  error: (message: string, options?: any) => void;
  info: (message: string, options?: any) => void;
  warning: (message: string, options?: any) => void;
};

// Declare types for any custom events
declare namespace CustomEventMap {
  interface WebSocketMessageEvent extends CustomEvent {
    detail: {
      type: string;
      payload: any;
    };
  }
  // Add more custom event types as needed
}

// Extend Jest types for Vitest compatibility
declare namespace jest {
  interface Mock<T = any, Y extends any[] = any> {
    (...args: Y): T;
    mock: MockContext<T, Y>;
    mockClear(): void;
    mockReset(): void;
    mockRestore(): void;
    mockImplementation(fn: (...args: Y) => T): this;
    mockImplementationOnce(fn: (...args: Y) => T): this;
    mockName(name: string): this;
    mockReturnThis(): this;
    mockReturnValue(value: T): this;
    mockReturnValueOnce(value: T): this;
    mockResolvedValue(value: Awaited<T>): this;
    mockResolvedValueOnce(value: Awaited<T>): this;
    mockRejectedValue(value: any): this;
    mockRejectedValueOnce(value: any): this;
    getMockName(): string;
    mockReturnThis(): this;
  }

  interface MockContext<T, Y extends any[]> {
    calls: Y[];
    instances: T[];
    invocationCallOrder: number[];
    results: Array<{
      type: 'return' | 'throw';
      value: T;
    }>;
    lastCall: Y | undefined;
  }
}

declare global {
  // Extend the global EventTarget interface to include our custom events
  interface WindowEventMap extends CustomEventMap {}
  interface HTMLElementEventMap extends CustomEventMap {}
  
  // Add any global utility types here
  type Nullable<T> = T | null;
  type Optional<T> = T | undefined;
  type Dictionary<T> = Record<string, T>;
  
  // Add any global utility functions here
  function debounce<T extends (...args: any[]) => any>(
    func: T,
    wait: number,
    immediate?: boolean
  ): (...args: Parameters<T>) => void;
}
