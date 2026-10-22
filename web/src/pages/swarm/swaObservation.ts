import type { SwaEnvelope } from '../../api/swa';
import type { components } from '../../api/schema';

function object(value: unknown): Record<string, unknown> {
  if (!value || typeof value !== 'object' || Array.isArray(value)) throw new Error('执行服务未返回可核对的原生对象');
  return value as Record<string, unknown>;
}

function text(value: unknown, field: string): string {
  if (typeof value !== 'string' || !value) throw new Error(`执行服务缺少 ${field}，保持未知`);
  return value;
}

export interface SwaWorker {
  dispatchID: string;
  taskID: string;
  runID: string;
  workerState: string;
  dispatchStatus: string;
  terminalState: string | null;
}

/** 字段来自固定 Orca workerList 的原生投影，不推断额外状态或协作连线。 */
export function readSwaFleet(reply: SwaEnvelope, projectID: string, controller: string, taskID: string) {
  const result = object(reply.result), scope = object(result.scope), native = object(result.native);
  if (scope.project_id !== projectID || scope.controller !== controller || scope.root_task_id !== taskID) throw new Error('任务观察范围与当前选择不一致');
  const runID = text(scope.run_id, 'run_id');
  if (!Array.isArray(native.workers)) throw new Error('执行服务未返回 Worker 清单');
  const workers: SwaWorker[] = native.workers.map(value => {
    const row = object(value);
    if (row.runId !== runID) throw new Error('Worker 不属于当前 Run');
    return { dispatchID: text(row.dispatchId, 'dispatchId'), taskID: text(row.taskId, 'taskId'), runID, workerState: text(row.workerState, 'workerState'), dispatchStatus: text(row.dispatchStatus, 'dispatchStatus'), terminalState: row.terminalState === null ? null : text(row.terminalState, 'terminalState') };
  });
  if (new Set(workers.map(worker => worker.dispatchID)).size !== workers.length) throw new Error('Worker 清单包含重复 Dispatch，保持未确认');
  const page = object(native.page);
  if (typeof page.hasMore !== 'boolean' || typeof page.total !== 'number') throw new Error('执行服务未提供清单完整性');
  return { workers, runID, hasMore: page.hasMore, total: page.total };
}

export function nativeObservation(reply: SwaEnvelope): unknown {
  return object(reply.result).native;
}

export function observationJSON(value: unknown): string {
  return JSON.stringify(value, null, 2) ?? '服务未提供';
}

type SwaObservation = components['schemas']['observation'];
type ProjectSettings = components['schemas']['LocalProjectSettingsV1'];

function readObservation(value: unknown): SwaObservation {
  const observation = object(value);
  if (typeof observation.status !== 'string' || !['available', 'unavailable', 'unknown'].includes(observation.status)) throw new Error('服务返回了未知的能力观察状态');
  return { status: observation.status as SwaObservation['status'], reason: text(observation.reason, 'observation.reason'), source: text(observation.source, 'observation.source') };
}

export function readSwaAgents(reply: SwaEnvelope) {
  const result = object(reply.result);
  if (result.schema !== 'swa.capabilities.v1' || !Array.isArray(result.agents)) throw new Error('服务未返回已知 Agent 能力合同');
  const agents = result.agents.map(value => {
    const agent = object(value), configured = object(agent.configured), control = object(agent.control), requirements = object(agent.requirements);
    if (agent.execution_mode !== 'structured' || !Number.isSafeInteger(requirements.settings_revision) || Number(requirements.settings_revision) < 0) throw new Error('服务未提供当前结构化 Agent 的权限版本');
    return {
      id: text(agent.id, 'agent.id'), model: typeof configured.model === 'string' ? configured.model : '', settingsRevision: Number(requirements.settings_revision),
      registered: readObservation(agent.registered), enabled: readObservation(agent.enabled), installed: readObservation(agent.installed), configured: readObservation(configured),
      launchable: readObservation(agent.launchable), start: readObservation(control.start), read: readObservation(control.read), tools: readObservation(control.tools),
    };
  });
  if (new Set(agents.map(agent => agent.id)).size !== agents.length) throw new Error('能力清单包含重复 Agent，保持未确认');
  return agents;
}

type SwaAgent = ReturnType<typeof readSwaAgents>[number];

/** 只判定是否可向原启动门提交请求，不改变任何原生能力观察。 */
export function swaStartRequestBlock(agent: SwaAgent, settings: ProjectSettings): string | null {
  if (agent.id !== settings.external_model_cli) return '当前项目未选择此 Agent';
  if (agent.settingsRevision !== settings.revision) return '能力清单的权限版本已过期，请刷新项目';
  if (!settings.allow_agent_control || !settings.allowed_actions.includes('start')) return '当前项目未允许启动';
  for (const [label, observation] of [['注册', agent.registered], ['启用', agent.enabled], ['安装', agent.installed], ['配置', agent.configured], ['启动支持', agent.start]] as const) {
    if (observation.status !== 'available') return `${label}尚未核实可用：${observation.reason}`;
  }
  if (!agent.model.trim()) return '服务未提供已配置模型';
  for (const [label, observation] of [['启动', agent.launchable], ['读取', agent.read], ['工具限制', agent.tools]] as const) {
    if (observation.status === 'unavailable') return `${label}不可用：${observation.reason}`;
  }
  return null;
}

