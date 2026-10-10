import { test, expect } from '@playwright/test';

test('search shortcut filters loaded records and clears with Escape', async ({ page }) => {
  await page.goto('/attention?fixture=1');
  await page.keyboard.press('/');
  const search = page.getByRole('textbox', { name: '搜索当前页资料、任务或 Agent' });
  await expect(search).toBeFocused();
  await search.fill('示例视频素材');
  await expect(page.locator('main [role="button"]:has-text("示例视频素材")')).toBeVisible();
  await expect(page.locator('main [role="button"]:has-text("示例论文")')).not.toBeVisible();
  await search.fill('没有匹配的随机词');
  await expect(page.getByRole('status', { name: '没有匹配的素材' })).toBeVisible();
  await page.keyboard.press('Escape');
  await expect(search).toHaveValue('');
  await expect(page.locator('main [role="button"]:has-text("示例论文")').first()).toBeVisible();
});

test('material filters preserve selectable details and unknown fields', async ({ page }) => {
  await page.goto('/attention?fixture=1');
  await page.getByRole('group', { name: '素材类型' }).getByRole('button', { name: '视频', exact: true }).click();
  const video = page.locator('main [role="button"]:has-text("示例视频素材")');
  await expect(video).toBeVisible();
  await expect(page.locator('main [role="button"]:has-text("示例论文")')).not.toBeVisible();
  await video.click();
  const dialog = page.getByRole('dialog');
  await expect(dialog).toContainText('来源定位');
  await expect(dialog).toContainText('人工使用');
  await expect(dialog.getByRole('button', { name: '继续沉淀' })).toBeDisabled();
});

test('explicit Attention fixture never queries real materials or sends commands', async ({ page }) => {
  const requests: string[] = [];
  page.on('request', request => {
    if (request.url().includes('/api/v1/materials') || request.url().includes('/api/v1/opportunities') || request.method() !== 'GET') requests.push(`${request.method()} ${request.url()}`);
  });
  await page.goto('/attention?fixture=1');
  await page.locator('main [role="button"]:has-text("示例论文")').first().click();
  await expect(page.getByRole('dialog')).toContainText('示例模式：所有写操作');
  await expect(page.getByRole('dialog').getByRole('button', { name: '继续沉淀' })).toBeDisabled();
  expect(requests).toEqual([]);
});

test('unknown dimensions and missing collection reasons remain explicit', async ({ page }) => {
  await page.goto('/attention?fixture=1');
  await expect(page.locator('main [role="button"]').filter({ hasText: 'Fixture Deferred Opportunity' })).toContainText('未知');
  await page.locator('main [role="button"]').filter({ hasText: 'Fixture Sample Paper:' }).click();
  await expect(page.getByRole('dialog')).toContainText('收藏理由');
  await expect(page.getByRole('dialog')).toContainText('未提供');
  await expect(page.getByRole('dialog')).toContainText('未提供真实位置');
});

test('tabs support arrow, Home, End and browser history without losing fixture mode', async ({ page }) => {
  await page.goto('/attention?fixture=1');
  const overview = page.getByRole('tab', { name: '资料与线索' });
  await overview.focus();
  await page.keyboard.press('ArrowRight');
  await expect(page.getByRole('tab', { name: '研究机会' })).toBeFocused();
  await expect(page.getByRole('tab', { name: '研究机会' })).toHaveAttribute('aria-selected', 'true');
  await expect(page.getByRole('heading', { name: /^素材/ })).not.toBeVisible();
  await page.keyboard.press('End');
  await expect(page.getByRole('status', { name: '收藏尚未接入' })).toBeVisible();
  await expect(page.getByRole('button', { name: '收藏操作尚未启用' })).toBeDisabled();
  await page.keyboard.press('Home');
  await expect(overview).toBeFocused();
  await page.goBack();
  await expect(page.getByRole('tab', { name: '我的收藏' })).toHaveAttribute('aria-selected', 'true');
  await expect(page.getByRole('alert').filter({ hasText: '示例数据模式' })).toBeVisible();
});

test('page shortcuts and sidebar keep explicit fixture boundary', async ({ page }) => {
  await page.goto('/attention?fixture=1');
  await page.keyboard.press('2');
  await expect(page).toHaveURL(/workspace\?fixture=1/);
  await page.getByRole('navigation', { name: '主导航' }).getByRole('button', { name: '蜂群空间' }).click();
  await expect(page).toHaveURL(/swarm\?fixture=1/);
  await page.getByRole('button', { name: '退出示例模式' }).click();
  await expect(page).not.toHaveURL(/fixture=1/);
  await page.getByRole('tab', { name: '任务时间线', exact: true }).click();
  await expect(page.getByRole('button', { name: '拓扑视图' })).toBeDisabled();
  await expect(page.getByText('T-11 · 示例', { exact: true })).not.toBeVisible();
});

