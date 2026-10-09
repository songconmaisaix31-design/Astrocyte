import type { components } from '../../api/schema';

export type Material = components['schemas']['MaterialV1'];
export type Opportunity = components['schemas']['OpportunityV1'];
export type Distillation = components['schemas']['DistillationV1'];
export type SourceRef = components['schemas']['SourceRefV1'];
export type Provenance = components['schemas']['ProvenanceV1'];
export type Dimensions = components['schemas']['DimensionsV1'];
export const stages = [{ value: 'content', label: '内容整理' }, { value: 'topic', label: '主题关联' }, { value: 'project', label: '项目关联' }];
export const feedbacks = [{ value: 'later', label: '以后再做' }, { value: 'reject', label: '拒绝' }, { value: 'revise', label: '修订' }, { value: 'adopt', label: '采用' }, { value: 'already_solved', label: '已解决' }];
export const dimensionLabels = { goal_progress: '目标进展', current_interest: '当前兴趣', project_improvement: '项目改善', originality: '创新性' };
export const lines = (text: string) => text.split('\n').map(line => line.trim()).filter(Boolean);
export function sourceRef(material: Material, revision = material.current_revision): SourceRef {
  return { material_id: material.id, revision, locator: material.source_locator, span: null };
}
export const sourceIdentity = (ref: SourceRef) => `${ref.material_id}@${ref.revision}:${ref.locator}:${ref.span ?? ''}`;
export const provenanceLabel = (p: Provenance) => ({ manual: '人工整理', arxiv: 'arXiv 导入', summarize_export: 'summarize 既有导出', summarize_extract: 'summarize 提取' })[p.mode];
export const unknownDimensions = (): Dimensions => ({ goal_progress: { value: null, reason: 'unknown' }, current_interest: { value: null, reason: 'unknown' }, project_improvement: { value: null, reason: 'unknown' }, originality: { value: null, reason: 'unknown' } });
