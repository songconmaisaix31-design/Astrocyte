import { test, expect } from '@playwright/test';

for (const narrow of [false, true]) {
  test(`redesigned task flows retain live empty states and local assets ${narrow ? 'narrow' : 'desktop'}`, async ({ page }, testInfo) => {
    if (narrow) await page.setViewportSize({ width: 390, height: 844 });
    const writes: string[] = [];
    page.on('request', request => { if (request.method() !== 'GET' && request.url().includes('/api/v1/')) writes.push(request.url()); });
    for (const [path, title] of [['attention', '资料沉淀'], ['workspace', '共同工作区'], ['swarm', '蜂群空间']]) {
      await page.goto(`/${path}`);
      await expect(page.getByRole('heading', { name: title, exact: true })).toBeVisible();
      await expect(page.getByRole('alert').filter({ hasText: '示例数据模式' })).toHaveCount(0);
      await expect(page.getByRole('status', { name: path === 'attention' ? '暂无素材' : path === 'workspace' ? '暂无本地项目' : '暂无开发空间', exact: true })).toBeVisible();
      await page.waitForLoadState('networkidle');
      await page.screenshot({ path: testInfo.outputPath(`${path}-initial-${page.viewportSize()?.width}.png`), fullPage: true, animations: 'disabled' });
      if (path === 'attention') {
        await expect(page.getByRole('status', { name: '暂无素材', exact: true })).toBeVisible();
        await expect(page.getByRole('heading', { name: '账号与更新清单', exact: true })).not.toBeVisible();
        await page.getByRole('button', { name: /查看来源更新/ }).click();
        await expect(page.getByRole('heading', { name: '账号与更新清单', exact: true })).toBeVisible();
        await expect(page.getByRole('status', { name: '尚未绑定来源', exact: true })).toBeVisible();
        await page.locator('summary').filter({ hasText: /^查看来源更新$/ }).click();
        await page.getByRole('button', { name: /查看处理队列/ }).click();
        await expect(page.getByRole('heading', { name: '导入与沉淀队列', exact: true })).toBeVisible();
        await page.locator('summary').filter({ hasText: /^查看处理队列$/ }).click();
      }
      if (path === 'workspace') {
        await expect(page.getByRole('status', { name: '暂无本地项目', exact: true })).toBeVisible();
        await page.getByRole('button', { name: '检查本机 Agent 清单', exact: true }).click();
        const dialog = page.getByRole('dialog', { name: '检查本机 Agent', exact: true });
        await expect(dialog.getByRole('heading', { name: '本机 Agent 清单', exact: true })).toBeVisible();
        await expect(dialog.locator('li')).toHaveCount(10);
        await dialog.getByRole('button', { name: '关闭', exact: true }).click();
      }
      if (path === 'swarm') await expect(page.getByRole('status', { name: '暂无开发空间', exact: true })).toBeVisible();
      expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
      const images = await page.locator('img[src^="/brands/"]').evaluateAll(elements => elements.map(element => ({ src: (element as HTMLImageElement).src, loaded: (element as HTMLImageElement).complete && (element as HTMLImageElement).naturalWidth > 0 })));
      expect(images.every(image => image.loaded)).toBe(true);
      await page.evaluate(() => window.scrollTo({ top: 0, behavior: 'instant' }));
      await page.screenshot({ path: testInfo.outputPath(`${path}-${page.viewportSize()?.width}.png`), fullPage: true, animations: 'disabled' });
    }
    expect(writes).toEqual([]);
  });
}
