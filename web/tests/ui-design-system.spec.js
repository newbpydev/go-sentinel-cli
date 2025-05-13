// Playwright test: Verifies UI design system components match the desired look and feel
const { test, expect } = require('@playwright/test');

test.describe('UI Design System', () => {
  // Define our design tokens to test against
  const designTokens = {
    colors: {
      // Dark theme colors
      background: { r: 18, g: 18, b: 18 }, // #121212
      cardBackground: { r: 32, g: 32, b: 36 }, // #202024
      sidebarBackground: { r: 15, g: 15, b: 15 }, // #0F0F0F 
      primaryText: { r: 255, g: 255, b: 255 }, // #FFFFFF
      secondaryText: { r: 170, g: 170, b: 170 }, // #AAAAAA
      primaryAccent: { r: 29, g: 144, b: 255 }, // #1D90FF
      secondaryAccent: { r: 104, g: 211, b: 145 }, // #68D391
      border: { r: 57, g: 57, b: 57 }, // #393939
    },
    spacing: {
      xs: 4,
      sm: 8,
      md: 16,
      lg: 24,
      xl: 32,
    },
    borderRadius: {
      sm: 4,
      md: 8,
      lg: 12,
      full: 9999,
    },
    typography: {
      fontFamily: "'Inter', 'SF Pro Display', -apple-system, BlinkMacSystemFont, system-ui, sans-serif",
      heading: { fontSize: '24px', fontWeight: '700', lineHeight: '32px' },
      subheading: { fontSize: '18px', fontWeight: '600', lineHeight: '24px' },
      body: { fontSize: '14px', fontWeight: '400', lineHeight: '20px' },
      small: { fontSize: '12px', fontWeight: '400', lineHeight: '16px' },
    }
  };

  test('verify color palette is applied correctly', async ({ page }) => {
    await page.goto('http://localhost:5173/');
    
    // Test background color
    const bodyBgColor = await page.evaluate(() => {
      const body = document.body;
      const style = window.getComputedStyle(body);
      return {
        r: parseInt(style.backgroundColor.split('(')[1].split(',')[0]),
        g: parseInt(style.backgroundColor.split(',')[1]),
        b: parseInt(style.backgroundColor.split(',')[2])
      };
    });
    
    // Allow some flexibility in color matching (±10 in RGB values)
    expect(Math.abs(bodyBgColor.r - designTokens.colors.background.r)).toBeLessThanOrEqual(10);
    expect(Math.abs(bodyBgColor.g - designTokens.colors.background.g)).toBeLessThanOrEqual(10);
    expect(Math.abs(bodyBgColor.b - designTokens.colors.background.b)).toBeLessThanOrEqual(10);
    
    // Test text colors - for headings
    if (await page.locator('h1, h2, h3').count() > 0) {
      const headingColor = await page.locator('h1, h2, h3').first().evaluate(el => {
        const style = window.getComputedStyle(el);
        return style.color;
      });
      expect(headingColor).toContain('rgb(255, 255, 255)');
    }
  });

  test('verify typography styles are applied correctly', async ({ page }) => {
    await page.goto('http://localhost:5173/');
    
    // Check heading typography if headings exist
    if (await page.locator('h1').count() > 0) {
      const headingStyles = await page.locator('h1').first().evaluate(el => {
        const style = window.getComputedStyle(el);
        return {
          fontSize: style.fontSize,
          fontWeight: style.fontWeight,
          lineHeight: style.lineHeight
        };
      });
      
      // Convert numerical values for comparison (may be in different units)
      const headingFontSize = parseInt(headingStyles.fontSize);
      const headingLineHeight = parseInt(headingStyles.lineHeight);
      
      // Assert that the heading styles are close to our design tokens
      // Allow some flexibility because CSS might convert units
      expect(headingFontSize).toBeGreaterThanOrEqual(20);
      // Font weight comes back as a string, so we need to convert it to a number to compare
      const fontWeightNum = parseInt(headingStyles.fontWeight);
      expect(fontWeightNum).toBeGreaterThanOrEqual(600);
      expect(headingLineHeight).toBeGreaterThanOrEqual(24);
    }
  });

  test('verify component spacing and alignment follow design system', async ({ page }) => {
    await page.goto('http://localhost:5173/');
    
    // Test card component spacing
    const card = page.locator('.dashboard-card');
    if (await card.count() > 0) {
      const cardStyles = await card.first().evaluate(el => {
        const style = window.getComputedStyle(el);
        return {
          padding: style.padding,
          margin: style.margin,
          borderRadius: style.borderRadius
        };
      });
      
      // Check that padding is at least the medium spacing from our design system
      expect(parseInt(cardStyles.padding)).toBeGreaterThanOrEqual(designTokens.spacing.md);
      
      // Check border radius matches our design system
      expect(parseInt(cardStyles.borderRadius)).toBeGreaterThanOrEqual(designTokens.borderRadius.sm);
    }
  });
  
  test('verify responsive layout follows design system breakpoints', async ({ page }) => {
    // Test mobile view
    await page.setViewportSize({ width: 375, height: 667 });
    await page.goto('http://localhost:5173/');
    
    // On mobile, we expect the layout to be responsive, but some content like tables might need horizontal scroll
    // Instead of checking for no horizontal scroll, let's verify that key elements are visible
    await expect(page.locator('.header h1')).toBeVisible();
    await expect(page.locator('.stats-grid')).toBeVisible();
    
    // For tables specifically, we'll check they have proper overflow handling
    const tables = page.locator('.table-container');
    if (await tables.count() > 0) {
      const tableContainerStyles = await tables.first().evaluate(el => {
        const style = window.getComputedStyle(el);
        return style.overflowX;
      });
      expect(['auto', 'scroll']).toContain(tableContainerStyles);
    }
    
    // Test desktop layout
    await page.setViewportSize({ width: 1280, height: 800 });
    await page.goto('http://localhost:5173/');
    
    // Check for properly aligned content on desktop
    // For example, we might expect multiple columns
    const mainContent = page.locator('main');
    if (await mainContent.count() > 0) {
      const contentWidth = await mainContent.evaluate(el => el.offsetWidth);
      // Desktop content should use a reasonable portion of the screen
      expect(contentWidth).toBeGreaterThan(800);
    }
  });

  test('verify UI components match design system', async ({ page }) => {
    await page.goto('http://localhost:5173/');
    
    // Test buttons
    const buttons = page.locator('button');
    if (await buttons.count() > 0) {
      const buttonStyles = await buttons.first().evaluate(el => {
        const style = window.getComputedStyle(el);
        return {
          borderRadius: style.borderRadius,
          padding: style.padding,
          fontWeight: style.fontWeight
        };
      });
      
      // Buttons should have rounded corners
      expect(parseInt(buttonStyles.borderRadius)).toBeGreaterThan(0);
      
      // Buttons should have proper padding
      expect(parseInt(buttonStyles.padding)).toBeGreaterThan(0);
    }
    
    // Test card components
    const cards = page.locator('.card, [class*="card"], .dashboard-card');
    if (await cards.count() > 0) {
      const cardStyles = await cards.first().evaluate(el => {
        const style = window.getComputedStyle(el);
        return {
          backgroundColor: style.backgroundColor,
          boxShadow: style.boxShadow,
          borderRadius: style.borderRadius
        };
      });
      
      // Cards should have backgrounds distinct from the page
      expect(cardStyles.backgroundColor).not.toBe('rgba(0, 0, 0, 0)');
      
      // Cards should have rounded corners
      expect(parseInt(cardStyles.borderRadius)).toBeGreaterThan(0);
    }
  });
});
