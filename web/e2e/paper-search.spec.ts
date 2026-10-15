import { test, expect, type Page } from '@playwright/test';
import type { components } from '../src/api/schema';
type Search = components['schemas']['PaperSearchResultV1'];
const panel = (page: Page) => page.locator('section[aria-label="论文检索"]');

async function search(page: Page, query = 'Attention Is All You Need') {
  await page.goto('/attention');
  await page.getByRole('button', { name: /检索论文/ }).first().click();
  await panel(page).getByLabel('检索论文', { exact: true }).fill(query);
  const completed = page.waitForResponse(response => response.url().includes('/api/v1/papers/search?'));
  await panel(page).getByRole('button', { name: '检索论文', exact: true }).click();
  return completed;
}

test('公开检索元数据和可得状态完整来自实际 HTTP', async ({ page }) => {
  test.setTimeout(120_000);
  const response = await search(page);
  expect(response.status()).toBe(200);
  const result = await response.json() as Search;
  expect(result.items.length).toBeGreaterThan(0);
  for (const [index, hit] of result.items.entries()) {
    const row = panel(page).locator('li').nth(index);
    await expect(row).toBeVisible();
    await expect(row.getByRole('heading')).toHaveText(hit.title);
    await expect(row).toContainText(hit.provider);
    await expect(row).toContainText('原文可得状态');
    expect(hit.source_key).toBeTruthy();
    expect(URL.canParse(hit.locator)).toBe(true);
    await expect(row.getByRole('link', { name: '查看原始来源', exact: true })).toHaveAttribute('href', hit.locator);
  }
});

test('勾选和清除实际检索结果均不自动提交导入', async ({ page }) => {
  test.setTimeout(120_000);
  const writes: string[] = [];
  page.on('request', request => {
    if (request.url().includes('/api/v1/') && !['GET', 'HEAD'].includes(request.method())) writes.push(request.url());
  });
  expect((await search(page)).status()).toBe(200);
  const checkbox = panel(page).getByRole('checkbox').first();
  await expect(checkbox).toBeEnabled();
  await checkbox.check();
  await expect(panel(page).getByRole('button', { name: '批量入库 1 条', exact: true })).toBeEnabled();
  await panel(page).getByRole('button', { name: '取消全部勾选', exact: true }).click();
  await expect(checkbox).not.toBeChecked();
  await expect(panel(page).getByRole('button', { name: /批量入库/ })).toBeDisabled();
  expect(writes).toEqual([]);
});

test('真实输入校验失败显示错误且不展示检索结果', async ({ page }) => {
  const response = await search(page, 'x'.repeat(1001));
  expect(response.status()).toBe(400);
  await expect(panel(page).getByText(/检索未完成/)).toBeVisible();
  await expect(panel(page).locator('li')).toHaveCount(0);
});

test('公开检索发送完整关键词并保留 HTTP 契约字段', async ({ page }) => {
  test.setTimeout(120_000);
  const query = 'Attention Is All You Need';
  const response = await search(page, query);
  expect(response.status()).toBe(200);
  expect(response.request().method()).toBe('GET');
  expect(new URL(response.url()).searchParams.get('q')).toBe(query);
  const result = await response.json() as Search;
  expect(result.schema_version).toBe(1);
  expect(result.query).toBe(query);
  expect(result.items.length).toBeGreaterThan(0);
  expect(result.items.every(hit => hit.content_state === 'abstract_only')).toBe(true);
  await expect(panel(page).getByText(/原文可得状态 · 仅摘要\/元数据/).first()).toBeVisible();
});

test('390px 布局完整呈现实际检索结果且没有水平溢出', async ({ page }) => {
  test.setTimeout(120_000);
  await page.setViewportSize({ width: 390, height: 844 });
  const response = await search(page);
  expect(response.status()).toBe(200);
  const result = await response.json() as Search;
  expect(result.items.length).toBeGreaterThan(0);
  await expect(panel(page).locator('li').first().getByRole('heading')).toHaveText(result.items[0].title);
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
});
