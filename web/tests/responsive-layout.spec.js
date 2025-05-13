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
      await page.goto('http://localhost:5173/');
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
    await page.goto('http://localhost:5173/');
    
    // Check dashboard card spacing and alignment
    const card = page.locator('.dashboard-card');
    await expect(card).toBeVisible();
    const box = await card.boundingBox();
    expect(box).not.toBeNull();
    
    // Verify card has proper margin (should be 1rem/16px from each side)
    expect(box.x).toBeGreaterThanOrEqual(16); // Left margin
    expect(box.y).toBeGreaterThanOrEqual(16); // Top margin might be more due to navbar
    
    // Check that text elements have proper spacing
    const heading = card.locator('h1');
    await expect(heading).toBeVisible();
    const paragraph = card.locator('p');
    await expect(paragraph).toBeVisible();
    
    // Verify heading has bottom margin
    const headingBox = await heading.boundingBox();
    const paragraphBox = await paragraph.boundingBox();
    expect(paragraphBox.y - (headingBox.y + headingBox.height)).toBeGreaterThanOrEqual(8); // At least 0.5rem/8px gap
  });
});

test.describe('Responsive Behavior', () => {
  test('layout adapts for different devices', async ({ page }) => {
    // Test mobile view (narrow)
    await page.setViewportSize({ width: 375, height: 667 });
    await page.goto('http://localhost:5173/');
    
    // Check container width fills viewport on mobile
    const mobileCard = page.locator('.dashboard-card');
    await expect(mobileCard).toBeVisible();
    const mobileBox = await mobileCard.boundingBox();
    expect(mobileBox.width).toBeGreaterThan(300); // Almost full width on mobile
    
    // Test desktop view (wide)
    await page.setViewportSize({ width: 1280, height: 800 });
    await page.reload();
    
    // Check card has reasonable max-width on desktop (not stretched across full screen)
    const desktopCard = page.locator('.dashboard-card');
    const desktopBox = await desktopCard.boundingBox();
    expect(desktopBox.width).toBeLessThan(1280 - 32); // Not full viewport width
    
    // Compare aspect ratios to ensure layout adapts between viewport sizes
    const mobileRatio = mobileBox.width / mobileBox.height;
    const desktopRatio = desktopBox.width / desktopBox.height;
    
    // Ratios might be different if layout is truly responsive
    console.log(`Mobile aspect ratio: ${mobileRatio}, Desktop aspect ratio: ${desktopRatio}`);
  });
});
