import type { JudgmentPolicyDraft } from './JudgmentPolicyFields';

export function policyDraftIssues(draft: JudgmentPolicyDraft, questionType: '' | 'choice' | 'score' | 'truth'): string[] {
  const issues: string[] = [];
  const languages = draft.languages.split('\n').map(value => value.trim()).filter(Boolean);
  if (new Set(languages).size !== languages.length) issues.push('允许的判断语言重复，请逐项核查。');
  const budgets = [
    { value: draft.timeoutMS, label: '判断超时上限', minimum: 1 },
    { value: draft.maxInputBytes, label: '判断输入字节上限', minimum: 1 },
    { value: draft.maxInputTokens, label: '判断输入 Token 上限', minimum: 1 },
    { value: draft.maxOutputTokens, label: '判断输出 Token 上限', minimum: 0 },
  ];
  for (const budget of budgets) {
    if (budget.value.trim() && (!Number.isSafeInteger(Number(budget.value)) || Number(budget.value) < budget.minimum)) {
      issues.push(`${budget.label}需要填写${budget.minimum === 0 ? '非负' : '正'}安全整数。`);
    }
  }
  const thresholds = [
    { value: draft.minConfidence, label: '模型置信下限' },
    { value: draft.truthAcceptMin, label: '真实性判断数值下限' },
    { value: draft.truthAcceptMax, label: '真实性判断数值上限' },
  ];
  for (const threshold of thresholds) {
    const value = Number(threshold.value);
    if (threshold.value.trim() && (!Number.isFinite(value) || value < 0 || value > 1)) issues.push(`${threshold.label}需要处于 0 至 1 范围。`);
  }
  const hasMinimum = !!draft.truthAcceptMin.trim(), hasMaximum = !!draft.truthAcceptMax.trim();
  if (hasMinimum !== hasMaximum) issues.push('真实性判断的可接受范围需要同时填写两个边界。');
  if (hasMinimum && hasMaximum && Number(draft.truthAcceptMin) > Number(draft.truthAcceptMax)) issues.push('真实性判断数值下限不得高于上限。');
  if (questionType !== 'truth' && (hasMinimum || hasMaximum)) issues.push('当前题型已变化，原真实性范围仍保留。请核查题型和范围后再保存。');
  return issues;
}
