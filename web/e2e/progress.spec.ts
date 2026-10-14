import { test, expect, type Page } from '@playwright/test';
import { humanAPI } from '../../tests/s1/api.mjs';
import { realService, registerCurrentProject, realProgress, currentStatus, permitProgressForm } from './realService';

async function openProgress(page: Page, webURL: string) {
  await page.goto(`${webURL}/workspace`);
  await page.locator('summary').filter({ hasText: /^项目接入、权限与原生操作/ }).click();
  await page.getByRole('button', { name: /查看项目与权限/ }).click();
  const progress = page.locator('section[aria-label="项目进度推断"]');
  await expect(progress).toBeVisible();
  return progress;
}

test('实际人工进度保留本轨 STATUS.md，并显示模型推断所需许可', async ({ page }) => {
  test.setTimeout(120_000);
  const { server, api } = await realService();
  try {
    const project = await registerCurrentProject(api);
    const summary = (await currentStatus()).split('\n').slice(0, 2).join('\n').trim();
    await api.write(`/local-projects/${project.id}/progress`, { status: '人工核查', summary, percent: null }, { method: 'PUT' });
    const saved = await realProgress(api, project.id);
    expect(saved.source).toBe('human');
    expect(saved.summary).toBe(summary);
    const progress = await openProgress(page, server.webURL);
    await expect(progress.getByText(/状态 · 人工核查/)).toBeVisible();
    await expect(progress.getByText(summary, { exact: true })).toBeVisible();
    await expect(progress.getByRole('button', { name: '需先许可项目模型客户端', exact: true })).toBeDisabled();
    await expect(progress.getByLabel('推断文件（相对路径，每行一个）', { exact: true })).toHaveValue('STATUS.md');
    await server.restart();
    await page.reload();
    await expect(progress.getByText(summary, { exact: true })).toBeVisible();
    expect((await realProgress(await humanAPI(server.apiURL), project.id)).summary).toBe(summary);
  } finally { await server.close({ preserveData: true }); }
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
