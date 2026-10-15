// 使用固定版本上游接口完成字幕提取与本地转写。
import { findPackageJSON } from 'node:module';
import { readFile, stat } from 'node:fs/promises';
import { dirname, resolve } from 'node:path';
import { pathToFileURL } from 'node:url';

const [cli, url] = process.argv.slice(2);
const output = { cliJSON: '', cliStderr: '', cliExit: -1, media: null, error: null };
let terminate = () => {};
let cancelled = false;
process.stdin.on('end', () => { cancelled = true; terminate(); process.exitCode = 130; });
process.stdin.resume();
try {
 const cliPackage = JSON.parse(await readFile(resolve(dirname(cli), '..', 'package.json'), 'utf8'));
 if (cliPackage.name !== '@steipete/summarize' || cliPackage.version !== '0.25.1') throw new Error('Pinned summarize CLI package mismatch');
 const coreManifest = findPackageJSON('@steipete/summarize-core', pathToFileURL(cli));
 if (!coreManifest) throw new Error('Pinned summarize core package unavailable');
 const coreRoot = dirname(coreManifest);
 const coreEntry = resolve(coreRoot, 'dist/esm/index.js');
 const corePackage = JSON.parse(await readFile(coreManifest, 'utf8'));
 if (corePackage.name !== '@steipete/summarize-core' || corePackage.version !== '0.25.1') throw new Error('Pinned summarize core package mismatch');
 const { assertNetworkTargetAllowed } = await import(pathToFileURL(resolve(dirname(coreEntry), 'content/network-guard.js')).href);
 await assertNetworkTargetAllowed(url, { targetLabel: 'selected public video' });
 const { spawnTracked, terminateTrackedProcesses, setProcessObserver } = await import(pathToFileURL(resolve(dirname(coreEntry), 'processes.js')).href);
 // 原观察器只转发实际进程标识和退出；环境、命令参数与原始输出保持本机私有。
 setProcessObserver({ register(info) {
  if (!['summarize', 'yt-dlp', 'ffmpeg', 'ffprobe', 'whisper.cpp'].includes(info.kind)) return null;
  let pid = null;
  const emit = (phase, fields = {}) => {
   if (pid !== null) process.stderr.write(`${JSON.stringify({ kind: info.kind, phase, pid, ...fields })}\n`);
  };
  return {
   setPid(actualPid) { pid = actualPid; emit('started'); },
   setProgress(progress) { emit('progress', { progress }); },
   appendOutput() {},
   finish(result) { emit('exited', { exitCode: result.exitCode }); },
  };
 } });
 terminate = () => terminateTrackedProcesses('SIGTERM');
 if (cancelled) throw new Error('Cancelled');
 const captionEnv = { ...process.env, SUMMARIZE_DISABLE_LOCAL_WHISPER_CPP: '1' };
 delete captionEnv.YT_DLP_PATH;
 // 先查询字幕，隔离 HOME 不包含配置或凭据。
 const args = [cli, url, '--extract', '--json', '--format', 'text', '--firecrawl', 'off', '--youtube', 'web', '--video-mode', 'transcript', '--embedded-video', 'off', '--timestamps', '--timeout', '30s', '--retries', '0', '--metrics', 'off'];
 const { proc } = spawnTracked(process.execPath, args, { env: captionEnv, stdio: ['ignore', 'pipe', 'pipe'], windowsHide: true, label: 'summarize captions', kind: 'summarize', captureOutput: false });
 const limit = 16 << 20;
 const collect = (stream, key) => stream.on('data', chunk => {
  if (Buffer.byteLength(output[key]) + Buffer.byteLength(chunk) > limit) { output.error = 'summarize output exceeds size limit'; terminate(); }
  else output[key] += chunk.toString('utf8');
 });
 // setEncoding 处理跨 chunk 的中文多字节字符。
 proc.stdout.setEncoding('utf8'); proc.stderr.setEncoding('utf8');
 collect(proc.stdout, 'cliJSON'); collect(proc.stderr, 'cliStderr');
 output.cliExit = await new Promise((done, reject) => { proc.once('error', reject); proc.once('close', code => done(code ?? -1)); });
 if (cancelled) throw new Error('Cancelled');
 if (output.error) throw new Error(output.error);
 if (output.cliExit !== 0) throw new Error(`summarize captions exited with ${output.cliExit}`);
 const raw = JSON.parse(output.cliJSON);
 const hasCaption = raw?.extracted?.transcriptSource && raw?.extracted?.diagnostics?.transcript?.textProvided !== false && raw?.extracted?.content?.trim();
 if (hasCaption && !raw.extracted.transcriptSegments?.length) throw new Error('Published transcript has no verified timestamp segments');
 if (!hasCaption) {
  for (const key of ['YT_DLP_PATH', 'FFMPEG_PATH', 'SUMMARIZE_WHISPER_CPP_BINARY', 'SUMMARIZE_WHISPER_CPP_MODEL_PATH']) {
   if (!process.env[key]) throw new Error(`No real video captions; configure local dependency ${key}`);
  }
  // 上游明确返回无字幕时，按已确认流程转写同一视频。
  const { acquireYtDlpMedia } = await import(pathToFileURL(resolve(dirname(coreEntry), 'content/transcript/providers/youtube/yt-dlp-media.js')).href);
  const acquired = await acquireYtDlpMedia({ url, ytDlpPath: process.env.YT_DLP_PATH, service: 'generic', mediaKind: 'video', downloadVideo: false, extraArgs: ['--ignore-config', '--no-playlist', '--playlist-items', '1', '--ffmpeg-location', dirname(process.env.FFMPEG_PATH)] });
  try {
   if (cancelled) throw new Error('Cancelled');
   const base = resolve(process.cwd(), 'whisper-original');
   // 固定上游纯文本转写缺少时间；同一原生程序一次生成正文与 JSON 时间。
   const whisperArgs = ['--model', process.env.SUMMARIZE_WHISPER_CPP_MODEL_PATH, '--language', 'auto', '-mc', '0', '--no-prints', '--output-json', '--output-txt', '--output-file', base, acquired.filePath];
   const { proc: whisper } = spawnTracked(process.env.SUMMARIZE_WHISPER_CPP_BINARY, whisperArgs, { env: process.env, stdio: ['ignore', 'ignore', 'pipe'], windowsHide: true, label: 'whisper.cpp timed transcript', kind: 'whisper.cpp', captureOutput: false });
   let detail = '';
   whisper.stderr.setEncoding('utf8');
   whisper.stderr.on('data', chunk => {
    if (Buffer.byteLength(detail) + Buffer.byteLength(chunk) > limit) { output.error = 'whisper.cpp diagnostics exceed size limit'; terminate(); }
    else detail += chunk;
   });
   const exit = await new Promise((done, reject) => { whisper.once('error', reject); whisper.once('close', code => done(code ?? -1)); });
   if (cancelled) throw new Error('Cancelled');
   if (output.error) throw new Error(output.error);
   if (exit !== 0) throw new Error(`whisper.cpp timed transcript exited with ${exit}`);
   for (const path of [`${base}.json`, `${base}.txt`]) {
    const info = await stat(path);
    if (!info.isFile() || info.size > limit) throw new Error('whisper.cpp output exceeds size limit or is not a regular file');
   }
   const native = JSON.parse(await readFile(`${base}.json`, 'utf8'));
   const text = (await readFile(`${base}.txt`, 'utf8')).trim();
   if (!text || !Array.isArray(native.transcription) || !native.transcription.length) throw new Error('whisper.cpp returned no timed transcript');
   let previousEnd = 0;
   const segments = native.transcription.map(segment => {
    const startMs = segment?.offsets?.from;
    const endMs = segment?.offsets?.to;
    if (!Number.isSafeInteger(startMs) || !Number.isSafeInteger(endMs) || startMs < previousEnd || endMs < startMs || typeof segment.text !== 'string' || !segment.text.trim() || !text.includes(segment.text.trim())) throw new Error('Invalid native whisper.cpp millisecond offsets or transcript text');
    previousEnd = endMs;
    return { startMs, endMs, text: segment.text.trim() };
   });
   // 原生 JSON 保留在本机文件，由 Go 读取为独立原件附件。
   output.media = { text, provider: 'whisper.cpp', error: null, notes: ['native whisper.cpp JSON offsets in milliseconds'], segments };
  } finally {
   await acquired.cleanup();
  }
 }
} catch (error) {
 output.error = error instanceof Error ? error.message : String(error);
 process.exitCode = cancelled ? 130 : 1;
} finally {
 terminate();
 process.stdin.pause();
 process.stdout.write(JSON.stringify(output));
}
