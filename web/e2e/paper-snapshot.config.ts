import { defineConfig, devices } from '@playwright/test';

// Real paper_snapshot browser acceptance uses the existing S1 helper to own the
// assigned API/Vite ports. No simultaneous default webServer, personal browser,
// model, or media process.
export default defineConfig({
  testDir: '.', testMatch: 's1-paper-snapshot.spec.ts', workers: 1, retries: 0,
  reporter: [['list'], ['html', { open: 'never' }]],
  use: { trace: 'retain-on-failure', screenshot: 'only-on-failure' },
  projects: [{ name: 'chromium-paper-snapshot', use: { ...devices['Desktop Chrome'], viewport: { width: 1920, height: 1080 } } }],
});
