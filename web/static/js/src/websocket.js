// WebSocket Client for Go Sentinel
// Handles WebSocket connections and message processing
class WebSocketClient {
  /**
   * Create a new WebSocketClient
   * @param {string} [url] - Optional WebSocket server URL to connect to
   */
  constructor(url = null) {
    this.socket = null;
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 5;
    this.reconnectDelay = 1000; // Start with 1 second
    this.maxReconnectDelay = 30000; // Max 30 seconds
    this.messageHandlers = new Map();
    this.connectionHandlers = {
      onOpen: [],
      onClose: [],
      onError: []
    };
  }

  /**
   * Initialize HTMX WebSocket integration
   */
  initHtmxIntegration() {
    if (typeof document === 'undefined' || !window.htmx) return;
    
    // Add HTMX WebSocket extension
    document.body.setAttribute('hx-ext', 'ws');
    
    // Process HTMX extensions
    if (typeof window.htmx.process === 'function') {
      window.htmx.process(document.body);
    }
  }

  /**
   * Connect to the WebSocket server
   * @param {string} url - WebSocket server URL
   * @returns {Promise<WebSocket>} The WebSocket instance
   */
  connect(url) {
    return new Promise((resolve, reject) => {
      // Clean up any existing connection
      if (this.socket) {
        if (this.socket.readyState === WebSocket.OPEN || 
            this.socket.readyState === WebSocket.CONNECTING) {
          console.warn('WebSocket already connected or connecting, closing existing connection');
          this.socket.close();
        }
        this.socket = null;
      }

      try {
        // Store the URL for reconnection attempts
        this.url = url || this.url;
        
        // Create a new WebSocket connection
        this.socket = new WebSocket(this.url);
        
        // Set up event listeners
        this.setupEventListeners();
        
        // Handle successful connection
        const onOpen = () => {
          this.socket.removeEventListener('open', onOpen);
          this.socket.removeEventListener('error', onError);
          resolve(this.socket);
        };
        
        // Handle connection error
        const onError = (error) => {
          this.socket.removeEventListener('open', onOpen);
          this.socket.removeEventListener('error', onError);
          reject(error);
        };
        
        this.socket.addEventListener('open', onOpen);
        this.socket.addEventListener('error', onError);
        
        // Expose the client globally for testing
        if (typeof window !== 'undefined') {
          window.webSocketClient = this;
        }
      } catch (error) {
        console.error('Failed to create WebSocket:', error);
        reject(error);
      }
    });
  }

  /**
   * Set up WebSocket event listeners
   */
  setupEventListeners() {
    if (!this.socket) return;

    // Define event handlers
    const handleOpen = (event) => {
      console.log('WebSocket connected to', this.url);
      this.reconnectAttempts = 0;
      this.reconnectDelay = 1000; // Reset reconnect delay
      this.connectionHandlers.onOpen.forEach(handler => {
        try {
          handler(event);
        } catch (err) {
          console.error('Error in open handler:', err);
        }
      });
    };

    const handleMessage = (event) => {
      try {
        // Parse the message if it's a string
        const message = typeof event.data === 'string' ? 
          (event.data.startsWith('{') ? JSON.parse(event.data) : { type: 'message', data: event.data }) : 
          event.data;
          
        this.handleMessage(message);
      } catch (error) {
        console.error('Error parsing WebSocket message:', error, event.data);
        this.connectionHandlers.onError.forEach(handler => {
          try {
            handler(error);
          } catch (err) {
            console.error('Error in error handler:', err);
          }
        });
      }
    };

    const handleClose = (event) => {
      console.log('WebSocket disconnected:', event.code, event.reason);
      this.connectionHandlers.onClose.forEach(handler => {
        try {
          handler(event);
        } catch (err) {
          console.error('Error in close handler:', err);
        }
      });
      
      // Only attempt to reconnect if the connection was not closed cleanly
      // and we're not already reconnecting
      if (!event.wasClean && this.socket && this.socket.readyState !== WebSocket.CONNECTING) {
        console.log('Attempting to reconnect...');
        this.handleReconnect(this.url);
      }
    };

    const handleError = (error) => {
      console.error('WebSocket error:', error);
      this.connectionHandlers.onError.forEach(handler => handler(error));
    };

    // Clean up any existing event listeners to prevent duplicates
    if (this.socket.removeEventListener) {
      this.socket.removeEventListener('open', this._handleOpen);
      this.socket.removeEventListener('message', this._handleMessage);
      this.socket.removeEventListener('close', this._handleClose);
      this.socket.removeEventListener('error', this._handleError);
    }
    
    // Store references to the handlers for cleanup
    this._handleOpen = handleOpen;
    this._handleMessage = handleMessage;
    this._handleClose = handleClose;
    this._handleError = handleError;
    
    // Set up event listeners
    this.socket.onopen = handleOpen;
    this.socket.onmessage = handleMessage;
    this.socket.onclose = handleClose;
    this.socket.onerror = handleError;
    
    // Also set up event listeners using addEventListener for better compatibility
    if (this.socket.addEventListener) {
      this.socket.addEventListener('open', handleOpen);
      this.socket.addEventListener('message', handleMessage);
      this.socket.addEventListener('close', handleClose);
      this.socket.addEventListener('error', handleError);
    }
  }

