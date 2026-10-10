/**
 * Project progress presentation (pure).
 *
 * Progress is a human-authored or model-inferred advisory stage read only from
 * approved fixed TASK/STATUS files. The local shapes mirror W0's published
 * `ProjectProgressV1`/`ProgressEvidenceV1` (contracts/openapi.yaml, owned by W0,
 * converged at `e39f8f1`). `source` and `observed_at` are null until a human
 * writes or an Agent inference concludes; an unobserved project is reported as
 * status=unknown with a null observed_at, never a fabricated timestamp. A
 * percent is a human-authored value or an evidence-backed Agent estimate, never
 * invented and never presented as human approval when it is not.
 */

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

/** Human label for a status; unknown/absent is presented as-is, never invented. */
export function progressStatusLabel(status: string | null | undefined): string {
  if (!status || status === 'unknown') return '未知';
  return status;
}

/** Human label for the evidence source (human-authored vs Agent-inferred). */
export function progressSourceLabel(source: ProjectProgress['source']): string {
  return source === 'human' ? '人工记录' : source === 'agent_inferred' ? 'Agent 推断' : '尚未记录';
}

/** Human label for a percent value's provenance; a model estimate is never shown as human input. */
export function progressPercentLabel(source: ProjectProgress['source']): string {
  return source === 'human' ? '人工记录' : source === 'agent_inferred' ? 'Agent 估计' : '未知';
}

/** Readable summary of the evidence's source files (path:version). */
export function progressSourceSummary(evidence: ProgressEvidence[] | undefined): string {
  if (!evidence || !evidence.length) return '无依据来源';
  return evidence.map(ref => (ref.version ? `${ref.source_path}:${ref.version}` : ref.source_path)).join('、');
}

/** Whether any progress has been recorded (human or inferred). */
export function hasProgress(progress: ProjectProgress): boolean {
  return !!progress.source || !!progress.summary || progress.percent != null || progress.evidence.length > 0;
}

/** Human label for the evidence kind; falls back to the raw value. */
export function progressKindLabel(kind: string): string {
  switch (kind) {
    case 'task': return '任务文件';
    case 'status': return '状态文件';
    case 'project_file': return '项目文件';
    default: return kind;
  }
}
