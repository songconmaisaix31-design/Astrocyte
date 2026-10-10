import { defineConfig, devices } from '@playwright/test';

// The live test uses the existing S1 helper to own the assigned API/Vite ports.
// No simultaneous default webServer, personal browser, model, or media process.
export default defineConfig({
  testDir: '.', testMatch: 's1-project-board.spec.ts', workers: 1, retries: 0,
  reporter: [['list'], ['html', { open: 'never' }]],
  use: { trace: 'retain-on-failure', screenshot: 'only-on-failure' },
  projects: [{ name: 'chromium-owned-board', use: { ...devices['Desktop Chrome'], viewport: { width: 1920, height: 1080 } } }],
});
