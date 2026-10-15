import { expect, test, type Locator, type Page, type Response } from '@playwright/test';
import { execFileSync } from 'node:child_process';
import { statSync } from 'node:fs';
import { mkdir, readdir, stat } from 'node:fs/promises';
import { isAbsolute, join } from 'node:path';
import { startS1Server } from '../../tests/s1/server.mjs';
import { humanAPI } from '../../tests/s1/api.mjs';
import { repositoryRoot } from './realService';
import type { components } from '../src/api/schema';

type S = components['schemas'];
type API = Awaited<ReturnType<typeof humanAPI>>;
const head = '4317d610cbe80aa6c4e22ea4da5e8dde013de4ed';

function input(name: string) {
  const value = process.env[`ASTROCYTE_TEST_S4_${name}`];
  if (!value || !isAbsolute(value)) throw new Error(`真实 S4 验证需要明确的绝对路径 ASTROCYTE_TEST_S4_${name}`);
  return value;
}

function git(root: string, ...args: string[]) {
  return execFileSync('git', ['-C', root, ...args], { encoding: 'utf8' }).trim();
}

async function configuredService() {
  const a = input('WORKTREE_A'), b = input('WORKTREE_B');
  const moduleCache = process.env.GOMODCACHE;
  if (!moduleCache || !isAbsolute(moduleCache)) throw new Error('真实 S4 验证需要明确的已有 Go 模块缓存绝对路径 GOMODCACHE');
  for (const root of [a, b]) {
    const remote = git(root, 'remote', 'get-url', 'origin');
    expect(['https://github.com/songconmaisaix31-design/astrocyte-s4-public-context-20261015', 'https://github.com/songconmaisaix31-design/astrocyte-s4-public-context-20261015.git'].includes(remote)).toBe(true);
    expect(git(root, 'rev-parse', 'HEAD')).toBe(head);
  }
  const before = [git(a, 'status', '--porcelain=v1', '--untracked-files=all'), git(b, 'status', '--porcelain=v1', '--untracked-files=all')];
  const governance = await Promise.all([aociFiles(a), aociFiles(b)]);
  const server = await startS1Server({ browser: true, env: async ({ temporary }) => {
    const profile = join(temporary, 'program-profile');
    const appdata = join(temporary, 'program-appdata');
    const localappdata = join(temporary, 'program-localappdata');
    await Promise.all([profile, appdata, localappdata].map(path => mkdir(path)));
    return {
      HOME: profile, USERPROFILE: profile, APPDATA: appdata, LOCALAPPDATA: localappdata,
      GOMODCACHE: moduleCache, GOPROXY: 'off',
      ASTROCYTE_PROJECT_DISCOVERY_ROOTS: JSON.stringify([repositoryRoot, a, b]),
      ASTROCYTE_REPOSITORY_ENVIRONMENT_ID: 'product-ui-s4-browser',
      ASTROCYTE_CODEGRAPH_NODE: input('NODE'), ASTROCYTE_CODEGRAPH_SDK: input('SDK'),
      ASTROCYTE_CODEGRAPH_SCRIPT: join(repositoryRoot, 'internal/adapters/repositorycontext/codegraph.mjs'),
      ASTROCYTE_AOCI_EXECUTABLE: input('AOCI'), ASTROCYTE_REPOSITORY_MAX_FILES: '100',
      ASTROCYTE_REPOSITORY_MAX_BYTES: '8388608', ASTROCYTE_REPOSITORY_MAX_RESULTS: '100', ASTROCYTE_REPOSITORY_MAX_DEPTH: '3',
    };
  } });
  const api = await humanAPI(server.apiURL);
  return { server, api, a, b, async unchanged() {
    expect(git(a, 'status', '--porcelain=v1', '--untracked-files=all') === before[0]).toBe(true);
    expect(git(b, 'status', '--porcelain=v1', '--untracked-files=all') === before[1]).toBe(true);
    expect(JSON.stringify(await Promise.all([aociFiles(a), aociFiles(b)])) === JSON.stringify(governance)).toBe(true);
  } };
}

