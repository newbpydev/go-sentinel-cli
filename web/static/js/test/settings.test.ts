import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import type { ValidationError } from '../src/settings';

// Note: We're testing the DOM-based implementation of settings.ts
// The validation and form submission are triggered via DOM events instead of direct function calls

describe('Settings Page', () => {
  let fetchSpy: any;
  let settingsForm: HTMLFormElement;
  let saveButton: HTMLButtonElement;
  let resetButton: HTMLButtonElement;
  let feedbackEl: HTMLElement;

  // Set up test DOM before each test
  beforeEach(() => {
    // Create mocked elements
    settingsForm = document.createElement('form');
    settingsForm.id = 'settings-form';
    
    saveButton = document.createElement('button');
    saveButton.id = 'save-all-settings';
    saveButton.textContent = 'Save Settings';
    
    resetButton = document.createElement('button');
    resetButton.id = 'reset-defaults';
    resetButton.textContent = 'Reset to Defaults';
    
    feedbackEl = document.createElement('div');
    feedbackEl.id = 'settings-feedback';
    
    // Add elements to document
    document.body.appendChild(settingsForm);
    document.body.appendChild(saveButton);
    document.body.appendChild(resetButton);
    document.body.appendChild(feedbackEl);
    
    // Create form fields
    createSettingsFormFields();
    
    // Mock fetch - use type assertion to handle the URL type issue
    fetchSpy = vi.spyOn(global, 'fetch').mockImplementation(mockFetch as any);
    
    // Mock setTimeout
    vi.useFakeTimers();
    
    // Trigger DOMContentLoaded to initialize event handlers
    const event = new Event('DOMContentLoaded');
    document.dispatchEvent(event);
  });
  
  afterEach(() => {
    // Clean up DOM
    document.body.innerHTML = '';
    
    // Restore mocks
    vi.restoreAllMocks();
    vi.useRealTimers();
  });

  // Helper to create form fields
  function createSettingsFormFields() {
    const formFields = [
      { id: 'test-timeout', type: 'number', value: '30', min: '1', max: '300' },
      { id: 'parallel-tests', type: 'number', value: '4', min: '1', max: '32' },
      { id: 'coverage-threshold', type: 'number', value: '80', min: '0', max: '100' },
      { id: 'notification-duration', type: 'number', value: '5', min: '1', max: '30' },
      { id: 'terminal-font-size', type: 'number', value: '14', min: '8', max: '24' }
    ];
    
    const selectFields = [
      { id: 'terminal-theme', options: ['light', 'dark', 'high-contrast'] }
    ];
    
    const checkboxFields = [
      { id: 'auto-run-tests', checked: true },
      { id: 'save-test-logs', checked: true },
      { id: 'show-failures-only', checked: false },
      { id: 'animate-results', checked: true },
      { id: 'use-websockets', checked: true }
    ];
    
    // Create number inputs
    formFields.forEach(field => {
      const formGroup = document.createElement('div');
      formGroup.className = 'form-group';
      
      const label = document.createElement('label');
      label.setAttribute('for', field.id);
      label.textContent = field.id;
      
      const input = document.createElement('input');
      input.id = field.id;
      input.type = field.type;
      input.value = field.value;
      if (field.min) input.min = field.min;
      if (field.max) input.max = field.max;
      
      formGroup.appendChild(label);
      formGroup.appendChild(input);
      settingsForm.appendChild(formGroup);
    });
    
    // Create select inputs
    selectFields.forEach(field => {
      const formGroup = document.createElement('div');
      formGroup.className = 'form-group';
      
      const label = document.createElement('label');
      label.setAttribute('for', field.id);
      label.textContent = field.id;
      
      const select = document.createElement('select');
      select.id = field.id;
      
      field.options.forEach(option => {
        const optionEl = document.createElement('option');
        optionEl.value = option;
        optionEl.textContent = option;
        select.appendChild(optionEl);
      });
      
      formGroup.appendChild(label);
      formGroup.appendChild(select);
      settingsForm.appendChild(formGroup);
    });
    
    // Create checkbox inputs
    checkboxFields.forEach(field => {
      const formGroup = document.createElement('div');
      formGroup.className = 'form-group';
      
      const label = document.createElement('label');
      label.setAttribute('for', field.id);
      label.textContent = field.id;
      
      const input = document.createElement('input');
      input.id = field.id;
      input.type = 'checkbox';
      input.checked = field.checked;
      
      formGroup.appendChild(label);
      formGroup.appendChild(input);
      settingsForm.appendChild(formGroup);
    });
  }
  
  // Mock fetch implementation
  function mockFetch(url: string, options?: RequestInit): Promise<Response> {
    if (url === '/api/settings') {
      if (options?.method === 'POST') {
        // Handle save settings
        const settings = JSON.parse(options.body as string);
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve(settings),
          status: 200,
          statusText: 'OK',
        } as Response);
      } else {
        // Handle load settings
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({
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
          }),
          status: 200,
          statusText: 'OK',
        } as Response);
      }
    }
    
    return Promise.reject(new Error('Not found'));
  }

  describe('Form validation', () => {
    it('should validate test timeout within range', async () => {
      // Set invalid value
      const testTimeout = document.getElementById('test-timeout') as HTMLInputElement;
      testTimeout.value = '500'; // Outside valid range
      
      // Submit form
      settingsForm.dispatchEvent(new Event('submit'));
      
      // Check for error state
      const formGroup = testTimeout.closest('.form-group');
      expect(formGroup?.classList.contains('has-error')).toBe(true);
      
      // Error message should be displayed
      const errorEl = formGroup?.querySelector('.form-error');
      expect(errorEl?.textContent).toContain('must be between 1 and 300');
      
      // Feedback should show error
      expect(feedbackEl.classList.contains('feedback-error')).toBe(true);
    });
    
    it('should validate parallel tests within range', async () => {
      // Set invalid value
      const parallelTests = document.getElementById('parallel-tests') as HTMLInputElement;
      parallelTests.value = '0'; // Below valid range
      
      // Submit form
      settingsForm.dispatchEvent(new Event('submit'));
      
      // Check for error state
      const formGroup = parallelTests.closest('.form-group');
      expect(formGroup?.classList.contains('has-error')).toBe(true);
    });
    
    it('should validate coverage threshold within range', async () => {
      // Set invalid value
      const coverageThreshold = document.getElementById('coverage-threshold') as HTMLInputElement;
      coverageThreshold.value = '101'; // Above valid range
      
      // Submit form
      settingsForm.dispatchEvent(new Event('submit'));
      
      // Check for error state
      const formGroup = coverageThreshold.closest('.form-group');
      expect(formGroup?.classList.contains('has-error')).toBe(true);
    });
    
    it('should validate notification duration within range', async () => {
      // Set invalid value
      const notificationDuration = document.getElementById('notification-duration') as HTMLInputElement;
      notificationDuration.value = 'abc'; // Not a number
      
      // Submit form
      settingsForm.dispatchEvent(new Event('submit'));
      
      // Check for error state
      const formGroup = notificationDuration.closest('.form-group');
      expect(formGroup?.classList.contains('has-error')).toBe(true);
    });
    
    it('should validate terminal font size within range', async () => {
      // Set invalid value
      const terminalFontSize = document.getElementById('terminal-font-size') as HTMLInputElement;
      terminalFontSize.value = '30'; // Above valid range
      
      // Submit form
      settingsForm.dispatchEvent(new Event('submit'));
      
      // Check for error state
      const formGroup = terminalFontSize.closest('.form-group');
      expect(formGroup?.classList.contains('has-error')).toBe(true);
    });
    
    it('should clear errors when valid values are submitted', async () => {
      // First set invalid values and submit
      const testTimeout = document.getElementById('test-timeout') as HTMLInputElement;
      testTimeout.value = '500';
      settingsForm.dispatchEvent(new Event('submit'));
      
      // Check error is shown
      let formGroup = testTimeout.closest('.form-group');
      expect(formGroup?.classList.contains('has-error')).toBe(true);
      
      // Now set valid value and resubmit
      testTimeout.value = '60';
      settingsForm.dispatchEvent(new Event('submit'));
      
      // Error should be cleared
      formGroup = testTimeout.closest('.form-group');
      expect(formGroup?.classList.contains('has-error')).toBe(false);
      expect(formGroup?.querySelector('.form-error')).toBeNull();
    });
  });
  
  describe('Settings API integration', () => {
    it('should load settings on page load', async () => {
      // Verify fetch was called to load settings
      expect(fetchSpy).toHaveBeenCalledWith('/api/settings');
    });
    
    it('should save settings on form submit', async () => {
      // Set up valid values
      const testTimeout = document.getElementById('test-timeout') as HTMLInputElement;
      testTimeout.value = '60';
      
      // Submit form
      settingsForm.dispatchEvent(new Event('submit'));
      
      // Verify fetch was called to save settings
      expect(fetchSpy).toHaveBeenCalledWith('/api/settings', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: expect.any(String)
      });
      
      // Verify the saved data includes the new value
      const lastCallBody = JSON.parse(fetchSpy.mock.calls[fetchSpy.mock.calls.length - 1][1].body);
      expect(lastCallBody.testTimeout).toBe(60);
    });
    
    it('should show success message when settings are saved', async () => {
      // Submit form
      settingsForm.dispatchEvent(new Event('submit'));
      
      // Verify success feedback is shown
      expect(feedbackEl.classList.contains('feedback-success')).toBe(true);
      expect(feedbackEl.textContent).toContain('Settings saved successfully');
    });
    
    it('should handle errors when saving settings', async () => {
      // Mock fetch to reject
      fetchSpy.mockImplementationOnce(() => Promise.resolve({
        ok: false,
        statusText: 'Server Error'
      }));
      
      // Submit form
      settingsForm.dispatchEvent(new Event('submit'));
      
      // Wait for promise rejection to be handled
      await vi.runAllTimersAsync();
      
      // Verify error feedback is shown
      expect(feedbackEl.classList.contains('feedback-error')).toBe(true);
      expect(feedbackEl.textContent).toContain('Failed to save settings');
    });
  });
  
  describe('Reset functionality', () => {
    it('should reset form fields to default values', async () => {
      // First modify values
      const testTimeout = document.getElementById('test-timeout') as HTMLInputElement;
      testTimeout.value = '60';
      
      const showFailuresOnly = document.getElementById('show-failures-only') as HTMLInputElement;
      showFailuresOnly.checked = true;
      
      // Click reset button
      resetButton.click();
      
      // Verify values are reset to defaults
      expect(testTimeout.value).toBe('30');
      expect(showFailuresOnly.checked).toBe(false);
    });
    
    it('should not save defaults until save button is clicked', async () => {
      // Click reset button
      resetButton.click();
      
      // Verify fetch was not called to save settings
      const saveApiCalls = fetchSpy.mock.calls.filter(
        (call: [string, RequestInit]) => call[0] === '/api/settings' && call[1]?.method === 'POST'
      );
      expect(saveApiCalls.length).toBe(0);
      
      // Verify success message indicates settings need to be saved
      expect(feedbackEl.textContent).toContain('Click save to apply');
    });
  });
});
