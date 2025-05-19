// WebSocket client implementation
export class WebSocketClient {
  private socket: WebSocket | null = null;
  private url: string;
  private reconnectAttempts: number = 0;
  private maxReconnectAttempts: number = 5;
  private reconnectDelay: number = 1000; // Start with 1 second
  private maxReconnectDelay: number = 30000; // Max 30 seconds
  private connectHandlers: ((event?: Event) => void)[] = [];
  private messageHandlers: ((data: any) => void)[] = [];
  private disconnectHandlers: ((event?: CloseEvent) => void)[] = [];
  private errorHandlers: ((error: Event) => void)[] = [];
  private onOpen?: () => void; // Legacy handler
  
  constructor(url: string) {
    this.url = url;
    this._connect();
  }

  private _connect(): void {
    if (!this.url) {
      console.error('WebSocket URL not provided');
      return;
    }

    try {
      this.socket = new WebSocket(this.url);

      this.socket.onopen = (event: Event) => {
        this.reconnectAttempts = 0;
        this.reconnectDelay = 1000; // Reset delay on successful connection
        this.connectHandlers.forEach(handler => handler(event));
        
        // Call onOpen handler if set (for backward compatibility)
        if (this.onOpen) this.onOpen();
      };

      this.socket.onmessage = (event: MessageEvent) => {
        try {
          const data = JSON.parse(event.data);
          this.messageHandlers.forEach(handler => handler(data));
        } catch (error) {
          console.error('Error parsing WebSocket message:', error);
          this.errorHandlers.forEach(handler => 
            handler(new ErrorEvent('error', { error, message: 'Failed to parse message' }))
          );
        }
      };

      this.socket.onclose = (event: CloseEvent) => {
        this.disconnectHandlers.forEach(handler => handler(event));
        
        // Don't reconnect if close was clean and intended
        if (!event.wasClean) {
          this._reconnect();
        }
      };

      this.socket.onerror = (error: Event) => {
        console.error('WebSocket error:', error);
        this.errorHandlers.forEach(handler => handler(error));
      };
    } catch (error) {
      console.error('Error creating WebSocket connection:', error);
      this._reconnect();
    }
  }

  private _reconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('Max reconnection attempts reached');
      return;
    }

    this.reconnectAttempts++;
    const delay = Math.min(this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1), this.maxReconnectDelay);
    
    console.log(`Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
    
    setTimeout(() => {
      this._connect();
    }, delay);
  }

  public send(data: any): void {
    if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
      console.error('WebSocket is not connected');
      return;
    }

    try {
      const message = typeof data === 'string' ? data : JSON.stringify(data);
      this.socket.send(message);
    } catch (error) {
      console.error('Error sending message:', error);
      this.errorHandlers.forEach(handler => 
        handler(new ErrorEvent('error', { error, message: 'Failed to send message' }))
      );
    }
  }

  public onConnect(handler: (event?: Event) => void): void {
    this.connectHandlers.push(handler);
  }

  public onMessage(handler: (data: any) => void): void {
    this.messageHandlers.push(handler);
  }

  public onDisconnect(handler: (event?: CloseEvent) => void): void {
    this.disconnectHandlers.push(handler);
  }

  public onError(handler: (error: Event) => void): void {
    this.errorHandlers.push(handler);
  }

  public disconnect(): void {
    if (this.socket) {
      this.socket.close(1000, 'Client disconnected');
      this.socket = null;
    }
  }

  public isConnected(): boolean {
    return this.socket !== null && this.socket.readyState === WebSocket.OPEN;
  }

  public getState(): number {
    return this.socket ? this.socket.readyState : WebSocket.CLOSED;
  }
}

export default WebSocketClient;
