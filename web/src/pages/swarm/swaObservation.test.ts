import { describe, expect, it } from 'vitest';
import type { SwaEnvelope } from '../../api/swa';
import { readSwaFleet, recoverSwaRequest, type PendingSwaRequest } from './swaObservation';

// 以下为解析器的明确参数输入，不表示存在原生执行、模型返回或安装验收结果。
const projectID = 'parser-project-input', controller = 'parser-controller-input', taskID = 'parser-task-input';
const requestID = '0ff0ec12-d039-4c27-b41a-f61b46cf7be7';

function commandReceipt(action: string, native: unknown, state = 'completed', childState = 'completed'): SwaEnvelope {
  return { ok: true, result: {
    requestId: requestID, state, method: 'swa.controllerDeliver',
    receipt: { project_id: projectID, controller, root_task_id: taskID, action, child_request: `${requestID}/native`, invocation: { command: { action, request_id: requestID, controller, task_id: taskID, dispatch_id: 'parser-dispatch-input' } }, native },
    children: [{ requestId: `${requestID}/native`, state: childState }],
  } };
}

describe('SWA 原生 DTO 解析边界，实际模型链未验', () => {
  it('资源状态为合法 null 时保留 Worker 和未观察值', () => {
    const input: SwaEnvelope = { ok: true, result: { scope: { project_id: projectID, controller, root_task_id: taskID, run_id: 'parser-run-input' }, native: { workers: [{ dispatchId: 'parser-dispatch-input', taskId: 'parser-child-task-input', runId: 'parser-run-input', workerState: 'starting', dispatchStatus: 'running', terminalState: null }], page: { hasMore: true, total: 3 } } } };
    const fleet = readSwaFleet(input, projectID, controller, taskID);
    expect(fleet.workers).toHaveLength(1);
    expect(fleet.workers[0].terminalState).toBeNull();
    expect(fleet).toMatchObject({ hasMore: true, total: 3 });
    expect(() => readSwaFleet(input, 'other-project-input', controller, taskID)).toThrow('范围');
  });

  it('absent、pending 及子回执 pending 都保持未确认', () => {
    const pending: PendingSwaRequest = { requestID, action: 'worker.start', controller, taskID };
    for (const state of ['absent', 'pending']) expect(recoverSwaRequest({ ok: true, result: { requestId: requestID, state } }, projectID, pending)).toBeNull();
    expect(recoverSwaRequest(commandReceipt(pending.action, { state: 'ready', dispatchId: 'parser-dispatch-input' }, 'completed', 'pending'), projectID, pending)).toBeNull();
  });

  it('完成回执只恢复原 controller/task/dispatch，不构造新请求', () => {
    const pending: PendingSwaRequest = { requestID, action: 'worker.start', controller, taskID };
    const input = commandReceipt(pending.action, { state: 'ready', dispatchId: 'parser-dispatch-input' });
    expect(recoverSwaRequest(input, projectID, pending)).toMatchObject({ request_id: requestID, controller, task_id: taskID, dispatch_id: 'parser-dispatch-input' });
    expect(recoverSwaRequest(commandReceipt(pending.action, { state: 'unknown' }), projectID, pending)).toBeNull();
    expect(() => recoverSwaRequest(input, 'other-project-input', pending)).toThrow('项目');
    expect(() => recoverSwaRequest(input, projectID, { ...pending, requestID: 'other-request-input' })).toThrow('请求身份');
    expect(() => recoverSwaRequest(input, projectID, { ...pending, taskID: 'other-task-input' })).toThrow('选择');
  });

  it('输入恢复核查当前 journal 结算身份，unknown 不解锁', () => {
    const pending: PendingSwaRequest = { requestID, action: 'worker.send', controller, dispatchID: 'parser-dispatch-input' };
    const input = commandReceipt(pending.action, { state: 'accepted' });
    const result = input.result as Record<string, unknown>;
    result.current_input = { request_id: requestID, dispatch_id: pending.dispatchID, state: 'unknown' };
    expect(recoverSwaRequest(input, projectID, pending)).toBeNull();
    result.current_input = { request_id: requestID, dispatch_id: pending.dispatchID, state: 'accepted' };
    expect(recoverSwaRequest(input, projectID, pending)?.request_id).toBe(requestID);
    result.current_input = { request_id: requestID, dispatch_id: 'other-dispatch-input', state: 'accepted' };
    expect(() => recoverSwaRequest(input, projectID, pending)).toThrow('结算身份');
  });

  it('Controller 创建恢复须从同项目完成回执取得真实选择字段', () => {
    const pending: PendingSwaRequest = { requestID, action: 'controller.open' };
    const input: SwaEnvelope = { ok: true, result: { requestId: requestID, state: 'completed', method: 'swa.controllerOpen', receipt: { project_id: projectID, controller }, children: [] } };
    expect(recoverSwaRequest(input, projectID, pending)?.controller).toBe(controller);
    expect(() => recoverSwaRequest({ ...input, result: { requestId: requestID, state: 'completed', method: 'swa.taskCreate', receipt: { project_id: projectID, controller }, children: [] } }, projectID, pending)).toThrow('动作');
  });
});
