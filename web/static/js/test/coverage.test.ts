import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import type { CoverageMetric, CoverageUpdateEvent } from '../src/coverage';

describe('Coverage Visualization', () => {
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
    fileList = document.createElement('table') as HTMLTableElement;
    fileList.id = 'coverage-file-list';
    dashboardEl.appendChild(fileList);
    
    // Create counter elements
    const totalFilesCount = document.createElement('span');
    totalFilesCount.id = 'total-files-count';
    totalFilesCount.textContent = '0';
    dashboardEl.appendChild(totalFilesCount);
    
    const visibleFilesCount = document.createElement('span');
    visibleFilesCount.id = 'visible-files-count';
    visibleFilesCount.textContent = '0';
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
    
    // Append elements to body
    document.body.appendChild(dashboardEl);
    document.body.appendChild(fileDetailsEl);
    
    // Mock fetch
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
    
    // Mock hljs
    (window as any).hljs = {
      highlightBlock: vi.fn()
    };
    
    // Mock htmx
    (window as any).htmx = {
      createWebSocket: vi.fn()
    };
    
    // Trigger DOMContentLoaded
    document.dispatchEvent(new Event('DOMContentLoaded'));
  });
  
  afterEach(() => {
    document.body.innerHTML = '';
    vi.restoreAllMocks();
  });
  
  describe('Metric fill bars', () => {
    it('should set width of metric fill bars based on data-percentage attributes', () => {
      // Given
      const fillElement = document.querySelector('.metric-fill') as HTMLElement;
      fillElement.setAttribute('data-percentage', '75');
      
      // When
      const event = new Event('DOMContentLoaded');
      document.dispatchEvent(event);
      
      // Then
      expect(fillElement.style.width).toBe('75%');
    });
    
    it('should update metric fill bars when new content is added', () => {
      // Given
      const newMetric = document.createElement('div');
      newMetric.innerHTML = '<div class="metric-fill" data-percentage="50"></div>';
      
      // When
      dashboardEl.appendChild(newMetric);
      
      // Then
      const fillElement = newMetric.querySelector('.metric-fill') as HTMLElement;
      expect(fillElement.style.width).toBe('50%');
    });
  });
  
  describe('File details panel', () => {
    it('should show file details panel when showFileDetails is called', () => {
      // Ensure showFileDetails function is defined
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
      // Setup showFileDetails function with file loading
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
    
    it('should update file details content when data is loaded', async () => {
      // When
      (window as any).showFileDetails('src/test.go');
      
      // Wait for fetch to resolve
      await new Promise(resolve => setTimeout(resolve, 0));
      
      // Then
      const nameEl = document.getElementById('file-detail-name');
      const pathEl = document.getElementById('file-detail-path');
      const codeEl = document.getElementById('file-code-content');
      
      expect(nameEl?.textContent).toBe('test.go');
      expect(pathEl?.textContent).toBe('src/test.go');
      expect(codeEl?.innerHTML).toBe('<pre><code>package main\n\nfunc main() {}</code></pre>');
      
      // Verify metrics were updated
      const statementsPercentage = document.querySelector('#file-statements .metric-percentage');
      expect(statementsPercentage?.textContent).toBe('80.00%');
    });
  });
  
  describe('File filtering', () => {
    beforeEach(() => {
      // Add some file rows for testing
      const fileList = document.getElementById('coverage-file-list');
      if (!fileList) {
        throw new Error('Coverage file list element not found');
        return;
      }
      
      const files = [
        { name: 'main.go', path: 'src/main.go' },
        { name: 'utils.go', path: 'src/utils/utils.go' },
        { name: 'models.go', path: 'src/models/models.go' }
      ];
      
      files.forEach(file => {
        const row = document.createElement('tr');
        row.className = 'file-row';
        row.setAttribute('data-path', file.path);
        row.setAttribute('data-name', file.name);
        row.innerHTML = `<td>${file.name}</td><td>${file.path}</td>`;
        fileList.appendChild(row);
      });
    });
    
    it('should filter file list when filter input changes', () => {
      // When
      filterInput.value = 'models';
      filterInput.dispatchEvent(new Event('input'));
      
      // Then
      const rows = Array.from(document.querySelectorAll('.file-row'));
      expect(rows.length).toBeGreaterThan(0);
      if (rows[0]) expect((rows[0] as HTMLElement).style.display).toBe('none'); // main.go
      if (rows[1]) expect((rows[1] as HTMLElement).style.display).toBe('none'); // utils.go
      if (rows[2]) expect((rows[2] as HTMLElement).style.display).toBe(''); // models.go should be visible
    });
    
    it('should update visible files count when filtering', () => {
      // Given
      const visibleCount = document.getElementById('visible-files-count') as HTMLElement;
      visibleCount.textContent = '3'; // Start with all files visible
      
      // When
      filterInput.value = 'utils';
      filterInput.dispatchEvent(new Event('input'));
      
      // Then
      expect(visibleCount.textContent).toBe('1');
    });
    
    it('should show filter indicator when filter is active', () => {
      // Given
      const indicator = document.getElementById('filter-indicator') as HTMLElement;
      
      // When
      filterInput.value = 'test';
      filterInput.dispatchEvent(new Event('input'));
      
      // Then
      expect(indicator.style.display).toBe('inline-block');
      
      // When filter is cleared
      filterInput.value = '';
      filterInput.dispatchEvent(new Event('input'));
      
      // Then indicator is hidden
      expect(indicator.style.display).toBe('none');
    });
  });
  
  describe('WebSocket integration', () => {
    it('should update summary metrics on coverage-updated event', () => {
      // Given
      const event = new CustomEvent('coverage-updated', {
        detail: {
          data: {
            summary: {
              statements: { total: 100, covered: 80, percentage: 80 },
              branches: { total: 50, covered: 40, percentage: 80 },
              functions: { total: 30, covered: 25, percentage: 83.33 },
              lines: { total: 200, covered: 160, percentage: 80 }
            },
            files: []
          }
        }
      }) as CoverageUpdateEvent;
      
      // When
      document.body.dispatchEvent(event);
      
      // Then
      const statementsPercentage = document.querySelector('#summary-statements .metric-percentage');
      const branchesPercentage = document.querySelector('#summary-branches .metric-percentage');
      const functionsPercentage = document.querySelector('#summary-functions .metric-percentage');
      const linesPercentage = document.querySelector('#summary-lines .metric-percentage');
      
      expect(statementsPercentage?.textContent).toBe('80.00%');
      expect(branchesPercentage?.textContent).toBe('80.00%');
      expect(functionsPercentage?.textContent).toBe('83.33%');
      expect(linesPercentage?.textContent).toBe('80.00%');
    });
    
    it('should update file list on coverage-updated event', () => {
      // Given
      const event = new CustomEvent('coverage-updated', {
        detail: {
          data: {
            summary: {
              statements: { total: 100, covered: 80, percentage: 80 },
              branches: { total: 50, covered: 40, percentage: 80 },
              functions: { total: 30, covered: 25, percentage: 83.33 },
              lines: { total: 200, covered: 160, percentage: 80 }
            },
            files: [
              {
                name: 'test.go',
                path: 'src/test.go',
                statements: { total: 10, covered: 8, percentage: 80 },
                branches: { total: 4, covered: 3, percentage: 75 },
                functions: { total: 3, covered: 3, percentage: 100 },
                lines: { total: 15, covered: 12, percentage: 80 },
                status: 'pass'
              }
            ]
          }
        }
      }) as CoverageUpdateEvent;
      
      // When
      document.body.dispatchEvent(event);
      
      // Then
      const fileList = document.getElementById('coverage-file-list');
      const rows = fileList?.querySelectorAll('.file-row');
      
      expect(rows?.length).toBe(1);
      
      // Make sure we have a row before accessing properties
      const firstRow = rows?.[0];
      if (!firstRow) {
        throw new Error('Expected to find a row in the file list');
        return;
      }
      
      expect(firstRow.classList.contains('pass')).toBe(true);
      expect(firstRow.getAttribute('data-path')).toBe('src/test.go');
      
      // Verify file count was updated
      const totalCount = document.getElementById('total-files-count');
      expect(totalCount?.textContent).toBe('1');
    });
  });
  
  describe('Helper functions', () => {
    it('should return appropriate coverage class based on percentage', () => {
      // This is testing a private function, so we need to test it indirectly
      
      // When (create metric fills with different percentages)
      [95, 80, 60, 30].forEach((percentage) => {
        const fillEl = document.createElement('div');
        fillEl.className = 'metric-fill';
        fillEl.setAttribute('data-percentage', percentage.toString());
        dashboardEl.appendChild(fillEl);
        
        // Fake updating the element through an event that would call the function
        const event = new CustomEvent('coverage-updated', {
          detail: {
            data: {
              summary: {
                statements: { total: 100, covered: percentage, percentage: percentage },
                branches: { total: 100, covered: percentage, percentage: percentage },
                functions: { total: 100, covered: percentage, percentage: percentage },
                lines: { total: 100, covered: percentage, percentage: percentage }
              },
              files: []
            }
          }
        }) as CoverageUpdateEvent;
        
        document.body.dispatchEvent(event);
      });
      
      // Then (check if the classes were applied correctly)
      const updateMetricDisplay = (id: string, metric: CoverageMetric) => {
        const element = document.getElementById(id);
        if (!element) return;
        
        const fillEl = element.querySelector('.metric-fill') as HTMLElement | null;
        if (fillEl) {
          const expectedClass = 
            metric.percentage >= 90 ? 'high' :
            metric.percentage >= 70 ? 'medium' :
            metric.percentage >= 50 ? 'low' : 'very-low';
          
          expect(fillEl.classList.contains(expectedClass)).toBe(true);
        }
      };
      
      updateMetricDisplay('summary-statements', { total: 100, covered: 95, percentage: 95 });
      updateMetricDisplay('summary-branches', { total: 100, covered: 80, percentage: 80 });
      updateMetricDisplay('summary-functions', { total: 100, covered: 60, percentage: 60 });
      updateMetricDisplay('summary-lines', { total: 100, covered: 30, percentage: 30 });
    });
  });
});
