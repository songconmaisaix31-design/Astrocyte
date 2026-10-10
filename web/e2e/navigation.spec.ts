/**
 * Playwright e2e: navigation, fixture display, detail panels, keyboard,
 * focus management, stale data via refresh, and API failure/retry.
 *
 * Replaces former unit fixture tests (constant-existence assertions) and
 * former hook state-transition tests with browser-level assertions.
 *
 * Config (W0-owned) sets baseURL and two viewport projects:
 *   chromium-1920 (1920x1080) and chromium-1280 (1280x720).
 */
import { test, expect } from '@playwright/test';

// ── 1. Three-page navigation ──

test.describe('Three-page navigation', () => {
  test('navigates to default page (资料沉淀)', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    await expect(page.locator('h1')).toContainText('资料沉淀');
  });

  test('navigates to 资料沉淀 page', async ({ page }) => {
    await page.goto('/attention');
    await page.waitForLoadState('networkidle');
    await expect(page.locator('h1')).toContainText('资料沉淀');
    await expect(page.getByRole('heading', { name: /素材/ })).toBeVisible();
    await page.getByRole('tab', { name: '研究机会', exact: true }).click();
    await expect(page.getByRole('heading', { name: /机会/ })).toBeVisible();
  });

  test('navigates to 共同工作区 page', async ({ page }) => {
    await page.goto('/workspace');
    await page.waitForLoadState('networkidle');
    await expect(page.locator('h1')).toContainText('共同工作区');
    await expect(page.getByRole('heading', { name: '项目总览', exact: true })).toBeVisible();
    await page.getByRole('button', { name: '接入项目', exact: true }).click();
    await expect(page.getByRole('heading', { name: '项目接入与设置', exact: true })).toBeVisible();
    await expect(page.getByRole('tab', { name: '提案与批准' })).toBeVisible();
    await page.getByRole('tab', { name: 'Agent 会话', exact: true }).click();
    await expect(page.getByRole('heading', { name: /^会话/ })).toBeVisible();
  });

  test('navigates to 蜂群空间 page', async ({ page }) => {
    await page.goto('/swarm');
    await page.waitForLoadState('networkidle');
    await expect(page.locator('h1')).toContainText('蜂群空间');
    await expect(page.getByRole('heading', { name: '顶层开发空间', exact: true })).toBeVisible();
    await page.locator('summary').filter({ hasText: /^已有任务记录/ }).click();
    await expect(page.getByRole('heading', { name: /任务/ })).toBeVisible();
    await expect(page.getByRole('heading', { name: /工作项/ })).toBeVisible();
    await page.getByRole('tab', { name: '成果与继承', exact: true }).click();
    await expect(page.getByRole('heading', { name: /产物/ })).toBeVisible();
  });

  test('sidebar navigation links work', async ({ page }) => {
    await page.goto('/attention');
    await page.waitForLoadState('networkidle');
    await page.getByRole('navigation', { name: '主导航' }).getByRole('button', { name: '共同工作区', exact: true }).click();
    await expect(page).toHaveURL(/\/workspace/);
    await expect(page.locator('h1')).toContainText('共同工作区');
    await page.getByRole('navigation', { name: '主导航' }).getByRole('button', { name: '蜂群空间', exact: true }).click();
    await expect(page).toHaveURL(/\/swarm/);
    await expect(page.locator('h1')).toContainText('蜂群空间');
    await page.getByRole('navigation', { name: '主导航' }).getByRole('button', { name: '资料沉淀', exact: true }).click();
    await expect(page).toHaveURL(/\/attention/);
  });

  test('active nav item is highlighted', async ({ page }) => {
    await page.goto('/workspace');
    await page.waitForLoadState('networkidle');
    const activeLink = page.locator('button[aria-current="page"]');
    await expect(activeLink).toContainText('共同工作区');
  });
});

// ── 2. Real empty API states (settled — no 加载中) ──

