import { mkdtemp, rm } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join, resolve } from 'node:path';
import { root, web, findGo, requireFile, wait, start, stop } from './process.mjs';
import { summarizeEnvironment } from './summarize.mjs';

await requireFile('cmd/server/main.go');
await requireFile('web/index.html');
const temporary = await mkdtemp(join(tmpdir(), 'astrocyte-dev-'));
const executable = join(temporary, process.platform === 'win32' ? 'server.exe' : 'server');
let env = { ...process.env };
if (process.argv.includes('--ephemeral')) {
  env.ASTROCYTE_DATA_DIR = join(temporary, 'data');
  // Ordinary browser checks must not inherit opt-in model/external extraction.
  env.ASTROCYTE_ENABLE_CODEX_DISTILLATION = 'false';
  env.ASTROCYTE_ENABLE_SUMMARIZE = 'false';
}
env = await summarizeEnvironment(env);
const children = [];
let stopping = false;
async function shutdown(code = 0) {
  if (stopping) return;
  stopping = true;
  const results = await Promise.allSettled(children.map(stop));
  const errors = results.filter(result => result.status === 'rejected');
  for (const error of errors) console.error(error.reason);
  if (!errors.length) {
    // This exact directory is created above, under OS temp, and contains only our process outputs.
    if (!resolve(temporary).startsWith(resolve(tmpdir()) + (process.platform === 'win32' ? '\\' : '/'))) throw new Error('Unsafe temporary cleanup path');
    await rm(temporary, { recursive: true, force: true });
  }
  process.exitCode = errors.length ? 1 : code;
}
for (const signal of ['SIGINT', 'SIGTERM']) process.on(signal, () => { void shutdown(); });
try {
  const build = start(await findGo(), ['build', '-mod=readonly', '-o', executable, './cmd/server'], { detached: process.platform !== 'win32' });
  children.push(build);
  await wait(build);
  if (!stopping) {
    for (const [command, args, cwd] of [
      [executable, [], root],
      [process.execPath, [join(web, 'node_modules/vite/bin/vite.js'), '--host', '127.0.0.1'], web],
    ]) {
      const child = start(command, args, { cwd, env, detached: process.platform !== 'win32' });
      children.push(child);
      child.once('error', error => { console.error(error); void shutdown(1); });
      child.once('exit', (code, signal) => { if (!stopping) { console.error(`Development child ${command} exited (${code ?? signal})`); void shutdown(code || 1); } });
    }
  }
} catch (error) { console.error(error); await shutdown(1); }
