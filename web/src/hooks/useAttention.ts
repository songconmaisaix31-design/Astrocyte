import { attentionApi } from '../api/client';
import { useReadApi } from './useReadApi';

export function useMaterialDetail(id: string, enabled = true) {
  return useReadApi(signal => attentionApi.getMaterial(id, { signal }), { enabled, key: id });
}
export function useOpportunityDetail(id: string, enabled = true) {
  return useReadApi(signal => attentionApi.getOpportunity(id, { signal }), { enabled, key: id });
}
export function useMaterialContent(id: string, revision: number, enabled = true) {
  return useReadApi(signal => attentionApi.getMaterialContent(id, revision, { signal }), { enabled, key: `${id}@${revision}` });
}
export function useDistillations(enabled = true) {
  return useReadApi(signal => attentionApi.listDistillations({ signal }), { enabled });
}
export function useJobs(enabled = true) {
  return useReadApi(signal => attentionApi.listJobs({ signal }), { enabled });
}
