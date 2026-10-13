import { describe, expect, it } from 'vitest';
import { relayProjectMatches, relayUnresolved } from './relayPresentation';
import type { RelayProject, RelayReceipt } from '../../api/relay';

const project: RelayProject = { project_id: 'test', name: 'Public test', root: 'C:/public/test', permission_level: 'C', project_settings_revision: 4, allowed_subdirs: ['README.md'], allowed_actions: ['start'], allow_agent_control: false, allow_directory: false, expand_references: true };
describe('relay projection safety', () => {
  it('matches an approved project without permitting prefix sibling escapes', () => {
    expect(relayProjectMatches('C:\\PUBLIC\\test\\child', project)).toBe(true);
    expect(relayProjectMatches('C:/public/test-secret', project)).toBe(false);
    expect(relayProjectMatches('/tmp/Public', { ...project, root: '/tmp/public' })).toBe(false);
  });
  it('unresolved send blocks send while leaving stop separately reviewable', () => {
    const receipts: RelayReceipt[] = [{ receipt_id: 'unknown', task_id: 'task-test', operation_id: '63d38b74-1655-43eb-98a7-5f1e2d9a4edb', action: 'send', operator: 'human', status: 'delivery_unknown', error: null, created_at: '2026-10-14T00:00:00Z', finished_at: null }];
    expect(relayUnresolved(receipts, 'send')).toBe(true);
    expect(relayUnresolved(receipts, 'stop')).toBe(false);
  });
});
