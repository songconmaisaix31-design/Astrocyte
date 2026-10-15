import { chromium, expect } from '@playwright/test';
import { readFile, writeFile } from 'node:fs/promises';
import { resolve } from 'node:path';
import { DatabaseSync } from 'node:sqlite';
import { humanAPI } from '../../tests/s1/api.mjs';
import type { components } from '../src/api/schema';

type S = components['schemas'];
type Operation = { operation_id: string; status: number };
const origin = process.env.ASTROCYTE_TEST_PRODUCT_URL;
const source = process.env.ASTROCYTE_TEST_PRODUCT_SOURCE;
const output = process.env.ASTROCYTE_VERIFICATION_OUTPUT_DIR;
if (origin !== 'http://127.0.0.1:64505' || !source || !/^[0-9a-f]{40}$/.test(source) || !output) throw new Error('需要当前同实例SOURCE及原公开候选结果目录');
const directory = resolve(output);
const original = JSON.parse(await readFile(resolve(directory, 'product-candidate-human.json'), 'utf8')) as {
  material_id: string; original_excerpt_id: string; topic_id: string; candidate_id: string; operations: Operation[];
};
const revision = JSON.parse(await readFile(resolve(directory, 'product-candidate-revision.json'), 'utf8')) as {
  candidate_id: string; operations: Operation[]; events: { id: string; type: string; aggregate_id: string }[];
};
expect(original.operations.map(item => item.status)).toEqual([200, 201, 201]);
expect(revision.operations.map(item => item.status)).toEqual([200]);
expect(revision.candidate_id).toBe(original.candidate_id);
const database = new DatabaseSync('D:/Astrocyte-agent-assets/20261015/relay/H/shared-product-ad8413/local/astrocyte-workbench/v0.2-access/app-state/state.sqlite', { readOnly: true });
const browser = await chromium.launch();
try {
  const api = await humanAPI(origin);
  const material = await api.get(`/materials/${original.material_id}`) as S['MaterialDetailV1'];
  const excerpt = material.distillations.find(item => item.id === original.original_excerpt_id);
  const topic = material.distillations.find(item => item.id === original.topic_id);
  if (!excerpt || !topic) throw new Error('原摘要或主题记录不存在，禁止重新创建');
  expect(topic.output_text).toBe(excerpt.output_text.trim());
  expect(topic.pending_questions).toEqual(['当前公开摘要与本项目的实际改进之间仍缺少任务结果和独立效果证据。']);
  expect(topic.input_refs.map(item => ({ material_id: item.material_id, revision: item.revision }))).toEqual([{ material_id: original.material_id, revision: 1 }]);
  expect(topic.provenance).toMatchObject({ processor: 'human', mode: 'manual' });
  const candidate = await api.get(`/opportunities/${original.candidate_id}`) as S['OpportunityDetailV1'];
  expect(candidate.opportunity).toMatchObject({ revision: 2, version: 3, state: 'ready_for_review', distillation_ids: [topic.id] });
  expect(candidate.revisions.map(item => item.revision)).toEqual([1, 2]);
  expect(candidate.reviews).toHaveLength(1);
  expect(candidate.reviews[0]).toMatchObject({ revision: 1, feedback: 'later', reason: '当前只保存公开来源候选，实际适用性和效果仍待后续核查。' });
  expect(candidate.revisions[0].evidence_refs).toEqual(candidate.revisions[1].evidence_refs);
  expect(candidate.opportunity.evidence_refs.map(item => ({ material_id: item.material_id, revision: item.revision }))).toEqual([{ material_id: original.material_id, revision: 1 }]);
  expect(candidate.revisions[0].dimensions).toEqual(candidate.revisions[1].dimensions);
  expect(Object.values(candidate.opportunity.dimensions)).toEqual(Array.from({ length: 4 }, () => ({ value: null, reason: '尚未进行实际任务和效果评估。' })));
  const saved = database.prepare('SELECT data FROM attention_opportunities WHERE id=?').get(original.candidate_id) as { data: string };
  const stored = JSON.parse(saved.data) as S['OpportunityDetailV1'];
  expect((await api.get('/attention-ranking-profile') as S['RankingProfileDetailV1']).configured).toBe(false);
  expect(candidate).toEqual({ ...stored, opportunity: { ...stored.opportunity,
    ranking_strategy: 'manual', ranking_reason: '未配置完整且启用的四维 profile，保留人工顺序',
  } });
  const operations = [...original.operations, ...revision.operations];
  expect(new Set(operations.map(item => item.operation_id)).size).toBe(4);
  for (const operation of operations) expect(database.prepare('SELECT key FROM attention_receipts WHERE key=?').get(operation.operation_id)).toEqual({ key: operation.operation_id });
  expect(revision.events.map(event => event.type).sort()).toEqual(['attention.distillation_recorded', 'attention.opportunity_created', 'attention.opportunity_reviewed', 'attention.opportunity_revised'].sort());
  for (const event of revision.events) expect(database.prepare('SELECT id,type,aggregate_id FROM attention_outbox WHERE id=?').get(event.id)).toEqual(event);
  expect(database.prepare('SELECT COUNT(*) AS count FROM workspace_approvals').get()).toEqual({ count: 1 });
  expect(database.prepare('SELECT COUNT(*) AS count FROM swarm_missions').get()).toEqual({ count: 0 });
  expect(database.prepare('SELECT COUNT(*) AS count FROM local_agent_sessions').get()).toEqual({ count: 0 });
  const page = await browser.newPage({ viewport: { width: 1280, height: 720 } });
  const writes: string[] = [];
  page.on('request', request => { if (new URL(request.url()).pathname.startsWith('/api/v1/') && request.method() !== 'GET') writes.push(request.method()); });
  await page.goto(`${origin}/attention`);
  await page.locator('main li[role="button"]').filter({ has: page.getByText(candidate.opportunity.title!, { exact: true }) }).click();
  const dialog = page.getByRole('dialog', { name: '机会详情', exact: true });
  await expect(dialog.getByText('以后再做 · r1', { exact: false })).toBeVisible();
  await dialog.getByLabel('候选历史版本', { exact: true }).selectOption('1');
  const missing = candidate.revisions[0].missing_evidence?.[0];
  if (!missing) throw new Error('原候选版本缺少已保存缺失依据');
  await expect(dialog.getByRole('region', { name: '候选版本历史', exact: true })).toContainText(missing);
  await dialog.getByLabel('候选历史版本', { exact: true }).selectOption('2');
  await expect(dialog).toContainText('人工暂缓记录已保存；实际任务、模型处理与独立效果证据仍未产生。');
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
  expect(writes).toEqual([]);
  const result = { source, material_id: original.material_id, topic_id: topic.id, candidate_id: candidate.opportunity.id,
    revision: candidate.opportunity.revision, version: candidate.opportunity.version, state: candidate.opportunity.state,
    title: candidate.opportunity.title, purpose: candidate.opportunity.purpose, next_step: candidate.opportunity.next_step,
    missing_evidence: candidate.opportunity.missing_evidence, fixed_source: candidate.opportunity.evidence_refs,
    goal_refs: candidate.opportunity.goal_refs, user_admission_decision: 'pending',
    revisions: candidate.revisions.map(item => ({ revision: item.revision, title: item.title, state: item.state,
      purpose: item.purpose, next_step: item.next_step, missing_evidence: item.missing_evidence, evidence_refs: item.evidence_refs })),
    four_dimensions_unknown: true, reviews: candidate.reviews.map(item => ({ id: item.id, revision: item.revision, feedback: item.feedback })),
    operations, events: revision.events, native_sessions: 0, mission_count: 0, approval_count: 1,
    writes, completed_at: new Date().toISOString() };
  await writeFile(resolve(directory, 'product-candidate-readonly.json'), JSON.stringify(result, null, 2), { flag: 'wx' });
  console.log(JSON.stringify(result));
} finally { await browser.close(); database.close(); }
