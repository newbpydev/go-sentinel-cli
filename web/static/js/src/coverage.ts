/**
 * Coverage Visualization TypeScript
 * Handles interactive elements for the coverage visualization page
 */

/**
 * Coverage data interface
 */
interface CoverageMetric {
  total: number;
  covered: number;
  percentage: number;
}

interface FileCoverage {
  path: string;
  name: string;
  statements: CoverageMetric;
  branches: CoverageMetric;
  functions: CoverageMetric;
  lines: CoverageMetric;
  status: 'pass' | 'warning' | 'fail';
}

interface CoverageUpdateEvent extends CustomEvent {
  detail: {
    data: {
      files: FileCoverage[];
      summary: {
        statements: CoverageMetric;
        branches: CoverageMetric;
        functions: CoverageMetric;
        lines: CoverageMetric;
      };
    };
  };
}

// Global variables
let fileDetails: HTMLElement | null = null;
let closeDetailsBtn: HTMLElement | null = null;
let observer: MutationObserver | null = null;

/**
 * Set width of metric fill bars based on data-percentage attributes
 */
function setMetricFillWidths(): void {
  document.querySelectorAll<HTMLElement>('.metric-fill[data-percentage]').forEach((el) => {
    const percentage = el.getAttribute('data-percentage');
    if (percentage) {
      el.style.width = `${percentage}%`;
    }
  });
}

/**
 * Helper function to format coverage percentage with appropriate color class
 * @param percentage - Coverage percentage (0-100)
 * @returns CSS class name based on coverage percentage
 */
function getCoverageClass(percentage: number): string {
  if (percentage >= 90) return 'high';
  if (percentage >= 70) return 'medium';
  if (percentage >= 50) return 'low';
  return 'very-low';
}

/**
 * Navigate to a specific page in the coverage file list
 * @param page - Page number to navigate to
 * @param filter - Optional filter to apply
 * @param search - Optional search term
 */
function goToPage(page: number, filter = 'all', search = ''): void {
  if (window.htmx?.ajax) {
    window.htmx.ajax('GET', `/api/coverage/files?page=${page}&filter=${filter}&search=${search}`, {
      target: '#coverage-list',
    });
  }
}

// Expose functions to window for global access
(window as any).setMetricFillWidths = setMetricFillWidths;
(window as any).getCoverageClass = getCoverageClass;
(window as any).goToPage = goToPage;
(globalThis as any).setMetricFillWidths = setMetricFillWidths;

