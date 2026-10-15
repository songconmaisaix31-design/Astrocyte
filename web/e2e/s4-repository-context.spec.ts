import { test, expect, type Locator } from '@playwright/test';
import { realService, registerCurrentProject } from './realService';
import type { components } from '../src/api/schema';

async function fillQuery(panel: Locator) {
  await panel.getByLabel('查询用途', { exact: true }).fill('核查当前公开仓库的仓库上下文能力和真实权限');
  await panel.getByLabel('查询项目相对路径', { exact: true }).fill('web/src/api\nweb/src/components');
  await panel.getByLabel('最多读取文件数', { exact: true }).fill('32');
  await panel.getByLabel('最多读取字节数', { exact: true }).fill('65536');
  await panel.getByLabel('最多返回结果数', { exact: true }).fill('10');
  await panel.getByLabel('关系展开深度', { exact: true }).fill('2');
}

test('真实未配置仓库上下文保留503、显式查询参数和人工权限恢复，维护保持禁用', async ({ page, context }) => {
  test.setTimeout(180_000);
  const { server, api } = await realService();
  try {
    const project = await registerCurrentProject(api);
    const path = `/api/v1/local-projects/${project.id}/repository-context`;
    const requests: string[] = [];
    page.on('request', request => { if (request.url().includes(path)) requests.push(`${request.method()} ${new URL(request.url()).pathname}`); });
    await page.goto(`${server.webURL}/workspace`);
    const panel = page.getByRole('region', { name: '仓库上下文', exact: true });
    await panel.getByLabel('仓库上下文项目', { exact: true }).selectOption(project.id);
    await fillQuery(panel);
    const maintenance = panel.getByRole('group', { name: '明确维护当前项目索引', exact: true });
    for (const button of await maintenance.getByRole('button').all()) await expect(button).toBeDisabled();
    expect(requests).toHaveLength(0);

    const mapResponse = page.waitForResponse(response => response.url().includes(`${path}/map?`) && response.request().method() === 'GET');
    await panel.getByRole('button', { name: '读取仓库上下文', exact: true }).focus();
    await page.keyboard.press('Enter');
    const map = await mapResponse;
    expect(map.status()).toBe(503);
    const error = await map.json() as components['schemas']['ErrorV1'];
    expect(error.error.code).toBe('provider_unavailable');
    expect(error.error.required_action).toBe('configure_the_verified_local_repository_adapter');
    expect(new URL(map.url()).searchParams.getAll('root')).toEqual(['web/src/api', 'web/src/components']);
    await expect(panel.getByRole('alert')).toContainText('暂未完成');
    await panel.getByRole('button', { name: '代码搜索', exact: true }).focus();
    await page.keyboard.press('Enter');
    await expect(panel.getByRole('button', { name: '代码搜索', exact: true })).toHaveAttribute('aria-pressed', 'true');
    await panel.getByLabel('代码搜索内容', { exact: true }).fill('QueryState');
    const searchResponse = page.waitForResponse(response => response.url().endsWith(`${path}/search`) && response.request().method() === 'POST');
    await panel.getByRole('button', { name: '读取仓库上下文', exact: true }).click();
    const search = await searchResponse;
    expect(search.status()).toBe(503);
    expect(search.request().postDataJSON()).toEqual({ schema_version: 1, purpose: '核查当前公开仓库的仓库上下文能力和真实权限', roots: ['web/src/api', 'web/src/components'], budget: { max_files: 32, max_bytes: 65536, max_results: 10, max_depth: 2 }, query: 'QueryState' });
    await expect(panel.getByRole('alert')).toContainText('暂未完成');
    await expect(panel.getByLabel('代码搜索内容', { exact: true })).toHaveValue('QueryState');
    for (const button of await maintenance.getByRole('button').all()) await expect(button).toBeDisabled();

    await context.setExtraHTTPHeaders({ Authorization: 'Bearer invalid' });
    const deniedResponse = page.waitForResponse(response => response.url().endsWith('/api/v1/local-projects') && response.request().method() === 'GET');
    await panel.getByRole('button', { name: '刷新项目与读取权限', exact: true }).click();
    expect((await deniedResponse).status()).toBe(403);
    await expect(panel.getByRole('alert').filter({ hasText: '数据可能已过期' })).toBeVisible();
    await expect(panel.getByLabel('仓库上下文项目', { exact: true })).toHaveValue(project.id);
    await expect(panel.getByLabel('查询用途', { exact: true })).toBeDisabled();
    await expect(panel.getByRole('button', { name: '读取仓库上下文', exact: true })).toBeDisabled();
    await context.setExtraHTTPHeaders({});
    const restoredResponse = page.waitForResponse(response => response.url().endsWith('/api/v1/local-projects') && response.request().method() === 'GET');
    await panel.getByRole('button', { name: '刷新项目与读取权限', exact: true }).click();
    expect((await restoredResponse).status()).toBe(200);
    await expect(panel.getByRole('alert').filter({ hasText: '数据可能已过期' })).not.toBeVisible();
    await expect(panel.getByLabel('查询用途', { exact: true })).toBeEnabled();
    await expect(panel.getByLabel('代码搜索内容', { exact: true })).toHaveValue('QueryState');
    expect(requests).toEqual([`GET ${path}/map`, `POST ${path}/search`]);
    expect(server.query('SELECT COUNT(*) AS count FROM repository_context_indexes')).toEqual([{ count: 0 }]);
    expect(server.query('SELECT COUNT(*) AS count FROM local_agent_sessions')).toEqual([{ count: 0 }]);
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
  } finally {
    await context.setExtraHTTPHeaders({});
    await server.close({ preserveData: true });
  }
});
