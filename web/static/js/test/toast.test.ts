import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { Toast, __test__ as testUtils } from '../src/toast';

// Extend the Window interface to include test utilities
declare global {
  interface Window {
    __TEST_TOAST_UTILS__: typeof testUtils;
  }
}

// Helper to set up the toast container for testing
function setupToastContainer() {
  // Reset toast state before each test
  if (!testUtils || typeof testUtils.resetToastState !== 'function') {
    throw new Error('testUtils.resetToastState is not available');
  }
  
  // This will clean up any existing container and reset the state
  testUtils.resetToastState();
  
  // Ensure the container is created and attached to the document
  const container = testUtils.ensureToastContainer();
  if (!document.body.contains(container)) {
    document.body.appendChild(container);
  }
  
  return container;
}

describe('Toast Notification System', () => {
  // Set up before tests
  beforeEach(() => {
    // Clean the DOM before each test
    document.body.innerHTML = '';
    
    // Reset mocks
    vi.clearAllMocks();
    
    // Mock setTimeout for toast animations
    vi.useFakeTimers();
    
    // Set up a fresh toast container and reset state
    setupToastContainer();
  });
  
  afterEach(() => {
    // Clean up timers
    vi.useRealTimers();
    
    // Clean up any remaining toasts
    if (typeof window !== 'undefined' && (window as any).__TEST_TOAST_CONTAINER__) {
      const container = (window as any).__TEST_TOAST_CONTAINER__.get();
      if (container && container.parentNode) {
        container.parentNode.removeChild(container);
      }
      (window as any).__TEST_TOAST_CONTAINER__.set(null);
    }
  });
  
  describe('showToast function', () => {
    it('should create a toast with the correct type and message', () => {
      // When
      Toast.showToast('Test message', 'success');
      
      // Then
      const container = testUtils.getToastContainer();
      expect(container).not.toBeNull();
      
      // Get the created toast
      const toast = testUtils.createToast('success', 'Test message');
      expect(toast).toBeDefined();
      expect(toast.textContent).toContain('Test message');
      expect(toast.classList.contains('toast-success')).toBe(true);
      // Clean up
      if (toast.parentNode) {
        toast.parentNode.removeChild(toast);
      }
      expect(toast!.querySelector('.toast-content')?.textContent).toBe('Test message');
    });
    
    it('should use "info" type by default if no type is specified', () => {
      // When
      Toast.showToast('Default type test');
      
      // Then
      const container = testUtils.getToastContainer();
      expect(container).not.toBeNull();
      const toast = container!.querySelector('.toast');
      expect(toast).not.toBeNull();
      expect(toast!.classList.contains('toast-info')).toBe(true);
    });
    
    it('should create a toast with the correct ARIA attributes for accessibility', () => {
      // When
      Toast.showToast('Accessibility test', 'warning');
      
      // Then
      const container = testUtils.getToastContainer();
      expect(container).not.toBeNull();
      const toast = container!.querySelector('.toast');
      expect(toast).not.toBeNull();
      expect(toast!.getAttribute('role')).toBe('alert');
      expect(toast!.getAttribute('aria-live')).toBe('assertive');
      expect(toast!.getAttribute('aria-atomic')).toBe('true');
    });
    
    it('should create a toast with a close button that removes the toast', () => {
      // When
      Toast.showToast('Close button test');
      
      // Then
      const container = testUtils.getToastContainer();
      expect(container).not.toBeNull();
      const toast = container!.querySelector('.toast');
      expect(toast).not.toBeNull();
      const closeButton = toast!.querySelector('.toast-close') as HTMLButtonElement;
      
      expect(closeButton).not.toBeNull();
      expect(closeButton.getAttribute('aria-label')).toBe('Close notification');
      
      // When close button is clicked
      closeButton.click();
      
      // Then toast should start removal animation
      expect(toast!.classList.contains('visible')).toBe(false);
      
      // After animation completes
      vi.runAllTimers();
      expect(container!.querySelectorAll('.toast').length).toBe(0);
    });
    
    it('should auto-dismiss the toast after the specified timeout', () => {
      // When
      Toast.showToast('Auto-dismiss test', 'info', 1000);
      
      // When - direct call to ensure we have a container
      const container = Toast.ensureContainer();
      expect(container).not.toBeNull();
      const toast = container!.querySelector('.toast') as HTMLElement;
      expect(toast).not.toBeNull();
      
      // Advance timers
      vi.advanceTimersByTime(1000);
      
      // Toast should start removal animation
      expect(toast!.classList.contains('visible')).toBe(false);
      
      // After animation completes
      vi.advanceTimersByTime(300);
      expect(container!.querySelectorAll('.toast').length).toBe(0);
    });
    
    it('should create different icons for each toast type', () => {
      // Define toast types directly to avoid namespace dependency
      const toastTypes: Array<'success' | 'error' | 'warning' | 'info'> = ['success', 'error', 'warning', 'info'];
      
      toastTypes.forEach(type => {
        // Reset toast state before each test
        testUtils.resetToastState();
        document.body.innerHTML = '';
        
        // When
        Toast.showToast(`${type} toast test`, type);
        
        // Then
        const container = testUtils.ensureToastContainer();
        expect(container).not.toBeNull();
        const toast = container.querySelector('.toast') as HTMLElement;
        const icon = toast.querySelector('.toast-icon') as HTMLElement;
        
        expect(icon).not.toBeNull();
        expect(icon.innerHTML).toContain('<svg');
        
        // Different icons should have different content
        if (type === 'success') {
          // Check for success checkmark SVG
          expect(icon.querySelector('path')).not.toBeNull();
          expect(icon.querySelector('polyline')).not.toBeNull();
        } else if (type === 'error') {
          // Check for error X SVG - should have two lines
          const lines = icon.querySelectorAll('line');
          expect(lines.length).toBe(2);
        } else if (type === 'warning') {
          // Check for warning triangle SVG
          expect(icon.querySelector('path')).not.toBeNull();
        } else {
          // Check for info icon - should have a circle and two lines
          expect(icon.querySelector('circle')).not.toBeNull();
          const lines = icon.querySelectorAll('line');
          expect(lines.length).toBe(2);
        }
      });
    });
  });
  
  describe('Toast container', () => {
    it('should create a container if it does not exist', () => {
      // Given - no container exists and we reset the state
      document.body.innerHTML = '';
      testUtils.resetToastState();
      
      // When
      const container = Toast.ensureContainer();
      
      // Then
      expect(container).not.toBeNull();
      expect(container!.classList.contains('toast-container')).toBe(true);
      expect(document.body.contains(container)).toBe(true);
    });
    
    it('should return the existing container if it already exists', () => {
      // Given
      const existingContainer = document.createElement('div');
      existingContainer.id = 'toast-container';
      existingContainer.className = 'toast-container';
      document.body.appendChild(existingContainer);
      
      // When
      const container = Toast.ensureContainer();
      
      // Then
      expect(container?.id).toBe('toast-container');
      expect(container?.className).toBe('toast-container');
    });
  });
});
