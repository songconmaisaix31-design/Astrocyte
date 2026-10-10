import { test, expect } from '@playwright/test';

// Transport contract check only: no personal browser, provider or content extraction.
test('browser folder selection hides technical identity and preserves explicit failed command', async ({ page }) => {
  const discoveries: { body: Record<string, unknown>; key: string }[] = [];
  const bindings: Record<string, unknown>[] = [];
  const folder = { platform: 'douyin', access_mode: 'browser_selected', owner_id: 'contract-local-owner', external_id: 'contract-local-folder', title: 'contract_local 选定收藏夹', locator: 'https://www.douyin.com/user/contract-local-owner?showTab=favorite&collection_id=contract-local-folder', item_count: 3 };
  await page.route('**/api/v1/source-collections/discover', route => {
    discoveries.push({ body: route.request().postDataJSON(), key: route.request().headers()['idempotency-key'] });
    if (discoveries.length === 1) return route.fulfill({ status: 503, json: { schema_version: 1, error: { code: 'provider_unavailable', message: 'contract_local 浏览器未连接', required_action: '正常连接扩展后再明确读取', retryable: true, request_id: 'picker-local' } } });
    return route.fulfill({ json: { schema_version: 1, items: [folder] } });
  });
  await page.route('**/api/v1/tracking-sources', route => {
    if (route.request().method() === 'GET') return route.fulfill({ json: { schema_version: 1, items: [], next_cursor: null } });
    bindings.push(route.request().postDataJSON());
    return route.fulfill({ status: 503, json: { schema_version: 1, error: { code: 'provider_unavailable', message: 'contract_local 保存失败', required_action: '核对来源后重试', retryable: true, request_id: 'bind-local' } } });
  });
  await page.goto('/attention');
  expect(discoveries).toEqual([]); expect(bindings).toEqual([]);
  await page.getByRole('button', { name: '查看来源更新', exact: true }).click();
  const accounts = page.locator('section').filter({ has: page.getByRole('heading', { name: '账号与更新清单', exact: true }) }).first();
  await accounts.locator('summary').filter({ hasText: /^绑定公开创作者 \/ 收藏夹$/ }).click();
  await accounts.getByLabel('来源平台', { exact: true }).selectOption('douyin');
  await accounts.getByLabel('追踪内容', { exact: true }).selectOption('favorites');
  await accounts.getByLabel('收藏夹读取范围', { exact: true }).selectOption('browser_selected');
  await expect(accounts.getByLabel('公开收藏夹 ID', { exact: true })).not.toBeVisible();
  await expect(accounts.getByLabel('收藏夹所属账号 ID', { exact: true })).not.toBeVisible();
  const discover = accounts.getByRole('button', { name: '读取已连接浏览器的选定收藏夹', exact: true });
  await discover.click();
  await expect(accounts.getByRole('alert')).toContainText('正常连接扩展后再明确读取');
  await discover.click();
  await expect(accounts.getByRole('button', { name: /contract_local 选定收藏夹/ })).toBeVisible();
  expect(discoveries).toHaveLength(2); expect(discoveries[1]).toEqual(discoveries[0]);
  expect(discoveries[0].body).toMatchObject({ platform: 'douyin', owner_id: 'self', access_mode: 'browser_selected' });
  await accounts.getByRole('button', { name: /contract_local 选定收藏夹/ }).click();
  await accounts.getByRole('button', { name: '绑定此选定浏览器收藏夹', exact: true }).click();
  await expect(accounts.getByRole('alert')).toContainText('核对来源后重试');
  expect(bindings).toHaveLength(1); expect(bindings[0]).toMatchObject({ owner_id: folder.owner_id, external_id: folder.external_id, access_mode: 'browser_selected', locator: folder.locator });
  expect(bindings[0].owner_id).not.toBe('self');
  await page.setViewportSize({ width: 390, height: 844 });
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
});