async function setPermission(api: API, project: S['LocalProjectV1'], allowDirectory: boolean) {
  const result = await api.write(`/local-projects/${project.id}/settings`, {
    expected_revision: project.settings.revision,
    settings: { allow_directory: allowDirectory, allowed_subdirs: ['.'], allowed_actions: ['read_context'], allowed_tools: [], external_model_cli: '', allow_agent_control: false },
  }, { method: 'PUT' }) as S['LocalProjectResultV1'];
  return result.project;
}

async function register(api: API, root: string, name: string) {
  const { space } = await api.write('/project-spaces', { title: name }, { status: 201 }) as S['ProjectSpaceResultV1'];
  const { project } = await api.write('/local-projects', { root, name, space_id: space.id }) as S['LocalProjectResultV1'];
  return setPermission(api, project, true);
}

async function openPanel(page: Page, webURL: string, projectID: string, mode: 'search' | 'map') {
  await page.goto(`${webURL}/workspace`);
  const panel = page.getByRole('region', { name: '仓库上下文', exact: true });
  await panel.getByLabel('仓库上下文项目', { exact: true }).selectOption(projectID);
  await fillQuery(panel, mode);
  return panel;
}

async function fillQuery(panel: Locator, mode: 'search' | 'map') {
  await panel.getByLabel('查询用途', { exact: true }).fill('s4_ui_acceptance');
  await panel.getByLabel('查询项目相对路径', { exact: true }).fill('.');
  for (const [label, value] of [['最多读取文件数', '100'], ['最多读取字节数', '8388608'], ['最多返回结果数', '100'], ['关系展开深度', '3']]) await panel.getByLabel(label, { exact: true }).fill(value);
  await panel.getByRole('button', { name: mode === 'search' ? '代码搜索' : '仓库说明', exact: true }).click();
  if (mode === 'search') await panel.getByLabel('代码搜索内容', { exact: true }).fill('Count');
}

function response(page: Page, id: string, endpoint: string) {
  return page.waitForResponse(value => new URL(value.url()).pathname === `/api/v1/local-projects/${id}/repository-context/${endpoint}`);
}

async function reread(page: Page, panel: Locator, id: string, endpoint: string) {
  const button = panel.getByRole('button', { name: '重新读取已提交查询', exact: true });
  await expect(button).toBeEnabled({ timeout: 30_000 });
  const queried = response(page, id, endpoint);
  await button.click();
  return await queried;
}

async function result(value: Response) {
  expect(value.status()).toBe(200);
  return (await value.json() as S['RepositoryResultEnvelopeV1']).repository_context;
}

async function indexFromMissing(page: Page, panel: Locator, project: S['LocalProjectV1']) {
  const missing = response(page, project.id, 'search');
  await panel.getByRole('button', { name: '读取仓库上下文', exact: true }).click();
  expect((await missing).status()).toBe(404);
  const indexed = response(page, project.id, 'index');
  await panel.getByRole('button', { name: '建立代码索引', exact: true }).click();
  const saved = await indexed;
  expect(saved.status()).toBe(200);
  const index = (await saved.json() as S['RepositoryIndexEnvelopeV1']).repository_context;
  expect(index.state).toBe('ready');
  expect(index.snapshot.head).toBe(head);
  expect(index.snapshot.settings_revision).toBe(project.settings.revision);
  expect(index.snapshot).not.toHaveProperty('root');
  const found = await result(await reread(page, panel, project.id, 'search'));
  const symbol = found.symbols?.find(item => item.name === 'Count' && item.source_span.path === 'textstats/stats.go');
  if (!symbol) throw new Error('实际 CodeGraph 没有返回公开统计项目的 Count 符号及来源位置');
  expect(symbol.source_span.start_line).toBeGreaterThan(0);
  expect(found.gaps?.includes('semantic_embedding_not_available')).toBe(true);
  await expect(panel.getByText(symbol.id, { exact: true })).toBeVisible();
  return { index, symbol };
}

