import { defineConfig } from 'vite';

export default defineConfig({
  server: {
    host: '127.0.0.1', port: 5178, strictPort: true,
    proxy: { '/api': { target: process.env.WORKBENCH_API_ORIGIN || 'http://127.0.0.1:17321' } },
  },
});
