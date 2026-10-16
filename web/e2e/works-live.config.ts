import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: '.', testMatch: 'works-live.spec.ts', workers: 1, retries: 0,
  reporter: [['list'], ['html', { open: 'never' }]],
  use: { trace: { mode: 'retain-on-failure', screenshots: false }, screenshot: 'off', video: 'off' },
  projects: [{ name: 'chromium-owned-works', use: { ...devices['Desktop Chrome'], viewport: { width: 1920, height: 1080 } } }],
});
