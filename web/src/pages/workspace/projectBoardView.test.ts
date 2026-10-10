import { describe, expect, it } from 'vitest';
import { boardSections, filterBoardProjects, initialBoardFilters, type BoardProject } from './projectBoardPresentation';

const base: BoardProject = { id: 'same-repo', name: '共同项目', folders: ['C:/project', 'C:/worktree'], clients: ['codex', 'claude'], sources: ['orca_registered', 'codex_index'], activityAt: '2026-10-10T10:00:00Z', activitySource: 'native session metadata', activityStatus: 'recorded', intent: 'continuing', group: '研究', notes: '下一步检查结果', review: '', archived: false };
describe('source-backed project board presentation', () => {
  it('keeps a multi-client project once in every layout and preserves complete provenance', () => {
    for (const view of ['flat', 'platform', 'group', 'timeline']) {
      const rendered = boardSections([base], view).flatMap(section => section.projects);
      expect(rendered).toEqual([base]);
      expect(rendered[0].clients).toEqual(['codex', 'claude']);
      expect(rendered[0].folders).toHaveLength(2);
    }
  });
  it('does not convert real activity into human intent or unknown completion', () => {
    const unknown = { ...base, id: 'unknown', name: '无活动', activityAt: null, activityStatus: 'unknown', intent: 'unknown' };
    const finished = { ...base, id: 'finished', intent: 'finished' };
    const items = [unknown, finished, base];
    expect(filterBoardProjects(items, initialBoardFilters).map(item => item.id)).toEqual(['finished', 'same-repo', 'unknown']);
    expect(filterBoardProjects(items, { ...initialBoardFilters, intent: 'continuing' })).toEqual([base]);
    expect(filterBoardProjects(items, { ...initialBoardFilters, status: 'unknown' })).toEqual([unknown]);
    expect(unknown.intent).toBe('unknown');
  });
  it('filters sources, clients, group, notes and reversible archives without mutating records', () => {
    const archived = { ...base, id: 'archived', archived: true };
    const items = [base, archived];
    expect(filterBoardProjects(items, { ...initialBoardFilters, client: 'claude', source: 'codex_index', group: '研究', search: '检查' })).toEqual([base]);
    expect(filterBoardProjects(items, { ...initialBoardFilters, archive: 'archived' })).toEqual([archived]);
    expect(filterBoardProjects(items, { ...initialBoardFilters, archive: 'all' })).toHaveLength(2);
    expect(items).toEqual([base, archived]);
  });
});
