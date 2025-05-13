/**
 * Go Sentinel Web Interface
 * Main JavaScript file
 */

document.addEventListener('DOMContentLoaded', function() {
    // Mobile menu toggle functionality
    setupMobileMenu();
    
    // Test selection functionality with enhanced features
    setupEnhancedTestSelection();
    
    // Mock WebSocket connection for demo
    setupMockWebSocket();
});

/**
 * Sets up mobile menu toggle for responsive design
 */
function setupMobileMenu() {
    // Check if we need to add the mobile menu toggle button
    if (window.innerWidth <= 768) {
        // Create toggle button if it doesn't exist
        if (!document.querySelector('.mobile-menu-toggle')) {
            const toggleBtn = document.createElement('button');
            toggleBtn.className = 'mobile-menu-toggle';
            toggleBtn.setAttribute('aria-label', 'Toggle navigation menu');
            toggleBtn.innerHTML = '☰';
            document.body.appendChild(toggleBtn);
            
            // Add event listener
            toggleBtn.addEventListener('click', function() {
                const sidebar = document.querySelector('.sidebar');
                sidebar.classList.toggle('open');
                
                // Update ARIA expanded state
                const isExpanded = sidebar.classList.contains('open');
                toggleBtn.setAttribute('aria-expanded', isExpanded);
            });
        }
    }
    
    // Handle window resize
    window.addEventListener('resize', function() {
        if (window.innerWidth <= 768) {
            setupMobileMenu(); // Ensure toggle exists
        }
    });
}

/**
 * Sets up enhanced test selection with clipboard and keyboard shortcuts
 */
