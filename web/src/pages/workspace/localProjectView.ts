import type { components } from '../../api/schema';

type Project = components['schemas']['ProjectV1'];
type Session = components['schemas']['SessionV1'];

/** Presentation fields only. Absent backend facts stay null until an adapter publishes them. */
export interface LocalProjectView {
  project: Project;
  platforms: string[];
  bindingStates: Session['binding_status'][];
  group: string | null;
  lastActivity: string | null;
  fileCount: number | null;
  handoff: string | null;
}

export function localProjectViews(projects: Project[], sessions: Session[]): LocalProjectView[] {
  return projects.map(project => {
    const linked = sessions.filter(session => session.project_id === project.id);
    return { project, platforms: [...new Set(linked.map(session => session.adapter))],
      bindingStates: [...new Set(linked.map(session => session.binding_status))],
      group: null, lastActivity: null, fileCount: null, handoff: null };
  });
}

export function filterLocalProjects(items: LocalProjectView[], platform: string, status: string, group: string, sort: string) {
  const filtered = items.filter(item => (platform === 'all' || (platform === 'unknown' ? !item.platforms.length : item.platforms.includes(platform))) &&
    (status === 'all' || (status === 'unknown' ? !item.bindingStates.length : item.bindingStates.includes(status as Session['binding_status']))) &&
    (group === 'all' || (group === 'unknown' ? item.group === null : item.group === group)));
  if (sort === 'name') return filtered.sort((a, b) => a.project.name.localeCompare(b.project.name, 'zh-CN'));
  if (sort === 'activity') return filtered.sort((a, b) => {
    const time = (value: string | null) => value && Number.isFinite(Date.parse(value)) ? Date.parse(value) : -Infinity;
    const left = time(a.lastActivity), right = time(b.lastActivity);
    return left === right ? 0 : right > left ? 1 : -1;
  });
  return filtered;
}
