// Real paper_snapshot ingest acceptance: import human-reviewed browser-plugin
// snapshots through the existing ImportMaterial/jobs/objects chain and verify
// dedup and durable restart. No synthetic fallback: the snapshot JSON must be a
// genuine plugin capture (see scripts/paper-snapshot-capture.mjs), never a fixture.
//
// Usage:
//   node tests/s1/paper-snapshot-live.mjs --snapshot-dir <dir>
//
// The directory must contain fulltext.snapshot.json and abstract.snapshot.json.
import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import { join, resolve } from 'node:path';
import { startS1Server } from './server.mjs';
import { humanAPI, importMaterial, waitJob } from './api.mjs';

const dirIndex = process.argv.indexOf('--snapshot-dir');
if (dirIndex < 0 || !process.argv[dirIndex + 1]) throw new Error('Pass --snapshot-dir with a directory containing real fulltext/abstract snapshots');
const snapshotDir = resolve(process.argv[dirIndex + 1]);

async function loadSnapshot(name) {
  const snap = JSON.parse(await readFile(join(snapshotDir, name), 'utf8'));
  assert.equal(snap.schema_version, 1, `${name} must be a schema 1 paper snapshot`);
  assert.ok(snap.source_url && snap.source_url.startsWith('https://'), `${name} source_url must be public HTTPS`);
  return snap;
}

const fulltext = await loadSnapshot('fulltext.snapshot.json');
const abstract = await loadSnapshot('abstract.snapshot.json');

const server = await startS1Server();
try {
  const api = await humanAPI(server.apiURL);

  // 1. Real full-text snapshot: no network, no model; text and original JSON preserved.
  const full = await importMaterial(api, {
    adapter: 'paper_snapshot', kind: 'paper', source_locator: fulltext.source_url,
    source_key: '', content_digest: '', title: fulltext.title, export_text: JSON.stringify(fulltext),
  });
  assert.equal(full.detail.revisions[0].provenance.processor, 'paper_snapshot');
  assert.match(full.detail.revisions[0].provenance.mode, /^browser_snapshot/);
  assert.equal(full.detail.revisions[0].source_key, fulltext.source_key);
  const fullContent = await api.get(`/materials/${full.detail.material.id}/revisions/1/content`);
  assert.ok(fullContent.text.length > 10000, 'full-text snapshot must retain the real body');
  assert.ok(fullContent.text.includes(fulltext.text.slice(0, 200)), 'full-text body must be the captured text, not a summary');
  const snapshotAttachment = full.detail.revisions[0].attachments.find(a => a.name === 'paper-snapshot.json');
  assert.ok(snapshotAttachment, 'the original snapshot JSON must be kept as an attachment');
  const kept = JSON.parse(await readFile(join(server.dataDir, 'objects', snapshotAttachment.object_ref), 'utf8'));
  assert.deepEqual(kept, fulltext, 'the stored snapshot must be byte-equivalent to the reviewed JSON');

  // 2. Real abstract-only snapshot: metadata only, no captured body impersonation.
  const abs = await importMaterial(api, {
    adapter: 'paper_snapshot', kind: 'paper', source_locator: abstract.source_url,
    source_key: '', content_digest: '', title: abstract.title, export_text: JSON.stringify(abstract),
  });
  assert.equal(abs.detail.revisions[0].provenance.mode, 'browser_snapshot_abstract_only');
  assert.equal(abs.detail.revisions[0].source_key, abstract.source_key);
  assert.notEqual(abs.detail.material.id, full.detail.material.id, 'different papers must be different materials');
  const absContent = await api.get(`/materials/${abs.detail.material.id}/revisions/1/content`);
  assert.ok(!absContent.text.includes('Full text'), 'an abstract-only snapshot must never present metadata as full text');

  // 3. Dedup: re-importing the same full-text snapshot reuses the same material.
  const again = await importMaterial(api, {
    adapter: 'paper_snapshot', kind: 'paper', source_locator: fulltext.source_url,
    source_key: '', content_digest: '', title: fulltext.title, export_text: JSON.stringify(fulltext),
  });
  assert.equal(again.detail.material.id, full.detail.material.id, 'same source_key must reuse the same material');

  // 4. Durable restart: fresh API process re-reads the same SQLite/objects.
  await server.restart();
  const refetched = await humanAPI(server.apiURL);
  const refetchedContent = await refetched.get(`/materials/${full.detail.material.id}/revisions/1/content`);
  assert.equal(refetchedContent.content_digest, fullContent.content_digest, 'full-text content must survive an actual restart');
  assert.equal((await refetched.get(`/materials/${abs.detail.material.id}/revisions/1/content`)).content_digest, absContent.content_digest);

  console.log(`PAPER_SNAPSHOT_OK fulltext=${full.detail.material.id} abstract=${abs.detail.material.id}`);
} finally {
  await server.close();
}
