import { copyFile, lstat, mkdir, mkdtemp, realpath, writeFile } from 'node:fs/promises';
import { isAbsolute, join, relative, sep } from 'node:path';
import { createServer } from 'node:net';
import { createInterface } from 'node:readline';
import { setTimeout as delay } from 'node:timers/promises';
import { DatabaseSync } from 'node:sqlite';
import { findGo, root, run, start, stop, web } from '../../scripts/process.mjs';

let compiledServer;

// 同一验证进程使用同一次真实构建，所有服务仍有独立可执行文件与数据目录。
async function buildServer(verification) {
  compiledServer ??= (async () => {
    const directory = await mkdtemp(join(verification, 'astrocyte-s1-build-'));
    const executable = join(directory, process.platform === 'win32' ? 'server.exe' : 'server');
    await run(await findGo(), ['build', '-mod=readonly', '-o', executable, './cmd/server'], { cwd: root });
    return executable;
  })();
  return await compiledServer;
}

function validatePort(requested, allowZero = false) {
  if (allowZero && requested === 0) return;
  if (requested !== undefined && (!Number.isInteger(requested) || requested < 1 || requested > 65535 || [5173, 8787].includes(requested))) {
    throw new Error('Acceptance port must be an integer from 1 to 65535, separate from personal 5173/8787');
  }
}

async function freePort(requested) {
  validatePort(requested);
  const socket = createServer();
  await new Promise((resolve, reject) => {
    socket.once('error', reject);
    socket.listen(requested ?? 0, '127.0.0.1', resolve);
  });
  const port = socket.address().port;
  await new Promise((resolve, reject) => socket.close(error => error ? reject(error) : resolve()));
  return port;
}

// 读取实际 Go 进程的一次绑定公告；该公告与 health 响应分别核实。
async function actualAPIURL(child, expectedPort) {
  return await new Promise((resolve, reject) => {
    const lines = createInterface({ input: child.stdout, crlfDelay: Infinity });
    let settled = false;
    const timer = setTimeout(() => finish(new Error('Actual Go listener announcement timed out')), 30_000);
    const onExit = () => finish(new Error('Actual Go listener announcement became unavailable', { cause: new Error('owned server exited before publishing its address') }));
    const onError = error => finish(error);
    const onClose = () => { if (!settled) finish(new Error('Actual Go listener announcement became unavailable', { cause: new Error('owned stdout closed before listener announcement') })); };
    const finish = (error, url) => {
      if (settled) return;
      settled = true;
      clearTimeout(timer);
      child.removeListener('exit', onExit);
      child.removeListener('error', onError);
      lines.removeListener('close', onClose);
      lines.close();
      child.stdout.resume();
      if (error) reject(error); else resolve(url);
    };
    child.once('exit', onExit);
    child.once('error', onError);
    lines.once('close', onClose);
    lines.on('line', line => {
      try {
        const record = JSON.parse(line);
        const url = new URL(record.api_url);
        if (record.type !== 'http_listening' || Object.keys(record).length !== 2 || url.protocol !== 'http:' || url.hostname !== '127.0.0.1' || url.username || url.password || url.pathname !== '/' || url.search || url.hash || record.api_url !== url.origin || !Number.isInteger(Number(url.port)) || Number(url.port) < 1 || Number(url.port) > 65535 || (expectedPort && Number(url.port) !== expectedPort)) throw new Error('Actual Go listener announcement has an invalid or mismatched loopback address');
        finish(undefined, url.origin);
      } catch (error) { finish(error); }
    });
    if (child.exitCode !== null || child.signalCode !== null) onExit();
  });
}

