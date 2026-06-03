import { defineConfig } from '@playwright/test';

export default defineConfig({
  testDir: '.',
  timeout: 30000,
  expect: {
    timeout: 10000,
  },
  use: {
    headless: true,
    // Mobile user agent for realistic mobile rendering.
    userAgent:
      'Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) ' +
      'AppleWebKit/605.1.15 (KHTML, like Gecko) ' +
      'Version/16.0 Mobile/15E148 Safari/604.1',
    // Allow running against localhost without HTTPS.
    ignoreHTTPSErrors: true,
  },
  // Output directories.
  outputDir: 'web/e2e/test-results',
});