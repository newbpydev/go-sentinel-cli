import { describe, it, expect, beforeAll, beforeEach, afterEach, afterAll, vi } from 'vitest';
import { WebSocketClient } from '../websocket';

// Mock Toastify
vi.mock('../toast', () => ({
  showToast: vi.fn()
}));

// Setup basic mock environment for browser globals
beforeAll(() => {
  // Mock HTMX
  global.htmx = {
    on: vi.fn(),
    trigger: vi.fn(),
    find: vi.fn(),
    ajax: vi.fn()
  };
  
  // Mock Toastify
  global.Toastify = vi.fn().mockImplementation(() => ({
    showToast: vi.fn()
  }));
});

describe('WebSocketClient', () => {
  let wsClient;
  let mockWebSocket;
  const testUrl = 'ws://test-server/socket';
  let originalWebSocket;

  // Save original WebSocket implementation
  beforeAll(() => {
    originalWebSocket = global.WebSocket;
  });

  // Restore original WebSocket implementation
  afterAll(() => {
    global.WebSocket = originalWebSocket;
  });

  beforeEach(() => {
    // Start with a clean state
    vi.clearAllMocks();
    
    // Create a mock WebSocket implementation
    mockWebSocket = {
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      send: vi.fn(),
      close: vi.fn(),
      readyState: WebSocket.OPEN
    };
    
    // Mock the WebSocket constructor but preserve constants
    global.WebSocket = vi.fn().mockImplementation(() => mockWebSocket);
    
    // Create a new WebSocketClient instance
    wsClient = new WebSocketClient();
  });

  afterEach(() => {
    // Clean up resources
    if (wsClient && typeof wsClient.close === 'function') {
      wsClient.close();
    }
    wsClient = null;
    mockWebSocket = null;
    vi.restoreAllMocks();
  });

  describe('connection handling', () => {
    it('should connect to WebSocket server', async () => {
      // Setup event handler capture mechanism
      const handlers = {};
      mockWebSocket.addEventListener.mockImplementation((event, handler) => {
        handlers[event] = handler;
      });
      
      // Begin connection process
      const connectionPromise = wsClient.connect(testUrl);
      
      // Manually trigger the 'open' event
      if (handlers.open) {
        handlers.open({ type: 'open' });
      }
      
      // Wait for the connection to complete
      await connectionPromise;
      
      // Verify connection was established properly
      expect(global.WebSocket).toHaveBeenCalledWith(testUrl);
      expect(wsClient.socket).toBe(mockWebSocket);
      expect(mockWebSocket.addEventListener).toHaveBeenCalledWith('open', expect.any(Function));
      expect(mockWebSocket.addEventListener).toHaveBeenCalledWith('message', expect.any(Function));
      expect(mockWebSocket.addEventListener).toHaveBeenCalledWith('close', expect.any(Function));
      expect(mockWebSocket.addEventListener).toHaveBeenCalledWith('error', expect.any(Function));
    });
    
    it('should handle connection errors', async () => {
      // Mock console.error to prevent test output pollution
      const originalConsoleError = console.error;
      console.error = vi.fn();
      
      try {
        // Create a mock WebSocket that throws an error when accessed
        global.WebSocket = vi.fn().mockImplementation(() => {
          throw new Error('Connection failed');
        });
        
        // Attempt to connect but catch the promise rejection
        // to avoid unhandled promise rejection warnings
        await wsClient.connect(testUrl).catch(error => {
          // Expected error
          expect(error.message).toBe('Connection failed');
        });
        
        // Verify error was logged
        expect(console.error).toHaveBeenCalled();
      } finally {
        // Restore console.error
        console.error = originalConsoleError;
      }
    });
    
    // This test verifies that abnormal WebSocket closures trigger a reconnection attempt
    it('should attempt reconnection after abnormal closure', () => {
      // Use fake timers to control the setTimeout behavior
      vi.useFakeTimers();
      
      try {
        // Start with a new WebSocketClient with minimal reconnection delay
        wsClient = new WebSocketClient();
        wsClient.reconnectDelay = 10;
        
        // Set up the initial connection
        wsClient.connect(testUrl);
        
        // Clear the WebSocket constructor calls from the initial connection
        global.WebSocket.mockClear();
        
        // Now directly call the internal reconnection handler
        // This is what's called when a non-clean closure happens
        wsClient.handleReconnect(testUrl);
        
        // Run all timers to trigger the setTimeout callback
        vi.runAllTimers();
        
        // Verify a new WebSocket connection was attempted (reconnection)
        expect(global.WebSocket).toHaveBeenCalledWith(testUrl);
      } finally {
        // Always restore real timers
        vi.useRealTimers();
      }
    });
  });
  
  describe('message handling', () => {
    it('should send messages', () => {
      // Connect first
      wsClient.connect(testUrl);
      
      // Send a test message
      const message = { action: 'test', data: 'payload' };
      wsClient.send(message);
      
      // Verify message was sent correctly
      expect(mockWebSocket.send).toHaveBeenCalledWith(JSON.stringify(message));
    });
    
    it('should receive and process valid JSON messages', () => {
      // Set up message event handler capture
      let messageHandler;
      mockWebSocket.addEventListener.mockImplementation((event, handler) => {
        if (event === 'message') {
          messageHandler = handler;
        }
      });
      
      // Set up message processing spy
      const handleMessageSpy = vi.spyOn(wsClient, 'handleMessage');
      
      // Connect to server
      wsClient.connect(testUrl);
      
      // Create valid JSON message
      const testMessage = { type: 'update', data: { test: 'value' } };
      const messageEvent = { data: JSON.stringify(testMessage) };
      
      // Ensure we have a message handler
      expect(messageHandler).toBeDefined();
      
      // Trigger the message event
      messageHandler(messageEvent);
      
      // Verify message was processed
      expect(handleMessageSpy).toHaveBeenCalledWith(testMessage);
    });
    
    it('should handle invalid JSON messages', () => {
      // Set up message event handler capture
      let messageHandler;
      mockWebSocket.addEventListener.mockImplementation((event, handler) => {
        if (event === 'message') {
          messageHandler = handler;
        }
      });
      
      // Mock console.error to prevent test output pollution
      const originalConsoleError = console.error;
      console.error = vi.fn();
      
      try {
        // Connect to server
        wsClient.connect(testUrl);
        
        // Ensure we have a message handler
        expect(messageHandler).toBeDefined();
        
        // Create invalid JSON message
        const invalidMessage = '{invalid json:}';
        
        // Trigger the message event with invalid JSON
        messageHandler({ data: invalidMessage });
        
        // Verify error was logged
        expect(console.error).toHaveBeenCalled();
      } finally {
        // Restore console.error
        console.error = originalConsoleError;
      }
    });
  });
  
  describe('HTMX integration', () => {
    it('should initialize HTMX integration', () => {
      // The actual method name in WebSocketClient is initHtmxIntegration
      expect(typeof wsClient.initHtmxIntegration).toBe('function');
      
      // Save original document and window
      const originalDocument = global.document;
      const originalWindow = global.window;
      
      try {
        // Create mock DOM elements
        global.document = {
          body: {
            setAttribute: vi.fn(),
            getAttribute: vi.fn()
          }
        };
        
        // Mock window with htmx
        global.window = {
          htmx: {
            process: vi.fn(),
            on: vi.fn()
          }
        };
        
        // Call the HTMX integration method
        wsClient.initHtmxIntegration();
        
        // Verify document body attributes were set correctly
        expect(document.body.setAttribute).toHaveBeenCalledWith('hx-ext', 'ws');
        
        // Verify htmx.process was called
        expect(window.htmx.process).toHaveBeenCalledWith(document.body);
      } finally {
        // Restore originals
        global.document = originalDocument;
        global.window = originalWindow;
      }
    });
  });
});
