import type { TasteCommand, TasteProfile, TasteRubric, TasteExample, TasteDelegation, DelegationRule } from './judgmentApi';

export type RubricDraft = Omit<TasteRubric, 'question_type'> & { question_type: TasteRubric['question_type'] | '' };
export interface TasteDraft {
  enabled: boolean | null;
  rubrics: RubricDraft[];
  examples: TasteExample[];
  delegations: TasteDelegation[];
}

export function profileToDraft(profile: TasteProfile | null): TasteDraft {
  return profile ? structuredClone({ enabled: profile.enabled, rubrics: profile.rubrics, examples: profile.examples, delegations: profile.delegations }) : { enabled: null, rubrics: [], examples: [], delegations: [] };
}

export function tasteInput(draft: TasteDraft, rules: DelegationRule[]): Omit<TasteCommand, 'request_id' | 'expected_version'> {
  if (draft.enabled === null) throw new Error('请明确选择启用或停用');
  if (draft.rubrics.length > 256 || draft.examples.length > 256 || draft.delegations.length > 256) throw new Error('准则、示例和委托引用分别最多 256 项');
  if (draft.enabled && (!draft.rubrics.length || !draft.examples.length)) throw new Error('启用配置需要至少一条评判准则和一个固定资料示例');
  const rubrics: TasteRubric[] = draft.rubrics.map(rubric => {
    if (!rubric.id.trim() || !rubric.criteria.trim() || !rubric.question_type) throw new Error('请完整填写评判准则并明确选择题型');
    if (rubric.question_type === 'choice') {
      if (rubric.options.length < 2 || rubric.options.length > 255 || rubric.levels.length) throw new Error('选择题需要 2 至 255 个选项');
      if (rubric.options.some(option => !option.id.trim() || !option.description.trim()) || new Set(rubric.options.map(option => option.id)).size !== rubric.options.length) throw new Error('选项标识必须唯一，标识和说明均须填写');
    } else if (rubric.question_type === 'score') {
      if (rubric.options.length || rubric.levels.length < 2 || rubric.levels.length > 10 || rubric.levels.some(level => !level.trim())) throw new Error('评分题需要 2 至 10 个非空等级');
    } else if (rubric.options.length || rubric.levels.length) throw new Error('确定性题的选项与等级必须为空');
    return { ...rubric, question_type: rubric.question_type };
  });
  if (new Set(rubrics.map(rubric => rubric.id)).size !== rubrics.length) throw new Error('评判准则标识不能重复');
  for (const example of draft.examples) {
    const rubric = rubrics.find(item => item.id === example.rubric_id);
    if (!rubric || !example.material_id || !Number.isSafeInteger(example.revision) || example.revision < 1 || !example.content_digest.trim()) throw new Error('请为每个示例明确选择准则及实际资料的固定版本');
    if (rubric.question_type === 'choice') {
      if (!rubric.options.some(option => option.id === example.expected)) throw new Error('请选择该示例的实际预期选项');
    } else {
      if (!example.expected.trim()) throw new Error('请填写示例的预期数值');
      const value: unknown = JSON.parse(example.expected);
      const limit = rubric.question_type === 'score' ? rubric.levels.length - 1 : 1;
      if (typeof value !== 'number' || !Number.isFinite(value) || value < 0 || value > limit) throw new Error(`示例预期值必须是 0 至 ${limit} 的数值`);
    }
  }
  for (const reference of draft.delegations) {
    if (!rules.some(rule => rule.id === reference.rule_id && rule.revision === reference.revision && rule.workflow_id === reference.workflow_id)) throw new Error('委托引用当前不可读取，请明确选择实际规则版本');
  }
  if (new Set(draft.delegations.map(reference => reference.rule_id)).size !== draft.delegations.length) throw new Error('同一委托规则只能引用一次');
  return structuredClone({ schema_version: 1, enabled: draft.enabled, rubrics, examples: draft.examples, delegations: draft.delegations });
}
