import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { showToast } from '../src/toast';

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
  // Backup original clipboard API
  const originalClipboard = { ...navigator.clipboard };
  
  beforeEach(() => {
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
    
    // Mock clipboard API
    Object.defineProperty(navigator, 'clipboard', {
      value: {
        writeText: vi.fn().mockResolvedValue(undefined)
      },
      configurable: true
    });
    
    // Dispatch DOMContentLoaded to initialize
    document.dispatchEvent(new Event('DOMContentLoaded'));
  });
  
  afterEach(() => {
    // Restore DOM
    document.body.innerHTML = '';
    
    // Restore clipboard
    Object.defineProperty(navigator, 'clipboard', {
      value: originalClipboard,
      configurable: true
    });
    
    // Reset mocks
    vi.clearAllMocks();
  });
  
  describe('Mobile Menu', () => {
    it('should toggle menu when mobile menu button is clicked', () => {
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
    
    it('should update icon when toggling menu', () => {
      // Given
      const menuToggle = document.getElementById('mobile-menu-toggle') as HTMLButtonElement;
      const icon = menuToggle.querySelector('i') as HTMLElement;
      
      // Initial state
      expect(icon.className).toBe('icon-menu');
      
      // When open
      menuToggle.click();
      
      // Then
      expect(icon.className).toBe('icon-x');
      expect(icon.getAttribute('aria-label')).toBe('Close menu');
      
      // When close
      menuToggle.click();
      
      // Then
      expect(icon.className).toBe('icon-menu');
      expect(icon.getAttribute('aria-label')).toBe('Open menu');
    });
  });
  
  describe('Test Selection', () => {
    it('should enter selection mode when "c" key is pressed', () => {
      // When
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'c' }));
      
      // Then
      expect(document.body.classList.contains('selection-mode')).toBe(true);
      
      const indicator = document.getElementById('selection-mode-indicator') as HTMLElement;
      expect(indicator.style.display).toBe('block');
      expect(indicator.textContent).toBe('Selection mode active');
    });
    
    it('should select a test item when clicked in selection mode', () => {
      // Given - enter selection mode
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'c' }));
      
      // When
      const testItem = document.querySelector('.test-item[data-id="test1"]') as HTMLElement;
      testItem.click();
      
      // Then
      expect(testItem.classList.contains('selected')).toBe(true);
      expect(testItem.getAttribute('aria-selected')).toBe('true');
      
      // Check selection count
      const selectionCount = document.getElementById('selection-count') as HTMLElement;
      expect(selectionCount.textContent).toBe('1');
      expect(selectionCount.classList.contains('has-selected')).toBe(true);
    });
    
    it('should support multi-select with Ctrl key', () => {
      // Given - enter selection mode
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'c' }));
      
      // When - select first item
      const testItem1 = document.querySelector('.test-item[data-id="test1"]') as HTMLElement;
      testItem1.click();
      
      // Then
      expect(testItem1.classList.contains('selected')).toBe(true);
      
      // When - select second item with Ctrl
      const testItem2 = document.querySelector('.test-item[data-id="test2"]') as HTMLElement;
      const clickEvent = new MouseEvent('click', { ctrlKey: true, bubbles: true });
      testItem2.dispatchEvent(clickEvent);
      
      // Then both should be selected
      expect(testItem1.classList.contains('selected')).toBe(true);
      expect(testItem2.classList.contains('selected')).toBe(true);
      
      // Selection count should be 2
      const selectionCount = document.getElementById('selection-count') as HTMLElement;
      expect(selectionCount.textContent).toBe('2');
    });
    
    it('should select all items when "a" key is pressed in selection mode', () => {
      // Given - enter selection mode
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'c' }));
      
      // When
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'a' }));
      
      // Then all items should be selected
      const testItems = document.querySelectorAll('.test-item');
      testItems.forEach(item => {
        expect(item.classList.contains('selected')).toBe(true);
      });
      
      // Selection count should match total items
      const selectionCount = document.getElementById('selection-count') as HTMLElement;
      expect(selectionCount.textContent).toBe(testItems.length.toString());
    });
    
    it('should copy selected test IDs when Enter key is pressed', async () => {
      // Given - enter selection mode and select items
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'c' }));
      
      const testItem1 = document.querySelector('.test-item[data-id="test1"]') as HTMLElement;
      const testItem2 = document.querySelector('.test-item[data-id="test2"]') as HTMLElement;
      testItem1.click();
      
      const clickEvent = new MouseEvent('click', { ctrlKey: true, bubbles: true });
      testItem2.dispatchEvent(clickEvent);
      
      // When - press Enter
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter' }));
      
      // Then clipboard should be called with selected IDs
      expect(navigator.clipboard.writeText).toHaveBeenCalledWith('test1\ntest2');
      
      // Should show success toast
      expect(showToast).toHaveBeenCalledWith(
        expect.stringContaining('Copied 2 test IDs'),
        'success'
      );
      
      // Should exit selection mode
      expect(document.body.classList.contains('selection-mode')).toBe(false);
    });
    
    it('should copy all failing tests when "C" (shift+c) is pressed', () => {
      // When
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'C' }));
      
      // Then clipboard should be called with failing test IDs
      expect(navigator.clipboard.writeText).toHaveBeenCalledWith('test3\ntest4');
      
      // Should show success toast
      expect(showToast).toHaveBeenCalledWith(
        expect.stringContaining('Copied 2 failing test IDs'),
        'success'
      );
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
      // Action bar should have buttons added
      const actionBar = document.querySelector('.test-actions') as HTMLElement;
      const buttons = actionBar.querySelectorAll('button');
      
      expect(buttons.length).toBe(2);
      const selectButton = buttons[0];
      const copyButton = buttons[1];
      
      if (selectButton && copyButton) {
        expect(selectButton.innerHTML).toContain('Select');
        expect(copyButton.innerHTML).toContain('Copy Failing');
      } else {
        fail('Selection buttons not found');
      }
    });
  });
});
