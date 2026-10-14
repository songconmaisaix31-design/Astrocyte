import { lstat, mkdir, mkdtemp, realpath } from 'node:fs/promises';
import { isAbsolute, join, relative, sep } from 'node:path';
import { createServer } from 'node:net';
import { setTimeout as delay } from 'node:timers/promises';
import { DatabaseSync } from 'node:sqlite';
import { findGo, root, run, start, stop, web } from '../../scripts/process.mjs';

async function freePort(requested) {
  if (requested !== undefined && (!Number.isInteger(requested) || requested < 1 || requested > 65535 || [5173, 8787].includes(requested))) {
    throw new Error('Acceptance port must be an integer from 1 to 65535, separate from personal 5173/8787');
  }
  const socket = createServer();
  await new Promise((resolve, reject) => {
    socket.once('error', reject);
    socket.listen(requested ?? 0, '127.0.0.1', resolve);
  });
  const port = socket.address().port;
  await new Promise((resolve, reject) => socket.close(error => error ? reject(error) : resolve()));
  return port;
}

async function ready(url, child) {
  const deadline = Date.now() + 30_000;
  while (Date.now() < deadline) {
    if (child.exitCode !== null || child.signalCode !== null) throw new Error(`Server exited before ${url}`);
    let response;
    try { response = await fetch(url, { signal: AbortSignal.timeout(1000) }); }
    catch (error) { if (error.cause?.code !== 'ECONNREFUSED') throw error; }
    if (response) {
      if (!response.ok) throw new Error(`Server readiness failed: ${url} returned ${response.status}`);
      return;
    }
    await delay(100);
  }
  throw new Error(`Server readiness timed out: ${url}`);
}

function inside(path, directory) {
  const rel = relative(directory, path);
  return rel === '' || (!rel.startsWith(`..${sep}`) && rel !== '..' && !isAbsolute(rel));
}

async function existingOwnedTemporary(choice) {
  // 调用者必须明确原 helper 目录及协调者授权的根目录。
  if (!choice || typeof choice.path !== 'string' || typeof choice.ownedRoot !== 'string' || !isAbsolute(choice.path) || !isAbsolute(choice.ownedRoot)) throw new Error('Reuse requires explicit absolute original path and approved ownedRoot');
  const approved = await realpath(choice.ownedRoot);
  const chosen = await realpath(choice.path);
  const verification = await realpath(join(root, '.verification'));
  if (!inside(approved, verification) || approved === verification || !inside(chosen, approved)) throw new Error('Original helper directory must stay within this worktree verification directory and its explicitly approved root');
  for (const [path, directory] of [[chosen, true], [join(chosen, 'data'), true], [join(chosen, 'data', 'state.sqlite'), false], [join(chosen, 'data', 'objects'), true], [join(chosen, process.platform === 'win32' ? 'server.exe' : 'server'), false]]) {
    const info = await lstat(path);
    if (info.isSymbolicLink() || (directory ? !info.isDirectory() : !info.isFile()) || !inside(await realpath(path), approved)) throw new Error('Original helper store must be an existing ordinary SQLite/objects directory within its approved root');
  }
  return { temporary: chosen, ownedRoot: approved };
}

