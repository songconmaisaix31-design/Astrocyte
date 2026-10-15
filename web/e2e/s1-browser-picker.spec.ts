import { test, expect } from '@playwright/test';

test('浏览器范围未交接时实际拒绝读取，保留命令身份和隐藏技术输入', async ({ page }) => {
  const discoveries: { body: unknown; key: string }[] = [];
  const bindings: string[] = [];
  page.on('request', request => {
    if (request.url().endsWith('/source-collections/discover') && request.method() === 'POST') discoveries.push({ body: request.postDataJSON(), key: request.headers()['idempotency-key'] });
    if (request.url().endsWith('/tracking-sources') && request.method() === 'POST') bindings.push(request.url());
  });
  await page.goto('/attention');
  expect(discoveries).toEqual([]);
  expect(bindings).toEqual([]);
  await page.getByRole('button', { name: '查看来源更新', exact: true }).click();
  const accounts = page.locator('section').filter({ has: page.getByRole('heading', { name: '账号与更新清单', exact: true }) }).first();
  await accounts.locator('summary').filter({ hasText: /^绑定创作者 \/ 收藏夹$/ }).click();
  await accounts.getByLabel('来源平台', { exact: true }).selectOption('douyin');
  await accounts.getByLabel('追踪内容', { exact: true }).selectOption('favorites');
  await accounts.getByLabel('收藏夹读取范围', { exact: true }).selectOption('browser_selected');
  await expect(accounts.getByLabel('公开收藏夹 ID', { exact: true })).not.toBeVisible();
  await expect(accounts.getByLabel('收藏夹所属账号 ID', { exact: true })).not.toBeVisible();
  const discover = accounts.getByRole('button', { name: '读取已连接浏览器的选定收藏夹', exact: true });
  for (let index = 0; index < 2; index++) {
    const response = page.waitForResponse(value => value.url().endsWith('/source-collections/discover') && value.request().method() === 'POST');
    await discover.click();
    const rejected = await response;
    expect(rejected.status()).toBe(503);
    expect((await rejected.json()).error.code).toBe('provider_unavailable');
    await expect(accounts.getByRole('alert')).toBeVisible();
    await expect(accounts.getByRole('button', { name: '绑定此选定浏览器收藏夹', exact: true })).toBeDisabled();
  }
  expect(discoveries).toHaveLength(2);
  expect(discoveries[1]).toEqual(discoveries[0]);
  expect(discoveries[0].body).toMatchObject({ platform: 'douyin', owner_id: 'self', access_mode: 'browser_selected' });
  expect(bindings).toHaveLength(0);
  await page.setViewportSize({ width: 390, height: 844 });
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
});
