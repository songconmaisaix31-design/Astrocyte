import assert from 'node:assert/strict';
import { after, before, test } from 'node:test';
import { readFile, writeFile } from 'node:fs/promises';
import { join } from 'node:path';
import { randomUUID } from 'node:crypto';
import { startS1Server } from './server.mjs';
import { command, humanAPI, importMaterial, waitJob } from './api.mjs';
import { changedPaperText, distillation, opportunity, paperImport, paperText, sourceRef, videoImport, videoTranscript } from './fixtures.mjs';
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
  const second = await importMaterial(api, { ...paperImport(locator, changedPaperText), refresh: true });
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
  // A -> B -> A reuses immutable A and preserves B as the current head (user 7A).
  const restored = await importMaterial(api, paperImport(locator));
  assert.equal(restored.detail.material.id, first.detail.material.id);
  assert.equal(restored.job.job_id, first.job.job_id);
  assert.equal(restored.job.material_revision, 1);
  assert.equal(restored.detail.revisions.length, 2);
  assert.equal(restored.detail.material.current_revision, 2);
  assert.equal((await api.get(`/materials/${first.detail.material.id}/revisions/1/content`)).text, paperText);
  assert.equal((await api.get(`/materials/${first.detail.material.id}/revisions/2/content`)).text, changedPaperText);
  const restoredRecord = await api.write('/distillations', body);
  assert.equal(restoredRecord.reused, true);
  assert.equal(restoredRecord.distillation.id, original.distillation.id);
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

test('AT02: Bilibili tracking and fragment aliases reuse one material and immutable revision', async () => {
  const locators = [
    'https://www.bilibili.com/video/BV1S1LOCAL01/',
    'https://www.bilibili.com/video/BV1S1LOCAL01/?spm_id_from=333.1#reply',
    'https://m.bilibili.com/video/BV1S1LOCAL01?p=1&spm=333.1',
  ];
  const imports = await Promise.all(locators.map(locator => importMaterial(api, videoImport(locator))));
  assert.equal(new Set(imports.map(imported => imported.detail.material.id)).size, 1);
  const id = imports[0].detail.material.id;
  assert.equal((await api.get(`/materials/${id}`)).revisions.length, 1);
  assert.equal(server.query('SELECT COUNT(*) AS n FROM attention_material_revisions WHERE material_id=?', id)[0].n, 1);
  assert.equal((await api.get(`/materials/${id}/revisions/1/content`)).text, videoTranscript);
});

test('page-only video diagnostic fails honestly and creates no Material', async () => {
  const locator = 'https://www.bilibili.com/video/BV1S1PAGEONLY/';
  const priorIDs = (await api.get('/materials')).items.map(item => item.id).sort();
  const pageOnly = {
    input: { url: locator },
    extracted: {
      url: locator, title: 'contract_local recommended page, not target video', content: 'contract_local recommendation titles only',
      transcriptSource: null, transcriptSegments: null,
      diagnostics: { strategy: 'html', transcript: { textProvided: false } },
    },
    summary: null, llm: null,
  };
  const receipt = await api.write('/materials/imports', { source_locator: locator, source_key: '', content_digest: '', kind: 'video', adapter: 'summarize_json', export_text: JSON.stringify(pageOnly), collection_reason: null }, { status: 202 });
  const failed = await waitJob(api, receipt.job_id, 'failed');
  assert.equal(failed.error.code, 'evidence_missing');
  assert.ok(failed.error.required_action);
  assert.equal(failed.material_id, null);
  assert.deepEqual((await api.get('/materials')).items.map(item => item.id).sort(), priorIDs);
  assert.equal(JSON.parse(server.query('SELECT data FROM attention_jobs WHERE id=?', failed.job_id)[0].data).status, 'failed');
});