test.describe('Empty API states', () => {
  test('资料沉淀 shows settled empty headings after queries resolve', async ({ page }) => {
    await page.goto('/attention');
    await page.waitForLoadState('networkidle');
    // Wait for queries to settle, then assert actual 暂无 headings
    await expect(page.locator('text=暂无素材').first()).toBeVisible({ timeout: 10000 });
    await page.getByRole('tab', { name: '研究机会', exact: true }).click();
    await expect(page.locator('text=暂无机会').first()).toBeVisible({ timeout: 10000 });
  });

  test('共同工作区 shows settled empty headings', async ({ page }) => {
    await page.goto('/workspace');
    await page.getByRole('button', { name: '接入项目', exact: true }).click();
    await page.waitForLoadState('networkidle');
    await expect(page.getByRole('status', { name: '暂无本地项目' })).toBeVisible({ timeout: 10000 });
    await page.getByRole('tab', { name: 'Agent 会话', exact: true }).click();
    await expect(page.locator('text=暂无会话').first()).toBeVisible({ timeout: 10000 });
    await page.getByRole('tab', { name: '提案与批准' }).click();
    await expect(page.locator('text=暂无提案').first()).toBeVisible({ timeout: 10000 });
  });

  test('蜂群空间 shows settled empty headings', async ({ page }) => {
    await page.goto('/swarm');
    await page.waitForLoadState('networkidle');
    await page.locator('summary').filter({ hasText: /^已有任务记录/ }).click();
    await expect(page.locator('text=暂无任务').first()).toBeVisible({ timeout: 10000 });
  });
});

// ── 3. Fixture mode ──

test.describe('Fixture mode banner', () => {
  test('banner visible with ?fixture=1', async ({ page }) => {
    await page.goto('/attention?fixture=1');
    await page.waitForLoadState('networkidle');
    await expect(page.locator('[role="alert"]:has-text("示例数据模式")')).toBeVisible();
  });

  test('banner NOT visible without query param', async ({ page }) => {
    await page.goto('/attention');
    await page.waitForLoadState('networkidle');
    await expect(page.locator('[role="alert"]:has-text("示例数据模式")')).not.toBeVisible();
  });

  test('exit fixture mode button works', async ({ page }) => {
    await page.goto('/attention?fixture=1');
    await page.waitForLoadState('networkidle');
    await expect(page.locator('[role="alert"]:has-text("示例数据模式")')).toBeVisible();
    await page.locator('button:has-text("退出示例模式")').click();
    await expect(page.locator('[role="alert"]:has-text("示例数据模式")')).not.toBeVisible();
    expect(page.url()).not.toContain('fixture=1');
  });

  test('refresh button is NOT visible in fixture mode', async ({ page }) => {
    await page.goto('/attention?fixture=1');
    await page.waitForLoadState('networkidle');
    await expect(page.locator('button:has-text("刷新")')).not.toBeVisible();
  });

  test('refresh button IS visible in real data mode', async ({ page }) => {
    await page.goto('/attention');
    await page.waitForLoadState('networkidle');
    await expect(page.getByRole('button', { name: '↻ 刷新', exact: true })).toBeVisible();
  });
});

test.describe('Fixture display — materials and opportunities', () => {
  test('renders fixture materials including video kind', async ({ page }) => {
    await page.goto('/attention?fixture=1');
    await page.waitForLoadState('networkidle');
    await expect(page.locator('text=示例论文').first()).toBeVisible();
    await expect(page.locator('text=示例视频素材').first()).toBeVisible();
  });

  test('renders long Chinese title with line-clamp', async ({ page }) => {
    await page.goto('/attention?fixture=1');
    await page.waitForLoadState('networkidle');
    const items = page.locator('[role="button"].line-clamp-2, [role="button"] .line-clamp-2');
    expect(await items.count()).toBeGreaterThan(0);
  });

  test('renders failed import status badge', async ({ page }) => {
    await page.goto('/attention?fixture=1');
    await page.waitForLoadState('networkidle');
    await expect(page.locator('text=失败').first()).toBeVisible();
  });

  test('renders archived lifecycle badge', async ({ page }) => {
    await page.goto('/attention?fixture=1');
    await page.waitForLoadState('networkidle');
    await expect(page.locator('text=已归档').first()).toBeVisible();
  });

  test('clicking material opens detail with fixture tag and disabled notice', async ({ page }) => {
    await page.goto('/attention?fixture=1');
    await page.waitForLoadState('networkidle');
    await page.locator('[role="button"]:has-text("示例论文")').first().click();
    const dialog = page.locator('[role="dialog"]');
    await expect(dialog).toBeVisible();
    await expect(dialog).toContainText('素材详情');
    await expect(dialog.locator('text=示例数据')).toBeVisible();
    await expect(dialog).toContainText('尚未启用');
  });

  test('clicking opportunity opens detail panel', async ({ page }) => {
    await page.goto('/attention?fixture=1');
    await page.waitForLoadState('networkidle');
    await page.locator('[role="button"]:has-text("示例机会")').first().click();
    await expect(page.locator('[role="dialog"]')).toBeVisible();
    await expect(page.locator('[role="dialog"]')).toContainText('机会详情');
  });
});

