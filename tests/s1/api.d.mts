import type { components } from '../../web/src/api/schema';
type Schemas = components['schemas'];
export type ImportInput = Omit<Schemas['ImportMaterialRequestV1'], 'schema_version' | 'request_id' | 'expected_version'>;
type ReadResult<P extends string> =
  P extends '/github-repositories' ? Schemas['GitHubRepositoryListV1'] :
  P extends '/local-projects' ? Schemas['LocalProjectListV1'] :
  P extends `/local-projects/${string}/sessions` ? Schemas['NativeSessionListV1'] :
  P extends '/tracking-sources' ? Schemas['TrackingSourceListV1'] :
  P extends `/tracking-sources/${string}` ? Schemas['TrackingSourceResultV1'] :
  P extends '/jobs' ? Schemas['JobListV1'] :
  P extends '/materials' ? Schemas['MaterialListV1'] :
  P extends '/opportunities' ? Schemas['OpportunityListV1'] :
  P extends '/missions' ? Schemas['MissionListV1'] :
  P extends '/material-domains' ? Schemas['MaterialDomainListV1'] :
  P extends '/project-spaces' ? Schemas['ProjectSpaceListV1'] :
  P extends `/project-spaces/${string}` ? Schemas['ProjectSpaceResultV1'] :
  P extends '/attention-ranking-profile' ? Schemas['RankingProfileDetailV1'] :
  P extends '/distillations/processor' ? Schemas['DistillerStatusV1'] :
  P extends `/materials/${string}/revisions/${number}/content` ? Schemas['ContentResultV1'] :
  P extends `/materials/${string}` ? Schemas['MaterialDetailV1'] :
  P extends `/opportunities/${string}` ? Schemas['OpportunityDetailV1'] :
  P extends `/jobs/${string}` ? Schemas['JobV1'] : unknown;
type WriteResult<P extends string> =
  P extends '/github-repositories/sync' ? Schemas['GitHubRepositoryListV1'] :
  P extends `/github-repositories/${string}/placement` ? Schemas['GitHubRepositoryResultV1'] :
  P extends '/project-spaces' ? Schemas['ProjectSpaceResultV1'] :
  P extends '/local-projects' ? Schemas['LocalProjectResultV1'] :
  P extends '/source-collections/discover' ? Schemas['SourceCollectionListV1'] : unknown;
export interface HumanAPI {
  session: Schemas['LocalSessionV1'];
  cookie: string;
  get<P extends string>(path: P): Promise<ReadResult<P>>;
  request(path: string, options?: { method?: string; body?: unknown; key?: string; headers?: Record<string, string> }): Promise<{ status: number; data: unknown }>;
  write<P extends string>(path: P, body: Record<string, unknown>, options?: { status?: number; key?: string; method?: string }): Promise<WriteResult<P>>;
}
export function command(body?: Record<string, unknown>): { schema_version: number; request_id: string; expected_version: number } & Record<string, unknown>;
export function humanAPI(baseURL: string): Promise<HumanAPI>;
export function waitJob(api: HumanAPI, id: string, expected?: string, timeoutMs?: number): Promise<Schemas['JobV1']>;
export function importMaterial(api: HumanAPI, body: ImportInput, options?: { status?: number; key?: string; timeoutMs?: number }): Promise<{
  receipt: Schemas['ImportJobResultV1'];
  job: Schemas['JobV1'];
  detail: Schemas['MaterialDetailV1'];
}>;
