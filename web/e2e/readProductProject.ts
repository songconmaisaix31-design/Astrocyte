import { chromium, expect } from '@playwright/test';
import { readFile, writeFile } from 'node:fs/promises';
import { resolve } from 'node:path';
import { DatabaseSync } from 'node:sqlite';
import { humanAPI } from '../../tests/s1/api.mjs';
import type { components } from '../src/api/schema';

type S = components['schemas'];
const origin = process.env.ASTROCYTE_TEST_PRODUCT_URL;
const output = process.env.ASTROCYTE_VERIFICATION_OUTPUT_DIR;
if (origin !== 'http://127.0.0.1:64505' || !output) throw new Error('需要同一产品入口和本轮项目原记录目录');
const directory = resolve(output);
const original = JSON.parse(await readFile(resolve(directory, 'product-project-human.json'), 'utf8')) as { project_id: string; space_id: string; settings_revision: number; denied_status: number; context_files: number; native_sessions: number };
expect(original.project_id).toBe('98389066-5e5a-4e82-982b-deab3e005273');
expect(original.denied_status).toBe(403);
expect(original.context_files).toBe(1);
expect(original.native_sessions).toBe(0);
const api = await humanAPI(origin);
const { items } = await api.get('/local-projects') as S['LocalProjectListV1'];
const project = items.find(item => item.id === original.project_id);
if (!project) throw new Error('同一产品当前项目列表没有原实际登记记录');
expect(project.settings.revision).toBe(original.settings_revision);
expect(project.space_id).toBe(original.space_id);
expect(project.settings).toMatchObject({ allow_directory: true, expand_references: false, allowed_subdirs: ['web/src/pages/workspace/proposals'], allowed_actions: ['read_context'], allowed_tools: [], external_model_cli: '', allow_agent_control: false });
const database = new DatabaseSync('D:/Astrocyte-agent-assets/20261015/relay/H/shared-product-ad8413/local/astrocyte-workbench/v0.2-access/app-state/state.sqlite', { readOnly: true });
let eventCount: number;
try {
  const saved = database.prepare('SELECT data FROM local_agent_projects WHERE id=?').get(project.id) as { data: string };
  expect((JSON.parse(saved.data) as S['LocalProjectV1']).settings).toEqual(project.settings);
  expect(database.prepare('SELECT COUNT(*) AS count FROM local_agent_sessions WHERE project_id=?').get(project.id)).toEqual({ count: 0 });
  const outbox = database.prepare('SELECT COUNT(*) AS count FROM foundation_outbox WHERE aggregate_id=?').get(project.id) as { count: number };
  expect(outbox.count).toBe(0);
  eventCount = outbox.count;
} finally { database.close(); }
const cognition: { path: string; status: number; code?: string; required_action?: string }[] = [];
for (const endpoint of ['/cognition/health', '/cognition/today', '/cognition/filters', '/relay/session-events?page=1&page_size=50']) {
  const result = await api.request(endpoint);
  const error = result.status === 200 ? undefined : (result.data as S['ErrorV1']).error;
  cognition.push({ path: endpoint, status: result.status, code: error?.code, required_action: error?.required_action });
}
const browser = await chromium.launch();
const commands: string[] = [];
try {
  const page = await browser.newPage({ viewport: { width: 1280, height: 720 } });
  page.on('request', request => {
    if (new URL(request.url()).pathname.startsWith('/api/v1/') && request.method() !== 'GET') commands.push(request.method());
  });
  await page.goto(`${origin}/workspace`);
  await page.getByRole('button', { name: '接入项目', exact: true }).click();
  const panel = page.locator('section').filter({ has: page.getByRole('heading', { name: '本地 Agent 与项目', exact: true }) }).first();
  await panel.getByRole('button', { name: /同实例公开前端项目.*查看项目与权限/ }).click();
  const manage = panel.getByRole('region', { name: '已登记项目管理' });
  await manage.getByText('项目小权限与模型处理许可', { exact: true }).click();
  await expect(manage.getByRole('checkbox', { name: 'B：允许读取明确目录 / 文件' })).toBeChecked();
  await expect(manage.getByRole('checkbox', { name: 'C：允许展开获准资料的引用' })).not.toBeChecked();
  await expect(manage.getByLabel('允许的项目子目录 / 文件（每行一项）', { exact: true })).toHaveValue('web/src/pages/workspace/proposals');
  await expect(manage.getByRole('checkbox', { name: '读取上下文', exact: true })).toBeChecked();
  await expect(manage.getByRole('button', { name: '启动获准原生会话', exact: true })).toBeDisabled();
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
  expect(commands).toEqual([]);
  const result = { project_id: project.id, settings_revision: project.settings.revision, native_sessions: 0, project_outbox_count: eventCount, project_event_contract: 'not_defined', writes: commands, cognition, completed_at: new Date().toISOString() };
  await writeFile(resolve(directory, 'product-project-readonly.json'), JSON.stringify(result, null, 2), { flag: 'wx' });
  console.log(JSON.stringify(result));
} finally { await browser.close(); }
