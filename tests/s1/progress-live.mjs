// 实际 HTTP 与 SQLite 验证人工输入、未知状态、重启及未授权推断的拒绝。
import assert from 'node:assert/strict';
import { randomUUID } from 'node:crypto';
import { startS1Server } from './server.mjs';
import { humanAPI } from './api.mjs';
import { root } from '../../scripts/process.mjs';

const server = await startS1Server();
try {
  let api = await humanAPI(server.apiURL);

  const space = await api.write('/project-spaces', { title: 'progress-acceptance-space' }, { status: 201 });
  const project = await api.write('/local-projects', { name: 'progress-acceptance', root, space_id: space.space.id });
  const id = project.project.id;

  // 初始进度明确保持未知。
  const absent = await api.get(`/local-projects/${id}/progress`);
  assert.equal(absent.progress.status, 'unknown', 'absent progress must be explicit unknown, not a fabricated stage');
  assert.equal(absent.progress.observed_at, null, 'absent progress must not fabricate a zero observed_at');
  assert.equal(absent.progress.source, null, 'absent progress must have a null source');
  assert.deepEqual(absent.progress.evidence, [], 'absent progress must have empty evidence');
  assert.equal(absent.progress.percent ?? null, null, 'absent progress must not invent a percent');

  // 下列状态与百分比是本次明确输入的人工字段，不描述实际工程进度。
  const written = await api.write(`/local-projects/${id}/progress`, { status: 'in progress', summary: 'human authored', percent: 40 }, { method: 'PUT' });
  assert.equal(written.progress.status, 'in progress');
  assert.equal(written.progress.source, 'human');
  assert.equal(written.progress.percent, 40);
  assert.ok(written.progress.observed_at, 'human write must carry a real observed_at');

  // 新进程重新读取同一人工输入。
  await server.restart();
  api = await humanAPI(server.apiURL);
  const refetched = await api.get(`/local-projects/${id}/progress`);
  assert.equal(refetched.progress.status, 'in progress');
  assert.equal(refetched.progress.source, 'human');
  assert.equal(refetched.progress.percent, 40);
  assert.equal(refetched.progress.observed_at, written.progress.observed_at);

  // 没有模型许可时，在读取实际 README 之前拒绝推断。
  const infer = await api.request(`/local-projects/${id}/progress/infer`, { method: 'POST', key: randomUUID(), body: { schema_version: 1, request_id: randomUUID(), expected_version: 1, files: ['README.md'] } });
  assert.equal(infer.status, 403, `infer without model consent must fail, got ${infer.status}`);
  assert.ok(infer.data.error && infer.data.error.code, 'infer error must be a structured error, not a fake 200 unknown');

  console.log(`PROGRESS_OK project=${id} status=${refetched.progress.status} source=${refetched.progress.source}`);
} finally {
  await server.close();
}
