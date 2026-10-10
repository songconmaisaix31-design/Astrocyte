// Stage standalone parser assets, without choosing final extension permissions.
// The final MV3 manifest/action/API coupling is gated on root product decisions.
import { createRequire, findPackageJSON } from 'node:module';
import { readFile, realpath, mkdir, copyFile, writeFile } from 'node:fs/promises';
import { dirname, resolve } from 'node:path';
import { pathToFileURL, fileURLToPath } from 'node:url';

const base = dirname(fileURLToPath(import.meta.url));
const cli = await realpath(resolve(base, '../../node_modules/@steipete/summarize/dist/cli.js'));
const core = findPackageJSON('@steipete/summarize-core', pathToFileURL(cli));
const corePackage = JSON.parse(await readFile(core, 'utf8'));
if (corePackage.version !== '0.25.1') throw new Error('summarize core pin differs from 0.25.1');
const require = createRequire(core);
const readability = require.resolve('@mozilla/readability');
const readabilityPackage = JSON.parse(await readFile(resolve(dirname(readability), 'package.json'), 'utf8'));
if (readabilityPackage.version !== '0.6.0') throw new Error('Readability pin differs from 0.6.0');
const output = resolve(base, 'dist');
await mkdir(output, { recursive: true });
await copyFile(resolve(dirname(readability), 'Readability.js'), resolve(output, 'Readability.js'));
await copyFile(resolve(base, '../../internal/adapters/importers/paper-dom.mjs'), resolve(output, 'paper-dom.mjs'));
await copyFile(resolve(base, 'vendor/readability-LICENSE.md'), resolve(output, 'readability-LICENSE.md'));
await copyFile(resolve(base, 'vendor/summarize-LICENSE'), resolve(output, 'summarize-LICENSE'));
await copyFile(resolve(base, 'THIRD_PARTY.md'), resolve(output, 'THIRD_PARTY.md'));
await writeFile(resolve(output, 'README.txt'), 'Standalone extraction assets built. Final extension action/permissions and application identity remain pending; this directory is not an installable extension. No provider credentials or automatic browsing.\n');
console.log(`Staged parser assets: ${output}`);
