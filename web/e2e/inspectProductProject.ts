import { chromium, expect } from '@playwright/test';
import { mkdir, readFile, writeFile } from 'node:fs/promises';
import { resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { DatabaseSync } from 'node:sqlite';
import { humanAPI } from '../../tests/s1/api.mjs';
import type { components } from '../src/api/schema';

type S = components['schemas'];
const origin = process.env.ASTROCYTE_TEST_PRODUCT_URL;
const output = process.env.ASTROCYTE_VERIFICATION_OUTPUT_DIR;
if (origin !== 'http://127.0.0.1:64505' || !output) throw new Error('需要明确的同一产品入口和独占验证目录');
const directory = resolve(output);
await mkdir(directory, { recursive: true });
const root = fileURLToPath(new URL('../../', import.meta.url));
const allowedRoot = 'web/src/pages/workspace/proposals';
const filename = `${allowedRoot}/proposalDraft.ts`;
const text = await readFile(resolve(root, filename), 'utf8');
const evidence: {
  stage: string; started_at: string; completed_at?: string; project_id?: string; space_id?: string;
  settings_revision?: number; context_files?: number; denied_status?: number;
  operations: { method: string; path: string; operation_id: string; status?: number }[];
  project_outbox_count?: number; native_sessions?: number;
} = { stage: 'started', started_at: new Date().toISOString(), operations: [] };
const path = resolve(directory, 'product-project-human.json');
await writeFile(path, JSON.stringify(evidence, null, 2), { flag: 'wx' });
const save = () => writeFile(path, JSON.stringify(evidence, null, 2));
const browser = await chromium.launch();
try {
  const api = await humanAPI(origin);
  const context = await browser.newContext({ viewport: { width: 1920, height: 1080 } });
  const page = await context.newPage();
  page.setDefaultTimeout(30_000);
  page.on('request', request => {
    const url = new URL(request.url());
    if (url.pathname.startsWith('/api/v1/') && request.method() !== 'GET') {
      evidence.operations.push({ method: request.method(), path: url.pathname, operation_id: request.headers()['idempotency-key'] ?? '' });
    }
  });
  page.on('response', response => {
    const key = response.request().headers()['idempotency-key'];
    const operation = evidence.operations.find(item => item.operation_id === key);
    if (operation) operation.status = response.status();
  });
  await page.goto(`${origin}/workspace`);
  await page.getByRole('button', { name: '接入项目', exact: true }).click();
  const panel = page.locator('section').filter({ has: page.getByRole('heading', { name: '本地 Agent 与项目', exact: true }) }).first();
  await panel.getByText('登记项目 / 发现子项目', { exact: true }).click();
  await panel.getByLabel('新项目顶层空间名称', { exact: true }).fill('同实例公开前端项目核查');
  evidence.stage = 'creating_project_space'; await save();
  await panel.getByRole('button', { name: '创建空间用于此项目', exact: true }).click();
  await expect(panel.getByLabel('关联项目顶层空间', { exact: true })).not.toHaveValue('');
  evidence.space_id = await panel.getByLabel('关联项目顶层空间', { exact: true }).inputValue();
  await panel.getByLabel('项目名称', { exact: true }).fill('同实例公开前端项目');
  await panel.getByLabel('项目绝对目录', { exact: true }).fill(root);
  evidence.stage = 'registering_public_project'; await save();
  const registerResponse = page.waitForResponse(response => response.url().endsWith('/api/v1/local-projects') && response.request().method() === 'POST');
  await panel.getByRole('button', { name: '登记此项目', exact: true }).click();
  const registered = await registerResponse;
  expect(registered.status()).toBe(200);
  const { project } = await registered.json() as S['LocalProjectResultV1'];
  evidence.project_id = project.id; await save();
  expect(project.settings).toMatchObject({ allow_directory: false, expand_references: false, external_model_cli: '', allow_agent_control: false });
  const manage = panel.getByRole('region', { name: '已登记项目管理' });
  await manage.getByText('项目小权限与模型处理许可', { exact: true }).click();
  await manage.getByRole('checkbox', { name: 'B：允许读取明确目录 / 文件' }).check();
  await expect(manage.getByRole('checkbox', { name: 'C：允许展开获准资料的引用' })).not.toBeChecked();
  await manage.getByLabel('允许的项目子目录 / 文件（每行一项）', { exact: true }).fill(allowedRoot);
  await manage.getByRole('checkbox', { name: '读取上下文', exact: true }).check();
  evidence.stage = 'saving_narrow_permission'; await save();
  const settingsResponse = page.waitForResponse(response => response.url().endsWith(`/local-projects/${project.id}/settings`) && response.request().method() === 'PUT');
  await manage.getByRole('button', { name: '保存此项目许可', exact: true }).click();
  const settings = await settingsResponse;
  expect(settings.status()).toBe(200);
  const { project: saved } = await settings.json() as S['LocalProjectResultV1'];
  expect(saved.settings).toMatchObject({ allow_directory: true, expand_references: false, allowed_subdirs: [allowedRoot], allowed_actions: ['read_context'], allowed_tools: [], external_model_cli: '', allow_agent_control: false });
  evidence.settings_revision = saved.settings.revision; await save();
  const native = manage.getByRole('region', { name: '项目原生 Agent 操作' });
  await native.getByLabel('本次读取的获准文件（每行一项）', { exact: true }).fill(filename);
  evidence.stage = 'reading_exact_public_file'; await save();
  const reading = page.waitForResponse(response => response.url().endsWith(`/local-projects/${project.id}/context`) && response.request().method() === 'POST');
  await native.getByRole('button', { name: '预览本次获准上下文', exact: true }).click();
  const response = await reading;
  expect(response.status()).toBe(200);
  const packet = await response.json() as S['LocalContextPacketV1'];
  expect(packet.files).toEqual([expect.objectContaining({ path: filename, text })]);
  expect(packet.materials).toEqual([]);
  evidence.context_files = packet.files.length; await save();
  await native.getByText(/^实际上下文 ·/).click();
  await expect(native).toContainText(text.trim());
  await expect(native.getByRole('button', { name: '启动获准原生会话', exact: true })).toBeDisabled();
  const deniedKey = crypto.randomUUID();
  const denied = await api.request(`/local-projects/${project.id}/context`, { method: 'POST', key: deniedKey, body: { schema_version: 1, request_id: deniedKey, expected_version: saved.settings.revision, files: ['AGENTS.md'], references: [] } });
  evidence.denied_status = denied.status; await save();
  expect(denied.status).toBe(403);
  await page.reload();
  const current = (await api.get('/local-projects') as S['LocalProjectListV1']).items.find(item => item.id === project.id);
  expect(current?.settings).toEqual(saved.settings);
  const database = new DatabaseSync('D:/Astrocyte-agent-assets/20261015/relay/H/shared-product-ad8413/local/astrocyte-workbench/v0.2-access/app-state/state.sqlite', { readOnly: true });
  try {
    const rows = database.prepare('SELECT data FROM local_agent_projects WHERE id=?').all(project.id) as { data: string }[];
    expect(rows).toHaveLength(1);
    expect((JSON.parse(rows[0].data) as S['LocalProjectV1']).settings).toEqual(saved.settings);
    const sessions = database.prepare('SELECT COUNT(*) AS count FROM local_agent_sessions WHERE project_id=?').get(project.id) as { count: number };
    expect(sessions.count).toBe(0);
    evidence.native_sessions = sessions.count;
    // 项目登记和设置直接保存项目；当前生产接口没有项目 outbox 事件契约。
    const events = database.prepare('SELECT COUNT(*) AS count FROM foundation_outbox WHERE aggregate_id=?').get(project.id) as { count: number };
    evidence.project_outbox_count = events.count;
    expect(events.count).toBe(0);
  } finally { database.close(); }
  expect(evidence.operations.every(item => item.operation_id && (item.status === 200 || item.path === '/api/v1/project-spaces' && item.status === 201))).toBe(true);
  evidence.stage = 'human_project_completed'; evidence.completed_at = new Date().toISOString(); await save();
  console.log(JSON.stringify(evidence));
} finally { await save(); await browser.close(); }