function setupEnhancedTestSelection() {
    // Cache DOM elements
    const testRows = document.querySelectorAll('.test-row');
    const testCheckboxes = document.querySelectorAll('.test-checkbox');
    const selectAllCheckbox = document.getElementById('select-all-checkbox');
    const selectAllBtn = document.getElementById('select-all-btn');
    const copySelectedBtn = document.getElementById('copy-selected-btn');
    const runSelectedBtn = document.getElementById('run-selected-btn');
    const selectionInfoEl = document.getElementById('selection-info');
    const copyAreaEl = document.getElementById('copy-area');
    
    // Track selection state
    let selectedTests = new Set();
    let lastSelectedIndex = -1;
    
    // Update selection count UI
    function updateSelectionInfo() {
        const count = selectedTests.size;
        selectionInfoEl.querySelector('.selection-count').textContent = 
            count === 0 ? 'No tests selected' : 
            count === 1 ? '1 test selected' : 
            `${count} tests selected`;
            
        // Update action button states
        copySelectedBtn.disabled = count === 0;
        runSelectedBtn.disabled = count === 0;
        
        // Update master checkbox state
        if (count === 0) {
            selectAllCheckbox.checked = false;
            selectAllCheckbox.indeterminate = false;
        } else if (count === testRows.length) {
            selectAllCheckbox.checked = true;
            selectAllCheckbox.indeterminate = false;
        } else {
            selectAllCheckbox.indeterminate = true;
        }
    }
    
    // Select/deselect a test row
    function toggleTestSelection(row, selected, updateCheckbox = true) {
        const testId = row.dataset.testId;
        
        if (selected) {
            selectedTests.add(testId);
            row.classList.add('selected');
            row.setAttribute('aria-selected', 'true');
            if (updateCheckbox) {
                row.querySelector('.test-checkbox').checked = true;
            }
        } else {
            selectedTests.delete(testId);
            row.classList.remove('selected');
            row.setAttribute('aria-selected', 'false');
            if (updateCheckbox) {
                row.querySelector('.test-checkbox').checked = false;
            }
        }
        
        updateSelectionInfo();
    }
    
    // Select/deselect all test rows
    function toggleAllTests(selected) {
        testRows.forEach(row => toggleTestSelection(row, selected));
    }
    
    // Handle click on individual checkboxes
    testCheckboxes.forEach((checkbox, index) => {
        checkbox.addEventListener('change', function(e) {
            e.stopPropagation(); // Prevent row click handler from firing
            const row = this.closest('.test-row');
            toggleTestSelection(row, this.checked, false);
            lastSelectedIndex = index;
        });
    });
    
    // Handle click on test rows
    testRows.forEach((row, index) => {
        row.addEventListener('click', function(e) {
            // Ignore clicks on checkbox and buttons
            if (e.target.type === 'checkbox' || e.target.tagName === 'BUTTON') return;
            
            const checkbox = row.querySelector('.test-checkbox');
            
            // Handle shift+click for range selection
            if (e.shiftKey && lastSelectedIndex !== -1) {
                const start = Math.min(lastSelectedIndex, index);
                const end = Math.max(lastSelectedIndex, index);
                
                for (let i = start; i <= end; i++) {
                    toggleTestSelection(testRows[i], true);
                }
            } 
            // Handle ctrl/cmd+click for toggling individual items
            else if (e.ctrlKey || e.metaKey) {
                toggleTestSelection(row, !checkbox.checked);
                lastSelectedIndex = index;
            } 
            // Normal click - deselect others and select this one
            else {
                toggleAllTests(false);
                toggleTestSelection(row, true);
                lastSelectedIndex = index;
            }
        });
        
        // Keyboard navigation for rows
        row.addEventListener('keydown', function(e) {
            if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                row.click();
            }
        });
    });
    
    // Handle select all checkbox
    selectAllCheckbox.addEventListener('change', function() {
        toggleAllTests(this.checked);
    });
    
    // Handle select all button
    selectAllBtn.addEventListener('click', function() {
        const allSelected = selectedTests.size === testRows.length;
        toggleAllTests(!allSelected);
    });
    
    // Handle copy selected button
    copySelectedBtn.addEventListener('click', function() {
        copySelectedTests();
    });
    
    // Handle run selected button
    runSelectedBtn.addEventListener('click', function() {
        // In a real implementation, this would trigger test runs
        alert(`Running ${selectedTests.size} selected tests...`);
    });
    
    // Copy selected tests to clipboard
    function copySelectedTests() {
        // Get test names from selected rows
        const selectedTestNames = [];
        testRows.forEach(row => {
            if (selectedTests.has(row.dataset.testId)) {
                selectedTestNames.push(row.dataset.testName);
            }
        });
        
        // Copy to clipboard
        if (selectedTestNames.length > 0) {
            copyAreaEl.value = selectedTestNames.join('\n');
            copyAreaEl.select();
            document.execCommand('copy');
            
            // Show success feedback
            const prevCount = selectionInfoEl.querySelector('.selection-count').textContent;
            selectionInfoEl.querySelector('.selection-count').textContent = 
                `✓ Copied ${selectedTests.size} tests to clipboard!`;
            
            // Reset after 2 seconds
            setTimeout(() => {
                selectionInfoEl.querySelector('.selection-count').textContent = prevCount;
            }, 2000);
        }
    }
    
    // Global keyboard shortcuts
    document.addEventListener('keydown', function(e) {
        // Only if we're within the test table container
        const testTableContainer = document.querySelector('.test-table-container');
        if (!testTableContainer.contains(document.activeElement) && 
            document.activeElement !== document.body) return;
        
        // Ctrl/Cmd+A to select all
        if (e.key === 'a' && (e.ctrlKey || e.metaKey)) {
            e.preventDefault();
            toggleAllTests(selectedTests.size !== testRows.length);
        }
        
        // Ctrl/Cmd+C to copy selected
        if (e.key === 'c' && (e.ctrlKey || e.metaKey) && selectedTests.size > 0) {
            if (window.getSelection().toString() === '') { // Only if no text is selected
                e.preventDefault();
                copySelectedTests();
            }
        }
        
        // Ctrl/Cmd+R to run selected
        if (e.key === 'r' && (e.ctrlKey || e.metaKey) && selectedTests.size > 0) {
            e.preventDefault();
            runSelectedBtn.click();
        }
        
        // Escape to clear selection
        if (e.key === 'Escape') {
            e.preventDefault();
            toggleAllTests(false);
        }
    });
    
    // Initialize selection info
    updateSelectionInfo();
}

/**
 * Sets up a mock WebSocket connection for demonstration
 */
function setupMockWebSocket() {
    // Update connection status to simulate connecting
    const statusIndicator = document.querySelector('.status-indicator');
    
    // Simulate connection after a delay
    setTimeout(() => {
        statusIndicator.classList.remove('connecting');
        statusIndicator.classList.add('connected');
        statusIndicator.textContent = 'Connected';
        
        // Register HTMX extension for WebSocket
        document.body.setAttribute('hx-ext', 'ws');
        
        // Enable WebSocket endpoint with re-connect on page
        document.body.setAttribute('ws-connect', '/ws');
    }, 2000);
    
    // Simulate occasional disconnections for testing
    setInterval(() => {
        const random = Math.random();
        
        if (random < 0.1) { // 10% chance of disconnect
            statusIndicator.classList.remove('connected');
            statusIndicator.classList.add('disconnected');
            statusIndicator.textContent = 'Disconnected';
            
            // Simulate reconnection attempt
            setTimeout(() => {
                statusIndicator.classList.remove('disconnected');
                statusIndicator.classList.add('connecting');
                statusIndicator.textContent = 'Reconnecting...';
                
                // Simulate successful reconnection
                setTimeout(() => {
                    statusIndicator.classList.remove('connecting');
                    statusIndicator.classList.add('connected');
                    statusIndicator.textContent = 'Connected';
                }, 1500);
            }, 1000);
        }
    }, 30000); // Check every 30 seconds
}
