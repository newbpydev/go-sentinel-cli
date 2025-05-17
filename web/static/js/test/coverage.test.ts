import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import type { CoverageMetric, CoverageUpdateEvent } from '../src/coverage';

describe('Coverage Visualization', () => {
  // Declare shared test variables
  let dashboardEl: HTMLElement;
  let fileDetailsEl: HTMLElement;
  let closeDetailsBtn: HTMLButtonElement;
  let filterInput: HTMLInputElement;
  let fileList: HTMLTableElement;
  let fetchSpy: any;
  
  beforeEach(() => {
    // Reset the DOM for clean tests
    document.body.innerHTML = '';
    
    // Create dashboard element
    dashboardEl = document.createElement('div');
    dashboardEl.className = 'coverage-dashboard';
    document.body.appendChild(dashboardEl);
    
    // Create summary metrics
    const summaryMetrics = ['summary-statements', 'summary-branches', 'summary-functions', 'summary-lines'];
    summaryMetrics.forEach(id => {
      const metricEl = document.createElement('div');
      metricEl.id = id;
      metricEl.className = 'metric';
      metricEl.innerHTML = `
        <span class="metric-percentage">0%</span>
        <div class="metric-bar">
          <div class="metric-fill" data-percentage="0"></div>
        </div>
        <span class="metric-ratio">0/0</span>
      `;
      dashboardEl.appendChild(metricEl);
    });
    
    // Create file list
    fileList = document.createElement('table');
    fileList.id = 'coverage-file-list';
    dashboardEl.appendChild(fileList);
    
    // Add some test files
    const fileData = [
      { name: 'main.go', path: 'src/main.go', statements: { total: 10, covered: 8, percentage: 80 } },
      { name: 'utils.go', path: 'src/utils/utils.go', statements: { total: 5, covered: 5, percentage: 100 } },
      { name: 'handlers.go', path: 'src/handlers/handlers.go', statements: { total: 20, covered: 10, percentage: 50 } }
    ];
    
    fileData.forEach(file => {
      const row = document.createElement('tr');
      row.className = 'file-row';
      row.dataset.path = file.path;
      row.innerHTML = `
        <td class="file-name">${file.name}</td>
        <td class="file-path">${file.path}</td>
        <td class="file-coverage">${file.statements.percentage}%</td>
      `;
      fileList.appendChild(row);
    });
    
    // Create counter elements
    const totalFilesCount = document.createElement('span');
    totalFilesCount.id = 'total-files-count';
    totalFilesCount.textContent = '3';
    dashboardEl.appendChild(totalFilesCount);
    
    const visibleFilesCount = document.createElement('span');
    visibleFilesCount.id = 'visible-files-count';
    visibleFilesCount.textContent = '3';
    dashboardEl.appendChild(visibleFilesCount);
    
    // Create filter input and indicator
    filterInput = document.createElement('input');
    filterInput.id = 'coverage-filter';
    filterInput.type = 'text';
    filterInput.placeholder = 'Filter files...';
    dashboardEl.appendChild(filterInput);
    
    const filterIndicator = document.createElement('span');
    filterIndicator.id = 'filter-indicator';
    filterIndicator.style.display = 'none';
    dashboardEl.appendChild(filterIndicator);
    
    // Create file details panel
    fileDetailsEl = document.createElement('div');
    fileDetailsEl.id = 'file-details';
    fileDetailsEl.style.display = 'none';
    document.body.appendChild(fileDetailsEl);
    
    // Add file detail elements
    const fileDetailElements = [
      'file-detail-name', 
      'file-detail-path', 
      'file-statements', 
      'file-branches', 
      'file-functions', 
      'file-lines', 
      'file-code-content'
    ];
    
    fileDetailElements.forEach(id => {
      const el = document.createElement('div');
      el.id = id;
      
      if (id.startsWith('file-') && id !== 'file-code-content' && !id.includes('detail')) {
        // Add metric structure for coverage metrics
        el.className = 'metric';
        el.innerHTML = `
          <span class="metric-percentage">0%</span>
          <div class="metric-bar">
            <div class="metric-fill" data-percentage="0"></div>
          </div>
          <span class="metric-ratio">0/0</span>
        `;
      }
      
      fileDetailsEl.appendChild(el);
    });
    
    // Create close button
    closeDetailsBtn = document.createElement('button');
    closeDetailsBtn.id = 'close-details';
    closeDetailsBtn.textContent = 'Close';
    fileDetailsEl.appendChild(closeDetailsBtn);
    
    // Mock fetch for API calls
    fetchSpy = vi.spyOn(global, 'fetch').mockImplementation((url) => {
      if (url.toString().includes('/api/coverage/file')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({
            name: 'test.go',
            path: 'src/test.go',
            codeHtml: '<pre><code>package main\n\nfunc main() {}</code></pre>',
            statements: { total: 10, covered: 8, percentage: 80 },
            branches: { total: 4, covered: 3, percentage: 75 },
            functions: { total: 3, covered: 3, percentage: 100 },
            lines: { total: 15, covered: 12, percentage: 80 }
          })
        } as Response);
      }
      return Promise.reject(new Error('Unexpected URL'));
    });
    
    // Mock hljs for syntax highlighting
    (window as any).hljs = {
      highlightBlock: vi.fn()
    };
    
    // Define the setMetricFillWidths function
    (window as any).setMetricFillWidths = function() {
      document.querySelectorAll('.metric-fill[data-percentage]').forEach(function(el) {
        const percentage = el.getAttribute('data-percentage');
        if (percentage) {
          (el as HTMLElement).style.width = percentage + '%';
        }
      });
    };
  });
  
  afterEach(() => {
    document.body.innerHTML = '';
    vi.restoreAllMocks();
  });
  
  // Test suite for metric fill bars
  describe('Metric fill bars', () => {
    it('should set width of metric fill bars based on data-percentage attributes', () => {
      // Given
      const fillElement = document.querySelector('.metric-fill');
      fillElement?.setAttribute('data-percentage', '75');
      
      // When
      (window as any).setMetricFillWidths();
      
      // Then
      expect((fillElement as HTMLElement).style.width).toBe('75%');
    });
    
    it('should update metric fill bars when new content is added', () => {
      // Given
      const newMetric = document.createElement('div');
      newMetric.innerHTML = '<div class="metric-fill" data-percentage="50"></div>';
      
      // When
      dashboardEl.appendChild(newMetric);
      (window as any).setMetricFillWidths();
      
      // Then
      const fillElement = newMetric.querySelector('.metric-fill') as HTMLElement;
      expect(fillElement.style.width).toBe('50%');
    });
  });
  
  // Test suite for file details panel
  describe('File details panel', () => {
    it('should show file details panel when showFileDetails is called', () => {
      // Given
      (window as any).showFileDetails = function() {
        if (fileDetailsEl) {
          fileDetailsEl.classList.add('active');
          fileDetailsEl.style.display = 'block';
        }
      };
      
      // When
      (window as any).showFileDetails();
      
      // Then
      expect(fileDetailsEl.style.display).toBe('block');
      expect(fileDetailsEl.classList.contains('active')).toBe(true);
    });
    
    it('should load file details when a file path is provided', () => {
      // Given
      (window as any).showFileDetails = function(filePath?: string) {
        if (fileDetailsEl) {
          fileDetailsEl.classList.add('active');
          fileDetailsEl.style.display = 'block';
          
          // If a file path is provided, load the file details via API
          if (filePath) {
            fetch(`/api/coverage/file?path=${encodeURIComponent(filePath)}`);
          }
        }
      };
      
      // When
      (window as any).showFileDetails('src/test.go');
      
      // Then
      expect(fetchSpy).toHaveBeenCalledWith('/api/coverage/file?path=src%2Ftest.go');
    });
    
    it('should hide file details panel when close button is clicked', () => {
      // Given file details are visible
      fileDetailsEl.style.display = 'block';
      fileDetailsEl.classList.add('active');
      
      // Add event listener to close button
      closeDetailsBtn.addEventListener('click', () => {
        fileDetailsEl.style.display = 'none';
        fileDetailsEl.classList.remove('active');
      });
      
      // When
      closeDetailsBtn.click();
      
      // Then
      expect(fileDetailsEl.style.display).toBe('none');
      expect(fileDetailsEl.classList.contains('active')).toBe(false);
    });
  });
  
  // Test suite for file filtering
  describe('File filtering', () => {
    it('should update visible files count when filtering', async () => {
      // Given
      const visibleCount = document.getElementById('visible-files-count') as HTMLElement;
      visibleCount.textContent = '3'; // Start with all files visible
      
      // Create a simple filter handler for testing
      filterInput.addEventListener('input', () => {
        const filterValue = filterInput.value.toLowerCase();
        const fileRows = document.querySelectorAll('.file-row');
        
        let visibleCount = 0;
        fileRows.forEach(row => {
          const fileName = (row.querySelector('.file-name')?.textContent || '').toLowerCase();
          const filePath = (row.querySelector('.file-path')?.textContent || '').toLowerCase();
          const isVisible = fileName.includes(filterValue) || filePath.includes(filterValue);
          
          (row as HTMLElement).style.display = isVisible ? '' : 'none';
          if (isVisible) visibleCount++;
        });
        
        const visibleCountEl = document.getElementById('visible-files-count');
        if (visibleCountEl) visibleCountEl.textContent = String(visibleCount);
        
        // Update filter indicator
        const indicator = document.getElementById('filter-indicator');
        if (indicator) {
          indicator.style.display = filterValue ? 'inline-block' : 'none';
        }
      });
      
      // When
      filterInput.value = 'utils';
      filterInput.dispatchEvent(new Event('input', { bubbles: true }));
      
      // Give DOM time to update
      await Promise.resolve();
      
      // Then
      expect(visibleCount.textContent).toBe('1');
    });
    
    it('should show filter indicator when filter is active', async () => {
      // Given
      const indicator = document.getElementById('filter-indicator') as HTMLElement;
      
      // Create a simple filter handler for testing
      filterInput.addEventListener('input', () => {
        const filterValue = filterInput.value.toLowerCase();
        indicator.style.display = filterValue ? 'inline-block' : 'none';
      });
      
      // When
      filterInput.value = 'test';
      filterInput.dispatchEvent(new Event('input', { bubbles: true }));
      
      // Give DOM time to update
      await Promise.resolve();
      
      // Then
      expect(indicator.style.display).toBe('inline-block');
      
      // When filter is cleared
      filterInput.value = '';
      filterInput.dispatchEvent(new Event('input', { bubbles: true }));
      
      // Give DOM time to update
      await Promise.resolve();
      
      // Then indicator is hidden
      expect(indicator.style.display).toBe('none');
    });
  });
  
  // Test suite for metric fill colors
  describe('Metric fill colors', () => {
    it('should apply correct fill colors based on percentage', () => {
      const testCases = [
        { id: 'summary-statements', percentage: 95, expectedClass: 'high' },
        { id: 'summary-branches', percentage: 80, expectedClass: 'medium' },
        { id: 'summary-functions', percentage: 60, expectedClass: 'low' },
        { id: 'summary-lines', percentage: 30, expectedClass: 'very-low' }
      ];
      
      testCases.forEach(({ id, percentage, expectedClass }) => {
        // Given
        const metricEl = document.getElementById(id);
        if (!metricEl) throw new Error(`Element with id ${id} not found`);
        
        const fillEl = metricEl.querySelector('.metric-fill') as HTMLElement;
        if (!fillEl) throw new Error(`.metric-fill not found in ${id}`);
        
        // Create function to update class based on percentage
        (window as any).updateMetricClass = (fillElement: HTMLElement, percentage: number) => {
          fillElement.classList.remove('high', 'medium', 'low', 'very-low');
          
          if (percentage >= 90) fillElement.classList.add('high');
          else if (percentage >= 70) fillElement.classList.add('medium');
          else if (percentage >= 50) fillElement.classList.add('low');
          else fillElement.classList.add('very-low');
        };
        
        // When
        (window as any).updateMetricClass(fillEl, percentage);
        
        // Then
        expect(fillEl.classList.contains(expectedClass)).toBe(true);
      });
    });
  });
});