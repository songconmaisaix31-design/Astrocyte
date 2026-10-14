import { api, ApiError } from './client';
import type { components, operations } from './schema';

export type RelaySessionEvents = components['schemas']['RelaySessionEventsV1'];
export type RelaySessionCoverage = components['schemas']['RelaySessionCoverageV1'];
export type RelaySessionQuery = NonNullable<operations['listRelaySessionEvents']['parameters']['query']>;

/** This client can only read the existing human/session/origin cache endpoint. */
export function createRelaySessionApi(client = api) {
  return {
    async events(query: RelaySessionQuery = {}, signal?: AbortSignal): Promise<RelaySessionEvents> {
      const { data, error, response } = await client.GET('/relay/session-events', { params: { query }, signal });
      if (error) throw new ApiError(response.status, error);
      if (!data) throw new Error('本机会话事件服务未返回数据');
      return data;
    },
  };
}
export const relaySessionApi = createRelaySessionApi();
