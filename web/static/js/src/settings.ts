/**
 * Settings Page TypeScript
 * Handles form validation, submission, and UI interactions
 */

// Define interfaces for our settings data
interface GoSentinelSettings {
  testTimeout: number;
  parallelTests: number;
  coverageThreshold: number;
  notificationDuration: number;
  terminalFontSize: number;
  terminalTheme: string;
  autoRunTests: boolean;
  saveTestLogs: boolean;
  showFailuresOnly: boolean;
  animateResults: boolean;
  useWebSockets: boolean;
  [key: string]: string | number | boolean;
}

interface ValidationError {
  field: string;
  message: string;
}

// Initialize event listeners when DOM is ready
document.addEventListener('DOMContentLoaded', function() {
  const settingsForm = document.getElementById('settings-form') as HTMLFormElement;
  const saveButton = document.getElementById('save-all-settings') as HTMLButtonElement;
  const resetButton = document.getElementById('reset-defaults') as HTMLButtonElement;
  const feedbackEl = document.getElementById('settings-feedback') as HTMLElement;
  
  /**
   * Form validation
   * @returns {boolean} Whether the form is valid
   */
  function validateForm(): { isValid: boolean; errors: ValidationError[] } {
    let isValid = true;
    const errors: ValidationError[] = [];
    
    // Clear previous validation state
    document.querySelectorAll('.form-group.has-error').forEach(el => {
      el.classList.remove('has-error');
    });
    document.querySelectorAll('.form-error').forEach(el => {
      el.remove();
    });
    
    // Validate test timeout
    const testTimeout = document.getElementById('test-timeout') as HTMLInputElement;
    const testTimeoutValue = parseInt(testTimeout?.value);
    if (testTimeout && (isNaN(testTimeoutValue) || testTimeoutValue < 1 || testTimeoutValue > 300)) {
      markFieldInvalid(testTimeout, 'Test timeout must be between 1 and 300 seconds');
      isValid = false;
      errors.push({ field: 'test-timeout', message: 'Invalid test timeout value' });
    }
    
    // Validate parallel tests
    const parallelTests = document.getElementById('parallel-tests') as HTMLInputElement;
    const parallelTestsValue = parseInt(parallelTests?.value);
    if (parallelTests && (isNaN(parallelTestsValue) || parallelTestsValue < 1 || parallelTestsValue > 32)) {
      markFieldInvalid(parallelTests, 'Parallel tests must be between 1 and 32');
      isValid = false;
      errors.push({ field: 'parallel-tests', message: 'Invalid parallel tests value' });
    }
    
    // Validate coverage threshold
    const coverageThreshold = document.getElementById('coverage-threshold') as HTMLInputElement;
    const coverageThresholdValue = parseInt(coverageThreshold?.value);
    if (coverageThreshold && (isNaN(coverageThresholdValue) || coverageThresholdValue < 0 || coverageThresholdValue > 100)) {
      markFieldInvalid(coverageThreshold, 'Coverage threshold must be between 0 and 100 percent');
      isValid = false;
      errors.push({ field: 'coverage-threshold', message: 'Invalid coverage threshold value' });
    }
    
    // Validate notification duration
    const notificationDuration = document.getElementById('notification-duration') as HTMLInputElement;
    const notificationDurationValue = parseInt(notificationDuration?.value);
    if (notificationDuration && (isNaN(notificationDurationValue) || notificationDurationValue < 1 || notificationDurationValue > 30)) {
      markFieldInvalid(notificationDuration, 'Notification duration must be between 1 and 30 seconds');
      isValid = false;
      errors.push({ field: 'notification-duration', message: 'Invalid notification duration value' });
    }
    
    // Validate terminal font size
    const terminalFontSize = document.getElementById('terminal-font-size') as HTMLInputElement;
    const terminalFontSizeValue = parseInt(terminalFontSize?.value);
    if (terminalFontSize && (isNaN(terminalFontSizeValue) || terminalFontSizeValue < 8 || terminalFontSizeValue > 24)) {
      markFieldInvalid(terminalFontSize, 'Terminal font size must be between 8 and 24 pixels');
      isValid = false;
      errors.push({ field: 'terminal-font-size', message: 'Invalid terminal font size value' });
    }
    
    return { isValid, errors };
  }

  /**
   * Helper to mark a field as invalid
   * @param field - The input field to mark as invalid
   * @param message - The error message to display
   */
  function markFieldInvalid(field: HTMLElement, message: string): void {
    const formGroup = field.closest('.form-group');
    if (formGroup) {
      formGroup.classList.add('has-error');
      
      // Add error message
      const errorEl = document.createElement('div');
      errorEl.className = 'form-error';
      errorEl.textContent = message;
      formGroup.appendChild(errorEl);
    }
  }
  
  /**
   * Show feedback message
   * @param type - The type of feedback (success, error, warning)
   * @param message - The message to display
   */
  function showFeedback(type: 'success' | 'error' | 'warning', message: string): void {
    if (!feedbackEl) return;
    
    // Clear previous feedback
    feedbackEl.className = '';
    feedbackEl.textContent = '';
    
    // Set new feedback
    feedbackEl.className = `feedback feedback-${type}`;
    feedbackEl.textContent = message;
    feedbackEl.style.display = 'block';
    
    // Auto-hide after 5 seconds for success messages
    if (type === 'success') {
      setTimeout(() => {
        feedbackEl.style.display = 'none';
      }, 5000);
    }
  }
  
  /**
   * Get settings from form
   * @returns Settings object from form values
   */
  function getSettingsFromForm(): GoSentinelSettings {
    const settings: Partial<GoSentinelSettings> = {};
    
    // Number inputs
    const numberInputs = ['test-timeout', 'parallel-tests', 'coverage-threshold', 'notification-duration', 'terminal-font-size'];
    numberInputs.forEach(id => {
      const input = document.getElementById(id) as HTMLInputElement;
      if (input) {
        const key = id.replace(/-([a-z])/g, (_, letter) => letter.toUpperCase()) as keyof GoSentinelSettings;
        settings[key] = parseInt(input.value);
      }
    });
    
    // Select inputs
    const selectInputs = ['terminal-theme'];
    selectInputs.forEach(id => {
      const input = document.getElementById(id) as HTMLSelectElement;
      if (input) {
        const key = id.replace(/-([a-z])/g, (_, letter) => letter.toUpperCase()) as keyof GoSentinelSettings;
        settings[key] = input.value;
      }
    });
    
    // Boolean inputs (checkboxes)
    const booleanInputs = ['auto-run-tests', 'save-test-logs', 'show-failures-only', 'animate-results', 'use-websockets'];
    booleanInputs.forEach(id => {
      const input = document.getElementById(id) as HTMLInputElement;
      if (input) {
        const key = id.replace(/-([a-z])/g, (_, letter) => letter.toUpperCase()) as keyof GoSentinelSettings;
        settings[key] = input.checked;
      }
    });
    
    return settings as GoSentinelSettings;
  }
  
  /**
   * Apply settings to form
   * @param settings - Settings object to apply to form
   */
  function applySettingsToForm(settings: GoSentinelSettings): void {
    // Number inputs
    for (const [key, value] of Object.entries(settings)) {
      const kebabKey = key.replace(/([A-Z])/g, '-$1').toLowerCase();
      const input = document.getElementById(kebabKey) as HTMLInputElement | HTMLSelectElement;
      
      if (input) {
        if (typeof value === 'boolean') {
          (input as HTMLInputElement).checked = value;
        } else {
          input.value = String(value);
        }
      }
    }
  }
  
  /**
   * Save settings to server via API
   * @param settings - Settings to save
   * @returns Promise that resolves when settings are saved
   */
  async function saveSettings(settings: GoSentinelSettings): Promise<boolean> {
    try {
      const response = await fetch('/api/settings', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(settings)
      });
      
      if (!response.ok) {
        throw new Error(`Failed to save settings: ${response.statusText}`);
      }
      
      return true;
    } catch (error) {
      console.error('Error saving settings:', error);
      return false;
    }
  }
  
  /**
   * Load settings from server via API
   * @returns Promise that resolves with settings object
   */
  async function loadSettings(): Promise<GoSentinelSettings | null> {
    try {
      const response = await fetch('/api/settings');
      
      if (!response.ok) {
        throw new Error(`Failed to load settings: ${response.statusText}`);
      }
      
      const settings = await response.json();
      return settings;
    } catch (error) {
      console.error('Error loading settings:', error);
      return null;
    }
  }
  
  /**
   * Reset settings to defaults
   */
  function resetSettings(): void {
    // Define default settings
    const defaults: GoSentinelSettings = {
      testTimeout: 30,
      parallelTests: 4,
      coverageThreshold: 80,
      notificationDuration: 5,
      terminalFontSize: 14,
      terminalTheme: 'dark',
      autoRunTests: true,
      saveTestLogs: true,
      showFailuresOnly: false,
      animateResults: true,
      useWebSockets: true
    };
    
    // Apply defaults to form
    applySettingsToForm(defaults);
    
    // Show feedback
    showFeedback('success', 'Default settings have been restored. Click save to apply them.');
  }
  
  // Handle form submission
  if (settingsForm) {
    settingsForm.addEventListener('submit', async function(e) {
      e.preventDefault();
      
      // Validate form
      const { isValid, errors } = validateForm();
      
      if (!isValid) {
        showFeedback('error', `Please fix the following errors: ${errors.map(e => e.message).join(', ')}`);
        return;
      }
      
      // Get settings from form
      const settings = getSettingsFromForm();
      
      // Disable save button during save
      if (saveButton) {
        saveButton.disabled = true;
        saveButton.textContent = 'Saving...';
      }
      
      // Save settings
      const success = await saveSettings(settings);
      
      // Re-enable save button
      if (saveButton) {
        saveButton.disabled = false;
        saveButton.textContent = 'Save Settings';
      }
      
      // Show feedback
      if (success) {
        showFeedback('success', 'Settings saved successfully!');
        
        // Reload settings to ensure consistency
        const updatedSettings = await loadSettings();
        if (updatedSettings) {
          applySettingsToForm(updatedSettings);
        }
      } else {
        showFeedback('error', 'Failed to save settings. Please try again.');
      }
    });
  }
  
  // Handle reset button
  if (resetButton) {
    resetButton.addEventListener('click', function(e) {
      e.preventDefault();
      resetSettings();
    });
  }
  
  // Load settings on page load
  loadSettings().then(settings => {
    if (settings) {
      applySettingsToForm(settings);
    }
  });
});

// Export types for testing
export type { GoSentinelSettings, ValidationError };
