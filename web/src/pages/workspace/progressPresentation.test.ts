import { expect, it } from 'vitest';
import {
  hasInference, progressPercentLabel, progressSourceSummary, progressStageLabel,
  type ProgressInference,
} from './progressPresentation';

const inference = (overrides: Partial<ProgressInference> = {}): ProgressInference => ({
  stage: '已实现论文检索入口', percent: null, source_refs: [{ path: 'tasks/S1-final-W3.md', line: 12 }],
  freshness: { observed_at: '2026-10-10T22:00:00Z', source: 'TASK/STATUS 推断' },
  ...overrides,
});

it('never fabricates a percentage for unknown progress', () => {
  expect(progressPercentLabel(null)).toBe('未知');
  expect(progressPercentLabel(undefined as unknown as null)).toBe('未知');
  expect(progressPercentLabel(0)).toBe('0%');
  expect(progressPercentLabel(62.4)).toBe('62%');
  expect(progressPercentLabel(140)).toBe('100%');
  expect(progressPercentLabel(-5)).toBe('0%');
});

it('summarises source references with line numbers when present', () => {
  expect(progressSourceSummary([])).toBe('无依据来源');
  expect(progressSourceSummary([{ path: 'tasks/S1-final-W3.md' }])).toBe('tasks/S1-final-W3.md');
  expect(progressSourceSummary([{ path: 'a.md', line: 3 }, { path: 'b.md' }])).toBe('a.md:3、b.md');
});

it('detects whether any inference is present', () => {
  expect(hasInference(inference())).toBe(true);
  expect(hasInference(inference({ stage: null, percent: null, source_refs: [] }))).toBe(false);
  expect(hasInference(inference({ stage: null, percent: 42, source_refs: [] }))).toBe(true);
});

it('labels a missing stage as unknown without guessing', () => {
  expect(progressStageLabel('已实现')).toBe('已实现');
  expect(progressStageLabel(null)).toBe('未知');
});