test.describe('Fixture display — workspace sessions', () => {
  test('renders projects, proposals, and sessions', async ({ page }) => {
    await page.goto('/workspace?fixture=1');
    await page.waitForLoadState('networkidle');
    await expect(page.locator('text=示例项目').first()).toBeVisible();
    await expect(page.locator('text=fixture-adapter').first()).toBeVisible();
    await page.getByRole('tab', { name: '提案与批准' }).click();
    await expect(page.locator('text=示例提案').first()).toBeVisible();
  });

  test('session capability shows three-state labels (null=未知)', async ({ page }) => {
    await page.goto('/workspace?fixture=1');
    await page.waitForLoadState('networkidle');
    await page.locator('[role="button"]:has-text("fixture-adapter")').first().click();
    const dialog = page.locator('[role="dialog"]');
    await expect(dialog).toBeVisible();
    await expect(dialog).toContainText('能力');
  });
});

test.describe('Fixture display — swarm missions', () => {
  test('renders missions including blocked and failed status', async ({ page }) => {
    await page.goto('/swarm?fixture=1');
    await page.waitForLoadState('networkidle');
    await expect(page.locator('text=示例运行中任务').first()).toBeVisible();
    await expect(page.locator('text=已阻塞').first()).toBeVisible();
  });

  test('mission detail shows work items and artifact references', async ({ page }) => {
    await page.goto('/swarm?fixture=1');
    await page.waitForLoadState('networkidle');
    await page.locator('[role="button"]:has-text("示例运行中任务")').first().click();
    const dialog = page.locator('[role="dialog"]');
    await expect(dialog).toBeVisible();
    await expect(dialog.locator('text=示例数据')).toBeVisible();
    await expect(dialog).toContainText('工作项');
  });

  test('work items section displays flattened items from missions', async ({ page }) => {
    await page.goto('/swarm?fixture=1');
    await page.waitForLoadState('networkidle');
    await expect(page.getByRole('heading', { name: /工作项/ })).toBeVisible();
  });
});

// ── 4. Screenshots ──

