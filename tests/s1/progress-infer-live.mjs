// 对本工作树实际公开 README 执行一次 Codex 推断，并核对 HTTP、SQLite 与重启。
import assert from 'node:assert/strict';
import { randomUUID } from 'node:crypto';
import { stat } from 'node:fs/promises';
import { join } from 'node:path';
import { startS1Server } from './server.mjs';
import { humanAPI } from './api.mjs';

const realRoot = process.cwd();
const files = ['README.md'];
const inputStat = await stat(join(realRoot, files[0]), { bigint: true });
const expectedVersion = `mtime:${inputStat.mtimeNs};size:${inputStat.size}`;

const server = await startS1Server();
let primaryError;
try {
  let api = await humanAPI(server.apiURL);

  const space = await api.write('/project-spaces', { title: 'progress-infer-space' }, { status: 201 });
  const project = await api.write('/local-projects', { name: 'astrocyte', root: realRoot, space_id: space.space.id });
  const id = project.project.id;

  await api.write(`/local-projects/${id}/settings`, {
    expected_revision: 1,
    settings: {
      allow_directory: true,
      allowed_subdirs: files,
      external_model_cli: 'codex',
      allowed_actions: ['start', 'stop', 'read_context', 'send', 'observe', 'discover', 'resume', 'reconcile'],
    },
  }, { method: 'PUT' });

  // 实际探测空会话并确认停止，推断仅使用当前配置。
  const probe = await api.write(`/local-projects/${id}/sessions/probe`, { cli: 'codex' });
  assert.ok(probe.session.stop_confirmed, 'probe must confirm owned Codex stop');

  // 操作身份来自 Idempotency-Key；未知交付保留原件且不重发。
  const op = randomUUID();
  const response = await fetch(`${server.apiURL}/api/v1/local-projects/${id}/progress/infer`, {
    method: 'POST',
    signal: AbortSignal.timeout(300_000),
    headers: {
      Cookie: api.cookie,
      Origin: server.apiURL,
      'Content-Type': 'application/json',
      'X-CSRF-Token': api.session.csrf_token,
      'Idempotency-Key': op,
    },
    body: JSON.stringify({ schema_version: 1, request_id: op, expected_version: 1, files }),
  });
  const body = await response.json();
  assert.equal(response.status, 200, JSON.stringify(body));
  const progress = body.progress;
  assert.equal(progress.source, 'agent_inferred', 'real infer must be agent_inferred');
  assert.ok(progress.native_id, 'real infer must carry the actual native session id');
  assert.ok(progress.model, 'real infer must report the actual model');
  assert.ok(progress.observed_at, 'real infer must carry a real observed time');
  assert.ok(progress.evidence.length > 0, 'real infer must carry file evidence');
  assert.ok(progress.evidence.every(e => e.source_path === files[0] && e.version === expectedVersion), 'evidence must name the actual README version');
  const persisted = JSON.parse(server.query('SELECT data FROM local_project_progress WHERE project_id=?', id)[0].data);
  assert.equal(persisted.source, 'agent_inferred');
  assert.equal(persisted.native_id, progress.native_id);
  assert.deepEqual(persisted.evidence, progress.evidence);

  // 新 API 进程读取已有推断，保持原 native_id 与来源版本。
  await server.restart();
  api = await humanAPI(server.apiURL);
  const refetched = await api.get(`/local-projects/${id}/progress`);
  assert.equal(refetched.progress.source, 'agent_inferred');
  assert.equal(refetched.progress.native_id, progress.native_id);
  assert.equal(refetched.progress.status, progress.status);
  assert.equal((await api.get('/missions')).items.length, 0, 'advisory inference must never create a Mission');

  console.log(JSON.stringify({
    phase: 'real_progress_infer_complete',
    project_id: id, status: progress.status, source: progress.source,
    native_id: progress.native_id, model: progress.model, observed_at: progress.observed_at,
    evidence: progress.evidence, files, cost: null,
  }, null, 2));
} catch (error) {
  primaryError = error;
  throw error;
} finally {
  await server.close({ preserveData: true, primaryError });
}
