import { describe, it, expect } from 'vitest';
import {
  isFixtureMode,
  fixtureMaterials,
  fixtureOpportunities,
  fixtureProjects,
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
});

describe('fixture materials conform to V1 schema', () => {
  it('have required fields', () => {
    for (const m of fixtureMaterials) {
      expect(m.id).toBeTruthy();
      expect(m.source_locator).toBeTruthy();
      expect(['paper', 'video', 'text', 'file']).toContain(m.kind);
      expect(typeof m.current_revision).toBe('number');
      expect(['active', 'archived', 'withdrawn']).toContain(m.lifecycle);
    }
  });

  it('includes edge cases: long Chinese titles', () => {
    expect(fixtureMaterials.some(m => (m.title?.length ?? 0) > 30)).toBe(true);
  });

  it('includes failed import status', () => {
    expect(fixtureMaterials.some(m => m.import_status === 'failed')).toBe(true);
  });

  it('includes video kind for summarize direction', () => {
    expect(fixtureMaterials.some(m => m.kind === 'video')).toBe(true);
  });

  it('includes archived lifecycle', () => {
    expect(fixtureMaterials.some(m => m.lifecycle === 'archived')).toBe(true);
  });

  it('includes queued import status', () => {
    expect(fixtureMaterials.some(m => m.import_status === 'queued')).toBe(true);
  });

  it('includes null optional fields', () => {
    expect(fixtureMaterials.some(m => m.collection_reason === null)).toBe(true);
    expect(fixtureMaterials.some(m => m.source_spans === undefined)).toBe(true);
  });
});

describe('fixture opportunities conform to V1 schema', () => {
  it('have required fields', () => {
    for (const o of fixtureOpportunities) {
      expect(o.id).toBeTruthy();
      expect(typeof o.revision).toBe('number');
      expect(['incubating', 'ready_for_review', 'admitted', 'deferred', 'rejected', 'withdrawn']).toContain(o.state);
      expect(o.dimensions).toBeDefined();
      expect(o.dimensions.goal_progress).toBeDefined();
      expect(o.dimensions.current_interest).toBeDefined();
      expect(o.dimensions.project_improvement).toBeDefined();
      expect(o.dimensions.originality).toBeDefined();
      expect(o.next_step).toBeTruthy();
    }
  });

  it('includes null dimension values (unknown)', () => {
    expect(fixtureOpportunities.some(o => o.dimensions.goal_progress.value === null)).toBe(true);
  });

  it('includes various states', () => {
    const states = new Set(fixtureOpportunities.map(o => o.state));
    expect(states.has('incubating')).toBe(true);
    expect(states.has('deferred')).toBe(true);
    expect(states.has('rejected')).toBe(true);
  });
});

describe('fixture projects conform to V1 schema', () => {
  it('have required fields', () => {
    for (const p of fixtureProjects) {
      expect(p.id).toBeTruthy();
      expect(p.name).toBeTruthy();
      expect(p.environment_id).toBeTruthy();
      expect(p.root_path).toBeTruthy();
    }
  });

  it('includes long Chinese names', () => {
    expect(fixtureProjects.some(p => p.name.length > 30)).toBe(true);
  });
});

describe('fixture proposals conform to V1 schema', () => {
  it('have required fields', () => {
    for (const p of fixtureProposals) {
      expect(p.id).toBeTruthy();
      expect(p.title).toBeTruthy();
      expect(p.goal).toBeTruthy();
      expect(['draft', 'in_review', 'approved', 'declined', 'superseded']).toContain(p.status);
      expect(p.scope).toBeDefined();
      expect(Array.isArray(p.deliverables)).toBe(true);
      expect(p.budget).toBeDefined();
      expect(Array.isArray(p.stop_conditions)).toBe(true);
    }
  });

  it('includes declined status', () => {
    expect(fixtureProposals.some(p => p.status === 'declined')).toBe(true);
  });
});

describe('fixture sessions conform to V1 schema', () => {
  it('have required fields', () => {
    for (const s of fixtureSessions) {
      expect(s.id).toBeTruthy();
      expect(s.adapter).toBeTruthy();
      expect(s.project_id).toBeTruthy();
      expect(['observed', 'bound', 'unavailable']).toContain(s.binding_status);
      expect(['unknown', 'current', 'stale']).toContain(s.context_state);
      expect(s.capabilities).toBeDefined();
    }
  });

  it('includes stale context state', () => {
    expect(fixtureSessions.some(s => s.context_state === 'stale')).toBe(true);
  });
});

describe('fixture missions conform to V1 schema', () => {
  it('have required fields', () => {
    for (const m of fixtureMissions) {
      expect(m.id).toBeTruthy();
      expect(m.goal).toBeTruthy();
      expect(typeof m.version).toBe('number');
      expect(m.grant_id).toBeTruthy();
      expect(['pending', 'running', 'paused', 'blocked', 'completed', 'cancelled', 'failed']).toContain(m.status);
    }
  });

  it('includes failed and blocked statuses', () => {
    expect(fixtureMissions.some(m => m.status === 'failed')).toBe(true);
    expect(fixtureMissions.some(m => m.status === 'blocked')).toBe(true);
  });

  it('includes nested work items', () => {
    const withWorkItems = fixtureMissions.filter(m => (m.work_items?.length ?? 0) > 0);
    expect(withWorkItems.length).toBeGreaterThan(0);
  });
});

describe('flattenWorkItems', () => {
  it('extracts all work items from missions', () => {
    const items = flattenWorkItems(fixtureMissions);
    expect(items.length).toBeGreaterThan(0);
    for (const wi of items) {
      expect(wi.id).toBeTruthy();
      expect(['open', 'claimed', 'in_progress', 'submitted', 'revision_needed', 'blocked', 'accepted', 'cancelled']).toContain(wi.status);
    }
  });
});

describe('flattenArtifactIds', () => {
  it('extracts all artifact IDs from missions', () => {
    const ids = flattenArtifactIds(fixtureMissions);
    expect(ids.length).toBeGreaterThan(0);
    for (const id of ids) {
      expect(id).toBeTruthy();
    }
  });
});
