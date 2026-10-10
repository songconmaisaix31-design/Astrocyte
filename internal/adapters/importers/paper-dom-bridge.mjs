// Reuse exact installed summarize parser dependencies; no new package/lockfile.
import { findPackageJSON, createRequire } from 'node:module';
import { realpathSync } from 'node:fs';
import { readFile } from 'node:fs/promises';
import { dirname, resolve } from 'node:path';
import { pathToFileURL } from 'node:url';
import { paperDOM } from './paper-dom.mjs';

const [cliArg, locator] = process.argv.slice(2);
const cli = realpathSync(cliArg);
const cliPackage = JSON.parse(await readFile(resolve(dirname(cli), '..', 'package.json'), 'utf8'));
const coreManifest = findPackageJSON('@steipete/summarize-core', pathToFileURL(cli));
const corePackage = JSON.parse(await readFile(coreManifest, 'utf8'));
if (cliPackage.version !== '0.25.1' || corePackage.version !== '0.25.1') throw new Error('summarize pin mismatch');
const require = createRequire(coreManifest);
if (require('@mozilla/readability/package.json').version !== '0.6.0') throw new Error('Readability pin mismatch');
const { Readability } = require('@mozilla/readability');
const { parseHTML } = require('linkedom');
let html = '';
process.stdin.setEncoding('utf8');
for await (const part of process.stdin) {
  html += part;
  if (Buffer.byteLength(html) > 16 * 1024 * 1024) throw new Error('paper HTML exceeds 16MiB');
}
const { document } = parseHTML(html);
process.stdout.write(JSON.stringify(paperDOM(document, locator, Readability)));
