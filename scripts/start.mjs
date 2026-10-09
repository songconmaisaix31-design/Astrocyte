import { join } from 'node:path';
import { root, requireFile, start, stop } from './process.mjs';

const filename = process.platform === 'win32' ? 'dist/astrocyte.exe' : 'dist/astrocyte';
await requireFile(filename);
const child = start(join(root, filename), [], {
  detached: process.platform !== 'win32',
  env: { ...process.env, ASTROCYTE_WEB_DIR: process.env.ASTROCYTE_WEB_DIR ?? join(root, 'web', 'dist') },
});
let stopping = false;
for (const signal of ['SIGINT', 'SIGTERM']) process.on(signal, async () => {
  if (stopping) return;
  stopping = true;
  try { await stop(child); } catch (error) { console.error(error); process.exitCode = 1; }
});
child.on('error', error => { console.error(error); process.exitCode = 1; });
child.on('exit', code => { if (!stopping) process.exitCode = code ?? 1; });
