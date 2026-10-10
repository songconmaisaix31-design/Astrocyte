import { test, expect, type Page } from '@playwright/test';

/**
 * Paper search entry in Attention: clear entry → metadata + availability →
 * checkbox → batch import (never full-auto). Fixture mode demonstrates the
 * interaction; real mode must surface an honest unavailable/error state and
 * never fabricate results.
 */

const panel = (page: Page) => page.locator('section[aria-label="论文检索"]');

async function openSearch(page: Page) {
  await page.getByRole('button', { name: /检索论文/ }).first().click();
  await expect(panel(page)).toBeVisible();
}

test.describe('论文检索（示例）', () => {
  test('renders results with availability, restriction note and selection', async ({ page }) => {
    await page.goto('/attention?fixture=1');
    await page.waitForLoadState('networkidle');
    await openSearch(page);
    await expect(panel(page).getByText('示例论文：全文可得的公开文献')).toBeVisible();
    await expect(panel(page).getByText('原文可得状态 · 原文可得').first()).toBeVisible();
    await expect(panel(page).getByText('示例受限论文：付费墙')).toBeVisible();
    await expect(panel(page).getByText(/受限说明 · 来源需要订阅/)).toBeVisible();
    await expect(panel(page).getByText(/原文可得性未知/)).toBeVisible();
    await expect(panel(page).getByText('示例已入库论文')).toBeVisible();
  });

  test('selects and clears hits explicitly without any write request', async ({ page }) => {
    const writes: string[] = [];
    page.on('request', request => {
      if (request.url().includes('/api/v1/') && !['GET', 'HEAD'].includes(request.method())) writes.push(request.url());
    });
    await page.goto('/attention?fixture=1');
    await page.waitForLoadState('networkidle');
    await openSearch(page);
    const fullRow = panel(page).locator('li').filter({ has: page.getByText('示例论文：全文可得的公开文献') });
    await fullRow.getByRole('checkbox').check();
    await expect(panel(page).getByRole('button', { name: /批量入库 1 条/ })).toBeDisabled();
    await panel(page).getByRole('button', { name: '取消全部勾选', exact: true }).click();
    await expect(panel(page).getByRole('button', { name: /批量入库/ })).toBeDisabled();
    expect(writes).toEqual([]);
  });

  test('unknown and already-imported hits cannot be selected', async ({ page }) => {
    await page.goto('/attention?fixture=1');
    await page.waitForLoadState('networkidle');
    await openSearch(page);
    const unknownRow = panel(page).locator('li').filter({ has: page.getByText('示例可得性未知论文') });
    await expect(unknownRow.getByRole('checkbox')).toBeDisabled();
    const importedRow = panel(page).locator('li').filter({ has: page.getByText('示例已入库论文') });
    await expect(importedRow.getByRole('checkbox')).toBeDisabled();
    await expect(importedRow.getByRole('button', { name: '查看已入库资料', exact: true })).toBeVisible();
  });
});

test.describe('论文检索（真实不可用）', () => {
  test('surfaces an honest unavailable state instead of fake results', async ({ page }) => {
    await page.route('**/api/v1/papers/search**', route => route.fulfill({ status: 501, contentType: 'application/json', body: JSON.stringify({ schema_version: 1, error: { code: 'unsupported_capability', message: 'Not implemented', request_id: 'r', retryable: false, required_action: '论文检索尚未接入' } }) }));
    await page.goto('/attention');
    await page.waitForLoadState('networkidle');
    await openSearch(page);
    await panel(page).getByLabel('检索论文', { exact: true }).fill('transformer');
    await panel(page).getByRole('button', { name: '检索论文', exact: true }).click();
    await expect(panel(page).getByText(/检索未完成/)).toBeVisible({ timeout: 10000 });
    await expect(panel(page).getByText('示例论文：全文可得的公开文献')).not.toBeVisible();
  });
});
