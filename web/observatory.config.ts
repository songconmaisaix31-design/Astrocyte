import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './e2e',
  testMatch: ['observatory.check.ts', 'observatory-focus.check.ts'],
  fullyParallel: false,
  workers: 1,
  retries: 0,
  reporter: 'list',
  outputDir: `../.verification/observatory-browser-${Date.now()}`,
  use: { screenshot: 'off', video: 'off', trace: 'off' },
  projects: [
    { name: 'chromium-1920', use: { ...devices['Desktop Chrome'], viewport: { width: 1920, height: 1080 } } },
    { name: 'chromium-1280', use: { ...devices['Desktop Chrome'], viewport: { width: 1280, height: 720 } } },
  ],
});
