import { test, expect } from '@playwright/test';

/**
 * Project progress panel: reads the cached human-authored or model-inferred
 * progress and lets the human trigger one inference through the published W0
 * progress routes. The inference files are human-entered project-relative
 * paths (default STATUS.md), passed verbatim to the backend which enforces the
 * approved-directory restriction. The routes are mocked with the exact
 * ProjectProgressResultV1 shape so the UI wiring is verified without a real
 * processor run.
 */

const project = {
  id: 'proj1', name: '示例项目', root: 'C:/tmp/example', space_id: 'space1',
  settings: { revision: 1, allow_directory: false, allowed_subdirs: [], expand_references: false, allowed_actions: [], allowed_tools: [], external_model_cli: 'codex', allow_agent_control: false, history_roots: {} },
  created_at: '2026-10-10T00:00:00Z',
};

async function openProgress(page: import('@playwright/test').Page) {
  await page.goto('/workspace');
  await page.waitForLoadState('networkidle');
  await page.locator('summary').filter({ hasText: /^项目接入、权限与原生操作/ }).click();
  await page.getByRole('button', { name: /查看项目与权限/ }).click();
  const progress = page.locator('section[aria-label="项目进度推断"]');
  await expect(progress).toBeVisible();
  return progress;
}

test('reads progress evidence and triggers an inference with the default STATUS.md', async ({ page }) => {
  await page.route(/\/api\/v1\/local-projects$/, route => route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ schema_version: 1, items: [project], next_cursor: null }) }));
  await page.route(/\/api\/v1\/local-projects\/[^/]+\/progress$/, route => route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ schema_version: 1, progress: { project_id: 'proj1', status: '开发中', summary: '已完成契约对接', source: 'agent_inferred', evidence: [{ source_path: 'STATUS.md', kind: 'status', version: 'abc123', excerpt: '契约已发布' }], observed_at: '2026-10-10T00:00:00Z', revision: 1 } }) }));
  let inferred = false;
  await page.route(/\/api\/v1\/local-projects\/[^/]+\/progress\/infer$/, route => {
    inferred = true;
    const request = route.request().postDataJSON();
    expect(request.schema_version).toBe(1);
    expect(request.files).toEqual(['STATUS.md']);
    return route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ schema_version: 1, progress: { project_id: 'proj1', status: '开发中', source: 'agent_inferred', evidence: [], observed_at: '2026-10-10T00:00:00Z', revision: 1 } }) });
  });

  const progress = await openProgress(page);
  await expect(progress.getByText(/状态 · 开发中/)).toBeVisible();
  await expect(progress.getByRole('heading', { name: 'STATUS.md' })).toBeVisible();
  await expect(progress.getByText(/依据来源 · STATUS\.md:abc123/)).toBeVisible();
  await progress.getByRole('button', { name: '发起进度推断', exact: true }).click();
  await expect.poll(() => inferred).toBe(true);
});

test('human-entered paths are passed verbatim (path change, no silent bypass)', async ({ page }) => {
  await page.route(/\/api\/v1\/local-projects$/, route => route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ schema_version: 1, items: [project], next_cursor: null }) }));
  await page.route(/\/api\/v1\/local-projects\/[^/]+\/progress$/, route => route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ schema_version: 1, progress: { project_id: 'proj1', status: 'unknown', source: null, evidence: [], observed_at: null, revision: 0 } }) }));
  const sent: string[][] = [];
  await page.route(/\/api\/v1\/local-projects\/[^/]+\/progress\/infer$/, route => {
    sent.push(route.request().postDataJSON().files);
    return route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ schema_version: 1, progress: { project_id: 'proj1', status: '开发中', source: 'agent_inferred', evidence: [], observed_at: '2026-10-10T00:00:00Z', revision: 1 } }) });
  });

  const progress = await openProgress(page);
  const filesField = progress.getByLabel('推断文件（相对路径，每行一个）', { exact: true });
  await expect(filesField).toHaveValue('STATUS.md');

  // A changed, multi-line set of relative paths is sent as the human entered it.
  await filesField.fill('docs/plan.md\nnotes.md');
  await progress.getByRole('button', { name: '发起进度推断', exact: true }).click();
  await expect.poll(() => sent.length).toBe(1);
  expect(sent[0]).toEqual(['docs/plan.md', 'notes.md']);

  // An out-of-bounds-looking path is passed verbatim (not silently normalized)
  // so the backend can reject it against the approved directory.
  await filesField.fill('../secret.md');
  await progress.getByRole('button', { name: '发起进度推断', exact: true }).click();
  await expect.poll(() => sent.length).toBe(2);
  expect(sent[1]).toEqual(['../secret.md']);
});

test('an over-limit file list shows a hint and disables inference (no silent drop)', async ({ page }) => {
  await page.route(/\/api\/v1\/local-projects$/, route => route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ schema_version: 1, items: [project], next_cursor: null }) }));
  await page.route(/\/api\/v1\/local-projects\/[^/]+\/progress$/, route => route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ schema_version: 1, progress: { project_id: 'proj1', status: 'unknown', source: null, evidence: [], observed_at: null, revision: 0 } }) }));
  let inferred = 0;
  await page.route(/\/api\/v1\/local-projects\/[^/]+\/progress\/infer$/, route => {
    inferred += 1;
    return route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ schema_version: 1, progress: { project_id: 'proj1', status: '开发中', source: 'agent_inferred', evidence: [], observed_at: '2026-10-10T00:00:00Z', revision: 1 } }) });
  });

  const progress = await openProgress(page);
  const filesField = progress.getByLabel('推断文件（相对路径，每行一个）', { exact: true });
  await filesField.fill(Array.from({ length: 9 }, (_, i) => `file${i}.md`).join('\n'));
  await expect(progress.getByText(/超过 8 个/)).toBeVisible();
  await expect(progress.getByRole('button', { name: '发起进度推断', exact: true })).toBeDisabled();
  expect(inferred).toBe(0);
});

test('presents unknown progress honestly without a fabricated number', async ({ page }) => {
  await page.route(/\/api\/v1\/local-projects$/, route => route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ schema_version: 1, items: [project], next_cursor: null }) }));
  await page.route(/\/api\/v1\/local-projects\/[^/]+\/progress$/, route => route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ schema_version: 1, progress: { project_id: 'proj1', status: 'unknown', source: null, evidence: [], observed_at: null, revision: 0 } }) }));

  const progress = await openProgress(page);
  await expect(progress.getByText(/状态 · 未知/)).toBeVisible();
  await expect(progress.getByText(/尚无进度记录/)).toBeVisible();
});
