// Mock WebSocket implementation for testing
export class WebSocketClient {
  constructor() {
    this.url = null;
    this.socket = null;
    this.messageHandlers = new Set();
    this.connectHandlers = new Set();
    this.disconnectHandlers = new Set();
    this.errorHandlers = new Set();
  }

  connect(url) {
    this.url = url;
    this.socket = new WebSocket(url);
    this.socket.onopen = this._handleOpen.bind(this);
    this.socket.onmessage = this._handleMessage.bind(this);
    this.socket.onclose = this._handleClose.bind(this);
    this.socket.onerror = this._handleError.bind(this);
  }

  disconnect() {
    if (this.socket) {
      this.socket.close();
      this.socket = null;
    }
  }

  send(message) {
    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
      this.socket.send(JSON.stringify(message));
      return true;
    }
    return false;
  }

  onMessage(handler) {
    this.messageHandlers.add(handler);
    return () => this.messageHandlers.delete(handler);
  }

  onConnect(handler) {
    this.connectHandlers.add(handler);
    return () => this.connectHandlers.delete(handler);
  }

  onDisconnect(handler) {
    this.disconnectHandlers.add(handler);
    return () => this.disconnectHandlers.delete(handler);
  }

  onError(handler) {
    this.errorHandlers.add(handler);
    return () => this.errorHandlers.delete(handler);
  }

  _handleOpen(event) {
    this.connectHandlers.forEach(handler => handler(event));
  }

  _handleMessage(event) {
    let message;
    try {
      message = JSON.parse(event.data);
    } catch (e) {
      console.error('Error parsing message:', e);
      return;
    }
    this.messageHandlers.forEach(handler => handler(message));
  }

  _handleClose(event) {
    this.disconnectHandlers.forEach(handler => handler(event));
  }

  _handleError(error) {
    this.errorHandlers.forEach(handler => handler(error));
  }
}

export default WebSocketClient;
