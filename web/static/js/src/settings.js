/**
 * Settings Page JavaScript
 * Handles form validation, submission, and UI interactions
 */

document.addEventListener('DOMContentLoaded', function() {
    const settingsForm = document.getElementById('settings-form');
    const saveButton = document.getElementById('save-all-settings');
    const resetButton = document.getElementById('reset-defaults');
    const feedbackEl = document.getElementById('settings-feedback');
    
    // Form validation
    function validateForm() {
        let isValid = true;
        const errors = [];
        
        // Clear previous validation state
        document.querySelectorAll('.form-group.has-error').forEach(el => {
            el.classList.remove('has-error');
        });
        document.querySelectorAll('.form-error').forEach(el => {
            el.remove();
        });
        
        // Validate test timeout
        const testTimeout = document.getElementById('test-timeout');
        if (testTimeout && (isNaN(testTimeout.value) || testTimeout.value < 1 || testTimeout.value > 300)) {
            markFieldInvalid(testTimeout, 'Test timeout must be between 1 and 300 seconds');
            isValid = false;
            errors.push('Invalid test timeout value');
        }
        
        // Validate parallel tests
        const parallelTests = document.getElementById('parallel-tests');
        if (parallelTests && (isNaN(parallelTests.value) || parallelTests.value < 1 || parallelTests.value > 32)) {
            markFieldInvalid(parallelTests, 'Parallel tests must be between 1 and 32');
            isValid = false;
            errors.push('Invalid parallel tests value');
        }
        
        // Validate coverage threshold
        const coverageThreshold = document.getElementById('coverage-threshold');
        if (coverageThreshold && (isNaN(coverageThreshold.value) || coverageThreshold.value < 0 || coverageThreshold.value > 100)) {
            markFieldInvalid(coverageThreshold, 'Coverage threshold must be between 0 and 100 percent');
            isValid = false;
            errors.push('Invalid coverage threshold value');
        }
        
        // Validate notification duration
        const notificationDuration = document.getElementById('notification-duration');
        if (notificationDuration && (isNaN(notificationDuration.value) || notificationDuration.value < 1 || notificationDuration.value > 30)) {
            markFieldInvalid(notificationDuration, 'Notification duration must be between 1 and 30 seconds');
            isValid = false;
            errors.push('Invalid notification duration value');
        }
        
        // Validate cache duration
        const cacheDuration = document.getElementById('cache-duration');
        if (cacheDuration && (isNaN(cacheDuration.value) || cacheDuration.value < 0 || cacheDuration.value > 1440)) {
            markFieldInvalid(cacheDuration, 'Cache duration must be between 0 and 1440 minutes');
            isValid = false;
            errors.push('Invalid cache duration value');
        }
        
        // Validate data directory
        const dataDirectory = document.getElementById('data-directory');
        if (dataDirectory && dataDirectory.value.trim() === '') {
            markFieldInvalid(dataDirectory, 'Data directory cannot be empty');
            isValid = false;
            errors.push('Data directory is required');
        }
        
        return { isValid, errors };
    }
    
    // Helper to mark a field as invalid
    function markFieldInvalid(field, message) {
        const formGroup = field.closest('.form-group');
        formGroup.classList.add('has-error');
        
        const errorEl = document.createElement('div');
        errorEl.className = 'form-error';
        errorEl.textContent = message;
        formGroup.appendChild(errorEl);
    }
    
    // Show feedback message
    function showFeedback(type, message) {
        feedbackEl.className = 'settings-feedback ' + type;
        feedbackEl.textContent = message;
        feedbackEl.style.display = 'block';
        
        // Auto-hide success messages after 5 seconds
        if (type === 'success') {
            setTimeout(() => {
                feedbackEl.style.display = 'none';
            }, 5000);
        }
    }
    
    // Handle form submission
    if (settingsForm) {
        settingsForm.addEventListener('submit', function(e) {
            e.preventDefault();
            
            const { isValid, errors } = validateForm();
            
            if (!isValid) {
                showFeedback('error', 'Please fix the validation errors before saving');
                return;
            }
            
            // Form is valid, HTMX will handle the actual submission
            // This is triggered by the save button's hx-post attribute
        });
    }
    
    // Handle client-side validation on input changes
    if (settingsForm) {
        const numericInputs = settingsForm.querySelectorAll('input[type="number"]');
        numericInputs.forEach(input => {
            input.addEventListener('input', function() {
                // Remove error state when user starts typing
                const formGroup = this.closest('.form-group');
                formGroup.classList.remove('has-error');
                const errorEl = formGroup.querySelector('.form-error');
                if (errorEl) errorEl.remove();
            });
        });
    }
    
    // Handle theme changes
    const themeSelect = document.getElementById('theme-select');
    if (themeSelect) {
        themeSelect.addEventListener('change', function() {
            // Apply theme immediately for preview
            document.body.className = this.value + '-theme';
            
            // Show preview message
            showFeedback('info', 'Theme preview applied. Save settings to make permanent.');
        });
    }
    
    // Handle font size changes
    const fontSizeSelect = document.getElementById('font-size');
    if (fontSizeSelect) {
        fontSizeSelect.addEventListener('change', function() {
            // Apply font size immediately for preview
            document.documentElement.setAttribute('data-font-size', this.value);
            
            // Show preview message
            showFeedback('info', 'Font size preview applied. Save settings to make permanent.');
        });
    }
    
    // Handle animations toggle
    const animationsToggle = document.getElementById('animations-enabled');
    if (animationsToggle) {
        animationsToggle.addEventListener('change', function() {
            // Apply animations setting immediately for preview
            document.documentElement.setAttribute('data-animations', this.checked ? 'enabled' : 'disabled');
            
            // Show preview message
            showFeedback('info', 'Animation setting preview applied. Save settings to make permanent.');
        });
    }
    
    // Handle HTMX events
    document.body.addEventListener('htmx:afterRequest', function(event) {
        const target = event.detail.target;
        
        // Only handle settings-related responses
        if (target.id === 'settings-feedback' || target.id === 'settings-form') {
            if (event.detail.successful) {
                if (event.detail.xhr.status === 200) {
                    showFeedback('success', 'Settings saved successfully');
                }
            } else {
                // Extract error message from response if available
                let errorMsg = 'Failed to save settings';
                try {
                    const response = JSON.parse(event.detail.xhr.responseText);
                    if (response.message) {
                        errorMsg = response.message;
                    }
                } catch (e) {
                    // Use default error message
                }
                showFeedback('error', errorMsg);
            }
        }
    });
});
