import { spawn, spawnSync } from 'node:child_process';
import { access, readFile, readdir } from 'node:fs/promises';
import { constants } from 'node:fs';
import { delimiter, join } from 'node:path';
import { fileURLToPath } from 'node:url';

export const root = fileURLToPath(new URL('../', import.meta.url));
export const web = join(root, 'web');

export async function findGo() {
  const windows = process.platform === 'win32';
  const fromPath = (process.env.PATH ?? '').split(delimiter).filter(Boolean)
    .map(dir => join(dir.replace(/^"|"$/g, ''), windows ? 'go.exe' : 'go'));
  const fallback = windows && process.env.LOCALAPPDATA
    ? join(process.env.LOCALAPPDATA, 'Programs', 'go', 'bin', 'go.exe') : undefined;
  const candidates = process.env.ASTROCYTE_GO ? [process.env.ASTROCYTE_GO] : [...fromPath, fallback].filter(Boolean);
  for (const candidate of candidates) {
    try { await access(candidate, windows ? constants.F_OK : constants.X_OK); return candidate; } catch { /* Try next location. */ }
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
    if (['.git', 'node_modules', 'dist'].includes(entry.name)) continue;
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
  const result = spawnSync(executable, ['-l', ...files], { cwd: root, encoding: 'utf8', windowsHide: true });
  if (result.error) throw result.error;
  if (result.status !== 0 || result.stdout.trim()) throw new Error(`gofmt check failed:\n${result.stdout}${result.stderr}`);
  console.log('gofmt: all Go files formatted');
}

/** Stop only children created by this invocation, including their tool subprocesses. */
export async function stop(child) {
  if (!child.pid || child.exitCode !== null || child.signalCode !== null) return;
  // Dispatch any queued exit before using this invocation's child PID. Windows
  // taskkill is synchronous, so exitCode cannot update while it is running.
  await new Promise(resolve => setImmediate(resolve));
  if (child.exitCode !== null || child.signalCode !== null) return;
  let onExit;
  const exited = new Promise(resolve => { onExit = resolve; child.once('exit', onExit); });
  let shutdownError;
  if (process.platform === 'win32') {
    const result = spawnSync('taskkill.exe', ['/pid', String(child.pid), '/T', '/F'], { encoding: 'utf8', windowsHide: true });
    if (result.status !== 0) shutdownError = new Error(`Child shutdown failed: ${result.stderr || result.stdout || result.error}`);
  } else {
    // Long-lived children are launched as process group leaders on POSIX.
    try { process.kill(-child.pid, 'SIGTERM'); } catch (error) { if (error.code !== 'ESRCH') throw error; }
  }
  let timer;
  try {
    // A nonzero taskkill result is harmless only after this owned child's exit
    // event confirms termination. A still-live/unconfirmed child remains failure.
    await Promise.race([exited, new Promise((_, reject) => { timer = setTimeout(() => reject(shutdownError ?? new Error(`Child ${child.pid} did not exit`)), 5000); })]);
  } finally { clearTimeout(timer); child.removeListener('exit', onExit); }
}

export async function requireFile(relative) {
  try { await readFile(join(root, relative)); } catch { throw new Error(`Missing ${relative}; downstream S0 worker files are pending`); }
}