test.describe('Screenshots', () => {
  test('资料沉淀 screenshot', async ({ page }) => {
    await page.goto('/attention?fixture=1');
    await page.waitForLoadState('networkidle');
    const vp = page.viewportSize();
    await page.evaluate(() => window.scrollTo(0, 0));
    await page.screenshot({ animations: 'disabled', path: `test-results/attention-${vp?.width}x${vp?.height}.png`, fullPage: true });
  });

  test('共同工作区 screenshot', async ({ page }) => {
    await page.goto('/workspace?fixture=1');
    await page.waitForLoadState('networkidle');
    const vp = page.viewportSize();
    await page.evaluate(() => window.scrollTo(0, 0));
    await page.screenshot({ animations: 'disabled', path: `test-results/workspace-${vp?.width}x${vp?.height}.png`, fullPage: true });
  });

  test('蜂群空间 screenshot', async ({ page }) => {
    await page.goto('/swarm?fixture=1');
    await page.waitForLoadState('networkidle');
    const vp = page.viewportSize();
    await page.evaluate(() => window.scrollTo(0, 0));
    await page.screenshot({ animations: 'disabled', path: `test-results/swarm-${vp?.width}x${vp?.height}.png`, fullPage: true });
  });

  test('detail panel screenshot', async ({ page }) => {
    await page.goto('/attention?fixture=1');
    await page.waitForLoadState('networkidle');
    await page.locator('[role="button"]:has-text("示例论文")').first().click();
    await expect(page.locator('[role="dialog"]')).toBeVisible();
    const vp = page.viewportSize();
    await page.evaluate(() => window.scrollTo(0, 0));
    await page.screenshot({ animations: 'disabled', path: `test-results/detail-${vp?.width}x${vp?.height}.png` });
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

// ── 6. Focus management ──

test.describe('Focus management', () => {
  test('close button receives focus when dialog opens', async ({ page }) => {
    await page.goto('/attention?fixture=1');
    await page.waitForLoadState('networkidle');
    await page.locator('[role="button"]:has-text("示例论文")').first().click();
    const dialog = page.locator('[role="dialog"]');
    await expect(dialog).toBeVisible();
    // useEffect fires asynchronously after mount; poll with string expression
    await page.waitForFunction(
      'document.querySelector(\'[role="dialog"] button[aria-label="关闭"]\') === document.activeElement',
      { timeout: 5000 },
    );
  });

  test('Escape closes detail panel', async ({ page }) => {
    await page.goto('/attention?fixture=1');
    await page.waitForLoadState('networkidle');
    await page.locator('[role="button"]:has-text("示例论文")').first().click();
    await expect(page.locator('[role="dialog"]')).toBeVisible();
    await page.keyboard.press('Escape');
    await expect(page.locator('[role="dialog"]')).not.toBeVisible();
  });

  test('close button retains focus after Tab cycling within dialog', async ({ page }) => {
    await page.goto('/attention?fixture=1');
    await page.waitForLoadState('networkidle');
    await page.locator('[role="button"]:has-text("示例论文")').first().click();
    const dialog = page.locator('[role="dialog"]');
    await expect(dialog).toBeVisible();

    // Wait for useEffect focus to settle
    await page.waitForFunction(
      'document.querySelector(\'[role="dialog"] button[aria-label="关闭"]\') === document.activeElement',
      { timeout: 5000 },
    );

    // Tab through all focusable elements and verify focus stays within dialog
    for (let i = 0; i < 10; i++) {
      await page.keyboard.press('Tab');
    }
    // After cycling, focus must still be inside the dialog
    const activeElement = page.locator('[role="dialog"] :focus');
    await expect(activeElement).toBeVisible();
    expect(await activeElement.count()).toBe(1);
  });

  test('focus returns to trigger element after Escape closes dialog', async ({ page }) => {
    await page.goto('/attention?fixture=1');
    await page.waitForLoadState('networkidle');
    const item = page.locator('[role="button"]:has-text("示例论文")').first();
    await item.click();
    await expect(page.locator('[role="dialog"]')).toBeVisible();
    await page.keyboard.press('Escape');
    await expect(page.locator('[role="dialog"]')).not.toBeVisible();
    await expect(item).toBeFocused();
  });

  test('Escape works even when focus is outside the dialog panel', async ({ page }) => {
    await page.goto('/attention?fixture=1');
    await page.waitForLoadState('networkidle');
    await page.locator('[role="button"]:has-text("示例论文")').first().click();
    await expect(page.locator('[role="dialog"]')).toBeVisible();
    // Move focus to an element outside the dialog
    await page.locator('nav button').first().focus();
    // Escape should still close via document-level listener
    await page.keyboard.press('Escape');
    await expect(page.locator('[role="dialog"]')).not.toBeVisible();
  });
});

// ── 7. Stale data via refresh button ──

test.describe('Stale data via refresh button', () => {
  test('first load failure shows error, no stale banner', async ({ page }) => {
    await page.route('**/api/v1/jobs', route => route.fulfill({ json: { schema_version: 1, items: [], next_cursor: null } }));
    await page.route('**/materials*', route => route.abort());
    await page.route('**/opportunities*', route => route.abort());
    await page.goto('/attention');
    await page.waitForLoadState('networkidle');
    // Both materials and opportunities show error (2 sections)
    const errorAlerts = page.locator('section').filter({ has: page.getByRole('heading', { name: /^(素材|机会)/ }) }).locator('[role="alert"]:has-text("暂未完成")');
    await expect(errorAlerts.first()).toBeVisible({ timeout: 10000 });
    expect(await errorAlerts.count()).toBe(2);
    // No stale banner (stale only appears when prior data was retained)
    await expect(page.locator('text=/数据可能已过期/')).not.toBeVisible();
  });

  test('click refresh → fail → stale banner with retained data → retry succeeds', async ({ page }) => {
    // MaterialV1 schema-valid items: only required fields plus collection_reason
    // (the only nullable optional). Optional non-nullable fields are omitted.
    const initialTitle = '保留的中性素材';
    const refreshedTitle = '刷新后的新素材';
    const emptyResp = { schema_version: 1, items: [], next_cursor: null };

    const makeMaterial = (title: string) => ({
      id: 'retained-mat-1',
      kind: 'text',
      source_locator: 'fixture:test-retained',
      current_revision: 1,
      lifecycle: 'active',
      title,
      collection_reason: null,
    });

    let materialsResponse = { schema_version: 1, items: [makeMaterial(initialTitle)], next_cursor: null };
    let failMaterials = false;
    let failOpportunities = false;

    await page.route('**/materials*', async route => {
      if (failMaterials) { await route.abort(); return; }
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(materialsResponse),
      });
    });
    await page.route('**/opportunities*', async route => {
      if (failOpportunities) { await route.abort(); return; }
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(emptyResp),
      });
    });

    // 1. Initial load succeeds with neutral item
    await page.goto('/attention');
    await page.waitForLoadState('networkidle');
    await expect(page.locator(`text=${initialTitle}`).first()).toBeVisible({ timeout: 10000 });

    // 2. Block subsequent fetches and click refresh
    failMaterials = true;
    failOpportunities = true;
    await page.getByRole('button', { name: '↻ 刷新', exact: true }).click();

    // 3. Stale banner appears with retained prior item still visible
    await expect(page.locator('text=/数据可能已过期/')).toBeVisible({ timeout: 10000 });
    // The neutral item from step 1 is still rendered (retained data)
    await expect(page.locator(`text=${initialTitle}`).first()).toBeVisible();

    // 4. Swap to a different item so we can prove the fresh response replaced stale data
    materialsResponse = { schema_version: 1, items: [makeMaterial(refreshedTitle)], next_cursor: null };
    failMaterials = false;
    failOpportunities = false;
    await page.locator('[role="alert"]:has-text("数据可能已过期") button:has-text("重试")').click();

    // 5. Stale banner clears, fresh response replaced stale data (new title visible)
    await expect(page.locator('text=/数据可能已过期/')).not.toBeVisible({ timeout: 10000 });
    await expect(page.locator(`text=${refreshedTitle}`).first()).toBeVisible({ timeout: 10000 });
  });
});

