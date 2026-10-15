import test from 'node:test';
import assert from 'node:assert/strict';
import { mkdir, mkdtemp, readFile } from 'node:fs/promises';
import { join } from 'node:path';
import { setTimeout as delay } from 'node:timers/promises';
import { actualListeningURL, launchLogged, root, stop, verificationOutputRoot, web } from '../../scripts/process.mjs';
import { startS1Server } from './server.mjs';

test('actual Vite serves early HTTP requests after connecting the actual API without losing them during initialization', { timeout: 90_000 }, async () => {
  const output = verificationOutputRoot();
  await mkdir(output, { recursive: true });
  const directory = await mkdtemp(join(output, 'astrocyte-vite-early-http-'));
  const server = await startS1Server();
  const env = { ...process.env, TEMP: directory, TMP: directory, GOTMPDIR: directory, ASTROCYTE_WEB_PORT: '0', ASTROCYTE_TEST_VITE_CACHE: join(directory, 'vite-cache') };
  let owned;
  try {
    owned = await launchLogged(process.execPath, [join(root, 'scripts/vite-server.mjs')], { cwd: web, env, ipc: true, detached: process.platform !== 'win32' }, directory, 'vite');
    const url = await actualListeningURL(owned.child, { type: 'vite_listening', field: 'web_url', label: 'Actual Vite', ipc: true });
    const earlyRequests = Promise.allSettled([url, `${url}/api/v1/health`].map(async path => {
      const response = await fetch(path, { signal: AbortSignal.timeout(5000) });
      return { status: response.status, body: await response.text() };
    }));
    // 真实 HTTP 请求先进入已绑定的 listener，再完成同一次 Vite 初始化。
    await delay(100);
    const initialized = actualListeningURL(owned.child, { type: 'vite_ready', field: 'web_url', label: 'Actual Vite', ipc: true });
    owned.child.send({ api_url: server.apiURL });
    await initialized;
    const results = await earlyRequests;
    for (const result of results) assert.equal(result.status, 'fulfilled', result.reason?.stack);
    assert.equal(results[0].value.status, 200);
    assert.match(results[0].value.body, /<html/);
    assert.equal(results[1].value.status, 200);
    assert.equal(JSON.parse(results[1].value.body).service, 'astrocyte');
  } finally {
    if (owned) { await stop(owned.child, { env }); await owned.drain(); }
    await server.close();
  }
});

test('actual Vite initialization failure closes early HTTP requests and preserves its original error', { timeout: 20_000 }, async () => {
  const output = verificationOutputRoot();
  await mkdir(output, { recursive: true });
  const directory = await mkdtemp(join(output, 'astrocyte-vite-early-failure-'));
  const env = { ...process.env, TEMP: directory, TMP: directory, GOTMPDIR: directory, ASTROCYTE_WEB_PORT: '0', ASTROCYTE_TEST_VITE_CACHE: join(directory, 'vite-cache') };
  const owned = await launchLogged(process.execPath, [join(root, 'scripts/vite-server.mjs')], { cwd: web, env, ipc: true, detached: process.platform !== 'win32' }, directory, 'vite');
  try {
    const url = await actualListeningURL(owned.child, { type: 'vite_listening', field: 'web_url', label: 'Actual Vite', ipc: true });
    const requested = Promise.allSettled([fetch(url, { signal: AbortSignal.timeout(5000) })]);
    await delay(100);
    // 实际发送不允许的 API origin，配置校验必须关闭本次已绑定的 listener。
    owned.child.send({ api_url: 'http://localhost:8787' });
    const exit = await owned.drain();
    assert.equal(exit.code, 1);
    const [result] = await requested;
    assert.equal(result.status, 'rejected');
    assert.match(await readFile(owned.log.stderr, 'utf8'), /Actual Vite requires its observed loopback API origin/);
    assert.throws(() => process.kill(owned.child.pid, 0), { code: 'ESRCH' });
    await assert.rejects(fetch(url, { signal: AbortSignal.timeout(1000) }));
  } finally { await stop(owned.child, { env }); await owned.drain(); }
});
