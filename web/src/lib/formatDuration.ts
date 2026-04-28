export function formatDuration(duration: number): string {
  if (duration <= 0) {
    return '--';
  }
  return `${(duration / 1_000_000).toFixed(2)}ms`;
}
