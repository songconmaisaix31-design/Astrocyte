import { test, expect, type Page } from '@playwright/test';
import { humanAPI } from '../../tests/s1/api.mjs';
import { realService, registerCurrentProject, realProgress, currentStatus, permitProgressForm } from './realService';
import type { components } from '../src/api/schema';

async function openProgress(page: Page, webURL: string) {
  await page.goto(`${webURL}/workspace`);
  await page.locator('summary').filter({ hasText: /^项目接入、权限与原生操作/ }).click();
  await page.getByRole('button', { name: /查看项目与权限/ }).click();
  const progress = page.locator('section[aria-label="项目进度推断"]');
  await expect(progress).toBeVisible();
  return progress;
}

test('实际人工进度表单保存本轨 STATUS.md，权限恢复与重启保留且不授权模型', async ({ page, context }) => {
  test.setTimeout(120_000);
  const { server, api } = await realService();
  try {
    const project = await registerCurrentProject(api);
    const summary = (await currentStatus()).split('\n').slice(0, 2).join('\n').trim();
    const progress = await openProgress(page, server.webURL);
    const manual = progress.getByRole('form', { name: '人工项目进度', exact: true });
    await manual.getByLabel('人工项目状态', { exact: true }).fill('人工核查');
    await manual.getByLabel('人工进度说明', { exact: true }).fill(summary);
    await expect(manual.getByLabel('人工完成度（0–100，未知留空）', { exact: true })).toHaveValue('');
    const writes: string[] = [];
    page.on('request', request => { if (new URL(request.url()).pathname.startsWith('/api/v1/') && request.method() !== 'GET') writes.push(new URL(request.url()).pathname); });
    const saving = page.waitForResponse(response => response.request().method() === 'PUT' && response.url().endsWith(`/local-projects/${project.id}/progress`));
    await manual.getByRole('button', { name: '保存人工进度', exact: true }).click();
    const savedHTTP = await saving;
    expect(savedHTTP.status()).toBe(200);
    expect(savedHTTP.request().postDataJSON()).toMatchObject({ status: '人工核查', summary, percent: null });
    expect(savedHTTP.request().postDataJSON().request_id).toBe(savedHTTP.request().headers()['idempotency-key']);
    const saved = await realProgress(api, project.id);
    expect(saved.source).toBe('human');
    expect(saved.summary).toBe(summary);
    expect(saved.percent == null).toBe(true);
    await expect(progress.getByText(/状态 · 人工核查/)).toBeVisible();
    await expect(progress.getByLabel('项目进度摘要', { exact: true })).toBeVisible();
    await expect(progress.getByLabel('项目进度摘要', { exact: true })).toHaveText(summary);
    await expect(progress.getByRole('button', { name: '需先许可项目模型客户端', exact: true })).toBeDisabled();
    await expect(progress.getByLabel('推断文件（相对路径，每行一个）', { exact: true })).toHaveValue('STATUS.md');
    await expect(progress.getByText(/完成度 ·/)).toHaveCount(0);
    await context.setExtraHTTPHeaders({ Authorization: 'Bearer invalid' });
    const denied = page.waitForResponse(response => response.url().endsWith(`/local-projects/${project.id}/progress`));
    await progress.getByRole('button', { name: '重载进度', exact: true }).click();
    expect((await denied).status()).toBe(403);
    await expect(progress.getByRole('alert')).toContainText('进度读取失败');
    await expect(manual.getByRole('button', { name: '保存人工进度', exact: true })).toBeDisabled();
    await context.setExtraHTTPHeaders({});
    await progress.getByRole('button', { name: '重载进度', exact: true }).click();
    await expect(progress.getByRole('alert')).toHaveCount(0);
    await expect(manual.getByRole('button', { name: '保存人工进度', exact: true })).toBeEnabled();
    await server.restart();
    await page.reload();
    const reopened = await openProgress(page, server.webURL);
    await expect(reopened.getByLabel('项目进度摘要', { exact: true })).toBeVisible();
    await expect(reopened.getByLabel('项目进度摘要', { exact: true })).toHaveText(summary);
    expect((await realProgress(await humanAPI(server.apiURL), project.id)).summary).toBe(summary);
    expect(writes).toEqual([`/api/v1/local-projects/${project.id}/progress`]);
    const projectList = await (await humanAPI(server.apiURL)).get('/local-projects') as components['schemas']['LocalProjectListV1'];
    const reopenedProject = projectList.items.find(item => item.id === project.id);
    if (!reopenedProject) throw new Error('真实重启后的项目列表缺少原已登记项目');
    expect(reopenedProject.settings).toEqual(project.settings);
    expect(server.query('SELECT COUNT(*) AS count FROM local_agent_sessions')).toEqual([{ count: 0 }]);
  } finally { await context.setExtraHTTPHeaders({}); await server.close({ preserveData: true }); }
});

