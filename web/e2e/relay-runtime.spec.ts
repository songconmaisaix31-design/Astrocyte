import { test, expect, type Page } from '@playwright/test';
import type { components } from '../src/api/schema';
import { fixtureWorksSnapshot } from '../src/pages/workspace/worksAlbumPresentation';

type S = components['schemas'];
const timestamp = '2026-10-14T00:00:00Z';
const capability = { status: 'supported' as const, reason: 'contract-local mock only', checked_at: timestamp };

async function mockRelay(page: Page) {
  const snapshot = fixtureWorksSnapshot();
  const project: S['RelayProjectV1'] = { project_id: 'public-project', name: 'Public fixture', root: snapshot.works[0].root, permission_level: 'A', project_settings_revision: 4, allowed_subdirs: ['README.md'], allowed_actions: ['start', 'send', 'observe', 'stop', 'reconcile'], allow_agent_control: false, allow_directory: false, expand_references: false };
  const spec: S['RelayAgentV1'] = { name: 'new-public-cli', launch_cmd: ['public-cli'], capabilities: { start: capability, send: capability, observe: capability, stop: capability }, drive_mode: 'terminal' };
  const observation = { status: 'available' as const, reason: 'contract-local mock only', checked_at: timestamp };
  const inventory: S['LocalAgentV1'] = { id: spec.name, display_name: spec.name, version: null, installed: observation, configured: observation, startable: observation, capabilities: { discover: capability, read_context: capability, start: capability, resume: capability, send: capability, stop: capability, observe: capability, reconcile: capability } };
  const state = {
    project, spec, inventory,
    views: [] as S['RelayTaskViewV1'][],
    receipts: [] as S['RelayReceiptV1'][],
    events: [{ event_id: 'derived-newer', kind: 'edit_file', topic: 'public-work', tokens: null, scope: 'public-project', ts: timestamp }, { event_id: 'derived-older', kind: 'agent_task', topic: 'public-work', tokens: 12, scope: 'public-project', ts: '2026-10-13T23:59:00Z' }] as S['WorkbenchEventV1'][],
    writes: [] as { action: string; body: Record<string, unknown>; key: string }[],
    confirmMode: 'success' as 'success' | 'denied' | 'duplicate' | 'unknown',
    sendMode: 'success' as 'success' | 'unknown',
  };
  const receipt = (action: S['RelayReceiptV1']['action'], key: string, status: S['RelayReceiptV1']['status'] = 'accepted', code?: S['RelayReceiptErrorV1']['code']): S['RelayReceiptV1'] => {
    const value: S['RelayReceiptV1'] = { receipt_id: `receipt-${action}-${state.receipts.length}`, task_id: state.views.at(-1)?.task.task_id || 'task-public', operation_id: key, action, operator: 'human', status, created_at: timestamp, finished_at: timestamp, error: code ? { code, message: action === 'dispatch' && code === 'version_conflict' ? 'existing active task task-current' : `contract-local ${code}`, request_id: key, retryable: false, required_action: 'read_task_receipts' } : null };
    state.receipts.push(value);
    return value;
  };
  await page.route('**/api/v1/works', route => route.fulfill({ json: { schema_version: 1, snapshot } }));
  await page.route('**/api/v1/relay/**', async route => {
    const request = route.request(), path = new URL(request.url()).pathname.replace('/api/v1/relay/', '');
    if (request.method() === 'GET') {
      const body = path === 'projects' ? { schema_version: 1, projects: [state.project] }
        : path === 'agents' ? { schema_version: 1, agents: [state.spec], inventory: [state.inventory] }
        : path === 'tasks' ? { schema_version: 1, tasks: state.views, dispatched_work_ids: state.receipts.some(item => item.action === 'dispatch' && item.status === 'accepted') ? ['work-a'] : [] }
        : path.endsWith('/events') ? { schema_version: 1, events: state.events, raw: 'PRIVATE_RAW_CANARY_NEVER_RENDER' }
        : path.endsWith('/receipts') ? { schema_version: 1, receipts: state.receipts.filter(receipt => receipt.task_id === path.split('/')[1]) }
        : { schema_version: 1, view: state.views.find(view => view.task.task_id === path.split('/')[1]) };
      await route.fulfill({ json: body }); return;
    }
    const body = request.postDataJSON() as Record<string, unknown>, key = request.headers()['idempotency-key'];
    expect(key).toMatch(/^[a-f\d-]{36}$/);
    expect(request.headers()['x-csrf-token']).toBeTruthy();
    const action = path === 'intents' ? 'intent' : path.split('/').at(-1)!;
    state.writes.push({ action, body, key });
    if (path === 'intents') {
      expect(body).toEqual({ work_id: 'work-a', project_id: 'public-project', agent: state.spec.name, task: 'Continue the public example' });
      const taskID = state.views.length ? `task-public-${state.views.length + 1}` : 'task-public';
      state.views.push({ task: { task_id: taskID, work_id: 'work-a', project_id: 'public-project', worktree_id: '', terminal_id: '', agent: state.spec.name, description: String(body.task), state: 'ready', condition: '', permission_level: state.project.permission_level, project_settings_revision: 4, degraded_launch: false, created_at: timestamp, updated_at: timestamp }, terminal: null, worktree: null, transitions: [{ task_id: taskID, from: '', to: 'ready', ts: timestamp, operator: 'human' }], event_publication: { status: 'none', pending_events: 0, published_events: 0 } });
      receipt('intent', key);
      await route.fulfill({ json: { schema_version: 1, view: state.views.at(-1) } }); return;
    }
    const view = state.views.find(view => view.task.task_id === path.split('/')[1])!;
    if (action === 'confirm') {
      expect(body.permission_confirmed).toBe(true);
      expect(body.permission_level).toBe(state.project.permission_level);
      expect(body.project_settings_revision).toBe(4);
      if (state.confirmMode !== 'success') {
        const code = state.confirmMode === 'denied' ? 'scope_denied' : state.confirmMode === 'duplicate' ? 'version_conflict' : 'delivery_unknown';
        const failed = receipt('dispatch', key, code === 'delivery_unknown' ? 'delivery_unknown' : 'rejected', code);
        await route.fulfill({ status: code === 'scope_denied' ? 403 : code === 'version_conflict' ? 409 : 503, json: { schema_version: 1, error: failed.error } }); return;
      }
      view.task.state = 'accepted'; view.task.terminal_id = 'terminal-stable'; view.task.worktree_id = 'worktree-neutral'; view.task.degraded_launch = true;
      view.terminal = { terminal_id: 'terminal-stable', worktree_id: 'worktree-neutral', agent: view.task.agent, drive_mode: 'terminal', liveness: 'live', last_event_at: timestamp, exited_at: null, last_sequence: 1, created_at: timestamp };
      view.worktree = { worktree_id: 'worktree-neutral', project_id: 'public-project', root: 'C:/public-neutral-fixture', branch: 'relay/public', prepared: true, active_task_id: view.task.task_id, created_at: timestamp };
      view.event_publication = { status: 'published', pending_events: 0, published_events: state.events.length };
      view.transitions.push({ task_id: view.task.task_id, from: 'ready', to: 'accepted', ts: timestamp, operator: 'human' });
      await route.fulfill({ json: { schema_version: 1, view, receipt: receipt('dispatch', key) } }); return;
    }
    if (action === 'transition') {
      expect(body.from).toBe(view.task.state);
      if (body.to === 'done') expect(body.completion_confirmed).toBe(true);
      if (body.to === 'released') {
        expect(view.terminal?.liveness).toBe('exited'); expect(view.terminal?.exited_at).toBeTruthy();
        view.worktree!.active_task_id = '';
      }
      view.transitions.push({ task_id: view.task.task_id, from: body.from as S['RelayTransitionV1']['from'], to: body.to as S['RelayTransitionV1']['to'], ts: timestamp, operator: 'human' });
      view.task.state = body.to as S['RelayTaskV1']['state'];
      receipt('transition', key);
      await route.fulfill({ json: { schema_version: 1, view } }); return;
    }
    if (action === 'reconcile') {
      view.event_publication = { status: 'published', pending_events: 0, published_events: state.events.length };
      receipt('reconcile_events', key);
      await route.fulfill({ json: { schema_version: 1, view } }); return;
    }
    if (action === 'send' && state.sendMode === 'unknown') {
      await route.fulfill({ json: { schema_version: 1, view, receipt: receipt('send', key, 'delivery_unknown', 'delivery_unknown') } }); return;
    }
    if (action === 'stop' && view.terminal) { view.terminal.liveness = 'exited'; view.terminal.exited_at = timestamp; if (view.task.state !== 'done') view.task.condition = 'terminal_exited'; }
    await route.fulfill({ json: { schema_version: 1, view, receipt: receipt(action as S['RelayReceiptV1']['action'], key) } });
  });
  return state;
}

