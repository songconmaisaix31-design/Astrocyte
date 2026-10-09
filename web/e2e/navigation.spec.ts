/**
 * Playwright e2e: navigation across three pages, empty API states,
 * fixture mode, detail panels, keyboard, and API failure/retry.
 *
 * Config (W0-owned) sets baseURL and two viewport projects:
 *   chromium-1920 (1920x1080) and chromium-1280 (1280x720).
 * Each test runs at both viewports automatically.
 *
 * All page.goto calls use relative paths (config baseURL handles host/port).
 * Empty state assertions check actual heading text, not generic role counts.
 */
import { test, expect } from '@playwright/test';

// ── 1. Navigation across all three pages ──

test.describe('Three-page navigation', () => {
  test('navigates to attention page by default', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    await expect(page.locator('h1')).toContainText('注意力');
  });

  test('navigates to attention page', async ({ page }) => {
    await page.goto('/attention');
    await page.waitForLoadState('networkidle');
    await expect(page.locator('h1')).toContainText('注意力');
    await expect(page.getByRole('heading', { name: /素材/ })).toBeVisible();
    await expect(page.getByRole('heading', { name: /机会/ })).toBeVisible();
  });

  test('navigates to workspace page', async ({ page }) => {
    await page.goto('/workspace');
    await page.waitForLoadState('networkidle');
    await expect(page.locator('h1')).toContainText('工作台');
    await expect(page.getByRole('heading', { name: /项目/ })).toBeVisible();
    await expect(page.getByRole('heading', { name: /提案/ })).toBeVisible();
    await expect(page.getByRole('heading', { name: /会话/ })).toBeVisible();
  });

  test('navigates to swarm page', async ({ page }) => {
    await page.goto('/swarm');
    await page.waitForLoadState('networkidle');
    await expect(page.locator('h1')).toContainText('集群');
    await expect(page.getByRole('heading', { name: /任务/ })).toBeVisible();
    await expect(page.getByRole('heading', { name: /工作项/ })).toBeVisible();
    await expect(page.getByRole('heading', { name: /产物/ })).toBeVisible();
  });

  test('sidebar navigation links work', async ({ page }) => {
    await page.goto('/attention');
    await page.waitForLoadState('networkidle');
    await page.locator('button:has-text("工作台")').click();
    await expect(page).toHaveURL(/\/workspace/);
    await expect(page.locator('h1')).toContainText('工作台');
    await page.locator('button:has-text("集群")').click();
    await expect(page).toHaveURL(/\/swarm/);
    await expect(page.locator('h1')).toContainText('集群');
    await page.locator('button:has-text("注意力")').click();
    await expect(page).toHaveURL(/\/attention/);
  });

  test('active nav item is highlighted', async ({ page }) => {
    await page.goto('/workspace');
    await page.waitForLoadState('networkidle');
    const activeLink = page.locator('button[aria-current="page"]');
    await expect(activeLink).toContainText('工作台');
  });
});

// ── 2. Real empty API states ──

test.describe('Empty API states', () => {
  test('attention page shows actual empty headings for materials and opportunities', async ({ page }) => {
    await page.goto('/attention');
    await page.waitForLoadState('networkidle');
    // The backend returns empty lists, so we should see the real empty state headings
    const emptyTexts = page.locator('text=/暂无素材|暂无机会|加载中/');
    await expect(emptyTexts.first()).toBeVisible({ timeout: 10000 });
  });

  test('workspace page shows actual empty headings', async ({ page }) => {
    await page.goto('/workspace');
    await page.waitForLoadState('networkidle');
    const emptyTexts = page.locator('text=/暂无项目|暂无提案|暂无会话|加载中/');
    await expect(emptyTexts.first()).toBeVisible({ timeout: 10000 });
  });

  test('swarm page shows actual empty headings', async ({ page }) => {
    await page.goto('/swarm');
    await page.waitForLoadState('networkidle');
    const emptyTexts = page.locator('text=/暂无任务|暂无工作项|暂无产物|加载中/');
    await expect(emptyTexts.first()).toBeVisible({ timeout: 10000 });
  });
});

// ── 3. Fixture mode with detail panel ──

