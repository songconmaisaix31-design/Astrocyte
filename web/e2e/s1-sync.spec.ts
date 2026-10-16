import { test, expect } from '@playwright/test';
import type { components } from '../src/api/schema';
import { realService, registerCurrentProject, repositoryRoot } from './realService';

test('local overview keeps real empty facts explicit at both sizes', async ({ page }) => {
  const projectsResponse = page.waitForResponse(response => response.url().endsWith('/api/v1/local-projects'));
  await page.goto('/workspace');
  await page.getByRole('button', { name: '接入项目', exact: true }).click();
  expect((await projectsResponse).status()).toBe(200);
  await expect(page.getByRole('status', { name: '暂无本地项目' })).toBeVisible();
  const panel = page.locator('section').filter({ has: page.getByRole('heading', { name: '本地 Agent 与项目', exact: true }) }).first();
  await expect(panel).toContainText('已登记项目');
  await expect(panel).toContainText('暂无本地项目');
  await expect(panel.getByRole('combobox', { name: '项目排序' }).locator('option[value="activity"]')).toHaveJSProperty('disabled', true);
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
});

test('实际登记项目可以筛选和排序，详情保留真实目录与独立默认权限', async ({ page }) => {
  test.setTimeout(120_000);
  const { server, api } = await realService();
  try {
    const project = await registerCurrentProject(api);
    await page.goto(`${server.webURL}/workspace`);
    await page.getByRole('button', { name: '接入项目', exact: true }).click();
    const panel = page.locator('section').filter({ has: page.getByRole('heading', { name: '本地 Agent 与项目', exact: true }) }).first();
    await expect(panel).toContainText('已登记项目');
    await expect(panel).toContainText('来源目录');
    await expect(panel).toContainText(repositoryRoot.replace(/[\\/]$/, ''));
    await expect(panel).toContainText('模型客户端 · 尚未许可');
    await panel.getByRole('combobox', { name: '项目排序' }).selectOption('name');
    const search = page.getByRole('textbox', { name: '搜索已加载项目、目录与备注' });
    await search.fill('不存在的项目名称');
    await expect(panel.getByRole('status', { name: '没有匹配的本地项目' })).toBeVisible();
    await search.fill(project.name);
    await panel.getByRole('button', { name: /查看项目与权限/ }).click();
    const manage = panel.getByRole('region', { name: '已登记项目管理' });
    await expect(manage).toContainText(project.root);
    await expect(manage.getByRole('button', { name: '启动获准原生会话', exact: true })).toBeDisabled();
    expect(project.settings).toMatchObject({ allow_directory: false, expand_references: false, allow_agent_control: false });
  } finally { await server.close({ preserveData: true }); }
});

test('real local inventory displays saved independent observations and refreshes without commands', async ({ page }, testInfo) => {
  const commands: string[] = [];
  page.on('request', request => { if (request.method() !== 'GET' && request.url().includes('/api/v1/')) commands.push(`${request.method()} ${request.url()}`); });
  const responsePromise = page.waitForResponse(response => response.url().endsWith('/api/v1/local-agents'));
  await page.goto('/workspace');
  const response = await responsePromise;
  expect(response.status()).toBe(200);
  const payload = await response.json() as components['schemas']['LocalAgentListV1'];
  await testInfo.attach('local-agents-response', { body: JSON.stringify(payload, null, 2), contentType: 'application/json' });
  console.log('Local inventory:', JSON.stringify({ clients: payload.items.length, installed: payload.items.filter(agent => agent.installed.status === 'available').length, configured: payload.items.filter(agent => agent.configured.status === 'available').length, startable: payload.items.filter(agent => agent.startable.status === 'available').length }));
  await page.getByRole('button', { name: '检查本机 Agent 清单', exact: true }).click();
  const panel = page.getByRole('dialog', { name: '检查本机 Agent', exact: true }).locator('section').filter({ has: page.getByRole('heading', { name: '本机 Agent 清单', exact: true }) }).first();
  await expect(panel.locator('li')).toHaveCount(payload.items.length);
  for (const agent of payload.items) {
    const card = panel.locator('li').filter({ has: page.getByRole('heading', { name: agent.display_name, exact: true }) });
    await expect(card).toContainText(agent.version ?? '未提供');
    const readinessReasons = await card.locator(':scope > dl dd > small[title]').evaluateAll(elements => elements.map(element => element.getAttribute('title')));
    for (const observation of [agent.installed, agent.configured, agent.startable]) {
      expect(readinessReasons).toContain(observation.reason);
      if (observation.status === 'unknown') await expect(card).toContainText('未知');
    }
    const native = card.locator('details').filter({ has: page.locator('summary').filter({ hasText: /^原生能力/ }) }).first();
    await native.locator(':scope > summary').click();
    const names = { discover: '发现会话', read_context: '读取上下文', start: '启动', resume: '原生接续', send: '发送输入', stop: '停止', observe: '观察状态', reconcile: '核对结果' };
    const labels = { supported: '已支持', unsupported: '不支持', unknown: '未知' };
    // 一次读取全部能力的标签、状态和原因，并逐项比较服务返回值。
    const expected = (Object.keys(names) as (keyof typeof names)[]).map(key => ({ label: names[key], status: labels[agent.capabilities[key].status], reason: agent.capabilities[key].reason }));
    await expect.poll(() => native.locator('dl > div').evaluateAll(elements => elements.map(element => ({ label: element.querySelector('dt')?.textContent, status: element.querySelector('dd > strong')?.textContent, reason: element.querySelector('dd > small[title]')?.getAttribute('title') })))).toEqual(expected);
  }
  const refresh = page.waitForResponse(response => response.url().endsWith('/api/v1/local-agents'));
  await panel.getByRole('button', { name: '刷新清单', exact: true }).click();
  expect((await refresh).status()).toBe(200);
  expect(commands).toEqual([]);
});
