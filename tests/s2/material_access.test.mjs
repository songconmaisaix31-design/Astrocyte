import assert from 'node:assert/strict';
import { execFile } from 'node:child_process';
import { randomUUID } from 'node:crypto';
import { join } from 'node:path';
import { promisify } from 'node:util';
import test from 'node:test';
import { humanAPI, importMaterial } from '../s1/api.mjs';
import { validateResponse } from '../s1/contracts.mjs';
import { startS1Server } from '../s1/server.mjs';
import { root } from '../../scripts/process.mjs';

const git = promisify(execFile);

test('actual public Git material references enforce current project access and fixed proposal revisions', { timeout: 120_000 }, async () => {
  const head = (await git('git', ['rev-parse', 'HEAD'], { cwd: root })).stdout.trim();
  const sourceText = (await git('git', ['show', `${head}:README.md`], { cwd: root, maxBuffer: 2 * 1024 * 1024 })).stdout;
  assert.ok(sourceText.length > 0);
  const application = await startS1Server();
  let failure;
  try {
    let api = await humanAPI(application.apiURL);
    const imported = await importMaterial(api, {
      kind: 'text', adapter: 'manual', source_key: '', content_digest: '',
      source_locator: `https://github.com/songconmaisaix31-design/Astrocyte/blob/${head}/README.md`,
      export_text: sourceText, title: 'Astrocyte public README at fixed Git revision',
      source_spans: [`Git ${head}:README.md`], collection_reason: null,
    });
    const material = imported.detail.material;
    const reference = { material_id: material.id, revision: material.current_revision };
    const firstSpace = (await api.write('/project-spaces', { title: 'Public fixed material S2 project' }, { status: 201 })).space;
    const secondSpace = (await api.write('/project-spaces', { title: 'Separate S2 project' }, { status: 201 })).space;
    let firstProject = (await api.write('/local-projects', { name: 'Public source reference review', root, space_id: firstSpace.id })).project;
    let secondProject = (await api.write('/local-projects', { name: 'Separate public domain review', root: join(root, 'internal', 'workspace', 'domain'), space_id: secondSpace.id })).project;
    firstProject = (await api.write(`/local-projects/${firstProject.id}/settings`, {
      expected_revision: firstProject.settings.revision,
      settings: { ...firstProject.settings, allow_directory: true, allowed_subdirs: ['internal/workspace/domain'], allowed_actions: ['read_context'] },
    }, { method: 'PUT' })).project;
    secondProject = (await api.write(`/local-projects/${secondProject.id}/settings`, {
      expected_revision: secondProject.settings.revision,
      settings: { ...secondProject.settings, allow_directory: true, allowed_subdirs: ['.'], allowed_actions: ['read_context'] },
    }, { method: 'PUT' })).project;
    const referenced = (await api.write(`/project-spaces/${firstSpace.id}/references`, {
      expected_version: firstSpace.version, ...reference,
    })).space;
    const goal = (await api.write('/goals', {
      expected_version: 0, goal: { project_id: firstProject.id, title: 'Review the fixed public README', status: 'active', success_criteria: ['Selected source access is explicit'], priority: 1 },
    })).goal;
    const otherGoal = (await api.write('/goals', {
      expected_version: 0, goal: { project_id: secondProject.id, title: 'Review the separate public project', status: 'active', success_criteria: ['Project boundaries are retained'], priority: 1 },
    })).goal;
    const content = {
      title: 'Review selected public source with current project permission',
      goal: { id: goal.id, revision: goal.revision }, inputs: [reference],
      scope: { project_id: firstProject.id, settings_revision: firstProject.settings.revision, roots: [secondProject.root], actions: ['read_context'], tools: [], external_targets: [], write_paths: [], max_concurrency: 1 },
      budget: { calls: 1, tokens: null, money_minor: null, currency: '' },
      stops: { deadline: new Date(Date.now() + 60_000).toISOString(), conditions: ['Stop after selected source review'], max_work_items: 1, max_attempts: 1, no_progress_seconds: 0, max_no_progress_rounds: 1 },
      deliverables: ['Fixed source review'], success_criteria: ['Current project permission is checked'],
    };
    const proposal = (await api.write('/proposals', { expected_version: 0, content })).proposal;
    const crossProject = await api.request('/proposals', { method: 'POST', body: {
      schema_version: 1, request_id: randomUUID(), expected_version: 0,
      content: { ...content, goal: { id: otherGoal.id, revision: otherGoal.revision } },
    } });
    assert.equal(crossProject.status, 403);
    assert.equal(crossProject.data.error.code, 'scope_denied');
    const context = await api.write(`/local-projects/${firstProject.id}/context`, { references: [reference], files: [] });
    assert.deepEqual(context.materials[0].reference, reference);
    assert.ok(context.materials[0].text === sourceText, 'Selected source text must equal the real fixed Git object');
    const agentID = randomUUID();
    await api.write(`/local-projects/${firstProject.id}/grants`, { agent_id: agentID, actions: ['read_context'] });
    const credential = (await api.write(`/local-projects/${firstProject.id}/agent-token`, { agent_id: agentID })).credential;
    async function agentContext(projectID) {
      const path = `/local-projects/${projectID}/context`;
      const response = await fetch(`${application.apiURL}/api/v1${path}`, {
        method: 'POST', signal: AbortSignal.timeout(10_000),
        headers: { Authorization: `Bearer ${credential.token}`, 'Content-Type': 'application/json', 'Idempotency-Key': randomUUID() },
        body: JSON.stringify({ schema_version: 1, request_id: randomUUID(), expected_version: 1, references: [reference], files: [] }),
      });
      const data = await response.json();
      validateResponse('POST', path, response.status, data);
      return { status: response.status, data };
    }
    const humanUsage = (await api.get(`/materials/${material.id}`)).material.human_usage_count;
    const allowed = await agentContext(firstProject.id);
    assert.equal(allowed.status, 200);
    assert.ok(allowed.data.materials[0].text === sourceText, 'Authorized Agent reads the selected real Git source');
    const denied = await agentContext(secondProject.id);
    assert.equal(denied.status, 403);
    assert.equal(denied.data.error.code, 'scope_denied');
    assert.equal((await api.get(`/materials/${material.id}`)).material.human_usage_count, humanUsage);
    await api.write(`/project-spaces/${firstSpace.id}/references/remove`, {
      expected_version: referenced.version, material_id: material.id,
    });
    const removed = await agentContext(firstProject.id);
    assert.equal(removed.status, 403);
    assert.equal(removed.data.error.code, 'scope_denied');
    const submit = await api.request(`/proposals/${proposal.id}/submit`, { method: 'POST', body: {
      schema_version: 1, request_id: randomUUID(), expected_version: proposal.version,
    } });
    assert.equal(submit.status, 403);
    assert.equal(submit.data.error.code, 'scope_denied');
    const retained = (await api.get(`/proposals/${proposal.id}?revision=1`)).proposal;
    assert.equal(retained.status, 'draft');
    assert.deepEqual(retained.content.inputs, [reference]);
    await application.restart();
    api = await humanAPI(application.apiURL);
    assert.deepEqual((await api.get(`/proposals/${proposal.id}?revision=1`)).proposal.content.inputs, [reference]);
    assert.equal((await agentContext(firstProject.id)).status, 403);
    assert.equal(application.query('SELECT count(*) AS n FROM workspace_approvals')[0].n, 0);
    assert.equal(application.query('SELECT count(*) AS n FROM local_agent_sessions')[0].n, 0);
    assert.equal(application.query('SELECT count(*) AS n FROM swarm_missions')[0].n, 0);
    console.log(JSON.stringify({ sourceHead: head, materialRevision: reference.revision, result: 'actual_material_access_pass', database: join(application.dataDir, 'state.sqlite') }));
  } catch (error) {
    failure = error;
    throw error;
  } finally {
    await application.close({ primaryError: failure });
  }
});