async function ready(url, child) {
  const started = Date.now();
  let requestStarted;
  let phase = 'before_request';
  let status;
  const failure = (message, cause) => {
    const readiness = { url, pid: child.pid ?? null, exit_code: child.exitCode, signal_code: child.signalCode,
      phase, http_status: status ?? null, started_at: new Date(started).toISOString(), elapsed_ms: Date.now() - started,
      request_started_at: requestStarted === undefined ? null : new Date(requestStarted).toISOString(), request_elapsed_ms: requestStarted === undefined ? null : Date.now() - requestStarted };
    const error = new Error(`${message}: ${url}; readiness=${JSON.stringify(readiness)}`, { cause });
    error.readiness = readiness;
    return error;
  };
  const deadline = Date.now() + 30_000;
  while (Date.now() < deadline) {
    if (child.exitCode !== null || child.signalCode !== null) throw failure('Owned server exited before readiness');
    let response;
    try {
      requestStarted = Date.now();
      phase = 'response_headers';
      status = undefined;
      response = await fetch(url, { signal: AbortSignal.timeout(1000) });
      phase = 'response_body';
      status = response.status;
      // 完整消费本次就绪响应，归还实际 HTTP 连接。
      await response.arrayBuffer();
      phase = 'response_complete';
    } catch (error) {
      if (error.cause?.code !== 'ECONNREFUSED') throw failure('Actual readiness GET failed', error);
    }
    if (response) {
      if (!response.ok) throw failure(`Actual readiness GET returned ${response.status}`);
      if (child.exitCode !== null || child.signalCode !== null) throw failure('Owned server exited during readiness');
      return;
    }
    await delay(100);
  }
  throw failure('Actual readiness GET timed out');
}

function inside(path, directory) {
  const rel = relative(directory, path);
  return rel === '' || (!rel.startsWith(`..${sep}`) && rel !== '..' && !isAbsolute(rel));
}

async function existingOwnedTemporary(choice, relay) {
  // 调用者必须明确原 helper 目录及协调者授权的根目录。
  if (!choice || typeof choice.path !== 'string' || typeof choice.ownedRoot !== 'string' || !isAbsolute(choice.path) || !isAbsolute(choice.ownedRoot)) throw new Error('Reuse requires explicit absolute original path and approved ownedRoot');
  const approved = await realpath(choice.ownedRoot);
  const chosen = await realpath(choice.path);
  const verification = await realpath(join(root, '.verification'));
  if (!inside(approved, verification) || approved === verification || !inside(chosen, approved)) throw new Error('Original helper directory must stay within this worktree verification directory and its explicitly approved root');
  const data = relay ? join(chosen, 'local-appdata', 'astrocyte-workbench', 'v0.2-access', 'app-state') : join(chosen, 'data');
  for (const [path, directory] of [[chosen, true], [data, true], [join(data, 'state.sqlite'), false], [join(data, 'objects'), true], [join(chosen, process.platform === 'win32' ? 'server.exe' : 'server'), false]]) {
    const info = await lstat(path);
    if (info.isSymbolicLink() || (directory ? !info.isDirectory() : !info.isFile()) || !inside(await realpath(path), approved)) throw new Error('Original helper store must be an existing ordinary SQLite/objects directory within its approved root');
  }
  return { temporary: chosen, ownedRoot: approved };
}

