import { test, expect } from '@playwright/test';

/**
 * Project progress panel: reads the cached human-authored or model-inferred
 * progress and lets the human trigger one inference through the published W0
 * progress routes. The routes are mocked with the exact ProjectProgressResultV1
 * shape so the UI wiring (GET + POST infer with operation_id + files) is
 * verified without a real processor run.
 */

const project = {
  id: 'proj1', name: '示例项目', root: 'C:/tmp/example', space_id: 'space1',
  settings: { revision: 1, allow_directory: false, allowed_subdirs: [], expand_references: false, allowed_actions: [], allowed_tools: [], external_model_cli: 'codex', allow_agent_control: false, history_roots: {} },
  created_at: '2026-10-10T00:00:00Z',
};

test('reads progress evidence and triggers an explicit inference', async ({ page }) => {
  await page.route(/\/api\/v1\/local-projects$/, route => route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ schema_version: 1, items: [project], next_cursor: null }) }));
  await page.route(/\/api\/v1\/local-projects\/[^/]+\/progress$/, route => route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ schema_version: 1, progress: { project_id: 'proj1', status: '开发中', summary: '已完成契约对接', source: 'agent_inferred', evidence: [{ source_path: 'tasks/S1.md', kind: 'task', version: 'abc123', excerpt: '契约已发布' }], observed_at: '2026-10-10T00:00:00Z', revision: 1 } }) }));
  let inferred = false;
  await page.route(/\/api\/v1\/local-projects\/[^/]+\/progress\/infer$/, route => {
    inferred = true;
    const request = route.request().postDataJSON();
    expect(request.schema_version).toBe(1);
    expect(request.files).toEqual(['TASK.md', 'STATUS.md']);
    expect(request.operation_id).toMatch(/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/);
    return route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ schema_version: 1, progress: { project_id: 'proj1', status: '开发中', source: 'agent_inferred', evidence: [], observed_at: '2026-10-10T00:00:00Z', revision: 1 } }) });
  });

  await page.goto('/workspace');
  await page.waitForLoadState('networkidle');
  await page.locator('summary').filter({ hasText: /^项目接入、权限与原生操作/ }).click();
  await page.getByRole('button', { name: /查看项目与权限/ }).click();
  const progress = page.locator('section[aria-label="项目进度推断"]');
  await expect(progress).toBeVisible();
  await expect(progress.getByText(/状态 · 开发中/)).toBeVisible();
  await expect(progress.getByRole('heading', { name: 'tasks/S1.md' })).toBeVisible();
  await expect(progress.getByText(/依据来源 · tasks\/S1\.md:abc123/)).toBeVisible();
  await progress.getByRole('button', { name: '发起进度推断', exact: true }).click();
  await expect.poll(() => inferred).toBe(true);
});

test('presents unknown progress honestly without a fabricated number', async ({ page }) => {
  await page.route(/\/api\/v1\/local-projects$/, route => route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ schema_version: 1, items: [project], next_cursor: null }) }));
  await page.route(/\/api\/v1\/local-projects\/[^/]+\/progress$/, route => route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ schema_version: 1, progress: { project_id: 'proj1', status: 'unknown', source: null, evidence: [], observed_at: null, revision: 0 } }) }));

  await page.goto('/workspace');
  await page.waitForLoadState('networkidle');
  await page.locator('summary').filter({ hasText: /^项目接入、权限与原生操作/ }).click();
  await page.getByRole('button', { name: /查看项目与权限/ }).click();
  const progress = page.locator('section[aria-label="项目进度推断"]');
  await expect(progress).toBeVisible();
  await expect(progress.getByText(/状态 · 未知/)).toBeVisible();
  await expect(progress.getByText(/尚无进度记录/)).toBeVisible();
});
