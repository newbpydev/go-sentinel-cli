/**
 * Go Sentinel Web Interface
 * Main JavaScript file
 */

document.addEventListener('DOMContentLoaded', function() {
    // Mobile menu toggle functionality
    setupMobileMenu();
    
    // Test selection functionality
    setupTestSelection();
    
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
 * Sets up test row selection with keyboard accessibility
 */
function setupTestSelection() {
    // Get all test rows
    const testRows = document.querySelectorAll('.test-row');
    
    testRows.forEach(row => {
        // Handle click selection
        row.addEventListener('click', function() {
            // Toggle selection state
            row.classList.toggle('selected');
            
            // Update ARIA selected state
            const isSelected = row.classList.contains('selected');
            row.setAttribute('aria-selected', isSelected);
        });
        
        // Keyboard support (Enter/Space to select)
        row.addEventListener('keydown', function(e) {
            if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                row.click(); // Trigger click event
            }
        });
    });
    
    // Add keyboard shortcut ('a' to select all, 'c' to copy)
    document.addEventListener('keydown', function(e) {
        // Only handle if we're focused within the test table
        const testTable = document.querySelector('.test-table');
        if (!testTable.contains(document.activeElement)) return;
        
        if (e.key === 'a' && (e.ctrlKey || e.metaKey)) {
            e.preventDefault();
            
            // Check if all are selected
            const allRows = document.querySelectorAll('.test-row');
            const selectedRows = document.querySelectorAll('.test-row.selected');
            const allSelected = allRows.length === selectedRows.length;
            
            // Toggle selection for all
            allRows.forEach(row => {
                row.classList.toggle('selected', !allSelected);
                row.setAttribute('aria-selected', !allSelected);
            });
        }
    });
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
