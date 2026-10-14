import { describe, expect, it } from 'vitest';
import type { WorksSnapshot } from './worksAlbumPresentation';
import { relayAlbumMetrics } from './relayAlbumMetrics';

// 参数只验证统计函数，不表示存在实际目录、手册或派发结果。
function inputSnapshot(): WorksSnapshot {
  return { status: 'complete', observed_at: '2026-10-14T00:00:00Z', failures: [], platforms: [{ platform: 'input-cli', works: 1, folders: 1, handed_over: 1, recent: 0 }], works: [{
    id: 'work-a', platform: 'input-cli', name: '输入', root: '输入路径', kind: 'folder', source: 'scan', attitude: 'continue', group: '', hidden: false,
    handoff: { content: '输入手册', updated_at: null, source: 'human' }, files: { source_path: '输入路径', mtime: null, file_count: 0, verified: false },
    updated_at: '2026-10-14T00:00:00Z', revision: 1,
  }] };
}

describe('relay album fact boundaries', () => {
  it('does not count handbooks as dispatches and deduplicates service evidence', () => {
    const snapshot = inputSnapshot();
    expect(relayAlbumMetrics(snapshot).find(item => item.label === '接力过')?.value).toBeNull();
    expect(relayAlbumMetrics(snapshot, ['work-a', 'work-a', 'foreign']).find(item => item.label === '接力过')?.value).toBe(1);
  });
  it('keeps counts unknown before scanning and recent unknown with unverified files', () => {
    const snapshot = inputSnapshot();
    expect(relayAlbumMetrics({ ...snapshot, observed_at: null }).every(item => item.value === null)).toBe(true);
    expect(relayAlbumMetrics(snapshot).find(item => item.label === '最近在动')?.value).toBeNull();
  });
  it('uses observed platform data including newly registered platforms', () => {
    const snapshot = inputSnapshot();
    snapshot.platforms = [{ platform: 'new-public-cli', works: 1, folders: 1, handed_over: 1, recent: 0 }];
    expect(relayAlbumMetrics(snapshot)[0].value).toBe(1);
  });
});
