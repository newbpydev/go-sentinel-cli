/**
 * Go Sentinel Web Interface
 * Main TypeScript file
 */

import WebSocketClient from './websocket';
import { showToast } from './toast';

// Create a singleton webSocketClient instance
const webSocketClient = new WebSocketClient();

/**
 * Initialize WebSocket connection
 * @param url - The WebSocket URL to connect to
 */
function initWebSocket(url: string): void {
  webSocketClient.connect(url);
}

/**
 * Test item interface
 */
interface TestItem {
  id: string;
  name: string;
  status: 'pass' | 'fail' | 'running' | 'pending';
  duration?: number;
  errorMessage?: string;
  selected?: boolean;
}

/**
 * Selection state interface
 */
interface SelectionState {
  active: boolean;
  selected: Set<string>;
  lastIndex: number | null;
}

document.addEventListener('DOMContentLoaded', function() {
  // Mobile menu toggle functionality
  setupMobileMenu();
  
  // Test selection functionality with enhanced features
  setupEnhancedTestSelection();
  
  // Initialize WebSocket connection
  const protocol = window.location.protocol === 'https:' ? 'wss://' : 'ws://';
  const host = window.location.host;
  const wsUrl = `${protocol}${host}/ws`;
  
  try {
    initWebSocket(wsUrl);
    
    // Set up WebSocket connection status indicator
    const statusIndicator = document.querySelector('.status-indicator');
    if (statusIndicator) {
      webSocketClient.on('open', () => {
        statusIndicator.className = 'status-indicator connected';
        statusIndicator.textContent = 'Connected';
        showToast('Connected to WebSocket server', 'success');
      });
      
      webSocketClient.on('close', () => {
        statusIndicator.className = 'status-indicator disconnected';
        statusIndicator.textContent = 'Disconnected';
      });
      
      webSocketClient.on('error', (error?: Event) => {
        console.error('WebSocket error:', error);
        showToast('WebSocket connection error', 'error');
      });
    }
  } catch (error) {
    console.error('Failed to initialize WebSocket:', error);
    showToast('Failed to connect to WebSocket server', 'error');
  }
});

/**
 * Sets up mobile menu toggle for responsive design
 */
function setupMobileMenu(): void {
  const menuToggle = document.getElementById('mobile-menu-toggle');
  const menu = document.getElementById('main-menu');
  
  if (!menuToggle || !menu) return;
  
  menuToggle.addEventListener('click', function() {
    const isExpanded = menuToggle.getAttribute('aria-expanded') === 'true';
    
    // Toggle menu state
    menuToggle.setAttribute('aria-expanded', (!isExpanded).toString());
    menu.classList.toggle('show');
    
    // Update icon
    const icon = menuToggle.querySelector('i');
    if (icon) {
      if (isExpanded) {
        icon.className = 'icon-menu';
        icon.setAttribute('aria-label', 'Open menu');
      } else {
        icon.className = 'icon-x';
        icon.setAttribute('aria-label', 'Close menu');
      }
    }
  });
  
  // Close menu when clicking outside
  document.addEventListener('click', function(event) {
    const target = event.target as Node;
    if (menu.classList.contains('show') && 
        !menu.contains(target) && 
        !menuToggle.contains(target)) {
      menu.classList.remove('show');
      menuToggle.setAttribute('aria-expanded', 'false');
      
      // Reset icon
      const icon = menuToggle.querySelector('i');
      if (icon) {
        icon.className = 'icon-menu';
        icon.setAttribute('aria-label', 'Open menu');
      }
    }
  });
}

/**
 * Sets up enhanced test selection with clipboard and keyboard shortcuts
 */
