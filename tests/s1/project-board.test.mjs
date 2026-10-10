import assert from 'node:assert/strict';
import { randomUUID } from 'node:crypto';
import { setTimeout as delay } from 'node:timers/promises';
import { test } from 'node:test';
import { startS1Server } from './server.mjs';
import { command, humanAPI } from './api.mjs';
import { validateResponse } from './contracts.mjs';

// Opt-in actual registered roots and native header observations. No synthetic
// board seeds, intercepted routes, model/media calls or personal browser reads.
test('actual project board: human metadata survives refresh and cold restart without granting access', {
  skip: process.env.ASTROCYTE_TEST_PROJECT_BOARD !== '1', timeout: 360_000,
}, async () => {
  const token = randomUUID();
  const server = await startS1Server({ env: { ASTROCYTE_AGENT_TOKEN: token } });
  let primaryError;
  try {
    let api = await humanAPI(server.apiURL);
    const deadline = Date.now() + 125_000;
    let snapshot;
    do {
      snapshot = (await api.get('/local-projects/registered')).snapshot;
      if (snapshot.observed_at) break;
      await delay(500);
    } while (Date.now() < deadline);
    assert.ok(snapshot.observed_at, 'Actual bounded startup discovery must finish');
    assert.ok(snapshot.board?.length, 'Actual registered projects must exist; empty is not a successful board acceptance');
    assert.ok(snapshot.sources?.length, 'CLI source coverage must be explicit');
    for (const project of snapshot.board) {
      for (const contributor of project.contributors) {
        assert.ok(Object.hasOwn(contributor, 'created_at'), 'Creation time must be distinct from activity');
        if (contributor.source === 'native_session_header' || contributor.source === 'native_project_metadata') {
          assert.equal(contributor.activity_at, null, 'Native creation headers cannot claim latest activity');
        }
      }
    }
    const selected = snapshot.board.find(project => project.roots.length > 0);
    assert.ok(selected, 'At least one actual registered root must be observed');
    assert.equal(selected.human.revision, 0);
    assert.equal(selected.human.updated_at, null);
    const path = `/local-projects/registered/${encodeURIComponent(selected.id)}/metadata`;
    const beforeProjects = await api.get('/local-projects');
    const grantCount = server.query('SELECT COUNT(*) AS n FROM local_agent_grants')[0].n;
    const unchanged = (await api.get('/local-projects/registered')).snapshot;
    assert.equal(unchanged.observed_at, snapshot.observed_at, 'Cached GET cannot refresh discovery');
    const fields = { notes: '真实观察项目的人类备注', review: '刷新与重启保留人工复盘', group: '本轮验收', intent: '保持未授予权限', archived: true, revision: 0 };
    for (const headers of [
      { 'X-CSRF-Token': '' },
      { Authorization: `Bearer ${token}` },
      { Origin: 'https://untrusted.invalid' },
      { Cookie: '' },
    ]) {
      const response = await api.request(path, { method: 'PUT', body: command(fields), headers });
      assert.equal(response.status, 403);
    }
    for (const forged of [{ updated_at: '2026-10-10T00:00:00Z' }, { actor_kind: 'human' }, { model_consent: true }, { progress: 100 }, { actions: ['start'] }]) {
      const response = await api.request(path, { method: 'PUT', body: command({ ...fields, ...forged }) });
      assert.equal(response.status, 400);
    }
    assert.equal(server.query('SELECT COUNT(*) AS n FROM local_project_metadata')[0].n, 0);
    const saved = await api.write(path, fields, { method: 'PUT' });
    assert.equal(saved.project.id, selected.id);
    assert.equal(saved.project.human.revision, 1);
    assert.ok(saved.project.human.updated_at);
    for (const [field, value] of Object.entries(fields)) {
      if (field !== 'revision') assert.equal(saved.project.human[field], value);
    }
    const stale = await api.request(path, { method: 'PUT', body: command({ ...fields, notes: '过期覆盖必须失败' }) });
    assert.equal(stale.status, 409);
    const raw = server.query('SELECT data,revision FROM local_project_metadata WHERE project_id=?', selected.id)[0];
    assert.equal(raw.revision, 1);
    assert.deepEqual(JSON.parse(raw.data), saved.project.human);

    const refresh = await fetch(`${server.apiURL}/api/v1/local-projects/registered/refresh`, {
      method: 'POST', signal: AbortSignal.timeout(130_000),
      headers: { Cookie: api.cookie, Origin: server.apiURL, 'Content-Type': 'application/json', 'X-CSRF-Token': api.session.csrf_token, 'Idempotency-Key': randomUUID() },
      body: JSON.stringify(command()),
    });
    const refreshed = await refresh.json();
    validateResponse('POST', '/local-projects/registered/refresh', refresh.status, refreshed);
    assert.equal(refresh.status, 200);
    assert.deepEqual(refreshed.snapshot.board.find(project => project.id === selected.id)?.human, saved.project.human);
    assert.deepEqual(await api.get('/local-projects'), beforeProjects);
    assert.equal(server.query('SELECT COUNT(*) AS n FROM local_agent_grants')[0].n, grantCount);

    await server.restart();
    const expired = await fetch(`${server.apiURL}/api/v1/local-projects/registered`, { headers: { Cookie: api.cookie } });
    assert.equal(expired.status, 403, 'Old in-memory human session is invalid after a real process restart');
    api = await humanAPI(server.apiURL);
    const restarted = (await api.get('/local-projects/registered')).snapshot;
    assert.deepEqual(restarted.board.find(project => project.id === selected.id)?.human, saved.project.human);
    assert.deepEqual(server.query('SELECT data,revision FROM local_project_metadata WHERE project_id=?', selected.id)[0], raw);
    assert.equal(server.query('SELECT COUNT(*) AS n FROM local_agent_grants')[0].n, grantCount);
    console.log(JSON.stringify({ scope: 'task_live_local_metadata', directory: server.temporary, project_id: selected.id, projects: snapshot.board.length, observed_roots: snapshot.projects.length, sources: snapshot.sources, revision: saved.project.human.revision }));
  } catch (error) {
    primaryError = error;
    console.error(`First project-board failure retained with SQLite: ${server.temporary}`);
    throw error;
  } finally {
    await server.close({ preserveData: true, primaryError });
  }
});
