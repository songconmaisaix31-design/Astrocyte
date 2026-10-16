import { test, expect } from '@playwright/test';
import { execFileSync } from 'node:child_process';
import { existsSync } from 'node:fs';
import { startS1Server } from '../../tests/s1/server.mjs';
import { humanAPI } from '../../tests/s1/api.mjs';
import type { components } from '../src/api/schema';
import { repositoryRoot, realService } from './realService';

type Repository = components['schemas']['GitHubRepositoryV1'];
// 仅按明确许可验证公开仓库，保存实际克隆和持久化观察。
test('actual public repository metadata precedes human placement, clone and cold persistence', async ({ page }, testInfo) => {
  test.skip(process.env.ASTROCYTE_TEST_GITHUB_REPOSITORY !== '1' || testInfo.project.name !== 'chromium-1920', 'Explicit live public GitHub opt-in, once only');
  test.setTimeout(180_000);
  const reusedPath = process.env.ASTROCYTE_GITHUB_REUSE_OWNED_TEMP;
  const approvedRoot = process.env.ASTROCYTE_GITHUB_REUSE_APPROVED_ROOT;
  if (!!reusedPath !== !!approvedRoot) throw new Error('Continuation requires both exact original helper path and its approved owned root');
  const server = await startS1Server({ browser: true, ...(reusedPath && approvedRoot ? { reuseOwnedTemporary: { path: reusedPath, ownedRoot: approvedRoot } } : {}) });
  let primaryError: unknown;
  try {
    let api = await humanAPI(server.apiURL);
    await page.goto(`${server.webURL}/workspace`);
    const metadata = page.locator('section').filter({ has: page.getByRole('heading', { name: 'GitHub 仓库信息', exact: true }) }).last();
    let records = await api.get('/github-repositories');
    if (!reusedPath) {
      expect(records.items).toHaveLength(0);
      await metadata.getByLabel('公开 GitHub 账号或仓库', { exact: true }).fill('songconmaisaix31-design/Astrocyte');
      const syncing = page.waitForResponse(response => response.url().endsWith('/api/v1/github-repositories/sync') && response.request().method() === 'POST');
      await metadata.getByRole('button', { name: '同步仓库信息', exact: true }).click();
      const response = await syncing; expect(response.status(), await response.text()).toBe(200);
      records = await api.get('/github-repositories');
    }
    expect(records.items).toHaveLength(1);
    const original = records.items[0];
    expect(original.metadata.full_name).toBe('songconmaisaix31-design/Astrocyte');
    if (!reusedPath) {
      expect(original.clone_status).toBe('not_placed');
      expect(original.root).toBe(''); expect(original.head).toBe(''); expect(original.project_id).toBe('');
      expect((await api.get('/local-projects')).items).toHaveLength(0);
    } else expect(original.clone_status).toBe('ready');
    await expect(metadata.getByRole('heading', { name: 'songconmaisaix31-design/Astrocyte', exact: true })).toBeVisible();
    if (original.metadata.unknown_fields?.includes('stars')) await expect(metadata).toContainText('★ 未知');
    await metadata.getByRole('button', { name: '去蜂群空间纳入代码', exact: true }).click();
    if (!reusedPath) {
      await page.locator('summary').filter({ hasText: /^创建开发空间$/ }).click();
      await page.getByLabel('开发空间名称', { exact: true }).fill('GitHub 实际公开仓库验收空间');
      await page.getByRole('button', { name: '创建开发空间', exact: true }).click();
    }
    await expect(page.getByRole('heading', { name: 'GitHub 实际公开仓库验收空间', exact: true })).toBeVisible();
    const space = (await api.get('/project-spaces')).items.find(item => reusedPath ? item.id === original.space_id : item.title === 'GitHub 实际公开仓库验收空间');
    if (!space) throw new Error('The original human-created development space was not found');
    const placement = page.locator('section').filter({ has: page.getByRole('heading', { name: 'GitHub 仓库纳入开发空间', exact: true }) }).last();
    await placement.getByRole('button', { name: '选择空间并纳入', exact: true }).click();
    if (!reusedPath) {
      const button = placement.getByRole('button', { name: '纳入所选空间并克隆代码', exact: true });
      await expect(button).toBeDisabled();
      await placement.getByLabel('为 songconmaisaix31-design/Astrocyte 选择开发空间', { exact: true }).selectOption(space.id);
      const cloning = page.waitForResponse(response => response.url().endsWith(`/api/v1/github-repositories/${original.id}/placement`) && response.request().method() === 'POST');
      await button.click();
      const response = await cloning; expect(response.status(), await response.text()).toBe(200);
    }
    await expect(placement.getByText('代码已克隆', { exact: true })).toBeVisible();
    const ready = (await api.get('/github-repositories')).items[0];
    expect(ready.clone_status).toBe('ready'); expect(ready.space_id).toBe(space.id); expect(ready.clone_attempts).toBe(1);
    expect(existsSync(ready.root)).toBe(true);
    expect(execFileSync('git', ['-C', ready.root, 'rev-parse', 'HEAD'], { encoding: 'utf8' }).trim()).toBe(ready.head);
    const local = (await api.get('/local-projects')).items.find(project => project.id === ready.project_id);
    if (!local) throw new Error('Ready repository did not create its actual local project');
    expect(local.settings).toMatchObject({ allow_directory: false, expand_references: false, allow_agent_control: false, external_model_cli: '', allowed_actions: [] });
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
    await placement.getByRole('button', { name: '选择空间并纳入', exact: true }).click();
    await expect(placement.getByRole('button', { name: '已纳入并克隆', exact: true })).toBeDisabled();
    for (const width of [1920, 1280, 390]) { await page.setViewportSize({ width, height: width === 390 ? 844 : 900 }); expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true); await page.evaluate(() => window.scrollTo({ top: 0, behavior: 'instant' })); }
    await testInfo.attach('actual-public-repository', { body: JSON.stringify({ continuation: !!reusedPath, repository: ready }, null, 2), contentType: 'application/json' });
    console.log('Actual GitHub repository data retained:', server.temporary, 'HEAD', ready.head, reusedPath ? 'same-store post-clone continuation' : 'complete human placement');
  } catch (error) { primaryError = error; throw error; }
  finally { await server.close({ preserveData: true, primaryError }); }
});

