// Capture a real human-reviewable paper snapshot from the built MV3 extension
// and save the raw JSON for later paper_snapshot ingest acceptance.
//
// This launches an isolated, throwaway Chromium profile (never the personal
// one) and loads the Astrocyte Paper extension; it then extracts snapshots from
// real public full-text and abstract-only pages using the extension's own
// injection files. The raw snapshot JSON is written verbatim to the output
// directory so it can be pasted into ImportMaterial(adapter=paper_snapshot)
// without any fixture substitution.
//
// Usage:
//   node scripts/paper-snapshot-capture.mjs [outputDir]
//
// Env:
//   ASTROCYTE_PAPER_SNAPSHOT_DIR  output directory (default: %TEMP%/astrocyte-paper-snapshot)
import { createRequire } from 'node:module';
import { readFile, mkdir, writeFile } from 'node:fs/promises';
import { dirname, resolve, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { existsSync } from 'node:fs';
import { execFileSync } from 'node:child_process';
import { tmpdir } from 'node:os';

const base = dirname(fileURLToPath(import.meta.url));
const repoRoot = resolve(base, '..');
const require = createRequire(resolve(repoRoot, 'web/package.json'));
const { chromium } = require('@playwright/test');

const dist = resolve(repoRoot, 'extensions/paper/dist');
if (!existsSync(resolve(dist, 'manifest.json'))) {
  execFileSync(process.execPath, [resolve(repoRoot, 'extensions/paper/build.mjs')], { stdio: 'inherit' });
}
const readability = await readFile(resolve(dist, 'Readability.js'), 'utf8');
const extract = await readFile(resolve(dist, 'extract.js'), 'utf8');

const outDir = process.argv[2] || process.env.ASTROCYTE_PAPER_SNAPSHOT_DIR || join(tmpdir(), 'astrocyte-paper-snapshot');
await mkdir(outDir, { recursive: true });

// Real public pages verified by W1; no credentials, no login.
const TARGETS = {
  fulltext: 'https://journals.plos.org/digitalhealth/article?id=10.1371/journal.pdig.0000514',
  abstract: 'https://aclanthology.org/2024.acl-long.1/',
};

const context = await chromium.launchPersistentContext('', {
  headless: false,
  args: [
    `--disable-extensions-except=${dist}`,
    `--load-extension=${dist}`,
    '--no-sandbox',
  ],
});

let extensionId = null;
for (let i = 0; i < 100 && !extensionId; i++) {
  for (const sw of context.serviceWorkers()) {
    const m = sw.url().match(/chrome-extension:\/\/([^/]+)\//);
    if (m) { extensionId = m[1]; break; }
  }
  if (!extensionId) await new Promise(r => setTimeout(r, 300));
}
if (!extensionId) {
  console.error('EXTENSION_LOAD_FAILED: no service worker registered');
  await context.close();
  process.exit(1);
}
console.log('EXTENSION_LOADED id=' + extensionId);

async function snapshot(url) {
  const page = await context.newPage();
  await page.goto(url, { waitUntil: 'domcontentloaded', timeout: 60000 });
  await page.addScriptTag({ content: readability });
  await page.addScriptTag({ content: extract });
  const snap = await page.evaluate((locator) => globalThis.astrocytePaperExtract(locator), url);
  await page.close();
  return snap;
}

let failed = false;
for (const [name, url] of Object.entries(TARGETS)) {
  const snap = await snapshot(url);
  const out = resolve(outDir, `${name}.snapshot.json`);
  await writeFile(out, JSON.stringify(snap, null, 2), 'utf8');
  console.log(`SAVED ${out} state=${snap.content_state} text_bytes=${(snap.text || '').length} source_key=${snap.source_key} pdf_urls=${(snap.pdf_urls || []).length}`);
  if (name === 'fulltext' && (snap.content_state !== 'readable_fulltext' || (snap.text || '').length < 10000)) failed = true;
  if (name === 'abstract' && (snap.content_state !== 'abstract_only' || (snap.text || '').length !== 0)) failed = true;
}
await context.close();
if (failed) {
  console.error('SNAPSHOT_WRONG: a real page did not yield the expected content_state');
  process.exit(2);
}
console.log('SNAPSHOT_OK output=' + outDir);
