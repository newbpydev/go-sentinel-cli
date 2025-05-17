// Toast notification system for Go Sentinel
// This script provides toast notifications for both HTMX events and direct JS calls

// Define toast types for better type safety
type ToastType = 'success' | 'error' | 'warning' | 'info';

// Define the toast interface for global usage
interface ToastAPI {
    success(message: string, timeout?: number): void;
    error(message: string, timeout?: number): void;
    warning(message: string, timeout?: number): void;
    info(message: string, timeout?: number): void;
}

// Create toast container if it doesn't exist
export let toastContainer: HTMLElement | null = null;

// For testing purposes
if (typeof window !== 'undefined') {
  // @ts-ignore - Allow setting toastContainer for testing
  window.__TEST_TOAST_CONTAINER__ = {
    set: (value: HTMLElement | null) => { toastContainer = value; },
    get: () => toastContainer,
    reset: resetToastState
  };
}

// Define the Toast object with all public methods
const Toast = {
  showToast,
  ensureContainer: () => {
    ensureToastContainer();
    return toastContainer;
  },
  createToast,
  removeToast,
  resetState: resetToastState,
  get container() {
    return toastContainer;
  },
  // Add test utilities to the Toast object
  __test__: {
    resetToastState,
    createToast,
    removeToast,
    ensureToastContainer: () => ensureToastContainer(),
    getToastContainer: () => toastContainer
  }
} as const;

// Export the Toast object and types
export { Toast };
export type { ToastAPI, ToastType };

// Export test utilities for testing
export const __test__ = Toast.__test__;

// Expose test utilities globally for testing
if (typeof window !== 'undefined') {
  (window as any).__TEST_TOAST_UTILS__ = __test__;
}

/**
 * Reset the toast state for testing purposes
 * Clears the toast container and removes it from the DOM
 * @internal
 */
export function resetToastState() {
    if (toastContainer && toastContainer.parentNode) {
        toastContainer.parentNode.removeChild(toastContainer);
    }
    toastContainer = null;
}


/**
 * Show a toast notification
 * @param message - The message to display
 * @param type - The type of toast (success, error, warning, info)
 * @param timeout - Time in milliseconds before the toast auto-dismisses
 */
export function showToast(message: string, type: ToastType = 'info', timeout: number = 3000): void {
    ensureToastContainer();
    createToast(type, message, timeout);
}

/**
 * Remove a toast with animation
 * @param toast - The toast element to remove
 * @export Exported for testing purposes
 */
export function removeToast(toast: HTMLElement): void {
    if (!toast) return;
    
    toast.classList.remove('visible');
    
    // Remove from DOM after animation completes
    setTimeout(() => {
        if (toast && toast.parentNode) {
            toast.parentNode.removeChild(toast);
        }
    }, 300); // Match this to your CSS transition time
}

/**
 * Create and show a toast notification
 * @export Exported for testing purposes
 */
export function createToast(level: ToastType, message: string, timeout: number = 3000): HTMLElement {
    // Ensure container exists
    const container = ensureToastContainer();
    
    // Create toast element
    const toast = document.createElement('div');
    toast.className = `toast toast-${level} visible`;
    toast.setAttribute('role', 'alert');
    toast.setAttribute('aria-live', 'assertive');
    toast.setAttribute('aria-atomic', 'true');

    // Create icon based on level
    const icon = document.createElement('span');
    icon.className = 'toast-icon';
    
    // Set appropriate icon class based on level
    switch (level) {
        case 'success':
            icon.innerHTML = '<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path><polyline points="22 4 12 14.01 9 11.01"></polyline></svg>';
            break;
        case 'error':
            icon.innerHTML = '<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><line x1="15" y1="9" x2="9" y2="15"></line><line x1="9" y1="9" x2="15" y2="15"></line></svg>';
            break;
        case 'warning':
            icon.innerHTML = '<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path><line x1="12" y1="9" x2="12" y2="13"></line><line x1="12" y1="17" x2="12.01" y2="17"></line></svg>';
            break;
        default: // info
            icon.innerHTML = '<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="16" x2="12" y2="12"></line><line x1="12" y1="8" x2="12.01" y2="8"></line></svg>';
            break;
    }
    
    // Create content
    const content = document.createElement('div');
    content.className = 'toast-content';
    content.innerHTML = message;
    
    // Create close button
    const closeBtn = document.createElement('button');
    closeBtn.className = 'toast-close';
    closeBtn.innerHTML = '×';
    closeBtn.setAttribute('aria-label', 'Close notification');
    closeBtn.onclick = () => removeToast(toast);
    
    // Assemble toast
    toast.appendChild(icon);
    toast.appendChild(content);
    toast.appendChild(closeBtn);
    
    // Add to container
    container.appendChild(toast);
    
    // Auto-dismiss after timeout
    if (timeout > 0) {
        setTimeout(() => removeToast(toast), timeout);
    }
    
    return toast;
}

/**
 * Ensures the toast container exists in the DOM
 * @returns The toast container element
 * @export Exported for testing purposes
 */
export function ensureToastContainer(): HTMLElement {
    if (toastContainer && document.body.contains(toastContainer)) {
        return toastContainer;
    }
    
    // Check if container exists in the DOM but isn't in our reference
    const existingContainer = document.getElementById('toast-container');
    if (existingContainer) {
        toastContainer = existingContainer;
        return toastContainer;
    }
    
    // Create a new container if none exists
    toastContainer = document.createElement('div');
    toastContainer.id = 'toast-container';
    document.body.appendChild(toastContainer);
    
    return toastContainer;
}

// Initialize the toast system when the DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    ensureToastContainer();
    
    // Create the toast API
    const toast: ToastAPI = {
        success: (message: string, timeout: number = 3000) => showToast(message, 'success', timeout),
        error: (message: string, timeout: number = 5000) => showToast(message, 'error', timeout),
        warning: (message: string, timeout: number = 4000) => showToast(message, 'warning', timeout),
        info: (message: string, timeout: number = 3000) => showToast(message, 'info', timeout)
    };
    
    // Expose the toast object globally
    (window as any).toast = toast;
});

// Export default toast API creator function for direct imports
export function createToastAPI(): ToastAPI {
    ensureToastContainer();
    return {
        success: (message: string, timeout: number = 3000) => showToast(message, 'success', timeout),
        error: (message: string, timeout: number = 5000) => showToast(message, 'error', timeout),
        warning: (message: string, timeout: number = 4000) => showToast(message, 'warning', timeout),
        info: (message: string, timeout: number = 3000) => showToast(message, 'info', timeout)
    };
}
