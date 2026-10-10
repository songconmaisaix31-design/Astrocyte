import { describe, expect, it } from 'vitest';
import type { components } from '../../api/schema';
import { jobPresentation } from './jobPresentation';

const now = Date.parse('2026-10-10T06:00:00Z');
const job: components['schemas']['JobV1'] = {
  schema_version: 1, job_id: 'j1', status: 'failed', kind: 'import', version: 1,
  dedupe_key: 'source', operation_id: 'op', material_id: null, material_revision: null,
  error: { code: 'provider_unavailable', message: 'upstream stack', request_id: 'r1', retryable: true, required_action: 'configure_public_source_network' },
  attempts: 1, max_attempts: 3, created_at: '2026-10-10T05:00:00Z', updated_at: '2026-10-10T05:10:00Z', deadline_at: '2026-10-10T07:00:00Z',
  cancel_requested: false, external_started: true, delivery_unknown: false,
};

describe('persisted job recovery', () => {
  it('permits a known recoverable failure only within both limits', () => {
    expect(jobPresentation(job, now)).toMatchObject({ label: '可以重试', retryable: true });
    expect(jobPresentation({ ...job, attempts: 3 }, now)).toMatchObject({ label: '需要你处理', retryable: false });
    expect(jobPresentation(job, Date.parse(job.deadline_at)).retryable).toBe(false);
    expect(jobPresentation({ ...job, deadline_at: 'unknown' }, now).retryable).toBe(false);
  });
  it('keeps an unknown external result from showing completion or allowing resend', () => {
    for (const status of ['failed', 'succeeded', 'running'] as const) {
      expect(jobPresentation({ ...job, status, delivery_unknown: true }, now)).toMatchObject({ label: '结果待核对', retryable: false });
    }
    expect(jobPresentation({ ...job, error: { ...job.error!, code: 'delivery_unknown' } }, now).retryable).toBe(false);
  });
  it('shows actual retry progress and actionable guidance without raw diagnostics', () => {
    expect(jobPresentation({ ...job, status: 'running', attempts: 2 }, now).label).toBe('正在重试');
    expect(jobPresentation(job, now).message).toContain('修复后再明确重试');
    expect(jobPresentation(job, now).message).not.toContain('upstream stack');
  });
  it('explains saved-result storage recovery while unknown still blocks retry', () => {
    const cached = { ...job, kind: 'distillation', error: { ...job.error!, code: 'internal_error' as const, required_action: 'repair_storage_then_retry_cached_result' } };
    expect(jobPresentation(cached, now)).toMatchObject({ retryable: true, message: expect.stringContaining('不再次调用模型') });
    expect(jobPresentation({ ...cached, delivery_unknown: true }, now)).toMatchObject({ retryable: false, message: expect.stringContaining('停止自动重发') });
  });
});
