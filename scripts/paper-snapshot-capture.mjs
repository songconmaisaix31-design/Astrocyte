// 使用自有 Chromium profile 和已构建 MV3 提取实际公开论文，原 JSON 保存在本树验证目录。
// 用法：
//   node scripts/paper-snapshot-capture.mjs [outputDir]
// ASTROCYTE_PAPER_SNAPSHOT_DIR 选择本树 .verification 内的专属输出目录。
import { createRequire } from 'node:module';
import { readFile, mkdir, writeFile, realpath, lstat } from 'node:fs/promises';
import { dirname, resolve, join, relative, isAbsolute, sep } from 'node:path';
import { fileURLToPath } from 'node:url';
import { existsSync } from 'node:fs';
import { execFileSync } from 'node:child_process';

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

const verification = await realpath(join(repoRoot, '.verification'));
const outDir = resolve(process.argv[2] || process.env.ASTROCYTE_PAPER_SNAPSHOT_DIR || join(verification, 'paper-snapshot'));
const inside = relative(verification, outDir);
if (!inside || inside === '..' || inside.startsWith(`..${sep}`) || isAbsolute(inside)) throw new Error('Paper snapshot output must stay in a dedicated worktree verification directory');
const parent = await realpath(dirname(outDir));
const parentInside = relative(verification, parent);
if (parentInside === '..' || parentInside.startsWith(`..${sep}`) || isAbsolute(parentInside)) throw new Error('Paper snapshot output parent redirects outside verification');
try {
  const info = await lstat(outDir);
  if (!info.isDirectory() || info.isSymbolicLink()) throw new Error('Paper snapshot output must be an owned ordinary directory');
} catch (error) {
  if (error.code !== 'ENOENT') throw error;
  await mkdir(outDir);
}
const canonical = await realpath(outDir);
if (process.platform === 'win32' ? canonical.toLowerCase() !== outDir.toLowerCase() : canonical !== outDir) throw new Error('Paper snapshot output directory redirects');

// 本任务已经完整读取的公开出版页面。
const TARGETS = {
  fulltext: 'https://journals.plos.org/digitalhealth/article?id=10.1371/journal.pdig.0000514',
  abstract: 'https://aclanthology.org/2024.acl-long.1/',
};

const context = await chromium.launchPersistentContext(join(outDir, 'browser-profile'), {
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
