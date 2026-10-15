import { chromium, expect } from '@playwright/test';
import { readFile, writeFile } from 'node:fs/promises';
import { join, resolve } from 'node:path';
import { DatabaseSync } from 'node:sqlite';
import { humanAPI } from '../../tests/s1/api.mjs';
import type { components } from '../src/api/schema';

type S = components['schemas'];
const origin = process.env.ASTROCYTE_TEST_PRODUCT_URL;
const output = process.env.ASTROCYTE_VERIFICATION_OUTPUT_DIR;
if (origin !== 'http://127.0.0.1:64505' || !output) throw new Error('需要同一产品入口和本轮公开项目、论文记录目录');
const directory = resolve(output);
const projectEvidence = JSON.parse(await readFile(resolve(directory, 'product-project-human.json'), 'utf8')) as { project_id: string; space_id: string };
const paperEvidence = JSON.parse(await readFile(resolve(directory, 'product-paper-human.json'), 'utf8')) as { imports: { material_id: string; revision: number }[] };
const evidence: {
  stage: string; started_at: string; completed_at?: string; project_id: string; material_id: string;
  goal_id?: string; proposal_id?: string; grant_id?: string; context_id?: string;
  context_status?: number; context_error?: { code: string; required_action?: string };
  operations: { method: string; path: string; operation_id: string; status?: number }[];
  events: { event_id: string; event_type: string; aggregate_id: string }[]; native_sessions?: number;
} = { stage: 'started', started_at: new Date().toISOString(), project_id: projectEvidence.project_id, material_id: paperEvidence.imports[2].material_id, operations: [], events: [] };
const path = resolve(directory, 'product-proposal-human.json');
await writeFile(path, JSON.stringify(evidence, null, 2), { flag: 'wx' });
const save = () => writeFile(path, JSON.stringify(evidence, null, 2));
const browser = await chromium.launch();
const api = await humanAPI(origin);
try {
  const project = (await api.get('/local-projects') as S['LocalProjectListV1']).items.find(item => item.id === evidence.project_id);
  if (!project) throw new Error('同一产品没有原已登记的公开前端项目');
  const { space } = await api.get(`/project-spaces/${project.space_id}`) as S['ProjectSpaceResultV1'];
  expect(space.material_refs).toEqual([]);
  const referenceKey = crypto.randomUUID();
  evidence.stage = 'including_existing_public_paper'; await save();
  await api.write(`/project-spaces/${project.space_id}/references`, { expected_version: space.version, material_id: evidence.material_id, revision: 1 }, { key: referenceKey });
  const page = await browser.newPage({ viewport: { width: 1920, height: 1080 } });
  page.setDefaultTimeout(30_000);
  page.on('request', request => {
    const url = new URL(request.url());
    if (url.pathname.startsWith('/api/v1/') && request.method() !== 'GET') evidence.operations.push({ method: request.method(), path: url.pathname, operation_id: request.headers()['idempotency-key'] ?? '' });
  });
  page.on('response', response => {
    const operation = evidence.operations.find(item => item.operation_id === response.request().headers()['idempotency-key']);
    if (operation) operation.status = response.status();
  });
  await page.goto(`${origin}/workspace?tab=proposals`);
  await page.getByRole('button', { name: '创建目标', exact: true }).click();
  const goalDialog = page.getByRole('dialog', { name: '创建目标', exact: true });
  await goalDialog.getByLabel('目标所属项目', { exact: true }).selectOption(project.id);
  await goalDialog.getByLabel('目标名称', { exact: true }).fill('同实例公开论文固定上下文核查');
  await goalDialog.getByLabel('目标成功条件', { exact: true }).fill('固定来源版本与项目权限可回查，原生会话保持未启动');
  await goalDialog.getByLabel('目标优先级', { exact: true }).fill('1');
  evidence.stage = 'creating_human_goal'; await save();
  const goalResponse = page.waitForResponse(response => response.url().endsWith('/api/v1/goals') && response.request().method() === 'POST');
  await goalDialog.getByRole('button', { name: '保存目标', exact: true }).click();
  const savedGoal = await goalResponse;
  expect(savedGoal.status()).toBe(200);
  const { goal } = await savedGoal.json() as S['S2GoalResultV1'];
  evidence.goal_id = goal.id; await save();
  await page.getByRole('button', { name: '创建提案', exact: true }).click();
  const form = page.getByRole('dialog', { name: '创建提案', exact: true });
  const title = '同实例公开论文固定来源核查';
  const fields: [string, string][] = [
    ['提案标题', title], ['目标 ID', goal.id], ['目标版本', String(goal.revision)], ['项目 ID', project.id],
    ['允许根目录', join(project.root, 'web/src/pages/workspace/proposals')], ['允许修改的路径', ''], ['允许动作', 'read_context'],
    ['最大并发数', '1'], ['调用上限', '3'], ['预期产出', '来自真实公开论文和当前项目快照的固定上下文包'],
    ['成功条件', 'HTTP、页面与SQLite均保留实际固定版本'], ['停止条件', '用户撤销或权限收窄后停止建立新上下文'],
    ['最大工作项数', '1'], ['最大尝试次数', '1'], ['连续无进展轮数上限', '1'], ['无进展时限（秒）', '300'],
  ];
  for (const [label, value] of fields) await form.getByLabel(label, { exact: true }).fill(value);
  const localMinute = (date: Date) => new Date(date.getTime() - date.getTimezoneOffset() * 60_000).toISOString().slice(0, 16);
  await form.getByLabel('截止时间', { exact: true }).fill(localMinute(new Date(Date.now() + 86_400_000)));
  const picker = form.getByRole('group', { name: '项目已纳入的固定资料版本', exact: true });
  await expect(picker.getByRole('checkbox')).toHaveCount(1);
  await picker.getByRole('checkbox').check();
  evidence.stage = 'creating_human_proposal'; await save();
  const proposalResponse = page.waitForResponse(response => response.url().endsWith('/api/v1/proposals') && response.request().method() === 'POST');
  await form.getByRole('button', { name: '保存提案草稿', exact: true }).click();
  const created = await proposalResponse;
  expect(created.status()).toBe(200);
  const { proposal } = await created.json() as S['S2ProposalResultV1'];
  expect(proposal.content.inputs).toEqual([{ material_id: evidence.material_id, revision: 1 }]);
  evidence.proposal_id = proposal.id; await save();
  await page.getByRole('button', { name: `查看提案 ${title}`, exact: true }).click();
  const dialog = page.getByRole('dialog', { name: '提案详情', exact: true });
  evidence.stage = 'submitting_human_review'; await save();
  const submitting = page.waitForResponse(response => response.url().endsWith(`/proposals/${proposal.id}/submit`));
  await dialog.getByRole('button', { name: '提交审阅', exact: true }).click();
  expect((await submitting).status()).toBe(200);
  await dialog.getByLabel('批准有效期', { exact: true }).fill(localMinute(new Date(Date.now() + 3_600_000)));
  await dialog.getByRole('checkbox', { name: '我已核对当前固定版本、范围、资源上限和停止条件', exact: true }).check();
  evidence.stage = 'approving_read_context_only'; await save();
  const approving = page.waitForResponse(response => response.url().endsWith(`/proposals/${proposal.id}/approvals`));
  await dialog.getByRole('button', { name: '批准当前版本', exact: true }).click();
  const approved = await approving;
  expect(approved.status()).toBe(200);
  const { approval } = await approved.json() as S['S2ApprovalResultV1'];
  evidence.grant_id = approval.id; await save();
  expect(approval.scope.actions).toEqual(['read_context']);
  expect(approval.scope.tools).toEqual([]);
  const panel = dialog.getByRole('region', { name: '批准范围内的固定上下文', exact: true });
  await panel.getByLabel('上下文预算（字节）', { exact: true }).fill('131072');
  await panel.getByLabel('上下文未解决问题', { exact: true }).fill('当前认知快照尚不可用；受限原生工具保持拒绝，尚未启动原生会话');
  evidence.stage = 'building_fixed_context_once'; await save();
  const building = page.waitForResponse(response => response.url().endsWith('/api/v1/context-packets') && response.request().method() === 'POST', { timeout: 30_000 });
  await panel.getByRole('button', { name: '建立固定上下文包', exact: true }).click();
  const built = await building;
  evidence.context_status = built.status();
  const body = await built.json() as S['S2ContextPacketResultV1'] | S['ErrorV1'];
  if ('error' in body) evidence.context_error = { code: body.error.code, required_action: body.error.required_action };
  await save();
  expect(built.status()).toBe(200);
  if (!('context_packet' in body)) throw new Error('上下文成功响应缺少准确固定包');
  const packet = body.context_packet;
  evidence.context_id = packet.id; await save();
  expect(packet.sources.map(item => item.reference)).toEqual([{ material_id: evidence.material_id, revision: 1 }]);
  expect(packet.scope.roots).toEqual(['web/src/pages/workspace/proposals']);
  expect(packet.grant_id).toBe(approval.id);
  expect(packet.snapshot.head).toMatch(/^[0-9a-f]{40}$/);
  const reference = panel.getByRole('region', { name: `固定上下文包 ${packet.id}`, exact: true });
  await reference.getByRole('button', { name: '展开固定上下文', exact: true }).click();
  await expect(reference).toContainText(packet.snapshot.head);
  expect((await api.get(`/context-packets/${packet.id}`) as S['S2ContextPacketResultV1']).context_packet).toEqual(packet);
  const database = new DatabaseSync('D:/Astrocyte-agent-assets/20261015/relay/H/shared-product-ad8413/local/astrocyte-workbench/v0.2-access/app-state/state.sqlite', { readOnly: true });
  try {
    // 批准事件以提案为 aggregate，固定包有独立事件；目标保存没有 outbox 事件。
    for (const id of [proposal.id, packet.id]) {
      const rows = database.prepare('SELECT event_id,event_type,aggregate_id FROM foundation_outbox WHERE aggregate_id=?').all(id) as typeof evidence.events;
      expect(rows.length).toBeGreaterThan(0);
      for (const event of rows) expect(database.prepare('SELECT aggregate_id FROM foundation_outbox WHERE event_id=?').get(event.event_id)).toEqual({ aggregate_id: id });
      evidence.events.push(...rows);
    }
    expect(evidence.events.map(event => event.event_type).sort()).toEqual(['workspace.context_packet_built', 'workspace.proposal_approved']);
    const approvedEvent = database.prepare('SELECT payload FROM foundation_outbox WHERE aggregate_id=? AND event_type=?').get(proposal.id, 'workspace.proposal_approved') as { payload: string };
    expect((JSON.parse(approvedEvent.payload) as { grant_id: string }).grant_id).toBe(approval.id);
    expect(database.prepare('SELECT COUNT(*) AS count FROM local_agent_sessions').get()).toEqual({ count: 0 });
    evidence.native_sessions = 0;
  } finally { database.close(); }
  evidence.stage = 'proposal_and_fixed_context_completed'; evidence.completed_at = new Date().toISOString(); await save();
  console.log(JSON.stringify(evidence));
} finally { await save(); await browser.close(); }
