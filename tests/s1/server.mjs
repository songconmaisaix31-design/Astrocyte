import { mkdtemp, rm } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join, resolve, sep } from 'node:path';
import { createServer } from 'node:net';
import { setTimeout as delay } from 'node:timers/promises';
import { DatabaseSync } from 'node:sqlite';
import { findGo, root, run, start, stop, web } from '../../scripts/process.mjs';

async function freePort() {
  const socket = createServer();
  await new Promise((resolve, reject) => {
    socket.once('error', reject);
    socket.listen(0, '127.0.0.1', resolve);
  });
  const port = socket.address().port;
  await new Promise((resolve, reject) => socket.close(error => error ? reject(error) : resolve()));
  return port;
}

async function ready(url, child) {
  const deadline = Date.now() + 30_000;
  while (Date.now() < deadline) {
    if (child.exitCode !== null || child.signalCode !== null) throw new Error(`Server exited before ${url}`);
    try {
      const response = await fetch(url, { signal: AbortSignal.timeout(1000) });
      if (response.ok) return;
    } catch { /* A startup readiness check, not a retry of a business operation. */ }
    await delay(100);
  }
  throw new Error(`Server readiness timed out: ${url}`);
}

/** Existing Go server and optional Vite, isolated from the Playwright/S0 service. */
export async function startS1Server({ browser = false, env: extraEnv = {} } = {}) {
  const temporary = await mkdtemp(join(tmpdir(), 'astrocyte-s1-'));
  const dataDir = join(temporary, 'data');
  const executable = join(temporary, process.platform === 'win32' ? 'server.exe' : 'server');
  const apiPort = await freePort();
  const webPort = browser ? await freePort() : apiPort;
  const apiURL = `http://127.0.0.1:${apiPort}`;
  const webURL = `http://127.0.0.1:${webPort}`;
  const suppliedEnv = typeof extraEnv === 'function' ? extraEnv({ dataDir, temporary }) : extraEnv;
  // Do not inherit a user's active private import roots, Agent scope, fixture mode,
  // ranking settings, or paid processor enablement into repeatable CI.
  const baseEnv = Object.fromEntries(Object.entries(process.env).filter(([name]) => !name.toUpperCase().startsWith('ASTROCYTE_')));
  const env = { ...baseEnv, ...suppliedEnv, ASTROCYTE_DATA_DIR: dataDir, ASTROCYTE_PORT: String(apiPort), ASTROCYTE_WEB_PORT: String(webPort) };
  let apiChild;
  let webChild;

  async function startAPI() {
    apiChild = start(executable, [], { env, detached: process.platform !== 'win32' });
    await ready(`${apiURL}/api/v1/health`, apiChild);
  }
  async function close() {
    // stop accepts only child handles created here; a stopped child is ignored.
    const results = await Promise.allSettled([webChild, apiChild].filter(Boolean).map(stop));
    const failures = results.filter(result => result.status === 'rejected');
    if (failures.length) throw new AggregateError(failures.map(result => result.reason), `S1 server cleanup failed; retained ${temporary}`);
    if (!resolve(temporary).startsWith(resolve(tmpdir()) + sep)) throw new Error('Unsafe S1 cleanup path');
    await rm(temporary, { recursive: true, force: true });
  }
  try {
    await run(await findGo(), ['build', '-mod=readonly', '-o', executable, './cmd/server'], { cwd: root });
    await startAPI();
    if (browser) {
      webChild = start(process.execPath, [join(web, 'node_modules/vite/bin/vite.js'), '--host', '127.0.0.1'], { cwd: web, env, detached: process.platform !== 'win32' });
      await ready(webURL, webChild);
    }
  } catch (error) {
    await close();
    throw error;
  }
  return {
    apiURL, webURL, dataDir, temporary, close,
    // Windows uses the existing helper's force termination of this owned handle.
    // This tests durable restart, not graceful shutdown or arbitrary PID recovery.
    async restart() { await stop(apiChild); await startAPI(); },
    query(sql, ...parameters) {
      const database = new DatabaseSync(join(dataDir, 'state.sqlite'), { readOnly: true });
      try { return database.prepare(sql).all(...parameters); }
      finally { database.close(); }
    },
  };
}