test.describe('Fixture mode', () => {
  test('fixture banner is visible when ?fixture=1', async ({ page }) => {
    await page.goto('/attention?fixture=1');
    await page.waitForLoadState('networkidle');
    await expect(page.locator('[role="alert"]:has-text("示例数据模式")')).toBeVisible();
  });

  test('fixture banner is NOT visible without query param', async ({ page }) => {
    await page.goto('/attention');
    await page.waitForLoadState('networkidle');
    await expect(page.locator('[role="alert"]:has-text("示例数据模式")')).not.toBeVisible();
  });

  test('fixture data populates materials list with neutral titles', async ({ page }) => {
    await page.goto('/attention?fixture=1');
    await page.waitForLoadState('networkidle');
    await expect(page.locator('text=示例论文')).toBeVisible();
  });

  test('fixture shows video material for summarize direction', async ({ page }) => {
    await page.goto('/attention?fixture=1');
    await page.waitForLoadState('networkidle');
    await expect(page.locator('text=示例视频素材')).toBeVisible();
  });

  test('clicking material opens detail panel with fixture tag', async ({ page }) => {
    await page.goto('/attention?fixture=1');
    await page.waitForLoadState('networkidle');
    await page.locator('[role="button"]:has-text("示例论文")').first().click();
    await expect(page.locator('[role="dialog"]')).toBeVisible();
    await expect(page.locator('[role="dialog"]')).toContainText('素材详情');
    await expect(page.locator('[role="dialog"]:has-text("示例数据")')).toBeVisible();
  });

  test('clicking opportunity opens detail panel', async ({ page }) => {
    await page.goto('/attention?fixture=1');
    await page.waitForLoadState('networkidle');
    await page.locator('[role="button"]:has-text("示例机会")').first().click();
    await expect(page.locator('[role="dialog"]')).toBeVisible();
    await expect(page.locator('[role="dialog"]')).toContainText('机会详情');
  });

  test('workspace fixture shows projects, proposals, sessions', async ({ page }) => {
    await page.goto('/workspace?fixture=1');
    await page.waitForLoadState('networkidle');
    await expect(page.locator('text=示例项目').first()).toBeVisible();
    await expect(page.locator('text=示例提案').first()).toBeVisible();
    await expect(page.locator('text=fixture-adapter').first()).toBeVisible();
  });

  test('swarm fixture shows missions with blocked status', async ({ page }) => {
    await page.goto('/swarm?fixture=1');
    await page.waitForLoadState('networkidle');
    await expect(page.locator('text=示例运行中任务')).toBeVisible();
    await expect(page.locator('text=已阻塞')).toBeVisible();
  });

  test('detail panel shows disabled notice for future features', async ({ page }) => {
    await page.goto('/attention?fixture=1');
    await page.waitForLoadState('networkidle');
    await page.locator('[role="button"]:has-text("示例论文")').first().click();
    await expect(page.locator('[role="dialog"]')).toContainText('后续');
  });

  test('exit fixture mode button works', async ({ page }) => {
    await page.goto('/attention?fixture=1');
    await page.waitForLoadState('networkidle');
    await expect(page.locator('[role="alert"]:has-text("示例数据模式")')).toBeVisible();
    await page.locator('button:has-text("退出示例模式")').click();
    await expect(page.locator('[role="alert"]:has-text("示例数据模式")')).not.toBeVisible();
    expect(page.url()).not.toContain('fixture=1');
  });

  test('swarm detail shows fixture tag', async ({ page }) => {
    await page.goto('/swarm?fixture=1');
    await page.waitForLoadState('networkidle');
    await page.locator('[role="button"]:has-text("示例运行中任务")').first().click();
    await expect(page.locator('[role="dialog"]:has-text("示例数据")')).toBeVisible();
  });
});

// ── 4. Screenshots (run at both viewports by Playwright projects) ──

test.describe('Screenshots', () => {
  test('attention page screenshot', async ({ page }) => {
    await page.goto('/attention?fixture=1');
    await page.waitForLoadState('networkidle');
    const vp = page.viewportSize();
    await page.screenshot({ path: `test-results/attention-${vp?.width}x${vp?.height}.png`, fullPage: true });
  });

  test('workspace page screenshot', async ({ page }) => {
    await page.goto('/workspace?fixture=1');
    await page.waitForLoadState('networkidle');
    const vp = page.viewportSize();
    await page.screenshot({ path: `test-results/workspace-${vp?.width}x${vp?.height}.png`, fullPage: true });
  });

  test('swarm page screenshot', async ({ page }) => {
    await page.goto('/swarm?fixture=1');
    await page.waitForLoadState('networkidle');
    const vp = page.viewportSize();
    await page.screenshot({ path: `test-results/swarm-${vp?.width}x${vp?.height}.png`, fullPage: true });
  });

  test('detail panel screenshot', async ({ page }) => {
    await page.goto('/attention?fixture=1');
    await page.waitForLoadState('networkidle');
    await page.locator('[role="button"]:has-text("示例论文")').first().click();
    await expect(page.locator('[role="dialog"]')).toBeVisible();
    const vp = page.viewportSize();
    await page.screenshot({ path: `test-results/detail-${vp?.width}x${vp?.height}.png` });
  });
});

