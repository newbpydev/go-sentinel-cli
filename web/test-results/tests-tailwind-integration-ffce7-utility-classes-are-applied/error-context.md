# Test info

- Name: Tailwind CSS Integration >> tailwind.build.css is loaded and utility classes are applied
- Location: C:\Users\Remym\pythonProject\__personal-projects\go-sentinel\web\tests\tailwind-integration.spec.js:5:3

# Error details

```
Error: expect(received).toBeTruthy()

Received: false
    at C:\Users\Remym\pythonProject\__personal-projects\go-sentinel\web\tests\tailwind-integration.spec.js:10:80
```

# Page snapshot

```yaml
- text: "Error: Failed to lookup view \"index\" in views directory \"C:\\Users\\Remym\\pythonProject\\__personal-projects\\go-sentinel\\web\\templates\\pages\" at Function.render (C:\\Users\\Remym\\pythonProject\\__personal-projects\\go-sentinel\\web\\node_modules\\express\\lib\\application.js:562:17) at ServerResponse.render (C:\\Users\\Remym\\pythonProject\\__personal-projects\\go-sentinel\\web\\node_modules\\express\\lib\\response.js:909:7) at C:\\Users\\Remym\\pythonProject\\__personal-projects\\go-sentinel\\web\\dev.server.js:24:7 at Layer.handleRequest (C:\\Users\\Remym\\pythonProject\\__personal-projects\\go-sentinel\\web\\node_modules\\router\\lib\\layer.js:152:17) at next (C:\\Users\\Remym\\pythonProject\\__personal-projects\\go-sentinel\\web\\node_modules\\router\\lib\\route.js:157:13) at Route.dispatch (C:\\Users\\Remym\\pythonProject\\__personal-projects\\go-sentinel\\web\\node_modules\\router\\lib\\route.js:117:3) at handle (C:\\Users\\Remym\\pythonProject\\__personal-projects\\go-sentinel\\web\\node_modules\\router\\index.js:435:11) at Layer.handleRequest (C:\\Users\\Remym\\pythonProject\\__personal-projects\\go-sentinel\\web\\node_modules\\router\\lib\\layer.js:152:17) at C:\\Users\\Remym\\pythonProject\\__personal-projects\\go-sentinel\\web\\node_modules\\router\\index.js:295:15 at processParams (C:\\Users\\Remym\\pythonProject\\__personal-projects\\go-sentinel\\web\\node_modules\\router\\index.js:582:12)"
```

# Test source

```ts
   1 | // Playwright test: Verifies Tailwind CSS is loaded and styles are applied
   2 | const { test, expect } = require('@playwright/test');
   3 |
   4 | test.describe('Tailwind CSS Integration', () => {
   5 |   test('tailwind.build.css is loaded and utility classes are applied', async ({ page }) => {
   6 |     await page.goto('http://localhost:5173/index.html');
   7 |
   8 |     // Check that the CSS file is loaded
   9 |     const cssHrefs = await page.$$eval('link[rel="stylesheet"]', links => links.map(l => l.getAttribute('href')));
> 10 |     expect(cssHrefs.some(href => href && href.includes('tailwind.build.css'))).toBeTruthy();
     |                                                                                ^ Error: expect(received).toBeTruthy()
  11 |
  12 |     // Check that a Tailwind utility class is applied and has effect
  13 |     const card = await page.locator('.dashboard-card');
  14 |     await expect(card).toBeVisible();
  15 |     // bg-white should yield a white background
  16 |     const bgColor = await card.evaluate(el => getComputedStyle(el).backgroundColor);
  17 |     // Accept both rgb(255,255,255) and #fff
  18 |     expect(["rgb(255, 255, 255)", "#fff", "#ffffff"]).toContain(bgColor.toLowerCase());
  19 |     // rounded should yield a border radius
  20 |     const borderRadius = await card.evaluate(el => getComputedStyle(el).borderRadius);
  21 |     expect(parseFloat(borderRadius)).toBeGreaterThan(0);
  22 |     // shadow should yield a box-shadow
  23 |     const boxShadow = await card.evaluate(el => getComputedStyle(el).boxShadow);
  24 |     expect(boxShadow).not.toBe('none');
  25 |     // text-2xl should yield a large font size
  26 |     const h1 = await card.locator('h1');
  27 |     const fontSize = await h1.evaluate(el => getComputedStyle(el).fontSize);
  28 |     expect(parseFloat(fontSize)).toBeGreaterThanOrEqual(24);
  29 |   });
  30 | });
  31 |
```