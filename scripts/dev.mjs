import { mkdir, mkdtemp, writeFile } from 'node:fs/promises';
import { join } from 'node:path';
import { setTimeout as delay } from 'node:timers/promises';
import { actualListeningURL, root, web, findGo, launchLogged, requireFile, stop, verificationOutputRoot } from './process.mjs';
import { summarizeEnvironment } from './summarize.mjs';

await requireFile('cmd/server/main.go');
await requireFile('web/index.html');
const verification = verificationOutputRoot();
await mkdir(verification, { recursive: true });
const temporary = await mkdtemp(join(verification, 'astrocyte-dev-'));
const executable = join(temporary, process.platform === 'win32' ? 'server.exe' : 'server');
const toolTemporary = join(temporary, 'tool-temp');
let env = { ...process.env, TEMP: toolTemporary, TMP: toolTemporary, GOTMPDIR: toolTemporary,
  GOCACHE: process.env.GOCACHE ?? join(verification, 'go-cache'), ASTROCYTE_TEST_VITE_CACHE: join(temporary, 'vite-cache') };
await mkdir(toolTemporary, { recursive: true });
await mkdir(env.GOCACHE, { recursive: true });
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
let shutdownPromise;
const readiness = new AbortController();
function shutdown(code = 0) {
  if (shutdownPromise) return shutdownPromise;
  stopping = true;
  readiness.abort();
  shutdownPromise = (async () => {
    const results = await Promise.allSettled(children.map(owned => stop(owned.child, { env })));
    const errors = results.filter(result => result.status === 'rejected').map(result => result.reason);
    if (!errors.length) {
      const output = await Promise.allSettled(children.map(owned => owned.drain()));
      errors.push(...output.filter(result => result.status === 'rejected').map(result => result.reason));
    }
    for (const error of errors) console.error(error);
    console.log(`Development data retained: ${temporary}`);
    if (process.connected) process.disconnect();
    process.exitCode = errors.length ? 1 : code;
  })();
  return shutdownPromise;
}
for (const signal of ['SIGINT', 'SIGTERM']) process.on(signal, () => { void shutdown(); });
if (process.send) process.once('message', record => {
  if (record.type !== 'development_stop' || Object.keys(record).length !== 1) throw new Error('Unexpected development lifecycle command');
  void shutdown();
});

async function launch(command, args, cwd, name, ipc = false, launchEnv = env) {
  const owned = await launchLogged(command, args, { cwd, env: launchEnv, ipc, detached: process.platform !== 'win32' }, temporary, name);
  children.push(owned);
  owned.child.once('error', error => { console.error(error); void shutdown(1); });
  owned.child.once('exit', (code, signal) => {
    if (!stopping) {
      console.error(`Development ${name} exited (${code ?? signal}); stdout=${owned.log.stdout}; stderr=${owned.log.stderr}`);
      void shutdown(code || 1);
    }
  });
  return owned.child;
}

async function waitForAPI(child, apiURL) {
  const url = `${apiURL}/api/v1/health`;
  const deadline = Date.now() + 30_000;
  while (!stopping && Date.now() < deadline) {
    if (child.exitCode !== null || child.signalCode !== null) throw new Error('API exited before becoming ready');
    let response;
    try { response = await fetch(url, { signal: AbortSignal.any([readiness.signal, AbortSignal.timeout(1000)]) }); }
    catch (error) {
      if (stopping && readiness.signal.aborted) throw error;
      if (error.cause?.code !== 'ECONNREFUSED') throw error;
    }
    if (response) {
      const health = await response.json();
      if (response.ok && health.service === 'astrocyte' && health.status === 'ok') return;
      throw new Error(`API readiness failed: ${response.status}`);
    }
    await delay(100, undefined, { signal: readiness.signal });
  }
  throw new Error('API health readiness timed out');
}

try {
  const build = await launchLogged(await findGo(), ['build', '-mod=readonly', '-o', executable, './cmd/server'], { cwd: root, env, detached: process.platform !== 'win32' }, temporary, 'build');
  children.push(build);
  const result = await build.drain();
  if (result.code !== 0) throw new Error(`Actual Go build failed (${result.code ?? result.signal}); stdout=${build.log.stdout}; stderr=${build.log.stderr}`);
  if (!stopping) {
    const webChild = await launch(process.execPath, [join(root, 'scripts/vite-server.mjs')], web, 'vite', true, { ...env, ASTROCYTE_WEB_PORT: env.ASTROCYTE_WEB_PORT ?? '5173' });
    const webURL = await actualListeningURL(webChild, { type: 'vite_listening', field: 'web_url', label: 'Actual Vite', expectedPort: Number(env.ASTROCYTE_WEB_PORT ?? 5173), ipc: true });
    env.ASTROCYTE_WEB_PORT = new URL(webURL).port;
    const apiChild = await launch(executable, [], root, 'api');
    const apiURL = await actualListeningURL(apiChild, { type: 'http_listening', field: 'api_url', label: 'Actual Go', expectedPort: Number(env.ASTROCYTE_PORT ?? 8787) });
    await waitForAPI(apiChild, apiURL);
    const initialized = actualListeningURL(webChild, { type: 'vite_ready', field: 'web_url', label: 'Actual Vite', expectedPort: Number(env.ASTROCYTE_WEB_PORT), ipc: true });
    webChild.send({ api_url: apiURL });
    await initialized;
    console.log(`Development web: ${webURL}; API: ${apiURL}; logs: ${temporary}`);
    if (process.send) process.send({ type: 'development_ready', api_url: apiURL, web_url: webURL, temporary, logs: children.map(owned => owned.log) });
  }
} catch (error) {
  console.error(error);
  try { await writeFile(join(temporary, 'startup-failure.json'), JSON.stringify({ message: error.message, stack: error.stack, logs: children.map(owned => owned.log) }, null, 2), { flag: 'wx', mode: 0o600 }); }
  catch (writeError) { console.error(writeError); }
  await shutdown(1);
}
