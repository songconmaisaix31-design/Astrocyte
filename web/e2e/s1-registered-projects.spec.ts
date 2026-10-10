/// <reference types="node" />
import { test, expect } from '@playwright/test';
import { stat, readFile } from 'node:fs/promises';
import { resolve, join } from 'node:path';
import { randomUUID } from 'node:crypto';
import process from 'node:process';
import { startS1Server } from '../../tests/s1/server.mjs';
import { humanAPI } from '../../tests/s1/api.mjs';
import type { components } from '../src/api/schema';

type S = components['schemas'];
test('actual Orca discovery remains a read-only candidate list until human registration and persists with default A permissions', async ({ page }, testInfo) => {
  test.skip(process.env.ASTROCYTE_TEST_REGISTERED_PROJECTS !== '1', 'Requires controller-assigned actual local-project/browser slot.');
  test.setTimeout(600_000);
  const reusePath = process.env.ASTROCYTE_PROJECT_REUSE_OWNED_TEMP;
  const ownedRoot = process.env.ASTROCYTE_PROJECT_REUSE_APPROVED_ROOT;
  if (reusePath) expect(ownedRoot && resolve(reusePath) === resolve('C:/Users/DW/AppData/Local/Temp/astrocyte-s1-1lHfyS') && resolve(ownedRoot) === resolve(reusePath)).toBe(true);
  const startupAt = Date.now();
  const server = await startS1Server({ browser: true, ...(reusePath ? { reuseOwnedTemporary: { path: reusePath, ownedRoot: ownedRoot! } } : {}) });
  let api = await humanAPI(server.apiURL);
  const read = async <T,>(path: string) => await api.get(path) as T;
  let primaryError: unknown;
  try {
    await page.goto(`${server.webURL}/workspace`);
    await page.getByRole('button', { name: '接入项目', exact: true }).click();
    await page.locator('summary').filter({ hasText: /^从 Orca 登记目录选择项目$/ }).click();
    const discovery = page.getByRole('region', { name: 'Orca 登记目录发现' });
    let actual = await read<S['RegisteredProjectDiscoveryResultV1']>('/local-projects/registered');
    await expect.poll(async () => { actual = await read<S['RegisteredProjectDiscoveryResultV1']>('/local-projects/registered'); return actual.snapshot.observed_at && Date.parse(actual.snapshot.observed_at) >= startupAt ? actual.snapshot.status : 'unknown'; }, { timeout: 130_000 }).toMatch(/complete|partial/);
    expect(actual.snapshot.observed_at).toBeTruthy();
    expect(actual.snapshot.projects.length).toBeGreaterThan(0);
    const initialProjects = (await read<S['LocalProjectListV1']>('/local-projects')).items;
    if (!reusePath) expect(initialProjects).toEqual([]);
    const initialProjectCount = server.query('SELECT COUNT(*) AS count FROM local_agent_projects')[0].count;
    await discovery.getByRole('button', { name: '重载已保存目录清单', exact: true }).click();
    await expect(discovery.locator('li')).toHaveCount(actual.snapshot.projects.length);
    const cachedBeforeGET = server.query('SELECT data FROM local_project_discovery');
    expect(await read<S['RegisteredProjectDiscoveryResultV1']>('/local-projects/registered')).toEqual(actual);
    expect(server.query('SELECT data FROM local_project_discovery')).toEqual(cachedBeforeGET);
    const selected = actual.snapshot.projects.find(item => item.source === 'orca_registered' && item.name === 'Astrocyte') ?? actual.snapshot.projects.find(item => resolve(item.root) === resolve(process.cwd(), '..')) ?? actual.snapshot.projects[0];
    expect((await stat(selected.root)).isDirectory()).toBe(true);
    const row = discovery.locator('li').filter({ has: page.getByText(selected.root, { exact: true }).and(page.locator('code')) });
    await expect(row).toHaveCount(1);
    await expect(row).toContainText(selected.activity.created_with_cli ?? '未记录');
    await expect(row).toContainText(selected.git.branch ?? '未知');
    const panel = page.locator('section').filter({ has: page.getByRole('heading', { name: '本地 Agent 与项目', exact: true }) }).first();
    const prior = initialProjects.find(item => resolve(item.root) === resolve(selected.root));
    let registered: S['LocalProjectResultV1'];
    if (prior) {
      registered = { schema_version: 1, project: prior };
      await panel.getByRole('button').filter({ has: page.getByRole('heading', { name: prior.name, exact: true }) }).click();
    } else {
      await row.getByRole('button', { name: '选择此目录登记', exact: true }).click();
      await expect(panel.getByLabel('项目绝对目录', { exact: true })).toHaveValue(selected.root);
      await expect(panel.getByLabel('项目名称', { exact: true })).toHaveValue(selected.name);
      expect(server.query('SELECT COUNT(*) AS count FROM local_agent_projects')[0].count).toBe(initialProjectCount);
      await panel.getByLabel('新项目顶层空间名称', { exact: true }).fill('Orca 实际目录人工登记');
      await panel.getByRole('button', { name: '创建空间用于此项目', exact: true }).click();
      await expect(panel.getByLabel('关联项目顶层空间', { exact: true })).not.toHaveValue('');
      const registeredResponse = page.waitForResponse(response => response.url().endsWith('/local-projects') && response.request().method() === 'POST');
      await panel.getByRole('button', { name: '登记此项目', exact: true }).click();
      const registerHTTP = await registeredResponse;
      expect(registerHTTP.status()).toBe(200);
      registered = await registerHTTP.json() as S['LocalProjectResultV1'];
    }
    expect(resolve(registered.project.root)).toBe(resolve(selected.root));
    if (!prior || prior.settings.revision === 1) expect(registered.project.settings).toMatchObject({ allow_directory: false, expand_references: false, external_model_cli: '', allow_agent_control: false, allowed_actions: [], allowed_tools: [] });
    await expect(row.getByRole('button', { name: '已登记此目录', exact: true })).toBeDisabled();
    let saved = server.query('SELECT data FROM local_agent_projects WHERE id=?', registered.project.id);
    expect(saved).toHaveLength(1);
    expect(JSON.parse(String(saved[0].data))).toEqual(registered.project);
    const refresh = page.waitForResponse(response => response.url().endsWith('/local-projects/registered/refresh') && response.request().method() === 'POST', { timeout: 140_000 });
    await discovery.getByRole('button', { name: '重新读取登记目录', exact: true }).click();
    expect((await refresh).status()).toBe(200);
    expect(server.query('SELECT data FROM local_agent_projects WHERE id=?', registered.project.id)).toEqual(saved);
    const manage = panel.getByRole('region', { name: '已登记项目管理' });
    await manage.getByText('项目小权限与模型处理许可', { exact: true }).click();
    await manage.getByRole('checkbox', { name: 'B：允许读取明确目录 / 文件' }).check();
    await expect(manage.getByRole('checkbox', { name: 'C：允许展开获准资料的引用' })).not.toBeChecked();
    await manage.getByLabel('允许的项目子目录 / 文件（每行一项）', { exact: true }).fill('README.md');
    const settingsResponse = page.waitForResponse(response => response.url().endsWith(`/local-projects/${registered.project.id}/settings`) && response.request().method() === 'PUT');
    await manage.getByRole('button', { name: '保存此项目许可', exact: true }).click();
    const settingsHTTP = await settingsResponse;
    expect(settingsHTTP.status()).toBe(200);
    const settings = await settingsHTTP.json() as S['LocalProjectResultV1'];
    expect(settings.project.settings).toMatchObject({ allow_directory: true, allowed_subdirs: ['README.md'], expand_references: false, external_model_cli: '', allow_agent_control: false, allowed_actions: [] });
    const native = manage.getByRole('region', { name: '项目原生 Agent 操作' });
    await native.getByLabel('本次读取的获准文件（每行一项）', { exact: true }).fill('README.md');
    const contextResponse = page.waitForResponse(response => response.url().endsWith(`/local-projects/${registered.project.id}/context`) && response.request().method() === 'POST');
    await native.getByRole('button', { name: '预览本次获准上下文', exact: true }).click();
    const contextHTTP = await contextResponse;
    expect(contextHTTP.status()).toBe(200);
    const context = await contextHTTP.json() as S['LocalContextPacketV1'];
    expect(context.files).toEqual([expect.objectContaining({ path: 'README.md', text: await readFile(join(selected.root, 'README.md'), 'utf8') })]);
    expect(context.materials).toEqual([]);
    saved = server.query('SELECT data FROM local_agent_projects WHERE id=?', registered.project.id);
    const agentAccess = manage.locator('details').filter({ has: page.getByText('由你授权的项目 Agent 身份', { exact: true }) });
    await agentAccess.getByText('由你授权的项目 Agent 身份', { exact: true }).click();
    await agentAccess.getByLabel('被授权 Agent 身份', { exact: true }).fill('actual-project-read-only');
    await agentAccess.getByRole('checkbox', { name: '读取已纳入上下文', exact: true }).check();
    const grantResponse = page.waitForResponse(response => response.url().endsWith(`/local-projects/${registered.project.id}/grants`) && response.request().method() === 'POST');
    await agentAccess.getByRole('button', { name: '保存此身份的动作授权', exact: true }).click();
    expect((await grantResponse).status()).toBe(200);
    async function humanWrite(endpoint: string, fields: Record<string, unknown>) {
      const key = randomUUID();
      const response = await fetch(`${server.apiURL}/api/v1${endpoint}`, { method: 'POST', headers: { Cookie: api.cookie, Origin: server.apiURL, 'Content-Type': 'application/json', 'X-CSRF-Token': api.session.csrf_token, 'Idempotency-Key': key }, body: JSON.stringify({ schema_version: 1, expected_version: 1, request_id: key, ...fields }) });
      expect(response.status).toBe(200);
      return await response.json() as Record<string, unknown>;
    }
    // Temporary API-issued test identity stays in memory; no token is attached,
    // displayed, logged or placed in browser request traces.
    const issued = await humanWrite(`/local-projects/${registered.project.id}/agent-token`, { expected_version: settings.project.settings.revision, agent_id: 'actual-project-read-only' }) as unknown as S['ProjectAgentTokenResultV1'];
    const token = issued.credential.token;
    async function agentRequest(endpoint: string, fields?: Record<string, unknown>, method = fields ? 'POST' : 'GET') {
      const key = randomUUID();
      return fetch(`${server.apiURL}/api/v1${endpoint}`, { method, headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json', 'Idempotency-Key': key }, ...(fields ? { body: JSON.stringify({ schema_version: 1, expected_version: 1, request_id: key, ...fields }) } : {}) });
    }
    const allowed = await agentRequest(`/local-projects/${registered.project.id}/context`, { references: [], files: ['README.md'] });
    expect(allowed.status).toBe(200);
    expect((await allowed.json() as S['LocalContextPacketV1']).files).toEqual(context.files);
    expect((await agentRequest(`/local-projects/${registered.project.id}/context`, { references: [], files: ['AGENTS.md'] })).status).toBe(403);
    expect((await agentRequest(`/local-projects/${registered.project.id}/sessions`)).status).toBe(403);
    expect((await agentRequest(`/local-projects/${registered.project.id}/settings`, { expected_revision: settings.project.settings.revision, settings: settings.project.settings }, 'PUT')).status).toBe(403);
    const other = actual.snapshot.projects.find(item => item.root !== selected.root)!;
    expect(other).toBeTruthy();
    const otherProject = await humanWrite('/local-projects', { name: other.name, root: other.root, space_id: registered.project.space_id }) as unknown as S['LocalProjectResultV1'];
    expect((await agentRequest(`/local-projects/${otherProject.project.id}/context`, { references: [], files: [] })).status).toBe(403);
    const revokeResponse = page.waitForResponse(response => response.url().endsWith(`/local-projects/${registered.project.id}/grants/revoke`) && response.request().method() === 'POST');
    await agentAccess.getByRole('button', { name: '撤销此项目身份许可', exact: true }).click();
    expect((await revokeResponse).status()).toBe(200);
    expect((await agentRequest(`/local-projects/${registered.project.id}/context`, { references: [], files: ['README.md'] })).status).toBe(403);
    for (const width of [1280, 1920]) {
      await page.setViewportSize({ width, height: width === 1280 ? 720 : 1080 });
      await discovery.scrollIntoViewIfNeeded();
      expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
      await page.screenshot({ path: testInfo.outputPath(`actual-registered-projects-${width}.png`) });
    }
    await testInfo.attach('actual-discovery-snapshot', { body: JSON.stringify(actual, null, 2), contentType: 'application/json' });
    await testInfo.attach('actual-human-project-registration', { body: JSON.stringify(registered, null, 2), contentType: 'application/json' });
    await server.restart();
    api = await humanAPI(server.apiURL);
    expect((await read<S['LocalProjectListV1']>('/local-projects')).items).toEqual(expect.arrayContaining([settings.project, otherProject.project]));
    expect(server.query('SELECT data FROM local_agent_projects WHERE id=?', registered.project.id)).toEqual(saved);
    expect(server.query('SELECT COUNT(*) AS count FROM local_agent_sessions')[0].count).toBe(0);
    expect(server.query('SELECT COUNT(*) AS count FROM attention_jobs')[0].count).toBe(0);
    expect((await read<S['MissionListV1']>('/missions')).items).toEqual([]);
    await page.reload();
    await page.getByRole('button', { name: '接入项目', exact: true }).click();
    await expect(discovery.getByRole('button', { name: '已登记此目录', exact: true })).toHaveCount(2);
  } catch (error) {
    primaryError = error;
    await testInfo.attach('actual-primary-error', { body: error instanceof Error ? `${error.name}: ${error.message}\n${error.stack ?? ''}` : String(error), contentType: 'text/plain' });
    throw error;
  } finally {
    await testInfo.attach('actual-owned-project-store', { body: JSON.stringify({ temporary: server.temporary, dataDir: server.dataDir }), contentType: 'application/json' });
    await server.close({ preserveData: true, primaryError });
  }
});
