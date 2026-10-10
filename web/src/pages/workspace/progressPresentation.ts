/**
 * Agent-inferred progress presentation (pure).
 *
 * Progress is only ever inferred from approved TASK/STATUS files with explicit
 * source references, a stage and a freshness marker. A missing percentage is
 * never replaced with a fabricated number: it renders as "未知". The endpoint
 * and generated DTO are W2-owned and not yet published; this module defines the
 * local view shape the board consumes.
 */

export interface ProgressSourceRef {
  path: string;
  line?: number | null;
}

export interface ProgressInference {
  /** Human-readable inferred stage; null when nothing can be concluded. */
  stage: string | null;
  /** Inferred completion percentage; null = unknown, never fabricated. */
  percent: number | null;
  /** TASK/STATUS file references that support the inference. */
  source_refs: ProgressSourceRef[];
  /** When the inference was observed and from which source. */
  freshness: { observed_at: string | null; source: string | null };
}

export function progressPercentLabel(percent: number | null): string {
  if (percent == null) return '未知';
  const clamped = Math.max(0, Math.min(100, percent));
  return `${Math.round(clamped)}%`;
}

export function progressSourceSummary(refs: ProgressSourceRef[]): string {
  if (!refs.length) return '无依据来源';
  return refs.map(ref => (ref.line != null ? `${ref.path}:${ref.line}` : ref.path)).join('、');
}

/** Whether any inference is present (stage, percent, or an evidence ref). */
export function hasInference(inference: ProgressInference): boolean {
  return inference.stage != null || inference.percent != null || inference.source_refs.length > 0;
}

/** Human label for a stage; falls back to the raw value rather than guessing. */
export function progressStageLabel(stage: string | null): string {
  return stage ?? '未知';
}
