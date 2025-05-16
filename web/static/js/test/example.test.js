import { describe, it, expect, beforeEach } from 'vitest';

describe('Example Test Suite', () => {
  it('should pass a simple test', () => {
    expect(1 + 1).toBe(2);
  });

  describe('DOM tests', () => {
    beforeEach(() => {
      document.body.innerHTML = `
        <div id="test">Hello World</div>
      `;
    });

    it('should find the test element', () => {
      const element = document.getElementById('test');
      expect(element).not.toBeNull();
      expect(element?.textContent).toBe('Hello World');
    });
  });
});
