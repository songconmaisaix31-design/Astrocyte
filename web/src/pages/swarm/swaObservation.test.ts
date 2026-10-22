import { describe, expect, it } from 'vitest';
import type { SwaEnvelope } from '../../api/swa';
import type { components } from '../../api/schema';
import { readSwaFleet, readSwaAgents, recoverSwaRequest, canRequestSwaStart, swaStartRequestBlock, type PendingSwaRequest } from './swaObservation';

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

type FormAgent = ReturnType<typeof readSwaAgents>[number];
type Observation = components['schemas']['observation'];
const formSettings: components['schemas']['LocalProjectSettingsV1'] = { revision: 3, allow_directory: true, allowed_subdirs: [], expand_references: false, allowed_actions: ['start'], allowed_tools: [], external_model_cli: 'form-agent-input', allow_agent_control: true, history_roots: {} };
const observationInput = (status: Observation['status']): Observation => ({ status, reason: 'explicit-form-parameter', source: 'parser-parameter' });
function formAgentInput(): FormAgent {
  return {
    id: 'form-agent-input', model: 'form-model-input', settingsRevision: formSettings.revision,
    registered: observationInput('available'), enabled: observationInput('available'), installed: observationInput('available'), configured: observationInput('available'), start: observationInput('available'),
    launchable: observationInput('unknown'), read: observationInput('unknown'), tools: observationInput('unknown'),
  };
}
const startInput = (agent: FormAgent | undefined = formAgentInput()) => ({ agent, settings: formSettings, disabled: false, pending: false, unresolved: false, name: 'explicit-worker-name-input', inputFile: '/explicit/server-input-path' });

describe('SWA 启动表单的明确参数规则，不表示原生发现或启动通过', () => {
  it('静态条件允许提交时，未握手观察仍是 unknown', () => {
    const agent = formAgentInput();
    expect(canRequestSwaStart(startInput(agent))).toBe(true);
    expect(swaStartRequestBlock(agent, formSettings)).toBeNull();
    expect([agent.launchable.status, agent.read.status, agent.tools.status]).toEqual(['unknown', 'unknown', 'unknown']);
  });

  it('任何静态条件未核实、模型缺失或明确不可用的未握手能力都禁止提交', () => {
    for (const key of ['registered', 'enabled', 'installed', 'configured', 'start'] as const) {
      for (const status of ['unknown', 'unavailable'] as const) {
        const agent = { ...formAgentInput(), [key]: observationInput(status) };
        expect(canRequestSwaStart(startInput(agent))).toBe(false);
      }
    }
    for (const key of ['launchable', 'read', 'tools'] as const) expect(canRequestSwaStart(startInput({ ...formAgentInput(), [key]: observationInput('unavailable') }))).toBe(false);
    for (const model of ['', '   ']) expect(canRequestSwaStart(startInput({ ...formAgentInput(), model }))).toBe(false);
  });

  it('当前项目、权限版本、刷新禁用和原请求未结使用同一个提交门', () => {
    const input = startInput();
    for (const settings of [
      { ...formSettings, external_model_cli: 'other-project-agent-input' },
      { ...formSettings, allow_agent_control: false },
      { ...formSettings, allowed_actions: [] },
      { ...formSettings, revision: formSettings.revision + 1 },
    ]) expect(canRequestSwaStart({ ...input, settings })).toBe(false);
    // 父组件把项目/状态/能力的loading或stale统一传入disabled，按钮和onSubmit共用本门。
    for (const key of ['disabled', 'pending', 'unresolved'] as const) expect(canRequestSwaStart({ ...input, [key]: true })).toBe(false);
    expect(canRequestSwaStart({ ...input, agent: undefined })).toBe(false);
    expect(canRequestSwaStart({ ...input, name: ' ' })).toBe(false);
    expect(canRequestSwaStart({ ...input, inputFile: ' ' })).toBe(false);
  });

  it('解析明确 DTO 参数时保留观察，重复身份和不明格式拒绝', () => {
    const agent = formAgentInput();
    const parameter = {
      id: agent.id, execution_mode: 'structured', registered: agent.registered, enabled: agent.enabled, installed: agent.installed,
      configured: { ...agent.configured, model: agent.model }, launchable: agent.launchable,
      control: { start: agent.start, read: agent.read, tools: agent.tools }, requirements: { settings_revision: formSettings.revision },
    };
    const parse = (agents: unknown[]) => readSwaAgents({ ok: true, result: { schema: 'swa.capabilities.v1', agents } });
    expect(parse([parameter])[0]).toEqual(agent);
    expect(() => parse([parameter, parameter])).toThrow('重复');
    for (const status of ['ready', ['available'], null]) expect(() => parse([{ ...parameter, registered: { ...agent.registered, status } }])).toThrow('观察状态');
    expect(() => parse([{ ...parameter, requirements: { settings_revision: '3' } }])).toThrow('权限版本');
    expect(() => parse([{ ...parameter, control: { start: agent.start, read: agent.read } }])).toThrow('原生对象');
  });
});
