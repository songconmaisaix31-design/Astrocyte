/// <reference types="node" />
import { execFileSync } from 'node:child_process';
import { readFile } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';
import { resolve } from 'node:path';
import { expect } from '@playwright/test';
import { startS1Server } from '../../tests/s1/server.mjs';
import { humanAPI, importMaterial } from '../../tests/s1/api.mjs';
import { loadPublicPaperInputs, loadPublicPaperSnapshots } from '../../tests/s1/public-inputs.mjs';
import type { components } from '../src/api/schema';

export const repositoryRoot = fileURLToPath(new URL('../../', import.meta.url));
type Schemas = components['schemas'];

export async function realService(relay = false, env?: { ASTROCYTE_REPOSITORY_ENVIRONMENT_ID: string }) {
  const server = await startS1Server({ browser: true, relay, env });
  const api = await humanAPI(server.apiURL);
  return { server, api };
}

/** Works命令使用同一操作身份写入正文和请求头。 */
export function writeWorks(api: Awaited<ReturnType<typeof humanAPI>>, path: string, body: Record<string, unknown>, options: { method?: string } = {}) {
  const key = crypto.randomUUID();
  return api.write(path, { ...body, request_id: key }, { ...options, key });
}

/** 来源是本轨已跟踪文件与精确 Git 提交，人工导入不调用提取器或模型。 */
export async function importTrackedDocument(api: Awaited<ReturnType<typeof humanAPI>>, filename = 'AGENTS.md') {
  const head = execFileSync('git', ['rev-parse', 'HEAD'], { cwd: repositoryRoot, encoding: 'utf8' }).trim();
  const text = execFileSync('git', ['show', `${head}:${filename}`], { cwd: repositoryRoot, encoding: 'utf8' });
  return importMaterial(api, {
    kind: 'text', adapter: 'manual', source_key: '', content_digest: '',
    source_locator: `https://github.com/songconmaisaix31-design/Astrocyte/blob/${head}/${filename}`,
    export_text: text, title: filename, collection_reason: null, source_spans: ['whole document'],
  });
}

function publicPaperDirectory() {
  const directory = process.env.ASTROCYTE_TEST_PUBLIC_PAPER_INPUTS_DIR;
  if (!directory) throw new Error('需要B交付实际公开HTML缓存的 ASTROCYTE_TEST_PUBLIC_PAPER_INPUTS_DIR');
  if (!/[\\/]\.verification[\\/]|[\\/]\.verification$/.test(directory)) throw new Error('公开HTML输入必须来自明确获准的验证目录');
  return directory;
}

/** 实际公开HTML解析复用生产paperDOM及安装的第三方库。 */
export function publicPaperInputs() { return loadPublicPaperInputs(publicPaperDirectory()); }
export function publicPaperSnapshots() { return loadPublicPaperSnapshots(publicPaperDirectory()); }

/** 两个实际arXiv版本正文通过生产手工导入，来源和版本完整保留。 */
export async function importPublicPaperVersions(api: Awaited<ReturnType<typeof humanAPI>>) {
  const inputs = await publicPaperInputs();
  const first = await importMaterial(api, inputs.original.import, { timeoutMs: 120_000 });
  const second = await importMaterial(api, inputs.revised.import, { timeoutMs: 120_000 });
  expect(second.detail.material.id).toBe(first.detail.material.id);
  expect(second.detail.material.current_revision).toBe(2);
  expect((await api.get(`/materials/${first.detail.material.id}/revisions/1/content`)).text).toBe(inputs.original.text);
  expect((await api.get(`/materials/${second.detail.material.id}/revisions/2/content`)).text).toBe(inputs.revised.text);
  return { first, second };
}

export async function registerCurrentProject(api: Awaited<ReturnType<typeof humanAPI>>) {
  const space = await api.write('/project-spaces', { title: '本轨实际验证空间' }, { status: 201 });
  const result = await api.write('/local-projects', {
    root: repositoryRoot, name: 'standards-frontend', space_id: space.space.id,
  });
  expect(result.project.root.toLowerCase()).toBe(resolve(repositoryRoot).toLowerCase().replace(/[\\/]$/, ''));
  return result.project;
}

export async function realProgress(api: Awaited<ReturnType<typeof humanAPI>>, id: string) {
  const result = await api.get(`/local-projects/${id}/progress`) as Schemas['ProjectProgressResultV1'];
  return result.progress;
}

/** 只保存验证表单所需许可；测试不提交可读取成功的推断命令。 */
export async function permitProgressForm(api: Awaited<ReturnType<typeof humanAPI>>, project: Schemas['LocalProjectV1']) {
  await api.write(`/local-projects/${project.id}/settings`, {
    expected_revision: project.settings.revision,
    settings: { allow_directory: true, allowed_subdirs: ['STATUS.md'], external_model_cli: 'codex', allowed_actions: [] },
  }, { method: 'PUT' });
}

export async function currentStatus() {
  return readFile(resolve(repositoryRoot, 'STATUS.md'), 'utf8');
}

/** 人工摘录实际协议并记录本轨待查问题；不生成模型执行事实。 */
export async function trackedCandidate(api: Awaited<ReturnType<typeof humanAPI>>) {
  const imported = await importTrackedDocument(api);
  const material = imported.detail.material;
  const refs = [{ material_id: material.id, revision: 1, locator: material.source_locator, span: 'whole document' }];
  const content = await api.get(`/materials/${material.id}/revisions/1/content`);
  const record = await api.write('/distillations', {
    input_refs: refs, stage: 'topic', output_text: content.text.split('\n').slice(0, 8).join('\n'),
    question: '人工核对本轨协议', processing_config: 'manual:v1',
    pending_questions: ['本轨真实 HTTP 和浏览器验证是否通过？'],
  }) as Schemas['DistillationResultV1'];
  const unknown = { value: null, reason: '本轨尚未测量该维度' };
  const result = await api.write('/opportunities', {
    title: '前端标准整改核查', evidence_refs: refs, goal_refs: [],
    dimensions: { goal_progress: unknown, current_interest: unknown, project_improvement: unknown, originality: unknown },
    next_step: '运行本轨实际 HTTP 和浏览器验证', purpose: '保留本轨执行协议与待查问题',
    missing_evidence: ['真实验证完成结果'], distillation_ids: [record.distillation.id],
  }, { status: 201 }) as Schemas['OpportunityDetailV1'];
  return result.opportunity;
}

/** 读取仓库已跟踪的真实 summarize 公共文档导出，保留来源和正文。 */
export async function capturedPublicDocument() {
  const raw = await readFile(resolve(repositoryRoot, 'internal/adapters/importers/testdata/summarize-arxiv-manual.json'), 'utf8');
  const result = JSON.parse(raw) as { extracted: { url: string; title: string; content: string } };
  if (!result.extracted.url || !result.extracted.content || !result.extracted.title) throw new Error('实际公共文档导出缺少来源或正文');
  return { raw, url: result.extracted.url, title: result.extracted.title, text: result.extracted.content };
}
