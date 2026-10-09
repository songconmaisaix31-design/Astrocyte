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
  async function read<P extends '/health' | '/foundation' | '/materials' | '/opportunities' | '/projects' | '/proposals' | '/sessions' | '/missions'>(path: P, options: ReadOptions = {}) {
    const { data, error, response } = await client.GET(path, { signal: options.signal });
    if (error) throw new ApiError(response.status, error);
    if (!data) throw new Error(`API returned no JSON data: ${path}`);
    return data;
  }
  return {
    getHealth: (options?: ReadOptions) => read('/health', options),
    getFoundation: (options?: ReadOptions) => read('/foundation', options),
    listMaterials: (options?: ReadOptions) => read('/materials', options),
    listOpportunities: (options?: ReadOptions) => read('/opportunities', options),
    listProjects: (options?: ReadOptions) => read('/projects', options),
    listProposals: (options?: ReadOptions) => read('/proposals', options),
    listSessions: (options?: ReadOptions) => read('/sessions', options),
    listMissions: (options?: ReadOptions) => read('/missions', options),
  };
}
export const { getHealth, getFoundation, listMaterials, listOpportunities, listProjects, listProposals, listSessions, listMissions } = createReadApi();
