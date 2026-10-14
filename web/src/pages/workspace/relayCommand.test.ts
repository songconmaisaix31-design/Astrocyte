import { expect, it } from 'vitest';
import { ApiError } from '../../api/client';
import { relayCommandError, relayCommandUncertain } from './relayCommand';

it('shows the sanitized HTTP conflict action and request without dropping its task ID', () => {
  const id = 'ae48c13e-86e4-4899-8c20-a618873ea209';
  const error = new ApiError(409, { schema_version: 1, error: { code: 'version_conflict', message: 'Relay state or workspace conflict', required_action: `view_existing_task:${id}`, request_id: 'public-request', retryable: false } });
  expect(relayCommandError(error)).toContain(`view_existing_task:${id}`);
  expect(relayCommandError(error)).toContain('public-request');
});

it('never displays arbitrary transport exception messages or provider paths', () => {
  expect(relayCommandError(new Error('PRIVATE_PROVIDER_CANARY /private/raw'))).toBe('接力请求结果未知，请只读取任务与收据核对。');
  expect(relayCommandError('PRIVATE_PROVIDER_CANARY')).not.toContain('PRIVATE_PROVIDER_CANARY');
});

it('503即使编码为provider_unavailable也保留未知；明确权限或未实现拒绝不冒充交付', () => {
  const error = (status: number, code: 'provider_unavailable' | 'scope_denied' | 'unsupported_capability') => new ApiError(status, { schema_version: 1, error: { code, message: '明确参数输入', required_action: 'inspect_local_relay_provider', request_id: 'input-id', retryable: false } });
  expect(relayCommandUncertain(error(503, 'provider_unavailable'))).toBe(true);
  expect(relayCommandUncertain(error(403, 'scope_denied'))).toBe(false);
  expect(relayCommandUncertain(error(501, 'unsupported_capability'))).toBe(false);
  expect(relayCommandUncertain(new TypeError('input transport failure'))).toBe(true);
});
