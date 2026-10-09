// These are synthetic contract_local records. They are never public paper/video evidence.
export const paperText = 'contract_local paper representative\nSection 1: queues preserve original input versions.\nSection 2: retries should reuse the same operation.';
export const changedPaperText = `${paperText}\nSection 3: an updated source retains both revisions.`;
export const videoTranscript = 'contract_local video representative\n[00:00-00:12] Human attention and machine retrieval are separate.\n[00:12-00:25] Deferring a candidate preserves its evidence.';

export function paperImport(locator = 'https://example.invalid/contract-local/paper', text = paperText) {
  return {
    source_locator: locator, source_key: '', content_digest: '', kind: 'paper', adapter: 'manual',
    export_text: text, title: 'contract_local 论文代表数据', source_spans: ['Section 1', 'Section 2'], collection_reason: null,
  };
}

export function videoImport(locator = 'https://www.youtube.com/watch?v=S1LOCAL001a', { summaryOnly = false } = {}) {
  return {
    source_locator: locator, source_key: '', content_digest: '', kind: 'video', adapter: 'summarize_json',
    export_text: JSON.stringify({
      input: { url: locator },
      extracted: {
        url: locator, title: 'contract_local 视频代表导出', content: summaryOnly ? '' : videoTranscript,
        transcriptSegments: summaryOnly ? [] : [
          { startMs: 0, endMs: 12000, text: 'Human attention and machine retrieval are separate.' },
          { startMs: 12000, endMs: 25000, text: 'Deferring a candidate preserves its evidence.' },
        ],
      },
      summary: 'contract_local synthetic summary; not a real summarize execution.',
    }),
    collection_reason: null,
  };
}

export function sourceRef(material, revision = material.current_revision, span = 'Section 1') {
  return { material_id: material.id, revision, locator: material.source_locator, span };
}

export function distillation(stage, refs, overrides = {}) {
  const base = {
    schema_version: 1, expected_version: 1,
    input_refs: refs, stage, output_text: `contract_local manual ${stage} output`,
    question: `contract_local ${stage} question`, processing_config: 'manual:contract_local:v1',
  };
  const stages = {
    content: {},
    topic: { related_ideas: ['contract_local reusable input versions'], pending_questions: ['Which source version changed?'] },
    project: {
      goal_refs: ['contract_local-goal'], existing_assets: ['contract_local existing queue'],
      expected_improvement: 'Keep source versions accessible after retry', minimum_artifact: 'A bounded queue comparison',
      missing_evidence: ['No real project performance measurement'],
    },
  };
  return { ...base, ...stages[stage], ...overrides };
}

export function opportunity(refs, distillationIDs, overrides = {}) {
  const unknown = { value: null, reason: 'contract_local: evidence not measured' };
  return {
    schema_version: 1, expected_version: 1,
    title: 'contract_local 候选：保留来源版本', evidence_refs: refs, goal_refs: ['contract_local-goal'],
    dimensions: { goal_progress: { ...unknown }, current_interest: { ...unknown }, project_improvement: { ...unknown }, originality: { ...unknown } },
    next_step: '人工对照两版输入，再决定是否继续', purpose: '检验来源版本与延后反馈',
    missing_evidence: ['真实项目收益尚未测量'], distillation_ids: distillationIDs,
    ...overrides,
  };
}
