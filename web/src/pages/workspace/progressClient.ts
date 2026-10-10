/**
 * Project progress client seam.
 *
 * W0 published `GET /api/v1/local-projects/{id}/progress` and
 * `POST /api/v1/local-projects/{id}/progress/infer` (contracts/openapi.yaml); the
 * generated `localProjectsApi.getProjectProgress`/`inferProjectProgress` wrappers
 * are added by W0 during final integration because `web/src/api/` is W0-owned.
 * Until then this module calls the published routes directly with the exact
 * `ProjectProgressResultV1` shape and never fabricates a status.
 */
import type { ProjectProgress, ProjectProgressResult } from './progressPresentation';

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
  source_path?: unknown; kind?: unknown; version?: unknown; freshness?: unknown; excerpt?: unknown;
}
interface RawProgress {
  project_id?: unknown; status?: unknown; summary?: unknown; evidence?: unknown;
  inferred?: unknown; inferred_at?: unknown; processor?: unknown; model?: unknown; warning?: unknown;
}

function mapProgress(raw: RawProgress): ProjectProgress {
  const evidence = Array.isArray(raw.evidence)
    ? (raw.evidence as RawEvidence[]).map(item => ({
      source_path: asString(item.source_path) ?? '',
      kind: asString(item.kind) ?? '',
      version: asString(item.version) ?? '',
      freshness: asString(item.freshness) ?? '',
      excerpt: asString(item.excerpt) ?? undefined,
    }))
    : undefined;
  return {
    project_id: asString(raw.project_id) ?? '',
    status: asString(raw.status) ?? 'unknown',
    summary: asString(raw.summary) ?? undefined,
    evidence,
    inferred: raw.inferred === true,
    inferred_at: asString(raw.inferred_at) ?? undefined,
    processor: asString(raw.processor) ?? undefined,
    model: asString(raw.model) ?? undefined,
    warning: asString(raw.warning) ?? undefined,
  };
}

export async function getProjectProgress(id: string, signal?: AbortSignal): Promise<ProjectProgressResult> {
  const response = await fetch(`${PROGRESS_PATH}/${encodeURIComponent(id)}/progress`, { credentials: 'same-origin', signal });
  if (!response.ok) throw new ProgressUnavailable(response.status);
  const body = (await response.json().catch(() => undefined)) as { progress?: unknown } | undefined;
  return { schema_version: 1, progress: mapProgress((body?.progress && typeof body.progress === 'object' ? body.progress : {}) as RawProgress) };
}

/** Human-session infer trigger; the operation identity is caller-owned and never replayed. */
export async function inferProjectProgress(id: string, key: string, signal?: AbortSignal): Promise<ProjectProgressResult> {
  const sessionResponse = await fetch('/api/v1/auth/session', { credentials: 'same-origin', signal });
  if (!sessionResponse.ok) throw new ProgressUnavailable(sessionResponse.status);
  const session = (await sessionResponse.json().catch(() => undefined)) as { csrf_token?: string } | undefined;
  const response = await fetch(`${PROGRESS_PATH}/${encodeURIComponent(id)}/progress/infer`, {
    method: 'POST',
    credentials: 'same-origin',
    signal,
    headers: { 'Content-Type': 'application/json', 'Idempotency-Key': key, 'X-CSRF-Token': session?.csrf_token ?? '' },
    body: JSON.stringify({ schema_version: 1, request_id: key, expected_version: 1 }),
  });
  if (!response.ok) throw new ProgressUnavailable(response.status);
  const body = (await response.json().catch(() => undefined)) as { progress?: unknown } | undefined;
  return { schema_version: 1, progress: mapProgress((body?.progress && typeof body.progress === 'object' ? body.progress : {}) as RawProgress) };
}
