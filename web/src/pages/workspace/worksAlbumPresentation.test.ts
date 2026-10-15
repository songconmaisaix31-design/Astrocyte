import { describe, expect, it } from 'vitest';
import { attitudeLabel, fileVerifiedLabel, filterWorks, handoffSourceLabel, initialWorksFilters, worksSections, type WorksWork } from './worksAlbumPresentation';

// 以下参数输入仅验证过滤和分组函数，不构成目录或执行观察。
const base: WorksWork = { id: 'input', platform: 'codex', name: '输入', root: '输入路径', kind: 'folder', source: 'scan', attitude: 'continue', group: '研究', hidden: false, handoff: { content: '', updated_at: null, source: 'human' }, files: { source_path: '输入路径', mtime: null, file_count: 0, verified: true }, updated_at: '2026-01-01T00:00:00Z', revision: 1 };
const snapshot = { works: [
  { ...base, id: 'work-a', name: '个人知识库' },
  { ...base, id: 'work-b', name: '短视频脚本生成器', platform: 'lingxi', attitude: 'park' as const, group: '媒体', files: { ...base.files, verified: false, reason: '目录尚未接入' } },
  { ...base, id: 'work-c', name: '周报聚合脚本', platform: 'claude', hidden: true, group: '工具' },
] };

describe('works album presentation keeps backend facts', () => {
  it('maps the three attitudes to Chinese labels', () => {
    expect(attitudeLabel('continue')).toBe('想继续做');
    expect(attitudeLabel('park')).toBe('暂时搁置');
    expect(attitudeLabel('done')).toBe('已经收尾');
    expect(attitudeLabel('')).toBe('尚未标记');
  });
  it('labels the file-verified badge honestly', () => {
    expect(fileVerifiedLabel(snapshot.works[0].files)).toBe('从文件核实');
    expect(fileVerifiedLabel(snapshot.works[1].files)).toBe('目录尚未接入');
  });
  it('labels the handoff source', () => {
    expect(handoffSourceLabel('generated')).toBe('AI 生成');
    expect(handoffSourceLabel('human')).toBe('人工编写');
  });
  it('hides hidden works by default and opts them back in', () => {
    const visible = filterWorks(snapshot.works, initialWorksFilters);
    expect(visible.map(work => work.id)).not.toContain('work-c');
    const withHidden = filterWorks(snapshot.works, { ...initialWorksFilters, includeHidden: true });
    expect(withHidden.map(work => work.id)).toContain('work-c');
  });
  it('filters by platform, attitude and search', () => {
    expect(filterWorks(snapshot.works, { ...initialWorksFilters, platform: 'codex' }).map(work => work.id)).toEqual(['work-a']);
    expect(filterWorks(snapshot.works, { ...initialWorksFilters, attitude: 'park' }).map(work => work.id)).toEqual(['work-b']);
    expect(filterWorks(snapshot.works, { ...initialWorksFilters, includeHidden: true, search: '周报' }).map(work => work.id)).toEqual(['work-c']);
  });
  it('sorts by name', () => {
    const sorted = filterWorks(snapshot.works, { ...initialWorksFilters, includeHidden: true, sort: 'name' });
    expect(sorted.map(work => work.name)).toEqual(['短视频脚本生成器', '个人知识库', '周报聚合脚本']);
  });
  it('sections by platform, group and timeline', () => {
    const byPlatform = worksSections(snapshot.works, 'platform');
    expect(byPlatform.length).toBe(3);
    const byGroup = worksSections(snapshot.works, 'group');
    expect(byGroup.map(section => section.label).sort()).toEqual(['媒体', '工具', '研究']);
    const flat = worksSections(snapshot.works, 'flat');
    expect(flat).toHaveLength(1);
    expect(flat[0].label).toBe('全部作品');
  });
});
