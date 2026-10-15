import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import { join, resolve } from 'node:path';
import { loadPublicPaperSnapshots } from './public-inputs.mjs';
import { startS1Server } from './server.mjs';
import { humanAPI, importMaterial } from './api.mjs';

const snapshots = await loadPublicPaperSnapshots(resolve(process.argv[2]));
const server = await startS1Server();
let primaryError;
try {
  const api = await humanAPI(server.apiURL);
  for (const name of ['secondary', 'abstract']) {
    const snapshot = snapshots[name];
    const original = JSON.stringify(snapshot);
    const imported = await importMaterial(api, {
      kind: 'paper', adapter: 'paper_snapshot', source_locator: snapshot.source_url,
      source_key: '', content_digest: '', export_text: original, collection_reason: null,
    });
    assert.equal(imported.detail.revisions[0].source_key, snapshot.source_key);
    const revision = imported.detail.revisions[0];
    assert.equal(revision.provenance.mode, name === 'secondary' ? 'browser_snapshot_fulltext' : 'browser_snapshot_abstract_only');
    const content = await api.get(`/materials/${imported.detail.material.id}/revisions/1/content`);
    const expected = name === 'secondary'
      ? `# ${snapshot.title}\n\nSource: ${snapshot.source_url}\n\n## Abstract\n\n${snapshot.abstract}\n\n## Full text (browser snapshot)\n\n${snapshot.text}`
      : `# ${snapshot.title}\n\nSource: ${snapshot.source_url}\n\n## Abstract (metadata only)\n\n${snapshot.abstract}`;
    assert.ok(content.text === expected, 'Actual DOM snapshot import must preserve the complete source observation');
    const attachment = revision.attachments.find(item => item.name === 'paper-snapshot.json');
    assert.ok(attachment, 'Actual snapshot original attachment must exist');
    assert.ok(await readFile(join(server.dataDir, 'objects', attachment.object_ref), 'utf8') === original, 'Actual DOM snapshot original bytes changed');
    console.log(JSON.stringify({ name, source_key: snapshot.source_key, schema_version: snapshot.schema_version, content_state: snapshot.content_state, original_bytes: Buffer.byteLength(original), imported_text_bytes: Buffer.byteLength(content.text), material_id: imported.detail.material.id }));
  }
} catch (error) {
  primaryError = error;
  throw error;
} finally {
  await server.close({ primaryError });
}
