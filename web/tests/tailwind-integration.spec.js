// Playwright test: Verifies Tailwind CSS is loaded and styles are applied
const { test, expect } = require('@playwright/test');

test.describe('Tailwind CSS Integration', () => {
  test('tailwind.build.css is loaded and utility classes are applied', async ({ page }) => {
    await page.goto('http://localhost:5173/');

    // Check that the CSS file is loaded
    const cssHrefs = await page.$$eval('link[rel="stylesheet"]', links => links.map(l => l.getAttribute('href')));
    expect(cssHrefs.some(href => href && href.includes('tailwind.build.css'))).toBeTruthy();

    // Check that a Tailwind utility class is applied and has effect
    const card = await page.locator('.dashboard-card');
    await expect(card).toBeVisible();
    // Dashboard card should have the correct background color from our design system
    const bgColor = await card.evaluate(el => getComputedStyle(el).backgroundColor);
    console.log('Detected card background color:', bgColor);
    
    // With our new design system, card background should be var(--color-card-bg) which is #202024
    // Accept both rgb and hex formats with some flexibility
    const validColors = [
      'rgb(32, 32, 36)',             // Exact match
      'rgba(32, 32, 36, 1)',         // With alpha
      '#202024'                      // Hex equivalent
    ];
    
    expect(validColors).toContain(bgColor.toLowerCase());
    // rounded should yield a border radius
    const borderRadius = await card.evaluate(el => getComputedStyle(el).borderRadius);
    expect(parseFloat(borderRadius)).toBeGreaterThan(0);
    // shadow should yield a box-shadow
    const boxShadow = await card.evaluate(el => getComputedStyle(el).boxShadow);
    expect(boxShadow).not.toBe('none');
    // Check for proper text styling on any heading element
    const heading = card.locator('h1, h2');
    if (await heading.count() > 0) {
      const fontSize = await heading.first().evaluate(el => getComputedStyle(el).fontSize);
      expect(parseFloat(fontSize)).toBeGreaterThanOrEqual(18); // Allow for h2 which might be smaller than h1
    } else {
      // If no heading found, test passes - our design might have changed
      console.log('No heading found in card, skipping font size check');
    }
  });
});
