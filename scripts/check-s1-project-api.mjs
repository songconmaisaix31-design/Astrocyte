// 实际公开论文原件通过生产 HTTP、SQLite、对象存储及项目权限处理。
import assert from 'node:assert/strict';
import { access, mkdir, writeFile } from 'node:fs/promises';
import { join, resolve } from 'node:path';
import { randomUUID } from 'node:crypto';
import { startS1Server } from '../tests/s1/server.mjs';
import { command, humanAPI, importMaterial } from '../tests/s1/api.mjs';
import { loadPublicPaperInputs } from '../tests/s1/public-inputs.mjs';
import { validateResponse } from '../tests/s1/contracts.mjs';

const papers = await loadPublicPaperInputs(resolve(process.argv[2] ?? '.verification'));
const paperText = papers.original.text;
let server = await startS1Server();
let stage = 'bootstrap';
let preserveOnExit = false;
try {
  let api = await humanAPI(server.apiURL);
  assert.deepEqual((await api.get('/tracking-sources')).items, []);
  stage = 'repository cache and human command boundary';
  assert.deepEqual((await api.get('/github-repositories')).items, []);
  assert.deepEqual((await api.get('/github-repositories')).items, []);
  assert.equal(server.query('SELECT COUNT(*) AS n FROM local_github_repositories')[0].n, 0);
  await assert.rejects(access(join(server.dataDir, 'repositories')), { code: 'ENOENT' });
  assert.equal((await api.request('/github-repositories/sync', { method: 'POST', body: command({ input: 'file:///unapproved' }) })).status, 400);
  assert.equal((await api.request('/github-repositories/missing/placement', { method: 'POST', body: command({ space_id: '' }) })).status, 400);
  assert.equal((await api.request('/github-repositories/sync', { method: 'POST', body: command({ input: 'example/repository' }), headers: { 'X-CSRF-Token': '' } })).status, 403);
  assert.equal(server.query('SELECT COUNT(*) AS n FROM local_github_repositories')[0].n, 0);
  await assert.rejects(access(join(server.dataDir, 'repositories')), { code: 'ENOENT' });
  stage = 'fixed source and explicit refresh';
  const imported = await importMaterial(api, papers.original.import);
  const material = imported.detail.material;
  await importMaterial(api, { ...papers.revised.import, refresh: true });
  const reused = await importMaterial(api, papers.original.import);
  assert.equal(reused.detail.material.current_revision, 2, 'A-B-A reuse must retain B as the head');
  assert.ok((await api.get(`/materials/${material.id}/revisions/1/content`)).text === paperText, 'actual original revision text differs');
  const space = (await api.write('/project-spaces', { title: '公开论文权限验收空间' }, { status: 201 })).space;
  const referenced = (await api.write(`/project-spaces/${space.id}/references`, {
    expected_version: space.version, material_id: material.id, revision: 1,
  })).space;
  const projectRoot = join(server.temporary, 'approved-project');
  await mkdir(projectRoot);
  await writeFile(join(projectRoot, 'selected.txt'), '人工明确选择的受控目录内容');
  stage = 'register project contract';
  const project = (await api.write('/local-projects', {
    name: '公开论文权限验收项目', root: projectRoot, space_id: space.id,
  })).project;
  assert.deepEqual(project.settings.history_roots, {});
  assert.equal(project.settings.allow_directory, false);
  assert.equal(project.settings.expand_references, false);
  assert.equal(project.settings.external_model_cli, '');
  const contextBody = { references: [{ material_id: material.id, revision: 1 }], files: [] };
  stage = 'human fixed reference read';
  assert.ok((await api.write(`/local-projects/${project.id}/context`, contextBody)).materials[0].text === paperText, 'human fixed reference text differs');
  const baseline = (await api.get(`/materials/${material.id}`)).material;
  stage = 'human grant and scoped Agent token';
  await api.write(`/local-projects/${project.id}/grants`, { agent_id: 'w0-contract-agent', actions: ['read_context', 'observe'] });
  const credential = (await api.write(`/local-projects/${project.id}/agent-token`, { agent_id: 'w0-contract-agent' })).credential;
  // 凭据只留在内存，不进入验证输出。
  async function agentRequest(path, body, method = 'POST') {
    const key = randomUUID();
    const response = await fetch(`${server.apiURL}/api/v1${path}`, {
      method, signal: AbortSignal.timeout(10_000),
      headers: { Authorization: `Bearer ${credential.token}`, 'Content-Type': 'application/json', 'Idempotency-Key': key },
      ...(body === undefined ? {} : { body: JSON.stringify(command({ request_id: key, ...body })) }),
    });
    const data = await response.json();
    validateResponse(method, path, response.status, data);
    return { status: response.status, data };
  }
  const packet = await agentRequest(`/local-projects/${project.id}/context`, contextBody);
  assert.equal(packet.status, 200);
  assert.ok(packet.data.materials[0].text === paperText, 'authorized Agent fixed reference text differs');
  const after = (await api.get(`/materials/${material.id}`)).material;
  assert.equal(after.human_usage_count, baseline.human_usage_count);
  assert.equal(after.agent_usage_count, baseline.agent_usage_count + 1);
  stage = 'scope and human-only denial';
  assert.equal((await agentRequest(`/local-projects/${project.id}/context`, { references: [], files: ['selected.txt'] })).status, 403);
  assert.equal((await agentRequest(`/local-projects/${project.id}/grants`, { agent_id: 'self', actions: ['start'] })).status, 403);
  assert.equal((await agentRequest('/tracking-sources', undefined, 'GET')).status, 403);
  assert.equal((await agentRequest('/github-repositories', undefined, 'GET')).status, 403);
  assert.equal((await agentRequest('/source-collections/discover', { platform: 'douyin', owner_id: 'self', access_mode: 'browser_selected' })).status, 403);
  stage = 'durable restart and grant reload';
  await server.restart();
  api = await humanAPI(server.apiURL);
  assert.deepEqual((await api.get('/github-repositories')).items, []);
  assert.equal(server.query('SELECT COUNT(*) AS n FROM local_github_repositories')[0].n, 0);
  await assert.rejects(access(join(server.dataDir, 'repositories')), { code: 'ENOENT' });
  assert.equal((await agentRequest(`/local-projects/${project.id}/context`, contextBody)).status, 200);
  assert.equal(server.query('SELECT COUNT(*) AS n FROM attention_material_revisions WHERE material_id=?', material.id)[0].n, 2);
  assert.equal(JSON.parse(server.query('SELECT data FROM attention_project_spaces WHERE id=?', space.id)[0].data).material_refs[0].revision, 1);
  stage = 'current Attention membership revocation';
  await api.write(`/project-spaces/${space.id}/references/remove`, { expected_version: referenced.version, material_id: material.id });
  assert.equal((await agentRequest(`/local-projects/${project.id}/context`, contextBody)).status, 403);
  stage = 'immediate project grant revocation';
  await api.write(`/local-projects/${project.id}/grants/revoke`, { agent_id: 'w0-contract-agent' });
  assert.equal((await agentRequest(`/local-projects/${project.id}/context`, { references: [], files: [] })).status, 403);
  stage = 'preserve original store and explicitly reuse its owned directory';
  preserveOnExit = true;
  const original = server.temporary;
  await server.close({ preserveData: true });
  server = await startS1Server({ reuseOwnedTemporary: { path: original, ownedRoot: original } });
  api = await humanAPI(server.apiURL);
  assert.equal((await api.get(`/materials/${material.id}`)).material.current_revision, 2);
  assert.ok((await api.get(`/materials/${material.id}/revisions/1/content`)).text === paperText, 'original revision changed after owned store reopen');
  assert.equal((await agentRequest(`/local-projects/${project.id}/context`, { references: [], files: [] })).status, 403);
  assert.equal(server.query('SELECT COUNT(*) AS n FROM attention_material_revisions WHERE material_id=?', material.id)[0].n, 2);
  preserveOnExit = true;
  console.log('PASS actual public-paper HTTP/SQLite: A-B-A, fixed objects, scoped Agent usage, human settings, restart, membership/grant revocation, owned store reuse');
} catch (error) {
  preserveOnExit = true;
  console.error(`FAIL actual public-paper API at ${stage}`);
  throw error;
} finally {
  await server.close({ preserveData: preserveOnExit });
}