export function canRequestSwaStart(input: { agent: SwaAgent | undefined; settings: ProjectSettings; disabled: boolean; pending: boolean; unresolved: boolean; name: string; inputFile: string }): boolean {
  return !input.disabled && !input.pending && !input.unresolved && !!input.name.trim() && !!input.inputFile.trim() && !!input.agent && swaStartRequestBlock(input.agent, input.settings) === null;
}

export function readSwaTranscript(reply: SwaEnvelope) {
  const native = object(nativeObservation(reply));
  if (native.source !== 'transcript') throw new Error('服务未返回原生 transcript，不能解释为聊天消息');
  const transcript = object(native.transcript);
  if (!Array.isArray(transcript.messages) || typeof transcript.limited !== 'boolean' || typeof transcript.nextCursor !== 'string') throw new Error('原生消息页或完整性字段缺失');
  const messages = transcript.messages.map(value => {
    const message = object(value);
    if (!Array.isArray(message.blocks)) throw new Error('原生消息缺少内容块');
    const role = text(message.role, 'message.role');
    if (!['user', 'assistant', 'tool', 'reasoning', 'system'].includes(role)) throw new Error('原生消息角色尚未支持，保留诊断');
    return { id: text(message.id, 'message.id'), role, blocks: message.blocks.map(value => {
      const block = object(value);
      if (block.type === 'text' && typeof block.text !== 'string') throw new Error('原生消息文字不可解析');
      return { type: text(block.type, 'block.type'), text: block.type === 'text' ? block.text as string : observationJSON(block) };
    }) };
  });
  if (new Set(messages.map(message => message.id)).size !== messages.length) throw new Error('消息页包含重复身份，保持未确认');
  return { messages, limited: transcript.limited, nextCursor: transcript.nextCursor, sourceIdentity: text(native.sourceIdentity, 'sourceIdentity'), complete: native.contentComplete === true };
}

export interface PendingSwaRequest {
  requestID: string;
  action: string;
  controller?: string;
  taskID?: string;
  dispatchID?: string;
}

/** 只恢复同范围已完成的原生回执；absent 和 pending 不能证明没有执行。 */
export function recoverSwaRequest(reply: SwaEnvelope, projectID: string, pending: PendingSwaRequest): SwaEnvelope | null {
  const request = object(reply.result);
  if (request.requestId !== pending.requestID) throw new Error('原请求查询返回了不同的请求身份');
  if (request.state !== 'completed') return null;
  const receipt = object(request.receipt);
  const methods: Record<string, string> = { 'controller.open': 'swa.controllerOpen', 'task.create': 'swa.taskCreate', 'worker.start': 'swa.controllerDeliver', 'worker.send': 'swa.controllerDeliver', 'worker.stop': 'swa.controllerDeliver' };
  if (!methods[pending.action] || request.method !== methods[pending.action] || receipt.project_id !== projectID || pending.controller && receipt.controller !== pending.controller) throw new Error('原请求回执不属于当前项目、动作或 Controller');
  if (pending.action !== 'controller.open') {
    const command = object(object(receipt.invocation).command);
    if (command.action !== pending.action || command.request_id !== pending.requestID || pending.taskID && command.task_id !== pending.taskID || pending.dispatchID && command.dispatch_id !== pending.dispatchID) throw new Error('原请求回执的任务或 Worker 选择不一致');
  }
  const expectedChildren = receipt.children ?? (receipt.child_request ? [receipt.child_request] : []);
  if (!Array.isArray(expectedChildren) || expectedChildren.some(id => typeof id !== 'string') || !Array.isArray(request.children) || request.children.length !== expectedChildren.length) throw new Error('原请求的子回执范围不完整');
  const children = request.children.map(object);
  if (children.some(child => !expectedChildren.includes(child.requestId)) || new Set(children.map(child => child.requestId)).size !== children.length) throw new Error('原请求的子回执身份不一致');
  if (children.some(child => child.state !== 'completed')) return null;
  const controller = text(receipt.controller, 'receipt.controller');
  let taskID = pending.taskID, dispatchID = pending.dispatchID;
  if (pending.action === 'task.create') taskID = text(receipt.root_task_id, 'receipt.root_task_id');
  if (pending.action.startsWith('worker.')) {
    if (receipt.action !== pending.action) throw new Error('原请求回执的动作不一致');
    const native = object(receipt.native);
    const states: Record<string, string[]> = { 'worker.start': ['ready'], 'worker.stop': ['stopped'], 'worker.send': ['journaled', 'queued', 'accepted', 'handed-to-terminal'] };
    if (!states[pending.action]?.includes(String(native.state))) return null;
    if (pending.action === 'worker.start') dispatchID = text(native.dispatchId, 'receipt.native.dispatchId');
    if (pending.action === 'worker.send') {
      const current = object(request.current_input);
      if (current.request_id !== pending.requestID || current.dispatch_id !== pending.dispatchID) throw new Error('原输入的当前结算身份不一致');
      if (!states['worker.send'].includes(String(current.state))) return null;
    }
  }
  return { ...reply, request_id: pending.requestID, controller, ...(taskID ? { task_id: taskID } : {}), ...(dispatchID ? { dispatch_id: dispatchID } : {}), result: receipt };
}
