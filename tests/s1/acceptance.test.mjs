import assert from 'node:assert/strict';
import { after, before, test } from 'node:test';
import { readFile, writeFile } from 'node:fs/promises';
import { join } from 'node:path';
import { randomUUID } from 'node:crypto';
import { startS1Server } from './server.mjs';
import { command, humanAPI, importMaterial, waitJob } from './api.mjs';
import { changedPaperText, distillation, paperImport, paperText, sourceRef, videoImport, videoTranscript } from './fixtures.mjs';
import { seedCandidate } from './scenarios.mjs';

// Explicit contract_local source records through the real HTTP, SQLite and object adapters.
// No public network, external summarizer, model call, or browser route interception.
const agentToken = randomUUID();
let server;
let api;
before(async () => {
  server = await startS1Server({ env: ({ temporary }) => ({ ASTROCYTE_AGENT_TOKEN: agentToken, ASTROCYTE_IMPORT_ROOTS: temporary }) });
  api = await humanAPI(server.apiURL);
}, { timeout: 120_000 });
after(async () => { if (server) await server.close(); });

test('AT01: source/version/positions and all three manual layers persist without a Mission', async () => {
  const missionsBefore = await api.get('/missions');
  const paper = await seedCandidate(api, 'https://example.invalid/contract-local/at01');
  const video = await importMaterial(api, videoImport());
  const videoRefs = [sourceRef(video.detail.material, 1, video.detail.revisions[0].source_spans[0])];
  for (const stage of ['content', 'topic', 'project']) {
    const result = await api.write('/distillations', distillation(stage, videoRefs));
    assert.equal(result.distillation.input_refs[0].revision, 1);
    assert.equal(result.distillation.input_refs[0].span, videoRefs[0].span);
  }
  assert.equal(paper.detail.material.collection_reason, null, 'Never invent the user collection reason');
  assert.deepEqual(paper.records.map(record => record.stage), ['content', 'topic', 'project']);
  for (const record of paper.records) {
    assert.equal(record.input_refs[0].revision, 1);
    assert.equal(record.input_refs[0].span, 'Section 1');
    assert.equal(record.provenance.mode, 'manual');
    assert.equal(await readFile(join(server.dataDir, 'objects', record.output_ref), 'utf8'), record.output_text);
  }
  const content = await api.get(`/materials/${paper.detail.material.id}/revisions/1/content`);
  assert.equal(content.text, paperText);
  const videoContent = await api.get(`/materials/${video.detail.material.id}/revisions/1/content`);
  assert.equal(videoContent.text, videoTranscript);
  assert.ok(video.detail.revisions[0].source_spans.some(span => span.includes('0.000s')));
  assert.ok(video.detail.revisions[0].attachments.some(attachment => attachment.name === 'summarize.json'));
  for (const dimension of Object.values(paper.candidate.opportunity.dimensions)) assert.equal(dimension.value, null);
  assert.equal(paper.candidate.opportunity.next_step, '人工对照两版输入，再决定是否继续');
  assert.deepEqual((await api.get('/missions')).items, missionsBefore.items);
  const stored = server.query('SELECT data FROM attention_materials WHERE id=?', paper.detail.material.id);
  assert.equal(JSON.parse(stored[0].data).material.current_revision, 1);
  assert.equal(server.query('SELECT COUNT(*) AS n FROM attention_outbox WHERE lower(type) LIKE ?', '%mission%')[0].n, 0);
});

