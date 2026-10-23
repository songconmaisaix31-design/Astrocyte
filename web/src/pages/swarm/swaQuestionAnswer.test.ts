import { describe, expect, it } from 'vitest';
import { createElement } from 'react';
import { renderToStaticMarkup } from 'react-dom/server';
import { SwaQuestionAnswerForm } from './SwaRunWorkspace';

// 纯表单边界；生产 SQLite 问题/答复与恢复器另由 question-answer.test.mjs 直接验证。
const question = { id: 'form-question', taskID: 'form-task', dispatchID: 'form-dispatch', question: '按哪个公开要求继续？' };
const onChange = () => { throw new Error('静态呈现不能改写用户回答'); };
const onSubmit = () => { throw new Error('静态呈现不能提交请求'); };
const render = (text: string, disabled = false) => renderToStaticMarkup(createElement(SwaQuestionAnswerForm, { question, text, disabled, onChange, onSubmit }));

describe('同 Run 问题回答表单', () => {
  it('保留正文空白与换行，展示实际问题身份和 UTF-8 字节数', () => {
    const body = '  按原公开要求继续。\n保留正文与换行。 \n';
    const html = render(body);
    expect(html).toContain(body);
    expect(html).toContain(question.id);
    expect(html).toContain(question.dispatchID);
    expect(html).toContain(`${new TextEncoder().encode(body).length}/32000`);
    expect(html).not.toContain('disabled=""');
  });

  it('旧权限、未知请求、空白及多字节越界时禁止提交，禁用不会丢弃正文', () => {
    const body = '需要用户确认的原回答';
    const locked = render(body, true);
    expect(locked).toContain(body);
    expect(locked).toContain('<textarea required="" disabled="">');
    expect(locked).toMatch(/<button[^>]*disabled=""/);
    for (const invalid of ['', ' \n\t', '界'.repeat(10667)]) expect(render(invalid)).toMatch(/<button[^>]*disabled=""/);
    expect(render('界'.repeat(10666) + 'ab')).not.toContain('disabled=""');
  });
});
