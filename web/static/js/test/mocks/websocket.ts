// Minimal WebSocket mock for testing
export class MockWebSocket extends EventTarget {
  // Static constants
  static readonly CONNECTING = 0;
  static readonly OPEN = 1;
  static readonly CLOSING = 2;
  static readonly CLOSED = 3;

  // Instance properties
  binaryType: BinaryType = 'arraybuffer';
  bufferedAmount = 0;
  extensions = '';
  protocol = '';
  url: string;
  readyState: number;

  // Event handlers
  onopen: ((this: WebSocket, ev: Event) => any) | null = null;
  onclose: ((this: WebSocket, ev: CloseEvent) => any) | null = null;
  onmessage: ((this: WebSocket, ev: MessageEvent) => any) | null = null;
  onerror: ((this: WebSocket, ev: Event) => any) | null = null;

  constructor(url: string | URL, _protocols?: string | string[]) {
    super();
    this.url = typeof url === 'string' ? url : url.toString();
    this.readyState = MockWebSocket.CONNECTING;
  }

  // Public methods
  send(_data: string | ArrayBufferLike | Blob | ArrayBufferView): void {
    if (this.readyState !== MockWebSocket.OPEN) {
      throw new Error('WebSocket is not open');
    }
    // Mock implementation - data is intentionally unused in tests
  }

  close(code?: number, reason?: string): void {
    if (this.readyState === MockWebSocket.CLOSED) {
      return;
    }
    
    const previousState = this.readyState;
    this.readyState = MockWebSocket.CLOSED;
    
    if (this.onclose && previousState !== MockWebSocket.CLOSING) {
      const event = new CloseEvent('close', { 
        code: code || 1000, 
        reason: reason || '', 
        wasClean: true 
      });
      this.onclose.call(this as unknown as WebSocket, event);
    }
  }

  // Test helper methods
  simulateOpen() {
    this.readyState = MockWebSocket.OPEN;
    if (this.onopen) {
      this.onopen.call(this as unknown as WebSocket, new Event('open'));
    }
  }

  simulateMessage(data: any) {
    const event = new MessageEvent('message', { data });
    
    // Call the onmessage handler if it exists
    if (this.onmessage) {
      this.onmessage.call(this as unknown as WebSocket, event);
    }
    
    // Dispatch the event to all listeners
    this.dispatchEvent(event);
  }

  simulateError() {
    if (this.onerror) {
      this.onerror.call(this as unknown as WebSocket, new Event('error'));
    }
  }

  // Override EventTarget methods
  override addEventListener(
    type: string,
    listener: EventListenerOrEventListenerObject | null,
    options?: boolean | AddEventListenerOptions
  ): void {
    super.addEventListener(type, listener as EventListener, options);
  }

  override removeEventListener(
    type: string,
    callback: EventListenerOrEventListenerObject | null,
    options?: boolean | EventListenerOptions
  ): void {
    super.removeEventListener(type, callback as EventListener, options);
  }
}

// Make MockWebSocket available globally for testing
declare global {
  interface Window {
    MockWebSocket: typeof MockWebSocket;
  }
}

if (typeof window !== 'undefined') {
  window.MockWebSocket = MockWebSocket;
}

export default MockWebSocket;
