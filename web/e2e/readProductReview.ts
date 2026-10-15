import { expect } from '@playwright/test';
import { readFile, writeFile } from 'node:fs/promises';
import { resolve } from 'node:path';
import { humanAPI } from '../../tests/s1/api.mjs';
import type { components } from '../src/api/schema';

type S = components['schemas'];
const origin = process.env.ASTROCYTE_TEST_PRODUCT_URL;
const source = process.env.ASTROCYTE_TEST_PRODUCT_SOURCE;
const output = process.env.ASTROCYTE_VERIFICATION_OUTPUT_DIR;
if (origin !== 'http://127.0.0.1:64505' || !source || !output) throw new Error('需要同一产品实际SOURCE和原候选记录目录');
const directory = resolve(output);
const original = JSON.parse(await readFile(resolve(directory, 'product-candidate-readonly.json'), 'utf8')) as { candidate_id: string; revision: number; version: number; topic_id: string };
const api = await humanAPI(origin);
const candidate = await api.get(`/opportunities/${original.candidate_id}`) as S['OpportunityDetailV1'];
expect(candidate.opportunity).toMatchObject({ revision: original.revision, version: original.version, state: 'ready_for_review' });
expect(Object.values(candidate.opportunity.dimensions).every(value => value.value === null)).toBe(true);
expect(candidate.opportunity.goal_refs).toHaveLength(1);
const { goal } = await api.get(`/goals/${candidate.opportunity.goal_refs![0]}`) as S['S2GoalResultV1'];
const project = (await api.get('/local-projects') as S['LocalProjectListV1']).items.find(item => item.id === goal.project_id);
if (!project) throw new Error('关联Goal的实际项目不存在');
expect(candidate.opportunity.evidence_refs).toHaveLength(1);
const fixed = candidate.opportunity.evidence_refs[0];
const material = await api.get(`/materials/${fixed.material_id}`) as S['MaterialDetailV1'];
expect(material.revisions.some(item => item.revision === fixed.revision)).toBe(true);
const excerpt = material.distillations.find(item => item.id === 'LD6A7RPZQJIQW6F224QALLQDR5');
if (!excerpt) throw new Error('原已存人工摘要不存在，禁止生成替代摘要');
expect(excerpt.provenance).toMatchObject({ processor: 'human', mode: 'manual' });
expect(excerpt.input_refs.map(item => ({ material_id: item.material_id, revision: item.revision }))).toEqual([{ material_id: fixed.material_id, revision: fixed.revision }]);
const result = {
  source, candidate: { id: candidate.opportunity.id, revision: candidate.opportunity.revision, version: candidate.opportunity.version,
    state: candidate.opportunity.state, title: candidate.opportunity.title, purpose: candidate.opportunity.purpose,
    next_step: candidate.opportunity.next_step, missing_evidence: candidate.opportunity.missing_evidence, dimensions: candidate.opportunity.dimensions },
  paper: { title: material.material.title, reference: fixed, manual_abstract_id: excerpt.id,
    manual_abstract: excerpt.output_text, provenance: excerpt.provenance },
  goal: { id: goal.id, revision: goal.revision, title: goal.title, success_criteria: goal.success_criteria },
  project: { id: project.id, name: project.name, space_id: project.space_id, settings_revision: project.settings.revision },
  admission: { decision: 'pending', production_contract: 'awaiting_authoritative_generated_contract', request_sent: false },
  writes: 0, new_model_turns: 0, completed_at: new Date().toISOString(),
};
await writeFile(resolve(directory, 'product-candidate-review.json'), JSON.stringify(result, null, 2), { flag: 'wx' });
console.log(JSON.stringify(result));
