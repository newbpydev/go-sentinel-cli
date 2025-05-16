import { describe, it, expect, vi, beforeEach, afterEach, beforeAll } from 'vitest';
import MockWebSocket from '../mocks/websocket';

// Make MockWebSocket available globally for testing
declare global {
  interface Window {
    MockWebSocket: typeof MockWebSocket;
  }
}

// Assign to window for global access
if (typeof window !== 'undefined') {
  window.MockWebSocket = MockWebSocket;
}

describe('MockWebSocket', () => {
  let mockWs: MockWebSocket;
  const testUrl = 'ws://test.com';

  // Mock the global WebSocket before all tests
  beforeAll(() => {
    // @ts-expect-error - Mocking global WebSocket
    global.WebSocket = MockWebSocket;
  });

  beforeEach(() => {
    mockWs = new MockWebSocket(testUrl);
  });

  afterEach(() => {
    // Ensure the WebSocket is closed after each test
    if (mockWs.readyState !== MockWebSocket.CLOSED) {
      mockWs.close();
    }
    vi.restoreAllMocks();
  });

  describe('Connection Lifecycle', () => {
    it('should initialize with CONNECTING state', () => {
      expect(mockWs.readyState).toBe(WebSocket.CONNECTING);
    });

    it('should transition to OPEN state after initialization', async () => {
      await new Promise<void>((resolve) => {
        mockWs.onopen = () => {
          expect(mockWs.readyState).toBe(WebSocket.OPEN);
          resolve();
        };
        mockWs.simulateOpen();
      });
    });

    it('should set the correct URL', () => {
      expect(mockWs.url).toBe(testUrl);
    });
  });

  describe('Message Handling', () => {
    it('should receive messages', async () => {
      const testMessage = 'test message';
      
      await new Promise<void>((resolve) => {
        mockWs.onmessage = (event: MessageEvent) => {
          expect(event.data).toBe(testMessage);
          resolve();
        };
        
        // Simulate receiving a message
        mockWs.simulateMessage(testMessage);
      });
    });

    it('should handle multiple message listeners', async () => {
      const testMessage = 'test message';
      const listener1 = vi.fn();
      const listener2 = vi.fn();

      // Add listeners
      mockWs.addEventListener('message', listener1);
      mockWs.addEventListener('message', listener2);

      // Simulate a message
      mockWs.simulateMessage(testMessage);

      // Verify both listeners were called with the correct data
      expect(listener1).toHaveBeenCalledTimes(1);
      expect(listener2).toHaveBeenCalledTimes(1);
      
      // Get the first call's first argument for each listener with type safety
      const listener1Calls = (listener1 as unknown as { mock: { calls: [MessageEvent][] } }).mock.calls;
      const listener2Calls = (listener2 as unknown as { mock: { calls: [MessageEvent][] } }).mock.calls;
      
      // Ensure we have calls before accessing them
      if (!listener1Calls[0] || !listener1Calls[0][0] || !listener2Calls[0] || !listener2Calls[0][0]) {
        throw new Error('Expected mock functions to have been called');
      }
      
      const listener1Call = listener1Calls[0][0];
      const listener2Call = listener2Calls[0][0];
      
      expect(listener1Call.data).toBe(testMessage);
      expect(listener2Call.data).toBe(testMessage);
      
      // Clean up
      mockWs.removeEventListener('message', listener1);
      mockWs.removeEventListener('message', listener2);
    });
  });

  describe('Error Handling', () => {
    it('should handle errors', async () => {
      await new Promise<void>((resolve) => {
        mockWs.onerror = (event: Event) => {
          expect(event).toBeInstanceOf(Event);
          expect(event.type).toBe('error');
          resolve();
        };

        mockWs.simulateError();
      });
    });
  });

  describe('Connection Closure', () => {
    it('should close with code and reason', async () => {
      const testCode = 1000;
      const testReason = 'test reason';

      await new Promise<void>((resolve) => {
        mockWs.onclose = (event: CloseEvent) => {
          expect(event.code).toBe(testCode);
          expect(event.reason).toBe(testReason);
          expect(mockWs.readyState).toBe(MockWebSocket.CLOSED);
          resolve();
        };

        mockWs.close(testCode, testReason);
      });
    });

    it('should not allow sending messages after close', () => {
      mockWs.close();
      expect(() => mockWs.send('test')).toThrow();
    });
  });

  describe('Event Listener Management', () => {
    it('should add and remove event listeners', () => {
      const listener = vi.fn();

      // Add listener
      mockWs.addEventListener('message', listener);
      mockWs.dispatchEvent(new MessageEvent('message', { data: 'test' }));
      expect(listener).toHaveBeenCalledTimes(1);

      // Remove listener
      mockWs.removeEventListener('message', listener);
      mockWs.dispatchEvent(new MessageEvent('message', { data: 'test' }));
      expect(listener).toHaveBeenCalledTimes(1); // Should not be called again
    });
  });

  describe('Static Properties', () => {
    it('should have correct static property values', () => {
      expect(MockWebSocket.CONNECTING).toBe(0);
      expect(MockWebSocket.OPEN).toBe(1);
      expect(MockWebSocket.CLOSING).toBe(2);
      expect(MockWebSocket.CLOSED).toBe(3);
    });
  });
});