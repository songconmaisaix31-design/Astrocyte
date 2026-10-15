import { defineConfig, devices } from '@playwright/test';

// 实际项目验收由既有 S1 helper 独占分配的 API/Vite 端口。
export default defineConfig({
  testDir: '.', testMatch: 's1-project-board.spec.ts', workers: 1, retries: 0,
  reporter: [['list'], ['html', { open: 'never' }]],
  use: { trace: 'retain-on-failure', screenshot: 'only-on-failure' },
  projects: [{ name: 'chromium-owned-board', use: { ...devices['Desktop Chrome'], viewport: { width: 1920, height: 1080 } } }],
});
