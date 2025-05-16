/**
 * WebSocket test helpers for Go Sentinel
 * Provides utilities for testing WebSocket functionality
 */

/**
 * Creates a mock WebSocket server for testing
 * @returns {Object} Mock WebSocket server with helper methods
 */
export function createMockWebSocketServer() {
  const clients = new Set();
  const messages = [];
  
  const mockServer = {
    /** @type {Function[]} */
    onConnectionCallbacks: [],
    
    /**
     * Simulate a client connecting to the server
     * @returns {Object} Mock WebSocket client
     */
    connectClient() {
      const client = {
        id: `client-${Math.random().toString(36).substr(2, 9)}`,
        messages: [],
        closed: false,
        closeCode: null,
        closeReason: null,
        
        send(message) {
          if (this.closed) {
            throw new Error('WebSocket is closed');
          }
          this.messages.push(message);
        },
        
        close(code = 1000, reason = '') {
          if (this.closed) return;
          this.closed = true;
          this.closeCode = code;
          this.closeReason = reason;
          if (this.onclose) {
            this.onclose({ code, reason, wasClean: true });
          }
        },
        
        // Client-side event handlers
        onopen: null,
        onmessage: null,
        onclose: null,
        onerror: null,
        
        // Test helpers
        _triggerMessage(data) {
          if (this.onmessage) {
            this.onmessage({ data });
          }
        },
        
        _triggerError(error) {
          if (this.onerror) {
            this.onerror(error || new Error('Test error'));
          }
        },
        
        _triggerClose(event = {}) {
          this.closed = true;
          this.closeCode = event.code || 1000;
          this.closeReason = event.reason || '';
          if (this.onclose) {
            this.onclose({
              code: event.code || 1000,
              reason: event.reason || '',
              wasClean: event.wasClean !== false
            });
          }
        }
      };
      
      clients.add(client);
      
      // Notify server of new connection
      mockServer.onConnectionCallbacks.forEach(callback => callback(client));
      
      // Automatically open the connection
      setTimeout(() => {
        if (!client.closed && client.onopen) {
          client.onopen({});
        }
      }, 0);
      
      return client;
    },
    
    /**
     * Register a callback for new client connections
     * @param {Function} callback - Callback function that receives the client
     */
    onConnection(callback) {
      this.onConnectionCallbacks.push(callback);
      return () => {
        const index = this.onConnectionCallbacks.indexOf(callback);
        if (index !== -1) {
          this.onConnectionCallbacks.splice(index, 1);
        }
      };
    },
    
    /**
     * Broadcast a message to all connected clients
     * @param {*} data - Data to send (will be JSON stringified)
     * @param {Function} [filter] - Optional filter function to select clients
     */
    broadcast(data, filter) {
      const message = typeof data === 'string' ? data : JSON.stringify(data);
      clients.forEach(client => {
        if (!client.closed && (!filter || filter(client))) {
          client._triggerMessage(message);
        }
      });
    },
    
    /**
     * Close all client connections
     * @param {number} [code] - Close code
     * @param {string} [reason] - Close reason
     */
    closeAll(code, reason) {
      clients.forEach(client => {
        if (!client.closed) {
          client._triggerClose({ code, reason });
        }
      });
      clients.clear();
    },
    
    /**
     * Get all connected clients
     * @returns {Array} Array of connected clients
     */
    getClients() {
      return Array.from(clients);
    },
    
    /**
     * Get all messages sent to the server
     * @returns {Array} Array of messages
     */
    getMessages() {
      return messages;
    },
    
    /**
     * Reset the mock server state
     */
    reset() {
      this.closeAll();
      messages.length = 0;
      this.onConnectionCallbacks = [];
    }
  };
  
  return mockServer;
}

/**
 * Wait for a specific WebSocket event
 * @param {Object} ws - WebSocket instance
 * @param {string} event - Event name ('open', 'message', 'close', 'error')
 * @param {number} [timeout=1000] - Timeout in milliseconds
 * @returns {Promise} Resolves when the event occurs or rejects on timeout
 */
export function waitForWebSocketEvent(ws, event, timeout = 1000) {
  return new Promise((resolve, reject) => {
    const timer = setTimeout(() => {
      reject(new Error(`Timeout waiting for WebSocket ${event} event`));
    }, timeout);
    
    const handler = (...args) => {
      clearTimeout(timer);
      resolve(args.length === 1 ? args[0] : args);
    };
    
    // Handle both event emitter and direct property styles
    if (ws.addEventListener) {
      ws.addEventListener(event, handler, { once: true });
    } else {
      const originalHandler = ws[`on${event}`];
      ws[`on${event}`] = function(...args) {
        if (originalHandler) originalHandler.apply(ws, args);
        handler.apply(null, args);
      };
    }
  });
}

/**
 * Create a mock WebSocket URL for testing
 * @returns {string} A mock WebSocket URL
 */
export function createMockWebSocketUrl() {
  return 'ws://test-websocket-server';
}
