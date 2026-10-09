const paths = {
  search: 'M21 21l-5.2-5.2M19 10.5a8.5 8.5 0 1 1-17 0 8.5 8.5 0 0 1 17 0Z',
  library: 'M4 4h5v16H4V4Zm8 0h4v16h-4V4Zm6 1 3-1 4 15-3 1-4-15',
  grid: 'M3 3h7v7H3V3Zm11 0h7v7h-7V3ZM3 14h7v7H3v-7Zm11 0h7v7h-7v-7',
  swarm: 'M6 6l6 6m0 0 6-6m-6 6-6 6m6-6 6 6M8 6a2 2 0 1 1-4 0 2 2 0 0 1 4 0Zm12 0a2 2 0 1 1-4 0 2 2 0 0 1 4 0ZM8 18a2 2 0 1 1-4 0 2 2 0 0 1 4 0Zm12 0a2 2 0 1 1-4 0 2 2 0 0 1 4 0Zm-6-6a2 2 0 1 1-4 0 2 2 0 0 1 4 0Z',
  arrow: 'M4 12h15m-6-6 6 6-6 6',
  chevron: 'm9 5 7 7-7 7',
  plus: 'M12 4v16M4 12h16',
  file: 'M14 2H5v20h14V7l-5-5Zm0 0v6h5M8 12h8m-8 4h6',
  video: 'M4 5h16v14H4V5Zm6 4 5 3-5 3V9Z',
  spark: 'm12 2 3 7 7 3-7 3-3 7-3-7-7-3 7-3 3-7Z',
  branch: 'M6 5v14m0-7h6a6 6 0 0 0 6-6M8 4a2 2 0 1 1-4 0 2 2 0 0 1 4 0Zm0 16a2 2 0 1 1-4 0 2 2 0 0 1 4 0Zm12-16a2 2 0 1 1-4 0 2 2 0 0 1 4 0Z',
  close: 'm6 6 12 12M6 18 18 6',
  layers: 'm12 3 10 5-10 5L2 8l10-5Zm-10 9 10 5 10-5M2 17l10 5 10-5',
  lock: 'M6 10h12v11H6V10Zm3 0V6a3 3 0 0 1 6 0v4',
  menu: 'M4 6h16M4 12h16M4 18h16',
  info: 'M12 11v6m0-10h.01M22 12a10 10 0 1 1-20 0 10 10 0 0 1 20 0Z',
} as const;

export function Icon({ name, size = 20 }: { name: keyof typeof paths; size?: number }) {
  return <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.65" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true"><path d={paths[name]} /></svg>;
}

/** Brand SVG from the supplied design, independent of its bundled runtime. */
export function Mark() {
  return <svg className="ac-mark" viewBox="0 0 40 40" fill="none" aria-hidden="true">
    <path d="M20 6v10M8 12l9 6M8 28l9-6m3 2v10m3-12 9 6m-9-10 9-6" stroke="currentColor" strokeWidth="3" />
    <circle cx="20" cy="20" r="6" fill="currentColor" />
    {[[20, 5], [7, 12], [7, 28], [20, 35], [33, 28], [33, 12]].map(([x, y]) => <circle key={`${x}-${y}`} cx={x} cy={y} r="3.5" fill="currentColor" />)}
  </svg>;
}

/** Decorative brand illustration from the design; never an execution graph. */
export function NetworkArt() {
  const nodes = [[62, 94], [118, 37], [136, 145], [210, 74], [234, 157], [305, 45], [339, 112], [391, 70], [389, 178]];
  const edges = [[0, 1], [0, 2], [0, 3], [1, 3], [2, 3], [2, 4], [3, 4], [3, 5], [3, 6], [4, 6], [4, 8], [5, 6], [5, 7], [6, 7], [6, 8], [7, 8]];
  return <svg className="ac-network-art" viewBox="0 0 440 215" aria-hidden="true"><g stroke="currentColor" fill="none">{edges.map(([a, b], index) => <path key={index} opacity={index % 3 === 0 ? .6 : .3} strokeWidth={index % 3 === 0 ? 2 : 1} d={`M${nodes[a][0]} ${nodes[a][1]} Q ${(nodes[a][0] + nodes[b][0]) / 2 + 8} ${(nodes[a][1] + nodes[b][1]) / 2 - 12} ${nodes[b][0]} ${nodes[b][1]}`} />)}{nodes.map(([x, y], index) => <g key={index}><circle cx={x} cy={y} r={index === 3 ? 15 : 9} fill="var(--ac-surface)" strokeWidth="1.5" /><circle cx={x} cy={y} r={index === 3 ? 6 : 3.5} fill="currentColor" stroke="none" /></g>)}</g></svg>;
}
