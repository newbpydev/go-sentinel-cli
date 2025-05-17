/**
 * Settings Page TypeScript
 * Handles form validation, submission, and UI interactions
 */

// Define interfaces for our settings data
export interface GoSentinelSettings {
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

export interface ValidationError {
  field: string;
  message: string;
}

// DOM Elements references
let settingsForm: HTMLFormElement | null = null;
let saveButton: HTMLButtonElement | null = null;
let resetButton: HTMLButtonElement | null = null;
let feedbackEl: HTMLElement | null = null;

/**
 * Handle form submission
 */
async function handleFormSubmit(e: Event): Promise<void> {
  e.preventDefault();
  
  if (!settingsForm || !saveButton || !feedbackEl) {
    console.error('Required DOM elements not found');
    return;
  }
  
  // Validate form
  const { isValid, errors } = validateForm();
  
  if (!isValid) {
    const errorMessage = errors.length > 0 
      ? errors[0]?.message || 'Validation error' 
      : 'Please fix the validation errors';
    showFeedback('error', errorMessage);
    return;
  }
  
  // Disable save button during save
  saveButton.disabled = true;
  saveButton.textContent = 'Saving...';
  
  try {
    // Get settings from form
    const settings = getSettingsFromForm();
    
    // Save settings
    const success = await saveSettings(settings);
    
    if (success) {
      showFeedback('success', 'Settings saved successfully!');
    } else {
      throw new Error('Failed to save settings');
    }
  } catch (error) {
    console.error('Error saving settings:', error);
    showFeedback('error', 'Failed to save settings. Please try again.');
  } finally {
    // Always re-enable the save button
    saveButton.disabled = false;
    saveButton.textContent = 'Save Settings';
  }
}

/**
 * Handle reset button click
 */
function handleResetClick(e: Event): void {
  e.preventDefault();
  resetSettings(true);
}

/**
 * Load and apply settings from server
 */
async function loadAndApplySettings(): Promise<void> {
  try {
    const settings = await loadSettings();
    if (settings) {
      applySettingsToForm(settings);
    } else {
      // If loading fails, show default settings
      resetSettings(false);
    }
  } catch (error) {
    console.error('Failed to load settings:', error);
    resetSettings(false);
  }
}

/**
 * Validate form inputs
 * @returns Validation result with errors if any
 */
function validateForm(): { isValid: boolean; errors: ValidationError[] } {
  const errors: ValidationError[] = [];
  let isValid = true;
  
  // Clear previous validation state
  document.querySelectorAll('.form-group.has-error').forEach(el => {
    el.classList.remove('has-error');
  });
  document.querySelectorAll('.form-error').forEach(el => {
    el.remove();
  });
  
  // Function to mark field as invalid
  function markFieldInvalid(field: HTMLElement, message: string): void {
    const formGroup = field.closest('.form-group');
    if (formGroup) {
      formGroup.classList.add('has-error');
      
      const errorEl = document.createElement('div');
      errorEl.className = 'form-error';
      errorEl.textContent = message;
      formGroup.appendChild(errorEl);
    }
  }

  // Validate test timeout
  const testTimeout = document.getElementById('test-timeout') as HTMLInputElement;
  const testTimeoutValue = parseInt(testTimeout?.value);
  if (testTimeout && (isNaN(testTimeoutValue) || testTimeoutValue < 1 || testTimeoutValue > 300)) {
    markFieldInvalid(testTimeout, 'Test timeout must be between 1 and 300 seconds');
    isValid = false;
    errors.push({ field: 'test-timeout', message: 'Test timeout must be between 1 and 300 seconds' });
  }
  
  // Validate parallel tests
  const parallelTests = document.getElementById('parallel-tests') as HTMLInputElement;
  const parallelTestsValue = parseInt(parallelTests?.value);
  if (parallelTests && (isNaN(parallelTestsValue) || parallelTestsValue < 1 || parallelTestsValue > 32)) {
    markFieldInvalid(parallelTests, 'Parallel tests must be between 1 and 32');
    isValid = false;
    errors.push({ field: 'parallel-tests', message: 'Parallel tests must be between 1 and 32' });
  }
  
  // Validate coverage threshold
  const coverageThreshold = document.getElementById('coverage-threshold') as HTMLInputElement;
  const coverageThresholdValue = parseInt(coverageThreshold?.value);
  if (coverageThreshold && (isNaN(coverageThresholdValue) || coverageThresholdValue < 0 || coverageThresholdValue > 100)) {
    markFieldInvalid(coverageThreshold, 'Coverage threshold must be between 0 and 100 percent');
    isValid = false;
    errors.push({ field: 'coverage-threshold', message: 'Coverage threshold must be between 0 and 100 percent' });
  }
  
  // Validate notification duration
  const notificationDuration = document.getElementById('notification-duration') as HTMLInputElement;
  const notificationDurationValue = parseInt(notificationDuration?.value);
  if (notificationDuration && (isNaN(notificationDurationValue) || notificationDurationValue < 1 || notificationDurationValue > 30)) {
    markFieldInvalid(notificationDuration, 'Notification duration must be between 1 and 30 seconds');
    isValid = false;
    errors.push({ field: 'notification-duration', message: 'Notification duration must be between 1 and 30 seconds' });
  }
  
  // Validate terminal font size
  const terminalFontSize = document.getElementById('terminal-font-size') as HTMLInputElement;
  const terminalFontSizeValue = parseInt(terminalFontSize?.value);
  if (terminalFontSize && (isNaN(terminalFontSizeValue) || terminalFontSizeValue < 8 || terminalFontSizeValue > 24)) {
    markFieldInvalid(terminalFontSize, 'Terminal font size must be between 8 and 24 pixels');
    isValid = false;
    errors.push({ field: 'terminal-font-size', message: 'Terminal font size must be between 8 and 24 pixels' });
  }
  
  return { isValid, errors };
}

/**
 * Show feedback message to the user
 */
function showFeedback(type: 'success' | 'error' | 'warning', message: string): void {
  if (!feedbackEl) return;
  
  // Clear previous feedback
  feedbackEl.className = 'feedback';
  feedbackEl.textContent = '';
  
  // Set new feedback - use class names that match test expectations
  feedbackEl.classList.add(`feedback-${type}`);
  feedbackEl.textContent = message;
  
  // Auto-hide after 5 seconds
  setTimeout(() => {
    if (feedbackEl) {
      feedbackEl.className = 'feedback';
      feedbackEl.textContent = '';
    }
  }, 5000);
}

/**
 * Get settings from form
 */
function getSettingsFromForm(): GoSentinelSettings {
  const settings: Partial<GoSentinelSettings> = {};
  
  // Helper to get input value by ID
  const getValue = (id: string): string => {
    const el = document.getElementById(id) as HTMLInputElement | HTMLSelectElement | null;
    return el?.value || '';
  };
  
  // Helper to get checkbox value by ID
  const getCheckbox = (id: string): boolean => {
    const el = document.getElementById(id) as HTMLInputElement | null;
    return el?.checked || false;
  };
  
  // Number inputs
  settings.testTimeout = parseInt(getValue('test-timeout')) || 30;
  settings.parallelTests = parseInt(getValue('parallel-tests')) || 4;
  settings.coverageThreshold = parseInt(getValue('coverage-threshold')) || 80;
  settings.notificationDuration = parseInt(getValue('notification-duration')) || 5;
  settings.terminalFontSize = parseInt(getValue('terminal-font-size')) || 14;
  
  // Select inputs
  settings.terminalTheme = getValue('terminal-theme') || 'dark';
  
  // Boolean inputs (checkboxes)
  settings.autoRunTests = getCheckbox('auto-run-tests');
  settings.saveTestLogs = getCheckbox('save-test-logs');
  settings.showFailuresOnly = getCheckbox('show-failures-only');
  settings.animateResults = getCheckbox('animate-results');
  settings.useWebSockets = getCheckbox('use-websockets');
  
  return settings as GoSentinelSettings;
}

/**
 * Apply settings to form inputs
 */
function applySettingsToForm(settings: GoSentinelSettings): void {
  // Apply settings to form fields
  for (const [key, value] of Object.entries(settings)) {
    const kebabKey = key.replace(/([A-Z])/g, '-$1').toLowerCase();
    const input = document.getElementById(kebabKey) as HTMLInputElement | HTMLSelectElement | null;
    
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
function resetSettings(showFeedbackMessage: boolean = true): void {
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
  
  // Show feedback if requested
  if (showFeedbackMessage) {
    showFeedback('success', 'Default settings have been restored. Click save to apply them.');
  }
}

/**
 * Initialize DOM elements and set up event listeners
 */
export function initSettings(): void {
  document.addEventListener('DOMContentLoaded', () => {
    // Initialize DOM elements
    settingsForm = document.getElementById('settings-form') as HTMLFormElement | null;
    saveButton = document.getElementById('save-all-settings') as HTMLButtonElement | null;
    resetButton = document.getElementById('reset-defaults') as HTMLButtonElement | null;
    feedbackEl = document.getElementById('settings-feedback') as HTMLElement | null;
    
    if (!settingsForm || !saveButton || !resetButton || !feedbackEl) {
      console.error('Required DOM elements not found');
      return;
    }
    
    // Set up event listeners
    settingsForm.addEventListener('submit', handleFormSubmit);
    resetButton.addEventListener('click', handleResetClick);
    
    // Load initial settings
    loadAndApplySettings();
  });
}

// Initialize settings when module is loaded
initSettings();

// Export functions and types for testing
export { 
  validateForm, 
  showFeedback, 
  getSettingsFromForm, 
  saveSettings, 
  loadSettings, 
  resetSettings,
  applySettingsToForm,
  handleFormSubmit,
  handleResetClick,
  loadAndApplySettings
};

// Default export for backward compatibility
export default initSettings;