document.addEventListener('DOMContentLoaded', () => {
  // Initial setup
  setMetricFillWidths();

  // Cache DOM elements
  fileDetails = document.getElementById('file-details');
  closeDetailsBtn = document.getElementById('close-details');

  // Set up observer for dynamic content
  observer = new MutationObserver((mutations) => {
    for (const mutation of mutations) {
      if (mutation.type === 'childList' && mutation.addedNodes.length > 0) {
        setMetricFillWidths();
      }
    }
  });

  // Observe the coverage dashboard for changes
  const dashboard = document.querySelector('.coverage-dashboard');
  if (dashboard) {
    observer.observe(dashboard, { childList: true, subtree: true });
  }

  // Function to show file details panel
  window.showFileDetails = (filePath?: string): void => {
    if (fileDetails) {
      fileDetails.classList.add('active');
      fileDetails.style.display = 'block';

      if (filePath) {
        // Load file details if file path is provided
        loadFileDetails(filePath);
      }
    }
  };

  // Close file details when close button is clicked
  if (closeDetailsBtn && fileDetails) {
    closeDetailsBtn.addEventListener('click', () => {
      if (fileDetails) {
        fileDetails.classList.remove('active');
        fileDetails.style.display = 'none';
      }
    });
  }

  // Add keyboard navigation for file list
  document.addEventListener('keydown', (event: KeyboardEvent) => {
    // Only handle keys when file details are visible
    if (fileDetails && fileDetails.style.display === 'block' && closeDetailsBtn) {
      if (event.key === 'Escape') {
        // Close details panel on Escape
        closeDetailsBtn.click();
      }
    }
  });

  // Handle threshold filter changes
  const thresholdFilter = document.getElementById('coverage-threshold') as HTMLSelectElement | null;
  if (thresholdFilter) {
    thresholdFilter.addEventListener('change', () => {
      // HTMX will handle the actual request
      // This is just for any additional UI updates
      updateFilterIndicator(thresholdFilter.value);
    });
  }

  // Handle the coverage filter
  const filterInput = document.getElementById('coverage-filter') as HTMLInputElement;
  if (filterInput) {
    filterInput.addEventListener('input', function () {
      const filterValue = filterInput.value.toLowerCase();

      // Filter the file list
      const fileRows = document.querySelectorAll('.file-row');
      let visibleCount = 0;

      fileRows.forEach(function (row) {
        const filePath = row.getAttribute('data-path')?.toLowerCase() || '';
        const fileName = row.getAttribute('data-name')?.toLowerCase() || '';

        
        if (filePath.includes(filterValue) || fileName.includes(filterValue)) {
          (row as HTMLElement).style.display = '';
          visibleCount++;
        } else {
          (row as HTMLElement).style.display = 'none';
        }
      });
      
      // Update filter indicator
      updateFilterIndicator(filterValue);
      
      // Update visible count
      const countEl = document.getElementById('visible-files-count');
      if (countEl) {
        countEl.textContent = visibleCount.toString();
      }
    });
  }
  
  /**
   * Load file details via AJAX
   * @param filePath - Path of the file to load details for
   */
  function loadFileDetails(filePath: string): void {
    if (!window.htmx?.ajax || !fileDetails) return;

    // Show loading state
    fileDetails.innerHTML = '<div class="loading">Loading file details...</div>';

    // Load file details via HTMX
    window.htmx.ajax('GET', `/api/coverage/files/${encodeURIComponent(filePath)}`, {
      target: '#file-details',
      swap: 'innerHTML'
    });
  }

  /**
   * Update metric display in the UI
   * @param elementId - ID of the element to update
   * @param metric - Metric object with coverage data
   */
  function updateMetricDisplay(elementId: string, metric: CoverageMetric): void {
    const element = document.getElementById(elementId);
    if (!element) return;
    
    const percentageEl = element.querySelector('.metric-percentage');
    const fillEl = element.querySelector('.metric-fill');
    const ratioEl = element.querySelector('.metric-ratio');
    
    if (percentageEl) {
      percentageEl.textContent = metric.percentage.toFixed(2) + '%';
    }
    
    if (fillEl) {
      (fillEl as HTMLElement).style.width = metric.percentage + '%';
      fillEl.className = 'metric-fill ' + getCoverageClass(metric.percentage);
    }
    
    if (ratioEl) {
      ratioEl.textContent = `${metric.covered}/${metric.total}`;
    }
  }

  /**
   * Function to update any UI indicators for active filters
   * @param filterValue - Current filter value
   */
  function updateFilterIndicator(filterValue: string): void {
    const indicator = document.getElementById('filter-indicator');
    if (indicator) {
      indicator.style.display = filterValue ? 'inline-block' : 'none';
    }
  }

  // Register for WebSocket events if available
  if (typeof (window as any).htmx !== 'undefined' && (window as any).htmx.createWebSocket) {
    document.body.addEventListener('coverage-updated', function(event: Event) {
      // Refresh coverage data when notified of updates
      const coverageEvent = event as CoverageUpdateEvent;
      if (coverageEvent.detail && coverageEvent.detail.data) {
        const data = coverageEvent.detail.data;
        
        // Update summary metrics
        updateMetricDisplay('summary-statements', data.summary.statements);
        updateMetricDisplay('summary-branches', data.summary.branches);
        updateMetricDisplay('summary-functions', data.summary.functions);
        updateMetricDisplay('summary-lines', data.summary.lines);
        
        // Update file list
        const fileListEl = document.getElementById('coverage-file-list');
        
        if (fileListEl && data.files) {
          // Clear existing file list
          fileListEl.innerHTML = '';
          
          // Add each file to the list
          data.files.forEach(file => {
            const rowEl = document.createElement('tr');
            rowEl.className = `file-row ${file.status}`;
            rowEl.setAttribute('data-path', file.path);
            rowEl.setAttribute('data-name', file.name);
            
            rowEl.innerHTML = `
              <td class="file-name">${file.name}</td>
              <td class="file-path">${file.path}</td>
              <td class="metric">${file.statements.percentage.toFixed(2)}%</td>
              <td class="metric">${file.branches.percentage.toFixed(2)}%</td>
              <td class="metric">${file.functions.percentage.toFixed(2)}%</td>
              <td class="metric">${file.lines.percentage.toFixed(2)}%</td>
            `;
            
            rowEl.addEventListener('click', () => {
              (window as any).showFileDetails(file.path);
            });
            
            fileListEl.appendChild(rowEl);
          });
          
          // Update total file count
          const countEl = document.getElementById('total-files-count');
          if (countEl) {
            countEl.textContent = data.files.length.toString();
          }
          
          // Apply any active filter
          const filterInput = document.getElementById('coverage-filter') as HTMLInputElement;
          if (filterInput && filterInput.value) {
            const event = new Event('input');
            filterInput.dispatchEvent(event);
          }
        }
      }
    });
  }
  
  /**
   * Pagination helper functions - will be fully implemented in future updates
   * This is a placeholder for the pagination feature. The associated variables and
   * functions are exported for documentation purposes and future development.
   */
  // Initialize pagination variables for future use
  const currentPage = 1; // Current active page
  
  // Update pagination UI - placeholder for future implementation
  function updatePagination(): void {
    // This function will handle updating the pagination UI
    // Implementation will be added in future updates
    console.debug(`Using pagination: page ${currentPage}`);
  }
  
  // Call initial pagination update
  updatePagination();
});

