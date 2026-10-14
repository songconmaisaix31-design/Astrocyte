import { spawnSync } from 'node:child_process';
import { mkdir, readFile, realpath, writeFile } from 'node:fs/promises';
import { isAbsolute, join, relative, resolve, sep } from 'node:path';
import { root } from './process.mjs';

// 通过安装的 CLI 取得公开来源原件，所有输出保留在工作树验证目录。
const directory = resolve(root, process.argv[2] ?? '.verification/public-inputs');
const inside = relative(join(root, '.verification'), directory);
if (!inside || inside === '..' || inside.startsWith(`..${sep}`) || isAbsolute(inside)) throw new Error('Public input output must stay inside this worktree verification directory');
await mkdir(directory);
const home = join(directory, 'native-home');
await mkdir(home, { recursive: true });
const cli = await realpath(join(root, 'node_modules', '@steipete', 'summarize', 'dist', 'cli.js'));
const environment = {
  HOME: home, USERPROFILE: home, TEMP: home, TMP: home,
  XDG_CONFIG_HOME: home, XDG_CACHE_HOME: home, APPDATA: home, LOCALAPPDATA: home,
  PATH: join(process.execPath, '..'), NO_COLOR: '1', SUMMARIZE_DISABLE_LOCAL_WHISPER_CPP: '1',
};
for (const key of ['SystemRoot', 'WINDIR', 'COMSPEC']) if (process.env[key]) environment[key] = process.env[key];
const version = spawnSync(process.execPath, [cli, '--version'], { env: environment, cwd: home, encoding: 'utf8', windowsHide: true });
if (version.error) throw version.error;
if (version.status !== 0 || version.stdout.trim() !== '0.25.1') throw new Error('Installed summarize 0.25.1 is required');
const sources = [
  { name: 'paper-v1', url: 'https://arxiv.org/html/2501.12948v1' },
  { name: 'paper-v2', url: 'https://arxiv.org/html/2501.12948v2' },
  { name: 'youtube', url: 'https://www.youtube.com/watch?v=aircAruvnKk' },
];
const selected = process.argv[3];
if (selected && !sources.some(input => input.name === selected)) throw new Error('Select paper-v1, paper-v2 or youtube');
for (const input of sources.filter(input => !selected || input.name === selected)) {
  const html = await fetch(input.url, { signal: AbortSignal.timeout(45_000) });
  if (!html.ok) throw new Error(`Public source ${input.url} returned ${html.status}`);
  await writeFile(join(directory, `${input.name}.html`), Buffer.from(await html.arrayBuffer()));
  const result = spawnSync(process.execPath, [cli, input.url, '--extract', '--json', '--format', 'text', '--firecrawl', 'off', '--youtube', 'web', '--video-mode', 'transcript', '--embedded-video', 'off', '--timestamps', '--timeout', '30s', '--retries', '0', '--metrics', 'off'], {
    env: environment, cwd: home, encoding: 'utf8', windowsHide: true, maxBuffer: 16 << 20, timeout: 60_000,
  });
  await writeFile(join(directory, `${input.name}.json`), result.stdout ?? '');
  await writeFile(join(directory, `${input.name}.stderr.txt`), result.stderr ?? '');
  if (result.error) throw result.error;
  if (result.status !== 0) throw new Error(`Public extraction ${input.name} exited with ${result.status}`);
  const actual = JSON.parse(await readFile(join(directory, `${input.name}.json`), 'utf8'));
  if (actual.input?.url !== input.url || actual.llm !== null || !actual.extracted?.content?.trim()) throw new Error(`Public extraction ${input.name} lacks matching original text`);
  if (input.name === 'youtube' && (!actual.extracted.transcriptSource || !actual.extracted.transcriptSegments?.length)) throw new Error('Actual YouTube caption timestamps are unavailable');
  console.log(JSON.stringify({ source: input.url, bytes: Buffer.byteLength(result.stdout), text_chars: actual.extracted.content.length, segments: actual.extracted.transcriptSegments?.length ?? 0 }));
}