test('workspace tabs retain unknown budget and disabled resume/handoff', async ({ page }) => {
  await page.goto('/workspace?fixture=1&tab=sessions');
  await expect(page.getByRole('heading', { name: /^项目/ })).not.toBeVisible();
  await page.locator('main [role="button"]:has-text("fixture-adapter")').first().click();
  const dialog = page.getByRole('dialog');
  await expect(dialog).toContainText('上下文');
  await expect(dialog.getByRole('button', { name: '原生接续' })).toBeDisabled();
  await expect(dialog.getByRole('button', { name: '显式交接' })).toBeDisabled();
  await page.keyboard.press('Escape');
  await page.getByRole('tab', { name: '提案与批准' }).click();
  await page.locator('main [role="button"]:has-text("示例已拒绝提案")').click();
  await expect(dialog).toContainText('未知');
  await expect(dialog.getByRole('button', { name: '批准', exact: true })).toBeDisabled();
});

test('topology/list design examples have explicit provenance and keyboard details', async ({ page }) => {
  await page.goto('/swarm?fixture=1&tab=timeline');
  await page.getByRole('button', { name: '拓扑视图' }).click();
  await expect(page.getByText('设计示意 · 非实时拓扑')).toBeVisible();
  const node = page.getByRole('button', { name: /T-14 · 示例/ });
  await node.focus();
  await page.keyboard.press('Enter');
  await expect(page.getByRole('dialog')).toContainText('不是 API 数据、真实会话或实验结论');
  await page.keyboard.press('Escape');
  await expect(node).toBeFocused();
  const vp = page.viewportSize();
  await page.evaluate(() => window.scrollTo(0, 0));
    await page.screenshot({ animations: 'disabled', path: `test-results/swarm-topology-${vp?.width}x${vp?.height}.png`, fullPage: true });
  await page.getByRole('button', { name: '列表视图' }).click();
  await expect(page.getByText('设计示意 · 非实时拓扑')).not.toBeVisible();
});

test('real timeline is explicitly unavailable and never renders sample events', async ({ page }) => {
  await page.goto('/swarm?tab=timeline');
  await expect(page.getByRole('status', { name: '事件时间线尚未接入' })).toBeVisible();
  await expect(page.getByText('14:32', { exact: true })).not.toBeVisible();
  await expect(page.getByText('Codex', { exact: true })).not.toBeVisible();
});

test('history switching from fixture to real closes selected fixture detail', async ({ page }) => {
  await page.goto('/attention');
  await page.evaluate(() => { history.pushState(null, '', '/attention?fixture=1'); window.dispatchEvent(new PopStateEvent('popstate')); });
  await page.locator('main [role="button"]:has-text("示例论文")').first().click();
  await expect(page.getByRole('dialog')).toContainText('示例数据');
  await page.goBack();
  await expect(page.getByRole('dialog')).not.toBeVisible();
  await expect(page.getByRole('alert').filter({ hasText: '示例数据模式' })).not.toBeVisible();
  await expect(page.locator('main [role="button"]:has-text("示例论文")')).not.toBeVisible();
});

test('drawer backdrop closes and restores trigger focus', async ({ page }) => {
  await page.goto('/attention?fixture=1');
  const trigger = page.locator('main [role="button"]:has-text("示例论文")').first();
  await trigger.click();
  await expect(page.getByRole('dialog')).toBeVisible();
  await page.mouse.click(10, 100);
  await expect(page.getByRole('dialog')).not.toBeVisible();
  await expect(trigger).toBeFocused();
});

for (const path of ['attention', 'workspace', 'swarm']) {
  test(`${path} has usable shell and no horizontal overflow`, async ({ page }) => {
    await page.goto(`/${path}?fixture=1`);
    await page.waitForLoadState('networkidle');
    await expect(page.getByRole('textbox', { name: path === 'workspace' ? '搜索已加载项目、目录与备注' : '搜索当前页资料、任务或 Agent' })).toBeVisible();
    await expect(page.getByRole('complementary', { name: '研究提示' })).toBeVisible();
    await expect(page.getByRole('tablist')).toBeVisible();
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
  });
}
