/// <reference types="node" />
import { test, expect } from '@playwright/test';
import { mkdir, readFile, writeFile, rename, unlink, stat } from 'node:fs/promises';
import { join, resolve, sep } from 'node:path';
import { randomUUID } from 'node:crypto';
import process from 'node:process';
import type { components } from '../src/api/schema';
import { startS1Server } from '../../tests/s1/server.mjs';
import { humanAPI, waitJob } from '../../tests/s1/api.mjs';
import { summarizeEnvironment } from '../../scripts/summarize.mjs';

type S = components['schemas'];
const paperExport = 'C:/Users/DW/AppData/Local/Temp/Astrocyte-S1-W2-ctx_00e6f40e6fe7/full-arxiv-import/summarize-original.json';
const videoExport = 'C:/Users/DW/orca/workspaces/Astrocyte/s1-attention-ui-1009/web/test-results/attention-video-import-sel-d52ae--an-export-or-model-request-chromium-1280/evidence/actual-video-content.json';
const paperURL = 'https://arxiv.org/html/2504.16054v1';
const videoURL = 'https://www.bilibili.com/video/BV1PReT6EEqR/';

// Explicitly assigned single live slot only. Inputs are previously captured real
// public exports, plus a fresh authorized URL extraction; no fixtures/routes.
test('real selected paper/video, scoped model rounds, ordinary reuse, authorized Agent reads and later persistence', async ({ page }, testInfo) => {
  test.skip(process.env.ASTROCYTE_TEST_REAL_S1_ACCEPTANCE !== '1', 'Requires coordinator-assigned live/model/media/native slot.');
  test.setTimeout(8_100_000);
  page.setDefaultTimeout(30_000);
  const rawPaper = await readFile(paperExport, 'utf8');
  const savedPaper = JSON.parse(rawPaper) as { extracted: { url: string; content: string } };
  expect(savedPaper.extracted.url).toBe(paperURL);
  expect(savedPaper.extracted.content.length).toBeGreaterThan(90_000);
  const savedVideo = JSON.parse(await readFile(videoExport, 'utf8')) as S['ContentV1'];
  expect(savedVideo.text.length).toBeGreaterThan(10_000);
  // HTML textarea input normalizes CRLF. Preserve the historical export itself.
  const videoInput = savedVideo.text.replaceAll('\r\n', '\n');
  const env = await summarizeEnvironment({ LOCALAPPDATA: process.env.LOCALAPPDATA ?? '', ASTROCYTE_ENABLE_SUMMARIZE: 'true', ASTROCYTE_ENABLE_CODEX_DISTILLATION: 'false' });
  const reusePath = process.env.ASTROCYTE_S1_REUSE_OWNED_TEMP;
  const ownedRoot = process.env.ASTROCYTE_S1_REUSE_APPROVED_ROOT;
  if (reusePath && !ownedRoot) throw new Error('Retained-store reuse requires the exact controller-approved owned root.');
  const server = await startS1Server({ browser: true, env, ...(reusePath ? { reuseOwnedTemporary: { path: reusePath, ownedRoot: ownedRoot! } } : {}) });
  let api = await humanAPI(server.apiURL);
  const projectRoot = join(server.temporary, 'explicit-public-material-project');
  await mkdir(projectRoot, { recursive: true });
  await writeFile(join(projectRoot, 'README.md'), 'Only explicitly selected public paper/video references are approved.\n');
  const jobs: S['JobV1'][] = [];
  const modelRecords: S['DistillationV1'][] = [];
  let completed = false;
  async function read<T>(path: string): Promise<T> { return (await api.get(path)) as T; }
  const originalUnknowns = (await read<S['JobListV1']>('/jobs')).items.filter(job => job.delivery_unknown);
  async function waitModelJob(id: string, expected: 'succeeded' | 'failed') {
    const job = await read<S['JobV1']>(`/jobs/${id}`);
    return waitJob(api, id, expected, Math.max(1000, Math.min(1_800_000, Date.parse(job.deadline_at) - Date.now()) + 1000));
  }
  async function importUI(kind: 'paper' | 'video', text?: string, refresh = false, sourceURL = kind === 'paper' ? paperURL : videoURL) {
    await page.goto(`${server.webURL}/attention`);
    await page.getByRole('main').getByRole('button', { name: '添加资料', exact: true }).click();
    const dialog = page.getByRole('dialog', { name: '添加资料' });
    await dialog.getByLabel('导入方式', { exact: true }).selectOption(text === undefined ? 'summarize_url' : 'summarize');
    if (text !== undefined) await dialog.getByLabel('既有导出资料类型', { exact: true }).selectOption(kind);
    await dialog.getByLabel(kind === 'paper' ? '论文原始来源' : text === undefined ? '视频链接' : '视频原始来源', { exact: true }).fill(sourceURL);
    await dialog.getByLabel('收藏理由（可选）').fill('人类选择的真实公开资料；核对证据缺口，不自动启动任务。');
    if (text !== undefined) await dialog.getByLabel('summarize 导出内容', { exact: true }).fill(text);
    await dialog.getByRole('checkbox', { name: '明确刷新来源，重新获取正文' }).setChecked(refresh);
    const accepted = page.waitForResponse(r => r.url().endsWith('/materials/imports') && r.request().method() === 'POST');
    await dialog.getByRole('button', { name: '导入资料', exact: true }).click();
    const response = await accepted;
    expect(response.status()).toBe(202);
    const receipt = await response.json() as S['ImportJobV1'];
    const acceptedJob = await read<S['JobV1']>(`/jobs/${receipt.job_id}`);
    const importBudget = kind === 'video' && text === undefined ? 900_000 : 30_000;
    const job = await waitJob(api, receipt.job_id, 'succeeded', Math.max(1000, Math.min(importBudget, Date.parse(acceptedJob.deadline_at) - Date.now() + 1000)));
    jobs.push(job);
    expect(job.delivery_unknown).not.toBe(true);
    expect(job.material_id).toBeTruthy();
    const detail = await read<S['MaterialDetailV1']>(`/materials/${job.material_id!}`);
    expect(detail.material.kind).toBe(kind);
    await dialog.getByRole('button', { name: '关闭', exact: true }).click();
    return { receipt, job, detail };
  }
  async function openMaterial(id: string) {
    const detail = await read<S['MaterialDetailV1']>(`/materials/${id}`);
    await page.goto(`${server.webURL}/attention`);
    await page.locator('main li[role="button"]').filter({ has: page.getByText(detail.material.title || detail.material.source_locator, { exact: true }) }).click();
    const dialog = page.getByRole('dialog');
    await expect(dialog.getByRole('button', { name: '继续沉淀', exact: true })).toBeVisible();
    return dialog;
  }
  try {
    const paper = await importUI('paper', rawPaper);
    const paperText = await read<S['ContentV1']>(`/materials/${paper.detail.material.id}/revisions/1/content`);
    expect(paperText.text).toBe(savedPaper.extracted.content);
    expect(paperText.provenance.mode).toBe('existing_json_export');
    const oldVideo = await importUI('video', videoInput);
    // A continuation reuses the genuinely completed first-run media job.
    const video = await importUI('video', undefined, !reusePath);
    expect(video.detail.material.id).toBe(oldVideo.detail.material.id);
    const originalVideo = await read<S['ContentV1']>(`/materials/${video.detail.material.id}/revisions/1/content`);
    expect(originalVideo.text).toBe(videoInput);
    const head = video.detail.material.current_revision;
    const videoText = await read<S['ContentV1']>(`/materials/${video.detail.material.id}/revisions/${head}/content`);
    expect(videoText.text.length).toBeGreaterThan(1000);
    const jobsBeforeReuse = server.query('SELECT COUNT(*) AS n FROM attention_jobs')[0].n;
    const reuse = await importUI('video');
    expect(reuse.receipt.job_id).toBe(video.receipt.job_id);
    expect(server.query('SELECT COUNT(*) AS n FROM attention_jobs')[0].n).toBe(jobsBeforeReuse);
    const oldReceipt = await importUI('video', videoInput);
    expect(oldReceipt.receipt.job_id).toBe(oldVideo.receipt.job_id);
    expect(oldReceipt.detail.material.current_revision).toBe(head);
    expect((await read<S['ContentV1']>(`/materials/${video.detail.material.id}/revisions/${head}/content`)).text).toBe(videoText.text);

    // Root-approved genuine historical v1/v2 exports: import-only, never model input.
    const versionObjects = 'C:/Users/DW/AppData/Local/Temp/astrocyte-s1-OHU8NF/data/objects';
    const rawV1 = await readFile(join(versionObjects, '159560bf4f2e4f2f82d4a1a03facd615d8a923345864d6cf30514a8249d237c1'), 'utf8');
    const rawV2 = await readFile(join(versionObjects, '75a089219c3dbeeb5ccef83ee8166ec63c4f956d9ff6e34fb8386f624584aae8'), 'utf8');
    const v1Body = (JSON.parse(rawV1) as { extracted: { content: string } }).extracted.content;
    const v2Body = (JSON.parse(rawV2) as { extracted: { content: string } }).extracted.content;
    expect(v1Body).not.toBe(v2Body);
    const versionA = await importUI('paper', rawV1, false, 'https://arxiv.org/html/2501.12948v1');
    const versionB = await importUI('paper', rawV2, false, 'https://arxiv.org/html/2501.12948v2');
    expect(versionB.detail.material.id).toBe(versionA.detail.material.id);
    expect(versionB.detail.material.current_revision).toBe(2);
    const reusedA = await importUI('paper', rawV1, false, 'https://arxiv.org/html/2501.12948v1');
    expect(reusedA.receipt.job_id).toBe(versionA.receipt.job_id);
    expect(reusedA.detail.material.current_revision).toBe(2);
    expect((await read<S['ContentV1']>(`/materials/${versionA.detail.material.id}/revisions/1/content`)).text).toBe(v1Body);
    expect((await read<S['ContentV1']>(`/materials/${versionA.detail.material.id}/revisions/2/content`)).text).toBe(v2Body);

    await page.goto(`${server.webURL}/workspace`);
    const panel = page.locator('section').filter({ has: page.getByRole('heading', { name: '本地 Agent 与项目', exact: true }) }).first();
    let registeredProject = reusePath ? (await read<S['LocalProjectListV1']>('/local-projects')).items.find(item => resolve(item.root) === resolve(projectRoot)) : undefined;
    if (!registeredProject) {
      await panel.getByText('登记项目 / 发现子项目', { exact: true }).click();
      await panel.getByLabel('新项目顶层空间名称', { exact: true }).fill('真实公开资料范围');
      await panel.getByRole('button', { name: '创建空间用于此项目', exact: true }).click();
      await expect(panel.getByLabel('关联项目顶层空间', { exact: true })).not.toHaveValue('');
      await panel.getByLabel('项目名称', { exact: true }).fill('真实公开资料项目');
      await panel.getByLabel('项目绝对目录', { exact: true }).fill(projectRoot);
      const registering = page.waitForResponse(r => r.url().endsWith('/local-projects') && r.request().method() === 'POST');
      await panel.getByRole('button', { name: '登记此项目', exact: true }).click();
      const registered = await registering;
      expect(registered.status()).toBe(200);
      registeredProject = (await registered.json() as S['LocalProjectResultV1']).project;
    }
    const project = registeredProject;
    const spaceID = project.space_id;
    if (reusePath) await panel.getByRole('button', { name: /真实公开资料项目.*查看项目与权限/ }).click();
    const manage = panel.getByRole('region', { name: '已登记项目管理' });
    if (project.settings.external_model_cli !== 'codex') {
      await manage.getByText('项目小权限与模型处理许可', { exact: true }).click();
      await manage.getByLabel('此项目允许模型处理的 CLI', { exact: true }).selectOption('codex');
      for (const action of ['启动', '停止']) await manage.getByRole('checkbox', { name: action, exact: true }).check();
      const saving = page.waitForResponse(r => r.url().endsWith(`/local-projects/${project.id}/settings`) && r.request().method() === 'PUT');
      await manage.getByRole('button', { name: '保存此项目许可', exact: true }).click();
      expect((await saving).status()).toBe(200);
    }
    const native = manage.getByRole('region', { name: '项目原生 Agent 操作' });
    await native.getByLabel('项目操作客户端', { exact: true }).selectOption('codex');
    const probing = page.waitForResponse(r => r.url().endsWith('/sessions/probe') && r.request().method() === 'POST');
    await native.getByRole('button', { name: '明确验证原生连接（启动后停止）', exact: true }).click();
    const probed = await probing;
    expect(probed.status()).toBe(200);
    expect((await probed.json() as S['NativeSessionResultV1']).session.stop_confirmed).toBe(true);

    // Human approval explicitly pins both actual objects, independently of model approval.
    await page.goto(`${server.webURL}/attention`);
    await page.getByLabel('当前项目顶层空间', { exact: true }).selectOption(spaceID);
    const refs = page.getByRole('region', { name: '空间资料引用', exact: true });
    const approvedRefs = (await read<S['ProjectSpaceResultV1']>(`/project-spaces/${spaceID}`)).space.material_refs;
    for (const [material, revision] of [[paper.detail.material, 1], [video.detail.material, head]] as const) {
      if (!approvedRefs.some(ref => ref.material_id === material.id && ref.revision === revision)) {
        await refs.getByLabel('@ 资料', { exact: true }).selectOption(material.id);
        await refs.getByLabel('@ 固定内容版本', { exact: true }).selectOption(String(revision));
        const reference = page.waitForResponse(r => r.url().endsWith(`/project-spaces/${spaceID}/references`) && r.request().method() === 'POST');
        await refs.getByRole('button', { name: '@ 引用文件', exact: true }).click();
        expect((await reference).status()).toBe(200);
      }
      await expect(refs.getByRole('button', { name: `移除引用 · ${material.title || material.id}`, exact: true })).toBeVisible();
    }
    async function modelRound(materialID: string, stage: 'content' | 'topic' | 'project', question: string, prior: string[] = [], recoverPublication = false) {
      const dialog = await openMaterial(materialID);
      await dialog.getByRole('button', { name: '继续沉淀', exact: true }).click();
      const automatic = dialog.getByRole('region', { name: '自动沉淀', exact: true });
      await automatic.getByLabel('自动沉淀的项目模型许可', { exact: true }).selectOption(project.id);
      await automatic.getByLabel('自动沉淀层次', { exact: true }).selectOption(stage);
      await automatic.getByLabel('自动沉淀本轮问题', { exact: true }).fill(question);
      for (const id of prior) await automatic.getByRole('checkbox', { name: new RegExp(id) }).check();
      const queued = page.waitForResponse(r => r.url().endsWith('/distillations/jobs') && r.request().method() === 'POST');
      await automatic.getByRole('button', { name: '提交自动沉淀', exact: true }).click();
      const response = await queued;
      expect(response.status()).toBe(202);
      const receipt = await response.json() as S['ImportJobV1'];
      const queuedJob = await read<S['JobV1']>(`/jobs/${receipt.job_id}`);
      expect(queuedJob.delivery_unknown, 'An earlier unknown operation must never be replayed').not.toBe(true);
      if (recoverPublication && queuedJob.status !== 'succeeded') {
        await expect.poll(async () => (await read<S['JobV1']>(`/jobs/${receipt.job_id}`)).external_started, { timeout: 30_000, intervals: [50] }).toBe(true);
        // Only this helper's explicitly owned temporary paths may be moved.
        const objectRoot = resolve(server.dataDir, 'objects');
        const retainedRoot = resolve(server.temporary, 'public-objects-retained');
        for (const path of [objectRoot, retainedRoot]) expect(path.startsWith(resolve(server.temporary) + sep)).toBe(true);
        expect((await stat(objectRoot)).isDirectory()).toBe(true);
        let moved = false;
        let blockedRoot = false;
        try {
          await rename(objectRoot, retainedRoot); moved = true;
          await writeFile(objectRoot, 'Owned temporary publication failure; restored before retry.', { flag: 'wx' }); blockedRoot = true;
          const failed = await waitModelJob(receipt.job_id, 'failed');
          jobs.push(failed);
          expect(failed.delivery_unknown).toBe(false);
          const payload = JSON.parse(String(server.query('SELECT payload FROM attention_jobs WHERE id=?', failed.job_id)[0].payload)) as { Result?: unknown; result?: unknown };
          const savedResult = payload.Result ?? payload.result;
          expect(savedResult, 'The actual paid result must be durable before retry').toBeTruthy();
          await testInfo.attach('actual-local-publication-failure', { body: JSON.stringify({ job: failed, savedResult }, null, 2), contentType: 'application/json' });
        } finally {
          if (blockedRoot) await unlink(objectRoot);
          if (moved) await rename(retainedRoot, objectRoot);
        }
        await server.restart();
        api = await humanAPI(server.apiURL);
        await page.goto(`${server.webURL}/attention`);
        await page.getByRole('button', { name: '刷新队列', exact: true }).click();
        const row = page.locator('li').filter({ has: page.locator('strong').filter({ hasText: receipt.job_id }) });
        await expect(row).toContainText('可以重试');
        const nativeBefore = server.query('SELECT COUNT(*) AS n FROM local_agent_sessions')[0].n;
        const startBefore = (await read<S['LocalAgentListV1']>('/local-agents')).items.find(item => item.id === 'codex')!.capabilities.start.checked_at;
        const resultBefore = String(server.query('SELECT payload FROM attention_jobs WHERE id=?', receipt.job_id)[0].payload);
        const retrying = page.waitForResponse(r => r.url().endsWith(`/jobs/${receipt.job_id}/retry`) && r.request().method() === 'POST');
        await row.getByRole('button', { name: '重试作业', exact: true }).click();
        expect((await retrying).status()).toBe(200);
        await waitJob(api, receipt.job_id, 'succeeded', 30_000);
        expect(server.query('SELECT COUNT(*) AS n FROM local_agent_sessions')[0].n).toBe(nativeBefore);
        expect((await read<S['LocalAgentListV1']>('/local-agents')).items.find(item => item.id === 'codex')!.capabilities.start.checked_at).toBe(startBefore);
        expect(String(server.query('SELECT payload FROM attention_jobs WHERE id=?', receipt.job_id)[0].payload)).toBe(resultBefore);
      }
      const job = await waitModelJob(receipt.job_id, 'succeeded');
      jobs.push(job);
      expect(job.delivery_unknown).not.toBe(true);
      if (recoverPublication) expect(job.attempts).toBe(2);
      const detail = await read<S['MaterialDetailV1']>(`/materials/${materialID}`);
      expect(detail.distillations.filter(item => item.question === question && item.stage === stage)).toHaveLength(1);
      const record = detail.distillations.find(item => item.question === question && item.stage === stage);
      expect(record, 'Actual successful job must have its persisted model record').toBeTruthy();
      expect(record!.provenance.mode).toBe('selected_project_fixed_text');
      expect(record!.provenance.processor).toBe('codex');
      expect(record!.output_text.length).toBeGreaterThan(20);
      expect(record!.prior_distillation_ids).toEqual([...prior].sort());
      modelRecords.push(record!);
      // A new request identity with the same fixed input/question reuses the receipt.
      if (!recoverPublication) {
        const body = response.request().postDataJSON() as Record<string, unknown>;
        const key = randomUUID();
        const repeated = await fetch(response.url(), { method: 'POST', headers: { Cookie: api.cookie, Origin: server.webURL, 'Content-Type': 'application/json', 'X-CSRF-Token': api.session.csrf_token, 'Idempotency-Key': key }, body: JSON.stringify({ ...body, request_id: key }) });
        expect(repeated.status).toBe(202);
        expect((await repeated.json() as S['ImportJobV1']).job_id).toBe(receipt.job_id);
        await dialog.getByRole('button', { name: '关闭', exact: true }).click();
      }
      return record!;
    }
    // A separately authorized new objective; never replay the earlier UNKNOWN.
    const contentQuestion = '梳理论文关键术语及正文定义依据，缺失定义标待查，不重做旧核心方法/局限作业。仅依据本次完整固定正文，不调用工具，全部文字限制1000中文字以内。';
    const content = await modelRound(paper.detail.material.id, 'content', contentQuestion);
    const videoContent = await modelRound(video.detail.material.id, 'content', '仅依据所选视频实际字幕提炼内容与尚待验证的主张，不补造时间或成果；全部文字限制1000中文字以内。');
    const paperDialog = await openMaterial(paper.detail.material.id);
    await paperDialog.getByRole('button', { name: '继续沉淀', exact: true }).click();
    const manual = paperDialog.locator('form').filter({ has: paperDialog.getByRole('heading', { name: '继续沉淀 · 人工记录', exact: true }) });
    await manual.getByLabel('沉淀层次', { exact: true }).selectOption('topic');
    await manual.getByLabel('本轮问题', { exact: true }).fill('人工选入两份公开资料，二者技术关联待核对');
    await manual.getByLabel('人工整理结果', { exact: true }).fill('将这两份实际公开资料并列作为待比较输入；是否有共同技术问题仍待查，无已验证结论。');
    await manual.getByLabel('添加关联资料', { exact: true }).selectOption(video.detail.material.id);
    await manual.getByLabel('待查问题（每行一项）', { exact: true }).fill('两份资料的核心问题是否可比较，哪些依据尚缺？');
    const bridging = page.waitForResponse(r => r.url().endsWith('/distillations') && r.request().method() === 'POST');
    await manual.getByRole('button', { name: '保存人工沉淀', exact: true }).click();
    const bridgeResponse = await bridging;
    expect(bridgeResponse.status()).toBe(201);
    const bridge = (await bridgeResponse.json() as S['DistillationResultV1']).distillation;
    expect(bridge.provenance.mode).toBe('manual');
    await paperDialog.getByRole('button', { name: '关闭', exact: true }).click();
    const topic = await modelRound(paper.detail.material.id, 'topic', '仅依据当前固定论文正文和本轮术语/定义记录，延续整理主题及明确待查问题；旧未知作业不作为前轮，其他资料未交付，不推断关联或研究成果；全部文字1000中文字以内。', [content.id]);
    const videoTopic = await modelRound(video.detail.material.id, 'topic', '延续当前所选视频固定正文和前轮内容，整理主题与待查问题；其他资料未交付，不编造跨资料关联或研究成果。不要启动任务，全部文字1000中文字以内。', [videoContent.id], true);
    expect(videoTopic.input_refs.every(ref => ref.material_id === video.detail.material.id && ref.revision === head)).toBe(true);
    const candidateDialog = await openMaterial(paper.detail.material.id);
    await candidateDialog.getByRole('button', { name: '形成候选', exact: true }).click();
    const candidateForm = candidateDialog.locator('form').filter({ has: candidateDialog.getByRole('heading', { name: '形成候选', exact: true }) });
    await candidateForm.getByLabel('候选标题', { exact: true }).fill('核对真实论文与视频的证据边界');
    await candidateForm.getByLabel('候选用途 / 为什么值得做', { exact: true }).fill('人工审核两份已选公开资料和实际模型输出，保留未确认的关联。');
    await candidateForm.getByLabel('最小下一步', { exact: true }).fill('人工逐项核对原文与前轮待查问题，暂不启动任务。');
    for (const id of [bridge.id, topic.id, videoTopic.id]) await candidateForm.getByRole('checkbox', { name: new RegExp(id) }).check();
    const forming = page.waitForResponse(r => r.url().endsWith('/opportunities') && r.request().method() === 'POST');
    await candidateForm.getByRole('button', { name: '保存候选', exact: true }).click();
    const formed = await forming;
    expect(formed.status()).toBe(201);
    const candidate = (await formed.json() as S['OpportunityDetailV1']).opportunity;
    expect(candidate.state).toBe('ready_for_review');
    const later = page.getByRole('dialog');
    await later.getByLabel('反馈类型', { exact: true }).selectOption('later');
    await later.getByLabel('反馈理由', { exact: true }).fill('以后再核对，保留公开原文与固定版本，不表示拒绝。');
    const reviewing = page.waitForResponse(r => r.url().endsWith(`/opportunities/${candidate.id}/reviews`) && r.request().method() === 'POST');
    await later.getByRole('button', { name: '保存人工反馈', exact: true }).click();
    expect((await reviewing).status()).toBe(201);
    const reviewed = await read<S['OpportunityDetailV1']>(`/opportunities/${candidate.id}`);
    expect(reviewed.opportunity.state).toBe('deferred');
    expect(reviewed.reviews.some(review => review.feedback === 'reject')).toBe(false);
    await later.getByRole('button', { name: '关闭', exact: true }).click();

    // Actual human grant then token-authenticated Agent reads, with no token artifacts.
    await page.goto(`${server.webURL}/workspace`);
    await panel.getByRole('button', { name: /真实公开资料项目.*查看项目与权限/ }).click();
    await manage.getByText('由你授权的项目 Agent 身份', { exact: true }).click();
    await manage.getByLabel('被授权 Agent 身份', { exact: true }).fill('w3-public-scope-reader');
    await manage.getByRole('checkbox', { name: '读取已纳入上下文', exact: true }).check();
    const granting = page.waitForResponse(r => r.url().endsWith(`/local-projects/${project.id}/grants`) && r.request().method() === 'POST');
    await manage.getByRole('button', { name: '保存此身份的动作授权', exact: true }).click();
    expect((await granting).status()).toBe(200);
    const tokenKey = randomUUID();
    const issued = await fetch(`${server.apiURL}/api/v1/local-projects/${project.id}/agent-token`, { method: 'POST', headers: { Cookie: api.cookie, Origin: server.apiURL, 'Content-Type': 'application/json', 'X-CSRF-Token': api.session.csrf_token, 'Idempotency-Key': tokenKey }, body: JSON.stringify({ schema_version: 1, request_id: tokenKey, expected_version: 2, agent_id: 'w3-public-scope-reader' }) });
    expect(issued.status).toBe(200);
    const credential = (await issued.json() as S['ProjectAgentTokenResultV1']).credential;
    const baseline = await Promise.all([paper.detail.material.id, video.detail.material.id].map(id => read<S['MaterialDetailV1']>(`/materials/${id}`)));
    for (let round = 0; round < 3; round++) {
      const key = randomUUID();
      const response = await fetch(`${server.apiURL}/api/v1/local-projects/${project.id}/context`, { method: 'POST', headers: { Authorization: `Bearer ${credential.token}`, 'Content-Type': 'application/json', 'Idempotency-Key': key }, body: JSON.stringify({ schema_version: 1, request_id: key, expected_version: 2, references: [{ material_id: paper.detail.material.id, revision: 1 }, { material_id: video.detail.material.id, revision: head }], files: [] }) });
      expect(response.status).toBe(200);
      const packet = await response.json() as S['LocalContextPacketV1'];
      expect(packet.materials.find(item => item.reference.material_id === paper.detail.material.id)?.text).toBe(paperText.text);
      expect(packet.materials.find(item => item.reference.material_id === video.detail.material.id)?.text).toBe(videoText.text);
    }
    const afterReads = await Promise.all(baseline.map(detail => read<S['MaterialDetailV1']>(`/materials/${detail.material.id}`)));
    afterReads.forEach((detail, index) => {
      expect(detail.material.human_usage_count).toBe(baseline[index].material.human_usage_count);
      expect(detail.uses.filter(use => use.actor_kind === 'human')).toEqual(baseline[index].uses.filter(use => use.actor_kind === 'human'));
      expect(detail.material.agent_usage_count).toBe(baseline[index].material.agent_usage_count! + 3);
      expect(detail.material.current_revision).toBe(baseline[index].material.current_revision);
    });
    const agentKey = randomUUID();
    const selfApproval = await fetch(`${server.apiURL}/api/v1/local-projects/${project.id}/grants`, { method: 'POST', headers: { Authorization: `Bearer ${credential.token}`, 'Content-Type': 'application/json', 'Idempotency-Key': agentKey }, body: JSON.stringify({ schema_version: 1, request_id: agentKey, expected_version: 2, agent_id: 'w3-public-scope-reader', actions: ['start'] }) });
    expect(selfApproval.status).toBe(403);
    const revoking = page.waitForResponse(r => r.url().endsWith(`/local-projects/${project.id}/grants/revoke`) && r.request().method() === 'POST');
    await manage.getByRole('button', { name: '撤销此项目身份许可', exact: true }).click();
    expect((await revoking).status()).toBe(200);
    await server.restart();
    api = await humanAPI(server.apiURL);
    expect(await read<S['OpportunityDetailV1']>(`/opportunities/${candidate.id}`)).toEqual(reviewed);
    for (const material of [paper.detail.material, video.detail.material]) expect((await read<S['MaterialDetailV1']>(`/materials/${material.id}`)).material.lifecycle).toBe('active');
    expect((await read<S['ContentV1']>(`/materials/${video.detail.material.id}/revisions/1/content`)).text).toBe(videoInput);
    expect((await read<S['MissionListV1']>('/missions')).items).toEqual([]);
    for (const original of originalUnknowns) expect(await read<S['JobV1']>(`/jobs/${original.job_id}`)).toEqual(original);
    await page.goto(`${server.webURL}/attention`);
    for (const width of [1280, 1920]) {
      await page.setViewportSize({ width, height: width === 1280 ? 720 : 1080 });
      expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
      await page.screenshot({ path: testInfo.outputPath(`actual-selected-materials-${width}.png`), fullPage: true });
    }
    await testInfo.attach('actual-public-scope-results', { body: JSON.stringify({ paperOriginal: paperExport, videoOriginal: videoExport, paper: afterReads[0], video: afterReads[1], modelRecords, candidate: reviewed, originalUnknowns, videoStoredRevisionChanged: head > 1, actualExtraPaperRevisions: [versionA.detail.material.id, 1, 2], limits: ['Paper reused genuine existing JSON export; no fresh arXiv URL retrieval.', 'Video export textarea normalizes CRLF; stored revision change does not establish an upstream video-content change.'] }, null, 2), contentType: 'application/json' });
    completed = true;
  } finally {
    await testInfo.attach('actual-job-states', { body: JSON.stringify(jobs, null, 2), contentType: 'application/json' });
    await testInfo.attach('actual-retained-store-path', { body: JSON.stringify({ temporary: server.temporary, dataDir: server.dataDir, completed }, null, 2), contentType: 'application/json' });
    await server.close({ preserveData: true });
  }
});
