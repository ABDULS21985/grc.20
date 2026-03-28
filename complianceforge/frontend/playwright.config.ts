import { defineConfig, devices } from '@playwright/test';

/**
 * ComplianceForge E2E Test Configuration
 *
 * Run all tests:  npx playwright test
 * Run with UI:    npx playwright test --ui
 * Run specific:   npx playwright test e2e/auth.spec.ts
 */
export default defineConfig({
  testDir: './e2e',
  outputDir: './e2e/test-results',

  /* Maximum time a single test can run */
  timeout: 30_000,

  /* Maximum time expect() assertions can wait */
  expect: {
    timeout: 5_000,
  },

  /* Run tests in files in parallel */
  fullyParallel: true,

  /* Fail the build on CI if test.only is left in source */
  forbidOnly: !!process.env.CI,

  /* Retry failed tests on CI */
  retries: process.env.CI ? 2 : 0,

  /* Limit parallel workers on CI to avoid resource contention */
  workers: process.env.CI ? 1 : undefined,

  /* Reporter configuration */
  reporter: process.env.CI
    ? [['html', { outputFolder: 'playwright-report' }], ['github']]
    : [['html', { outputFolder: 'playwright-report', open: 'never' }]],

  /* Shared settings for all projects */
  use: {
    baseURL: process.env.BASE_URL || 'http://localhost:3000',

    /* Collect trace on first retry of a failed test */
    trace: 'on-first-retry',

    /* Capture screenshot on failure */
    screenshot: 'only-on-failure',

    /* Record video on first retry */
    video: 'on-first-retry',

    /* Extra HTTP headers */
    extraHTTPHeaders: {
      'Accept': 'application/json',
    },
  },

  /* Browser projects */
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],

  /* Start the dev server before running tests (local development only) */
  webServer: process.env.CI
    ? undefined
    : {
        command: 'npm run dev',
        url: 'http://localhost:3000',
        reuseExistingServer: true,
        timeout: 120_000,
      },
});
