/**
 * WebSocket Server for Go Sentinel test data
 * Provides real-time updates to connected clients
 */

const WebSocket = require('ws');
const http = require('http');
const express = require('express');

// Sample test data for demonstration
const testData = {
  stats: {
    totalTests: 128,
    passing: 119,
    failing: 9,
    avgDuration: '1.2s'
  },
  recentTests: [
    { name: 'TestParseConfig', status: 'Passed', duration: '0.8s', lastRun: '2 min ago' },
    { name: 'TestValidateInput', status: 'Passed', duration: '1.2s', lastRun: '2 min ago' },
    { name: 'TestProcessResults', status: 'Failed', duration: '2.1s', lastRun: '2 min ago' },
    { name: 'TestExportReport', status: 'Passed', duration: '0.9s', lastRun: '2 min ago' }
  ],
  failingTests: [
    { name: 'TestProcessResults', error: 'Expected 3 items, got 2', failedSince: 'Today' },
    { name: 'TestIntegrationAPI', error: 'Connection timeout', failedSince: 'Yesterday' }
  ]
};

/**
 * Create a WebSocket server attached to an HTTP server
 * @param {http.Server} server - HTTP server to attach to
 * @returns {WebSocket.Server} WebSocket server instance
 */
function setupWebSocketServer(server) {
  const wss = new WebSocket.Server({ server, path: '/ws' });
  
  // Handle WebSocket connections
  wss.on('connection', function connection(ws) {
    console.log('WebSocket client connected');
    
    // Send initial stats
    ws.send(JSON.stringify({
      type: 'stats-update',
      data: testData.stats
    }));
    
    // Setup message handling
    ws.on('message', function incoming(message) {
      try {
        const data = JSON.parse(message);
        handleClientMessage(ws, data);
      } catch (error) {
        console.error('Error processing message:', error);
      }
    });
    
    // Handle disconnection
    ws.on('close', function close() {
      console.log('WebSocket client disconnected');
    });
    
    // Send periodic updates for demonstration
    setupDemoUpdates(ws);
  });
  
  return wss;
}

/**
 * Handle messages from clients
 * @param {WebSocket} ws - WebSocket connection
 * @param {Object} data - Message data
 */
function handleClientMessage(ws, data) {
  switch (data.action) {
    case 'run-test':
      if (data.testName) {
        simulateTestRun(ws, data.testName);
      }
      break;
      
    case 'run-all-tests':
      simulateAllTestsRun(ws);
      break;
      
    case 'get-failing-tests':
      ws.send(JSON.stringify({
        type: 'failing-tests-update',
        data: {
          tests: testData.failingTests
        }
      }));
      break;
      
    default:
      console.log('Unknown client action:', data.action);
  }
}

/**
 * Simulate running a single test
 * @param {WebSocket} ws - WebSocket connection
 * @param {string} testName - Name of test to run
 */
function simulateTestRun(ws, testName) {
  // Find test in the test data
  const test = testData.recentTests.find(t => t.name === testName);
  
  if (!test) return;
  
  // Send "running" status
  ws.send(JSON.stringify({
    type: 'test-result',
    data: {
      testName: test.name,
      status: 'Running',
      lastRun: 'Just now'
    }
  }));
  
  // Wait a short time then send result
  setTimeout(() => {
    // Simulate random result (80% pass rate)
    const passed = Math.random() < 0.8;
    const duration = (parseFloat(test.duration) + Math.random() * 0.2).toFixed(1) + 's';
    
    // Update test data
    test.status = passed ? 'Passed' : 'Failed';
    test.duration = duration;
    test.lastRun = 'Just now';
    
    // Send result
    ws.send(JSON.stringify({
      type: 'test-result',
      data: {
        testName: test.name,
        status: test.status,
        duration: test.duration,
        lastRun: test.lastRun
      }
    }));
    
    // Update stats
    updateStats(ws, passed);
    
    // If test failed, add to failing tests
    if (!passed) {
      const failingTest = testData.failingTests.find(t => t.name === test.name);
      if (!failingTest) {
        testData.failingTests.push({
          name: test.name,
          error: 'Random test failure',
          failedSince: 'Just now'
        });
        
        // Send updated failing tests
        ws.send(JSON.stringify({
          type: 'failing-tests-update',
          data: {
            tests: testData.failingTests
          }
        }));
      }
    } else {
      // If test passed, remove from failing tests if it was there
      const failingIndex = testData.failingTests.findIndex(t => t.name === test.name);
      if (failingIndex >= 0) {
        testData.failingTests.splice(failingIndex, 1);
        
        // Send updated failing tests
        ws.send(JSON.stringify({
          type: 'failing-tests-update',
          data: {
            tests: testData.failingTests
          }
        }));
      }
    }
  }, 1000 + Math.random() * 2000); // Random duration between 1-3 seconds
}

/**
 * Simulate running all tests
 * @param {WebSocket} ws - WebSocket connection
 */
function simulateAllTestsRun(ws) {
  // Notify that all tests are starting
  ws.send(JSON.stringify({
    type: 'test-run-start',
    data: {
      totalTests: testData.recentTests.length
    }
  }));
  
  // Run each test with a delay between them
  let passed = 0;
  let failed = 0;
  
  testData.recentTests.forEach((test, index) => {
    setTimeout(() => {
      simulateTestRun(ws, test.name);
      
      // After last test, send completion
      if (index === testData.recentTests.length - 1) {
        setTimeout(() => {
          ws.send(JSON.stringify({
            type: 'test-run-complete',
            data: {
              passed,
              failed
            }
          }));
        }, 3000);
      }
      
      // Track pass/fail stats
      if (test.status === 'Passed') passed++;
      else failed++;
    }, index * 1500); // Run a new test every 1.5 seconds
  });
}

/**
 * Update stats after a test run
 * @param {WebSocket} ws - WebSocket connection
 * @param {boolean} passed - Whether the test passed
 */
function updateStats(ws, passed) {
  if (passed) {
    testData.stats.passing++;
    testData.stats.failing = Math.max(0, testData.stats.failing - 1);
  } else {
    testData.stats.passing = Math.max(0, testData.stats.passing - 1);
    testData.stats.failing++;
  }
  
  // Send updated stats
  ws.send(JSON.stringify({
    type: 'stats-update',
    data: testData.stats
  }));
}

/**
 * Set up periodic updates for demo purposes
 * @param {WebSocket} ws - WebSocket connection
 */
function setupDemoUpdates(ws) {
  // Periodically simulate a random test run
  const interval = setInterval(() => {
    if (ws.readyState === WebSocket.OPEN) {
      // 10% chance of running a random test
      if (Math.random() < 0.1) {
        const randomTest = testData.recentTests[Math.floor(Math.random() * testData.recentTests.length)];
        simulateTestRun(ws, randomTest.name);
      }
    } else {
      clearInterval(interval);
    }
  }, 10000); // Check every 10 seconds
  
  // Clean up on close
  ws.on('close', () => {
    clearInterval(interval);
  });
}

// Export for use in dev.server.js
module.exports = { setupWebSocketServer };
