import { attentionApi } from '../../../api/client';
import { useReadApi } from '../../../hooks/useReadApi';
import type { EvidenceReference } from './judgmentApi';

export interface ProjectEvidence {
  reference: EvidenceReference;
  label: string;
}

export const evidenceKey = (reference: EvidenceReference) => `${reference.material_id}:${reference.revision}:${reference.digest}`;

export function useProjectEvidence(projectID: string, spaceID: string) {
  return useReadApi(async signal => {
    const detail = await attentionApi.getProjectSpace(spaceID, { signal });
    const references = Array.from(new Map(detail.space.material_refs.map(reference => [`${reference.material_id}:${reference.revision}`, reference])).values());
    return Promise.all(references.map(async reference => {
      const material = await attentionApi.getMaterial(reference.material_id, { signal });
      const revision = material.revisions.find(value => value.revision === reference.revision);
      if (!revision) throw new Error('项目已纳入的固定资料版本当前无法读取，请核查版本和权限');
      return { reference: { material_id: reference.material_id, revision: reference.revision, digest: revision.content_digest }, label: `${material.material.title} · r${reference.revision} · ${reference.locator}` };
    }));
  }, { key: `${projectID}:${spaceID}` });
}
