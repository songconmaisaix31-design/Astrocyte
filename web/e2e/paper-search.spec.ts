import { test, expect, type Page } from '@playwright/test';

/**
 * Paper search entry in Attention: clear entry → metadata + availability note →
 * checkbox → batch import (never full-auto). Fixture mode demonstrates the
 * interaction; real mode must surface an honest unavailable/error state and
 * never fabricate results. Real search is a GET /papers/search returning hits
 * with source_key/provider/content_state/pdf_urls.
 */

const panel = (page: Page) => page.locator('section[aria-label="论文检索"]');

async function openSearch(page: Page) {
  await page.getByRole('button', { name: /检索论文/ }).first().click();
  await expect(panel(page)).toBeVisible();
}

test.describe('论文检索（示例）', () => {
  test('renders metadata results with an availability note', async ({ page }) => {
    await page.goto('/attention?fixture=1');
    await page.waitForLoadState('networkidle');
    await openSearch(page);
    await expect(panel(page).getByText('示例论文：公开 arXiv 文献（用于验证检索与勾选流程）')).toBeVisible();
    await expect(panel(page).getByText(/arXiv 2504\.16054/)).toBeVisible();
    await expect(panel(page).getByText('示例期刊论文：经 DOI 元数据索引')).toBeVisible();
    await expect(panel(page).getByText(/原文可得状态 · 仅摘要\/元数据/).first()).toBeVisible();
  });

  test('selects and clears hits explicitly without any write request', async ({ page }) => {
    const writes: string[] = [];
    page.on('request', request => {
      if (request.url().includes('/api/v1/') && !['GET', 'HEAD'].includes(request.method())) writes.push(request.url());
    });
    await page.goto('/attention?fixture=1');
    await page.waitForLoadState('networkidle');
    await openSearch(page);
    const fullRow = panel(page).locator('li').filter({ has: page.getByText('示例论文：公开 arXiv 文献（用于验证检索与勾选流程）') });
    await fullRow.getByRole('checkbox').check();
    await expect(panel(page).getByRole('button', { name: /批量入库 1 条/ })).toBeDisabled();
    await panel(page).getByRole('button', { name: '取消全部勾选', exact: true }).click();
    await expect(panel(page).getByRole('button', { name: /批量入库/ })).toBeDisabled();
    expect(writes).toEqual([]);
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
    await expect(panel(page).getByText('示例论文：公开 arXiv 文献（用于验证检索与勾选流程）')).not.toBeVisible();
  });
});

test.describe('论文检索（真实契约对接）', () => {
  test('renders a real PaperSearchResultV1 payload and GETs the search', async ({ page }) => {
    let searched = false;
    await page.route('**/api/v1/papers/search**', route => {
      searched = true;
      expect(route.request().url()).toContain('q=transformer');
      return route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ schema_version: 1, query: 'transformer', items: [{ source_key: 'arxiv:2504.16054', provider: 'arxiv', title: 'Attention 论文', authors: ['甲'], year: 2024, arxiv_id: '2504.16054', abstract: '摘要', locator: 'https://arxiv.org/abs/2504.16054', content_state: 'abstract_only', pdf_urls: [] }], next_cursor: null, has_more: false, warnings: [] }) });
    });
    await page.goto('/attention');
    await page.waitForLoadState('networkidle');
    await openSearch(page);
    await panel(page).getByLabel('检索论文', { exact: true }).fill('transformer');
    await panel(page).getByRole('button', { name: '检索论文', exact: true }).click();
    await expect(panel(page).getByText('Attention 论文')).toBeVisible();
    await expect(panel(page).getByText(/原文可得状态 · 仅摘要\/元数据/)).toBeVisible();
    expect(searched).toBe(true);
  });
});

test.describe('论文检索（移动布局）', () => {
  test('renders results without horizontal overflow at 390px', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 });
    await page.goto('/attention?fixture=1');
    await page.waitForLoadState('networkidle');
    await openSearch(page);
    await expect(panel(page).getByText('示例论文：公开 arXiv 文献（用于验证检索与勾选流程）')).toBeVisible();
    await expect(panel(page).getByText('示例期刊论文：经 DOI 元数据索引')).toBeVisible();
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
  });
});