test('人工进度实际离线后保留输入和未知锁，读取或重开不重发', async ({ page, context }) => {
  test.setTimeout(120_000);
  const { server, api } = await realService();
  try {
    const project = await registerCurrentProject(api);
    let progress = await openProgress(page, server.webURL);
    let manual = progress.getByRole('form', { name: '人工项目进度', exact: true });
    await manual.getByLabel('人工项目状态', { exact: true }).fill('人工核查待完成');
    await manual.getByLabel('人工进度说明', { exact: true }).fill('当前获准项目尚未保存人工进度。');
    const writes: string[] = [];
    page.on('request', request => { if (request.method() === 'PUT' && request.url().endsWith(`/local-projects/${project.id}/progress`)) writes.push(request.url()); });
    await context.setOffline(true);
    await manual.getByRole('button', { name: '保存人工进度', exact: true }).click();
    await expect(manual.getByRole('alert').filter({ hasText: '关闭或刷新后保持锁定' })).toBeVisible();
    await expect(manual.getByLabel('人工项目状态', { exact: true })).toHaveValue('人工核查待完成');
    await expect(manual.getByRole('button', { name: '保存人工进度', exact: true })).toBeDisabled();
    await context.setOffline(false);
    await progress.getByRole('button', { name: '重载进度', exact: true }).click();
    await expect(manual.getByRole('button', { name: '保存人工进度', exact: true })).toBeDisabled();
    await page.reload();
    progress = await openProgress(page, server.webURL);
    manual = progress.getByRole('form', { name: '人工项目进度', exact: true });
    await expect(manual.getByRole('button', { name: '保存人工进度', exact: true })).toBeDisabled();
    expect(writes).toEqual([]);
    expect(await realProgress(api, project.id)).toMatchObject({ status: 'unknown', source: null, revision: 0 });
    expect(server.query('SELECT COUNT(*) AS count FROM local_project_progress WHERE project_id=?', project.id)).toEqual([{ count: 0 }]);
    expect(server.query('SELECT COUNT(*) AS count FROM local_agent_sessions')).toEqual([{ count: 0 }]);
  } finally { await context.setOffline(false); await server.close({ preserveData: true }); }
});

test('纯输入解析完整保留多行路径和越界路径，实际 API 拒绝未许可推断', async ({ page }) => {
  test.setTimeout(120_000);
  const { parseProgressFiles } = await import('../src/pages/workspace/progressPresentation');
  expect(parseProgressFiles('docs/plan.md\nnotes.md')).toEqual(['docs/plan.md', 'notes.md']);
  expect(parseProgressFiles('../secret.md')).toEqual(['../secret.md']);
  const { server, api } = await realService();
  try {
    const project = await registerCurrentProject(api);
    await permitProgressForm(api, project);
    const progress = await openProgress(page, server.webURL);
    const field = progress.getByLabel('推断文件（相对路径，每行一个）', { exact: true });
    await expect(field).toHaveValue('STATUS.md');
    for (const files of [['docs/plan.md', 'notes.md'], ['../secret.md']]) {
      await field.fill(files.join('\n'));
      const response = page.waitForResponse(value => value.url().endsWith('/progress/infer') && value.request().method() === 'POST');
      await progress.getByRole('button', { name: '发起进度推断', exact: true }).click();
      const rejected = await response;
      expect(rejected.request().postDataJSON().files).toEqual(files);
      expect(rejected.status()).toBe(403);
      await expect(progress.getByRole('alert')).toBeVisible();
    }
    expect((await realProgress(api, project.id)).source).toBeNull();
  } finally { await server.close({ preserveData: true }); }
});

test('超限文件完整保留，实际 API 校验拒绝且不开始模型推断', async ({ page }) => {
  test.setTimeout(120_000);
  const { parseProgressFiles, PROGRESS_FILES_LIMIT } = await import('../src/pages/workspace/progressPresentation');
  const files = Array.from({ length: 9 }, (_, i) => `file${i}.md`);
  expect(parseProgressFiles(files.join('\n'))).toEqual(files);
  expect(files.length).toBeGreaterThan(PROGRESS_FILES_LIMIT);
  const { server, api } = await realService();
  try {
    const project = await registerCurrentProject(api);
    await permitProgressForm(api, project);
    const progress = await openProgress(page, server.webURL);
    await progress.getByLabel('推断文件（相对路径，每行一个）', { exact: true }).fill(files.join('\n'));
    await expect(progress.getByText(/超过 8 个/)).toBeVisible();
    await expect(progress.getByRole('button', { name: '发起进度推断', exact: true })).toBeDisabled();
    const result = await api.request(`/local-projects/${project.id}/progress/infer`, {
      method: 'POST', body: { schema_version: 1, request_id: crypto.randomUUID(), expected_version: 1, files },
    });
    expect(result.status).toBe(400);
    expect((await realProgress(api, project.id)).revision).toBe(0);
  } finally { await server.close({ preserveData: true }); }
});

test('实际未记录进度显示 unknown 和空观察时间', async ({ page }) => {
  test.setTimeout(120_000);
  const { server, api } = await realService();
  try {
    const project = await registerCurrentProject(api);
    expect(await realProgress(api, project.id)).toMatchObject({ status: 'unknown', source: null, observed_at: null, revision: 0, evidence: [] });
    const progress = await openProgress(page, server.webURL);
    await expect(progress.getByText(/状态 · 未知/)).toBeVisible();
    await expect(progress.getByText(/尚无进度记录/)).toBeVisible();
    await expect(progress.getByText(/完成度 ·/)).toHaveCount(0);
  } finally { await server.close({ preserveData: true }); }
});
