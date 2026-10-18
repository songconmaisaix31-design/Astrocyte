import { test, expect } from '@playwright/test';
import { readFile } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';
import { startS1Server } from '../../tests/s1/server.mjs';
import { humanAPI } from '../../tests/s1/api.mjs';
import { importTrackedDocument } from './realService';
import type { components } from '../src/api/schema';

test.describe('新版页面的真实服务与品牌资源', () => {
  let server: Awaited<ReturnType<typeof startS1Server>>;
  let api: Awaited<ReturnType<typeof humanAPI>>;
  let spaceID: string;
  let refValue: string;
  let sourceText: string;
  test.beforeAll(async () => {
    test.setTimeout(120_000);
    server = await startS1Server({ env: { ASTROCYTE_WEB_DIR: fileURLToPath(new URL('../dist/', import.meta.url)), OPENVIKING_BASE_URL: '', OPENVIKING_API_KEY: '', OPENVIKING_ACCOUNT: '', OPENVIKING_USER: '', OPENVIKING_ROOT_URI: '' } });
    api = await humanAPI(server.apiURL);
    const imported = await importTrackedDocument(api, 'README.md');
    const space = await api.write('/project-spaces', { title: '新版页面公开来源验证' }, { status: 201 });
    spaceID = space.space.id;
    const material = imported.detail.material;
    await api.write(`/project-spaces/${spaceID}/references`, { material_id: material.id, revision: material.current_revision });
    refValue = `${material.id}@${material.current_revision}`;
    sourceText = (await api.get(`/materials/${material.id}/revisions/${material.current_revision}/content`)).text;
  });
  test.afterAll(async () => { if (server) await server.close({ preserveData: true }); });

  test('新版侧栏保留生产页面、真实集合与双视口布局', async ({ page }) => {
    const failures: string[] = [];
    page.on('pageerror', error => failures.push(error.message));
    for (const [path, title] of [['attention', '资料沉淀'], ['attention/sources', '资料沉淀'], ['workspace', '共同工作区'], ['swarm', '蜂群空间'], ['cognition/focus', '认知工作台'], ['cognition/evidence', '认知工作台']]) {
      await page.goto(`${server.apiURL}/${path}`);
      await expect(page.locator('h1')).toHaveText(title);
      await expect(page.getByRole('navigation', { name: '主导航' })).toBeVisible();
      expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
    }
    expect(failures).toEqual([]);
    expect((await api.get('/missions')).items).toEqual([]);
  });

  test('Ctrl K、清除搜索、侧栏和原批准入口使用键盘交互', async ({ page }) => {
    await page.goto(`${server.apiURL}/attention`);
    await expect(page.locator('main li[role="button"]').filter({ hasText: 'README.md' })).toBeVisible();
    await page.keyboard.press('Control+k');
    const search = page.getByRole('textbox', { name: '搜索当前页资料、任务或 Agent' });
    await expect(search).toBeFocused();
    await search.fill('README.md');
    await expect(page.locator('main li[role="button"]').filter({ hasText: 'README.md' })).toBeVisible();
    await page.keyboard.press('Escape');
    await expect(search).toHaveValue('');
    const workspace = page.getByRole('navigation', { name: '主导航' }).getByRole('button', { name: '共同工作区', exact: true });
    await workspace.focus();
    await page.keyboard.press('Enter');
    await expect(page).toHaveURL(/\/workspace$/);
    await page.getByRole('tab', { name: '提案与批准' }).click();
    await expect(page.getByRole('status', { name: '暂无提案', exact: true })).toBeVisible();
    await page.getByRole('navigation', { name: '主导航' }).getByRole('button', { name: '作品寻回册' }).click();
    await expect(page).toHaveURL(/\/workspace\/works/);
    await expect(page.locator('h1')).toBeVisible();
  });

  test('可选仓库未配置时可查看实际固定正文且外发保持禁用', async ({ page }) => {
    const sends: string[] = [];
    page.on('request', request => { if (request.method() !== 'GET' && new URL(request.url()).pathname.includes('/openviking/')) sends.push(request.url()); });
    await page.goto(`${server.apiURL}/attention`);
    const repository = page.getByRole('region', { name: 'OpenViking 资料仓库' });
    await expect(repository.getByText('未配置', { exact: true })).toBeVisible();
    await repository.getByLabel('OpenViking 项目空间', { exact: true }).selectOption(spaceID);
    await repository.getByLabel('OpenViking 固定资料版本', { exact: true }).selectOption(refValue);
    const preview = repository.locator('details').filter({ has: page.locator('summary').filter({ hasText: '核对实际固定版本正文' }) });
    await expect(preview).toBeVisible();
    await preview.locator('summary').click();
    expect(await preview.locator('pre').textContent()).toBe(sourceText);
    await expect(repository.getByRole('checkbox')).toBeDisabled();
    await expect(repository.getByRole('button', { name: '确认并提交到 OpenViking' })).toBeDisabled();
    await expect(repository.getByLabel('OpenViking 语义检索')).toBeDisabled();
    expect(sends).toEqual([]);
    const submissions = await api.request(`/project-spaces/${spaceID}/openviking/submissions`);
    expect(submissions.status).toBe(200);
    expect((submissions.data as components['schemas']['OpenVikingSubmissionListV1']).items).toEqual([]);
  });

  test('官方品牌本地资源真实响应并可由浏览器解码', async ({ page }) => {
    await page.goto(`${server.apiURL}/workspace`);
    const manifest = JSON.parse(await readFile(fileURLToPath(new URL('../public/brands/provenance.json', import.meta.url)), 'utf8')) as { name: string; file?: string; status: string }[];
    const decoded: string[] = [];
    for (const entry of manifest.filter(item => item.status === 'downloaded')) {
      expect(entry.file).toBeTruthy();
      const url = `${server.apiURL}/brands/${entry.file}`;
      const response = await page.request.get(url);
      expect(response.status(), entry.name).toBe(200);
      expect((await response.body()).length, entry.name).toBeGreaterThan(0);
      expect(response.headers()['content-type'], entry.name).toMatch(/^image\//);
      const dimensions = await page.evaluate(async source => { const image = new Image(); image.src = source; await image.decode(); return [image.naturalWidth, image.naturalHeight]; }, url);
      expect(dimensions[0], entry.name).toBeGreaterThan(0);
      expect(dimensions[1], entry.name).toBeGreaterThan(0);
      decoded.push(entry.name);
    }
    expect(decoded).toContain('kimi');
    expect(decoded).toContain('codex');
    await page.getByRole('tab', { name: 'Agent 会话', exact: true }).click();
    await expect(page.getByRole('status', { name: '暂无会话', exact: true })).toBeVisible();
  });
});