// Handle threshold filter changes
const thresholdFilter = document.getElementById('coverage-threshold') as HTMLSelectElement | null;
if (thresholdFilter) {
  thresholdFilter.addEventListener('change', () => {
    // HTMX will handle the actual request
    // This is just for any additional UI updates
    updateFilterIndicator(thresholdFilter.value);
  });
}

// Handle the coverage filter
const filterInput = document.getElementById('coverage-filter') as HTMLInputElement;
if (filterInput) {
  filterInput.addEventListener('input', function () {
    const filterValue = filterInput.value.toLowerCase();

    // Filter the file list
    const fileRows = document.querySelectorAll('.file-row');
    let visibleCount = 0;

    fileRows.forEach(function (row) {
      const filePath = row.getAttribute('data-path')?.toLowerCase() || '';
      const fileName = row.getAttribute('data-name')?.toLowerCase() || '';

      
      if (filePath.includes(filterValue) || fileName.includes(filterValue)) {
        (row as HTMLElement).style.display = '';
        visibleCount++;
      } else {
        (row as HTMLElement).style.display = 'none';
      }
    });
    
    // Update filter indicator
    updateFilterIndicator(filterValue);
    
    // Update visible count
    const countEl = document.getElementById('visible-files-count');
    if (countEl) {
      countEl.textContent = visibleCount.toString();
    }
  });
}

/**
 * Load file details via AJAX
 * @param filePath - Path of the file to load details for
 */
function loadFileDetails(filePath: string): void {
  if (!window.htmx?.ajax || !fileDetails) return;

  // Show loading state
  fileDetails.innerHTML = '<div class="loading">Loading file details...</div>';

  // Load file details via HTMX
  window.htmx.ajax('GET', `/api/coverage/files/${encodeURIComponent(filePath)}`, {
    target: '#file-details',
    swap: 'innerHTML'
  });
}

