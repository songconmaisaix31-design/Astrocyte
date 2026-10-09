import { mkdir } from 'node:fs/promises';
import { join } from 'node:path';
import { findGo, root, run, runTool, requireFile } from './process.mjs';

await requireFile('cmd/server/main.go');
await requireFile('web/index.html');
await mkdir(join(root, 'dist'), { recursive: true });
await run(await findGo(), ['build', '-mod=readonly', '-trimpath', '-o', join(root, 'dist', process.platform === 'win32' ? 'astrocyte.exe' : 'astrocyte'), './cmd/server']);
await runTool('tsc', ['-b']);
await runTool('vite', ['build']);
