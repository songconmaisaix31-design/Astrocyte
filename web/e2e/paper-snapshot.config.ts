import { defineConfig, devices } from '@playwright/test';

// paper_snapshot 验收由既有 S1 helper 独占分配的 API/Vite 端口。
export default defineConfig({
  testDir: '.', testMatch: 's1-paper-snapshot.spec.ts', workers: 1, retries: 0,
  reporter: [['list'], ['html', { open: 'never' }]],
  use: { trace: { mode: 'retain-on-failure', screenshots: false }, screenshot: 'off', video: 'off' },
  projects: [{ name: 'chromium-paper-snapshot', use: { ...devices['Desktop Chrome'], viewport: { width: 1920, height: 1080 } } }],
});
