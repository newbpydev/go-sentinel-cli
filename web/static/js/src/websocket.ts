// WebSocket client implementation
export class WebSocketClient {
  private socket: WebSocket | null = null;
  private url: string | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectInterval = 1000; // 1 second
  private messageHandlers: Array<(data: any) => void> = [];
  private connectHandlers: Array<() => void> = [];
  private disconnectHandlers: Array<() => void> = [];
  private errorHandlers: Array<(error: Event) => void> = [];
  
  // Public property for backward compatibility
  public onOpen: (() => void) | null = null;

  constructor() {}

  connect(url: string): void {
    this.url = url;
    this._connect();
  }

  private _connect(): void {
    if (!this.url) return;

    this.socket = new WebSocket(this.url);

    this.socket.onopen = () => {
      this.reconnectAttempts = 0;
      this.connectHandlers.forEach(handler => handler());
      
      // Call onOpen handler if set (for backward compatibility)
      if (this.onOpen) this.onOpen();
    };

    this.socket.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        this.messageHandlers.forEach(handler => handler(data));
      } catch (error) {
        console.error('Error parsing WebSocket message:', error);
      }
    };

    this.socket.onclose = () => {
      this.disconnectHandlers.forEach(handler => handler());
      this._reconnect();
    };

    this.socket.onerror = (error) => {
      console.error('WebSocket error:', error);
      this.errorHandlers.forEach(handler => handler(error));
    };
  }

  private _reconnect(): void {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      setTimeout(() => {
        console.log(`Reconnecting (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})...`);
        this._connect();
      }, this.reconnectInterval * this.reconnectAttempts);
    } else {
      console.error('Max reconnection attempts reached');
    }
  }

  send(data: any): boolean {
    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
      this.socket.send(JSON.stringify(data));
      return true;
    }
    return false;
  }

  disconnect(): void {
    if (this.socket) {
      this.socket.close();
      this.socket = null;
    }
  }

  /**
   * Register a message handler
   * @param callback - Function to call when a message is received
   * @returns Function to unregister the handler
   */
  onMessage(callback: (data: any) => void): () => void {
    this.messageHandlers.push(callback);
    return () => {
      this.messageHandlers = this.messageHandlers.filter(handler => handler !== callback);
    };
  }
  
  /**
   * Register an event handler for connection events
   * @param event - Event type ('open', 'close', 'error')
   * @param callback - Function to call when the event occurs
   * @returns Function to unregister the handler
   */
  on(event: 'open' | 'close' | 'error', callback: (error?: Event) => void): () => void {
    switch (event) {
      case 'open':
        this.connectHandlers.push(callback as () => void);
        return () => {
          this.connectHandlers = this.connectHandlers.filter(handler => handler !== callback);
        };
      case 'close':
        this.disconnectHandlers.push(callback as () => void);
        return () => {
          this.disconnectHandlers = this.disconnectHandlers.filter(handler => handler !== callback);
        };
      case 'error':
        this.errorHandlers.push(callback as (error: Event) => void);
        return () => {
          this.errorHandlers = this.errorHandlers.filter(handler => handler !== callback);
        };
      default:
        console.warn(`Unknown event type: ${event}`);
        return () => {};
    }
  }

  onConnect(handler: () => void): () => void {
    this.connectHandlers.push(handler);
    return () => {
      this.connectHandlers = this.connectHandlers.filter(h => h !== handler);
    };
  }

  onDisconnect(handler: () => void): () => void {
    this.disconnectHandlers.push(handler);
    return () => {
      this.disconnectHandlers = this.disconnectHandlers.filter(h => h !== handler);
    };
  }

  onError(handler: (error: Event) => void): () => void {
    this.errorHandlers.push(handler);
    return () => {
      this.errorHandlers = this.errorHandlers.filter(h => h !== handler);
    };
  }
}

export default WebSocketClient;
