import { describe, expect, it } from 'vitest';
import { renderToStaticMarkup } from 'react-dom/server';
import { JudgmentPolicyFields, type JudgmentPolicyDraft } from './JudgmentPolicyFields';
import { policyDraftIssues } from './policyDraftIssues';

// 仅验证明确表单参数的显示与输入约束，不代表实际策略、后端或调用。
const emptyDraft: JudgmentPolicyDraft = {
  backend: '', languages: '', modelRevision: '', policyVersion: '', externalScope: '',
  timeoutMS: '', maxInputBytes: '', maxInputTokens: '', maxOutputTokens: '',
  minConfidence: '', truthAcceptMin: '', truthAcceptMax: '',
};
const change = () => { throw new Error('静态显示测试不得触发保存'); };

describe('判断策略表单参数', () => {
  it('保留明确选择、空预算与未知置信信息', () => {
    const html = renderToStaticMarkup(<JudgmentPolicyFields draft={emptyDraft} questionType="" disabled={false} onChange={change} />);
    expect(html).toContain('请选择明确的判断后端');
    expect(html).toContain('未提供模型置信信息时保留未知');
    expect(html).toContain('max="9007199254740991"');
    expect(policyDraftIssues(emptyDraft, '')).toEqual([]);
  });

  it('拒绝预算中的小数、溢出、负数和零输入上限', () => {
    expect(policyDraftIssues({ ...emptyDraft, timeoutMS: '0.5', maxInputBytes: '9007199254740992', maxInputTokens: '0', maxOutputTokens: '-1' }, 'choice')).toHaveLength(4);
    expect(policyDraftIssues({ ...emptyDraft, timeoutMS: '1', maxInputBytes: '1', maxInputTokens: '1', maxOutputTokens: '0' }, 'choice')).toEqual([]);
  });

  it('提示重复语言、范围边界和非有限置信输入', () => {
    expect(policyDraftIssues({ ...emptyDraft, languages: '中文\n 中文 ', minConfidence: 'Infinity', truthAcceptMin: '0' }, 'truth')).toEqual([
      '允许的判断语言重复，请逐项核查。',
      '模型置信下限需要处于 0 至 1 范围。',
      '真实性判断的可接受范围需要同时填写两个边界。',
    ]);
    expect(policyDraftIssues({ ...emptyDraft, truthAcceptMin: '0.9', truthAcceptMax: '0.1' }, 'truth')).toContain('真实性判断数值下限不得高于上限。');
  });

  it('题型变化保留原值并提示复核，禁用时阻止编辑', () => {
    const draft = { ...emptyDraft, backend: 'jev' as const, externalScope: '明确输入的范围', truthAcceptMin: '0', truthAcceptMax: '1' };
    const html = renderToStaticMarkup(<JudgmentPolicyFields draft={draft} questionType="choice" disabled onChange={change} />);
    expect(draft.truthAcceptMin).toBe('0');
    expect(html).toContain('role="alert"');
    expect(html).toContain('原真实性范围仍保留');
    expect(html).toContain('<fieldset');
    expect(html).toContain('disabled=""');
    expect(html).toContain('明确输入的范围');
    expect(html).toContain('服务仍核查每项资料的实际外发许可');
  });
});
