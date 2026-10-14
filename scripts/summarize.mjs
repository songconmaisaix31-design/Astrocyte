import { readFile, access, realpath, stat } from 'node:fs/promises';
import { join, isAbsolute } from 'node:path';
import { root } from './process.mjs';

export const summarizeVersion = '0.25.1';

// 使用项目固定版本的实际安装目录。
export async function summarizeEnvironment(env = process.env) {
  const enabled = env.ASTROCYTE_ENABLE_SUMMARIZE ?? '';
  if (enabled === 'false') return { ...env };
  if (enabled !== '' && enabled !== 'true') throw new Error('ASTROCYTE_ENABLE_SUMMARIZE must be true or false');
  const pkgRoot = join(root, 'node_modules', '@steipete', 'summarize');
  const pkg = JSON.parse(await readFile(join(pkgRoot, 'package.json'), 'utf8'));
  if (pkg.version !== summarizeVersion) throw new Error(`Project summarize must be ${summarizeVersion}; run pnpm install --frozen-lockfile`);
  const cliLocation = env.ASTROCYTE_SUMMARIZE_CLI || join(pkgRoot, pkg.bin.summarize);
  if (!isAbsolute(cliLocation)) throw new Error('Summarize CLI and Node paths must be absolute');
  const cli = await realpath(cliLocation);
  const node = env.ASTROCYTE_NODE || process.execPath;
  for (const path of [cli, node]) {
    if (!isAbsolute(path)) throw new Error('Summarize CLI and Node paths must be absolute');
    await access(path);
  }
  const configured = { ...env, ASTROCYTE_SUMMARIZE_CLI: cli, ASTROCYTE_NODE: node };
  const mediaRoot = env.ASTROCYTE_MEDIA_DIR || (process.platform === 'win32' && env.LOCALAPPDATA ? join(env.LOCALAPPDATA, 'Astrocyte', 'media') : undefined);
  if (mediaRoot) {
    if (!isAbsolute(mediaRoot)) throw new Error('ASTROCYTE_MEDIA_DIR must be absolute');
    const pins = JSON.parse(await readFile(join(root, 'dependencies.lock.json'), 'utf8')).external_media;
    for (const [key, relative] of [
      ['ASTROCYTE_YT_DLP_PATH', join(pins.yt_dlp.directory, pins.yt_dlp.executable)],
      ['ASTROCYTE_LISTING_PYTHON', join(pins.yt_dlp.directory, 'Scripts', 'python.exe')],
      ['ASTROCYTE_FFMPEG_PATH', join(pins.ffmpeg.directory, pins.ffmpeg.executable)],
      ['ASTROCYTE_WHISPER_BINARY', join(pins.whisper.directory, pins.whisper.executable)],
      ['ASTROCYTE_WHISPER_MODEL', pins.model.file],
    ]) {
      const candidate = configured[key] || join(mediaRoot, relative);
      if (!isAbsolute(candidate)) throw new Error(`${key} must be absolute`);
      let info;
      try { info = await stat(candidate); }
      catch (error) {
        if (!configured[key] && !env.ASTROCYTE_MEDIA_DIR && error.code === 'ENOENT') continue;
        throw error;
      }
      if (!info.isFile()) throw new Error(`${key} must select a regular file`);
      configured[key] = candidate;
    }
  }
  return configured;
}
