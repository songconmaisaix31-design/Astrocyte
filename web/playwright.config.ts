import { defineConfig, devices } from '@playwright/test';

const port = process.env.ASTROCYTE_E2E_WEB_PORT ?? '15173';
export default defineConfig({
  testDir: './e2e',
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: 0,
  reporter: [['list'], ['html', { open: 'never' }]],
  use: { baseURL: `http://127.0.0.1:${port}`, trace: 'retain-on-failure', screenshot: 'only-on-failure' },
  projects: [
    { name: 'chromium-1920', use: { ...devices['Desktop Chrome'], viewport: { width: 1920, height: 1080 } } },
    { name: 'chromium-1280', use: { ...devices['Desktop Chrome'], viewport: { width: 1280, height: 720 } } },
  ],
  webServer: {
    command: 'node ../scripts/dev.mjs --ephemeral',
    url: `http://127.0.0.1:${port}`,
    timeout: 120_000,
    reuseExistingServer: false,
    // POSIX must let dev.mjs stop its detached API/Vite process groups before
    // their inherited stdout/stderr pipes can close. Windows uses tree kill.
    gracefulShutdown: { signal: 'SIGTERM', timeout: 15_000 },
    env: { ASTROCYTE_WEB_PORT: port, ASTROCYTE_PORT: process.env.ASTROCYTE_E2E_API_PORT ?? '18787' },
  },
});
