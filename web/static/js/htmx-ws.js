/**
 * HTMX WebSocket Extension for Go Sentinel
 * Implements WebSocket connection with automatic reconnection
 * and message type routing for real-time test updates
 */

// Initialize the HTMX WebSocket extension
(function() {
  // Create HTMX extension
  htmx.defineExtension('ws-connect', {
    
    init: function(apiRef) {
      // Store reference to the API
      this.apiRef = apiRef;
      
      // Store WebSocket connections by id
      this.sockets = {};
      
      // Register events so we can trigger them elsewhere
      this.apiRef.onEvent('htmx:wsBeforeConnect', function(evt) {
        console.log('WebSocket connection starting:', evt.detail.socketId);
      });
      
      this.apiRef.onEvent('htmx:wsConnected', function(evt) {
        console.log('WebSocket connected:', evt.detail.socketId);
        // Update connection status indicator
        const statusEl = document.querySelector('.connection-status');
        if (statusEl) {
          statusEl.className = 'connection-status connected';
          statusEl.querySelector('.status-indicator').className = 'status-indicator connected';
          statusEl.querySelector('.status-text').textContent = 'Connected';
        }
      });
      
      this.apiRef.onEvent('htmx:wsReconnecting', function(evt) {
        console.log('WebSocket reconnecting:', evt.detail.socketId);
        // Update connection status indicator
        const statusEl = document.querySelector('.connection-status');
        if (statusEl) {
          statusEl.className = 'connection-status connecting';
          statusEl.querySelector('.status-indicator').className = 'status-indicator connecting';
          statusEl.querySelector('.status-text').textContent = 'Reconnecting...';
        }
      });
      
      this.apiRef.onEvent('htmx:wsBeforeMessage', function(evt) {
        console.log('WebSocket message received:', evt.detail.socketId, evt.detail.message);
      });
    },
    
    onEvent: function(name, evt) {
      // Initialize WebSockets on HTMX afterLoad
      if (name === 'htmx:afterOnLoad') {
        this.processWebSocketElements(document.body);
      }
      
      // Handle WebSocket elements added to the DOM dynamically
      if (name === 'htmx:afterSettle') {
        this.processWebSocketElements(evt.detail.elt);
      }
    },
    
    // Process all elements with hx-ws attribute
    processWebSocketElements: function(elt) {
      const wsElements = elt.querySelectorAll('[hx-ws]');
      for (const wsElt of wsElements) {
        this.initWebSocket(wsElt);
      }
    },
    
    // Initialize WebSocket for an element
    initWebSocket: function(elt) {
      const socketId = elt.getAttribute('id') || 'socket-' + Math.random().toString(36).substring(2);
      const wsUrl = elt.getAttribute('hx-ws');
      
      // Skip if already connected or no URL provided
      if (this.sockets[socketId] || !wsUrl.startsWith('connect:')) {
        return;
      }
      
      // Extract WebSocket URL
      const socketUrl = wsUrl.substring(8);
      
      // Store element reference for events
      elt.socketId = socketId;
      
      // Trigger connection event
      this.apiRef.triggerEvent(elt, 'htmx:wsBeforeConnect', {
        socketId: socketId,
        socketUrl: socketUrl
      });
      
      // Create and configure WebSocket
      this.createWebSocket(socketId, socketUrl, elt);
    },
    
    // Create WebSocket connection with automatic reconnection
    createWebSocket: function(socketId, socketUrl, elt) {
      const self = this;
      const socket = new WebSocket(socketUrl);
      
      // Store socket in global window for test access
      window.socket = socket;
      
      // Also store in our local collection
      this.sockets[socketId] = socket;
      
      // Configure socket event handlers
      socket.onopen = function() {
        self.apiRef.triggerEvent(elt, 'htmx:wsConnected', {
          socketId: socketId,
          socket: socket
        });
      };
      
      socket.onmessage = function(event) {
        const message = event.data;
        
        // Process message before triggering event
        self.apiRef.triggerEvent(elt, 'htmx:wsBeforeMessage', {
          socketId: socketId,
          message: message
        });
        
        // Parse the message if it's JSON
        let messageData = message;
        try {
          messageData = JSON.parse(message);
        } catch(e) {
          // Keep as string if not valid JSON
        }
        
        // Route message based on type
        self.routeMessage(elt, messageData);
        
        // Trigger after message event
        self.apiRef.triggerEvent(elt, 'htmx:wsAfterMessage', {
          socketId: socketId,
          message: message
        });
      };
      
      socket.onclose = function(event) {
        // Remove from collection
        delete self.sockets[socketId];
        
        // Update connection status indicator
        const statusEl = document.querySelector('.connection-status');
        if (statusEl) {
          statusEl.className = 'connection-status disconnected';
          statusEl.querySelector('.status-indicator').className = 'status-indicator disconnected';
          statusEl.querySelector('.status-text').textContent = 'Disconnected';
        }
        
        // Trigger event
        self.apiRef.triggerEvent(elt, 'htmx:wsClosed', {
          socketId: socketId,
          code: event.code,
          reason: event.reason
        });
        
        // Reconnect after delay if not explicitly closed
        if (event.code !== 1000) {
          self.apiRef.triggerEvent(elt, 'htmx:wsReconnecting', {
            socketId: socketId
          });
          
          // Reconnect after delay
          setTimeout(function() {
            self.createWebSocket(socketId, socketUrl, elt);
          }, 3000); // 3-second reconnection delay
        }
      };
      
      socket.onerror = function(error) {
        console.error('WebSocket error:', error);
        self.apiRef.triggerEvent(elt, 'htmx:wsError', {
          socketId: socketId,
          error: error
        });
      };
      
      // Expose reconnect function for tests
      window.reconnectWebSocket = function() {
        if (socket) {
          socket.close();
        }
      };
    },
    
    // Route message based on type to the appropriate UI element
    routeMessage: function(elt, messageData) {
      // Handle different message types
      if (typeof messageData === 'object' && messageData.type) {
        switch(messageData.type) {
          // Frontend demo message types
          case 'stats-update':
            this.updateStats(messageData.data);
            break;
          case 'test-run-complete':
            this.handleTestRunComplete(messageData.data);
            break;
          case 'failing-tests-update':
            this.updateFailingTests(messageData.data);
            break;
            
          // Go backend message types (match internal/api/websocket/message_types.go)
          case 'test_result':
            this.handleTestResult(messageData.payload);
            break;
          case 'command':
            this.handleCommand(messageData.payload);
            break;
            
          // Extended message types for richer UI
          case 'test_suite_started':
            this.handleTestSuiteStarted(messageData.payload);
            break;
          case 'test_suite_completed':
            this.handleTestSuiteCompleted(messageData.payload);
            break;
          case 'test_metrics':
            this.handleTestMetrics(messageData.payload);
            break;
          default:
            console.log('Unknown message type:', messageData.type);
        }
      }
    },
    
    // Update stats grid with new data
    updateStats: function(data) {
      if (data.totalTests) {
        const totalTestsEl = document.querySelector('.stats-grid .metric-value:nth-child(1)');
        if (totalTestsEl) totalTestsEl.textContent = data.totalTests;
      }
      
      if (data.passing) {
        const passingEl = document.querySelector('.stats-grid .metric-card:nth-child(2) .metric-value');
        if (passingEl) passingEl.textContent = data.passing;
      }
      
      if (data.failing) {
        const failingEl = document.querySelector('.stats-grid .metric-card:nth-child(3) .metric-value');
        if (failingEl) failingEl.textContent = data.failing;
      }
      
      if (data.avgDuration) {
        const durationEl = document.querySelector('.stats-grid .metric-card:nth-child(4) .metric-value');
        if (durationEl) durationEl.textContent = data.avgDuration;
      }
    },
    
    // Handle test result from our Go backend
    handleTestResult: function(payload) {
      if (!payload || !payload.test_id) return;
      
      // Look for the test row in the recent tests table
      const testRows = document.querySelectorAll('table tbody tr');
      for (const row of testRows) {
        const nameCell = row.querySelector('td:first-child');
        if (nameCell && nameCell.textContent.trim() === payload.test_id) {
          // Update status
          if (payload.status) {
            const statusCell = row.querySelector('td:nth-child(2)');
            if (statusCell) {
              const isPassed = payload.status.toLowerCase() === 'passed';
              statusCell.innerHTML = `<span class="badge badge-${isPassed ? 'success' : 'error'}">${payload.status}</span>`;
            }
          }
          
          // Update timestamp with current time
          const lastRunCell = row.querySelector('td:nth-child(4)');
          if (lastRunCell) lastRunCell.textContent = 'Just now';
          
          // Highlight the row briefly
          row.classList.add('row-updated');
          setTimeout(() => {
            row.classList.remove('row-updated');
          }, 2000);
          
          // Update failing tests list if applicable
          if (payload.status.toLowerCase() === 'failed') {
            this.addToFailingTests(payload.test_id, payload.error || 'Unknown error');
          } else {
            this.removeFromFailingTests(payload.test_id);
          }
          
          break;
        }
      }
    },
    
    // Handle command messages from the backend
    handleCommand: function(payload) {
      if (!payload || !payload.command) return;
      
      console.log('Received command:', payload.command, payload.args);
      
      // Execute appropriate action based on command
      switch(payload.command) {
        case 'refresh':
          // Refresh the page or specific component
          window.location.reload();
          break;
          
        case 'update_ui':
          // Trigger an update of specific UI elements
          if (payload.args && payload.args.length > 0) {
            const targetId = payload.args[0];
            const el = document.getElementById(targetId);
            if (el && typeof htmx !== 'undefined') {
              htmx.trigger(el, 'refresh');
            }
          }
          break;
          
        default:
          console.log('Unknown command type:', payload.command);
      }
    },
    
    // Update a single test result (for backward compatibility with demo)
    updateTestResult: function(data) {
      if (!data.testName) return;
      
      // Look for the test row in the recent tests table
      const testRows = document.querySelectorAll('table tbody tr');
      for (const row of testRows) {
        const nameCell = row.querySelector('td:first-child');
        if (nameCell && nameCell.textContent.trim() === data.testName) {
          // Update status
          if (data.status) {
            const statusCell = row.querySelector('td:nth-child(2)');
            if (statusCell) {
              statusCell.innerHTML = `<span class="badge badge-${data.status.toLowerCase() === 'passed' ? 'success' : 'error'}">${ data.status}</span>`;
            }
          }
          
          // Update duration
          if (data.duration) {
            const durationCell = row.querySelector('td:nth-child(3)');
            if (durationCell) durationCell.textContent = data.duration;
          }
          
          // Update last run time
          if (data.lastRun) {
            const lastRunCell = row.querySelector('td:nth-child(4)');
            if (lastRunCell) lastRunCell.textContent = data.lastRun;
          }
          
          // Highlight the row briefly
          row.classList.add('row-updated');
          setTimeout(() => {
            row.classList.remove('row-updated');
          }, 2000);
          
          break;
        }
      }
    },
    
    // Handle test run completion message
    handleTestRunComplete: function(data) {
      // Flash a notification or update UI to show run completion
      const notification = document.createElement('div');
      notification.className = 'notification';
      notification.textContent = `Test run complete. ${data.passed} passed, ${data.failed} failed.`;
      document.body.appendChild(notification);
      
      // Remove notification after a delay
      setTimeout(() => {
        notification.remove();
      }, 5000);
    },
    
    // Extended message handlers for richer UI
    handleTestSuiteStarted: function(payload) {
      console.log('Test suite started:', payload);
      
      // Update UI to show test suite is running
      const notification = document.createElement('div');
      notification.className = 'notification';
      notification.textContent = `Starting test suite: ${payload.name || 'Unnamed suite'}`;
      document.body.appendChild(notification);
      
      // Remove notification after a delay
      setTimeout(() => {
        notification.remove();
      }, 5000);
      
      // Update run all tests button to show it's in progress
      const runAllButton = document.querySelector('button[hx-ws="send:{action:\'run-all-tests\'}"]');
      if (runAllButton) {
        runAllButton.innerHTML = '<span class="spinner"></span> Running...';
        runAllButton.disabled = true;
      }
    },
    
    handleTestSuiteCompleted: function(payload) {
      console.log('Test suite completed:', payload);
      
      // Show completion notification
      const notification = document.createElement('div');
      notification.className = 'notification';
      notification.textContent = `Test suite completed: ${payload.passed} passed, ${payload.failed} failed in ${payload.duration || '0s'}`;
      document.body.appendChild(notification);
      
      // Remove notification after a delay
      setTimeout(() => {
        notification.remove();
      }, 5000);
      
      // Update run all tests button
      const runAllButton = document.querySelector('button[hx-ws="send:{action:\'run-all-tests\'}"]');
      if (runAllButton) {
        runAllButton.innerHTML = 'Run All Tests';
        runAllButton.disabled = false;
      }
      
      // Update statistics
      if (payload.stats) {
        this.updateStats({
          totalTests: payload.stats.total || 0,
          passing: payload.stats.passed || 0,
          failing: payload.stats.failed || 0,
          avgDuration: payload.stats.avgDuration || '0s'
        });
      }
    },
    
    handleTestMetrics: function(payload) {
      console.log('Test metrics received:', payload);
      
      // Update stats with metrics data
      if (payload) {
        this.updateStats({
          totalTests: payload.totalTests || 0,
          passing: payload.passing || 0,
          failing: payload.failing || 0,
          avgDuration: payload.avgDuration || '0s'
        });
      }
    },
    
    // Helper methods for failing tests management
    addToFailingTests: function(testId, errorMessage) {
      const failingTestsContainer = document.querySelector('.dashboard-card:nth-child(2) tbody');
      if (!failingTestsContainer) return;
      
      // Check if test is already in failing tests
      const existingRows = failingTestsContainer.querySelectorAll('tr');
      for (const row of existingRows) {
        const nameCell = row.querySelector('td:first-child');
        if (nameCell && nameCell.textContent.trim() === testId) {
          // Update error message
          const errorCell = row.querySelector('td:nth-child(2)');
          if (errorCell) errorCell.textContent = errorMessage;
          return;
        }
      }
      
      // Add new failing test
      const row = document.createElement('tr');
      row.innerHTML = `
        <td>${testId}</td>
        <td>${errorMessage}</td>
        <td>Just now</td>
        <td>
          <button class="btn btn-secondary">Debug</button>
          <button class="btn btn-primary">Fix</button>
        </td>
      `;
      failingTestsContainer.appendChild(row);
    },
    
    removeFromFailingTests: function(testId) {
      const failingTestsContainer = document.querySelector('.dashboard-card:nth-child(2) tbody');
      if (!failingTestsContainer) return;
      
      // Find and remove the test from failing tests
      const existingRows = failingTestsContainer.querySelectorAll('tr');
      for (const row of existingRows) {
        const nameCell = row.querySelector('td:first-child');
        if (nameCell && nameCell.textContent.trim() === testId) {
          row.remove();
          break;
        }
      }
    },
    
    // Update failing tests section (for backward compatibility with demo)
    updateFailingTests: function(data) {
      if (!data.tests || !Array.isArray(data.tests)) return;
      
      const failingTestsContainer = document.querySelector('.dashboard-card:nth-child(2) tbody');
      if (!failingTestsContainer) return;
      
      // Clear existing failing tests
      failingTestsContainer.innerHTML = '';
      
      // Add new failing tests
      data.tests.forEach(test => {
        const row = document.createElement('tr');
        row.innerHTML = `
          <td>${test.name}</td>
          <td>${test.error}</td>
          <td>${test.failedSince}</td>
          <td>
            <button class="btn btn-secondary">Debug</button>
            <button class="btn btn-primary">Fix</button>
          </td>
        `;
        failingTestsContainer.appendChild(row);
      });
    }
  });
  
  // -----------------------------------------------------------
  // Test Selection Mode (similar to CLI interactive selection)
  // -----------------------------------------------------------
  
  // Initialize test selection mode system
  window.testSelectionMode = {
    active: false,
    selectedTests: {},
    visibleTests: [],
    
    // Enter selection mode
    enter: function() {
      this.active = true;
      this.selectedTests = {};
      
      // Get all currently visible test rows
      const testRows = document.querySelectorAll('table tbody tr');
      this.visibleTests = Array.from(testRows);
      
      // Add selection mode styling
      document.body.classList.add('selection-mode');
      
      // Add index indicators to rows
      this.visibleTests.forEach((row, index) => {
        // Skip if index is greater than 9 (0-9 keys only)
        if (index > 9) return;
        
        const indexIndicator = document.createElement('div');
        indexIndicator.className = 'selection-index';
        indexIndicator.textContent = index;
        row.classList.add('selectable');
        row.setAttribute('data-index', index);
        row.prepend(indexIndicator);
      });
      
      // Show selection mode UI
      this.showSelectionUI();
      
      // Set up key handlers
      document.addEventListener('keydown', this.handleKeyPress);
    },
    
    // Exit selection mode
    exit: function() {
      this.active = false;
      
      // Remove selection mode styling
      document.body.classList.remove('selection-mode');
      
      // Remove index indicators
      document.querySelectorAll('.selection-index').forEach(el => el.remove());
      document.querySelectorAll('.selectable').forEach(row => {
        row.classList.remove('selectable', 'selected');
        row.removeAttribute('data-index');
      });
      
      // Hide selection UI
      this.hideSelectionUI();
      
      // Remove key handler
      document.removeEventListener('keydown', this.handleKeyPress);
    },
    
    // Toggle selection of a test by index
    toggleSelection: function(index) {
      if (index < 0 || index >= this.visibleTests.length) return;
      
      const row = this.visibleTests[index];
      const isSelected = row.classList.contains('selected');
      
      if (isSelected) {
        row.classList.remove('selected');
        delete this.selectedTests[index];
      } else {
        row.classList.add('selected');
        this.selectedTests[index] = row;
      }
      
      // Update selection UI
      this.updateSelectionUI();
    },
    
    // Toggle all selections
    toggleAll: function() {
      const hasSelected = Object.keys(this.selectedTests).length > 0;
      
      if (hasSelected) {
        // Deselect all
        this.visibleTests.forEach(row => row.classList.remove('selected'));
        this.selectedTests = {};
      } else {
        // Select all
        this.visibleTests.forEach((row, index) => {
          row.classList.add('selected');
          this.selectedTests[index] = row;
        });
      }
      
      // Update selection UI
      this.updateSelectionUI();
    },
    
    // Handle keypress events in selection mode
    handleKeyPress: function(event) {
      const mode = window.testSelectionMode;
      
      // Handle number keys (0-9)
      if (/^[0-9]$/.test(event.key)) {
        const index = parseInt(event.key);
        mode.toggleSelection(index);
        return;
      }
      
      // Handle other special keys
      switch(event.key.toLowerCase()) {
        case 'a': // Toggle all
          mode.toggleAll();
          break;
        case 'enter': // Copy selected to clipboard
          mode.copySelectedToClipboard();
          break;
        case 'q': // Exit selection mode
        case 'escape':
          mode.exit();
          break;
      }
    },
    
    // Copy selected tests to clipboard
    copySelectedToClipboard: function() {
      const selectedIndices = Object.keys(this.selectedTests);
      if (selectedIndices.length === 0) return;
      
      const selectedTests = selectedIndices.map(index => {
        const row = this.selectedTests[index];
        const testName = row.querySelector('td:first-child').textContent.trim();
        return testName;
      });
      
      // Copy to clipboard
      navigator.clipboard.writeText(selectedTests.join(' ')).then(() => {
        // Show feedback
        this.showCopyFeedback(selectedTests.length);
      });
    },
    
    // Show selection mode UI
    showSelectionUI: function() {
      const selectionUI = document.createElement('div');
      selectionUI.className = 'selection-ui';
      selectionUI.innerHTML = `
        <div class="selection-ui-header">Selection Mode</div>
        <div class="selection-ui-body">
          <div class="selection-count">0 tests selected</div>
          <div class="selection-help">
            <p><kbd>0-9</kbd> Toggle selection</p>
            <p><kbd>a</kbd> Select/deselect all</p>
            <p><kbd>Enter</kbd> Copy selected to clipboard</p>
            <p><kbd>q</kbd> Exit selection mode</p>
          </div>
        </div>
      `;
      document.body.appendChild(selectionUI);
    },
    
    // Hide selection mode UI
    hideSelectionUI: function() {
      const selectionUI = document.querySelector('.selection-ui');
      if (selectionUI) selectionUI.remove();
    },
    
    // Update selection UI count
    updateSelectionUI: function() {
      const countEl = document.querySelector('.selection-count');
      if (!countEl) return;
      
      const count = Object.keys(this.selectedTests).length;
      countEl.textContent = `${count} test${count === 1 ? '' : 's'} selected`;
    },
    
    // Show copy feedback
    showCopyFeedback: function(count) {
      const feedback = document.createElement('div');
      feedback.className = 'copy-feedback';
      feedback.textContent = `${count} test${count === 1 ? '' : 's'} copied to clipboard`;
      document.body.appendChild(feedback);
      
      setTimeout(() => {
        feedback.remove();
      }, 2000);
    }
  };
  
  // Add keyboard shortcut to enter selection mode
  document.addEventListener('keydown', function(event) {
    // Press 'c' to enter selection mode (similar to CLI)
    if (event.key.toLowerCase() === 'c' && !window.testSelectionMode.active) {
      window.testSelectionMode.enter();
    }
  });
  
})();
