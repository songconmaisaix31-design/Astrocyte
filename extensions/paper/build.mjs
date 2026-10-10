// Build a loadable Manifest V3 extension under extensions/paper/dist.
// It reuses the pinned installed summarize core to resolve the vendored
// Readability 0.6.0 (Apache-2.0) and copies the browser-native extraction
// assets. No new package, lockfile or build framework is introduced; no
// provider credentials, automatic browsing or daemon are shipped.
import { createRequire, findPackageJSON } from 'node:module';
import { readFile, realpath, mkdir, copyFile, writeFile, rm } from 'node:fs/promises';
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
await rm(output, { recursive: true, force: true });
await mkdir(output, { recursive: true });
await copyFile(resolve(dirname(readability), 'Readability.js'), resolve(output, 'Readability.js'));
for (const name of ['manifest.json', 'background.js', 'extract.js', 'popup.html', 'popup.js']) {
  await copyFile(resolve(base, name), resolve(output, name));
}
await copyFile(resolve(base, 'vendor/readability-LICENSE.md'), resolve(output, 'readability-LICENSE.md'));
await copyFile(resolve(base, 'vendor/Apache-2.0.txt'), resolve(output, 'Apache-2.0.txt'));
await copyFile(resolve(base, 'vendor/summarize-LICENSE'), resolve(output, 'summarize-LICENSE'));
await copyFile(resolve(base, 'THIRD_PARTY.md'), resolve(output, 'THIRD_PARTY.md'));
await writeFile(resolve(output, 'README.txt'),
  'Astrocyte Paper (MV3). Load this directory as an unpacked extension. ' +
  'Extraction runs only on the human-clicked current page (activeTab). ' +
  'The same-origin application import transport is owned by W0 and is not shipped here; ' +
  'the popup copies a metadata/fulltext snapshot to the clipboard for human review. ' +
  'No provider credentials, automatic tab reading or daemon.\n');
console.log(`Built loadable MV3 extension: ${output}`);
