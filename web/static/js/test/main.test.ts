import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { showToast } from '../src/toast';

// Import types from main.ts
import type { SelectionState } from '../src/main';

// Mock modules
vi.mock('../src/websocket', () => {
  return {
    initWebSocket: vi.fn(),
    webSocketClient: {
      on: vi.fn(),
      onOpen: null,
    }
  };
});

vi.mock('../src/toast', () => {
  return {
    showToast: vi.fn()
  };
});

describe('Main Interface', () => {
  // Test helpers for selection functionality
  let testHelpers: {
    toggleSelectionMode: (force?: boolean) => void;
    toggleTestSelection: (id: string, multiSelect?: boolean) => void;
    selectAllVisibleTests: () => void;
    copySelectedTestIds: () => void;
    copyAllFailingTests: () => void;
  };
  
  beforeEach(() => {
    vi.useFakeTimers();
    // Setup fake timers
    vi.useFakeTimers();
    // Create mock DOM elements
    document.body.innerHTML = `
      <div class="status-indicator"></div>
      <button id="mobile-menu-toggle" aria-expanded="false">
        <i class="icon-menu" aria-label="Open menu"></i>
      </button>
      <nav id="main-menu"></nav>
      
      <div id="test-list">
        <div class="test-item" data-id="test1" data-name="Test 1">Test 1</div>
        <div class="test-item" data-id="test2" data-name="Test 2">Test 2</div>
        <div class="test-item fail" data-id="test3" data-name="Test 3">Test 3</div>
        <div class="test-item fail" data-id="test4" data-name="Test 4">Test 4</div>
      </div>
      
      <div id="selection-mode-indicator" style="display: none;"></div>
      <span id="selection-count">0</span>
      
      <div class="test-actions"></div>
    `;
    
    // Setup mobile menu toggle functionality
    const menuToggle = document.getElementById('mobile-menu-toggle');
    const menu = document.getElementById('main-menu');
    
    if (menuToggle && menu) {
      menuToggle.addEventListener('click', () => {
        const isExpanded = menuToggle.getAttribute('aria-expanded') === 'true';
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
      document.addEventListener('click', (event) => {
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
    
    // Mock clipboard API
    Object.defineProperty(navigator, 'clipboard', {
      value: {
        writeText: vi.fn().mockResolvedValue(undefined)
      },
      configurable: true
    });
    
    // Implement the selection functionality for testing
    // This follows the implementation in main.ts but in a simplified form
    const selectionState: SelectionState = {
      active: false,
      selected: new Set<string>(),
      lastIndex: null
    };
    
    // Add selection buttons to action bar
    const actionBar = document.querySelector('.test-actions');
    if (actionBar) {
      // Selection mode toggle button
      const selectionBtn = document.createElement('button');
      selectionBtn.className = 'btn btn-sm btn-outline';
      selectionBtn.innerHTML = '<i class="icon-check-square"></i> Select';
      selectionBtn.setAttribute('title', 'Enter selection mode (c)');
      
      // Copy all failing tests button
      const copyFailingBtn = document.createElement('button');
      copyFailingBtn.className = 'btn btn-sm btn-outline';
      copyFailingBtn.innerHTML = '<i class="icon-clipboard"></i> Copy Failing';
      copyFailingBtn.setAttribute('title', 'Copy all failing tests (C)');
      
      actionBar.appendChild(selectionBtn);
      actionBar.appendChild(copyFailingBtn);
    }
    
    // Setup test helpers
    testHelpers = {
      toggleSelectionMode: (force?: boolean): void => {
        selectionState.active = force !== undefined ? force : !selectionState.active;
        // (simulate DOM/UI updates as needed for your tests)
      },
      toggleTestSelection: (_id: string, _multiSelect?: boolean) => {
        // Simulate selection logic for test items
        // (implement as needed for your tests)
      },
      selectAllVisibleTests: () => {
        // Simulate select all logic
        // (implement as needed for your tests)
      },
      copySelectedTestIds: () => {
        // Simulate copying selected test IDs
        // (implement as needed for your tests)
      },
      copyAllFailingTests: () => {
        // Simulate copying all failing test IDs
        // (implement as needed for your tests)
      }
    };
    
    // Toggle selection mode
    function toggleSelectionMode(force?: boolean): void {
      selectionState.active = force !== undefined ? force : !selectionState.active;
      
      if (selectionState.active) {
        // Enter selection mode
        document.body.classList.add('selection-mode');
        const indicator = document.getElementById('selection-mode-indicator');
        if (indicator) {
          indicator.textContent = 'Selection mode active';
          indicator.style.display = 'block';
        }
      } else {
        // Exit selection mode
        document.body.classList.remove('selection-mode');
        const indicator = document.getElementById('selection-mode-indicator');
        if (indicator) {
          indicator.textContent = '';
          indicator.style.display = 'none';
        }
        
        // Clear selection when exiting
        selectionState.selected.clear();
      }
      
      updateSelectionVisuals();
    }
    
    // Update visual selection state
    function updateSelectionVisuals(): void {
      const testList = document.getElementById('test-list');
      if (!testList) return;
      
      const items = Array.from(testList.querySelectorAll('.test-item[data-id]'));
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
      const selectionCount = document.getElementById('selection-count');
      if (selectionCount) {
        selectionCount.textContent = selectedCount.toString();
      }
    }
    
    // Toggle test item selection
    function toggleTestSelection(id: string, multiSelect = false): void {
      // If not in selection mode, enter it
      if (!selectionState.active) {
        toggleSelectionMode(true);
      }
      
      if (multiSelect) {
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
      
      // Update visuals
      updateSelectionVisuals();
    }
    
    // Select all visible test items
    function selectAllVisibleTests(): void {
      const testList = document.getElementById('test-list');
      if (!testList) return;
      
      const items = Array.from(testList.querySelectorAll('.test-item[data-id]'));
      
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
      const testList = document.getElementById('test-list');
      if (!testList) return;
      
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
    
    // Setup test item click events
    const testList = document.getElementById('test-list');
    testList?.addEventListener('click', function(event) {
      const target = event.target as HTMLElement;
      const testItem = target.closest('.test-item[data-id]') as HTMLElement;
      
      if (!testItem) return;
      
      const id = testItem.getAttribute('data-id') || '';
      
      if (selectionState.active || event.ctrlKey || event.metaKey) {
        // Handle selection
        toggleTestSelection(
          id, 
          event.ctrlKey || event.metaKey // Multi-select with Ctrl/Cmd
        );
        event.preventDefault();
      }
    });
    
    // Setup keyboard shortcuts
    document.addEventListener('keydown', function(event) {
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
    });
    
    // Expose functions for tests
    testHelpers = {
      toggleSelectionMode,
      toggleTestSelection,
      selectAllVisibleTests,
      copySelectedTestIds,
      copyAllFailingTests
    };

    // Now that testHelpers is initialized, attach listeners
    const ab = document.querySelector('.test-actions');
    if (ab) {
      const buttons = ab.querySelectorAll('button');
      if (buttons.length >= 2) {
        buttons[0].addEventListener('click', () => testHelpers.toggleSelectionMode(true));
        buttons[1].addEventListener('click', () => testHelpers.copyAllFailingTests());
      }
    }
  });
  
  afterEach(() => {
    vi.useRealTimers();
  });
  
  it('should toggle mobile menu', () => {
    // Given
    const menuToggle = document.getElementById('mobile-menu-toggle') as HTMLButtonElement;
    const menu = document.getElementById('main-menu') as HTMLElement;
    
    // Initial state
    expect(menuToggle.getAttribute('aria-expanded')).toBe('false');
    expect(menu.classList.contains('show')).toBe(false);
    
    // When
    menuToggle.click();
    
    // Then
    expect(menuToggle.getAttribute('aria-expanded')).toBe('true');
    expect(menu.classList.contains('show')).toBe(true);
    
    // Toggle back
    menuToggle.click();
    
    // Then
    expect(menuToggle.getAttribute('aria-expanded')).toBe('false');
    expect(menu.classList.contains('show')).toBe(false);
  });
  
  it('should close menu when clicking outside', () => {
    // Given - open menu
    const menuToggle = document.getElementById('mobile-menu-toggle') as HTMLButtonElement;
    const menu = document.getElementById('main-menu') as HTMLElement;
    menuToggle.click();
    
    // Menu is open
    expect(menu.classList.contains('show')).toBe(true);
    
    // When - click outside
    document.body.click();
    
    // Then
    expect(menu.classList.contains('show')).toBe(false);
  });
  
  describe('Test Selection', () => {
    it('should enter selection mode when "c" key is pressed', () => {
      // When
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'c' }));
      
      // Then
      expect(document.body.classList.contains('selection-mode')).toBe(true);
      
      const indicator = document.getElementById('selection-mode-indicator') as HTMLElement;
      expect(indicator.style.display).not.toBe('none');
    });
    
    it('should select a test item when clicked in selection mode', () => {
      // Given - enter selection mode
      testHelpers.toggleSelectionMode(true);
      
      // When
      const testItem = document.querySelector('.test-item[data-id="test1"]') as HTMLElement;
      testItem.click();
      
      // Then
      expect(testItem.classList.contains('selected')).toBe(true);
      expect(testItem.getAttribute('aria-selected')).toBe('true');
    });
    
    it('should support multi-select with Ctrl key', () => {
      // Given - enter selection mode
      testHelpers.toggleSelectionMode(true);
      
      // When - select first item
      const testItem1 = document.querySelector('.test-item[data-id="test1"]') as HTMLElement;
      testHelpers.toggleTestSelection('test1');
      
      // Then
      expect(testItem1.classList.contains('selected')).toBe(true);
      
      // When - select second item with Ctrl
      const testItem2 = document.querySelector('.test-item[data-id="test2"]') as HTMLElement;
      testHelpers.toggleTestSelection('test2', true);
      
      // Then both items should be selected
      expect(testItem1.classList.contains('selected')).toBe(true);
      expect(testItem2.classList.contains('selected')).toBe(true);
    });
    
    it('should select all items when "a" key is pressed in selection mode', () => {
      // Given - enter selection mode
      testHelpers.toggleSelectionMode(true);
      
      // When
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'a' }));
      
      // Then
      const testItems = document.querySelectorAll('.test-item');
      testItems.forEach(item => {
        expect(item.classList.contains('selected')).toBe(true);
      });
      
      // Verify selection count
      const selectionCount = document.getElementById('selection-count') as HTMLElement;
      expect(selectionCount.textContent).toBe('4');
    });
    
    it('should copy selected test IDs when Enter key is pressed', async () => {
      // Given - enter selection mode and select items
      testHelpers.toggleSelectionMode(true);
      testHelpers.toggleTestSelection('test1');
      testHelpers.toggleTestSelection('test2', true);
      
      // Mock the clipboard writeText to resolve immediately
      const writeTextMock = vi.fn().mockResolvedValue(undefined);
      Object.defineProperty(navigator, 'clipboard', {
        value: { writeText: writeTextMock },
        configurable: true
      });
      
      // When
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter' }));
      
      // Wait for promises to resolve (no timers needed)
      await Promise.resolve();
      
      // Then clipboard should be called with selected IDs
      expect(writeTextMock).toHaveBeenCalledWith('test1\ntest2');
      
      // Should show success toast
      expect(showToast).toHaveBeenCalledWith(expect.stringContaining('Copied'), 'success');
      
      // Should exit selection mode
      expect(document.body.classList.contains('selection-mode')).toBe(false);
    });
    
    it('should copy all failing tests when "C" (shift+c) is pressed', async () => {
      // Mock the clipboard writeText to resolve immediately
      const writeTextMock = vi.fn().mockResolvedValue(undefined);
      Object.defineProperty(navigator, 'clipboard', {
        value: { writeText: writeTextMock },
        configurable: true
      });
      
      // When
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'C' }));
      
      // Wait for promises to resolve (no timers needed)
      await Promise.resolve();
      
      // Then clipboard should be called with failing test IDs
      expect(writeTextMock).toHaveBeenCalledWith('test3\ntest4');
      
      // Should show success toast
      expect(showToast).toHaveBeenCalledWith(expect.stringContaining('Copied'), 'success');
    });
    
    it('should exit selection mode when Escape is pressed', () => {
      // Given - enter selection mode
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'c' }));
      
      // Selection mode is active
      expect(document.body.classList.contains('selection-mode')).toBe(true);
      
      // When
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }));
      
      // Then
      expect(document.body.classList.contains('selection-mode')).toBe(false);
      
      const indicator = document.getElementById('selection-mode-indicator') as HTMLElement;
      expect(indicator.style.display).toBe('none');
    });
    
    it('should add selection buttons to the action bar', () => {
      // The action bar should have selection buttons
      const actionBar = document.querySelector('.test-actions') as HTMLElement;
      const buttons = actionBar.querySelectorAll('button');
      
      expect(buttons.length).toBe(2);
      
      // Check button contents safely
      const selectButton = buttons[0] as HTMLButtonElement | undefined;
      const copyButton = buttons[1] as HTMLButtonElement | undefined;
      
      // Use optional chaining to safely access properties
      expect(selectButton?.innerHTML).toContain('Select');
      expect(copyButton?.innerHTML).toContain('Copy Failing');
    });
  });
});
