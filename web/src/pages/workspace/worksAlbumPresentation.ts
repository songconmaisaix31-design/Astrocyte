/** Presentation-only works album types and pure helpers. The transport DTO is
 * supplied by the backend once W0 lands the endpoint; these mirror it so the
 * panel and its unit tests compile before the generated client exists. */

export type WorksAttitude = 'continue' | 'park' | 'done';

export interface WorksHandoff {
  content: string;
  updated_at: string | null;
  source: string;
}

export interface WorksFileFact {
  source_path: string;
  mtime: string | null;
  file_count: number;
  verified: boolean;
  reason?: string;
}

export interface WorksWork {
  id: string;
  platform: string;
  name: string;
  root: string;
  kind: string;
  source: string;
  attitude: WorksAttitude | '';
  group: string;
  hidden: boolean;
  handoff: WorksHandoff;
  files: WorksFileFact;
  updated_at: string | null;
  revision: number;
}

export interface WorksPlatformStat {
  platform: string;
  works: number;
  folders: number;
  handed_over: number;
  recent: number;
}

export interface WorksSnapshot {
  works: WorksWork[];
  platforms: WorksPlatformStat[];
  status: string;
  observed_at: string | null;
  failures: { root: string; reason: string }[];
}

export const attitudeLabels: Record<WorksAttitude, string> = {
  continue: '想继续做',
  park: '暂时搁置',
  done: '已经收尾',
};

export const attitudeOptions: { value: WorksAttitude; label: string }[] = [
  { value: 'continue', label: attitudeLabels.continue },
  { value: 'park', label: attitudeLabels.park },
  { value: 'done', label: attitudeLabels.done },
];

export const attitudeLabel = (attitude: string) => attitudeLabels[attitude as WorksAttitude] ?? '尚未标记';

export const platformLabel = (platform: string) => platform && platform !== 'unknown' ? platform : '平台未知';

export const handoffSourceLabel = (source: string) => (source === 'generated' ? 'AI 生成' : source === 'human' ? '人工编写' : '来源未知');

export interface WorksFilters {
  view: 'flat' | 'platform' | 'group' | 'timeline';
  platform: string;
  group: string;
  attitude: string;
  includeHidden: boolean;
  sort: 'name' | 'recent' | 'updated';
  search: string;
}

export const initialWorksFilters: WorksFilters = {
  view: 'flat', platform: 'all', group: 'all', attitude: 'all', includeHidden: false, sort: 'recent', search: '',
};

/** Hide is honored by default; a hidden work only reappears when the user opts in. */
export function filterWorks(items: WorksWork[], filters: WorksFilters) {
  const query = filters.search.trim().toLocaleLowerCase();
  return items
    .filter(item =>
      (filters.includeHidden || !item.hidden) &&
      (filters.platform === 'all' || item.platform === filters.platform) &&
      (filters.group === 'all' || item.group === filters.group) &&
      (filters.attitude === 'all' || item.attitude === filters.attitude) &&
      (!query || [item.name, item.platform, item.group, item.root, item.handoff.content].some(value => (value ?? '').toLocaleLowerCase().includes(query))))
    .sort((left, right) => {
      if (filters.sort === 'name') return left.name.localeCompare(right.name, 'zh-CN');
      if (filters.sort === 'updated') return (Date.parse(right.updated_at ?? '') || 0) - (Date.parse(left.updated_at ?? '') || 0);
      return (Date.parse(right.files.mtime ?? '') || 0) - (Date.parse(left.files.mtime ?? '') || 0) || left.name.localeCompare(right.name, 'zh-CN');
    });
}

/** Grouping keyed by the chosen view; flat keeps a single section. */
export function worksSections(items: WorksWork[], view: WorksFilters['view']) {
  const groups = new Map<string, WorksWork[]>();
  for (const item of items) {
    const label = view === 'platform' ? platformLabel(item.platform)
      : view === 'group' ? item.group || '未分组'
      : view === 'timeline' ? item.files.mtime?.slice(0, 10) || '活动时间未知'
      : '全部作品';
    groups.set(label, [...(groups.get(label) ?? []), item]);
  }
  return [...groups].map(([label, works]) => ({ label, works }));
}

/** The "从文件核实" badge is only honest when the scanner read the folder. */
export const fileVerifiedLabel = (fact: WorksFileFact) => fact.verified ? '从文件核实' : fact.reason || '待核实';

export function fixtureWorksSnapshot(): WorksSnapshot {
  const now = new Date().toISOString();
  return {
    works: [
      { id: 'work-a', platform: 'codex', name: '个人知识库', root: 'C:\\works\\kb', kind: 'folder', source: 'scan', attitude: 'continue', group: '研究', hidden: false, handoff: { content: '下一步：接入检索重排，先跑 M4 复算。', updated_at: now, source: 'human' }, files: { source_path: 'C:\\works\\kb', mtime: now, file_count: 42, verified: true }, updated_at: now, revision: 3 },
      { id: 'work-b', platform: 'lingxi', name: '短视频脚本生成器', root: 'C:\\works\\video-gen', kind: 'folder', source: 'json', attitude: 'park', group: '媒体', hidden: false, handoff: { content: 'prompt 保留在根目录，输出样例在 samples/。', updated_at: now, source: 'generated' }, files: { source_path: 'C:\\works\\video-gen', mtime: null, file_count: 0, verified: false, reason: '目录尚未接入' }, updated_at: now, revision: 1 },
      { id: 'work-c', platform: 'claude', name: '周报聚合脚本', root: 'C:\\works\\weekly', kind: 'folder', source: 'scan', attitude: 'done', group: '工具', hidden: true, handoff: { content: '', updated_at: null, source: 'human' }, files: { source_path: 'C:\\works\\weekly', mtime: now, file_count: 8, verified: true }, updated_at: now, revision: 2 },
    ],
    platforms: [
      { platform: 'codex', works: 1, folders: 1, handed_over: 1, recent: 1 },
      { platform: 'lingxi', works: 1, folders: 1, handed_over: 1, recent: 0 },
      { platform: 'claude', works: 1, folders: 1, handed_over: 0, recent: 1 },
    ],
    status: 'complete', observed_at: now, failures: [],
  };
}