test('AT03: Agent without material authorization is denied; human refresh never heats and explicit reread does', async () => {
  const imported = await importMaterial(api, paperImport('https://example.invalid/contract-local/attention'));
  await importMaterial(api, { ...paperImport('https://example.invalid/contract-local/attention', changedPaperText), refresh: true });
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

// e9a44a63 profile contract: wait for actual W1/W2 runtime before claiming this path passed.
test('explicit candidate ranking profile preserves immutable versions, rejects stale/partial settings, and keeps unknown scores unknown', async () => {
  const empty = await api.get('/attention-ranking-profile');
  assert.equal(empty.configured, false);
  assert.equal(empty.profile, null);
  assert.deepEqual(empty.versions, []);
  const seeded = await seedCandidate(api, 'https://example.invalid/contract-local/profile');
  const unknownID = seeded.candidate.opportunity.id;
  const initialUnknown = (await api.get('/opportunities')).items.find(item => item.id === unknownID);
  assert.ok(initialUnknown.composite_score == null, 'No profile or unknown dimensions must not produce zero');
  const weights = { goal_progress: 4, current_interest: 3, project_improvement: 2, originality: 1 };
  const first = await api.write('/attention-ranking-profile', { enabled: true, weights }, { method: 'PUT' });
  assert.equal(first.configured, true);
  assert.equal(first.profile.version, 1);
  assert.deepEqual(first.versions, [first.profile]);
  const knownIDs = [];
  for (const [name, value] of [['low', 0.1], ['high', 0.9]]) {
    const dimensions = Object.fromEntries(Object.keys(weights).map(key => [key, { value, reason: `contract_local explicit ${name} score; not real project evidence` }]));
    const result = await api.write('/opportunities', opportunity([sourceRef(seeded.detail.material)], seeded.records.map(record => record.id), { title: `contract_local ${name} ranked candidate`, dimensions }), { status: 201 });
    knownIDs.push(result.opportunity.id);
  }
  const listed = (await api.get('/opportunities')).items;
  assert.ok(listed.findIndex(item => item.id === knownIDs[1]) < listed.findIndex(item => item.id === knownIDs[0]));
  for (const id of knownIDs) {
    const item = listed.find(value => value.id === id);
    assert.equal(typeof item.composite_score, 'number');
    assert.equal(item.ranking_profile_version, 1);
  }
  const unknown = listed.find(item => item.id === unknownID);
  assert.ok(unknown.composite_score == null);
  for (const dimension of Object.values(unknown.dimensions)) assert.equal(dimension.value, null);
  for (const invalidWeights of [{ ...weights, originality: 0 }, { goal_progress: 4 }, { ...weights, current_interest: 5 }]) {
    const invalid = await api.request('/attention-ranking-profile', { method: 'PUT', body: command({ expected_version: 1, enabled: true, weights: invalidWeights }) });
    assert.equal(invalid.status, 400, JSON.stringify(invalid.data));
  }
  const changedWeights = { ...weights, goal_progress: 5 };
  const second = await api.write('/attention-ranking-profile', { expected_version: 1, enabled: true, weights: changedWeights }, { method: 'PUT' });
  assert.equal(second.profile.version, 2);
  assert.deepEqual(second.versions[0], first.profile);
  assert.deepEqual(second.versions[1].weights, changedWeights);
  const stale = await api.request('/attention-ranking-profile', { method: 'PUT', body: command({ expected_version: 1, enabled: false, weights }) });
  assert.equal(stale.status, 409, JSON.stringify(stale.data));
  const denied = await api.request('/attention-ranking-profile', { method: 'PUT', body: command({ expected_version: 2, enabled: false, weights: changedWeights }), headers: { Authorization: `Bearer ${agentToken}` } });
  assert.equal(denied.status, 403);
  const rows = server.query('SELECT version,data FROM attention_ranking_profile_revisions WHERE profile_id=? ORDER BY version', first.profile.id);
  assert.deepEqual(rows.map(row => row.version), [1, 2]);
  assert.deepEqual(JSON.parse(rows[0].data).weights, weights);
  await server.restart();
  api = await humanAPI(server.apiURL);
  assert.deepEqual(await api.get('/attention-ranking-profile'), second);
  const disabled = await api.write('/attention-ranking-profile', { expected_version: 2, enabled: false, weights: changedWeights }, { method: 'PUT' });
  assert.equal(disabled.profile.version, 3);
  assert.equal(disabled.versions.length, 3);
  for (const item of (await api.get('/opportunities')).items) assert.ok(item.composite_score == null, 'Disabling explicit ranking must restore manual ordering semantics');
  assert.equal((await api.get('/missions')).items.length, 0);
});

// Published 6e34233 human classification/@ contract; run only on its complete runtime integration.
test('default disabled automatic processing denies ungranted input without creating a fake generated record', async () => {
  const imported = await importMaterial(api, paperImport('https://example.invalid/contract-local/no-processor'));
  const jobsBefore = (await api.get('/jobs')).items;
  const recordsBefore = (await api.get('/distillations')).items;
  const status = await api.get('/distillations/processor');
  assert.equal(status.available, false);
  assert.equal(status.configuration_id, null);
  assert.equal(status.model, null);
  assert.deepEqual(status.allowed_source_keys, []);
  assert.ok(status.reason && status.required_action);
  const body = command({ input_refs: [sourceRef(imported.detail.material)], stage: 'content', processing_config: 'unconfigured:contract_local', question: 'contract_local: no external processor is enabled' });
  const result = await api.request('/distillations/jobs', { method: 'POST', body });
  assert.equal(result.status, 403, JSON.stringify(result.data));
  assert.equal(result.data.error.code, 'scope_denied');
  assert.ok(result.data.error.required_action);
  assert.deepEqual((await api.get('/jobs')).items, jobsBefore);
  assert.deepEqual((await api.get('/distillations')).items, recordsBefore);
  const denied = await api.request('/distillations/jobs', { method: 'POST', body, headers: { Authorization: `Bearer ${agentToken}` } });
  assert.equal(denied.status, 403);
});

test('actual ranking orders human attention and pinning; list refresh is neutral and explains configured strategy', async () => {
  const first = await importMaterial(api, paperImport('https://example.invalid/contract-local/rank-a'));
  const second = await importMaterial(api, paperImport('https://example.invalid/contract-local/rank-b'));
  await api.write(`/materials/${first.detail.material.id}/uses`, { expected_version: first.detail.material.version, action: 'reread' });
  const ordered = (await api.get('/materials')).items.filter(item => [first.detail.material.id, second.detail.material.id].includes(item.id));
  assert.equal(ordered[0].id, first.detail.material.id);
  assert.ok(ordered[0].attention_score > ordered[1].attention_score);
  for (const item of ordered) {
    assert.ok(item.ranking_strategy);
    assert.ok(item.ranking_reason);
    assert.ok(item.attention_half_life_seconds > 0);
    assert.ok(item.attention_weights.reread > 0);
  }
  const beforePin = await api.get(`/materials/${second.detail.material.id}`);
  await api.write(`/materials/${second.detail.material.id}`, { expected_version: beforePin.material.version, lifecycle: 'active', pinned: true, collection_reason: null }, { method: 'PATCH' });
  for (let index = 0; index < 3; index++) await api.get('/materials');
  const pinnedFirst = (await api.get('/materials')).items.filter(item => [first.detail.material.id, second.detail.material.id].includes(item.id));
  assert.equal(pinnedFirst[0].id, second.detail.material.id, 'Manual pin must be retained ahead of decaying activity');
  assert.equal(pinnedFirst[0].human_usage_count, beforePin.material.human_usage_count);
  assert.equal(pinnedFirst[1].human_usage_count, ordered[0].human_usage_count);
  assert.equal((await api.request('/materials', { headers: { Authorization: `Bearer ${agentToken}` } })).status, 403);
});

test('human domains and @ references pin the original revision without copying or granting Agent access', async () => {
  const domain = (await api.write('/material-domains', { title: 'contract_local 输入版本域', description: '人工分类；不授予读取范围' }, { status: 201 })).domain;
  const imported = await importMaterial(api, paperImport('https://example.invalid/contract-local/space-original'));
  const material = imported.detail.material;
  const originalRevision = imported.detail.revisions[0];
  const classified = await api.write(`/materials/${material.id}/domains`, { expected_version: material.version, domain_ids: [domain.id] }, { method: 'PUT' });
  assert.deepEqual(classified.material.domain_ids, [domain.id]);
  assert.equal(classified.material.current_revision, 1);
  assert.equal(classified.material.agent_usage_count, material.agent_usage_count);
  const staleClassification = await api.request(`/materials/${material.id}/domains`, { method: 'PUT', body: command({ expected_version: material.version, domain_ids: [] }) });
  assert.equal(staleClassification.status, 409, JSON.stringify(staleClassification.data));
  const space = (await api.write('/project-spaces', { title: 'contract_local 人工空间' }, { status: 201 })).space;
  assert.deepEqual(space.material_refs, []);
  const referenceCommand = command({ expected_version: space.version, material_id: material.id, revision: 1 });
  const referenceKey = randomUUID();
  const referenced = await api.request(`/project-spaces/${space.id}/references`, { method: 'POST', body: referenceCommand, key: referenceKey });
  assert.equal(referenced.status, 200);
  assert.deepEqual(await api.request(`/project-spaces/${space.id}/references`, { method: 'POST', body: referenceCommand, key: referenceKey }), referenced);
  assert.equal(referenced.data.space.material_refs.length, 1);
  assert.equal(referenced.data.space.material_refs[0].material_id, material.id);
  assert.equal(referenced.data.space.material_refs[0].revision, 1);
  const afterHumanActions = await api.get(`/materials/${material.id}`);
  await importMaterial(api, { ...paperImport(material.source_locator, changedPaperText), refresh: true });
  const updated = await api.get(`/materials/${material.id}`);
  assert.deepEqual(updated.material.domain_ids, [domain.id], '@ must retain the original classification');
  assert.deepEqual(updated.revisions[0], originalRevision);
  assert.equal(updated.material.current_revision, 2);
  assert.equal(updated.material.human_usage_count, afterHumanActions.material.human_usage_count, 'Imports and read-only inspection must not add human signals');
  const savedSpace = (await api.get(`/project-spaces/${space.id}`)).space;
  assert.equal(savedSpace.material_refs[0].revision, 1, 'An @ reference must not silently track new source content');
  assert.equal((await api.get(`/materials/${savedSpace.material_refs[0].material_id}/revisions/1/content`)).text, paperText);
  assert.equal(server.query('SELECT COUNT(*) AS n FROM attention_materials WHERE source_key=?', originalRevision.source_key)[0].n, 1);
  assert.equal(server.query('SELECT COUNT(*) AS n FROM attention_material_revisions WHERE material_id=?', material.id)[0].n, 2);
  assert.deepEqual(JSON.parse(server.query('SELECT data FROM attention_project_spaces WHERE id=?', space.id)[0].data), savedSpace);
  for (const path of ['/material-domains', '/project-spaces', `/project-spaces/${space.id}`, `/materials/${material.id}/revisions/1/content`]) {
    const denied = await api.request(path, { headers: { Authorization: `Bearer ${agentToken}` } });
    assert.equal(denied.status, 403, `${path}: ${JSON.stringify(denied.data)}`);
  }
  await server.restart();
  api = await humanAPI(server.apiURL);
  assert.deepEqual((await api.get(`/project-spaces/${space.id}`)).space, savedSpace);
  assert.ok((await api.get('/material-domains')).items.some(item => item.id === domain.id));
  const staleRemove = await api.request(`/project-spaces/${space.id}/references/remove`, { method: 'POST', body: command({ expected_version: space.version, material_id: material.id }) });
  assert.equal(staleRemove.status, 409, JSON.stringify(staleRemove.data));
  const removed = await api.write(`/project-spaces/${space.id}/references/remove`, { expected_version: savedSpace.version, material_id: material.id });
  assert.deepEqual(removed.space.material_refs, []);
  assert.deepEqual((await api.get(`/materials/${material.id}`)).material.domain_ids, [domain.id]);
  assert.equal((await api.get(`/materials/${material.id}/revisions/1/content`)).text, paperText);
  assert.equal((await api.get('/missions')).items.length, 0);
});