/** 使用现有 Go 服务及可选 Vite，数据仅放在本工作树的验证目录。 */
export async function startS1Server({ browser = false, relay = false, env: extraEnv = {}, reuseOwnedTemporary, apiPort: requestedAPIPort, webPort: requestedWebPort } = {}) {
  if (typeof relay !== 'boolean') throw new Error('relay must be an explicit boolean');
  if (!browser && requestedWebPort !== undefined) throw new Error('A web acceptance port requires browser:true');
  // API 由同一 Go listener 原子绑定；显式错误配置在创建数据之前失败。
  validatePort(requestedAPIPort, true);
  const webPort = browser ? await freePort(requestedWebPort) : undefined;
  if (browser && requestedAPIPort === webPort) throw new Error('API and Vite acceptance ports must differ');
  const verification = join(root, '.verification');
  await mkdir(verification, { recursive: true });
  const original = reuseOwnedTemporary ? await existingOwnedTemporary(reuseOwnedTemporary, relay) : null;
  const temporary = original?.temporary ?? await mkdtemp(join(verification, 'astrocyte-s1-'));
  const localAppData = relay ? join(temporary, 'local-appdata') : undefined;
  const relayRoot = relay ? join(localAppData, 'astrocyte-workbench', 'v0.2-access') : undefined;
  const dataDir = relay ? join(relayRoot, 'app-state') : join(temporary, 'data');
  if (relay) await mkdir(relayRoot, { recursive: true });
  const toolTemporary = join(temporary, 'tool-temp');
  await mkdir(toolTemporary, { recursive: true });
  const executable = join(temporary, process.platform === 'win32' ? 'server.exe' : 'server');
  let apiURL;
  const browserURL = browser ? `http://127.0.0.1:${webPort}` : undefined;
  const suppliedEnv = typeof extraEnv === 'function' ? await extraEnv({ dataDir, temporary }) : extraEnv;
  // 仅接收本次明确传入的产品配置，防止继承私人源范围或模型授权。
  const baseEnv = Object.fromEntries(Object.entries(process.env).filter(([name]) => !name.toUpperCase().startsWith('ASTROCYTE_')));
  const env = { ...baseEnv, ASTROCYTE_ENABLE_CODEX_DISTILLATION: 'false', ASTROCYTE_ENABLE_SUMMARIZE: 'false', ASTROCYTE_SUMMARIZE_CLI: '', ASTROCYTE_PROJECT_DISCOVERY_ROOTS: JSON.stringify([root]), ...suppliedEnv,
    ...(relay ? { LOCALAPPDATA: localAppData, APPDATA: join(temporary, 'appdata'), HOME: join(temporary, 'profile'), USERPROFILE: join(temporary, 'profile'), ASTROCYTE_RELAY: '1', ASTROCYTE_RELAY_LIVE_SECONDS: '' } : {}),
    TEMP: toolTemporary, TMP: toolTemporary,
    ASTROCYTE_DATA_DIR: dataDir, ASTROCYTE_PORT: String(requestedAPIPort ?? 0), ASTROCYTE_WEB_PORT: String(webPort ?? 5173) };
  let apiChild;
  let webChild;
  let disposed = false;

  async function startAPI() {
    if (disposed) throw new Error('S1 server has been closed');
    if (apiChild && apiChild.exitCode === null && apiChild.signalCode === null) throw new Error('Owned API process is still running');
    apiChild = start(executable, [], { env, stdio: ['ignore', 'pipe', 'inherit'], detached: process.platform !== 'win32' });
    apiURL = await actualAPIURL(apiChild, Number(env.ASTROCYTE_PORT));
    // 后续明确的重启继续绑定原实际端口，Vite 使用同一准确地址。
    env.ASTROCYTE_PORT = new URL(apiURL).port;
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
    await copyFile(await buildServer(verification), executable);
    await startAPI();
    if (browser) {
      webChild = start(process.execPath, [join(web, 'node_modules/vite/bin/vite.js'), '--host', '127.0.0.1', '--strictPort'], { cwd: web, env, detached: process.platform !== 'win32' });
      await ready(browserURL, webChild);
    }
  } catch (error) {
    let failure = error;
    try {
      const causes = [];
      for (let cause = error; cause; cause = cause.cause) causes.push({ name: cause.name, message: cause.message, code: cause.code ?? null });
      await writeFile(join(temporary, 'startup-failure.json'), JSON.stringify({ readiness: error.readiness ?? null, causes }, null, 2));
    } catch (writeError) {
      failure = new AggregateError([error, writeError], `Startup failure could not be retained in ${temporary}`, { cause: error });
    }
    await close({ primaryError: failure });
    throw failure;
  }
  return {
    get apiURL() { return apiURL; },
    get webURL() { return browserURL ?? apiURL; },
    dataDir, temporary, localAppData, relayRoot, close,
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