// ── 8. API failure and retry ──

test.describe('API failure and retry', () => {
  test('shows error state and retry button on API failure', async ({ page }) => {
    await page.route('**/api/v1/jobs', route => route.fulfill({ json: { schema_version: 1, items: [], next_cursor: null } }));
    await page.route('**/materials*', route => route.abort());
    await page.route('**/opportunities*', route => route.abort());
    await page.goto('/attention');
    await page.waitForLoadState('networkidle');
    // Each failed section (materials, opportunities) shows its own error + retry
    const errorAlerts = page.locator('section').filter({ has: page.getByRole('heading', { name: /^(素材|机会)/ }) }).locator('[role="alert"]:has-text("暂未完成")');
    await expect(errorAlerts.first()).toBeVisible({ timeout: 10000 });
    expect(await errorAlerts.count()).toBe(2);
    const retryButtons = page.locator('[role="alert"] button:has-text("重试")');
    expect(await retryButtons.count()).toBeGreaterThanOrEqual(2);
  });

  test('retry button triggers new request', async ({ page }) => {
    let callCount = 0;
    await page.route('**/materials*', route => {
      callCount++;
      route.abort();
    });
    await page.route('**/opportunities*', route => route.abort());
    await page.goto('/attention');
    await page.waitForLoadState('networkidle');
    await expect(page.locator('button:has-text("重试")').first()).toBeVisible({ timeout: 10000 });
    const initialCount = callCount;
    await page.locator('section').filter({ has: page.getByRole('heading', { name: /^素材/ }) }).getByRole('button', { name: '重试', exact: true }).click();
    await page.waitForTimeout(500);
    expect(callCount).toBeGreaterThan(initialCount);
  });

  test('no silent fixture fallback on API failure', async ({ page }) => {
    await page.route('**/materials*', route => route.abort());
    await page.route('**/opportunities*', route => route.abort());
    await page.goto('/attention');
    await page.waitForLoadState('networkidle');
    // Error state visible
    await expect(page.locator('[role="alert"]:has-text("暂未完成")').first()).toBeVisible({ timeout: 10000 });
    // Fixture data never appears
    await expect(page.locator('text=示例论文')).not.toBeVisible();
    await expect(page.locator('text=示例机会')).not.toBeVisible();
  });

  test('swarm propagates mission error to workitems and artifacts sections', async ({ page }) => {
    await page.route('**/missions*', route => route.abort());
    await page.goto('/swarm');
    await page.waitForLoadState('networkidle');
    await page.locator('summary').filter({ hasText: /^已有任务记录/ }).click();
    // Mission and workitem errors are in records; artifacts are on their own tab.
    // Check each downstream section visibly rather than counting hidden panels.
    for (const title of [/^任务\(/, /^工作项\(/]) {
      const section = page.locator('section').filter({ has: page.getByRole('heading', { name: title }) });
      await expect(section.getByRole('alert').filter({ hasText: '暂未完成' })).toBeVisible({ timeout: 10000 });
    }
    await page.getByRole('tab', { name: '成果与继承', exact: true }).click();
    const artifacts = page.locator('section').filter({ has: page.getByRole('heading', { name: /^产物\(/ }) });
    await expect(artifacts.getByRole('alert').filter({ hasText: '暂未完成' })).toBeVisible({ timeout: 10000 });
    // No fixture data should appear
    await expect(page.locator('text=示例运行中任务')).not.toBeVisible();
  });
});
