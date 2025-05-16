/**
 * A simple utility function that greets a user
 * @param name - The name of the user
 * @returns A greeting message
 */
export function greet(name: string): string {
  if (!name) {
    throw new Error('Name is required');
  }
  return `Hello, ${name}! Welcome to Go Sentinel!`;
}

/**
 * A simple utility function that adds two numbers
 * @param a - First number
 * @param b - Second number
 * @returns The sum of a and b
 */
export function add(a: number, b: number): number {
  return a + b;
}

/**
 * A simple utility function that checks if a value is defined
 * @param value - The value to check
 * @returns True if the value is defined, false otherwise
 */
export function isDefined<T>(value: T | undefined | null): value is T {
  return value !== undefined && value !== null;
}

/**
 * A simple utility function that creates a debounced version of a function
 * @param func - The function to debounce
 * @param wait - The number of milliseconds to delay
 * @returns A debounced version of the function
 */
export function debounce<T extends (...args: any[]) => any>(
  func: T,
  wait: number
): (...args: Parameters<T>) => void {
  let timeoutId: ReturnType<typeof setTimeout> | null = null;

  return function debounced(...args: Parameters<T>) {
    if (timeoutId !== null) {
      clearTimeout(timeoutId);
    }

    timeoutId = setTimeout(() => {
      func(...args);
      timeoutId = null;
    }, wait);
  };
}

/**
 * A simple utility function that formats a date
 * @param date - The date to format
 * @returns A formatted date string
 */
export function formatDate(date: Date | string | number): string {
  const d = new Date(date);
  return d.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}
