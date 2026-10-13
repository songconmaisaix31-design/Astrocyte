import type { RelayAgent, RelayProject, RelayReceipt, RelayTaskView } from '../../api/relay';
import type { components } from '../../api/schema';

export const relayStateLabel = { ready: '等待权限确认', accepted: '已接受', working: '工作中', done: '任务已完成', released: '已释放' };
export const relayLivenessLabel = { live: '近期活动确证 · live', unverifiable: '活性无法核实 · unverifiable', exited: '终端已退出 · exited' };
export const relayActionLabel = { intent: '声明意图', dispatch: '派发', send: '发送输入', read: '读取会话', stop: '停止', resume: '原生恢复', transition: '状态迁移', observe: '服务端观察', reconcile_events: '核对事件接入' };
export const relayReceiptLabel = { pending: '结果待确认', accepted: '动作已确认', rejected: '被拒绝 · 首 RED 保留', delivery_unknown: '交付未知 · 不重试' };

/** Only the HTTP service's known-task conflict hint can offer a read-only lookup. */
export function relayExistingTaskID(error: RelayReceipt['error']) {
  if (error?.code !== 'version_conflict') return '';
  const match = /^view_existing_task:([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})$/.exec(error.required_action);
  return match && match[1] !== '00000000-0000-0000-0000-000000000000' ? match[1] : '';
}

export function relaySingleInput(terminal: RelayTaskView['terminal'], agent: RelayAgent | undefined) {
  return terminal?.drive_mode === 'terminal' && agent?.input_mode === 'stdin_eof';
}

export function relayDisplayState(view: RelayTaskView) {
  if (view.task.condition === 'terminal_exited') return 'terminal_exited · 终端退出，任务未完成';
  return `${relayStateLabel[view.task.state]} · ${view.task.state}${view.task.condition ? ` · ${view.task.condition}` : ''}`;
}
export function relayElapsed(view: RelayTaskView, now = Date.now()) {
  const start = Date.parse(view.task.created_at);
  const end = ['done', 'released'].includes(view.task.state) ? Date.parse(view.task.updated_at) : now;
  if (!Number.isFinite(start) || !Number.isFinite(end)) return '未知';
  const seconds = Math.max(0, Math.floor((end - start) / 1000));
  return `${Math.floor(seconds / 60)} 分 ${seconds % 60} 秒`;
}
export function relayProjectMatches(root: string, project: RelayProject) {
  const normalize = (value: string) => {
    const normalized = value.replaceAll('\\', '/').replace(/\/$/, '');
    return /^[a-z]:/i.test(normalized) ? normalized.toLowerCase() : normalized;
  };
  const workRoot = normalize(root), projectRoot = normalize(project.root);
  return workRoot === projectRoot;
}
export function relayUnresolved(receipts: readonly RelayReceipt[], action?: RelayReceipt['action']) {
  return receipts.some(receipt => (!action || receipt.action === action) && ['pending', 'delivery_unknown'].includes(receipt.status));
}
export function relayAgentAvailability(agent: RelayAgent, inventory: readonly components['schemas']['LocalAgentV1'][]) {
  const item = inventory.find(item => item.id === agent.name);
  const labels = { available: '已核实可用', unavailable: '不可用', unknown: '未知', unsupported: '不支持' };
  return item ? `安装：${labels[item.installed.status]} · 配置：${labels[item.configured.status]} · 可启动：${labels[item.startable.status]}` : '安装 / 配置 / 可启动：尚无独立观察';
}
