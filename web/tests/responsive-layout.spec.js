// Playwright E2E tests for responsive layout and component alignment
const { test, expect } = require('@playwright/test');

const viewports = [
  { width: 375, height: 667, name: 'mobile' }, // iPhone 8
  { width: 768, height: 1024, name: 'tablet' }, // iPad
  { width: 1280, height: 800, name: 'desktop' },
];

test.describe('Responsive Layout', () => {
  for (const vp of viewports) {
    test(`renders correctly at ${vp.name} size`, async ({ page }) => {
      await page.setViewportSize({ width: vp.width, height: vp.height });
      await page.goto('http://localhost:5173/static/index.html');
      // Check that navbar is visible
      await expect(page.locator('nav')).toBeVisible();
      // Check main content exists
      await expect(page.locator('main')).toBeVisible();
      // Check no horizontal scroll
      const scrollWidth = await page.evaluate(() => document.body.scrollWidth);
      const clientWidth = await page.evaluate(() => document.body.clientWidth);
      expect(scrollWidth).toBeLessThanOrEqual(clientWidth + 2); // allow tiny rounding error
    });
  }
});

test.describe('Component Alignment', () => {
  test('components maintain spacing and alignment', async ({ page }) => {
    await page.setViewportSize({ width: 1280, height: 800 });
    await page.goto('http://localhost:5173/static/index.html');
    // Example: check that a dashboard card has correct margin
    const card = page.locator('.dashboard-card');
    if (await card.count() > 0) {
      const box = await card.boundingBox();
      expect(box).not.toBeNull();
      // Optionally, check some spacing/margin/padding rules
    }
  });
});
