import { mkdir, mkdtemp } from 'node:fs/promises';
import { join } from 'node:path';
import { setTimeout as delay } from 'node:timers/promises';
import { root, web, findGo, requireFile, wait, start, stop } from './process.mjs';
import { summarizeEnvironment } from './summarize.mjs';

await requireFile('cmd/server/main.go');
await requireFile('web/index.html');
const verification = join(root, '.verification');
await mkdir(verification, { recursive: true });
const temporary = await mkdtemp(join(verification, 'astrocyte-dev-'));
const executable = join(temporary, process.platform === 'win32' ? 'server.exe' : 'server');
let env = { ...process.env };
if (process.argv.includes('--ephemeral')) {
  env.ASTROCYTE_DATA_DIR = join(temporary, 'data');
  // 浏览器检查仅使用本次显式授权的提取与模型配置。
  env.ASTROCYTE_ENABLE_CODEX_DISTILLATION = 'false';
  env.ASTROCYTE_ENABLE_SUMMARIZE = 'false';
  env.ASTROCYTE_PROJECT_DISCOVERY_ROOTS = JSON.stringify([root]);
  // 真实采集的 Chrome 范围通过 S1 helper 显式传入。
  for (const name of Object.keys(env)) {
    if (name.toUpperCase().startsWith('ASTROCYTE_DOUYIN_')) delete env[name];
  }
}
env = await summarizeEnvironment(env);
const children = [];
let stopping = false;
const readiness = new AbortController();
async function shutdown(code = 0) {
  if (stopping) return;
  stopping = true;
  readiness.abort();
  const results = await Promise.allSettled(children.map(stop));
  const errors = results.filter(result => result.status === 'rejected');
  for (const error of errors) console.error(error.reason);
  console.log(`Development data retained: ${temporary}`);
  process.exitCode = errors.length ? 1 : code;
}
for (const signal of ['SIGINT', 'SIGTERM']) process.on(signal, () => { void shutdown(); });

function launch(command, args, cwd) {
  const child = start(command, args, { cwd, env, detached: process.platform !== 'win32' });
  children.push(child);
  child.once('error', error => { console.error(error); void shutdown(1); });
  child.once('exit', (code, signal) => { if (!stopping) { console.error(`Development child ${command} exited (${code ?? signal})`); void shutdown(code || 1); } });
  return child;
}

async function waitForAPI(child) {
  const url = `http://127.0.0.1:${env.ASTROCYTE_PORT || '8787'}/api/v1/health`;
  const deadline = Date.now() + 30_000;
  while (!stopping && Date.now() < deadline) {
    if (child.exitCode !== null || child.signalCode !== null) throw new Error('API exited before becoming ready');
    let response;
    try { response = await fetch(url, { signal: AbortSignal.any([readiness.signal, AbortSignal.timeout(1000)]) }); }
    catch (error) {
      if (stopping && readiness.signal.aborted) return;
      if (error.cause?.code !== 'ECONNREFUSED') throw error;
    }
    if (response) {
      const health = await response.json();
      if (response.ok && health.service === 'astrocyte' && health.status === 'ok') return;
      throw new Error(`API readiness failed: ${response.status}`);
    }
    if (!stopping) await delay(100, undefined, { signal: readiness.signal });
  }
  if (!stopping) throw new Error('API health readiness timed out; frontend was not started');
}

try {
  const build = start(await findGo(), ['build', '-mod=readonly', '-o', executable, './cmd/server'], { detached: process.platform !== 'win32' });
  children.push(build);
  await wait(build);
  if (!stopping) {
    const api = launch(executable, [], root);
    await waitForAPI(api);
    if (!stopping) launch(process.execPath, [join(web, 'node_modules/vite/bin/vite.js'), '--host', '127.0.0.1'], web);
  }
} catch (error) { if (!stopping) { console.error(error); await shutdown(1); } }
