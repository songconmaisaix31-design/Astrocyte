/** Presentation only; HTTP identity and aggregation are supplied by the backend. */
export interface BoardProject {
  id: string;
  name: string;
  folders: string[];
  clients: string[];
  sources: string[];
  activityAt: string | null;
  activitySource: string;
  activityStatus: string;
  intent: string;
  group: string;
  notes: string;
  review: string;
  archived: boolean;
}

export const intentLabels: Record<string, string> = { continuing: '想继续做', on_hold: '暂时搁置', finished: '已经收尾', unknown: '尚未标记' };
export interface BoardFilters { client: string; intent: string; group: string; source: string; status: string; archive: string; search: string; sort: string }
export const initialBoardFilters: BoardFilters = { client: 'all', intent: 'all', group: 'all', source: 'all', status: 'all', archive: 'active', search: '', sort: 'activity' };

export function filterBoardProjects(items: BoardProject[], filters: BoardFilters) {
  const query = filters.search.trim().toLocaleLowerCase();
  return items.filter(item =>
    (filters.client === 'all' || (filters.client === 'unknown' ? !item.clients.length : item.clients.includes(filters.client))) &&
    (filters.intent === 'all' || item.intent === filters.intent) &&
    (filters.group === 'all' || (filters.group === 'unknown' ? !item.group : item.group === filters.group)) &&
    (filters.source === 'all' || item.sources.includes(filters.source)) &&
    (filters.status === 'all' || item.activityStatus === filters.status) &&
    (filters.archive === 'all' || item.archived === (filters.archive === 'archived')) &&
    (!query || [item.name, item.notes, item.review, item.group, ...item.folders, ...item.clients, ...item.sources].some(value => value.toLocaleLowerCase().includes(query)))
  ).sort((left, right) => filters.sort === 'name' ? left.name.localeCompare(right.name, 'zh-CN') :
    filters.sort === 'activity' ? (Date.parse(right.activityAt ?? '') || 0) - (Date.parse(left.activityAt ?? '') || 0) || left.name.localeCompare(right.name, 'zh-CN') : 0);
}

/** Grouping never duplicates a multi-client project; its complete client list stays on its card. */
export function boardSections(items: BoardProject[], view: string) {
  const groups = new Map<string, BoardProject[]>();
  for (const item of items) {
    const label = view === 'platform' ? [...item.clients].sort().join(' + ') || '客户端未知' :
      view === 'group' ? item.group || '未分组' :
      view === 'timeline' ? item.activityAt?.slice(0, 10) || '活动时间未知' : '全部项目';
    groups.set(label, [...(groups.get(label) ?? []), item]);
  }
  return [...groups].map(([label, projects]) => ({ label, projects }));
}
