import createClient from 'openapi-fetch';
import type { components, paths } from './schema';

export type { components, operations, paths } from './schema';
export type Health = components['schemas']['HealthV1'];
export type Foundation = components['schemas']['FoundationV1'];
export type Material = components['schemas']['MaterialV1'];
export type Opportunity = components['schemas']['OpportunityV1'];
export type Project = components['schemas']['ProjectV1'];
export type Proposal = components['schemas']['ProposalV1'];
export type Session = components['schemas']['SessionV1'];
export type Mission = components['schemas']['MissionV1'];
export type ApiErrorBody = components['schemas']['ErrorV1'];

/** Uses OpenAPI-generated path, request, response and error types. No retries. */
export const createApiClient = (baseUrl = '/api/v1', fetchImpl?: typeof fetch) =>
  createClient<paths>({ baseUrl, ...(fetchImpl ? { fetch: fetchImpl } : {}) });
export const api = createApiClient();

export class ApiError extends Error {
  constructor(readonly status: number, readonly detail: ApiErrorBody) {
    super(detail.error.message);
    this.name = 'ApiError';
  }
  get code() { return this.detail.error.code; }
  get requestId() { return this.detail.error.request_id; }
}

type ReadOptions = { signal?: AbortSignal };
export function createReadApi(client = api) {
  async function unwrap<T>(result: Promise<{ data?: T; error?: ApiErrorBody; response: Response }>): Promise<T> {
    const { data, error, response } = await result;
    if (error) throw new ApiError(response.status, error);
    if (!data) throw new Error('API returned no JSON data');
    return data;
  }
  return {
    getHealth: (options?: ReadOptions) => unwrap(client.GET('/health', options)),
    getFoundation: (options?: ReadOptions) => unwrap(client.GET('/foundation', options)),
    listMaterials: (options?: ReadOptions) => unwrap(client.GET('/materials', options)),
    listOpportunities: (options?: ReadOptions) => unwrap(client.GET('/opportunities', options)),
    listProjects: (options?: ReadOptions) => unwrap(client.GET('/projects', options)),
    listProposals: (options?: ReadOptions) => unwrap(client.GET('/proposals', options)),
    listSessions: (options?: ReadOptions) => unwrap(client.GET('/sessions', options)),
    listMissions: (options?: ReadOptions) => unwrap(client.GET('/missions', options)),
  };
}
export const { getHealth, getFoundation, listMaterials, listOpportunities, listProjects, listProposals, listSessions, listMissions } = createReadApi();