test('AT02: concurrent same request and same source converge; same key with another body is rejected', async () => {
  const input = command(paperImport('https://example.invalid/contract-local/parallel'));
  const key = randomUUID();
  const responses = await Promise.all(Array.from({ length: 6 }, () => api.request('/materials/imports', { method: 'POST', body: input, key })));
  for (const response of responses) assert.equal(response.status, 202, JSON.stringify(response.data));
  assert.equal(new Set(responses.map(response => response.data.job_id)).size, 1);
  const firstJob = await waitJob(api, responses[0].data.job_id);
  const conflict = await api.request('/materials/imports', { method: 'POST', body: { ...input, export_text: changedPaperText }, key });
  assert.equal(conflict.status, 409, JSON.stringify(conflict.data));
  assert.equal(conflict.data.error.code, 'version_conflict');
  const aliases = await Promise.all(Array.from({ length: 4 }, (_, index) => importMaterial(api, { ...paperImport(`${input.source_locator}#section-${index}`), title: input.title })));
  for (const imported of aliases) assert.equal(imported.detail.material.id, firstJob.material_id);
  assert.equal(server.query('SELECT COUNT(*) AS n FROM attention_materials WHERE id=?', firstJob.material_id)[0].n, 1);
  assert.equal(server.query('SELECT COUNT(*) AS n FROM attention_material_revisions WHERE material_id=?', firstJob.material_id)[0].n, 1);
});

test('AT02: changed source keeps immutable old/new content; unchanged distillation retry reuses output', async () => {
  const locator = 'https://example.invalid/contract-local/revisions';
  const first = await importMaterial(api, paperImport(locator));
  const initialRevision = first.detail.revisions[0];
  const body = distillation('content', [sourceRef(first.detail.material)]);
  const original = await api.write('/distillations', body);
  const repeated = await api.write('/distillations', { ...body, output_text: 'This retry must not replace existing output' });
  assert.equal(repeated.reused, true);
  assert.equal(repeated.distillation.id, original.distillation.id);
  assert.equal(repeated.distillation.output_text, original.distillation.output_text);
  const second = await importMaterial(api, paperImport(locator, changedPaperText));
  assert.equal(second.detail.material.id, first.detail.material.id);
  assert.equal(second.detail.material.current_revision, 2);
  assert.deepEqual(second.detail.revisions[0], initialRevision);
  assert.equal((await api.get(`/materials/${first.detail.material.id}/revisions/1/content`)).text, paperText);
  assert.equal((await api.get(`/materials/${first.detail.material.id}/revisions/2/content`)).text, changedPaperText);
  const changedInput = await api.write('/distillations', distillation('content', [sourceRef(second.detail.material, 2)]));
  assert.notEqual(changedInput.distillation.id, original.distillation.id);
  const rows = server.query('SELECT revision,data FROM attention_material_revisions WHERE material_id=? ORDER BY revision', first.detail.material.id);
  assert.deepEqual(rows.map(row => row.revision), [1, 2]);
  for (const row of rows) {
    const revision = JSON.parse(row.data);
    assert.equal(await readFile(join(server.dataDir, 'objects', revision.object_ref), 'utf8'), row.revision === 1 ? paperText : changedPaperText);
  }
});

test('AT02: YouTube URL aliases deduplicate real adapter input; legacy summary never invents timestamps', async () => {
  const locators = ['https://youtu.be/S1LOCAL002a', 'https://www.youtube.com/watch?v=S1LOCAL002a&t=12', 'https://m.youtube.com/watch?v=S1LOCAL002a', 'https://www.youtube.com/embed/S1LOCAL002a'];
  const imports = await Promise.all(locators.map(locator => importMaterial(api, videoImport(locator))));
  assert.equal(new Set(imports.map(imported => imported.detail.material.id)).size, 1);
  assert.equal((await api.get(`/materials/${imports[0].detail.material.id}`)).revisions.length, 1);
  const legacy = await importMaterial(api, videoImport('https://youtu.be/S1LOCAL003a', { summaryOnly: true }));
  assert.deepEqual(legacy.detail.revisions[0].source_spans, []);
  assert.deepEqual(legacy.detail.material.source_spans, []);
});

