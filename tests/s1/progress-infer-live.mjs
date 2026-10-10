// One authorized real Codex project-progress inference over the real Astrocyte
// worktree's existing maintained task document. This is a single paid model
// turn, never replayed; it reads no personal history and fabricates no TASK.md.
//
// Usage:
//   node tests/s1/progress-infer-live.mjs
import assert from 'node:assert/strict';
import { randomUUID } from 'node:crypto';
import { startS1Server } from './server.mjs';
import { humanAPI } from './api.mjs';

const realRoot = process.cwd();
const files = ['tasks/S1-final-W2.md'];

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
      allowed_subdirs: ['tasks'],
      external_model_cli: 'codex',
      allowed_actions: ['start', 'stop', 'read_context', 'send', 'observe', 'discover', 'resume', 'reconcile'],
    },
  }, { method: 'PUT' });

  // Observe the Codex native configuration (empty session + owned stop, no model turn).
  const probe = await api.write(`/local-projects/${id}/sessions/probe`, { cli: 'codex' });
  assert.ok(probe.session.stop_confirmed, 'probe must confirm owned Codex stop');

  // One real model turn. Operation identity is the Idempotency-Key header (a UUID).
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
  assert.ok(progress.evidence.every(e => e.source_path === 'tasks/S1-final-W2.md' && e.version), 'evidence must name the real file with its version');

  // Durable: fresh API process re-reads the persisted inference (no new model turn).
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
