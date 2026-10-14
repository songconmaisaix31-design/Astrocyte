import assert from 'node:assert/strict';
import { randomUUID } from 'node:crypto';
import { join } from 'node:path';
import { DatabaseSync } from 'node:sqlite';
import test from 'node:test';
import { humanAPI } from '../s1/api.mjs';
import { startS1Server } from '../s1/server.mjs';
import { root } from '../../scripts/process.mjs';

test('actual production HTTP approval, revision, idempotency, revoke and process restart', { timeout: 120_000 }, async () => {
  const application = await startS1Server();
  let failure;
  try {
    let api = await humanAPI(application.apiURL);
    const space = await api.write('/project-spaces', { title: 'Public S2 HTTP validation' }, { status: 201 });
    const registered = await api.write('/local-projects', { name: 'Astrocyte public S2 HTTP', root, space_id: space.space.id });
    let project = registered.project;
    const configured = await api.write(`/local-projects/${project.id}/settings`, {
      expected_revision: project.settings.revision,
      settings: { ...project.settings, allow_directory: true, allowed_subdirs: ['.'], allowed_actions: ['read_context'] },
    }, { method: 'PUT' });
    project = configured.project;
    const goalCreated = await api.write('/goals', {
      expected_version: 0,
      goal: { project_id: project.id, title: 'Review public approval behavior', status: 'active', success_criteria: ['Durable approval and revocation'], priority: 1 },
    });
    const goal = goalCreated.goal;
    assert.equal(goal.revision, 1);
    assert.ok((await api.get('/goals')).items.some(item => item.id === goal.id));
    const deadline = new Date(Date.now() + 60_000).toISOString();
    const content = {
      title: 'Review Astrocyte S2 approval implementation',
      goal: { id: goal.id, revision: goal.revision }, inputs: [],
      scope: { project_id: project.id, settings_revision: project.settings.revision, roots: [project.root], actions: ['read_context'], tools: [], external_targets: [], write_paths: [], max_concurrency: 2 },
      budget: { calls: 2, tokens: null, money_minor: null, currency: '' },
      stops: { deadline, conditions: ['Stop after verification'], max_work_items: 2, max_attempts: 2, no_progress_seconds: 0, max_no_progress_rounds: 2 },
      deliverables: ['Public approval verification'], success_criteria: ['No executor starts on approval'],
    };
    const createKey = randomUUID();
    const created = await api.write('/proposals', { expected_version: 0, content }, { key: createKey });
    let proposal = created.proposal;
    const repeated = await api.write('/proposals', { expected_version: 0, content }, { key: createKey });
    assert.equal(repeated.proposal.id, proposal.id);
    const changed = await api.request('/proposals', { method: 'POST', key: createKey, body: { schema_version: 1, request_id: randomUUID(), expected_version: 0, content: { ...content, title: 'Revised public approval input' } } });
    assert.equal(changed.status, 409);
    assert.equal(changed.data.error.code, 'version_conflict');
    proposal = (await api.write(`/proposals/${proposal.id}/submit`, { expected_version: proposal.version })).proposal;
    assert.equal(proposal.status, 'in_review');
    const agentID = randomUUID();
    await api.write(`/local-projects/${project.id}/grants`, { agent_id: agentID, actions: ['read_context'] });
    const issued = await api.write(`/local-projects/${project.id}/agent-token`, { agent_id: agentID });
    const agentResponse = await fetch(`${application.apiURL}/api/v1/proposals/${proposal.id}/approvals`, {
      method: 'POST', headers: { Authorization: `Bearer ${issued.credential.token}`, 'Content-Type': 'application/json', 'Idempotency-Key': randomUUID() },
      body: JSON.stringify({ schema_version: 1, request_id: randomUUID(), expected_version: proposal.version, expected_revision: proposal.revision, expires_at: deadline }),
      signal: AbortSignal.timeout(10_000),
    });
    assert.equal(agentResponse.status, 403);
    assert.equal(application.query('SELECT count(*) AS n FROM workspace_approvals')[0].n, 0);
    const stale = await api.request(`/proposals/${proposal.id}/approvals`, { method: 'POST', body: { schema_version: 1, request_id: randomUUID(), expected_version: proposal.version - 1, expected_revision: proposal.revision, expires_at: deadline } });
    assert.equal(stale.status, 409);
    assert.equal(stale.data.error.code, 'version_conflict');
    const approvalKey = randomUUID();

    // 真实SQLite约束在Grant写入后拒绝事件，核完整批准事务回滚。
    const databasePath = join(application.dataDir, 'state.sqlite');
    let database = new DatabaseSync(databasePath);
    try {
      database.exec("CREATE TRIGGER s2_approval_event_constraint BEFORE INSERT ON foundation_outbox WHEN NEW.event_type='workspace.proposal_approved' BEGIN SELECT RAISE(ABORT, 'S2 approval event constraint'); END");
    } finally {
      database.close();
    }
    const failedApproval = await api.request(`/proposals/${proposal.id}/approvals`, {
      method: 'POST', key: approvalKey,
      body: { schema_version: 1, request_id: randomUUID(), expected_version: proposal.version, expected_revision: proposal.revision, expires_at: deadline },
    });
    assert.equal(failedApproval.status, 500);
    assert.equal(application.query('SELECT count(*) AS n FROM workspace_approvals')[0].n, 0);
    assert.equal(application.query('SELECT count(*) AS n FROM foundation_idempotency WHERE command_type=?', 'workspace.proposal.approve')[0].n, 0);
    assert.equal((await api.get(`/proposals/${proposal.id}`)).proposal.status, 'in_review');
    database = new DatabaseSync(databasePath);
    try {
      database.exec('DROP TRIGGER s2_approval_event_constraint');
    } finally {
      database.close();
    }
    const approved = await api.write(`/proposals/${proposal.id}/approvals`, { expected_version: proposal.version, expected_revision: proposal.revision, expires_at: deadline }, { key: approvalKey });
    const grant = approved.approval;
    assert.equal(grant.status, 'active');
    const replayed = await api.write(`/proposals/${proposal.id}/approvals`, { expected_version: proposal.version, expected_revision: proposal.revision, expires_at: deadline }, { key: approvalKey });
    assert.equal(replayed.approval.id, grant.id);
    const detail = await api.get(`/proposals/${proposal.id}`);
    assert.equal(detail.proposal.grant_id, grant.id);
    const events = application.query('SELECT event_id, aggregate_version, payload FROM foundation_outbox WHERE event_type=?', 'workspace.proposal_approved');
    assert.equal(events.length, 1);
    assert.equal(JSON.parse(events[0].payload).grant_id, grant.id);
    assert.equal(application.query('SELECT count(*) AS n FROM local_agent_sessions')[0].n, 0);
    assert.equal(application.query('SELECT count(*) AS n FROM swarm_missions')[0].n, 0);
    await application.restart();
    api = await humanAPI(application.apiURL);
    assert.deepEqual((await api.get(`/approvals/${grant.id}`)).approval, grant);
    assert.equal((await api.get(`/proposals/${proposal.id}?revision=1`)).proposal.content.goal.revision, goal.revision);
    const revoked = await api.write(`/approvals/${grant.id}/revoke`, { expected_version: grant.epoch, reason: 'End actual public HTTP validation' });
    assert.equal(revoked.approval.status, 'revoked');
    assert.equal(revoked.approval.epoch, grant.epoch + 1);
    await application.restart();
    api = await humanAPI(application.apiURL);
    assert.equal((await api.get(`/approvals/${grant.id}`)).approval.status, 'revoked');
    assert.equal(application.query('SELECT count(*) AS n FROM foundation_outbox WHERE event_type=?', 'workspace.approval_revoked')[0].n, 1);
    console.log(`S2 actual HTTP database: ${join(application.dataDir, 'state.sqlite')}`);
  } catch (error) {
    failure = error;
    throw error;
  } finally {
    await application.close({ primaryError: failure });
  }
});