test('AT03: Agent without material authorization is denied; human refresh never heats and explicit reread does', async () => {
  const imported = await importMaterial(api, paperImport('https://example.invalid/contract-local/attention'));
  await importMaterial(api, paperImport('https://example.invalid/contract-local/attention', changedPaperText));
  const id = imported.detail.material.id;
  const baseline = await api.get(`/materials/${id}`);
  for (let index = 0; index < 4; index++) {
    const response = await api.request(`/materials/${id}/revisions/1/content`, { headers: { Authorization: `Bearer ${agentToken}` } });
    assert.equal(response.status, 403, JSON.stringify(response.data));
  }
  for (let index = 0; index < 3; index++) {
    await api.get('/materials');
    await api.get(`/materials/${id}`);
  }
  const machineRead = await api.get(`/materials/${id}`);
  assert.equal(machineRead.material.human_usage_count, baseline.material.human_usage_count);
  assert.ok(machineRead.material.attention_score <= baseline.material.attention_score + 1e-9);
  assert.equal(machineRead.material.agent_usage_count, baseline.material.agent_usage_count);
  assert.equal(machineRead.uses.filter(use => use.actor_kind === 'agent').length, baseline.uses.filter(use => use.actor_kind === 'agent').length);
  const deniedDetail = await api.request(`/materials/${id}`, { headers: { Authorization: `Bearer ${agentToken}` } });
  assert.equal(deniedDetail.status, 403, JSON.stringify(deniedDetail.data));
  const deniedList = await api.request('/materials', { headers: { Authorization: `Bearer ${agentToken}` } });
  assert.equal(deniedList.status, 403, JSON.stringify(deniedList.data));
  for (const path of ['/distillations', '/opportunities', '/jobs', `/materials/${id}/revisions/1/attachments/original.pdf`]) {
    const denied = await api.request(path, { headers: { Authorization: `Bearer ${agentToken}` } });
    assert.equal(denied.status, 403, `${path}: ${JSON.stringify(denied.data)}`);
  }
  const forbiddenWrite = await api.request(`/materials/${id}/uses`, { method: 'POST', body: command({ expected_version: machineRead.material.version, action: 'reread' }), headers: { Authorization: `Bearer ${agentToken}` } });
  assert.equal(forbiddenWrite.status, 403, JSON.stringify(forbiddenWrite.data));
  const forbiddenBootstrap = await fetch(`${server.apiURL}/api/v1/auth/session`, { headers: { Authorization: `Bearer ${agentToken}` } });
  assert.equal(forbiddenBootstrap.status, 403);
  const key = randomUUID();
  const use = command({ action: 'reread', expected_version: machineRead.material.version });
  const reread = await api.request(`/materials/${id}/uses`, { method: 'POST', body: use, key });
  assert.equal(reread.status, 200, JSON.stringify(reread.data));
  const replay = await api.request(`/materials/${id}/uses`, { method: 'POST', body: use, key });
  assert.deepEqual(replay, reread);
  const result = await api.get(`/materials/${id}`);
  assert.equal(result.material.human_usage_count, baseline.material.human_usage_count + 1);
  assert.ok(result.material.attention_score > machineRead.material.attention_score);
  const stored = JSON.parse(server.query('SELECT data FROM attention_materials WHERE id=?', id)[0].data);
  assert.equal(stored.material.human_usage_count, result.material.human_usage_count);
  assert.equal(stored.uses.filter(entry => entry.actor_kind === 'human' && entry.action === 'reread').length, 1);
});

test('AT02: failed local export recovers under the original job/operation after its file is corrected', async () => {
  const locator = 'https://youtu.be/S1LOCAL004a';
  const exported = videoImport(locator);
  const file = join(server.temporary, 'contract-local-summarize.json');
  await writeFile(file, '{bad contract_local JSON');
  const input = { ...exported, local_file_ref: file };
  delete input.export_text;
  const receipt = await api.write('/materials/imports', input, { status: 202 });
  const failed = await waitJob(api, receipt.job_id, 'failed');
  assert.ok(failed.error?.required_action);
  assert.equal(failed.material_id, null);
  await writeFile(file, exported.export_text);
  await api.write(`/jobs/${failed.job_id}/retry`, { expected_version: failed.version });
  const recovered = await waitJob(api, failed.job_id);
  assert.equal(recovered.job_id, failed.job_id);
  assert.equal(recovered.operation_id, failed.operation_id);
  assert.equal(recovered.attempts, failed.attempts + 1);
  assert.equal((await api.get(`/materials/${recovered.material_id}/revisions/1/content`)).text, videoTranscript);
  assert.equal(server.query('SELECT COUNT(*) AS n FROM attention_jobs WHERE id=?', failed.job_id)[0].n, 1);
});

