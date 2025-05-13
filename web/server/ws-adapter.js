/**
 * WebSocket Adapter for Go Sentinel
 * 
 * This module connects the frontend WebSocket to the Go backend WebSocket server.
 * It handles message translation, connection management, and error handling.
 */

const WebSocket = require('ws');

class GoWebSocketAdapter {
  constructor(serverPort = 8080, path = '/ws') {
    this.backendUrl = `ws://localhost:${serverPort}${path}`;
    this.clients = new Map(); // clientId -> client connection
    this.backendSocket = null;
    this.isConnected = false;
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 10;
    this.reconnectInterval = 3000; // ms
  }

  /**
   * Start the adapter server
   * @param {WebSocketServer} wss - The WebSocket server instance for frontend clients
   */
  start(wss) {
    console.log('[WS Adapter] Starting WebSocket adapter');
    
    // Connect to backend
    this.connectToBackend();
    
    // Handle frontend client connections
    wss.on('connection', (client, req) => {
      const clientId = this.generateClientId();
      console.log(`[WS Adapter] Frontend client connected: ${clientId}`);
      
      // Store client connection
      this.clients.set(clientId, client);
      
      // Send connection status
      this.sendToClient(client, {
        type: 'connection_status',
        payload: {
          status: this.isConnected ? 'connected' : 'disconnected',
          backendUrl: this.backendUrl
        }
      });
      
      // Handle client messages
      client.on('message', (message) => {
        this.handleClientMessage(clientId, message);
      });
      
      // Handle client disconnect
      client.on('close', () => {
        console.log(`[WS Adapter] Frontend client disconnected: ${clientId}`);
        this.clients.delete(clientId);
      });
    });
    
    // Set up periodic health check
    setInterval(() => this.healthCheck(), 30000);
  }
  
  /**
   * Connect to the Go backend WebSocket server
   */
  connectToBackend() {
    if (this.backendSocket && (this.backendSocket.readyState === WebSocket.OPEN || 
        this.backendSocket.readyState === WebSocket.CONNECTING)) {
      return;
    }
    
    console.log(`[WS Adapter] Connecting to backend: ${this.backendUrl}`);
    
    try {
      this.backendSocket = new WebSocket(this.backendUrl);
      
      this.backendSocket.on('open', () => {
        console.log('[WS Adapter] Connected to backend');
        this.isConnected = true;
        this.reconnectAttempts = 0;
        
        // Notify all clients about successful connection
        this.broadcastToClients({
          type: 'connection_status',
          payload: { status: 'connected' }
        });
      });
      
      this.backendSocket.on('message', (message) => {
        this.handleBackendMessage(message);
      });
      
      this.backendSocket.on('error', (error) => {
        console.error('[WS Adapter] Backend socket error:', error.message);
      });
      
      this.backendSocket.on('close', () => {
        console.log('[WS Adapter] Backend socket closed');
        this.isConnected = false;
        
        // Notify all clients about disconnection
        this.broadcastToClients({
          type: 'connection_status',
          payload: { status: 'disconnected' }
        });
        
        // Attempt to reconnect
        this.scheduleReconnect();
      });
    } catch (error) {
      console.error('[WS Adapter] Failed to connect to backend:', error.message);
      this.scheduleReconnect();
    }
  }
  
  /**
   * Schedule a reconnection attempt to the backend
   */
  scheduleReconnect() {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.log('[WS Adapter] Max reconnect attempts reached, stopping reconnect');
      
      // Notify all clients about reconnection failure
      this.broadcastToClients({
        type: 'connection_status',
        payload: { 
          status: 'failed',
          message: 'Failed to connect to test server after multiple attempts' 
        }
      });
      return;
    }
    
    this.reconnectAttempts++;
    
    // Notify all clients about reconnection attempt
    this.broadcastToClients({
      type: 'connection_status',
      payload: { 
        status: 'reconnecting',
        attempt: this.reconnectAttempts,
        maxAttempts: this.maxReconnectAttempts
      }
    });
    
    console.log(`[WS Adapter] Scheduling reconnect attempt ${this.reconnectAttempts}`);
    setTimeout(() => this.connectToBackend(), this.reconnectInterval);
  }
  
  /**
   * Handle a message from a frontend client
   * @param {string} clientId - The client ID
   * @param {Buffer|string} message - The message data
   */
  handleClientMessage(clientId, message) {
    try {
      const msgString = message.toString();
      console.log(`[WS Adapter] Message from client ${clientId}:`, msgString);
      
      // Parse the message
      const parsedMessage = JSON.parse(msgString);
      
      // If not connected to backend, queue or reject
      if (!this.isConnected) {
        console.log('[WS Adapter] Not connected to backend, cannot forward message');
        
        // Send error response to client
        const client = this.clients.get(clientId);
        if (client) {
          this.sendToClient(client, {
            type: 'error',
            payload: { 
              message: 'Not connected to test server',
              originalRequest: parsedMessage
            }
          });
        }
        return;
      }
      
      // Translate and forward message to backend
      this.sendToBackend(parsedMessage);
    } catch (error) {
      console.error('[WS Adapter] Error handling client message:', error.message);
    }
  }
  
  /**
   * Handle a message from the backend
   * @param {Buffer|string} message - The message data
   */
  handleBackendMessage(message) {
    try {
      const msgString = message.toString();
      console.log('[WS Adapter] Message from backend:', msgString);
      
      // Parse the message
      const parsedMessage = JSON.parse(msgString);
      
      // Broadcast to all clients
      this.broadcastToClients(parsedMessage);
    } catch (error) {
      console.error('[WS Adapter] Error handling backend message:', error.message);
    }
  }
  
  /**
   * Send a message to a specific client
   * @param {WebSocket} client - The client WebSocket
   * @param {object} message - The message object
   */
  sendToClient(client, message) {
    if (client.readyState === WebSocket.OPEN) {
      client.send(JSON.stringify(message));
    }
  }
  
  /**
   * Broadcast a message to all connected clients
   * @param {object} message - The message object
   */
  broadcastToClients(message) {
    const messageStr = JSON.stringify(message);
    this.clients.forEach((client) => {
      if (client.readyState === WebSocket.OPEN) {
        client.send(messageStr);
      }
    });
  }
  
  /**
   * Send a message to the backend
   * @param {object} message - The message object
   */
  sendToBackend(message) {
    if (this.backendSocket && this.backendSocket.readyState === WebSocket.OPEN) {
      this.backendSocket.send(JSON.stringify(message));
    } else {
      console.error('[WS Adapter] Cannot send to backend: not connected');
    }
  }
  
  /**
   * Perform a health check and reconnect if needed
   */
  healthCheck() {
    if (!this.isConnected && this.reconnectAttempts < this.maxReconnectAttempts) {
      console.log('[WS Adapter] Health check: reconnecting to backend');
      this.connectToBackend();
    }
  }
  
  /**
   * Generate a unique client ID
   * @returns {string} The generated client ID
   */
  generateClientId() {
    return `client-${Date.now()}-${Math.floor(Math.random() * 1000)}`;
  }
  
  /**
   * Stop the adapter
   */
  stop() {
    console.log('[WS Adapter] Stopping adapter');
    
    // Close backend connection
    if (this.backendSocket) {
      this.backendSocket.close();
      this.backendSocket = null;
    }
    
    // Close all client connections
    this.clients.forEach((client) => {
      client.close();
    });
    
    this.clients.clear();
  }
}

module.exports = { GoWebSocketAdapter };
