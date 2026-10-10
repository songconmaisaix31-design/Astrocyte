/**
 * Project progress presentation (pure).
 *
 * Progress is only ever a model-inferred observation read from approved fixed
 * TASK/STATUS files; the local shapes below mirror W0's published
 * `ProjectProgressV1`/`ProgressEvidenceV1` (contracts/openapi.yaml, owned by W0).
 * The service contract has no percentage field — status is a free string and
 * "unknown" means no inference was concluded. Nothing here fabricates a number.
 */

export interface ProgressEvidence {
  source_path: string;
  kind: string;
  version: string;
  freshness: string;
  excerpt?: string;
}

export interface ProjectProgress {
  project_id: string;
  status: string;
  summary?: string;
  evidence?: ProgressEvidence[];
  inferred: boolean;
  inferred_at?: string | null;
  processor?: string;
  model?: string | null;
  warning?: string;
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

/** Readable summary of the evidence's source files (path:version). */
export function progressSourceSummary(evidence: ProgressEvidence[] | undefined): string {
  if (!evidence || !evidence.length) return '无依据来源';
  return evidence.map(ref => (ref.version ? `${ref.source_path}:${ref.version}` : ref.source_path)).join('、');
}

/** Whether any inference evidence is present (status, summary, or an evidence ref). */
export function hasInference(progress: ProjectProgress): boolean {
  return progress.inferred || !!progress.summary || (progress.evidence?.length ?? 0) > 0;
}

/** Human label for the evidence kind; falls back to the raw value. */
export function progressKindLabel(kind: string): string {
  switch (kind) {
    case 'task': return '任务文件';
    case 'status': return '状态文件';
    default: return kind;
  }
}
