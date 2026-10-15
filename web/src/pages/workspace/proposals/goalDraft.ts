import type { Goal } from '../../../api/proposals';

export const goalStatuses: { value: Goal['status']; label: string }[] = [
  { value: 'active', label: '进行中' }, { value: 'paused', label: '已暂停' },
  { value: 'completed', label: '已完成' }, { value: 'archived', label: '已归档' },
];

export function goalStatusLabel(status: Goal['status']) {
  const item = goalStatuses.find(value => value.value === status);
  if (!item) throw new Error('目标状态不在已生成契约的可选范围内');
  return item.label;
}
