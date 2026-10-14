import { defineConfig } from '@playwright/test';
import base from '../playwright.config';

// 每个用例通过既有 S1 helper 持有真实 Go / SQLite / Vite 服务。
export default defineConfig({
  ...base,
  testDir: '.',
  testMatch: ['s2-*.spec.ts', 's3-*.spec.ts'],
  workers: 1,
  retries: 0,
  webServer: undefined,
});
