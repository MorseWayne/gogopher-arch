import { describe, expect, it } from 'vitest';
import { formatDuration } from './formatDuration';

describe('formatDuration', () => {
  it('returns -- for zero or negative', () => {
    expect(formatDuration(0)).toBe('--');
    expect(formatDuration(-1)).toBe('--');
  });

  it('formats nanoseconds to milliseconds', () => {
    expect(formatDuration(1_500_000)).toBe('1.50ms');
    expect(formatDuration(12_340_000)).toBe('12.34ms');
  });
});