test('真实克隆进程中断后保留原空间，人工恢复核对实际 HEAD', async ({ page }) => {
  test.setTimeout(240_000);
  expect(execFileSync('git', ['remote', 'get-url', 'origin'], { cwd: repositoryRoot, encoding: 'utf8' }).trim()).toBe('https://github.com/songconmaisaix31-design/Astrocyte');
  const { server, api } = await realService();
  try {
    await api.write('/github-repositories/sync', { input: 'songconmaisaix31-design/Astrocyte' });
    const original = (await api.get('/github-repositories')).items[0];
    expect(original.metadata.full_name).toBe('songconmaisaix31-design/Astrocyte');
    const space = await api.write('/project-spaces', { title: '本轨真实克隆恢复空间' }, { status: 201 });
    const requests: { body: unknown; key: string }[] = [];
    page.on('request', request => {
      if (request.method() === 'POST' && request.url().endsWith('/placement')) requests.push({ body: request.postDataJSON(), key: request.headers()['idempotency-key'] });
    });
    await page.goto(`${server.webURL}/swarm`);
    const placement = page.locator('section').filter({ has: page.getByRole('heading', { name: 'GitHub 仓库纳入开发空间', exact: true }) }).last();
    expect(requests).toEqual([]);
    await placement.getByRole('button', { name: '选择空间并纳入', exact: true }).click();
    await placement.getByLabel('为 songconmaisaix31-design/Astrocyte 选择开发空间', { exact: true }).selectOption(space.space.id);
    await placement.getByRole('button', { name: '纳入所选空间并克隆代码', exact: true }).click();
    await expect.poll(() => server.query('SELECT json_extract(data,\'$.clone_status\') AS status FROM local_github_repositories WHERE id=?', original.id)[0]?.status, { intervals: [20], timeout: 10_000 }).toBe('cloning');
    await server.stopAPI();
    await expect(placement.getByRole('alert')).toBeVisible();
    expect(requests).toHaveLength(1);
    await server.startAPI();
    const reconnected = await humanAPI(server.apiURL);
    const pending = (await reconnected.get('/github-repositories')).items[0];
    expect(pending.space_id).toBe(space.space.id);
    expect(pending.clone_status).toBe('cloning');
    await page.reload();
    await placement.getByRole('button', { name: '选择空间并纳入', exact: true }).click();
    const choice = placement.getByLabel('为 songconmaisaix31-design/Astrocyte 选择开发空间', { exact: true });
    await expect(choice).toHaveValue(space.space.id);
    await expect(choice).toBeDisabled();
    const recovered = page.waitForResponse(value => value.url().endsWith('/placement') && value.request().method() === 'POST');
    await placement.getByRole('button', { name: '核对并恢复此空间克隆', exact: true }).click();
    const recoveryResponse = await recovered;
    expect(recoveryResponse.status()).toBe(200);
    const recoveryBody = await recoveryResponse.json();
    await expect(placement.getByText('代码已克隆', { exact: true })).toBeVisible();
    const ready = (await reconnected.get('/github-repositories')).items[0] as Repository;
    expect(ready.clone_status).toBe('ready');
    expect(ready.space_id).toBe(space.space.id);
    expect(execFileSync('git', ['-C', ready.root, 'rev-parse', 'HEAD'], { encoding: 'utf8' }).trim()).toBe(ready.head);
    expect(requests).toHaveLength(2);
    expect(requests[0].body).toMatchObject({ expected_version: original.revision, space_id: space.space.id });
    expect(requests[1].body).toMatchObject({ expected_version: pending.revision, space_id: space.space.id });
    const recoveryRequest = recoveryResponse.request();
    const replay = await page.request.post(recoveryRequest.url(), { headers: recoveryRequest.headers(), data: recoveryRequest.postDataJSON() });
    expect(replay.status()).toBe(200);
    expect(await replay.json()).toEqual(recoveryBody);
    const replayed = (await reconnected.get('/github-repositories')).items[0] as Repository;
    expect(replayed.clone_attempts).toBe(ready.clone_attempts);
    expect(replayed.root).toBe(ready.root);
    expect(replayed.head).toBe(ready.head);
    await expect(placement.getByRole('button', { name: '已纳入并克隆', exact: true })).toBeDisabled();
  } finally { await server.close({ preserveData: true }); }
});
