import { describe, expect, it } from 'vitest';
import { renderToStaticMarkup } from 'react-dom/server';
import { JudgmentQuestionFields, type JudgmentQuestionDraft } from './JudgmentQuestionFields';
import { questionDraftIssues } from './questionDraftIssues';

// 仅验证明确表单输入和显示约束，不代表真实来源、策略或后端调用。
const emptyDraft: JudgmentQuestionDraft = { id: '', type: '', instructions: '', rule: '', options: [], levels: [] };
const change = () => { throw new Error('静态参数检查不得修改题目'); };

describe('判断题目表单参数', () => {
  it('保留未选择的题型，选择后提示空白题目和不足的选项', () => {
    expect(questionDraftIssues(emptyDraft)).toEqual([]);
    const draft = { ...emptyDraft, type: 'choice' as const, id: ' ', instructions: '\n' };
    const html = renderToStaticMarkup(<JudgmentQuestionFields draft={draft} evidence={[]} disabled={false} onChange={change} />);
    expect(html).toContain('判断题目标识需要填写完整内容');
    expect(html).toContain('判断题目与要求需要填写完整内容');
    expect(html).toContain('选项判断需要填写 2 至 255 个选项');
    expect(html).toContain('role="alert"');
  });

  it('提示重复和空白标识，以及规则缺少必要依据', () => {
    const draft: JudgmentQuestionDraft = {
      ...emptyDraft, id: 'choice-input', type: 'choice', instructions: '检查当前明确填写的候选选项', rule: 'single_supported_option',
      options: [{ id: '重复', description: '', evidenceKeys: [] }, { id: ' 重复 ', description: '', evidenceKeys: [] }, { id: ' ', description: '', evidenceKeys: [] }],
    };
    expect(questionDraftIssues(draft)).toEqual([
      '每个判断选项都需要填写完整标识。',
      '判断选项的标识重复，请为各选项填写独立标识。',
      '依据齐全的唯一选项规则要求每个选项选择固定依据。',
    ]);
    const withoutRule = { ...draft, rule: 'none', options: [{ id: '保留', description: '', evidenceKeys: [] }, { id: '修改', description: '', evidenceKeys: [] }] };
    expect(questionDraftIssues(withoutRule)).toEqual([]);
  });

  it('检查评分等级的数量与空白说明，题型切换继续保留输入', () => {
    const draft: JudgmentQuestionDraft = { ...emptyDraft, id: 'score-input', type: 'score', instructions: '按明确等级评分', levels: ['\n'] };
    expect(questionDraftIssues(draft)).toEqual(['分级评分需要填写 2 至 10 个有序等级。', '每个评分等级都需要填写完整说明。']);
    expect(questionDraftIssues({ ...draft, levels: ['未达到', '已达到'] })).toEqual([]);
    const html = renderToStaticMarkup(<JudgmentQuestionFields draft={{ ...draft, type: 'truth' }} evidence={[]} disabled onChange={change} />);
    expect(draft.levels).toEqual(['\n']);
    expect(html).toContain('disabled=""');
    expect(html).toContain('按明确等级评分');
  });
});
