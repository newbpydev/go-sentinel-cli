// Playwright test file for WebSocket connection tests
const { test, expect } = require('@playwright/test');

test.describe('WebSocket Connection', () => {
  test('establishes WebSocket connection on page load', async ({ page }) => {
    // Create a promise that resolves when a WebSocket connection is established
    const wsConnectionPromise = page.waitForEvent('websocket');
    
    // Navigate to the page
    await page.goto('http://localhost:5173/');
    
    // Wait for the WebSocket connection and verify it
    const wsConnection = await wsConnectionPromise;
    expect(wsConnection.url()).toContain('/ws'); // Assuming the WebSocket endpoint is at /ws
  });
  
  test('reconnects when WebSocket connection is dropped', async ({ page }) => {
    // First establish a connection
    const wsConnectionPromise = page.waitForEvent('websocket');
    await page.goto('http://localhost:5173/');
    const wsConnection = await wsConnectionPromise;
    
    // Listen for a new WebSocket connection after closing the current one
    const wsReconnectionPromise = page.waitForEvent('websocket');
    
    // Simulate connection drop by injecting code that closes WebSocket and triggers reconnect
    await page.evaluate(() => {
      // Assuming the app stores the socket instance in a global variable or we can access it
      if (window.socket && window.socket.close) {
        window.socket.close();
        // Trigger reconnect mechanism (this will depend on your implementation)
        if (typeof window.reconnectWebSocket === 'function') {
          window.reconnectWebSocket();
        }
      }
    });
    
    // Verify reconnection occurred
    const newConnection = await wsReconnectionPromise;
    expect(newConnection.url()).toContain('/ws');
  });
  
  test('binds WebSocket events to DOM updates', async ({ page }) => {
    // Navigate to the page
    await page.goto('http://localhost:5173/');
    
    // Wait for initial content to load
    await expect(page.locator('.stats-grid')).toBeVisible();
    
    // Initial test count
    const initialTestCount = await page.locator('.stats-grid .metric-value').first().textContent();
    
    // Simulate receiving a WebSocket message that updates test count
    await page.evaluate(() => {
      // Create a custom event that mimics what HTMX will respond to
      const event = new CustomEvent('htmx:wsAfterMessage', {
        detail: {
          message: JSON.stringify({
            type: 'stats-update',
            data: {
              totalTests: '129', // One more than initial 128
              passing: '120',    // One more than initial 119
              failing: '9',
              avgDuration: '1.2s'
            }
          })
        }
      });
      document.body.dispatchEvent(event);
    });
    
    // Check that the UI has been updated with the new value
    await expect(page.locator('.stats-grid .metric-value').first()).toHaveText('129');
  });
});
