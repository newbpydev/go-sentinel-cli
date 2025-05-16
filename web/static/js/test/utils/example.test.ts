import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { greet, add, isDefined, debounce, formatDate } from '../../src/utils/example';

describe('Example Utility Functions', () => {
  describe('greet', () => {
    it('should return a greeting with the provided name', () => {
      const name = 'John';
      const result = greet(name);
      expect(result).toBe(`Hello, ${name}! Welcome to Go Sentinel!`);
    });

    it('should throw an error if name is not provided', () => {
      expect(() => greet('')).toThrow('Name is required');
    });
  });

  describe('add', () => {
    it('should add two numbers correctly', () => {
      expect(add(2, 3)).toBe(5);
      expect(add(-1, 5)).toBe(4);
      expect(add(0, 0)).toBe(0);
    });
  });

  describe('isDefined', () => {
    it('should return false for null or undefined', () => {
      expect(isDefined(null)).toBe(false);
      expect(isDefined(undefined)).toBe(false);
    });

    it('should return true for defined values', () => {
      expect(isDefined('')).toBe(true);
      expect(isDefined(0)).toBe(true);
      expect(isDefined(false)).toBe(true);
      expect(isDefined({})).toBe(true);
      expect(isDefined([])).toBe(true);
    });
  });

  describe('debounce', () => {
    beforeEach(() => {
      vi.useFakeTimers();
    });

    afterEach(() => {
      vi.restoreAllMocks();
    });

    it('should call the function only once after the delay', () => {
      const mockFn = vi.fn();
      const debouncedFn = debounce(mockFn, 100);

      // Call it multiple times
      debouncedFn('first');
      debouncedFn('second');
      debouncedFn('third');

      // Fast-forward time
      vi.advanceTimersByTime(50);
      expect(mockFn).not.toHaveBeenCalled();

      // Fast-forward until all timers have been executed
      vi.advanceTimersByTime(100);
      expect(mockFn).toHaveBeenCalledTimes(1);
      expect(mockFn).toHaveBeenCalledWith('third');
    });
  });

  describe('formatDate', () => {
    // Use a simpler approach by mocking the actual date formatting functions
    // instead of trying to mock the Date constructor
    
    it('should format a date string correctly', () => {
      // The actual implementation uses toLocaleString with specific options
      // Let's test the actual behavior instead of mocking it
      const result = formatDate('2023-04-01T12:34:56');
      
      // The exact format might vary by environment, so we'll check for the key components
      expect(result).toContain('April');
      expect(result).toContain('1');
      expect(result).toContain('2023');
      expect(result).toMatch(/1[02]:34 (AM|PM)/);
    });

    it('should handle different date formats', () => {
      // Use a specific timezone for testing to avoid timezone issues
      const dateString = '2023-12-25T00:00:00.000Z'; // UTC time
      const dateNumber = new Date(dateString).getTime();
      const dateObject = new Date(dateString);
      
      // All formats should produce the same formatted output
      const result1 = formatDate(dateString);
      const result2 = formatDate(dateNumber);
      const result3 = formatDate(dateObject);
      
      // Check that all results are the same
      expect(result1).toBe(result2);
      expect(result2).toBe(result3);
      
      // Check that the result contains the expected date components
      // The exact format might vary by timezone, so we'll be more flexible with the day
      expect(result1).toContain('December');
      expect(result1).toMatch(/(24|25)/); // Could be 24th or 25th depending on timezone
      expect(result1).toContain('2023');
      expect(result1).toMatch(/\d{1,2}:\d{2} (AM|PM)/); // Match any time format
    });
  });
});
