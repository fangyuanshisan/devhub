import { defineConfig, devices } from '@playwright/test';

const baseURL = process.env.DEVHUB_E2E_ORIGIN || 'http://127.0.0.1:8090';

export default defineConfig({
  testDir: './tests/e2e',
  globalSetup: './tests/e2e/support/global-setup.js',
  timeout: 45_000,
  expect: { timeout: 8_000 },
  fullyParallel: false,
  // The backend is a shared in-memory service during docker-runner E2E. Many specs mutate global
  // plugin/community state (enable/disable/archive) for governance verification, so we keep a
  // single worker to avoid cross-file state races.
  workers: 1,
  retries: process.env.CI ? 1 : 0,
  reporter: [['list'], ['html', { open: 'never', outputFolder: 'playwright-report' }]],
  use: {
    baseURL,
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
});
