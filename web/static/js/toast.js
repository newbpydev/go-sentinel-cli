// Toast notification system for Go Sentinel
// This script listens for HTMX events and creates toast notifications

document.addEventListener('DOMContentLoaded', function() {
    // Create toast container if it doesn't exist
    let toastContainer = document.getElementById('toast-container');
    if (!toastContainer) {
        toastContainer = document.createElement('div');
        toastContainer.id = 'toast-container';
        document.body.appendChild(toastContainer);
    }

    // Listen for showToast events from HTMX
    document.body.addEventListener('showToast', function(event) {
        const toast = event.detail;
        createToast(toast.level, toast.message, toast.timeout);
    });

    // Create and show a toast notification
    function createToast(level, message, timeout = 3000) {
        // Create toast element
        const toast = document.createElement('div');
        toast.className = `toast toast-${level}`;
        toast.setAttribute('role', 'alert');
        toast.setAttribute('aria-live', 'assertive');
        toast.setAttribute('aria-atomic', 'true');

        // Create icon based on level
        const icon = document.createElement('span');
        icon.className = 'toast-icon';
        switch (level) {
            case 'success':
                icon.innerHTML = '✓';
                break;
            case 'error':
                icon.innerHTML = '✕';
                break;
            case 'warning':
                icon.innerHTML = '⚠';
                break;
            case 'info':
            default:
                icon.innerHTML = 'ℹ';
                break;
        }
        toast.appendChild(icon);

        // Create message element
        const messageEl = document.createElement('span');
        messageEl.className = 'toast-message';
        messageEl.textContent = message;
        toast.appendChild(messageEl);

        // Create close button
        const closeBtn = document.createElement('button');
        closeBtn.className = 'toast-close';
        closeBtn.innerHTML = '&times;';
        closeBtn.setAttribute('aria-label', 'Close notification');
        closeBtn.onclick = function() {
            removeToast(toast);
        };
        toast.appendChild(closeBtn);

        // Add to container
        toastContainer.appendChild(toast);

        // Add visible class after a small delay (for animation)
        setTimeout(() => {
            toast.classList.add('visible');
        }, 10);

        // Auto-remove after timeout
        if (timeout > 0) {
            setTimeout(() => {
                removeToast(toast);
            }, timeout);
        }
    }

    // Remove a toast with animation
    function removeToast(toast) {
        toast.classList.remove('visible');
        
        // Remove from DOM after animation completes
        setTimeout(() => {
            if (toast.parentNode) {
                toast.parentNode.removeChild(toast);
            }
        }, 300); // Match this to your CSS transition time
    }

    // Helper functions that can be called from anywhere
    window.toast = {
        success: (message, timeout = 3000) => createToast('success', message, timeout),
        error: (message, timeout = 5000) => createToast('error', message, timeout),
        warning: (message, timeout = 4000) => createToast('warning', message, timeout),
        info: (message, timeout = 3000) => createToast('info', message, timeout)
    };
});
