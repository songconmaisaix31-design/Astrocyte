import { test, expect, type Page } from '@playwright/test';
import { realService, repositoryRoot, writeWorks } from './realService';
import type { components } from '../src/api/schema';
type S = components['schemas'];

async function actualWork() {
  const { server, api } = await realService();
  await expect.poll(async () => {
    const discovery = await api.get('/local-projects/registered') as S['RegisteredProjectDiscoveryResultV1'];
    return discovery.snapshot.projects.some(project => project.root.replace(/[\\/]$/, '').toLowerCase() === repositoryRoot.replace(/[\\/]$/, '').toLowerCase());
  }, { timeout: 120_000 }).toBe(true);
  const result = await writeWorks(api, '/works/refresh', { roots: [repositoryRoot] }) as S['WorksSnapshotResultV1'];
  expect(result.snapshot.works).toHaveLength(1);
  const work = result.snapshot.works[0];
  expect(work.files.verified).toBe(true);
  expect(work.files.file_count).toBeGreaterThan(0);
  return { server, api, work, snapshot: result.snapshot };
}
async function openAlbum(page: Page, webURL: string) {
  await page.goto(`${webURL}/workspace`);
  return page.getByRole('region', { name: '作品寻回册', exact: true });
}

test('实际网络断开保留未确认草稿，禁止重发并允许恢复读取', async ({ page, context }) => {
  test.setTimeout(180_000);
  const { server, api, work } = await actualWork();
  try {
    const album = await openAlbum(page, server.webURL);
    await album.getByRole('button', { name: `查看作品 ${work.name}`, exact: true }).click();
    const dialog = page.getByRole('dialog');
    const draft = '核查网络断开时保留人工交接草稿';
    await dialog.getByLabel('编辑交接手册').fill(draft);
    const writes: string[] = [];
    page.on('request', request => { if (request.method() === 'PUT' && request.url().endsWith('/handoff')) writes.push(request.url()); });
    await context.setOffline(true);
    await dialog.getByRole('button', { name: '保存交接手册', exact: true }).click();
    await expect(dialog.getByRole('alert')).toContainText('请求结果未知');
    await expect(dialog.getByLabel('编辑交接手册')).toHaveValue(draft);
    await expect(dialog.getByRole('button', { name: '保存交接手册', exact: true })).toBeDisabled();
    await context.setOffline(false);
    await dialog.getByRole('button', { name: '重新读取保存结果', exact: true }).click();
    await expect(dialog.getByRole('status')).toContainText('已重新读取保存结果');
    await expect(dialog.getByLabel('编辑交接手册')).toHaveValue(draft);
    expect(writes).toHaveLength(0);
    const current = await api.get('/works') as S['WorksSnapshotResultV1'];
    expect(current.snapshot.works[0].handoff).toEqual(work.handoff);
  } finally { await context.setOffline(false); await server.close({ preserveData: true }); }
});

test('实际文件扫描返回平台统计与核实卡片', async ({ page }) => {
  test.setTimeout(180_000);
  const { server, work, snapshot } = await actualWork();
  try {
    const album = await openAlbum(page, server.webURL);
    await expect(album.getByLabel('平台统计')).toBeVisible();
    await expect(album.getByRole('button', { name: `查看作品 ${work.name}`, exact: true })).toBeVisible();
    await expect(album.getByText('从文件核实').first()).toBeVisible();
    expect(snapshot.platforms.reduce((count, stat) => count + stat.works, 0)).toBe(1);
  } finally { await server.close({ preserveData: true }); }
});

test('实际人工分组与态度保存后控制分组和筛选', async ({ page }) => {
  test.setTimeout(180_000);
  const { server, work } = await actualWork();
  try {
    const album = await openAlbum(page, server.webURL);
    const card = album.getByRole('button', { name: `查看作品 ${work.name}`, exact: true });
    await card.click();
    const dialog = page.getByRole('dialog');
    await dialog.getByLabel('编辑分组').fill('标准整改');
    await dialog.getByRole('button', { name: '保存分组', exact: true }).click();
    await expect(dialog.getByRole('status')).toContainText('人工记录已保存');
    await dialog.getByLabel('记录态度').selectOption('park');
    await dialog.getByRole('button', { name: '保存态度', exact: true }).click();
    await expect(dialog.getByRole('status')).toContainText('人工记录已保存');
    await page.keyboard.press('Escape');
    await album.getByRole('button', { name: '按分组', exact: true }).click();
    await expect(album.getByRole('region', { name: '标准整改', exact: true }).getByRole('heading', { name: '标准整改 1', exact: true })).toBeVisible();
    await album.getByLabel('态度', { exact: true }).selectOption('park');
    await expect(card).toBeVisible();
    await album.getByLabel('态度', { exact: true }).selectOption('continue');
    await expect(card).toBeHidden();
  } finally { await server.close({ preserveData: true }); }
});

test('实际作品详情保留文件观察和交接字段', async ({ page }) => {
  test.setTimeout(180_000);
  const { server, work } = await actualWork();
  try {
    const album = await openAlbum(page, server.webURL);
    await album.getByRole('button', { name: `查看作品 ${work.name}`, exact: true }).click();
    const dialog = page.getByRole('dialog');
    await expect(dialog).toContainText('文件核实');
    await expect(dialog).toContainText('交接手册');
    await expect(dialog).toContainText(work.files.source_path);
    await expect(dialog).toContainText(`${work.files.file_count} 个文件`);
    await page.keyboard.press('Escape');
    await expect(dialog).not.toBeVisible();
  } finally { await server.close({ preserveData: true }); }
});

test('实际作品卡片通过 Enter 打开详情', async ({ page }) => {
  test.setTimeout(180_000);
  const { server, work } = await actualWork();
  try {
    const album = await openAlbum(page, server.webURL);
    await album.getByRole('button', { name: `查看作品 ${work.name}`, exact: true }).focus();
    await page.keyboard.press('Enter');
    await expect(page.getByRole('dialog')).toBeVisible();
  } finally { await server.close({ preserveData: true }); }
});

for (const size of [{ width: 1920, height: 1080 }, { width: 1280, height: 720 }, { width: 390, height: 844 }]) {
  test(`实际作品在 ${size.width}x${size.height} 没有水平溢出或页面异常`, async ({ page }) => {
    test.setTimeout(180_000);
    const { server, work } = await actualWork();
    try {
      const errors: string[] = [];
      page.on('pageerror', error => errors.push(error.message));
      await page.setViewportSize(size);
      const album = await openAlbum(page, server.webURL);
      await expect(album.getByRole('button', { name: `查看作品 ${work.name}`, exact: true })).toBeVisible();
      expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
      expect(errors).toEqual([]);
    } finally { await server.close({ preserveData: true }); }
  });
}
