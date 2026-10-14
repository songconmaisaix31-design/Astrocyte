import { spawn, spawnSync } from 'node:child_process';
import { access, readFile, readdir } from 'node:fs/promises';
import { constants } from 'node:fs';
import { delimiter, join, relative } from 'node:path';
import { fileURLToPath } from 'node:url';

export const root = fileURLToPath(new URL('../', import.meta.url));
export const web = join(root, 'web');

export async function findGo() {
  const windows = process.platform === 'win32';
  const fromPath = (process.env.PATH ?? '').split(delimiter).filter(Boolean)
    .map(dir => join(dir.replace(/^"|"$/g, ''), windows ? 'go.exe' : 'go'));
  const installed = windows && process.env.LOCALAPPDATA
    ? join(process.env.LOCALAPPDATA, 'Programs', 'go', 'bin', 'go.exe') : undefined;
  const candidates = process.env.ASTROCYTE_GO ? [process.env.ASTROCYTE_GO] : [...fromPath, installed].filter(Boolean);
  for (const candidate of candidates) {
    try {
      await access(candidate, windows ? constants.F_OK : constants.X_OK);
      return candidate;
    } catch (error) {
      if (process.env.ASTROCYTE_GO || error.code !== 'ENOENT') throw error;
    }
  }
  throw new Error('Go not found. Install Go 1.27.2 or set ASTROCYTE_GO to its executable.');
}

export function start(command, args, options = {}) {
  console.log(`> ${command} ${args.join(' ')}`);
  return spawn(command, args, { cwd: root, stdio: 'inherit', windowsHide: true, ...options });
}

export function wait(child) {
  return new Promise((resolve, reject) => {
    child.once('error', reject);
    child.once('exit', (code, signal) => code === 0 ? resolve() : reject(new Error(`Command failed (${code ?? signal})`)));
  });
}

export async function run(command, args, options) { await wait(start(command, args, options)); }
export async function runTool(name, args, options = {}) {
  const locations = {
    tsc: [web, 'typescript/bin/tsc'],
    eslint: [web, 'eslint/bin/eslint.js'],
    vitest: [web, 'vitest/vitest.mjs'],
    vite: [web, 'vite/bin/vite.js'],
    redocly: [root, '@redocly/cli/bin/cli.js'],
  };
  const [base, relative] = locations[name];
  await run(process.execPath, [join(base, 'node_modules', relative), ...args], { cwd: base, ...options });
}

export async function goFiles(dir = root) {
  const files = [];
  for (const entry of await readdir(dir, { withFileTypes: true })) {
    if (['.git', '.verification', 'node_modules', 'dist'].includes(entry.name)) continue;
    const path = join(dir, entry.name);
    if (entry.isDirectory()) files.push(...await goFiles(path));
    else if (entry.isFile() && entry.name.endsWith('.go')) files.push(path);
  }
  return files;
}

export async function checkGofmt(go) {
  const files = await goFiles();
  if (!files.length) throw new Error('No Go files; W1 backend is pending');
  const executable = join(go, '..', process.platform === 'win32' ? 'gofmt.exe' : 'gofmt');
  // 分批检查全部 Go 文件，保持命令长度在 Windows 限制以内。
  for (let offset = 0; offset < files.length; offset += 40) {
    const batch = files.slice(offset, offset + 40).map(file => relative(root, file));
    const result = spawnSync(executable, ['-l', ...batch], { cwd: root, encoding: 'utf8', windowsHide: true });
    if (result.error) throw result.error;
    if (result.status !== 0 || result.stdout.trim()) throw new Error(`gofmt check failed:\n${result.stdout}${result.stderr}`);
  }
  console.log('gofmt: all Go files formatted');
}

// 只终止本次调用启动的子进程及其工具进程。
export async function stop(child) {
  if (!child.pid || child.exitCode !== null || child.signalCode !== null) return;
  // 先处理已排队的退出事件，再使用本次调用持有的子进程 PID。
  await new Promise(resolve => setImmediate(resolve));
  if (child.exitCode !== null || child.signalCode !== null) return;
  let onExit;
  const exited = new Promise(resolve => { onExit = resolve; child.once('exit', onExit); });
  let shutdownError;
  if (process.platform === 'win32') {
    const result = spawnSync('taskkill.exe', ['/pid', String(child.pid), '/T', '/F'], { encoding: 'utf8', windowsHide: true });
    if (result.status !== 0) shutdownError = new Error(`Child shutdown failed: ${result.stderr || result.stdout || result.error}`);
  } else {
    // POSIX 的长期子进程以进程组启动。
    try { process.kill(-child.pid, 'SIGTERM'); } catch (error) { if (error.code !== 'ESRCH') throw error; }
  }
  let timer;
  try {
    // 只有本次子进程退出事件能够确认终止，未确认退出仍返回失败。
    await Promise.race([exited, new Promise((_, reject) => { timer = setTimeout(() => reject(shutdownError ?? new Error(`Child ${child.pid} did not exit`)), 5000); })]);
  } finally { clearTimeout(timer); child.removeListener('exit', onExit); }
}

export async function requireFile(relative) {
  await readFile(join(root, relative));
}
