import { describe, it, expect } from 'vitest';
import {
  isFixtureMode,
  fixtureMaterials,
  fixtureOpportunities,
  fixtureProposals,
  fixtureSessions,
  fixtureMissions,
  flattenWorkItems,
  flattenArtifactIds,
} from '../index';

describe('isFixtureMode', () => {
  it('returns true when fixture=1', () => {
    expect(isFixtureMode(new URLSearchParams('fixture=1'))).toBe(true);
  });

  it('returns false when fixture param absent', () => {
    expect(isFixtureMode(new URLSearchParams())).toBe(false);
  });

  it('returns false for fixture=0', () => {
    expect(isFixtureMode(new URLSearchParams('fixture=0'))).toBe(false);
  });

  it('returns false for empty fixture value', () => {
    expect(isFixtureMode(new URLSearchParams('fixture='))).toBe(false);
  });

  it('returns false for fixture=true (only "1" activates)', () => {
    expect(isFixtureMode(new URLSearchParams('fixture=true'))).toBe(false);
  });
});

describe('source locators are explicitly fixture-prefixed', () => {
  it('all material source_locators start with "fixture:"', () => {
    for (const m of fixtureMaterials) {
      expect(m.source_locator).toMatch(/^fixture:/);
    }
  });

  it('all source ref locators start with "fixture-"', () => {
    const allRefs = fixtureOpportunities.flatMap(o => o.evidence_refs);
    for (const ref of allRefs) {
      expect(ref.locator).toMatch(/^fixture-/);
    }
  });

  it('no invented arXiv, PubMed, DOI, or bioRxiv citations', () => {
    const text = JSON.stringify([...fixtureMaterials, ...fixtureOpportunities]);
    expect(text).not.toMatch(/arxiv:/i);
    expect(text).not.toMatch(/pubmed:/i);
    expect(text).not.toMatch(/10\.\d{4}/);
    expect(text).not.toMatch(/biorxiv/i);
    expect(text).not.toMatch(/doi\.org/i);
  });
});

describe('fixtures exercise edge cases', () => {
  it('includes long Chinese titles (>30 chars) for truncation testing', () => {
    const longTitles = fixtureMaterials.filter(m => (m.title?.length ?? 0) > 30);
    expect(longTitles.length).toBeGreaterThan(0);
  });

  it('includes video kind for summarize direction', () => {
    expect(fixtureMaterials.some(m => m.kind === 'video')).toBe(true);
  });

  it('includes failed import status', () => {
    expect(fixtureMaterials.some(m => m.import_status === 'failed')).toBe(true);
  });

  it('includes archived lifecycle', () => {
    expect(fixtureMaterials.some(m => m.lifecycle === 'archived')).toBe(true);
  });

  it('includes queued import status', () => {
    expect(fixtureMaterials.some(m => m.import_status === 'queued')).toBe(true);
  });

  it('includes null optional fields (collection_reason, source_spans)', () => {
    expect(fixtureMaterials.some(m => m.collection_reason === null)).toBe(true);
    expect(fixtureMaterials.some(m => m.source_spans === undefined)).toBe(true);
  });

  it('includes null dimension values (unknown scores)', () => {
    expect(fixtureOpportunities.some(o => o.dimensions.goal_progress.value === null)).toBe(true);
  });

  it('includes declined proposal status', () => {
    expect(fixtureProposals.some(p => p.status === 'declined')).toBe(true);
  });

  it('includes stale context state in sessions', () => {
    expect(fixtureSessions.some(s => s.context_state === 'stale')).toBe(true);
  });

  it('includes null capability values (unprobed)', () => {
    expect(fixtureSessions.some(s => s.capabilities.native_resume === null)).toBe(true);
  });

  it('includes failed and blocked mission statuses', () => {
    expect(fixtureMissions.some(m => m.status === 'failed')).toBe(true);
    expect(fixtureMissions.some(m => m.status === 'blocked')).toBe(true);
  });

  it('includes missions with blockers', () => {
    expect(fixtureMissions.some(m => (m.blockers?.length ?? 0) > 0)).toBe(true);
  });
});

describe('flattenWorkItems', () => {
  it('extracts work items from missions that have them', () => {
    const items = flattenWorkItems(fixtureMissions);
    expect(items.length).toBeGreaterThan(0);
    // Missions without work_items contribute nothing
    const expected = fixtureMissions
      .filter(m => m.work_items)
      .reduce((sum, m) => sum + (m.work_items?.length ?? 0), 0);
    expect(items.length).toBe(expected);
  });

  it('returns empty for missions with no work items', () => {
    expect(flattenWorkItems([])).toEqual([]);
  });
});

describe('flattenArtifactIds', () => {
  it('extracts artifact IDs from missions', () => {
    const ids = flattenArtifactIds(fixtureMissions);
    expect(ids.length).toBeGreaterThan(0);
    for (const id of ids) {
      expect(id).toMatch(/^fixture-/);
    }
  });

  it('returns empty for missions with no artifacts', () => {
    expect(flattenArtifactIds([])).toEqual([]);
  });
});
