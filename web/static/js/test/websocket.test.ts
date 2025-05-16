import { describe, it, expect, beforeAll, beforeEach, afterEach, vi } from 'vitest';
import { WebSocketClient } from '../src/websocket';

// Mock Toastify
vi.mock('../src/toast', () => ({
  showToast: vi.fn()
}));

// Global types are now handled through type assertions in the test code

// Setup basic mock environment for browser globals
beforeAll(() => {
  // Mock HTMX
  const mockHtmx = {
    process: vi.fn(),
    on: vi.fn(),
    trigger: vi.fn(),
    find: vi.fn(),
    ajax: vi.fn()
  };

  // Mock window with type assertion
  (globalThis as any).window = {
    htmx: mockHtmx,
    Toastify: vi.fn().mockImplementation(() => ({
      showToast: vi.fn()
    })),
    WebSocket: class MockWebSocket {
      static CONNECTING = 0;
      static OPEN = 1;
      static CLOSING = 2;
      static CLOSED = 3;

      url: string;
      onopen: ((this: WebSocket, ev: Event) => any) | null = null;
      onclose: ((this: WebSocket, ev: CloseEvent) => any) | null = null;
      onerror: ((this: WebSocket, ev: Event) => any) | null = null;
      onmessage: ((this: WebSocket, ev: MessageEvent) => any) | null = null;
      readyState: number = WebSocket.CONNECTING;
      binaryType: BinaryType = 'blob';
      bufferedAmount: number = 0;
      extensions: string = '';
      protocol: string = '';

      constructor(url: string | URL, _protocols?: string | string[] | undefined) {
        this.url = url.toString();
      }

      send(_data: string | ArrayBufferLike | Blob | ArrayBufferView): void {
        // Mock implementation
      }

      close(_code?: number, _reason?: string): void {
        this.readyState = WebSocket.CLOSED;
      }
    }
  } as unknown as Window & typeof globalThis;
});

describe('WebSocketClient', () => {
  let wsClient: WebSocketClient;
  let mockWebSocket: any;
  const testUrl = 'ws://test-server/socket';

  beforeEach(() => {
    // Create a new WebSocketClient for each test
    wsClient = new WebSocketClient();
    
    // Spy on the WebSocket constructor
    mockWebSocket = vi.spyOn(globalThis, 'WebSocket');
    
    // Mock the WebSocket instance methods
    const mockSocket = {
      send: vi.fn(),
      close: vi.fn(),
      readyState: WebSocket.CONNECTING,
      onopen: null as ((this: WebSocket, ev: Event) => any) | null,
      onclose: null as ((this: WebSocket, ev: CloseEvent) => any) | null,
      onerror: null as ((this: WebSocket, ev: Event) => any) | null,
      onmessage: null as ((this: WebSocket, ev: MessageEvent) => any) | null,
    };
    
    // Make the mock return our mock socket
    mockWebSocket.mockImplementation(() => mockSocket);
    
    // Return the mock socket for assertions
    return mockSocket;
  });

  afterEach(() => {
    // Clear all mocks
    vi.clearAllMocks();
  });

  describe('connection', () => {
    it('should connect to the WebSocket server', () => {
      wsClient.connect(testUrl);
      expect(mockWebSocket).toHaveBeenCalledWith(testUrl);
    });

    it('should handle connection open', () => {
      const onOpen = vi.fn();
      const removeHandler = wsClient.onConnect(onOpen);
      
      // Connect and get the mock WebSocket instance
      const mockSocket = wsClient.connect(testUrl);
      
      // Simulate the WebSocket opening
      if (mockSocket.onopen) {
        mockSocket.onopen(new Event('open'));
      }
      
      expect(onOpen).toHaveBeenCalled();
      
      // Clean up
      removeHandler();
    });

    it('should handle connection close', () => {
      const onClose = vi.fn();
      const removeHandler = wsClient.onDisconnect(onClose);
      
      // Connect and get the mock WebSocket instance
      const mockSocket = wsClient.connect(testUrl);
      
      // Simulate the WebSocket closing
      if (mockSocket.onclose) {
        mockSocket.onclose(new CloseEvent('close'));
      }
      
      expect(onClose).toHaveBeenCalled();
      
      // Clean up
      removeHandler();
    });
  });

  describe('message handling', () => {
    it('should send messages', () => {
      // Connect and get the mock WebSocket instance
      const mockSocket = wsClient.connect(testUrl);
      
      // Send a test message
      const testMessage = { type: 'test', data: 'Hello, WebSocket!' };
      const result = wsClient.send(testMessage);
      
      // Check if send was called with the correct message
      expect(mockSocket.send).toHaveBeenCalledWith(JSON.stringify(testMessage));
      expect(result).toBe(true);
    });

    it('should receive and handle messages', () => {
      const testMessage = { type: 'test', data: 'Hello, WebSocket!' };
      const messageHandler = vi.fn();
      
      // Set up the message handler
      const removeHandler = wsClient.onMessage(messageHandler);
      
      // Connect and get the mock WebSocket instance
      const mockSocket = wsClient.connect(testUrl);
      
      // Simulate a message
      if (mockSocket.onmessage) {
        mockSocket.onmessage(new MessageEvent('message', {
          data: JSON.stringify(testMessage)
        }));
      }
      
      expect(messageHandler).toHaveBeenCalledWith(testMessage);
      
      // Clean up
      removeHandler();
    });
  });

  describe('error handling', () => {
    it('should handle WebSocket errors', () => {
      const errorHandler = vi.fn();
      const removeHandler = wsClient.onError(errorHandler);
      
      // Connect and get the mock WebSocket instance
      const mockSocket = wsClient.connect(testUrl);
      
      // Simulate an error
      const errorEvent = new Event('error');
      if (mockSocket.onerror) {
        mockSocket.onerror(errorEvent);
      }
      
      expect(errorHandler).toHaveBeenCalledWith(errorEvent);
      
      // Clean up
      removeHandler();
    });
  });

  describe('disconnection', () => {
    it('should close the WebSocket connection', () => {
      // Connect and get the mock WebSocket instance
      const mockSocket = wsClient.connect(testUrl);
      
      // Disconnect
      wsClient.disconnect();
      
      // Check if close was called
      expect(mockSocket.close).toHaveBeenCalled();
    });
  });
});
