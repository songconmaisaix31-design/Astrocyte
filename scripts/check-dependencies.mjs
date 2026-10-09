import { readFile } from 'node:fs/promises';
import { join } from 'node:path';
import { parse } from 'yaml';
import { execFileSync } from 'node:child_process';
import { dirname } from 'node:path';
import { root, web } from './process.mjs';
import { summarizeEnvironment, summarizeVersion } from './summarize.mjs';

const json = async path => JSON.parse(await readFile(path, 'utf8'));
const inventory = await json(join(root, 'dependencies.lock.json'));
const lock = parse(await readFile(join(root, 'pnpm-lock.yaml'), 'utf8'));
for (const [name, base, importer] of [['root', root, '.'], ['web', web, 'web']]) {
  const pkg = await json(join(base, 'package.json'));
  const expected = inventory.direct_dependencies[name];
  const actual = { ...pkg.dependencies, ...pkg.devDependencies };
  if (JSON.stringify(Object.entries(actual).sort()) !== JSON.stringify(Object.entries(expected).sort())) {
    throw new Error(`${name} dependency inventory drift; update dependencies.lock.json`);
  }
  for (const [dependency, version] of Object.entries(actual)) {
    if (!/^\d+\.\d+\.\d+$/.test(version)) throw new Error(`${dependency} must use an exact stable version`);
    const installed = await json(join(base, 'node_modules', dependency, 'package.json'));
    const entry = lock.importers[importer].dependencies?.[dependency] ?? lock.importers[importer].devDependencies?.[dependency];
    if (installed.version !== version || entry?.specifier !== version) throw new Error(`${dependency} installed/lock version differs from ${version}; run pnpm install --frozen-lockfile`);
  }
}
const goMod = await readFile(join(root, 'go.mod'), 'utf8');
if (!goMod.includes(`require modernc.org/sqlite ${inventory.direct_dependencies.go['modernc.org/sqlite']}`)) throw new Error('SQLite dependency inventory differs from go.mod');
if (!goMod.includes(`go ${inventory.toolchain.go}`)) throw new Error('Go toolchain inventory differs from go.mod');
const extraction = await summarizeEnvironment({});
const version = execFileSync(extraction.ASTROCYTE_NODE, [extraction.ASTROCYTE_SUMMARIZE_CLI, '--version'], {
  encoding: 'utf8', windowsHide: true, timeout: 60000,
  env: { PATH: dirname(extraction.ASTROCYTE_NODE), SystemRoot: process.env.SystemRoot, WINDIR: process.env.WINDIR },
}).trim();
if (version !== summarizeVersion) throw new Error(`Project summarize CLI reports ${version}, expected ${summarizeVersion}`);
console.log(`Project summarize CLI loads and reports ${version}; media extraction is checked separately`);
console.log('Exact dependency pins, installed versions and lock inventory agree');
