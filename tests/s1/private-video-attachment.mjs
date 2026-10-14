import assert from 'node:assert/strict';
import { randomBytes } from 'node:crypto';
import { backup, DatabaseSync } from 'node:sqlite';
import { constants } from 'node:fs';
import { copyFile, lstat, readFile, realpath } from 'node:fs/promises';
import { isAbsolute, join } from 'node:path';
import { startS1Server } from './server.mjs';
import { humanAPI } from './api.mjs';

const sourceRoot = process.env.ASTROCYTE_TEST_VIDEO_PUBLICATION;
if (!sourceRoot || !isAbsolute(sourceRoot)) throw new Error('An explicit approved actual video publication is required');
for (const [path, directory] of [[sourceRoot, true], [join(sourceRoot, 'state.sqlite'), false], [join(sourceRoot, 'objects'), true]]) {
  const info = await lstat(path);
  if (info.isSymbolicLink() || (directory ? !info.isDirectory() : !info.isFile()) || await realpath(path) !== path) throw new Error('Actual video source must use ordinary exact paths');
}
const original = new DatabaseSync(join(sourceRoot, 'state.sqlite'), { readOnly: true });
let server;
let primaryError;
try {
  const rows = original.prepare('SELECT data FROM attention_materials').all().map(row => JSON.parse(row.data));
  const material = rows.find(row => row.material.source_locator.includes('BV1PReT6EEqR'));
  assert.ok(material, 'actual approved video material missing');
  const revision = material.revisions.find(row => row.attachments.some(attachment => attachment.name === 'whisper-original.json'));
  assert.ok(revision, 'actual native attachment missing');
  const privateAttachment = revision.attachments.find(row => row.name === 'whisper-original.json');
  const objectBytes = await readFile(join(sourceRoot, 'objects', privateAttachment.object_ref));
  const native = JSON.parse(objectBytes);
  assert.ok(native.params && native.transcription.length > 0, 'actual native JSON envelope missing');
  const token = randomBytes(32).toString('hex');
  server = await startS1Server({ env: { ASTROCYTE_AGENT_TOKEN: token } });
  await server.stopAPI();
  await backup(original, join(server.dataDir, 'state.sqlite'));
  const references = new Set(material.revisions.flatMap(row => [row.object_ref, ...row.attachments.map(attachment => attachment.object_ref)]));
  for (const reference of references) {
    assert.match(reference, /^[a-f0-9]{64}$/, 'actual object reference format changed');
    const object = join(sourceRoot, 'objects', reference);
    assert.ok((await lstat(object)).isFile(), 'actual source object must be an ordinary file');
    await copyFile(object, join(server.dataDir, 'objects', reference), constants.COPYFILE_EXCL);
  }
  await server.startAPI();
  const before = server.query('SELECT data FROM attention_materials WHERE id=?', material.material.id)[0].data;
  for (const restart of [false, true]) {
    if (restart) await server.restart();
    const api = await humanAPI(server.apiURL);
    const detail = await api.get(`/materials/${material.material.id}`);
    assert.equal(detail.material.current_revision, material.material.current_revision);
    for (const pathName of ['whisper-original.json', '%77hisper-original.json']) {
      const path = `/materials/${material.material.id}/revisions/${revision.revision}/attachments/${pathName}`;
      for (const headers of [{}, { Authorization: `Bearer ${token}` }]) {
        const result = await api.request(path, { headers });
        assert.equal(result.status, 403, 'native object must be denied for actual human and agent identities');
        assert.equal(result.data.error.code, 'scope_denied');
      }
    }
    for (const attachment of revision.attachments.filter(row => row.name !== 'whisper-original.json')) {
      const response = await fetch(`${server.apiURL}/api/v1/materials/${material.material.id}/revisions/${revision.revision}/attachments/${encodeURIComponent(attachment.name)}`, { headers: { Cookie: api.cookie } });
      assert.equal(response.status, 200, 'public attachment semantics changed');
      assert.deepEqual(Buffer.from(await response.arrayBuffer()), await readFile(join(sourceRoot, 'objects', attachment.object_ref)));
    }
    assert.deepEqual(await readFile(join(server.dataDir, 'objects', privateAttachment.object_ref)), objectBytes, 'local immutable original changed');
    assert.equal(server.query('SELECT data FROM attention_materials WHERE id=?', material.material.id)[0].data, before, 'denied native reads recorded successful machine use');
  }
  console.log(JSON.stringify({ actual_material_id: material.material.id, revision: revision.revision, native_segments: native.transcription.length, private_http_denials: 8, public_attachment_reads: 6, new_asr_runs: 0, native_local_bytes_preserved: true, restart_verified: true }));
} catch (error) {
  primaryError = error;
  throw error;
} finally {
  original.close();
  if (server) await server.close({ primaryError });
}
