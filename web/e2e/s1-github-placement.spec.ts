import { test, expect } from '@playwright/test';
import { execFileSync } from 'node:child_process';
import { existsSync } from 'node:fs';
import { startS1Server } from '../../tests/s1/server.mjs';
import { humanAPI } from '../../tests/s1/api.mjs';
import type { components } from '../src/api/schema';

type Repository = components['schemas']['GitHubRepositoryV1'];
// Explicit opt-in public reference; never bind a private account or run a model/media job.
test('actual public repository metadata precedes human placement, clone and cold persistence', async ({ page }, testInfo) => {
  test.skip(process.env.ASTROCYTE_TEST_GITHUB_REPOSITORY !== '1' || testInfo.project.name !== 'chromium-1920', 'Explicit live public GitHub opt-in, once only');
  test.setTimeout(180_000);
  const server = await startS1Server({ browser: true });
  let primaryError: unknown;
  try {
    let api = await humanAPI(server.apiURL);
    await page.goto(`${server.webURL}/workspace`);
    const metadata = page.locator('section').filter({ has: page.getByRole('heading', { name: 'GitHub 仓库信息', exact: true }) }).first();
    await metadata.getByLabel('公开 GitHub 账号或仓库', { exact: true }).fill('octocat/Hello-World');
    const syncedResponse = page.waitForResponse(response => response.url().endsWith('/api/v1/github-repositories/sync') && response.request().method() === 'POST');
    await metadata.getByRole('button', { name: '同步仓库信息', exact: true }).click();
    const response = await syncedResponse;
    expect(response.status(), await response.text()).toBe(200);
    await expect(metadata.getByRole('heading', { name: 'octocat/Hello-World', exact: true })).toBeVisible();
    let records = await api.get('/github-repositories') as components['schemas']['GitHubRepositoryListV1'];
    expect(records.items).toHaveLength(1);
    const original = records.items[0];
    expect(original.clone_status).toBe('not_placed');
    expect(original.root).toBe(''); expect(original.head).toBe(''); expect(original.project_id).toBe('');
    expect((await api.get('/local-projects')).items).toHaveLength(0);
    if (original.metadata.unknown_fields?.includes('stars')) await expect(metadata).toContainText('★ 未知');
    await metadata.getByRole('button', { name: '去蜂群空间纳入代码', exact: true }).click();
    await page.locator('summary').filter({ hasText: /^创建开发空间$/ }).click();
    await page.getByLabel('开发空间名称', { exact: true }).fill('GitHub 实际公开仓库验收空间');
    await page.getByRole('button', { name: '创建开发空间', exact: true }).click();
    await expect(page.getByRole('heading', { name: 'GitHub 实际公开仓库验收空间', exact: true })).toBeVisible();
    const placement = page.locator('section').filter({ has: page.getByRole('heading', { name: 'GitHub 仓库纳入开发空间', exact: true }) }).first();
    await placement.getByRole('button', { name: '选择空间并纳入', exact: true }).click();
    const button = placement.getByRole('button', { name: '纳入所选空间并克隆代码', exact: true });
    await expect(button).toBeDisabled();
    const space = (await api.get('/project-spaces')).items[0];
    await placement.getByLabel('为 octocat/Hello-World 选择开发空间', { exact: true }).selectOption(space.id);
    const clonedResponse = page.waitForResponse(response => response.url().endsWith(`/api/v1/github-repositories/${original.id}/placement`) && response.request().method() === 'POST');
    await button.click();
    const cloneResponse = await clonedResponse;
    expect(cloneResponse.status(), await cloneResponse.text()).toBe(200);
    await expect(placement.getByText('代码已克隆', { exact: true })).toBeVisible();
    records = await api.get('/github-repositories') as components['schemas']['GitHubRepositoryListV1'];
    const ready: Repository = records.items[0];
    expect(ready.clone_status).toBe('ready'); expect(ready.space_id).toBe(space.id); expect(ready.clone_attempts).toBe(1);
    expect(existsSync(ready.root)).toBe(true);
    expect(execFileSync('git', ['-C', ready.root, 'rev-parse', 'HEAD'], { encoding: 'utf8' }).trim()).toBe(ready.head);
    const local = (await api.get('/local-projects')).items.find((project: { id: string }) => project.id === ready.project_id);
    expect(local.settings).toMatchObject({ allow_directory: false, expand_references: false, allow_agent_control: false, external_model_cli: '', allowed_actions: [] });
    await placement.getByRole('button', { name: '选择空间并纳入', exact: true }).click();
    await expect(placement.getByRole('button', { name: '已纳入并克隆', exact: true })).toBeDisabled();
    const duplicate = await api.write(`/github-repositories/${ready.id}/placement`, { expected_version: ready.revision, space_id: space.id });
    expect(duplicate.repository).toEqual(ready);
    await server.restart(); api = await humanAPI(server.apiURL);
    expect((await api.get('/github-repositories')).items[0]).toEqual(ready);
    expect((await api.get('/missions')).items).toHaveLength(0);
    expect((await api.get('/jobs')).items).toHaveLength(0);
    expect((await api.get(`/local-projects/${ready.project_id}/sessions`)).items).toHaveLength(0);
    await page.reload();
    await expect(placement.getByText('代码已克隆', { exact: true })).toBeVisible();
    for (const width of [1920, 1280, 390]) { await page.setViewportSize({ width, height: width === 390 ? 844 : 900 }); expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true); await page.screenshot({ path: testInfo.outputPath(`github-placed-${width}.png`), fullPage: true, animations: 'disabled' }); }
    await testInfo.attach('actual-public-repository', { body: JSON.stringify(ready, null, 2), contentType: 'application/json' });
    console.log('Actual GitHub repository data retained:', server.temporary, 'HEAD', ready.head);
  } catch (error) { primaryError = error; throw error; }
  finally { await server.close({ preserveData: true, primaryError }); }
});
