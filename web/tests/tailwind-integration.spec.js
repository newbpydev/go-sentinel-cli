// Playwright test: Verifies Tailwind CSS is loaded and styles are applied
const { test, expect } = require('@playwright/test');

test.describe('Tailwind CSS Integration', () => {
  test('tailwind.build.css is loaded and utility classes are applied', async ({ page }) => {
    await page.goto('http://localhost:5173/index.html');

    // Check that the CSS file is loaded
    const cssHrefs = await page.$$eval('link[rel="stylesheet"]', links => links.map(l => l.getAttribute('href')));
    expect(cssHrefs.some(href => href && href.includes('tailwind.build.css'))).toBeTruthy();

    // Check that a Tailwind utility class is applied and has effect
    const card = await page.locator('.dashboard-card');
    await expect(card).toBeVisible();
    // bg-white should yield a white background
    const bgColor = await card.evaluate(el => getComputedStyle(el).backgroundColor);
    // Accept both rgb(255,255,255) and #fff
    expect(["rgb(255, 255, 255)", "#fff", "#ffffff"]).toContain(bgColor.toLowerCase());
    // rounded should yield a border radius
    const borderRadius = await card.evaluate(el => getComputedStyle(el).borderRadius);
    expect(parseFloat(borderRadius)).toBeGreaterThan(0);
    // shadow should yield a box-shadow
    const boxShadow = await card.evaluate(el => getComputedStyle(el).boxShadow);
    expect(boxShadow).not.toBe('none');
    // text-2xl should yield a large font size
    const h1 = await card.locator('h1');
    const fontSize = await h1.evaluate(el => getComputedStyle(el).fontSize);
    expect(parseFloat(fontSize)).toBeGreaterThanOrEqual(24);
  });
});
