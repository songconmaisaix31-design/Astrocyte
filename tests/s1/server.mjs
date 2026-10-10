import { lstat, mkdtemp, realpath, rm } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { isAbsolute, join, relative, resolve, sep } from 'node:path';
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

function inside(path, directory) {
  const rel = relative(directory, path);
  return rel === '' || (!rel.startsWith(`..${sep}`) && rel !== '..' && !isAbsolute(rel));
}

async function existingOwnedTemporary(choice) {
  // The caller must name the original helper directory and the exact root
  // approved by the controller. A filename prefix never establishes ownership.
  if (!choice || typeof choice.path !== 'string' || typeof choice.ownedRoot !== 'string' || !isAbsolute(choice.path) || !isAbsolute(choice.ownedRoot)) throw new Error('Reuse requires explicit absolute original path and approved ownedRoot');
  const approved = await realpath(choice.ownedRoot);
  const chosen = await realpath(choice.path);
  const tempRoot = await realpath(tmpdir());
  if (!inside(approved, tempRoot) || approved === tempRoot || !inside(chosen, approved)) throw new Error('Original helper directory must stay within the explicitly approved temporary root');
  for (const [path, directory] of [[chosen, true], [join(chosen, 'data'), true], [join(chosen, 'data', 'state.sqlite'), false], [join(chosen, 'data', 'objects'), true], [join(chosen, process.platform === 'win32' ? 'server.exe' : 'server'), false]]) {
    const info = await lstat(path);
    if (info.isSymbolicLink() || (directory ? !info.isDirectory() : !info.isFile()) || !inside(await realpath(path), approved)) throw new Error('Original helper store must be an existing ordinary SQLite/objects directory within its approved root');
  }
  return { temporary: chosen, ownedRoot: approved };
}

/** Existing Go server and optional Vite, isolated from the Playwright/S0 service. */
export async function startS1Server({ browser = false, env: extraEnv = {}, reuseOwnedTemporary } = {}) {
  const original = reuseOwnedTemporary ? await existingOwnedTemporary(reuseOwnedTemporary) : null;
  const temporary = original?.temporary ?? await mkdtemp(join(tmpdir(), 'astrocyte-s1-'));
  const ownedRoot = original?.ownedRoot ?? temporary;
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
  const env = { ...baseEnv, ASTROCYTE_ENABLE_CODEX_DISTILLATION: 'false', ASTROCYTE_ENABLE_SUMMARIZE: 'false', ASTROCYTE_SUMMARIZE_CLI: '', ...suppliedEnv, ASTROCYTE_DATA_DIR: dataDir, ASTROCYTE_PORT: String(apiPort), ASTROCYTE_WEB_PORT: String(webPort) };
  let apiChild;
  let webChild;
  let disposed = false;

  async function startAPI() {
    apiChild = start(executable, [], { env, detached: process.platform !== 'win32' });
    await ready(`${apiURL}/api/v1/health`, apiChild);
  }
  async function close({ preserveData = Boolean(original) } = {}) {
    if (disposed) return;
    // stop accepts only child handles created here; a stopped child is ignored.
    const results = await Promise.allSettled([webChild, apiChild].filter(Boolean).map(stop));
    const failures = results.filter(result => result.status === 'rejected');
    if (failures.length) throw new AggregateError(failures.map(result => result.reason), `S1 server cleanup failed; retained ${temporary}`);
    if (preserveData) return;
    const target = await realpath(temporary);
    const approved = await realpath(ownedRoot);
    if (!inside(target, approved) || !resolve(target).startsWith(resolve(await realpath(tmpdir())) + sep)) throw new Error('Unsafe S1 cleanup path');
    await rm(temporary, { recursive: true, force: true });
    disposed = true;
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
