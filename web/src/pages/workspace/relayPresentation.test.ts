import { describe, expect, it } from 'vitest';
import { relayProjectMatches, relayUnresolved, relayExistingTaskID, relaySingleInput } from './relayPresentation';
import type { RelayProject, RelayReceipt, RelayTaskView, RelayAgent } from '../../api/relay';

const project: RelayProject = { project_id: 'test', name: 'Public test', root: 'C:/public/test', permission_level: 'C', project_settings_revision: 4, allowed_subdirs: ['README.md'], allowed_actions: ['start'], allow_agent_control: false, allow_directory: false, expand_references: true };
describe('relay projection safety', () => {
  it('uses explicit EOF registration only for the actual terminal drive, preserving native structured send', () => {
    const terminal: NonNullable<RelayTaskView['terminal']> = { terminal_id: 'public-terminal', worktree_id: 'public-worktree', agent: 'public-cli', drive_mode: 'terminal', liveness: 'unverifiable', last_event_at: null, exited_at: null, last_sequence: 0, created_at: '2026-10-14T00:00:00Z' };
    const agent: RelayAgent = { name: 'public-cli', launch_cmd: ['public-cli'], capabilities: {}, drive_mode: 'terminal', input_mode: 'stdin_eof' };
    expect(relaySingleInput(terminal, agent)).toBe(true);
    expect(relaySingleInput({ ...terminal, drive_mode: 'structured' }, agent)).toBe(false);
    expect(relaySingleInput(terminal, { ...agent, input_mode: 'line' })).toBe(false);
    expect(relaySingleInput(terminal, { ...agent, input_mode: undefined })).toBe(false);
    expect(relaySingleInput(null, agent)).toBe(false);
    expect(relaySingleInput(terminal, undefined)).toBe(false);
  });
  it('accepts only the canonical nonnil existing-task conflict action, never message text', () => {
    const id = 'ae48c13e-86e4-4899-8c20-a618873ea209';
    const error: NonNullable<RelayReceipt['error']> = { code: 'version_conflict', message: 'Relay state or workspace conflict', required_action: `view_existing_task:${id}`, request_id: 'public-test', retryable: false };
    expect(relayExistingTaskID(error)).toBe(id);
    for (const action of ['refresh_relay_state', `view_existing_task:${id}/PRIVATE_PROVIDER_CANARY`, `view_existing_task:${id.toUpperCase()}`, 'view_existing_task:00000000-0000-0000-0000-000000000000', 'view_existing_task:PRIVATE_PROVIDER_CANARY']) {
      expect(relayExistingTaskID({ ...error, required_action: action })).toBe('');
    }
    expect(relayExistingTaskID({ ...error, code: 'context_stale' })).toBe('');
    expect(relayExistingTaskID({ ...error, code: 'delivery_unknown' })).toBe('');
    expect(relayExistingTaskID({ ...error, required_action: 'refresh_relay_state', message: `view_existing_task:${id}` })).toBe('');
    expect(relayExistingTaskID(null)).toBe('');
  });
  it('matches an approved project without permitting prefix sibling escapes', () => {
    expect(relayProjectMatches('C:\\PUBLIC\\test', project)).toBe(true);
    expect(relayProjectMatches('C:\\PUBLIC\\test\\child', project)).toBe(false);
    expect(relayProjectMatches('C:/public/test-secret', project)).toBe(false);
    expect(relayProjectMatches('/tmp/Public', { ...project, root: '/tmp/public' })).toBe(false);
  });
  it('unresolved send blocks send while leaving stop separately reviewable', () => {
    const receipts: RelayReceipt[] = [{ receipt_id: 'unknown', task_id: 'task-test', operation_id: '63d38b74-1655-43eb-98a7-5f1e2d9a4edb', action: 'send', operator: 'human', status: 'delivery_unknown', error: null, created_at: '2026-10-14T00:00:00Z', finished_at: null }];
    expect(relayUnresolved(receipts, 'send')).toBe(true);
    expect(relayUnresolved(receipts, 'stop')).toBe(false);
  });
  it('prepare的待定、未知或带错误接受收据持续阻止重新初始化', () => {
    const receipt: RelayReceipt = { receipt_id: 'input-receipt', task_id: 'input-task', operation_id: '63d38b74-1655-43eb-98a7-5f1e2d9a4edb', action: 'prepare', operator: 'human', status: 'accepted', error: null, created_at: '2026-10-14T00:00:00Z', finished_at: null };
    expect(relayUnresolved([receipt], 'prepare')).toBe(false);
    expect(relayUnresolved([{ ...receipt, status: 'pending' }], 'prepare')).toBe(true);
    expect(relayUnresolved([{ ...receipt, status: 'delivery_unknown' }], 'prepare')).toBe(true);
    expect(relayUnresolved([{ ...receipt, error: { code: 'delivery_unknown', message: '明确参数输入', request_id: receipt.operation_id, retryable: false, required_action: 'inspect_original_receipt_do_not_resend' } }], 'prepare')).toBe(true);
    expect(relayUnresolved([{ ...receipt, status: 'rejected' }], 'prepare')).toBe(false);
    expect(relayUnresolved([{ ...receipt, status: 'pending' }], 'dispatch')).toBe(false);
  });
});
