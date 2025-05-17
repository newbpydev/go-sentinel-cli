import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest';

// Import the WebSocketClient class, not the singleton instance
// Import using require to ensure we get the actual runtime implementation
const { WebSocketClient } = require('../src/websocket');

// Type declaration for the JS implementation
type WebSocketClientType = {
  connect(url: string): Promise<WebSocket>;
  disconnect(): void;
  close(): void;
  send(data: any): boolean;
  on(event: string, handler: Function): () => void;
  off(event: string, handler: Function): boolean;
  onMessage(messageType: string, handler: Function): () => void;
}

// Explicitly don't mock the WebSocketClient
vi.mock('../src/toast', () => ({
  showToast: vi.fn()
}));

// Use constants from the WebSocket API
const WS_CONNECTING = 0;
const WS_OPEN = 1;
const WS_CLOSING = 2;
const WS_CLOSED = 3;

describe('WebSocketClient', () => {
  let wsClient: WebSocketClientType; // Using our type definition
  let mockWs: any;
  const testUrl = 'ws://test-server/socket';
  
  beforeEach(() => {
    // Reset all mocks
    vi.clearAllMocks();
    
    // Create mockWs with spies for the WebSocket methods
    mockWs = {
      send: vi.fn(),
      close: vi.fn(),
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      readyState: WS_CONNECTING,
      url: testUrl,
      binaryType: 'blob'
    };
    
    // Mock the global WebSocket
    global.WebSocket = vi.fn().mockImplementation(() => mockWs) as any;
    Object.defineProperty(global.WebSocket, 'CONNECTING', { value: WS_CONNECTING });
    Object.defineProperty(global.WebSocket, 'OPEN', { value: WS_OPEN });
    Object.defineProperty(global.WebSocket, 'CLOSING', { value: WS_CLOSING });
    Object.defineProperty(global.WebSocket, 'CLOSED', { value: WS_CLOSED });
    
    // Setup the browser window mock
    global.window = {
      Toastify: vi.fn().mockImplementation(() => ({
        showToast: vi.fn()
      }))
    } as any;
    
    // Create a new instance of WebSocketClient for each test
    wsClient = new WebSocketClient();
  });
  
  describe('connect method', () => {
    it('should create a WebSocket with the given URL', () => {
      // When
      wsClient.connect(testUrl);
      
      // Then
      expect(WebSocket).toHaveBeenCalledWith(testUrl);
    });
  });
  
  describe('event handlers', () => {
    beforeEach(() => {
      // Connect to ensure the socket is initialized, but don't await it
      // since we're mocking the implementation
      wsClient.connect(testUrl);
      
      // Resolve the promise by simulating open event
      mockWs.onopen && mockWs.onopen(new Event('open'));
    });
    
    afterEach(() => {
      vi.clearAllMocks();
    });
    
    it('should register open event handlers', () => {
      // Given
      const openHandler = vi.fn();
      const removeHandler = wsClient.on('open', openHandler);
      
      // When - simulate socket open event
      mockWs.onopen && mockWs.onopen(new Event('open'));
      
      // Then
      expect(openHandler).toHaveBeenCalled();
      
      // Cleanup
      removeHandler();
    });
    
    it('should register close event handlers', () => {
      // Given
      const closeHandler = vi.fn();
      const removeHandler = wsClient.on('close', closeHandler);
      
      // When - simulate socket close event
      mockWs.onclose && mockWs.onclose(new CloseEvent('close'));
      
      // Then
      expect(closeHandler).toHaveBeenCalled();
      
      // Cleanup
      removeHandler();
    });
    
    it('should register and handle message events', () => {
      // Given
      const messageHandler = vi.fn();
      const testMessage = { type: 'test', data: 'message' };
      
      // Register a message handler for the specific message type
      const removeHandler = wsClient.onMessage('test', messageHandler);
      
      // When - simulate message event
      mockWs.onmessage && mockWs.onmessage(new MessageEvent('message', {
        data: JSON.stringify(testMessage)
      }));
      
      // Then
      expect(messageHandler).toHaveBeenCalled();
      
      // Cleanup
      removeHandler();
    });
    
    it('should register error event handlers', () => {
      // Given
      const errorHandler = vi.fn();
      const removeHandler = wsClient.on('error', errorHandler);
      
      // When - simulate error event
      mockWs.onerror && mockWs.onerror(new Event('error'));
      
      // Then
      expect(errorHandler).toHaveBeenCalled();
      
      // Cleanup
      removeHandler();
    });
  });
  
  describe('send method', () => {
    // Setup the test with a shorter timeout
    beforeEach(() => {
      // Connect without awaiting
      wsClient.connect(testUrl);
      // Mock the WebSocket to be open
      mockWs.readyState = WS_OPEN;
      // Manually resolve the connect promise
      mockWs.onopen && mockWs.onopen(new Event('open'));
    });
    
    it('should send JSON-stringified data when socket is open', async () => {
      // Given socket is open
      mockWs.readyState = WS_OPEN;
      
      // When
      const testData = { type: 'test', payload: 'data' };
      const resultPromise = wsClient.send(testData);
      
      // Then
      expect(mockWs.send).toHaveBeenCalledWith(JSON.stringify(testData));
      
      // Wait for the promise to resolve and check its value
      const result = await resultPromise;
      expect(result).toBe(true);
    });
    
    it('should not send data when socket is not open', async () => {
      // Given socket is connecting (not open)
      mockWs.readyState = WS_CONNECTING;
      
      // When
      const resultPromise = wsClient.send({ type: 'test' });
      
      // Then
      expect(mockWs.send).not.toHaveBeenCalled();
      
      // Wait for the promise to resolve and check its value
      const result = await resultPromise;
      expect(result).toBe(false);
    });
  });
  
  describe('disconnect method', () => {
    // Testing the close functionality without waiting for promises
    it('should close the WebSocket connection', () => {
      // Given
      wsClient.connect(testUrl);
      // Directly trigger the mock WebSocket's onopen handler
      mockWs.onopen && mockWs.onopen(new Event('open'));
      
      // When
      // Try both methods - in the actual implementation, one might be an alias for the other
      wsClient.disconnect ? wsClient.disconnect() : wsClient.close();
      
      // Then
      expect(mockWs.close).toHaveBeenCalled();
    });
  });
});
