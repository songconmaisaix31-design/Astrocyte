// Real project progress HTTP acceptance: human Set / unknown Get / absent
// explicit null / durable restart / unavailable-infer error. No fixture status,
// no fake percent, no synthetic progress. Registers an isolated local project
// (throwaway temp root, no personal history) through the real HTTP stack.
//
// Usage:
//   node tests/s1/progress-live.mjs
import assert from 'node:assert/strict';
import { mkdtemp, writeFile, readFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { randomUUID } from 'node:crypto';
import { startS1Server } from './server.mjs';
import { humanAPI } from './api.mjs';

const root = await mkdtemp(join(tmpdir(), 'astrocyte-progress-root-'));
await writeFile(join(root, 'TASK.md'), '# Real progress task\n\nStage: contract convergence.\n', 'utf8');
await writeFile(join(root, 'STATUS.md'), '# Real progress status\n\nStatus: in progress.\n', 'utf8');

const server = await startS1Server();
try {
  let api = await humanAPI(server.apiURL);

  const space = await api.write('/project-spaces', { title: 'progress-acceptance-space' }, { status: 201 });
  const project = await api.write('/local-projects', { name: 'progress-acceptance', root, space_id: space.space.id });
  const id = project.project.id;

  // 1. Unknown/absent Get: explicit unknown status, null observed_at/source, no fake timestamp.
  const absent = await api.get(`/local-projects/${id}/progress`);
  assert.equal(absent.progress.status, 'unknown', 'absent progress must be explicit unknown, not a fabricated stage');
  assert.equal(absent.progress.observed_at, null, 'absent progress must not fabricate a zero observed_at');
  assert.equal(absent.progress.source, null, 'absent progress must have a null source');
  assert.deepEqual(absent.progress.evidence, [], 'absent progress must have empty evidence');
  assert.equal(absent.progress.percent ?? null, null, 'absent progress must not invent a percent');

  // 2. Human Set: advisory only, persisted, source=human, observed_at set.
  const written = await api.write(`/local-projects/${id}/progress`, { status: 'in progress', summary: 'human authored', percent: 40 }, { method: 'PUT' });
  assert.equal(written.progress.status, 'in progress');
  assert.equal(written.progress.source, 'human');
  assert.equal(written.progress.percent, 40);
  assert.ok(written.progress.observed_at, 'human write must carry a real observed_at');

  // 3. Durable restart: fresh API process re-reads the same persisted progress.
  await server.restart();
  api = await humanAPI(server.apiURL);
  const refetched = await api.get(`/local-projects/${id}/progress`);
  assert.equal(refetched.progress.status, 'in progress');
  assert.equal(refetched.progress.source, 'human');
  assert.equal(refetched.progress.percent, 40);
  assert.equal(refetched.progress.observed_at, written.progress.observed_at);

  // 4. Infer without model consent: real error (approval required), never a 200 unknown.
  const infer = await api.request(`/local-projects/${id}/progress/infer`, { method: 'POST', key: randomUUID(), body: { schema_version: 1, request_id: randomUUID(), expected_version: 1, files: ['TASK.md', 'STATUS.md'] } });
  assert.equal(infer.status, 403, `infer without model consent must fail, got ${infer.status}`);
  assert.ok(infer.data.error && infer.data.error.code, 'infer error must be a structured error, not a fake 200 unknown');

  console.log(`PROGRESS_OK project=${id} status=${refetched.progress.status} source=${refetched.progress.source}`);
} finally {
  await server.close();
}