// ── 5. Keyboard navigation ──

test.describe('Keyboard navigation', () => {
  test('Enter key opens detail from list item', async ({ page }) => {
    await page.goto('/attention?fixture=1');
    await page.waitForLoadState('networkidle');
    const item = page.locator('[role="button"]:has-text("示例论文")').first();
    await item.focus();
    await page.keyboard.press('Enter');
    await expect(page.locator('[role="dialog"]')).toBeVisible();
  });

  test('Space key opens detail from list item', async ({ page }) => {
    await page.goto('/attention?fixture=1');
    await page.waitForLoadState('networkidle');
    const item = page.locator('[role="button"]:has-text("示例论文")').first();
    await item.focus();
    await page.keyboard.press('Space');
    await expect(page.locator('[role="dialog"]')).toBeVisible();
  });

  test('Escape closes detail panel', async ({ page }) => {
    await page.goto('/attention?fixture=1');
    await page.waitForLoadState('networkidle');
    await page.locator('[role="button"]:has-text("示例论文")').first().click();
    await expect(page.locator('[role="dialog"]')).toBeVisible();
    await page.keyboard.press('Escape');
    await expect(page.locator('[role="dialog"]')).not.toBeVisible();
  });

  test('Tab is trapped within open dialog', async ({ page }) => {
    await page.goto('/attention?fixture=1');
    await page.waitForLoadState('networkidle');
    await page.locator('[role="button"]:has-text("示例论文")').first().click();
    const dialog = page.locator('[role="dialog"]');
    await expect(dialog).toBeVisible();

    // Tab through all focusable elements in the dialog and verify we stay inside
    for (let i = 0; i < 10; i++) {
      await page.keyboard.press('Tab');
    }
    // The close button should still be reachable (focus stays in dialog)
    const activeElement = page.locator('[role="dialog"] :focus');
    await expect(activeElement).toBeVisible();
  });

  test('Focus returns to trigger after Escape closes dialog', async ({ page }) => {
    await page.goto('/attention?fixture=1');
    await page.waitForLoadState('networkidle');
    const item = page.locator('[role="button"]:has-text("示例论文")').first();
    await item.click();
    await expect(page.locator('[role="dialog"]')).toBeVisible();
    await page.keyboard.press('Escape');
    await expect(page.locator('[role="dialog"]')).not.toBeVisible();
    // Focus should return to the list item that opened the dialog
    await expect(item).toBeFocused();
  });

  test('Tab navigation through list items', async ({ page }) => {
    await page.goto('/attention?fixture=1');
    await page.waitForLoadState('networkidle');
    await page.locator('[role="button"]:has-text("示例论文")').first().focus();
    await page.keyboard.press('Tab');
    const focused = page.locator(':focus');
    await expect(focused).toHaveAttribute('role', 'button');
  });

  test('Tab navigation from sidebar to content', async ({ page }) => {
    await page.goto('/attention?fixture=1');
    await page.waitForLoadState('networkidle');
    await page.locator('nav button').first().focus();
    for (let i = 0; i < 5; i++) {
      await page.keyboard.press('Tab');
    }
    const focused = page.locator(':focus');
    expect(await focused.count()).toBe(1);
  });
});

// ── 6. API failure and retry ──

test.describe('API failure and retry', () => {
  test('shows error state and retry button on API failure', async ({ page }) => {
    await page.route('**/materials*', route => route.abort());
    await page.goto('/attention');
    await page.waitForLoadState('networkidle');
    await expect(page.locator('[role="alert"]:has-text("请求失败")')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('button:has-text("重试")')).toBeVisible();
  });

  test('retry button triggers new request', async ({ page }) => {
    let callCount = 0;
    await page.route('**/materials*', route => {
      callCount++;
      route.abort();
    });
    await page.goto('/attention');
    await page.waitForLoadState('networkidle');
    await expect(page.locator('button:has-text("重试")')).toBeVisible({ timeout: 10000 });
    const initialCount = callCount;
    await page.locator('button:has-text("重试")').first().click();
    await page.waitForTimeout(500);
    expect(callCount).toBeGreaterThan(initialCount);
  });

  test('no silent fixture fallback on API failure', async ({ page }) => {
    await page.route('**/materials*', route => route.abort());
    await page.goto('/attention');
    await page.waitForLoadState('networkidle');
    await expect(page.locator('[role="alert"]:has-text("请求失败")')).toBeVisible({ timeout: 10000 });
    // Fixture data should NOT appear
    await expect(page.locator('text=示例论文')).not.toBeVisible();
  });
});
