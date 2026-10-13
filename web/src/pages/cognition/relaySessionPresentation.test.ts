import { expect, it } from 'vitest';
import type { RelaySessionCoverage } from '../../api/relaySession';
import { relaySessionGap, relaySessionHasGap, relaySessionQueryError } from './relaySessionPresentation';

const coverage: RelaySessionCoverage = { task_id: 'cf1497b0-e092-43c0-80a9-33449bb57a40', terminal_id: 'public-terminal', worktree_id: 'public-worktree', project_id: 'public-project', status: 'complete', gap: '', lines: 3, projected: 3, duplicates: 0, unknown: 0, errors: 0, unknown_tokens: 1, remaining_bytes: 0, tail_bytes: 0, snapshot_complete: true, session_complete: false };
it('keeps unknown records, first errors and incomplete tails as explicit coverage gaps', () => {
  for (const field of ['unknown', 'errors', 'tail_bytes', 'remaining_bytes'] as const) expect(relaySessionHasGap({ ...coverage, [field]: 1 })).toBe(true);
  expect(relaySessionHasGap({ ...coverage, gap: 'first_capture_failed' })).toBe(true);
  expect(coverage.snapshot_complete).not.toBe(coverage.session_complete);
  expect(relaySessionGap('PRIVATE_RAW_CANARY /native/path')).toBe('存在未识别的覆盖缺口，保持未知');
  expect(relaySessionGap('prefix_unverified')).toContain('前缀尚未核实');
});
it('accepts only canonical task UUIDs and ordered offset-aware time filters', () => {
  expect(relaySessionQueryError({ task_id: coverage.task_id, start: '2026-10-14T00:00:00Z', end: '2026-10-15T08:00:00+08:00' })).toBe('');
  expect(relaySessionQueryError({ task_id: 'public/private/path' })).not.toBe('');
  expect(relaySessionQueryError({ task_id: '00000000-0000-0000-0000-000000000000' })).not.toBe('');
  expect(relaySessionQueryError({ start: '2026-10-14' })).not.toBe('');
  expect(relaySessionQueryError({ start: '2026-10-15T00:00:00Z', end: '2026-10-14T00:00:00Z' })).not.toBe('');
});
