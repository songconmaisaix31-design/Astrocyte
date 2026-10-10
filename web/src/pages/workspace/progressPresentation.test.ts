import { expect, it } from 'vitest';
import { hasInference, progressKindLabel, progressSourceSummary, progressStatusLabel, type ProjectProgress } from './progressPresentation';

const progress = (overrides: Partial<ProjectProgress> = {}): ProjectProgress => ({
  project_id: 'p1', status: 'unknown', inferred: false,
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
    { source_path: 'tasks/S1.md', kind: 'task', version: 'abc123', freshness: '2026-10-10' },
    { source_path: 'STATUS.md', kind: 'status', version: '', freshness: '2026-10-10' },
  ])).toBe('tasks/S1.md:abc123、STATUS.md');
});

it('detects inference only from actual evidence', () => {
  expect(hasInference(progress())).toBe(false);
  expect(hasInference(progress({ inferred: true }))).toBe(true);
  expect(hasInference(progress({ summary: '有进展' }))).toBe(true);
  expect(hasInference(progress({ evidence: [{ source_path: 'tasks/S1.md', kind: 'task', version: 'v', freshness: 'now' }] }))).toBe(true);
});

it('labels evidence kinds without guessing', () => {
  expect(progressKindLabel('task')).toBe('任务文件');
  expect(progressKindLabel('status')).toBe('状态文件');
  expect(progressKindLabel('other')).toBe('other');
});