/**
 * Update metric display in the UI
 * @param elementId - ID of the element to update
 * @param metric - Metric object with coverage data
 */
function updateMetricDisplay(elementId: string, metric: CoverageMetric): void {
  const element = document.getElementById(elementId);
  if (!element) return;
  
  const percentageEl = element.querySelector('.metric-percentage');
  const fillEl = element.querySelector('.metric-fill');
  const ratioEl = element.querySelector('.metric-ratio');
  
  if (percentageEl) {
    percentageEl.textContent = metric.percentage.toFixed(2) + '%';
  }
  
  if (fillEl) {
    (fillEl as HTMLElement).style.width = metric.percentage + '%';
    fillEl.className = 'metric-fill ' + getCoverageClass(metric.percentage);
  }
  
  if (ratioEl) {
    ratioEl.textContent = `${metric.covered}/${metric.total}`;
  }
}

/**
 * Function to update any UI indicators for active filters
 * @param filterValue - Current filter value
 */
function updateFilterIndicator(filterValue: string): void {
  const indicator = document.getElementById('filter-indicator');
  if (indicator) {
    indicator.style.display = filterValue ? 'inline-block' : 'none';
  }
}

// Register for WebSocket events if available
if (typeof (window as any).htmx !== 'undefined' && (window as any).htmx.createWebSocket) {
  document.body.addEventListener('coverage-updated', function(event: Event) {
    // Refresh coverage data when notified of updates
    const coverageEvent = event as CoverageUpdateEvent;
    if (coverageEvent.detail && coverageEvent.detail.data) {
      const data = coverageEvent.detail.data;
      
      // Update summary metrics
      updateMetricDisplay('summary-statements', data.summary.statements);
      updateMetricDisplay('summary-branches', data.summary.branches);
      updateMetricDisplay('summary-functions', data.summary.functions);
      updateMetricDisplay('summary-lines', data.summary.lines);
      
      // Update file list
      const fileListEl = document.getElementById('coverage-file-list');
      
      if (fileListEl && data.files) {
        // Clear existing file list
        fileListEl.innerHTML = '';
        
        // Add each file to the list
        data.files.forEach(file => {
          const rowEl = document.createElement('tr');
          rowEl.className = `file-row ${file.status}`;
          rowEl.setAttribute('data-path', file.path);
          rowEl.setAttribute('data-name', file.name);
          
          rowEl.innerHTML = `
            <td class="file-name">${file.name}</td>
            <td class="file-path">${file.path}</td>
            <td class="metric">${file.statements.percentage.toFixed(2)}%</td>
            <td class="metric">${file.branches.percentage.toFixed(2)}%</td>
            <td class="metric">${file.functions.percentage.toFixed(2)}%</td>
            <td class="metric">${file.lines.percentage.toFixed(2)}%</td>
          `;
          
          rowEl.addEventListener('click', () => {
            (window as any).showFileDetails(file.path);
          });
          
          fileListEl.appendChild(rowEl);
        });
        
        // Update total file count
        const countEl = document.getElementById('total-files-count');
        if (countEl) {
          countEl.textContent = data.files.length.toString();
        }
        
        // Apply any active filter
        const filterInput = document.getElementById('coverage-filter') as HTMLInputElement;
        if (filterInput && filterInput.value) {
          const event = new Event('input');
          filterInput.dispatchEvent(event);
        }
      }
    }
  });
}

// Using the module-scoped goToPage function at the top of the file

// Export types and functions for testing
export type { CoverageMetric, FileCoverage, CoverageUpdateEvent };

// Export functions for testing
export { 
  loadFileDetails,
  updateMetricDisplay,
  updateFilterIndicator,
  setMetricFillWidths,
  getCoverageClass,
  goToPage
};