async function declareIntent(page: Page) {
  await page.goto('/workspace/works');
  await page.getByRole('region', { name: '作品寻回册', exact: true }).getByLabel('个人知识库 的操作', { exact: true }).getByRole('button', { name: '接力', exact: true }).click();
  const dialog = page.getByRole('dialog');
  await dialog.getByLabel('获准项目', { exact: true }).selectOption('public-project');
  await dialog.getByLabel('偏好 Agent', { exact: true }).selectOption('new-public-cli');
  await dialog.getByLabel('任务描述', { exact: true }).fill('Continue the public example');
  await dialog.getByRole('button', { name: '保存意图并查看权限', exact: true }).click();
  await expect(dialog.getByRole('region', { name: '权限确认页' })).toBeVisible();
  return dialog;
}

function seedProjectedTask(state: Awaited<ReturnType<typeof mockRelay>>, status: 'working' | 'done' | 'released' = 'working') {
  const exited = status === 'released';
  state.views.push({
    task: { task_id: 'task-public', work_id: 'work-a', project_id: 'public-project', worktree_id: 'worktree-neutral', terminal_id: 'terminal-stable', agent: state.spec.name, description: 'Continue the public example', state: status, condition: '', permission_level: state.project.permission_level, project_settings_revision: 4, degraded_launch: false, created_at: timestamp, updated_at: timestamp },
    terminal: { terminal_id: 'terminal-stable', worktree_id: 'worktree-neutral', agent: state.spec.name, drive_mode: 'terminal', liveness: exited ? 'exited' : 'live', last_event_at: timestamp, exited_at: exited ? timestamp : null, last_sequence: 1, created_at: timestamp },
    worktree: { worktree_id: 'worktree-neutral', project_id: 'public-project', root: 'C:/public-neutral-fixture', branch: 'relay/public', prepared: true, active_task_id: exited ? '' : 'task-public', created_at: timestamp },
    transitions: [{ task_id: 'task-public', from: 'accepted', to: 'working', ts: timestamp, operator: 'service' }],
    event_publication: { status: 'published', pending_events: 0, published_events: 2 },
  });
}

