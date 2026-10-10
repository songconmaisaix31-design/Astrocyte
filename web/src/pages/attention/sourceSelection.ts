import type { components } from '../../api/schema';
type Item = components['schemas']['SourceItemV1'];
export const sourceItemKey = (item: Item) => `${item.external_id}@${item.revision}`;
export function canSelectSourceItem(item: Item) {
  return !item.stale && !item.selected && !item.metadata.unavailable_reason && item.recommendation?.status === 'succeeded' && !!item.recommendation.text.trim() && item.recommendation.metadata_revision === item.revision;
}
export function canRecommendSourceItem(item: Item) {
  if (item.stale || item.selected || item.metadata.unavailable_reason) return false;
  const recommendation = item.recommendation;
  if (!recommendation) return true;
  if (['unknown', 'running', 'pending'].includes(recommendation.status) || recommendation.error?.code === 'delivery_unknown') return false;
  if (recommendation.status === 'failed') return !!recommendation.error?.retryable;
  return recommendation.status !== 'succeeded' || recommendation.metadata_revision !== item.revision;
}
