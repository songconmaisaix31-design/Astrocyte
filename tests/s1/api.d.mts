import type { components } from '../../web/src/api/schema';
type Schemas = components['schemas'];
export type ImportInput = Omit<Schemas['ImportMaterialRequestV1'], 'schema_version' | 'request_id' | 'expected_version'>;
type ReadResult<P extends string> =
  P extends '/materials' ? Schemas['MaterialListV1'] :
  P extends '/opportunities' ? Schemas['OpportunityListV1'] :
  P extends '/missions' ? Schemas['MissionListV1'] :
  P extends `/materials/${string}/revisions/${number}/content` ? Schemas['ContentResultV1'] :
  P extends `/materials/${string}` ? Schemas['MaterialDetailV1'] :
  P extends `/opportunities/${string}` ? Schemas['OpportunityDetailV1'] :
  P extends `/jobs/${string}` ? Schemas['JobV1'] : unknown;
export interface HumanAPI {
  session: Schemas['LocalSessionV1'];
  cookie: string;
  get<P extends string>(path: P): Promise<ReadResult<P>>;
}
export function humanAPI(baseURL: string): Promise<HumanAPI>;
export function waitJob(api: HumanAPI, id: string, expected?: string, timeoutMs?: number): Promise<Schemas['JobV1']>;
export function importMaterial(api: HumanAPI, body: ImportInput, options?: { status?: number; key?: string; timeoutMs?: number }): Promise<{
  receipt: Schemas['ImportJobResultV1'];
  job: Schemas['JobV1'];
  detail: Schemas['MaterialDetailV1'];
}>;
