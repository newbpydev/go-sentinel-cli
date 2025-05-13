const { test, expect } = require('@playwright/test');

test('API health endpoint is reachable', async ({ request }) => {
  const response = await request.get(process.env.API_URL || 'http://localhost:8080/health');
  expect(response.status()).toBe(200);
  expect(await response.text()).toContain('ok');
});
