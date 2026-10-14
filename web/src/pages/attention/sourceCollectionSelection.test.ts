import { expect, it } from 'vitest';
import { sourceCollectionSelection } from './sourceCollectionSelection';

it('明确输入的收藏夹选择完整保留身份、范围和定位', () => {
  const input = { platform: 'douyin' as const, access_mode: 'browser_selected' as const, owner_id: 'input-owner', external_id: 'input-folder', locator: 'input-locator', title: '输入标题', item_count: 3 };
  expect(sourceCollectionSelection(input)).toEqual({ accessMode: 'browser_selected', externalID: 'input-folder', ownerID: 'input-owner', locator: 'input-locator', title: '输入标题' });
  expect(() => sourceCollectionSelection({ ...input, owner_id: 'self' })).toThrow('明确来源身份');
  expect(() => sourceCollectionSelection({ ...input, external_id: '' })).toThrow('明确来源身份');
});
