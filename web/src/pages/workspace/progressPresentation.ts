/** 项目进度的纯显示类型；进度、来源、百分数与观察时间均由服务提供。 */

export interface ProgressEvidence {
  source_path: string;
  kind: string;
  version: string;
  excerpt?: string;
}

export interface ProjectProgress {
  project_id: string;
  status: string;
  summary?: string;
  percent?: number | null;
  source?: 'human' | 'agent_inferred' | null;
  evidence: ProgressEvidence[];
  native_id?: string;
  model?: string | null;
  observed_at?: string | null;
  warning?: string;
  revision: number;
}

export interface ProjectProgressResult {
  schema_version: 1;
  progress: ProjectProgress;
}

/** 未知或缺失状态明确显示为未知。 */
export function progressStatusLabel(status: string | null | undefined): string {
  if (!status || status === 'unknown') return '未知';
  return status;
}

/** 分别显示人工记录与 Agent 推断的来源。 */
export function progressSourceLabel(source: ProjectProgress['source']): string {
  return source === 'human' ? '人工记录' : source === 'agent_inferred' ? 'Agent 推断' : '尚未记录';
}

/** 百分数保留人工记录或 Agent 估计的来源。 */
export function progressPercentLabel(source: ProjectProgress['source']): string {
  return source === 'human' ? '人工记录' : source === 'agent_inferred' ? 'Agent 估计' : '未知';
}

/** 显示依据文件及其版本。 */
export function progressSourceSummary(evidence: ProgressEvidence[] | undefined): string {
  if (!evidence || !evidence.length) return '无依据来源';
  return evidence.map(ref => (ref.version ? `${ref.source_path}:${ref.version}` : ref.source_path)).join('、');
}

/** 判断是否已有进度记录。 */
export function hasProgress(progress: ProjectProgress): boolean {
  return !!progress.source || !!progress.summary || progress.percent != null || progress.evidence.length > 0;
}

/** 已知依据类型使用中文标签，其他类型保留原值。 */
export function progressKindLabel(kind: string): string {
  switch (kind) {
    case 'task': return '任务文件';
    case 'status': return '状态文件';
    case 'project_file': return '项目文件';
    default: return kind;
  }
}

/** 单次推断接受的人工输入文件数量上限。 */
export const PROGRESS_FILES_LIMIT = 8;

/** 保留每个非空路径；调用方提示数量超限，服务负责核对已批准目录。 */
export function parseProgressFiles(text: string): string[] {
  return text.split(/\r?\n/).map(line => line.trim()).filter(Boolean);
}
