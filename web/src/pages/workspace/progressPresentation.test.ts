import { expect, it } from 'vitest';
import { hasProgress, progressKindLabel, progressSourceLabel, progressSourceSummary, progressStatusLabel, type ProjectProgress } from './progressPresentation';

const progress = (overrides: Partial<ProjectProgress> = {}): ProjectProgress => ({
  project_id: 'p1', status: 'unknown', evidence: [], revision: 0,
  ...overrides,
});

it('presents unknown/absent status as 未知, never a fabricated number', () => {
  expect(progressStatusLabel('unknown')).toBe('未知');
  expect(progressStatusLabel(null)).toBe('未知');
  expect(progressStatusLabel(undefined)).toBe('未知');
  expect(progressStatusLabel('开发中')).toBe('开发中');
});

it('summarizes evidence source files with version', () => {
  expect(progressSourceSummary(undefined)).toBe('无依据来源');
  expect(progressSourceSummary([])).toBe('无依据来源');
  expect(progressSourceSummary([
    { source_path: 'tasks/S1.md', kind: 'task', version: 'abc123' },
    { source_path: 'STATUS.md', kind: 'status', version: '' },
  ])).toBe('tasks/S1.md:abc123、STATUS.md');
});

it('detects recorded progress only from actual evidence', () => {
  expect(hasProgress(progress())).toBe(false);
  expect(hasProgress(progress({ source: 'agent_inferred' }))).toBe(true);
  expect(hasProgress(progress({ source: 'human' }))).toBe(true);
  expect(hasProgress(progress({ summary: '有进展' }))).toBe(true);
  expect(hasProgress(progress({ percent: 42 }))).toBe(true);
  expect(hasProgress(progress({ evidence: [{ source_path: 'tasks/S1.md', kind: 'task', version: 'v' }] }))).toBe(true);
});

it('labels evidence source without guessing', () => {
  expect(progressSourceLabel('human')).toBe('人工记录');
  expect(progressSourceLabel('agent_inferred')).toBe('Agent 推断');
  expect(progressSourceLabel(null)).toBe('尚未记录');
  expect(progressSourceLabel(undefined)).toBe('尚未记录');
});

it('labels evidence kinds without guessing', () => {
  expect(progressKindLabel('task')).toBe('任务文件');
  expect(progressKindLabel('status')).toBe('状态文件');
  expect(progressKindLabel('project_file')).toBe('项目文件');
  expect(progressKindLabel('other')).toBe('other');
});