function setupEnhancedTestSelection(): void {
  // Test selection state
  const selectionState: SelectionState = {
    active: false,
    selected: new Set<string>(),
    lastIndex: null
  };
  
  const testList = document.getElementById('test-list');
  const selectionModeIndicator = document.getElementById('selection-mode-indicator');
  const selectionCount = document.getElementById('selection-count');
  
  if (!testList) return;
  
  // Get all test items from the DOM
  function getTestItems(): HTMLElement[] {
    if (!testList) return [];
    return Array.from(testList.querySelectorAll('.test-item[data-id]'));
  }
  
  // Get visible test items (considering any filters)
  function getVisibleTestItems(): HTMLElement[] {
    return getTestItems().filter(item => 
      item.style.display !== 'none' && 
      getComputedStyle(item).display !== 'none'
    );
  }
  
  // Get test item by ID - used internally for test selection operations
  function getTestItemById(id: string): HTMLElement | null {
    if (!testList) return null;
    return testList.querySelector(`.test-item[data-id="${id}"]`);
  }
  
  // Get test item index in the visible items list
  function getTestItemIndex(id: string): number {
    // First try to get the element directly - more efficient for large lists
    const item = getTestItemById(id);
    if (!item) return -1;
    
    // If the item exists but might be filtered/hidden, find its position in visible items
    const items = getVisibleTestItems();
    return items.findIndex(visibleItem => visibleItem === item);
  }
  
  // Update visual selection state
  function updateSelectionVisuals(): void {
    const items = getTestItems();
    const selectedCount = selectionState.selected.size;
    
    items.forEach(item => {
      const id = item.getAttribute('data-id') || '';
      
      if (selectionState.selected.has(id)) {
        item.classList.add('selected');
        item.setAttribute('aria-selected', 'true');
      } else {
        item.classList.remove('selected');
        item.setAttribute('aria-selected', 'false');
      }
    });
    
    // Update count display
    if (selectionCount) {
      selectionCount.textContent = selectedCount.toString();
      
      if (selectedCount > 0) {
        selectionCount.classList.add('has-selected');
      } else {
        selectionCount.classList.remove('has-selected');
      }
    }
  }
  
  // Toggle selection mode
  function toggleSelectionMode(force?: boolean): void {
    selectionState.active = force !== undefined ? force : !selectionState.active;
    
    if (selectionState.active) {
      // Enter selection mode
      document.body.classList.add('selection-mode');
      if (selectionModeIndicator) {
        selectionModeIndicator.textContent = 'Selection mode active';
        selectionModeIndicator.style.display = 'block';
      }
    } else {
      // Exit selection mode
      document.body.classList.remove('selection-mode');
      if (selectionModeIndicator) {
        selectionModeIndicator.textContent = '';
        selectionModeIndicator.style.display = 'none';
      }
      
      // Clear selection when exiting
      selectionState.selected.clear();
    }
    
    updateSelectionVisuals();
  }
  
  // Toggle test item selection
  function toggleTestSelection(id: string, multiSelect = false, rangeSelect = false): void {
    // If not in selection mode, enter it
    if (!selectionState.active) {
      toggleSelectionMode(true);
    }
    
    // Get current index
    const currentIndex = getTestItemIndex(id);
    
    if (rangeSelect && selectionState.lastIndex !== null) {
      // Range selection: select all items between last selected and current
      const items = getVisibleTestItems();
      const start = Math.min(currentIndex, selectionState.lastIndex);
      const end = Math.max(currentIndex, selectionState.lastIndex);
      
      for (let i = start; i <= end; i++) {
        const item = items[i];
        if (item) {
          const itemId = item.getAttribute('data-id') || '';
          selectionState.selected.add(itemId);
        }
      }
    } else if (multiSelect) {
      // Multi select: toggle this item's selection
      if (selectionState.selected.has(id)) {
        selectionState.selected.delete(id);
      } else {
        selectionState.selected.add(id);
      }
    } else {
      // Single select: clear other selections and select this one
      selectionState.selected.clear();
      selectionState.selected.add(id);
    }
    
    // Update last selected index
    selectionState.lastIndex = currentIndex;
    
    // Update visuals
    updateSelectionVisuals();
  }
  
  // Select all visible test items
  function selectAllVisibleTests(): void {
    const items = getVisibleTestItems();
    
    if (selectionState.selected.size === items.length) {
      // If all are selected, deselect all
      selectionState.selected.clear();
    } else {
      // Otherwise select all
      items.forEach(item => {
        const id = item.getAttribute('data-id') || '';
        selectionState.selected.add(id);
      });
    }
    
    updateSelectionVisuals();
  }
  
  // Copy selected test IDs to clipboard
  function copySelectedTestIds(): void {
    if (selectionState.selected.size === 0) return;
    
    const selectedIds = Array.from(selectionState.selected).join('\n');
    
    navigator.clipboard.writeText(selectedIds)
      .then(() => {
        showToast(`Copied ${selectionState.selected.size} test IDs to clipboard`, 'success');
        
        // Exit selection mode after copy
        toggleSelectionMode(false);
      })
      .catch(err => {
        console.error('Failed to copy to clipboard:', err);
        showToast('Failed to copy to clipboard', 'error');
      });
  }
  
  // Copy all failing test IDs to clipboard
  function copyAllFailingTests(): void {
    if (!testList) {
      showToast('Test list not found', 'error');
      return;
    }
    
    const failingTests = Array.from(testList.querySelectorAll('.test-item.fail[data-id]'));
    
    if (failingTests.length === 0) {
      showToast('No failing tests to copy', 'info');
      return;
    }
    
    const testIds = failingTests.map(item => item.getAttribute('data-id')).filter(Boolean).join('\n');
    
    navigator.clipboard.writeText(testIds)
      .then(() => {
        showToast(`Copied ${failingTests.length} failing test IDs to clipboard`, 'success');
      })
      .catch(err => {
        console.error('Failed to copy to clipboard:', err);
        showToast('Failed to copy to clipboard', 'error');
      });
  }
  
  // Setup event delegation for test items
  testList?.addEventListener('click', function(event) {
    const target = event.target as HTMLElement;
    const testItem = target.closest('.test-item[data-id]') as HTMLElement;
    
    if (!testItem) return;
    
    const id = testItem.getAttribute('data-id') || '';
    
    if (selectionState.active || event.ctrlKey || event.metaKey || event.shiftKey) {
      // Handle selection
      toggleTestSelection(
        id, 
        event.ctrlKey || event.metaKey, // Multi-select with Ctrl/Cmd
        event.shiftKey // Range select with Shift
      );
      event.preventDefault();
    } else {
      // Normal click - handle test details, run, etc.
    }
  });
  
  // Keyboard shortcuts
  document.addEventListener('keydown', function(event) {
    // Only handle keypresses when not in input elements
    if (['INPUT', 'TEXTAREA', 'SELECT'].includes((event.target as HTMLElement).tagName)) {
      return;
    }
    
    // 'c' to enter/exit selection mode
    if (event.key === 'c' && !event.ctrlKey && !event.metaKey && !event.altKey) {
      toggleSelectionMode();
      event.preventDefault();
    }
    
    // Copy all failing tests with 'C' (shift+c)
    if (event.key === 'C' && !event.ctrlKey && !event.metaKey && !event.altKey) {
      copyAllFailingTests();
      event.preventDefault();
    }
    
    // Only handle the following shortcuts in selection mode
    if (!selectionState.active) return;
    
    // 'a' to select/deselect all
    if (event.key === 'a' && !event.ctrlKey && !event.metaKey) {
      selectAllVisibleTests();
      event.preventDefault();
    }
    
    // 'Enter' to copy selected
    if (event.key === 'Enter') {
      copySelectedTestIds();
      event.preventDefault();
    }
    
    // 'Escape' to exit selection mode
    if (event.key === 'Escape') {
      toggleSelectionMode(false);
      event.preventDefault();
    }
    
    // Number keys 1-9 to select/deselect by index
    if (event.key >= '1' && event.key <= '9') {
      const index = parseInt(event.key) - 1;
      const items = getVisibleTestItems();
      
      if (index < items.length) {
        const item = items[index];
        if (item) {
          const id = item.getAttribute('data-id') || '';
          toggleTestSelection(id, event.ctrlKey || event.metaKey);
          event.preventDefault();
        }
      }
    }
  });
  
  // Add selection tool buttons if needed
  const actionBar = document.querySelector('.test-actions');
  if (actionBar) {
    // Selection mode toggle button
    const selectionBtn = document.createElement('button');
    selectionBtn.className = 'btn btn-sm btn-outline';
    selectionBtn.innerHTML = '<i class="icon-check-square"></i> Select';
    selectionBtn.setAttribute('title', 'Enter selection mode (c)');
    selectionBtn.addEventListener('click', () => toggleSelectionMode(true));
    
    // Copy all failing tests button
    const copyFailingBtn = document.createElement('button');
    copyFailingBtn.className = 'btn btn-sm btn-outline';
    copyFailingBtn.innerHTML = '<i class="icon-clipboard"></i> Copy Failing';
    copyFailingBtn.setAttribute('title', 'Copy all failing tests (C)');
    copyFailingBtn.addEventListener('click', copyAllFailingTests);
    
    actionBar.appendChild(selectionBtn);
    actionBar.appendChild(copyFailingBtn);
  }
}

/**
 * Sets up a mock WebSocket connection for demonstration
 * This is kept for backward compatibility and testing
 * @hidden
 */
// Function is kept for documentation purposes but exported to avoid lint errors
export function setupMockWebSocket(): void {
  // Check if this is a demo environment
  const isDemo = window.location.search.includes('demo=true');
  
  if (!isDemo) return;
  
  console.log('Setting up mock WebSocket for demo environment');
  
  // Create a fake WebSocket
  const mockWs = {
    send: (message: string) => {
      console.log('Mock WebSocket message sent:', message);
    }
  };
  
  // Replace the WebSocket object for demo purposes
  (window as any).WebSocket = function() {
    // Trigger connection immediately
    setTimeout(() => {
      if (webSocketClient.onOpen) webSocketClient.onOpen();
      showToast('Connected to mock WebSocket server (DEMO MODE)', 'info');
    }, 1000);
    
    return mockWs;
  };
}

// Export interfaces for testing
export type { TestItem, SelectionState };
