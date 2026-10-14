import { expect, it } from 'vitest';
import { renderToStaticMarkup } from 'react-dom/server';
import type { RelayTaskView, RelayTaskResult } from '../../api/relay';
import { RelayTaskSummary } from './relayTaskPanel';
import { RelayTransitionControls } from './relayTransitionControls';
import { relayElapsed } from './relayPresentation';

// 明确参数输入只验证组件显示与确认入口，不表示实际任务、终端或执行结果。
function inputView(state: RelayTaskView['task']['state'], liveness: NonNullable<RelayTaskView['terminal']>['liveness']): RelayTaskView {
  return {
    task: { task_id: 'input-task', work_id: 'input-work', project_id: 'input-project', worktree_id: 'input-worktree', terminal_id: 'input-terminal', agent: 'input-agent', description: '输入描述', state, condition: '', permission_level: 'A', project_settings_revision: 1, degraded_launch: false, created_at: '2026-10-14T00:00:00Z', updated_at: '2026-10-14T00:01:00Z' },
    terminal: { terminal_id: 'input-terminal', worktree_id: 'input-worktree', agent: 'input-agent', drive_mode: 'structured', liveness, last_event_at: null, exited_at: liveness === 'exited' ? '2026-10-14T00:01:00Z' : null, last_sequence: 0, created_at: '2026-10-14T00:00:00Z' },
    worktree: { worktree_id: 'input-worktree', project_id: 'input-project', root: '输入工作区', branch: 'input-branch', prepared: false, active_task_id: state === 'released' ? '' : 'input-task', created_at: '2026-10-14T00:00:00Z' },
    transitions: [], event_publication: { status: 'none', pending_events: 0, published_events: 0 },
  };
}

it('明确非法时间输入立即拒绝，合法时间保留实际时长', () => {
  const view = inputView('working', 'live');
  expect(relayElapsed(view, Date.parse('2026-10-14T00:01:30Z'))).toBe('1 分 30 秒');
  expect(() => relayElapsed(view, Date.parse('2026-10-13T23:59:59Z'))).toThrow('时间顺序无法核查');
  view.task.created_at = 'invalid-input-time';
  expect(() => relayElapsed(view)).toThrow('接力任务时间无效');
});

for (const [liveness, label] of [['live', '近期活动确证 · live'], ['unverifiable', '活性无法核实 · unverifiable'], ['exited', '终端已退出 · exited']] as const) {
  it(`参数 ${liveness} 完整保留三值活性与独立任务状态`, () => {
    const view = inputView('working', liveness);
    const html = renderToStaticMarkup(<RelayTaskSummary view={view} />);
    expect(html).toContain(label);
    expect(html).toContain('工作中 · working');
    expect(html).toContain('input-terminal');
    expect(html).toContain('input-branch');
    expect(html).toContain('input-agent');
    expect(html).toContain('input-work');
  });
}

it('参数中的终端退出条件、降级与发布状态完整显示', () => {
  const view = inputView('working', 'exited');
  view.task.condition = 'terminal_exited'; view.task.degraded_launch = true;
  view.event_publication = { status: 'error', pending_events: 2, published_events: 0 };
  const html = renderToStaticMarkup(<RelayTaskSummary view={view} />);
  expect(html).toContain('terminal_exited · 终端退出，任务未完成');
  expect(html).toContain('degraded_launch');
  expect(html).toContain('事件接入失败');
  expect(html).toContain('待接入 2');
  expect(html).toContain('已接入 0');
});

const unreachable = async (): Promise<RelayTaskResult> => { throw new Error('静态组件测试不执行迁移命令'); };
const unchanged = () => { throw new Error('静态组件测试不执行状态变更'); };
for (const [state, liveness, label] of [['working', 'live', '确认任务完成'], ['done', 'live', '确认释放工作区'], ['done', 'exited', '确认释放工作区']] as const) {
  it(`参数 ${state}/${liveness} 需要人工确认并保留取消入口`, () => {
    const html = renderToStaticMarkup(<RelayTransitionControls view={inputView(state, liveness)} disabled={false} transition={unreachable} onChanged={unchanged} />);
    expect(html).toContain('取消确认');
    expect(html).toContain(label);
    expect(html).toContain('disabled=""');
    expect(html).toContain('type="checkbox"');
    if (state === 'done' && liveness !== 'exited') expect(html).toContain('尚无终端退出确证，不能释放工作区');
  });
}

it('参数 released 完整显示工作区已释放，迁移入口关闭', () => {
  const view = inputView('released', 'exited');
  expect(renderToStaticMarkup(<RelayTaskSummary view={view} />)).toContain('已释放 · released');
  expect(renderToStaticMarkup(<RelayTransitionControls view={view} disabled={false} transition={unreachable} onChanged={unchanged} />)).toBe('');
  expect(view.worktree!.active_task_id).toBe('');
});
