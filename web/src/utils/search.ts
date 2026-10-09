/** Local presentation filter; it never changes or fabricates API records. */
export function matchesQuery(query: string, values: unknown[]): boolean {
  const needle = query.trim().toLocaleLowerCase();
  return !needle || values.some(value => typeof value === 'string' && value.toLocaleLowerCase().includes(needle));
}
