import { readFile, access, realpath } from 'node:fs/promises';
import { join, isAbsolute } from 'node:path';
import { root } from './process.mjs';

export const summarizeVersion = '0.25.1';

// Resolve the locked project installation, never a global PATH command.
export async function summarizeEnvironment(env = process.env) {
  const enabled = env.ASTROCYTE_ENABLE_SUMMARIZE ?? '';
  if (enabled === 'false') return { ...env };
  if (enabled !== '' && enabled !== 'true') throw new Error('ASTROCYTE_ENABLE_SUMMARIZE must be true or false');
  const pkgRoot = join(root, 'node_modules', '@steipete', 'summarize');
  const pkg = JSON.parse(await readFile(join(pkgRoot, 'package.json'), 'utf8').catch(() => {
    throw new Error('Project summarize is missing; run pnpm install --frozen-lockfile');
  }));
  if (pkg.version !== summarizeVersion) throw new Error(`Project summarize must be ${summarizeVersion}; run pnpm install --frozen-lockfile`);
  const cli = env.ASTROCYTE_SUMMARIZE_CLI || await realpath(join(pkgRoot, pkg.bin.summarize));
  const node = env.ASTROCYTE_NODE || process.execPath;
  for (const path of [cli, node]) {
    if (!isAbsolute(path)) throw new Error('Summarize CLI and Node paths must be absolute');
    await access(path);
  }
  return { ...env, ASTROCYTE_SUMMARIZE_CLI: cli, ASTROCYTE_NODE: node };
}
