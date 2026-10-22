import { fileURLToPath } from 'node:url';
import { test, expect } from '@playwright/test';
import { startS1Server } from '../../tests/s1/server.mjs';
import { humanAPI } from '../../tests/s1/api.mjs';

test('真实服务未配置原生运行时时显示错误并阻止创建和附着', async ({ page }) => {
  const server = await startS1Server({ browser: true });
  try {
    const human = await humanAPI(server.apiURL);
    const { space } = await human.write('/project-spaces', { title: 'SWA 浏览器边界核查' }, { status: 201 });
    const { project } = await human.write('/local-projects', { root: fileURLToPath(new URL('../../', import.meta.url)), name: 'SWA 实际公开项目', space_id: space.id });
    const commands: { project: string; action: string }[] = [];
    page.on('request', request => {
      if (request.method() === 'POST' && request.url().includes('/orca/commands')) commands.push({ project: new URL(request.url()).pathname, action: request.postDataJSON().action });
    });
    await page.goto(`${server.webURL}/swarm`);
    // SectionCard 使用语义 section；通过标题定位所属容器，不修改生产可访问性来迁就测试。
    const workspace = page.locator('section').filter({ has: page.getByRole('heading', { name: 'SWA 执行工作台', exact: true }) }).first();
    await workspace.getByLabel('SWA 项目').selectOption(project.id);
    await expect(workspace.getByRole('alert').first()).toContainText('provider_unavailable');
    await expect(workspace.getByRole('button', { name: '连接已有任务', exact: true })).toBeDisabled();
    await workspace.getByText('新建 Controller 与固定任务', { exact: true }).click();
    await expect(workspace.getByRole('button', { name: '创建 Controller', exact: true })).toBeDisabled();
    expect(commands).toEqual([{ project: `/api/v1/local-projects/${project.id}/orca/commands`, action: 'status' }]);
    expect(server.query('SELECT COUNT(*) AS count FROM orca_execution_links')[0].count).toBe(0);
    await page.reload();
    await expect(workspace.getByLabel('SWA 项目')).toHaveValue('');
    expect(commands).toHaveLength(1);
    expect(server.query('SELECT COUNT(*) AS count FROM orca_execution_links')[0].count).toBe(0);
  } finally { await server.close(); }
});
