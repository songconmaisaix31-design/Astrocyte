import { describe, it, expect } from 'vitest';
import {
  formatDate,
  formatDateTime,
  formatBytes,
  formatDimScore,
  formatTokens,
  formatBudget,
  lifecycleLabel,
  importStatusLabel,
  opportunityStateLabel,
  proposalStatusLabel,
  missionStatusLabel,
  workItemStatusLabel,
  bindingStatusLabel,
  contextStateLabel,
  materialKindLabel,
  capabilityLabel,
} from '../format';

describe('formatDate', () => {
  it('formats valid ISO date', () => {
    const result = formatDate('2025-10-09T08:30:00Z');
    expect(result).toContain('2025');
    expect(result).toContain('10');
    expect(result).toContain('09');
  });

  it('returns dash for null', () => {
    expect(formatDate(null)).toBe('—');
  });

  it('returns dash for undefined', () => {
    expect(formatDate(undefined)).toBe('—');
  });

  it('returns dash for invalid date string', () => {
    expect(formatDate('not-a-date')).toBe('—');
  });
});

describe('formatDateTime', () => {
  it('formats valid ISO datetime with time', () => {
    const result = formatDateTime('2025-10-09T14:30:00Z');
    expect(result).toContain('2025');
    expect(result.length).toBeGreaterThan(10);
  });

  it('returns dash for null', () => {
    expect(formatDateTime(null)).toBe('—');
  });
});

describe('formatBytes', () => {
  it('returns dash for null', () => {
    expect(formatBytes(null)).toBe('—');
  });

  it('returns dash for undefined', () => {
    expect(formatBytes(undefined)).toBe('—');
  });

  it('formats 0 as "0 B"', () => {
    expect(formatBytes(0)).toBe('0 B');
  });

  it('formats bytes under 1KB', () => {
    expect(formatBytes(500)).toBe('500 B');
  });

  it('formats kilobytes', () => {
    expect(formatBytes(1024)).toBe('1.0 KB');
    expect(formatBytes(245760)).toBe('240.0 KB');
  });

  it('formats megabytes', () => {
    expect(formatBytes(1048576)).toBe('1.0 MB');
  });

  it('formats gigabytes', () => {
    expect(formatBytes(1073741824)).toBe('1.0 GB');
  });
});

describe('formatDimScore', () => {
  it('returns dash for null', () => {
    expect(formatDimScore(null)).toBe('—');
  });

  it('converts 0-1 to percentage', () => {
    expect(formatDimScore(0.82)).toBe('82%');
    expect(formatDimScore(0)).toBe('0%');
    expect(formatDimScore(1)).toBe('100%');
    expect(formatDimScore(0.455)).toBe('46%');
  });
});

describe('formatTokens', () => {
  it('returns dash for null', () => {
    expect(formatTokens(null)).toBe('—');
  });

  it('formats with locale separators', () => {
    const result = formatTokens(45200);
    expect(result).toMatch(/45.*200|45200/);
  });
});

describe('formatBudget', () => {
  it('returns dash for null', () => {
    expect(formatBudget(null)).toBe('—');
  });

  it('formats budget with all fields', () => {
    const result = formatBudget({ calls: 100, tokens: 500000, money: 50, currency: 'USD' });
    expect(result).toContain('100');
    expect(result).toContain('500');
    expect(result).toContain('50');
  });

  it('formats budget with null fields', () => {
    expect(formatBudget({ calls: null, tokens: null, money: null, currency: null })).toBe('—');
  });
});

describe('lifecycleLabel', () => {
  it('translates known values', () => {
    expect(lifecycleLabel('active')).toBe('活跃');
    expect(lifecycleLabel('archived')).toBe('已归档');
    expect(lifecycleLabel('withdrawn')).toBe('已撤回');
  });
  it('returns raw for unknown', () => {
    expect(lifecycleLabel('xyz')).toBe('xyz');
  });
});

describe('importStatusLabel', () => {
  it('translates known values', () => {
    expect(importStatusLabel('queued')).toBe('排队中');
    expect(importStatusLabel('succeeded')).toBe('已完成');
    expect(importStatusLabel('failed')).toBe('失败');
  });
});

describe('opportunityStateLabel', () => {
  it('translates known values', () => {
    expect(opportunityStateLabel('incubating')).toBe('孵化中');
    expect(opportunityStateLabel('ready_for_review')).toBe('待审核');
    expect(opportunityStateLabel('rejected')).toBe('已拒绝');
  });
});

describe('proposalStatusLabel', () => {
  it('translates known values', () => {
    expect(proposalStatusLabel('draft')).toBe('草稿');
    expect(proposalStatusLabel('in_review')).toBe('审核中');
    expect(proposalStatusLabel('approved')).toBe('已通过');
    expect(proposalStatusLabel('declined')).toBe('已拒绝');
    expect(proposalStatusLabel('superseded')).toBe('已替代');
  });
});

describe('missionStatusLabel', () => {
  it('translates known values', () => {
    expect(missionStatusLabel('pending')).toBe('等待中');
    expect(missionStatusLabel('running')).toBe('执行中');
    expect(missionStatusLabel('blocked')).toBe('已阻塞');
    expect(missionStatusLabel('completed')).toBe('已完成');
    expect(missionStatusLabel('failed')).toBe('失败');
    expect(missionStatusLabel('paused')).toBe('已暂停');
  });
});

describe('workItemStatusLabel', () => {
  it('translates known values', () => {
    expect(workItemStatusLabel('open')).toBe('开放');
    expect(workItemStatusLabel('claimed')).toBe('已认领');
    expect(workItemStatusLabel('in_progress')).toBe('进行中');
    expect(workItemStatusLabel('accepted')).toBe('已通过');
    expect(workItemStatusLabel('revision_needed')).toBe('需修改');
  });
});

describe('bindingStatusLabel', () => {
  it('translates known values', () => {
    expect(bindingStatusLabel('bound')).toBe('已绑定');
    expect(bindingStatusLabel('unavailable')).toBe('不可用');
    expect(bindingStatusLabel('observed')).toBe('已观察');
  });
});

describe('contextStateLabel', () => {
  it('translates known values', () => {
    expect(contextStateLabel('current')).toBe('当前');
    expect(contextStateLabel('stale')).toBe('过期');
    expect(contextStateLabel('unknown')).toBe('未知');
  });
});

describe('materialKindLabel', () => {
  it('translates known values', () => {
    expect(materialKindLabel('paper')).toBe('论文');
    expect(materialKindLabel('video')).toBe('视频');
    expect(materialKindLabel('text')).toBe('文本');
    expect(materialKindLabel('file')).toBe('文件');
  });
});

describe('capabilityLabel', () => {
  it('returns 可用 for true', () => {
    expect(capabilityLabel(true)).toBe('● 可用');
  });

  it('returns 不可用 for false', () => {
    expect(capabilityLabel(false)).toBe('○ 不可用');
  });

  it('returns 未知 for null (unprobed)', () => {
    expect(capabilityLabel(null)).toBe('◇ 未知');
  });

  it('null and false produce different labels', () => {
    expect(capabilityLabel(null)).not.toBe(capabilityLabel(false));
  });
});