test('version conflict: metadata CAS is independent of content revision and stale feedback cannot overwrite', async () => {
  const seeded = await seedCandidate(api, 'https://example.invalid/contract-local/cas');
  const material = seeded.detail.material;
  await api.write(`/materials/${material.id}`, { expected_version: material.version, lifecycle: 'active', collection_reason: 'contract_local human note', pinned: true }, { method: 'PATCH' });
  const stale = await api.request(`/materials/${material.id}`, { method: 'PATCH', body: command({ expected_version: material.version, lifecycle: 'withdrawn', collection_reason: null, pinned: false }) });
  assert.equal(stale.status, 409, JSON.stringify(stale.data));
  const current = await api.get(`/materials/${material.id}`);
  assert.equal(current.material.current_revision, 1);
  assert.equal(current.material.pinned, true);
  const candidate = seeded.candidate.opportunity;
  await api.write(`/opportunities/${candidate.id}/reviews`, { expected_version: candidate.version, feedback: 'later', reason: 'contract_local postpone' }, { status: 201 });
  const staleReview = await api.request(`/opportunities/${candidate.id}/reviews`, { method: 'POST', body: command({ expected_version: candidate.version, feedback: 'reject', reason: 'stale must not take effect' }) });
  assert.equal(staleReview.status, 409, JSON.stringify(staleReview.data));
  const result = await api.get(`/opportunities/${candidate.id}`);
  assert.equal(result.opportunity.state, 'deferred');
  assert.deepEqual(result.reviews.map(review => review.feedback), ['later']);
});

test('AT04/AT02: later retains feedback/version/material and import receipt across an actual process restart', async () => {
  const seeded = await seedCandidate(api, 'https://example.invalid/contract-local/restart');
  const candidate = seeded.candidate.opportunity;
  await api.write(`/opportunities/${candidate.id}/reviews`, { expected_version: candidate.version, feedback: 'later', reason: '以后再做；缺少真实项目证据' }, { status: 201 });
  const saved = await api.get(`/opportunities/${candidate.id}`);
  assert.equal(saved.opportunity.state, 'deferred');
  assert.equal(saved.reviews[0].revision, candidate.revision);
  assert.equal(saved.reviews[0].feedback, 'later');
  assert.equal(saved.reviews.some(review => review.feedback === 'reject'), false);
  const stored = JSON.parse(server.query('SELECT data FROM attention_opportunities WHERE id=?', candidate.id)[0].data);
  assert.deepEqual(stored.reviews, saved.reviews);
  const input = command(paperImport('https://example.invalid/contract-local/persistent-receipt'));
  const key = randomUUID();
  const original = await api.request('/materials/imports', { method: 'POST', body: input, key });
  assert.equal(original.status, 202, JSON.stringify(original.data));
  const job = await waitJob(api, original.data.job_id);
  const durableJob = server.query('SELECT payload,caller FROM attention_jobs WHERE id=?', job.job_id)[0];
  assert.equal(JSON.parse(durableJob.payload).source_locator, input.source_locator);
  assert.equal(JSON.parse(durableJob.caller).Kind, 'human');
  const actorID = api.session.actor_id;
  await server.restart();
  api = await humanAPI(server.apiURL);
  assert.equal(api.session.actor_id, actorID, 'Human identity must remain stable for durable idempotency receipts');
  assert.deepEqual(await api.get(`/opportunities/${candidate.id}`), saved);
  assert.equal((await api.get(`/materials/${seeded.detail.material.id}/revisions/1/content`)).text, paperText);
  const replay = await api.request('/materials/imports', { method: 'POST', body: input, key });
  assert.deepEqual(replay, original);
  assert.equal((await api.get(`/jobs/${job.job_id}`)).operation_id, job.operation_id);
  assert.equal((await api.get('/missions')).items.length, 0);
});
