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

// Set width of metric fill bars based on data-percentage attributes
function setMetricFillWidths(): void {
  document.querySelectorAll('.metric-fill[data-percentage]').forEach(function(el) {
    const percentage = el.getAttribute('data-percentage');
    if (percentage) {
      (el as HTMLElement).style.width = percentage + '%';
    }
  });
}

// Expose for testing and runtime as early as possible
// Expose for testing and runtime in both browser and Vitest/JSDOM environments
(window as any).setMetricFillWidths = setMetricFillWidths;
(globalThis as any).setMetricFillWidths = setMetricFillWidths;

document.addEventListener('DOMContentLoaded', function() {
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
  const fileDetails = document.getElementById('file-details') as HTMLElement;
  const closeDetailsBtn = document.getElementById('close-details');
  
  // Function to show file details panel
  (window as any).showFileDetails = function(filePath?: string): void {
    if (fileDetails) {
      fileDetails.classList.add('active');
      fileDetails.style.display = 'block';
      
      // If a file path is provided, load the file details
      if (filePath) {
        loadFileDetails(filePath);
      }
    }
  };
  
  // Close file details when close button is clicked
  if (closeDetailsBtn) {
    closeDetailsBtn.addEventListener('click', function() {
      if (fileDetails) {
        fileDetails.classList.remove('active');
        fileDetails.style.display = 'none';
      }
    });
  }
  
  // Handle the coverage filter
  const filterInput = document.getElementById('coverage-filter') as HTMLInputElement;
  if (filterInput) {
    filterInput.addEventListener('input', function() {
      const filterValue = filterInput.value.toLowerCase();
      
      // Filter the file list
      const fileRows = document.querySelectorAll('.file-row');
      let visibleCount = 0;
      
      fileRows.forEach(function(row) {
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
   * Function to update any UI indicators for active filters
   * @param filterValue - Current filter value
   */
  function updateFilterIndicator(filterValue: string): void {
    const indicator = document.getElementById('filter-indicator');
    if (indicator) {
      indicator.style.display = filterValue ? 'inline-block' : 'none';
    }
  }
  
  /**
   * Load file details via AJAX
   * @param filePath - Path of the file to load details for
   */
  function loadFileDetails(filePath: string): void {
    fetch(`/api/coverage/file?path=${encodeURIComponent(filePath)}`)
      .then(response => response.json())
      .then(data => {
        // Update file details panel
        const fileNameEl = document.getElementById('file-detail-name');
        const filePathEl = document.getElementById('file-detail-path');
        const codeContentEl = document.getElementById('file-code-content');
        
        if (fileNameEl) {
          fileNameEl.textContent = data.name;
        }
        
        if (filePathEl) {
          filePathEl.textContent = data.path;
        }
        
        if (codeContentEl) {
          codeContentEl.innerHTML = data.codeHtml;
          highlightCode();
        }
        
        // Update metrics
        updateMetricDisplay('file-statements', data.statements);
        updateMetricDisplay('file-branches', data.branches);
        updateMetricDisplay('file-functions', data.functions);
        updateMetricDisplay('file-lines', data.lines);
      })
      .catch(error => {
        console.error('Error loading file details:', error);
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
   * Add syntax highlighting to code blocks (if using a library like highlight.js)
   */
  function highlightCode(): void {
    // Type safety for hljs - this is a global library loaded via CDN
    const highlightJs = (window as any).hljs;
    if (typeof highlightJs !== 'undefined') {
      document.querySelectorAll('pre code').forEach((block) => {
        highlightJs.highlightBlock(block);
      });
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
   * Helper function to format coverage percentage with appropriate color class
   * @param percentage - Coverage percentage
   * @returns CSS class name based on coverage percentage
   */
  function getCoverageClass(percentage: number): string {
    if (percentage >= 90) return 'high';
    if (percentage >= 70) return 'medium';
    if (percentage >= 50) return 'low';
    return 'very-low';
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
  
  // Call initial pagination update on load
  // This ensures the function is used and establishes the pattern for future development
  updatePagination();
});

/**
 * Navigate to a specific coverage page via URL
 * @param page The page number to go to
 */
export function goToPage(page: number): void {
  const url = `/coverage?page=${page}`;
  window.location.href = url;
}

// Export types for testing
export type { CoverageMetric, FileCoverage, CoverageUpdateEvent };
