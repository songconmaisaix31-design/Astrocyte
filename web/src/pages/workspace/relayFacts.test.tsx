import { renderToStaticMarkup } from 'react-dom/server';
import { expect, it } from 'vitest';
import type { RelayFacts } from '../../api/relay';
import { RelayRuntimeFacts } from './relayFacts';

const input: RelayFacts = { schema_version: 1, task_id: 'explicit-input-task', facts: [], session_relations: [], current_ownership: { status: 'unknown', reason: 'controller_binding_unrecorded' } };

it('明确空输入保持归属、输入与会话关系未知', () => {
  const html = renderToStaticMarkup(<RelayRuntimeFacts data={input} />);
  expect(html).toContain('当前原控制器归属未知');
  expect(html).toContain('输入与原生执行证据未知');
  expect(html).toContain('不能据历史终端或工作区推断交接');
});

it('明确写入事实输入保留原收据字符串且不生成开始事实', () => {
  const data: RelayFacts = { ...input, facts: [{ fact_id: 'explicit-input-fact', task_id: input.task_id, kind: 'input_written', status: 'confirmed', source: 'owned_stdin_pipe', operation_id: 'explicit-input-operation', receipt_id: 'human:local-human:send:explicit-input-operation', sequence: 1, ts: '2026-10-15T00:00:00Z' }] };
  const html = renderToStaticMarkup(<RelayRuntimeFacts data={data} />);
  expect(html).toContain('输入通道写入');
  expect(html).toContain(data.facts[0].receipt_id);
  expect(html).not.toContain('<strong>原生回合开始</strong>');
  expect(html).toContain('缺少的初始化、开始或消费证据保持未知');
});

it('明确人工交接关系输入不生成新执行者会话证明', () => {
  const data: RelayFacts = { ...input, session_relations: [{ task_id: input.task_id, previous_task_id: 'explicit-input-previous', operation_id: 'explicit-input-operation', kind: 'handoff', status: 'confirmed', source: 'human_intent', ts: '2026-10-15T00:00:00Z' }] };
  const html = renderToStaticMarkup(<RelayRuntimeFacts data={data} />);
  expect(html).toContain('跨 Agent 上下文交接');
  expect(html).toContain('explicit-input-previous');
  expect(html).toContain('新执行者的原生会话确认仍须另行核查');
  expect(html).toContain('当前原控制器归属未知');
});
