/**
 * Project progress client seam.
 *
 * W0 published `GET/PUT /api/v1/local-projects/{id}/progress` and
 * `POST /api/v1/local-projects/{id}/progress/infer` (contracts/openapi.yaml,
 * converged at `e39f8f1`); the generated `localProjectsApi.getProjectProgress`/
 * `setProjectProgress`/`inferProjectProgress` wrappers are added by W0 during
 * final integration because `web/src/api/` is W0-owned. Until then this module
 * calls the published routes directly with the exact wire shape and never
 * fabricates a status or a percent.
 */
import type { ProgressEvidence, ProjectProgress, ProjectProgressResult } from './progressPresentation';

export const PROGRESS_PATH = '/api/v1/local-projects';

export class ProgressUnavailable extends Error {
  constructor(readonly status: number) {
    super(status === 404 || status === 501 ? '项目进度尚未接入（服务未提供该能力）' : `项目进度暂不可用（HTTP ${status}）`);
    this.name = 'ProgressUnavailable';
  }
}

function asString(value: unknown): string | null {
  return typeof value === 'string' ? value : null;
}

interface RawEvidence {
  source_path?: unknown; kind?: unknown; version?: unknown; excerpt?: unknown;
}
interface RawProgress {
  project_id?: unknown; status?: unknown; summary?: unknown; percent?: unknown; source?: unknown;
  evidence?: unknown; native_id?: unknown; model?: unknown; observed_at?: unknown;
  warning?: unknown; revision?: unknown;
}

function mapProgress(raw: RawProgress): ProjectProgress {
  const evidence: ProgressEvidence[] = Array.isArray(raw.evidence)
    ? (raw.evidence as RawEvidence[]).map(item => ({
      source_path: asString(item.source_path) ?? '',
      kind: asString(item.kind) ?? '',
      version: asString(item.version) ?? '',
      excerpt: asString(item.excerpt) ?? undefined,
    }))
    : [];
  const source = asString(raw.source);
  return {
    project_id: asString(raw.project_id) ?? '',
    status: asString(raw.status) || 'unknown',
    summary: asString(raw.summary) ?? undefined,
    percent: typeof raw.percent === 'number' ? raw.percent : null,
    source: source === 'human' || source === 'agent_inferred' ? source : null,
    evidence,
    native_id: asString(raw.native_id) ?? undefined,
    model: asString(raw.model) ?? null,
    observed_at: asString(raw.observed_at) ?? null,
    warning: asString(raw.warning) ?? undefined,
    revision: typeof raw.revision === 'number' ? raw.revision : 0,
  };
}

export async function getProjectProgress(id: string, signal?: AbortSignal): Promise<ProjectProgressResult> {
  const response = await fetch(`${PROGRESS_PATH}/${encodeURIComponent(id)}/progress`, { credentials: 'same-origin', signal });
  if (!response.ok) throw new ProgressUnavailable(response.status);
  const body = (await response.json().catch(() => undefined)) as { progress?: unknown } | undefined;
  return { schema_version: 1, progress: mapProgress((body?.progress && typeof body.progress === 'object' ? body.progress : {}) as RawProgress) };
}

/** Human-session infer trigger; the caller-supplied Idempotency-Key is the operation identity. */
export async function inferProjectProgress(id: string, key: string, files: string[], signal?: AbortSignal): Promise<ProjectProgressResult> {
  const sessionResponse = await fetch('/api/v1/auth/session', { credentials: 'same-origin', signal });
  if (!sessionResponse.ok) throw new ProgressUnavailable(sessionResponse.status);
  const session = (await sessionResponse.json().catch(() => undefined)) as { csrf_token?: string } | undefined;
  const response = await fetch(`${PROGRESS_PATH}/${encodeURIComponent(id)}/progress/infer`, {
    method: 'POST',
    credentials: 'same-origin',
    signal,
    headers: { 'Content-Type': 'application/json', 'Idempotency-Key': key, 'X-CSRF-Token': session?.csrf_token ?? '' },
    body: JSON.stringify({ schema_version: 1, request_id: key, expected_version: 1, files }),
  });
  if (!response.ok) throw new ProgressUnavailable(response.status);
  const body = (await response.json().catch(() => undefined)) as { progress?: unknown } | undefined;
  return { schema_version: 1, progress: mapProgress((body?.progress && typeof body.progress === 'object' ? body.progress : {}) as RawProgress) };
}
