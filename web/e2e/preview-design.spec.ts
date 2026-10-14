import { test, expect } from '@playwright/test';
import { realService, importTrackedDocument, trackedCandidate, registerCurrentProject } from './realService';
import type { components } from '../src/api/schema';

test.describe('实际资料、项目与协作边界', () => {
  let owned: Awaited<ReturnType<typeof realService>>;
  test.beforeAll(async () => {
    test.setTimeout(120_000);
    owned = await realService();
    await trackedCandidate(owned.api);
    await importTrackedDocument(owned.api, 'README.md');
    await registerCurrentProject(owned.api);
  });
  test.afterAll(async () => { if (owned) await owned.server.close({ preserveData: true }); });
  const go = (path: string) => `${owned.server.webURL}/${path}`;

  test('搜索快捷键过滤实际集合，Escape 清除搜索', async ({ page }) => {
    await page.goto(go('attention'));
    await page.keyboard.press('/');
    const search = page.getByRole('textbox', { name: '搜索当前页资料、任务或 Agent' });
    await expect(search).toBeFocused();
    await search.fill('AGENTS.md');
    await expect(page.locator('main [role="button"]').filter({ hasText: 'AGENTS.md' })).toBeVisible();
    await expect(page.locator('main [role="button"]').filter({ hasText: 'README.md' })).not.toBeVisible();
    await search.fill('没有匹配的随机词');
    await expect(page.getByRole('status', { name: '没有匹配的素材' })).toBeVisible();
    await page.keyboard.press('Escape');
    await expect(search).toHaveValue('');
    await expect(page.locator('main [role="button"]').filter({ hasText: 'README.md' })).toBeVisible();
  });
  test('类型筛选保留真实资料详情和来源定位', async ({ page }) => {
    await page.goto(go('attention'));
    await page.getByRole('group', { name: '素材类型' }).getByRole('button', { name: '文本', exact: true }).click();
    await page.locator('main [role="button"]').filter({ hasText: 'AGENTS.md' }).click();
    const dialog = page.getByRole('dialog');
    await expect(dialog).toContainText('来源定位');
    await expect(dialog).toContainText('人工使用');
    await expect(dialog).toContainText('收藏理由');
    await expect(dialog).toContainText('未提供');
  });
  test('历史 fixture 参数读取真实集合且查看不发送命令', async ({ page }) => {
    const writes: string[] = [];
    page.on('request', request => { if (request.method() !== 'GET' && request.url().includes('/api/v1/')) writes.push(request.url()); });
    const response = page.waitForResponse(value => value.url().endsWith('/materials'));
    await page.goto(go('attention?fixture=1'));
    expect((await response).status()).toBe(200);
    await page.locator('main [role="button"]').filter({ hasText: 'AGENTS.md' }).click();
    await expect(page.getByRole('dialog')).not.toContainText('示例数据');
    expect(writes).toEqual([]);
  });
  test('实际未测量候选维度和缺失收藏理由保持未知', async ({ page }) => {
    await page.goto(go('attention'));
    await expect(page.locator('main [role="button"]').filter({ hasText: '前端标准整改核查' })).toContainText('未知');
    await page.locator('main [role="button"]').filter({ hasText: 'AGENTS.md' }).click();
    await expect(page.getByRole('dialog')).toContainText('收藏理由');
    await expect(page.getByRole('dialog')).toContainText('未提供');
  });
  test('Arrow、Home、End 和浏览器历史保持实际页签状态', async ({ page }) => {
    await page.goto(go('attention'));
    const overview = page.getByRole('tab', { name: '资料与线索' });
    await overview.focus();
    await page.keyboard.press('ArrowRight');
    const opportunities = page.getByRole('tab', { name: '研究机会' });
    await expect(opportunities).toBeFocused();
    await expect(opportunities).toHaveAttribute('aria-selected', 'true');
    await expect(page.getByRole('heading', { name: /^素材/ })).not.toBeVisible();
    await page.keyboard.press('End');
    await expect(page.getByRole('status', { name: '暂无收藏', exact: true })).toBeVisible();
    await page.keyboard.press('Home');
    await expect(overview).toBeFocused();
    await page.goBack();
    await expect(page.getByRole('tab', { name: '我的收藏' })).toHaveAttribute('aria-selected', 'true');
  });
  test('页面快捷键与侧栏保持真实协作来源', async ({ page }) => {
    await page.goto(go('attention'));
    await page.keyboard.press('2');
    await expect(page).toHaveURL(/\/workspace/);
    await page.getByRole('navigation', { name: '主导航' }).getByRole('button', { name: '蜂群空间' }).click();
    await expect(page).toHaveURL(/\/swarm/);
    await page.getByRole('tab', { name: '任务时间线', exact: true }).click();
    await expect(page.getByRole('button', { name: '拓扑视图' })).toBeDisabled();
    await expect(page.getByText('T-11 · 示例', { exact: true })).not.toBeVisible();
  });
  test('实际未连接会话与未提交提案显示空集合', async ({ page }) => {
    await page.goto(go('workspace?tab=sessions'));
    await expect(page.getByRole('status', { name: '暂无会话', exact: true })).toBeVisible();
    await page.getByRole('tab', { name: '提案与批准' }).click();
    await expect(page.getByRole('status', { name: '暂无提案', exact: true })).toBeVisible();
    expect((await owned.api.get('/sessions') as components['schemas']['SessionListV1']).items).toHaveLength(0);
  });
  test('缺少协作事件时保持禁用拓扑和完整任务快照说明', async ({ page }) => {
    await page.goto(go('swarm?tab=timeline'));
    await expect(page.getByRole('button', { name: '拓扑视图' })).toBeDisabled();
    await expect(page.getByText('拓扑尚未接入；下面显示真实任务与工作项快照。')).toBeVisible();
    await expect(page.getByText('设计示意 · 非实时拓扑')).not.toBeVisible();
  });
  test('真实时间线接口缺口保持可见且不生成动态', async ({ page }) => {
    await page.goto(go('swarm?tab=timeline'));
    await expect(page.getByRole('status', { name: '事件时间线尚未接入' })).toBeVisible();
    await expect(page.getByText('14:32', { exact: true })).not.toBeVisible();
    await expect(page.getByText('Codex', { exact: true })).not.toBeVisible();
  });
  test('切换实际页面关闭详情，浏览器返回后重新读取实际来源', async ({ page }) => {
    await page.goto(go('attention'));
    await page.locator('main [role="button"]').filter({ hasText: 'AGENTS.md' }).click();
    await expect(page.getByRole('dialog')).toBeVisible();
    await page.evaluate(() => { history.pushState(null, '', '/workspace'); window.dispatchEvent(new PopStateEvent('popstate')); });
    await expect(page.getByRole('dialog')).not.toBeVisible();
    await page.goBack();
    await expect(page.locator('main [role="button"]').filter({ hasText: 'AGENTS.md' })).toBeVisible();
    await expect(page.getByRole('dialog')).not.toBeVisible();
  });
  test('抽屉背景关闭详情并恢复真实卡片焦点', async ({ page }) => {
    await page.goto(go('attention'));
    const trigger = page.locator('main [role="button"]').filter({ hasText: 'AGENTS.md' });
    await trigger.click();
    await expect(page.getByRole('dialog')).toBeVisible();
    await page.mouse.click(10, 100);
    await expect(page.getByRole('dialog')).not.toBeVisible();
    await expect(trigger).toBeFocused();
  });
  for (const path of ['attention', 'workspace', 'swarm']) {
    test(`${path} 的实际页面外壳没有水平溢出`, async ({ page }) => {
      await page.goto(go(path));
      await expect(page.getByRole('textbox', { name: path === 'workspace' ? '搜索已加载项目、目录与备注' : '搜索当前页资料、任务或 Agent' })).toBeVisible();
      await page.locator('summary').filter({ hasText: /^研究提示与概览$/ }).click();
      await expect(page.getByRole('complementary', { name: '研究提示' })).toBeVisible();
      await expect(page.getByRole('tablist')).toBeVisible();
      expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
    });
  }
});
