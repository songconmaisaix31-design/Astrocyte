import { api, ApiError, type ApiErrorBody } from './client';
import type { components, operations } from './schema';

type S = components['schemas'];
export type RepositoryBudget = S['RepositoryBudgetV1'];
export type RepositoryMapQuery = NonNullable<operations['readRepositoryContextMap']['parameters']['query']>;
export type RepositoryQueryRequest = S['RepositoryQueryRequestV1'];
export type RepositorySearchRequest = S['RepositorySearchRequestV1'];
export type RepositoryImpactRequest = S['RepositoryImpactRequestV1'];
export type RepositoryIndexRequest = S['RepositoryIndexRequestV1'];
export type RepositoryResult = S['RepositoryResultV1'];
export type RepositoryIndex = S['RepositoryIndexV1'];
export type RepositoryVersion = S['RepositoryVersionV1'];
export type RepositorySymbol = S['RepositorySymbolV1'];
export type RepositoryRelation = S['RepositoryRelationV1'];
export type RepositoryMapEntry = S['RepositoryMapEntryV1'];
export type RepositoryResultEnvelope = S['RepositoryResultEnvelopeV1'];
export type RepositoryIndexEnvelope = S['RepositoryIndexEnvelopeV1'];

/** 仓库查询与维护均使用准确生成契约；POST 沿既有人工会话传递 CSRF。 */
export function createRepositoryContextApi(client = api) {
  async function unwrap<T>(request: Promise<{ data?: T; error?: ApiErrorBody; response: Response }>): Promise<T> {
    const { data, error, response } = await request;
    if (error) throw new ApiError(response.status, error);
    if (!data) throw new Error('仓库上下文服务未返回数据');
    return data;
  }
  async function headers(signal?: AbortSignal) {
    const session = await unwrap(client.GET('/auth/session', { signal }));
    return { 'X-CSRF-Token': session.csrf_token };
  }
  return {
    map: (id: string, query: RepositoryMapQuery, signal?: AbortSignal) => unwrap(client.GET('/local-projects/{id}/repository-context/map', { params: { path: { id }, query }, signal })),
    search: async (id: string, body: RepositorySearchRequest, signal?: AbortSignal) => unwrap(client.POST('/local-projects/{id}/repository-context/search', { params: { path: { id } }, headers: await headers(signal), body, signal })),
    impact: async (id: string, body: RepositoryImpactRequest, signal?: AbortSignal) => unwrap(client.POST('/local-projects/{id}/repository-context/impact', { params: { path: { id } }, headers: await headers(signal), body, signal })),
    index: async (id: string, body: RepositoryIndexRequest, signal?: AbortSignal) => unwrap(client.POST('/local-projects/{id}/repository-context/index', { params: { path: { id } }, headers: await headers(signal), body, signal })),
    refreshMap: async (id: string, body: RepositoryQueryRequest, signal?: AbortSignal) => unwrap(client.POST('/local-projects/{id}/repository-context/map/refresh', { params: { path: { id } }, headers: await headers(signal), body, signal })),
  };
}

export const repositoryContextApi = createRepositoryContextApi();
