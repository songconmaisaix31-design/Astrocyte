import { describe, expect, it } from 'vitest';
import { fixtureWorksSnapshot } from './worksAlbumPresentation';
import { relayAlbumMetrics } from './relayAlbumMetrics';

describe('relay album fact boundaries', () => {
  it('does not count handbooks as dispatches and deduplicates service evidence', () => {
    const snapshot = fixtureWorksSnapshot();
    expect(relayAlbumMetrics(snapshot).find(item => item.label === '接力过')?.value).toBeNull();
    expect(relayAlbumMetrics(snapshot, ['work-a', 'work-a', 'foreign']).find(item => item.label === '接力过')?.value).toBe(1);
  });
  it('keeps counts unknown before scanning and recent unknown with unverified files', () => {
    const snapshot = fixtureWorksSnapshot();
    expect(relayAlbumMetrics({ ...snapshot, observed_at: null }).every(item => item.value === null)).toBe(true);
    expect(relayAlbumMetrics(snapshot).find(item => item.label === '最近在动')?.value).toBeNull();
  });
  it('uses observed platform data including newly registered platforms', () => {
    const snapshot = fixtureWorksSnapshot();
    snapshot.platforms = [{ platform: 'new-public-cli', works: 1, folders: 1, handed_over: 1, recent: 0 }];
    expect(relayAlbumMetrics(snapshot)[0].value).toBe(1);
  });
});