  /**
   * Handle reconnection logic
   * @param {string} url - WebSocket server URL
   */
  handleReconnect(url) {
    // Clear any existing reconnect timer
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
      this.reconnectTimeout = null;
    }

    // Don't attempt to reconnect if we're already reconnecting
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('Max reconnection attempts reached');
      const error = new Error('Max reconnection attempts reached');
      this.connectionHandlers.onError.forEach(handler => handler(error));
      return;
    }

    // Calculate delay with exponential backoff and jitter
    const baseDelay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts);
    const jitter = Math.random() * 1000;
    const delay = Math.min(baseDelay + jitter, this.maxReconnectDelay);

    console.log(`Attempting to reconnect in ${Math.round(delay / 1000)} seconds...`);
    
    // Clear any existing timeout to prevent multiple reconnection attempts
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
    }
    
    this.reconnectTimeout = setTimeout(() => {
      this.reconnectAttempts++;
      this.connect(url);
      this.isReconnecting = false;
      
      // Only attempt to reconnect if we're not already connected
      if (!this.socket || this.socket.readyState === WebSocket.CLOSED) {
        this.connect(url);
      }
    }, delay);
    
    // Return the delay for testing purposes
    return delay;
  }

  /**
   * Register a message handler
   * @param {string} messageType - Type of message to handle
   * @param {Function} handler - Handler function
   * @returns {Function} Unsubscribe function
   */
  onMessage(messageType, handler) {
    if (!this.messageHandlers.has(messageType)) {
      this.messageHandlers.set(messageType, new Set());
    }
    
    const handlers = this.messageHandlers.get(messageType);
    handlers.add(handler);
    
    // Return unsubscribe function
    return () => {
      handlers.delete(handler);
      if (handlers.size === 0) {
        this.messageHandlers.delete(messageType);
      }
    };
  }

  /**
   * Handle incoming messages
   * @param {Object} message - Parsed message object
   */
  handleMessage(message) {
    if (!message) {
      console.warn('Received empty message');
      return;
    }

    // Handle different message formats
    let messageType, messageData;
    
    if (typeof message === 'object' && message.type) {
      // Standard format: { type: 'event', data: {...} }
      messageType = message.type;
      messageData = message.data || {};
    } else if (typeof message === 'object') {
      // Simple object format: treat the message itself as data
      messageType = 'message';
      messageData = message;
    } else {
      console.warn('Received message with invalid format:', message);
      return;
    }

    // Get handlers for this message type
    const handlers = this.messageHandlers.get(messageType) || [];
    
    // Also check for wildcard handlers
    const wildcardHandlers = this.messageHandlers.get('*') || [];
    
    // Combine all handlers
    const allHandlers = [...handlers, ...wildcardHandlers];
    
    // Call all handlers
    allHandlers.forEach(handler => {
      try {
        handler(messageData, messageType);
      } catch (error) {
        console.error(`Error in ${messageType} handler:`, error);
      }
    });
    
    // Also trigger HTMX events if HTMX is available
    if (typeof window !== 'undefined' && window.htmx && window.htmx.trigger) {
      try {
        window.htmx.trigger(document.body, `ws:${messageType}`, messageData);
      } catch (error) {
        console.error('Error triggering HTMX event:', error);
      }
    }
  }

  /**
   * Send a message through the WebSocket
   * @param {string|Object} message - Message to send
   * @returns {Promise<boolean>} Whether the message was sent successfully
   */
  send(message) {
    return new Promise((resolve) => {
      if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
        console.error('WebSocket is not connected');
        resolve(false);
        return;
      }

      try {
        const messageStr = typeof message === 'string' ? message : JSON.stringify(message);
        this.socket.send(messageStr);
        resolve(true);
      } catch (error) {
        console.error('Failed to send WebSocket message:', error);
        resolve(false);
      }
    });
  }

  /**
   * Close the WebSocket connection
   * @param {number} code - Close code
   * @param {string} reason - Close reason
   */
  close(code = 1000, reason = '') {
    if (this.socket && typeof this.socket.close === 'function') {
      this.socket.close(code, reason);
    } else if (this.socket && this.socket.client && typeof this.socket.client.close === 'function') {
      // Handle our mock WebSocket implementation
      this.socket.client.close(code, reason);
    }
    this.socket = null;
  }

  /**
   * Register a connection event handler
   * @param {'open'|'close'|'error'} event - Event type
   * @param {Function} handler - Event handler
   * @returns {Function} Unsubscribe function
   */
  on(event, handler) {
    const eventKey = `on${event.charAt(0).toUpperCase() + event.slice(1)}`;
    if (!this.connectionHandlers[eventKey]) {
      throw new Error(`Invalid event type: ${event}`);
    }
    
    this.connectionHandlers[eventKey].push(handler);
    
    // Return unsubscribe function
    return () => {
      const index = this.connectionHandlers[eventKey].indexOf(handler);
      if (index !== -1) {
        this.connectionHandlers[eventKey].splice(index, 1);
      }
    };
  }
}

