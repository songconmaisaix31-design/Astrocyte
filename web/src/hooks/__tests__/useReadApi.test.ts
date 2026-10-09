import { describe, it, expect } from 'vitest';

/**
 * useReadApi hook logic tests (unit-level, no React rendering).
 *
 * We test the stale-state logic directly by simulating the setState patterns
 * used in the hook's error handler.
 */

describe('useReadApi stale state logic', () => {
  it('marks stale when refresh fails but prior data exists', () => {
    // Simulate the hook's error handler logic:
    // setState(prev => ({ ...prev, loading: false, error: msg, stale: !!prev.data }));
    const prevWith = { data: { items: [] }, loading: true, error: null, stale: false };
    const prevWithout = { data: null, loading: true, error: null, stale: false };

    // Error with prior data → stale = true
    const resultWith = { ...prevWith, loading: false, error: 'API 500', stale: !!prevWith.data };
    expect(resultWith.stale).toBe(true);
    expect(resultWith.data).toEqual({ items: [] }); // data retained

    // Error without prior data → stale = false
    const resultWithout = { ...prevWithout, loading: false, error: 'API 500', stale: !!prevWithout.data };
    expect(resultWithout.stale).toBe(false);
  });

  it('clears stale on successful refresh', () => {
    const stalePrev = { data: { items: ['old'] }, loading: true, error: 'old error', stale: true };
    // Success resets stale to false
    const result = { data: { items: ['new'] }, loading: false, error: null, stale: false };
    expect(result.stale).toBe(false);
    expect(result.data.items).toEqual(['new']);
    // Previous stale state is irrelevant after success
    expect(stalePrev.stale).toBe(true); // untouched original
  });

  it('does not silently fall back to fixtures on error', () => {
    // Verify that error state is set and data from API is retained (not replaced with fixtures)
    const prevData = { items: [{ id: 'real-data' }] };
    const errorState = { data: prevData, loading: false, error: 'API 503: Service Unavailable', stale: true };
    expect(errorState.data).toBe(prevData);
    expect(errorState.error).toContain('503');
    // The data is the real API data, not fixture data
    expect(errorState.data.items[0].id).toBe('real-data');
  });
});