test('真实配置下双worktree独立建图、源码关系、权限撤销与过期重建及SQLite重启', async ({ page }) => {
  test.setTimeout(240_000);
  const owned = await configuredService();
  const { server, api } = owned;
  try {
    let a = await register(api, owned.a, '公开统计工作树 A');
    const b = await register(api, owned.b, '公开统计工作树 B');
    let panel = await openPanel(page, server.webURL, a.id, 'search');
    const first = await indexFromMissing(page, panel, a);
    const symbolRow = panel.getByRole('region', { name: '实际仓库上下文结果', exact: true }).locator('li').filter({ has: page.getByText(first.symbol.id, { exact: true }) }).filter({ has: page.getByRole('button', { name: '选择此符号查询影响', exact: true }) });
    await expect(symbolRow).toHaveCount(1);
    await symbolRow.getByRole('button', { name: '选择此符号查询影响', exact: true }).click();
    await expect(panel.getByLabel('待查影响的符号 ID', { exact: true })).toHaveValue(first.symbol.id);
    const impact = response(page, a.id, 'impact');
    await panel.getByRole('button', { name: '读取仓库上下文', exact: true }).click();
    const relations = await result(await impact);
    expect((relations.relations?.length ?? 0) > 0).toBe(true);
    expect(relations.relations?.every(item => item.evidence_kind === 'codegraph_static_heuristic')).toBe(true);
    await expect(panel.getByText('codegraph_static_heuristic', { exact: true }).first()).toBeVisible();
    await panel.getByLabel('仓库上下文项目', { exact: true }).selectOption(b.id);
    await fillQuery(panel, 'search');
    const second = await indexFromMissing(page, panel, b);
    expect(second.index.snapshot.repo_id).toBe(first.index.snapshot.repo_id);
    expect(second.index.snapshot.worktree_id).not.toBe(first.index.snapshot.worktree_id);
    expect(second.index.generation).not.toBe(first.index.generation);
    await panel.getByLabel('仓库上下文项目', { exact: true }).selectOption(a.id);
    await fillQuery(panel, 'search');
    let queried = response(page, a.id, 'search');
    await panel.getByRole('button', { name: '读取仓库上下文', exact: true }).click();
    await result(await queried);
    a = await setPermission(api, a, false);
    queried = response(page, a.id, 'search');
    await panel.getByRole('button', { name: '重新读取已提交查询', exact: true }).click();
    expect((await queried).status()).toBe(403);
    await expect(panel.getByRole('alert').filter({ hasText: '数据可能已过期' })).toBeVisible();
    await expect(panel.getByRole('button', { name: '重建代码索引', exact: true })).toBeDisabled();
    a = await setPermission(api, a, true);
    queried = response(page, a.id, 'search');
    await panel.getByRole('button', { name: '重新读取已提交查询', exact: true }).click();
    expect((await queried).status()).toBe(409);
    const rebuilt = response(page, a.id, 'index');
    await panel.getByRole('button', { name: '重建代码索引', exact: true }).click();
    expect((await rebuilt).status()).toBe(200);
    const current = await result(await reread(page, panel, a.id, 'search'));
    expect(current.index.generation).not.toBe(first.index.generation);
    expect(current.index.snapshot.settings_revision).toBe(a.settings.revision);
    await server.restart();
    const restarted = await humanAPI(server.apiURL);
    const saved = await restarted.request(`/local-projects/${b.id}/repository-context/search`, { method: 'POST', body: { schema_version: 1, purpose: 's4_ui_acceptance', roots: ['.'], budget: { max_files: 100, max_bytes: 8388608, max_results: 100, max_depth: 3 }, query: 'Count' } });
    expect(saved.status).toBe(200);
    expect((saved.data as S['RepositoryResultEnvelopeV1']).repository_context.index.generation).toBe(second.index.generation);
    panel = await openPanel(page, server.webURL, b.id, 'search');
    queried = response(page, b.id, 'search');
    await panel.getByRole('button', { name: '读取仓库上下文', exact: true }).click();
    expect((await result(await queried)).index.generation).toBe(second.index.generation);
    expect(server.query('SELECT COUNT(*) AS count FROM local_agent_sessions')).toEqual([{ count: 0 }]);
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
  } finally { await server.close({ preserveData: true }); await owned.unchanged(); }
});

