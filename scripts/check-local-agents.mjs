import assert from 'node:assert/strict';
import { randomUUID } from 'node:crypto';
import { startS1Server } from '../tests/s1/server.mjs';
import { humanAPI } from '../tests/s1/api.mjs';

// Actual server/HTTP/schema smoke check in an isolated empty SQLite directory.
// Reuse the S1 lifecycle helpers; no browser, model request or media extraction.
const agentToken = randomUUID();
const server = await startS1Server({ env: { ASTROCYTE_AGENT_TOKEN: agentToken } });
try {
  let api = await humanAPI(server.apiURL);
  const foundation = await api.get('/foundation');
  assert.equal(foundation.stage, 'S1');
  assert.deepEqual(foundation.capabilities, {
    imports: true, approvals: false, execution: false, native_resume: false, handoff: false,
  });
  const first = await api.get('/local-agents');
  assert.ok(first.items.length, 'Assembled inventory must retain declared CLI observations');
  assert.equal(new Set(first.items.map(item => item.id)).size, first.items.length);
  const nativeKeys = ['discover', 'read_context', 'start', 'resume', 'send', 'stop', 'observe', 'reconcile'].sort();
  for (const item of first.items) {
    assert.equal(item.configured.status, 'unknown', 'Public CLI probes do not inspect private configuration');
    assert.equal(item.startable.status, 'unknown', 'Version/help is not a native start');
    assert.deepEqual(Object.keys(item.capabilities).sort(), nativeKeys);
    for (const capability of Object.values(item.capabilities)) {
      assert.equal(capability.status, 'unknown');
      assert.equal(capability.checked_at, null, 'Unrun native checks must not receive invented timestamps');
    }
  }
  for (let read = 0; read < 3; read++) {
    assert.deepEqual(await api.get('/local-agents'), first, 'GET must not refresh CLI probes');
  }
  const unauthenticated = await fetch(`${server.apiURL}/api/v1/local-agents`);
  assert.equal(unauthenticated.status, 403);
  assert.equal((await api.request('/local-agents', { headers: { Authorization: `Bearer ${agentToken}` } })).status, 403);
  assert.equal((await api.request('/local-agents', { headers: { Origin: 'https://example.invalid' } })).status, 403);
  const resume = await api.request('/sessions/unbound/resume', { method: 'POST', body: {} });
  assert.equal(resume.status, 501, 'Discovery must not activate native control');
  assert.equal(resume.data.error.code, 'unsupported_capability');
  const counts = () => Object.fromEntries(['attention_materials', 'attention_jobs', 'attention_distillations', 'attention_outbox'].map(table => [table, server.query(`SELECT COUNT(*) AS n FROM ${table}`)[0].n]));
  assert.deepEqual(Object.values(counts()), [0, 0, 0, 0], 'Discovery must not create Attention records');
  await server.restart();
  api = await humanAPI(server.apiURL);
  const restarted = await api.get('/local-agents');
  assert.deepEqual(restarted.items.map(item => item.id), first.items.map(item => item.id));
  assert.deepEqual(Object.values(counts()), [0, 0, 0, 0]);
  console.log(JSON.stringify({
    result: 'PASS', scope: 'installed_cli_http', cached_reads: 4,
    unauthenticated: 403, agent: 403, foreign_origin: 403, resume: 501,
    restart: 'PASS', attention_counts: counts(),
    cli: restarted.items.map(({ id, version, installed, configured, startable }) => ({
      id, version, installed: installed.status, probe: installed.reason,
      configured: configured.status, startable: startable.status,
    })),
    native_capabilities: 'NOT_RUN / unknown',
  }, null, 2));
} finally {
  await server.close();
}
