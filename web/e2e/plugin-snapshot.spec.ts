import { test, expect, type Page } from '@playwright/test';

/**
 * Plugin snapshot paste-review: human pastes the extension's JSON snapshot
 * (W1 shape: source_url / content_state / host_family), reviews metadata only,
 * and never treats the snapshot/abstract as the body. arXiv routes to the arxiv
 * adapter; every other public HTTPS page routes to W1's paper_url adapter.
 */

const panel = (page: Page) => page.locator('section[aria-label="插件快照复核导入"]');

async function openPanel(page: Page) {
  await page.locator('summary').filter({ hasText: /^插件快照复核导入/ }).click();
  await expect(panel(page)).toBeVisible();
}

test.describe('插件快照复核导入（示例）', () => {
  test('rejects non-JSON and a snapshot without a source_url', async ({ page }) => {
    await page.goto('/attention?fixture=1');
    await page.waitForLoadState('networkidle');
    await openPanel(page);
    await panel(page).getByLabel('粘贴插件快照 JSON', { exact: true }).fill('not json');
    await panel(page).getByRole('button', { name: '复核快照', exact: true }).click();
    await expect(panel(page).getByText(/不是有效的 JSON/)).toBeVisible();
    await panel(page).getByLabel('粘贴插件快照 JSON', { exact: true }).fill('{}');
    await panel(page).getByRole('button', { name: '复核快照', exact: true }).click();
    await expect(panel(page).getByText(/缺少来源链接/)).toBeVisible();
  });

  test('reviews metadata and never enables import in fixture mode', async ({ page }) => {
    const writes: string[] = [];
    page.on('request', request => {
      if (request.url().includes('/api/v1/') && !['GET', 'HEAD'].includes(request.method())) writes.push(request.url());
    });
    await page.goto('/attention?fixture=1');
    await page.waitForLoadState('networkidle');
    await openPanel(page);
    await panel(page).getByLabel('粘贴插件快照 JSON', { exact: true }).fill(JSON.stringify({ schema_version: 1, source_url: 'https://arxiv.org/abs/2504.16054', host_family: 'arxiv', title: '快照标题', authors: ['甲'], arxiv_id: '2504.16054', abstract: '快照摘要', content_state: 'abstract_only' }));
    await panel(page).getByRole('button', { name: '复核快照', exact: true }).click();
    await expect(panel(page).getByRole('heading', { name: '快照标题' })).toBeVisible();
    await expect(panel(page).getByText(/不是正文/).first()).toBeVisible();
    await expect(panel(page).getByRole('button', { name: '按来源导入正文（arXiv）', exact: true })).toBeDisabled();
    expect(writes).toEqual([]);
  });
});

test.describe('插件快照复核导入（真实）', () => {
  test('routes a non-arXiv snapshot to the paper_url adapter instead of pretending it is unsupported', async ({ page }) => {
    await page.goto('/attention');
    await page.waitForLoadState('networkidle');
    await openPanel(page);
    await panel(page).getByLabel('粘贴插件快照 JSON', { exact: true }).fill(JSON.stringify({ schema_version: 1, source_url: 'https://journals.plos.org/plosone/article?id=x', host_family: 'plos', title: 'PLOS 论文', source_key: 'doi:10.1/example', content_state: 'readable_fulltext' }));
    await panel(page).getByRole('button', { name: '复核快照', exact: true }).click();
    await expect(panel(page).getByRole('heading', { name: 'PLOS 论文' })).toBeVisible();
    await expect(panel(page).getByRole('button', { name: '按来源导入正文（paper_url）', exact: true })).toBeVisible();
  });

  test('gates a paywalled snapshot honestly without a bypass', async ({ page }) => {
    await page.goto('/attention');
    await page.waitForLoadState('networkidle');
    await openPanel(page);
    await panel(page).getByLabel('粘贴插件快照 JSON', { exact: true }).fill(JSON.stringify({ schema_version: 1, source_url: 'https://example.com/p', host_family: 'generic', title: '付费墙论文', content_state: 'paywall', warning: 'Access restriction detected' }));
    await panel(page).getByRole('button', { name: '复核快照', exact: true }).click();
    await expect(panel(page).getByRole('heading', { name: '付费墙论文' })).toBeVisible();
    await expect(panel(page).getByText(/无绕过/)).toBeVisible();
    await expect(panel(page).getByRole('button', { name: /按来源导入正文/ })).not.toBeVisible();
  });
});

test.describe('插件快照复核导入（移动布局）', () => {
  test('renders without horizontal overflow at 390px', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 });
    await page.goto('/attention?fixture=1');
    await page.waitForLoadState('networkidle');
    await openPanel(page);
    await panel(page).getByLabel('粘贴插件快照 JSON', { exact: true }).fill(JSON.stringify({ schema_version: 1, source_url: 'https://arxiv.org/abs/2504.16054', host_family: 'arxiv', title: '窄屏快照标题', authors: ['甲'] }));
    await panel(page).getByRole('button', { name: '复核快照', exact: true }).click();
    await expect(panel(page).getByRole('heading', { name: '窄屏快照标题' })).toBeVisible();
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
  });
});