async function aociFiles(root: string) {
  const directory = join(root, '.aoci');
  if (!statSync(directory, { throwIfNoEntry: false })) return null;
  const entries = await readdir(directory, { recursive: true, withFileTypes: true });
  return Promise.all(entries.filter(entry => entry.isFile()).map(async entry => {
    const metadata = await stat(join(entry.parentPath, entry.name));
    return `${entry.parentPath}/${entry.name}:${metadata.size}:${metadata.mtimeMs}`;
  })).then(values => values.sort());
}

test('真实AOCI人工刷新及保存地图读取，来源治理文件保持只读', async ({ page }) => {
  test.setTimeout(180_000);
  const owned = await configuredService();
  const { server, api } = owned;
  try {
    const before = await aociFiles(owned.a);
    expect(before?.length).toBe(10);
    const project = await register(api, owned.a, '公开工作树 A 的实际治理地图');
    const panel = await openPanel(page, server.webURL, project.id, 'map');
    const missing = response(page, project.id, 'map');
    await panel.getByRole('button', { name: '读取仓库上下文', exact: true }).click();
    expect((await missing).status()).toBe(404);
    const refreshed = response(page, project.id, 'map/refresh');
    await panel.getByRole('button', { name: '刷新仓库说明', exact: true }).click();
    expect((await refreshed).status()).toBe(200);
    const map = await result(await reread(page, panel, project.id, 'map'));
    expect(map.map_entries?.length).toBe(6);
    expect(map.map_entries?.every(entry => entry.evidence_kind === 'aoci_model_authored_fras')).toBe(true);
    for (const entry of map.map_entries ?? []) {
      const row = panel.getByRole('region', { name: '实际仓库上下文结果', exact: true }).locator('li').filter({ has: page.getByText(entry.path, { exact: true }) });
      await expect(row).toHaveCount(1);
      expect(await row.getByText('职责', { exact: true }).locator('..').innerText().then(value => value.includes(entry.responsibility))).toBe(true);
      expect(await row.getByText('约束', { exact: true }).locator('..').innerText().then(value => value.includes(entry.constraints))).toBe(true);
    }
    expect(JSON.stringify(await aociFiles(owned.a)) === JSON.stringify(before)).toBe(true);
  } finally { await server.close({ preserveData: true }); await owned.unchanged(); }
});

test('公开B工作树缺少真实AOCI治理索引时保留失败与空存储', async ({ page }) => {
  test.setTimeout(180_000);
  const owned = await configuredService();
  const { server, api } = owned;
  try {
    const project = await register(api, owned.b, '公开工作树 B 的治理缺口');
    const panel = await openPanel(page, server.webURL, project.id, 'map');
    const missing = response(page, project.id, 'map');
    await panel.getByRole('button', { name: '读取仓库上下文', exact: true }).click();
    expect((await missing).status()).toBe(404);
    const refreshed = response(page, project.id, 'map/refresh');
    await panel.getByRole('button', { name: '刷新仓库说明', exact: true }).click();
    const refused = await refreshed;
    expect(refused.status()).toBe(400);
    expect((await refused.json() as S['ErrorV1']).error.code).toBe('evidence_missing');
    expect(server.query('SELECT COUNT(*) AS count FROM repository_context_indexes WHERE project_id=? AND backend=?', project.id, 'aoci')).toEqual([{ count: 0 }]);
    await expect(panel.getByRole('alert').filter({ hasText: '缺少可核查的仓库上下文' })).toBeVisible();
  } finally { await server.close({ preserveData: true }); await owned.unchanged(); }
});

