import { api, ApiError, type ApiErrorBody, type components } from './client';

type Options = { signal?: AbortSignal };
async function unwrap<T>(result: Promise<{ data?: T; error?: ApiErrorBody; response: Response }>): Promise<T> {
  const { data, error, response } = await result;
  if (error) throw new ApiError(response.status, error);
  if (!data) throw new Error('API returned no JSON data');
  return data;
}
async function headers(key: string) {
  if (!key.trim()) throw new Error('Idempotency key is required');
  const session = await unwrap(api.GET('/auth/session'));
  return { 'Idempotency-Key': key, 'X-CSRF-Token': session.csrf_token };
}
// Writes keep the caller's operation identity and never replay failed/unknown effects.
export const trackingApi = {
	listCollections: (platform: string, owner_id: string, options?: Options) => unwrap(api.GET('/source-collections', { params: { query: { platform, owner_id } }, ...options })),
  listSources: (options?: Options) => unwrap(api.GET('/tracking-sources', options)),
  getSource: (id: string, options?: Options) => unwrap(api.GET('/tracking-sources/{id}', { params: { path: { id } }, ...options })),
  bindSource: async (body: components['schemas']['BindTrackingSourceRequestV1'], key: string, options?: Options) => unwrap(api.POST('/tracking-sources', { body, params: { header: await headers(key) }, ...options })),
  syncSource: async (id: string, body: components['schemas']['SyncTrackingSourceRequestV1'], key: string, options?: Options) => unwrap(api.POST('/tracking-sources/{id}/sync', { body, params: { path: { id }, header: await headers(key) }, ...options })),
  recommendItems: async (id: string, body: components['schemas']['RecommendSourceItemsRequestV1'], key: string, options?: Options) => unwrap(api.POST('/tracking-sources/{id}/recommend', { body, params: { path: { id }, header: await headers(key) }, ...options })),
  selectItems: async (id: string, body: components['schemas']['SelectSourceItemsRequestV1'], key: string, options?: Options) => unwrap(api.POST('/tracking-sources/{id}/select', { body, params: { path: { id }, header: await headers(key) }, ...options })),
};
export const localProjectsApi = {
	issueAgentToken: async (id: string, body: components['schemas']['IssueProjectAgentTokenRequestV1'], key: string, options?: Options) => unwrap(api.POST('/local-projects/{id}/agent-token', { body, params: { path: { id }, header: await headers(key) }, ...options })),
	discoverSessions: async (id: string, body: components['schemas']['DiscoverNativeSessionsRequestV1'], key: string, options?: Options) => unwrap(api.POST('/local-projects/{id}/sessions/discover', { body, params: { path: { id }, header: await headers(key) }, ...options })),
	readNativeContext: (id: string, session_id: string, options?: Options) => unwrap(api.GET('/local-projects/{id}/sessions/{session_id}/context', { params: { path: { id, session_id } }, ...options })),
  listProjects: (options?: Options) => unwrap(api.GET('/local-projects', options)),
  registerProject: async (body: components['schemas']['RegisterLocalProjectRequestV1'], key: string, options?: Options) => unwrap(api.POST('/local-projects', { body, params: { header: await headers(key) }, ...options })),
  discoverProjects: async (body: components['schemas']['DiscoverLocalProjectsRequestV1'], key: string, options?: Options) => unwrap(api.POST('/local-projects/discover', { body, params: { header: await headers(key) }, ...options })),
  setSettings: async (id: string, body: components['schemas']['LocalProjectSettingsRequestV1'], key: string, options?: Options) => unwrap(api.PUT('/local-projects/{id}/settings', { body, params: { path: { id }, header: await headers(key) }, ...options })),
  grantAgent: async (id: string, body: components['schemas']['GrantLocalProjectAgentRequestV1'], key: string, options?: Options) => unwrap(api.POST('/local-projects/{id}/grants', { body, params: { path: { id }, header: await headers(key) }, ...options })),
  revokeAgent: async (id: string, body: components['schemas']['RevokeLocalProjectAgentRequestV1'], key: string, options?: Options) => unwrap(api.POST('/local-projects/{id}/grants/revoke', { body, params: { path: { id }, header: await headers(key) }, ...options })),
  readContext: async (id: string, body: components['schemas']['ReadLocalProjectContextRequestV1'], key: string, options?: Options) => unwrap(api.POST('/local-projects/{id}/context', { body, params: { path: { id }, header: await headers(key) }, ...options })),
  listSessions: (id: string, options?: Options) => unwrap(api.GET('/local-projects/{id}/sessions', { params: { path: { id } }, ...options })),
  startSession: async (id: string, body: components['schemas']['NativeSessionRequestV1'], key: string, options?: Options) => unwrap(api.POST('/local-projects/{id}/sessions', { body, params: { path: { id }, header: await headers(key) }, ...options })),
  resumeSession: async (id: string, session_id: string, body: components['schemas']['NativeSessionRequestV1'], key: string, options?: Options) => unwrap(api.POST('/local-projects/{id}/sessions/{session_id}/resume', { body, params: { path: { id, session_id }, header: await headers(key) }, ...options })),
  sendMessage: async (id: string, session_id: string, body: components['schemas']['NativeMessageRequestV1'], key: string, options?: Options) => unwrap(api.POST('/local-projects/{id}/sessions/{session_id}/send', { body, params: { path: { id, session_id }, header: await headers(key) }, ...options })),
  stopSession: async (id: string, session_id: string, body: components['schemas']['JobCommandV1'], key: string, options?: Options) => unwrap(api.POST('/local-projects/{id}/sessions/{session_id}/stop', { body, params: { path: { id, session_id }, header: await headers(key) }, ...options })),
  observeSession: (id: string, session_id: string, options?: Options) => unwrap(api.GET('/local-projects/{id}/sessions/{session_id}/observe', { params: { path: { id, session_id } }, ...options })),
};