test.describe('relay contract-local mock evidence; not CLI or durable acceptance', () => {
  test('registered CLI, mandatory default A confirmation and permission refusal', async ({ page }) => {
    const state = await mockRelay(page); state.confirmMode = 'denied';
    const dialog = await declareIntent(page);
    expect(state.writes.map(item => item.action)).toEqual(['intent']);
    await expect(dialog.getByLabel('权限级别', { exact: true })).toHaveValue('A');
    await expect(dialog).toContainText(state.project.root);
    await expect(dialog).toContainText('README.md');
    await expect(dialog).toContainText('start、send、observe、stop、reconcile');
    await expect(dialog.getByRole('button', { name: '确认资料权限并请求接力' })).toBeDisabled();
    await dialog.getByRole('checkbox').check();
    await dialog.getByRole('button', { name: '确认资料权限并请求接力' }).click();
    await expect(dialog.getByRole('alert')).toContainText('scope_denied');
    expect(state.writes.map(item => item.action)).toEqual(['intent', 'confirm']);
    expect(state.views[0].task.state).toBe('ready');
    await page.keyboard.press('Escape');
    await expect(page.getByText('new-public-cli', { exact: true }).first()).toBeVisible();
    expect(state.receipts[1].status).toBe('rejected');
  });

  test('duplicate dispatch reports existing task and does not create another writer', async ({ page }) => {
    const state = await mockRelay(page); state.confirmMode = 'duplicate';
    const dialog = await declareIntent(page);
    await dialog.getByRole('checkbox').check();
    await dialog.getByRole('button', { name: '确认资料权限并请求接力' }).click();
    await expect(dialog.getByRole('alert')).toContainText('existing active task task-current');
    expect(state.writes.filter(item => item.action === 'confirm')).toHaveLength(1);
    expect(state.views[0].terminal).toBeNull();
  });

  test('existing independent B and C grants confirm their actual service summary', async ({ page }) => {
    const state = await mockRelay(page);
    state.project.permission_level = 'C'; state.project.allow_directory = true; state.project.expand_references = true;
    const dialog = await declareIntent(page);
    await expect(dialog.getByLabel('权限级别', { exact: true })).toHaveValue('C');
    await expect(dialog.getByLabel('权限级别', { exact: true })).toHaveAttribute('readonly', '');
    await dialog.getByRole('checkbox').check();
    await dialog.getByRole('button', { name: '确认资料权限并请求接力' }).click();
    await expect(dialog.getByRole('status').filter({ hasText: '派发收据' })).toContainText('已确认');
    expect(state.writes[1].body.permission_level).toBe('C');
  });

  test('changed project revision blocks confirmation while preserving the intent', async ({ page }) => {
    const state = await mockRelay(page);
    const dialog = await declareIntent(page);
    await dialog.getByRole('checkbox').check();
    state.project.project_settings_revision = 5;
    await dialog.getByRole('button', { name: '读取权限、任务和收据' }).click();
    await expect(dialog.getByRole('alert')).toContainText('任务的权限版本已过期');
    await expect(dialog.getByRole('button', { name: '确认资料权限并请求接力' })).toBeDisabled();
    expect(state.writes.map(item => item.action)).toEqual(['intent']);
  });

  test('unknown dispatch stays blocked after reads, closing and reopening', async ({ page }) => {
    const state = await mockRelay(page); state.confirmMode = 'unknown';
    const dialog = await declareIntent(page);
    await dialog.getByRole('checkbox').check();
    await dialog.getByRole('button', { name: '确认资料权限并请求接力' }).click();
    await expect(dialog.getByText('派发结果未知，禁止重派。只读取任务和收据核对结果。')).toBeVisible();
    await dialog.getByRole('button', { name: '读取权限、任务和收据' }).click();
    await expect(dialog.getByRole('button', { name: '确认资料权限并请求接力' })).toBeDisabled();
    await page.keyboard.press('Escape');
    await page.getByRole('button', { name: '查看意图与权限', exact: true }).click();
    await expect(page.getByRole('dialog').getByRole('button', { name: '确认资料权限并请求接力' })).toBeDisabled();
    expect(state.writes.filter(item => item.action === 'confirm')).toHaveLength(1);
  });

  test('service projections retain state across reload and show all three liveness values', async ({ page }) => {
    const state = await mockRelay(page);
    const dialog = await declareIntent(page);
    await dialog.getByRole('checkbox').check();
    await dialog.getByRole('button', { name: '确认资料权限并请求接力' }).click();
    await expect(dialog.getByRole('status').filter({ hasText: '派发收据' })).toContainText('已确认');
    await page.keyboard.press('Escape');
    await page.getByRole('button', { name: '查看任务、事件和收据' }).click();
    const taskDialog = page.getByRole('dialog');
    await expect(taskDialog).toContainText('近期活动确证 · live');
    await expect(taskDialog).toContainText('已接受 · accepted');
    await expect(taskDialog).toContainText('degraded_launch');
    await expect(taskDialog.getByRole('region', { name: '任务事件流' }).locator('tbody tr').first()).toContainText('derived-newer');
    await expect(page.getByText('PRIVATE_RAW_CANARY_NEVER_RENDER')).toHaveCount(0);
    state.views[0].terminal!.liveness = 'unverifiable';
    await taskDialog.getByRole('button', { name: '只读取任务、事件和收据' }).click();
    await expect(taskDialog).toContainText('活性无法核实 · unverifiable');
    await expect(taskDialog).toContainText('已接受 · accepted');
    state.views[0].task.state = 'working';
    state.views[0].transitions.push({ task_id: 'task-public', from: 'accepted', to: 'working', ts: timestamp, operator: 'service' });
    await taskDialog.getByRole('button', { name: '只读取任务、事件和收据' }).click();
    await expect(taskDialog).toContainText('工作中 · working');
    await taskDialog.getByLabel('发送输入', { exact: true }).fill('public follow-up');
    await taskDialog.getByRole('button', { name: '发送输入', exact: true }).click();
    await expect(taskDialog.getByRole('status').filter({ hasText: '发送输入收据' })).toContainText('已确认');
    await taskDialog.getByRole('button', { name: '读取会话并留收据' }).click();
    await expect(taskDialog.getByRole('status').filter({ hasText: '读取会话收据' })).toContainText('已确认');
    await taskDialog.getByRole('button', { name: '停止 Agent' }).click();
    await expect(taskDialog).toContainText('终端已退出 · exited');
    await expect(taskDialog).toContainText('terminal_exited');
    await page.keyboard.press('Escape');
    await page.reload();
    await expect(page.getByText('terminal_exited · 终端退出，任务未完成', { exact: true })).toBeVisible();
    expect(state.writes.map(item => item.action)).toEqual(['intent', 'confirm', 'send', 'read', 'stop']);
    expect(state.views[0].task.state).toBe('working');
    expect(state.views[0].terminal!.terminal_id).toBe('terminal-stable');
    expect(state.receipts.map(item => item.action)).toEqual(['intent', 'dispatch', 'send', 'read', 'stop']);
  });

  test('unknown send preserves input, prevents replay and keeps stop available', async ({ page }) => {
    const state = await mockRelay(page); state.sendMode = 'unknown';
    const dialog = await declareIntent(page);
    await dialog.getByRole('checkbox').check();
    await dialog.getByRole('button', { name: '确认资料权限并请求接力' }).click();
    await expect(dialog.getByRole('status').filter({ hasText: '派发收据' })).toContainText('已确认');
    await page.keyboard.press('Escape');
    await page.getByRole('button', { name: '查看任务、事件和收据' }).click();
    const taskDialog = page.getByRole('dialog');
    await taskDialog.getByLabel('发送输入', { exact: true }).fill('retain unknown input');
    await taskDialog.getByRole('button', { name: '发送输入', exact: true }).click();
    await expect(taskDialog.getByRole('button', { name: '发送输入', exact: true })).toBeDisabled();
    await expect(taskDialog.getByLabel('发送输入', { exact: true })).toHaveValue('retain unknown input');
    await taskDialog.getByRole('button', { name: '只读取任务、事件和收据' }).click();
    await expect(taskDialog.getByRole('button', { name: '发送输入', exact: true })).toBeDisabled();
    await expect(taskDialog.getByRole('button', { name: '停止 Agent' })).toBeEnabled();
    await taskDialog.getByRole('button', { name: '停止 Agent' }).click();
    await expect(taskDialog).toContainText('terminal_exited');
    expect(state.writes.filter(item => item.action === 'send')).toHaveLength(1);
  });

  test('human completion cancellation, done stop/read and exited release are reachable', async ({ page }) => {
    const state = await mockRelay(page); seedProjectedTask(state);
    await page.goto('/workspace/works');
    await page.getByRole('button', { name: '查看任务、事件和收据', exact: true }).click();
    const dialog = page.getByRole('dialog');
    const transition = dialog.getByRole('region', { name: '人工完成与交接' });
    await expect(transition.getByRole('button', { name: '确认任务完成', exact: true })).toBeDisabled();
    await transition.getByRole('checkbox').check();
    await transition.getByRole('button', { name: '取消确认', exact: true }).click();
    await expect(transition.getByRole('checkbox')).not.toBeChecked();
    expect(state.writes).toEqual([]);
    await transition.getByRole('checkbox').check();
    await transition.getByRole('button', { name: '确认任务完成', exact: true }).click();
    await expect(dialog).toContainText('任务已完成 · done');
    await expect(dialog.getByRole('button', { name: '发送输入', exact: true })).toBeDisabled();
    await expect(dialog.getByRole('button', { name: '读取会话并留收据' })).toBeEnabled();
    await expect(dialog.getByRole('button', { name: '停止 Agent' })).toBeEnabled();
    await expect(transition.getByRole('button', { name: '确认释放工作区', exact: true })).toBeDisabled();
    expect(state.writes.map(item => item.action)).toEqual(['transition']);
    await dialog.getByRole('button', { name: '停止 Agent' }).click();
    await expect(dialog).toContainText('终端已退出 · exited');
    await transition.getByRole('checkbox').check();
    await transition.getByRole('button', { name: '确认释放工作区', exact: true }).click();
    await expect(dialog).toContainText('已释放 · released');
    expect(state.writes.map(item => item.action)).toEqual(['transition', 'stop', 'transition']);
    expect(state.views[0].worktree!.active_task_id).toBe('');
    expect(state.receipts.filter(receipt => receipt.action === 'transition')).toHaveLength(2);
  });

  test('released work can pass through a fresh permission page to another registered Agent', async ({ page }) => {
    const state = await mockRelay(page); seedProjectedTask(state, 'released');
    const workspace = { ...state.views[0].worktree! };
    state.spec.name = 'alternate-public-cli'; state.inventory.id = state.spec.name;
    await page.goto('/workspace/works');
    const album = page.getByRole('region', { name: '作品寻回册', exact: true });
    await album.getByLabel('个人知识库 的操作', { exact: true }).getByRole('button', { name: '接力', exact: true }).click();
    const dialog = page.getByRole('dialog');
    await dialog.getByLabel('获准项目', { exact: true }).selectOption('public-project');
    await dialog.getByLabel('偏好 Agent', { exact: true }).selectOption('alternate-public-cli');
    await dialog.getByLabel('任务描述', { exact: true }).fill('Continue the public example');
    await dialog.getByRole('button', { name: '保存意图并查看权限' }).click();
    await expect(dialog.getByRole('region', { name: '权限确认页' })).toBeVisible();
    expect(state.writes.map(item => item.action)).toEqual(['intent']);
    await dialog.getByRole('checkbox').check();
    await dialog.getByRole('button', { name: '确认资料权限并请求接力' }).click();
    await expect(dialog.getByRole('status').filter({ hasText: '派发收据' })).toContainText('已确认');
    expect(state.views[1].task.agent).toBe('alternate-public-cli');
    expect(state.views[1].worktree).toMatchObject({ worktree_id: workspace.worktree_id, root: workspace.root, branch: workspace.branch, created_at: workspace.created_at });
    expect(state.writes.map(item => item.action)).toEqual(['intent', 'confirm']);
  });

  test('event reconciliation is manual and publication is not presented as cognition refresh', async ({ page }) => {
    const state = await mockRelay(page); seedProjectedTask(state);
    state.views[0].event_publication = { status: 'error', pending_events: 2, published_events: 0 };
    await page.goto('/workspace/works');
    await page.getByRole('button', { name: '查看任务、事件和收据', exact: true }).click();
    const dialog = page.getByRole('dialog');
    await expect(dialog).toContainText('事件接入失败');
    await expect(dialog).toContainText('认知数值仍来自已发布快照');
    expect(state.writes).toEqual([]);
    await dialog.getByRole('button', { name: '核对事件接入', exact: true }).click();
    await expect(dialog).toContainText('已接入本机应用事件');
    await expect(dialog.getByRole('region', { name: '任务收据' })).toContainText('核对事件接入 · 动作已确认');
    expect(state.writes.map(item => item.action)).toEqual(['reconcile']);
    expect(state.views[0].task.state).toBe('working');
  });

  test('nested project root is not offered as an exact authorized work root', async ({ page }) => {
    const state = await mockRelay(page); state.project.root = 'C:/works';
    await page.goto('/workspace/works');
    await page.getByRole('region', { name: '作品寻回册', exact: true }).getByLabel('个人知识库 的操作', { exact: true }).getByRole('button', { name: '接力', exact: true }).click();
    const dialog = page.getByRole('dialog');
    await expect(dialog).toContainText('作品尚无根目录完全匹配的获准项目');
    await expect(dialog.getByRole('button', { name: '保存意图并查看权限' })).toBeDisabled();
    expect(state.writes).toEqual([]);
  });
});