test('真实同步与预算拒绝保留输入，离线维护未知锁跨重开与刷新且无重发', async ({ page, context }) => {
  test.setTimeout(240_000);
  const owned = await configuredService();
  const { server, api } = owned;
  try {
    const project = await register(api, owned.b, '公开工作树 B 的同步与预算核查');
    const mutations: string[] = [];
    page.on('request', request => {
      if (new URL(request.url()).pathname === `/api/v1/local-projects/${project.id}/repository-context/index`) mutations.push(request.method());
    });
    let panel = await openPanel(page, server.webURL, project.id, 'search');
    const first = await indexFromMissing(page, panel, project);
    const syncing = response(page, project.id, 'index');
    await panel.getByRole('button', { name: '同步代码索引', exact: true }).click();
    const synced = await syncing;
    expect(synced.status()).toBe(200);
    expect(synced.request().postDataJSON().action).toBe('sync');
    const current = (await synced.json() as S['RepositoryIndexEnvelopeV1']).repository_context;
    expect(current.state).toBe('ready');
    expect(current.generation).not.toBe(first.index.generation);
    expect((await result(await reread(page, panel, project.id, 'search'))).index.generation).toBe(current.generation);
    expect(mutations).toEqual(['POST', 'POST']);

    await panel.getByLabel('最多读取文件数', { exact: true }).fill('101');
    const limited = response(page, project.id, 'search');
    await panel.getByRole('button', { name: '读取仓库上下文', exact: true }).click();
    const refused = await limited;
    expect(refused.status()).toBe(409);
    expect((await refused.json() as S['ErrorV1']).error.code).toBe('budget_exhausted');
    await expect(panel.getByLabel('最多读取文件数', { exact: true })).toHaveValue('101');
    const queriedResult = panel.getByRole('region', { name: '实际仓库上下文结果', exact: true });
    await expect(queriedResult.getByRole('alert')).toContainText('API 409 [budget_exhausted]');
    await expect(queriedResult).toContainText('101 个文件');
    await expect(queriedResult.getByText(first.symbol.id, { exact: true })).toHaveCount(0);
    await expect(panel.getByRole('region', { name: '已保存的索引结果', exact: true })).toContainText(current.generation);
    expect(server.query('SELECT generation FROM repository_context_indexes WHERE project_id=? AND backend=?', project.id, 'codegraph')).toEqual([{ generation: current.generation }]);
    await expect(panel.getByRole('button', { name: '同步代码索引', exact: true })).toBeDisabled();
    await panel.getByLabel('最多读取文件数', { exact: true }).fill('100');
    const recovered = response(page, project.id, 'search');
    await panel.getByRole('button', { name: '读取仓库上下文', exact: true }).click();
    expect((await result(await recovered)).index.generation).toBe(current.generation);
    await expect(queriedResult.getByRole('alert')).toHaveCount(0);

    await context.setOffline(true);
    await panel.getByRole('button', { name: '同步代码索引', exact: true }).click();
    await expect(panel.getByRole('alert').filter({ hasText: '维护结果未知' })).toBeVisible();
    await context.setOffline(false);
    await panel.getByLabel('仓库上下文项目', { exact: true }).selectOption('');
    await panel.getByLabel('仓库上下文项目', { exact: true }).selectOption(project.id);
    await expect(panel.getByRole('alert').filter({ hasText: '维护结果未知' })).toBeVisible();
    await page.reload();
    panel = page.getByRole('region', { name: '仓库上下文', exact: true });
    await panel.getByLabel('仓库上下文项目', { exact: true }).selectOption(project.id);
    await expect(panel.getByRole('alert').filter({ hasText: '维护结果未知' })).toBeVisible();
    for (const button of await panel.getByRole('group', { name: '明确维护当前项目索引', exact: true }).getByRole('button').all()) await expect(button).toBeDisabled();
    await fillQuery(panel, 'search');
    const rereadOnly = response(page, project.id, 'search');
    await panel.getByRole('button', { name: '读取仓库上下文', exact: true }).click();
    expect((await result(await rereadOnly)).index.generation).toBe(current.generation);
    expect(mutations).toEqual(['POST', 'POST']);
    expect(server.query('SELECT generation FROM repository_context_indexes WHERE project_id=? AND backend=?', project.id, 'codegraph')).toEqual([{ generation: current.generation }]);
    expect(server.query('SELECT COUNT(*) AS count FROM local_agent_sessions')).toEqual([{ count: 0 }]);
  } finally {
    await context.setOffline(false);
    await server.close({ preserveData: true });
    await owned.unchanged();
  }
});
