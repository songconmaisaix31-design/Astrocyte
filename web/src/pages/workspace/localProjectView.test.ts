import { describe, expect, it } from 'vitest';
import { filterLocalProjects, localProjectViews } from './localProjectView';
import type { components } from '../../api/schema';

const projects: components['schemas']['ProjectV1'][] = [
  { id: 'a', name: 'Alpha', root_path: '/authorized/alpha', environment_id: 'local' },
  { id: 'b', name: 'Beta', root_path: '/authorized/beta', environment_id: 'local' },
];
const sessions: components['schemas']['SessionV1'][] = [
  { id: 's1', adapter: 'codex', project_id: 'a', native_session_id: null, binding_status: 'observed', context_state: 'unknown', capabilities: { native_resume: null, handoff: null } },
  { id: 's2', adapter: 'codex', project_id: 'a', native_session_id: null, binding_status: 'bound', context_state: 'unknown', capabilities: { native_resume: false, handoff: null } },
  { id: 's3', adapter: 'pi', project_id: 'outside', native_session_id: null, binding_status: 'bound', context_state: 'unknown', capabilities: { native_resume: true, handoff: true } },
];

describe('project overview preserves backend facts', () => {
  it('associates only sessions for the project and deduplicates platforms', () => {
    const views = localProjectViews(projects, sessions);
    expect(views[0].platforms).toEqual(['codex']);
    expect(views[0].bindingStates).toEqual(['observed', 'bound']);
    expect(views[1].platforms).toEqual([]);
    expect(views.every(view => view.lastActivity === null && view.fileCount === null && view.handoff === null && view.group === null)).toBe(true);
  });
  it('combines actual platform and binding filters without granting capability', () => {
    const views = localProjectViews(projects, sessions);
    expect(filterLocalProjects(views, 'codex', 'observed', 'all', 'source').map(view => view.project.id)).toEqual(['a']);
    expect(filterLocalProjects(views, 'pi', 'bound', 'all', 'source')).toEqual([]);
    expect(filterLocalProjects(views, 'unknown', 'unknown', 'unknown', 'source').map(view => view.project.id)).toEqual(['b']);
  });
  it('sorts actual timestamps with unknown and invalid last, preserving input order', () => {
    const views = localProjectViews(projects, sessions);
    const timed = [{ ...views[0], lastActivity: 'invalid' }, { ...views[1], lastActivity: '2026-10-10T03:00:00Z' }, { ...views[0], lastActivity: null }, { ...views[1], lastActivity: '2026-10-09T03:00:00Z' }];
    expect(filterLocalProjects(timed, 'all', 'all', 'all', 'activity').map(view => view.lastActivity)).toEqual(['2026-10-10T03:00:00Z', '2026-10-09T03:00:00Z', 'invalid', null]);
    expect(timed[0].lastActivity).toBe('invalid');
  });
});
