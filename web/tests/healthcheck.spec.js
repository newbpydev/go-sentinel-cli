// Playwright E2E test for Go Sentinel API health endpoint
// Reads API_URL from .env if available, otherwise defaults to localhost:8080
const { test, expect } = require('@playwright/test');
require('dotenv').config({ path: require('path').resolve(__dirname, '../.env') });

test('API health endpoint is reachable', async ({ request }) => {
  const apiUrl = process.env.API_URL || 'http://localhost:8080';
  const response = await request.get(`${apiUrl}/health`);
  expect(response.status()).toBe(200);
  expect(await response.text()).toContain('ok');
});
