// Explicit public-network acceptance only. Not discovered by node --test or ordinary CI.
import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import { join } from 'node:path';
import { startS1Server } from './server.mjs';
import { humanAPI, importMaterial, waitJob } from './api.mjs';
import { distillation, opportunity, sourceRef } from './fixtures.mjs';

const exportIndex = process.argv.indexOf('--summarize-export');
if (exportIndex < 0 || !process.argv[exportIndex + 1]) throw new Error('Pass --summarize-export with the real W2 export path; no synthetic fallback');
const exported = await readFile(process.argv[exportIndex + 1], 'utf8');
const actualExport = JSON.parse(exported);
assert.equal(actualExport.input.url, 'https://www.bilibili.com/video/BV1PReT6EEqR/');
const requireFullText = process.argv.includes('--require-full-text');
if (requireFullText && !process.env.ASTROCYTE_SUMMARIZE_CLI) throw new Error('Set ASTROCYTE_SUMMARIZE_CLI to the trusted installed 0.21.8 absolute CLI path');
const server = await startS1Server({ env: requireFullText ? { ASTROCYTE_NODE: process.execPath, ASTROCYTE_SUMMARIZE_CLI: process.env.ASTROCYTE_SUMMARIZE_CLI } : {} });
try {
  const api = await humanAPI(server.apiURL);
  const paper = await importMaterial(api, {
    source_locator: 'https://arxiv.org/abs/2504.16054v1', source_key: '', content_digest: '', kind: 'paper', adapter: 'arxiv', collection_reason: null,
  }, { timeoutMs: 120_000 });
  const revision = paper.detail.revisions[0];
  assert.match(revision.provenance.version, /^2504\.16054v1(?:; summarize 0\.21\.8)?$/);
  const pdf = revision.attachments.find(attachment => attachment.media_type === 'application/pdf');
  assert.ok(pdf, 'Actual original paper PDF must be preserved');
  const pdfResponse = await fetch(`${server.apiURL}/api/v1/materials/${paper.detail.material.id}/revisions/1/attachments/${encodeURIComponent(pdf.name)}`, { headers: { Cookie: api.cookie }, signal: AbortSignal.timeout(30_000) });
  assert.equal(pdfResponse.status, 200);
  const bytes = Buffer.from(await pdfResponse.arrayBuffer());
  assert.ok(bytes.subarray(0, 5).equals(Buffer.from('%PDF-')));
  assert.deepEqual(bytes, await readFile(join(server.dataDir, 'objects', pdf.object_ref)));
  let extractedFullTextCharacters = null;
  if (requireFullText) {
    assert.equal(revision.provenance.mode, 'official_atom_pdf_and_html_text', 'A metadata-only import is not full paper text');
    const html = revision.attachments.find(attachment => attachment.name === '2504.16054v1.html');
    const originalExport = revision.attachments.find(attachment => attachment.name === 'summarize-original.json');
    assert.ok(html && originalExport, 'Official original HTML and genuine extraction export must be retained');
    const extracted = JSON.parse(await readFile(join(server.dataDir, 'objects', originalExport.object_ref), 'utf8'));
    assert.equal(extracted.input.url, 'https://arxiv.org/html/2504.16054v1');
    assert.equal(extracted.llm, null, 'The original text extraction must not call a model');
    assert.ok(extracted.extracted.content.length > 1000, 'The selected full paper must contain more than a metadata abstract');
    const paperContent = await api.get(`/materials/${paper.detail.material.id}/revisions/1/content`);
    assert.ok(paperContent.text.includes(extracted.extracted.content));
    assert.ok((await readFile(join(server.dataDir, 'objects', html.object_ref), 'utf8')).includes('2504.16054'));
    extractedFullTextCharacters = extracted.extracted.content.length;
  }
  const videoInput = {
    source_locator: actualExport.input.url, source_key: '', content_digest: '', kind: 'video', adapter: 'summarize', export_text: exported, collection_reason: null,
  };
  const hasTranscript = Array.isArray(actualExport.extracted.transcriptSegments) && actualExport.extracted.transcriptSegments.length > 0;
  let video;
  if (!hasTranscript && !actualExport.llm) {
    // The chosen real export contains recommended-page titles, not video evidence.
    const receipt = await api.write('/materials/imports', videoInput, { status: 202 });
    const failed = await waitJob(api, receipt.job_id, 'failed');
    assert.equal(failed.error?.code, 'evidence_missing');
    assert.equal(failed.material_id, null);
    assert.equal((await api.get('/materials')).items.some(item => item.kind === 'video'), false);
  } else {
    video = await importMaterial(api, videoInput);
    const videoContent = await api.get(`/materials/${video.detail.material.id}/revisions/1/content`);
    assert.equal(videoContent.text, actualExport.extracted.content);
    const videoSource = video.detail.revisions[0];
    if (!hasTranscript) assert.deepEqual(videoSource.source_spans, []);
    const exportedAttachment = videoSource.attachments.find(attachment => attachment.name === 'summarize.json');
    assert.ok(exportedAttachment);
    assert.equal(await readFile(join(server.dataDir, 'objects', exportedAttachment.object_ref), 'utf8'), exported);
  }

  // Honest manual traceability records, not a model summary or claims about paper/video findings.
  const refs = [sourceRef(paper.detail.material, 1, revision.source_spans[0]), ...(video ? [{ material_id: video.detail.material.id, revision: 1, locator: video.detail.material.source_locator }] : [])];
  const records = [];
  const missing = ['PDF 尚未逐页细读', ...(!hasTranscript ? ['视频只取得网页文本，字幕与时间位置缺失'] : []), '未选择具体项目改进任务'];
  for (const stage of ['content', 'topic', 'project']) {
    const output = stage === 'content'
      ? '人工采集记录：论文原始 PDF 与 metadata 分别保留；无有效视频证据时不创建视频资料。收藏理由未提供，不把网页文本当视频字幕。'
      : stage === 'topic'
        ? '人工待查关联：需要细读论文及取得视频字幕，才能评价这些来源是否支持当前项目改进。尚无研究结论。'
        : '人工项目关联：现有 Astrocyte 导入、SQLite 与对象层可保存来源；最小成果是可回查来源版本的候选。具体改进依据仍缺失。';
    const body = distillation(stage, refs, {
      question: `真实材料人工采集的 ${stage} 待查问题`, processing_config: 'manual:selected-public-sources:v1', output_text: output,
      ...(stage === 'topic' ? { related_ideas: [], pending_questions: missing } : {}),
      ...(stage === 'project' ? {
        goal_refs: ['Astrocyte:S1:material-to-candidate'], existing_assets: ['Astrocyte Go/SQLite/objects importer'],
        expected_improvement: '保存可回查来源的候选供人工继续研究', minimum_artifact: '一个带缺失依据的人工候选', missing_evidence: missing,
      } : {}),
    });
    records.push((await api.write('/distillations', body)).distillation);
  }
  const candidate = await api.write('/opportunities', opportunity(refs, records.map(record => record.id), {
    title: '真实公开材料：待查关联', purpose: '保存原文与采集限制，待人工补证', next_step: '细读 PDF 并补充视频字幕后人工决定下一步',
    goal_refs: ['Astrocyte:S1:material-to-candidate'], missing_evidence: missing,
    dimensions: Object.fromEntries(['goal_progress', 'current_interest', 'project_improvement', 'originality'].map(name => [name, { value: null, reason: '真实材料尚待人工补证；未评估' }])),
  }), { status: 201 });
  assert.equal((await api.get('/missions')).items.length, 0);
  await server.restart();
  const reconnected = await humanAPI(server.apiURL);
  assert.deepEqual((await reconnected.get(`/materials/${paper.detail.material.id}`)).revisions[0].attachments, revision.attachments);
  assert.equal((await reconnected.get(`/opportunities/${candidate.opportunity.id}`)).opportunity.id, candidate.opportunity.id);
  console.log(JSON.stringify({
    paper_import: 'PASS', paper_version: revision.provenance.version, original_pdf_bytes: bytes.length,
    paper_full_text: requireFullText ? 'PASS' : 'NOT_RUN', extracted_full_text_characters: extractedFullTextCharacters,
    video_existing_export_import: video ? 'PASS' : 'REJECTED_PAGE_ONLY', page_only_rejection: video ? 'NOT_APPLICABLE' : 'PASS', video_extracted_characters: actualExport.extracted.totalCharacters,
    video_transcript: hasTranscript ? 'PASS' : 'BLOCKED', video_summary: actualExport.summary ? 'available_in_export' : 'NOT_RUN',
    manual_traceability_layers: records.map(record => record.stage), automatic_distillation: 'NOT_RUN',
    model: actualExport.llm ?? null, full_AT01: hasTranscript ? 'NOT_RUN' : 'BLOCKED',
    reason: '只验导入/手工采集记录/来源回查/重启；原文细读与自动整理不因这些检查而通过',
  }, null, 2));
  process.exitCode = 2;
} finally { await server.close(); }
