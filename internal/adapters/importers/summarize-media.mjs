// 使用固定版本上游接口完成字幕提取与本地转写。
import { findPackageJSON } from 'node:module';
import { readFile } from 'node:fs/promises';
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
 const { spawnTracked, terminateTrackedProcesses } = await import(pathToFileURL(resolve(dirname(coreEntry), 'processes.js')).href);
 terminate = () => terminateTrackedProcesses('SIGTERM');
 if (cancelled) throw new Error('Cancelled');
 const captionEnv = { ...process.env, SUMMARIZE_DISABLE_LOCAL_WHISPER_CPP: '1' };
 delete captionEnv.YT_DLP_PATH;
 // 先查询字幕，隔离 HOME 不包含配置或凭据。
 const args = [cli, url, '--extract', '--json', '--format', 'text', '--firecrawl', 'off', '--youtube', 'web', '--video-mode', 'transcript', '--embedded-video', 'off', '--timestamps', '--timeout', '30s', '--retries', '0', '--metrics', 'off'];
 const { proc } = spawnTracked(process.execPath, args, { env: captionEnv, stdio: ['ignore', 'pipe', 'pipe'], windowsHide: true, label: 'summarize captions' });
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
 if (!hasCaption) {
  for (const key of ['YT_DLP_PATH', 'FFMPEG_PATH', 'SUMMARIZE_WHISPER_CPP_BINARY', 'SUMMARIZE_WHISPER_CPP_MODEL_PATH']) {
   if (!process.env[key]) throw new Error(`No real video captions; configure local dependency ${key}`);
  }
  // 上游明确返回无字幕时，按已确认流程转写同一视频。
  const { fetchTranscriptWithYtDlp } = await import(pathToFileURL(resolve(dirname(coreEntry), 'content/transcript/providers/youtube/yt-dlp.js')).href);
  const result = await fetchTranscriptWithYtDlp({ url, ytDlpPath: process.env.YT_DLP_PATH, env: process.env, service: 'generic', mediaKind: 'video', extraArgs: ['--ignore-config', '--no-playlist', '--playlist-items', '1', '--ffmpeg-location', dirname(process.env.FFMPEG_PATH)] });
  output.media = { ...result, error: result.error ? result.error.message : null };
  if (result.error) throw result.error;
 }
} catch (error) {
 output.error = error instanceof Error ? error.message : String(error);
 process.exitCode = cancelled ? 130 : 1;
} finally {
 terminate();
 process.stdin.pause();
 process.stdout.write(JSON.stringify(output));
}
