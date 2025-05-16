/**
 * Coverage Visualization JavaScript
 * Handles interactive elements for the coverage visualization page
 */

document.addEventListener('DOMContentLoaded', function() {
    // Set width of metric fill bars based on data-percentage attributes
    function setMetricFillWidths() {
        document.querySelectorAll('.metric-fill[data-percentage]').forEach(function(el) {
            const percentage = el.getAttribute('data-percentage');
            if (percentage) {
                el.style.width = percentage + '%';
            }
        });
    }
    
    // Initial setup
    setMetricFillWidths();
    
    // Set up observer for dynamic content
    const observer = new MutationObserver(function(mutations) {
        mutations.forEach(function(mutation) {
            if (mutation.type === 'childList' && mutation.addedNodes.length > 0) {
                setMetricFillWidths();
            }
        });
    });
    
    // Observe the coverage dashboard for changes
    const dashboard = document.querySelector('.coverage-dashboard');
    if (dashboard) {
        observer.observe(dashboard, { childList: true, subtree: true });
    }

    // Initialize file details section visibility
    const fileDetails = document.getElementById('file-details');
    const closeDetailsBtn = document.getElementById('close-details');
    
    // Function to show file details panel
    window.showFileDetails = function() {
        fileDetails.classList.add('active');
        fileDetails.style.display = 'block';
    };
    
    // Close file details when close button is clicked
    if (closeDetailsBtn) {
        closeDetailsBtn.addEventListener('click', function() {
            fileDetails.classList.remove('active');
            fileDetails.style.display = 'none';
        });
    }
    
    // Handle threshold filter changes
    const thresholdFilter = document.getElementById('coverage-threshold');
    if (thresholdFilter) {
        thresholdFilter.addEventListener('change', function() {
            // HTMX will handle the actual request
            // This is just for any additional UI updates
            updateFilterIndicator(thresholdFilter.value);
        });
    }
    
    // Handle search input
    const searchInput = document.getElementById('coverage-search');
    if (searchInput) {
        // Clear button for search
        searchInput.addEventListener('search', function() {
            if (this.value === '') {
                // If search is cleared, reset to show all files
                htmx.trigger('#coverage-list', 'hx-get', {url: '/api/coverage/files'});
            }
        });
    }
    
    // Function to update any UI indicators for active filters
    function updateFilterIndicator(filterValue) {
        // Could add visual indicators for active filters here
        console.log('Filter applied:', filterValue);
    }
    
    // Add syntax highlighting to code blocks (if using a library like highlight.js)
    // This would be called after loading file details
    function highlightCode() {
        // If using highlight.js:
        // document.querySelectorAll('pre code').forEach((block) => {
        //     hljs.highlightBlock(block);
        // });
    }
    
    // Register for WebSocket events if available
    if (typeof htmx !== 'undefined' && htmx.createWebSocket) {
        document.body.addEventListener('coverage-updated', function(event) {
            // Refresh coverage data when notified of updates
            htmx.trigger('#coverage-summary', 'hx-get', {url: '/api/coverage/summary'});
            htmx.trigger('#coverage-list', 'hx-get', {url: '/api/coverage/files'});
            
            // Show notification
            showToast('success', 'Coverage data has been updated');
        });
    }
    
    // Add keyboard navigation for file list
    document.addEventListener('keydown', function(event) {
        // Only handle keys when file details are visible
        if (fileDetails.style.display === 'block') {
            if (event.key === 'Escape') {
                // Close details panel on Escape
                closeDetailsBtn.click();
            }
        }
    });
});

// Helper function to format coverage percentage with appropriate color class
function getCoverageClass(percentage) {
    if (percentage >= 80) return 'high-coverage';
    if (percentage >= 50) return 'medium-coverage';
    return 'low-coverage';
}

// Pagination helper functions
function goToPage(page, filter, search) {
    htmx.ajax('GET', `/api/coverage/files?page=${page}&filter=${filter || 'all'}&search=${search || ''}`, {
        target: '#coverage-list'
    });
}
