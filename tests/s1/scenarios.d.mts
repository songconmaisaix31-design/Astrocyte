import type { components } from '../../web/src/api/schema';
import type { HumanAPI } from './api.mjs';
type Schemas = components['schemas'];
export function seedCandidate(api: HumanAPI, locator?: string): Promise<{
  receipt: Schemas['ImportJobResultV1'];
  job: Schemas['JobV1'];
  detail: Schemas['MaterialDetailV1'];
  refs: Schemas['SourceRefV1'][];
  records: Schemas['DistillationV1'][];
  candidate: Schemas['OpportunityDetailV1'];
}>;