// Create a singleton instance for the default export
const webSocketClient = new WebSocketClient();

// Export the WebSocketClient class and the singleton instance
export { WebSocketClient, webSocketClient };

export default webSocketClient;

/**
 * Initialize WebSocket connection
 * @param {string} url - WebSocket server URL
 */
export function initWebSocket(url) {
  if (!url) {
    console.error('WebSocket URL is required');
    return null;
  }
  
  // Set up default message handlers
  setupDefaultMessageHandlers(webSocketClient);
  
  // Connect to the WebSocket server
  webSocketClient.connect(url);
  
  return webSocketClient;
}

/**
 * Set up default message handlers
 * @param {WebSocketClient} wsClient - WebSocket client instance
 */
function setupDefaultMessageHandlers(wsClient) {
  // Test result update handler
  wsClient.onMessage('test_update', (data) => {
    const { testId, status, output, duration } = data;
    const testElement = document.querySelector(`[data-test-id="${testId}"]`);
    
    if (testElement) {
      // Update test status
      const statusElement = testElement.querySelector('.test-status');
      if (statusElement) {
        statusElement.textContent = status;
        statusElement.className = `test-status status-${status.toLowerCase()}`;
      }
      
      // Update test output if available
      const outputElement = testElement.querySelector('.test-output');
      if (outputElement && output) {
        outputElement.textContent = output;
      }
      
      // Update test duration if available
      const durationElement = testElement.querySelector('.test-duration');
      if (durationElement && duration) {
        durationElement.textContent = duration;
      }
      
      // Trigger animation
      testElement.classList.add('updated');
      setTimeout(() => testElement.classList.remove('updated'), 1000);
    }
  });
  
  // Test suite started handler
  wsClient.onMessage('test_suite_started', (data) => {
    const { suiteId, timestamp } = data;
    console.log(`Test suite ${suiteId} started at ${new Date(timestamp).toISOString()}`);
    // Update UI to show test suite is running
    document.dispatchEvent(new CustomEvent('test-suite-started', { detail: data }));
  });
  
  // Test suite completed handler
  wsClient.onMessage('test_suite_completed', (data) => {
    const { suiteId, passed, failed, skipped, duration } = data;
    console.log(`Test suite ${suiteId} completed: ${passed} passed, ${failed} failed, ${skipped} skipped in ${duration}`);
    // Update UI with test suite results
    document.dispatchEvent(new CustomEvent('test-suite-completed', { detail: data }));
  });
  
  // Error handler
  wsClient.onMessage('error', (error) => {
    console.error('Received error from server:', error);
    // Show error notification to user
    document.dispatchEvent(new CustomEvent('show-toast', { 
      detail: { 
        message: error.message || 'An error occurred',
        type: 'error',
        duration: 5000
      } 
    }));
  });
}

/**
 * Send a test run command via WebSocket
 * @param {string|string[]} testIds - Single test ID or array of test IDs to run
 * @returns {boolean} Whether the message was sent successfully
 */
export function runTests(testIds) {
  if (!testIds || (Array.isArray(testIds) && testIds.length === 0)) {
    console.warn('No test IDs provided');
    return false;
  }
  
  const tests = Array.isArray(testIds) ? testIds : [testIds];
  return webSocketClient.send('run_tests', { tests });
}

/**
 * Send a test debug command via WebSocket
 * @param {string} testId - Test ID to debug
 * @returns {boolean} Whether the message was sent successfully
 */
export function debugTest(testId) {
  if (!testId) {
    console.warn('No test ID provided');
    return false;
  }
  
  return webSocketClient.send('debug_test', { test: testId });
}

/**
 * Send a request to get test details
 * @param {string} testId - Test ID to get details for
 * @returns {boolean} Whether the message was sent successfully
 */
export function getTestDetails(testId) {
  if (!testId) {
    console.warn('No test ID provided');
    return false;
  }
  
  return webSocketClient.send('get_test_details', { test: testId });
}
