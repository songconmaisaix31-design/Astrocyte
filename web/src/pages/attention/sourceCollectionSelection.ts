import type { components } from '../../api/schema';
type Collection = components['schemas']['SourceCollectionV1'];

/** 选择函数完整保留服务提供的来源身份，不从 self 或显示标题推断身份。 */
export function sourceCollectionSelection(collection: Collection) {
  if (!collection.owner_id || collection.owner_id === 'self' || !collection.external_id || !collection.locator || !collection.access_mode) throw new Error('收藏夹缺少明确来源身份或读取范围');
  return { accessMode: collection.access_mode, externalID: collection.external_id, ownerID: collection.owner_id, locator: collection.locator, title: collection.title };
}
