import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { showToast, ToastType } from '../src/toast';

// Direct import of the createToast function for testing
// This is necessary because it's private in the module
let createToastFn: any;

// Override the module's functions to expose them for testing
// This is a testing-only technique that wouldn't be used in production
vi.mock('../src/toast', async (importOriginal) => {
  const originalModule = await importOriginal<typeof import('../src/toast')>();
  
  return {
    ...originalModule,
    showToast: (message: string, type?: ToastType, timeout?: number) => {
      const result = originalModule.showToast(message, type, timeout);
      
      // Find and expose the private functions for testing if needed
      const toastModule = originalModule as any;
      if (!createToastFn) {
        // Walk the module to find the implementation functions
        Object.keys(toastModule).forEach(key => {
          if (typeof toastModule[key] === 'function' && key.includes('createToast')) {
            createToastFn = toastModule[key];
          }
        });
      }
      
      return result;
    }
  };
});

// Helper to set up the toast container for testing
function setupToastContainer() {
  // Create a fresh container for testing
  const container = document.createElement('div');
  container.id = 'toast-container';
  document.body.appendChild(container);
  return container;
}

describe('Toast Notification System', () => {
  // Set up before tests
  beforeEach(() => {
    // Clean the DOM
    document.body.innerHTML = '';
    
    // Reset mocks 
    vi.clearAllMocks();
    
    // Mock setTimeout for toast animations
    vi.useFakeTimers();
  });
  
  afterEach(() => {
    // Clean up DOM
    document.body.innerHTML = '';
    
    // Reset mocks
    vi.restoreAllMocks();
    vi.useRealTimers();
  });
  
  describe('showToast function', () => {
    it('should create a toast with the correct type and message', () => {
      // Set up container for this test
      const container = setupToastContainer();
      
      // When
      showToast('Test message', 'success');
      
      // Then
      const toasts = container.querySelectorAll('.toast');
      expect(toasts.length).toBe(1);
      
      const toast = toasts[0] as HTMLElement;
      expect(toast.classList.contains('toast-success')).toBe(true);
      expect(toast.querySelector('.toast-content')?.innerHTML).toBe('Test message');
    });
    
    it('should use "info" type by default if no type is specified', () => {
      // Set up container for this test
      const container = setupToastContainer();
      
      // When
      showToast('Default type test');
      
      // Then
      const toast = container.querySelector('.toast') as HTMLElement;
      expect(toast.classList.contains('toast-info')).toBe(true);
    });
    
    it('should create a toast with the correct ARIA attributes for accessibility', () => {
      // Set up container for this test
      const container = setupToastContainer();
      
      // When
      showToast('Accessibility test', 'warning');
      
      // Then
      const toast = container.querySelector('.toast') as HTMLElement;
      expect(toast.getAttribute('role')).toBe('alert');
      expect(toast.getAttribute('aria-live')).toBe('assertive');
      expect(toast.getAttribute('aria-atomic')).toBe('true');
    });
    
    it('should create a toast with a close button that removes the toast', () => {
      // Set up container for this test
      const container = setupToastContainer();
      
      // When
      showToast('Close button test');
      
      // Then
      const toast = container.querySelector('.toast') as HTMLElement;
      const closeButton = toast.querySelector('.toast-close') as HTMLButtonElement;
      
      expect(closeButton).not.toBeNull();
      expect(closeButton.getAttribute('aria-label')).toBe('Close notification');
      
      // When close button is clicked
      closeButton.click();
      
      // Then toast should start removal animation
      expect(toast.classList.contains('visible')).toBe(false);
      
      // After animation completes
      vi.runAllTimers();
      expect(container.querySelectorAll('.toast').length).toBe(0);
    });
    
    it('should auto-dismiss the toast after the specified timeout', () => {
      // Set up container for this test
      const container = setupToastContainer();
      
      // When
      showToast('Auto-dismiss test', 'info', 1000);
      
      // Then toast should be created
      const toast = container.querySelector('.toast') as HTMLElement;
      expect(toast).not.toBeNull();
      
      // Advance timers
      vi.advanceTimersByTime(1000);
      
      // Toast should start removal animation
      expect(toast.classList.contains('visible')).toBe(false);
      
      // After animation completes
      vi.advanceTimersByTime(300);
      expect(container.querySelectorAll('.toast').length).toBe(0);
    });
    
    it('should create different icons for each toast type', () => {
      const toastTypes: ToastType[] = ['success', 'error', 'warning', 'info'];
      
      toastTypes.forEach(type => {
        // Clean previous toasts and create a fresh container
        document.body.innerHTML = '';
        const container = setupToastContainer();
        
        // When
        showToast(`${type} toast test`, type);
        
        // Then
        const toast = container.querySelector('.toast') as HTMLElement;
        const icon = toast.querySelector('.toast-icon') as HTMLElement;
        
        expect(icon).not.toBeNull();
        expect(icon.innerHTML).toContain('<svg');
        
        // Different icons should have different content
        if (type === 'success') {
          expect(icon.innerHTML).toContain('22 4 12 14.01 9 11.01'); // Success checkmark
        } else if (type === 'error') {
          expect(icon.innerHTML).toContain('15 9 9 15'); // Error X
        } else if (type === 'warning') {
          expect(icon.innerHTML).toContain('10.29 3.86'); // Warning triangle
        } else {
          expect(icon.innerHTML).toContain('12 8 12.01 8'); // Info dot
        }
      });
    });
  });
  
  describe('Toast container', () => {
    it('should create a toast container if one does not exist', () => {
      // Given - empty DOM without container
      document.body.innerHTML = ''; 
      
      // When - direct call to ensure we have a container
      showToast('Create container test');
      
      // Then - a container should be created
      const container = document.getElementById('toast-container');
      expect(container).not.toBeNull();
    });
    
    it('should reuse an existing toast container if one exists', () => {
      // Given - a container exists
      document.body.innerHTML = '';
      const container = setupToastContainer();
      
      // When - show a toast
      const appendChildSpy = vi.spyOn(container, 'appendChild');
      showToast('Reuse container test');
      
      // Then - existing container should be used (appendChild will be called)
      expect(appendChildSpy).toHaveBeenCalled();
    });
    
    it('should stack multiple toasts in the container', () => {
      // Given - clean container 
      document.body.innerHTML = '';
      const container = setupToastContainer();
      
      // When - add multiple toasts
      showToast('First toast');
      showToast('Second toast');
      showToast('Third toast');
      
      // Then - all toasts should be in the container
      const toasts = container.querySelectorAll('.toast');
      expect(toasts.length).toBe(3);
    });
  });
});
