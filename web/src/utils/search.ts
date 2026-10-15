/** 按输入关键词筛选已有 API 记录。 */
export function matchesQuery(query: string, values: unknown[]): boolean {
  const needle = query.trim().toLocaleLowerCase();
  return !needle || values.some(value => typeof value === 'string' && value.toLocaleLowerCase().includes(needle));
}