/** 使用现有 Go 服务及可选 Vite，数据仅放在本工作树的验证目录。 */
export async function startS1Server({ browser = false, env: extraEnv = {}, reuseOwnedTemporary, apiPort: requestedAPIPort, webPort: requestedWebPort } = {}) {
  if (!browser && requestedWebPort !== undefined) throw new Error('A web acceptance port requires browser:true');
  // 创建数据目录和进程前先检查分配端口，已占用的端口立即失败。
  const apiPort = await freePort(requestedAPIPort);
  const webPort = browser ? await freePort(requestedWebPort) : apiPort;
  if (browser && apiPort === webPort) throw new Error('API and Vite acceptance ports must differ');
  const verification = join(root, '.verification');
  await mkdir(verification, { recursive: true });
  const original = reuseOwnedTemporary ? await existingOwnedTemporary(reuseOwnedTemporary) : null;
  const temporary = original?.temporary ?? await mkdtemp(join(verification, 'astrocyte-s1-'));
  const dataDir = join(temporary, 'data');
  const executable = join(temporary, process.platform === 'win32' ? 'server.exe' : 'server');
  const apiURL = `http://127.0.0.1:${apiPort}`;
  const webURL = `http://127.0.0.1:${webPort}`;
  const suppliedEnv = typeof extraEnv === 'function' ? await extraEnv({ dataDir, temporary }) : extraEnv;
  // 仅接收本次明确传入的产品配置，防止继承私人源范围或模型授权。
  const baseEnv = Object.fromEntries(Object.entries(process.env).filter(([name]) => !name.toUpperCase().startsWith('ASTROCYTE_')));
  const env = { ...baseEnv, ASTROCYTE_ENABLE_CODEX_DISTILLATION: 'false', ASTROCYTE_ENABLE_SUMMARIZE: 'false', ASTROCYTE_SUMMARIZE_CLI: '', ASTROCYTE_PROJECT_DISCOVERY_ROOTS: JSON.stringify([root]), ...suppliedEnv, ASTROCYTE_DATA_DIR: dataDir, ASTROCYTE_PORT: String(apiPort), ASTROCYTE_WEB_PORT: String(webPort) };
  let apiChild;
  let webChild;
  let disposed = false;

  async function startAPI() {
    if (disposed) throw new Error('S1 server has been closed');
    if (apiChild && apiChild.exitCode === null && apiChild.signalCode === null) throw new Error('Owned API process is still running');
    apiChild = start(executable, [], { env, detached: process.platform !== 'win32' });
    await ready(`${apiURL}/api/v1/health`, apiChild);
  }
  async function close({ primaryError } = {}) {
    if (disposed) return;
    try {
      // 仅停止本次创建的子进程，已退出的进程由 stop 确认。
      const results = await Promise.allSettled([webChild, apiChild].filter(Boolean).map(stop));
      const failures = results.filter(result => result.status === 'rejected');
      if (failures.length) throw new AggregateError(failures.map(result => result.reason), `S1 server cleanup failed; retained ${temporary}`);
      console.log(`S1 verification data retained: ${temporary}`);
      disposed = true;
    } catch (cleanupError) {
      if (primaryError === undefined) throw cleanupError;
      // 调用者将抛出首个错误；清理失败另行保留并报告。
      console.error('S1 cleanup failed; original error retained and data preserved:', cleanupError);
      if (primaryError !== null && (typeof primaryError === 'object' || typeof primaryError === 'function')) {
        if (Object.isExtensible(primaryError)) Object.defineProperty(primaryError, 'cleanupError', { value: cleanupError, configurable: true });
      }
    }
  }
  try {
    await run(await findGo(), ['build', '-mod=readonly', '-o', executable, './cmd/server'], { cwd: root });
    await startAPI();
    if (browser) {
      webChild = start(process.execPath, [join(web, 'node_modules/vite/bin/vite.js'), '--host', '127.0.0.1'], { cwd: web, env, detached: process.platform !== 'win32' });
      await ready(webURL, webChild);
    }
  } catch (error) {
    await close({ primaryError: error });
    throw error;
  }
  return {
    apiURL, webURL, dataDir, temporary, close,
    startAPI,
    async stopAPI() { await stop(apiChild); },
    // 通过本次实际服务进程的退出和新进程启动验证持久化恢复。
    async restart() { await stop(apiChild); await startAPI(); },
    query(sql, ...parameters) {
      const database = new DatabaseSync(join(dataDir, 'state.sqlite'), { readOnly: true });
      try { return database.prepare(sql).all(...parameters); }
      finally { database.close(); }
    },
  };
}
