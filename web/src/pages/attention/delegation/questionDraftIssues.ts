import type { JudgmentQuestionDraft } from './JudgmentQuestionFields';

export function questionDraftIssues(draft: JudgmentQuestionDraft): string[] {
  if (!draft.type) return [];
  const issues: string[] = [];
  if (!draft.id.trim()) issues.push('判断题目标识需要填写完整内容。');
  if (!draft.instructions.trim()) issues.push('判断题目与要求需要填写完整内容。');
  if (draft.type === 'choice') {
    if (draft.options.length < 2 || draft.options.length > 255) issues.push('选项判断需要填写 2 至 255 个选项。');
    const ids = draft.options.map(option => option.id.trim());
    if (ids.some(id => !id)) issues.push('每个判断选项都需要填写完整标识。');
    const populated = ids.filter(Boolean);
    if (new Set(populated).size !== populated.length) issues.push('判断选项的标识重复，请为各选项填写独立标识。');
    if (draft.rule === 'single_supported_option' && draft.options.some(option => option.evidenceKeys.length === 0)) {
      issues.push('依据齐全的唯一选项规则要求每个选项选择固定依据。');
    }
  }
  if (draft.type === 'score') {
    if (draft.levels.length < 2 || draft.levels.length > 10) issues.push('分级评分需要填写 2 至 10 个有序等级。');
    if (draft.levels.some(level => !level.trim())) issues.push('每个评分等级都需要填写完整说明。');
  }
  return issues;
}
