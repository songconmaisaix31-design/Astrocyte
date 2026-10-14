/// <reference types="node" />
import { writeFile } from 'node:fs/promises';
import { join } from 'node:path';
import { expect, type Page } from '@playwright/test';
import { startS1Server } from '../../tests/s1/server.mjs';
import { humanAPI } from '../../tests/s1/api.mjs';
import { registerCurrentProject, repositoryRoot } from './realService';
import type { components } from '../src/api/schema';

type S = components['schemas'];
/** 注册实际安装的Codex，仅保存意图或拒绝场景；此环境没有模型与原生启动许可。 */
export async function relayIntentService({ nestedRoot = false } = {}) {
  const server = await startS1Server({ browser: true, relay: true, env: async ({ temporary }) => {
    await writeFile(join(temporary, 'local-appdata/astrocyte-workbench/v0.2-access/agents.json'), JSON.stringify([{ name: 'codex', launch_cmd: ['codex'], capabilities: {}, drive_mode: 'structured' }]));
    return {};
  } });
  const api = await humanAPI(server.apiURL);
  let project: S['LocalProjectV1'];
  if (nestedRoot) {
    const { space } = await api.write('/project-spaces', { title: '本轨嵌套根目录权限核查' }, { status: 201 }) as S['ProjectSpaceResultV1'];
    const result = await api.write('/local-projects', { root: join(repositoryRoot, 'web'), name: '本轨web子目录', space_id: space.id }) as S['LocalProjectResultV1'];
    project = result.project;
  } else project = await registerCurrentProject(api);
  await expect.poll(async () => (await api.get('/local-projects/registered') as S['RegisteredProjectDiscoveryResultV1']).snapshot.projects.length).toBeGreaterThan(0);
  const { snapshot } = await api.write('/works/refresh', { roots: [repositoryRoot] }) as S['WorksSnapshotResultV1'];
  expect(snapshot.works).toHaveLength(1);
  const work = snapshot.works[0];
  await api.write(`/works/${work.id}/handoff`, { content: '本轨前端真实HTTP与资料权限核查；此环境未授权原生启动。' }, { method: 'PUT' });
  return { server, api, project, work };
}

export async function declareIntent(page: Page, owned: Awaited<ReturnType<typeof relayIntentService>>) {
  await page.goto(`${owned.server.webURL}/workspace/works`);
  const album = page.getByRole('region', { name: '作品寻回册', exact: true });
  await album.getByLabel(`${owned.work.name} 的操作`, { exact: true }).getByRole('button', { name: '接力', exact: true }).click();
  const dialog = page.getByRole('dialog');
  await dialog.getByLabel('获准项目', { exact: true }).selectOption(owned.project.id);
  await dialog.getByLabel('偏好 Agent', { exact: true }).selectOption('codex');
  await dialog.getByLabel('任务描述', { exact: true }).fill('核对本轨固定公开协议的资料权限，不启动原生会话。');
  await dialog.getByRole('button', { name: '保存意图并查看权限', exact: true }).click();
  await expect(dialog.getByRole('region', { name: '权限确认页', exact: true })).toBeVisible();
  return dialog;
}
