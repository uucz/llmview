export function formatDuration(ms: number): string {
  if (ms === 0) return '...';
  if (ms < 1000) return `${ms}ms`;
  if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
  return `${(ms / 60000).toFixed(1)}m`;
}

export function formatCost(cost: number): string {
  if (cost === 0) return 'free';
  if (cost < 0.001) return `$${cost.toFixed(5)}`;
  if (cost < 0.01) return `$${cost.toFixed(4)}`;
  if (cost < 1) return `$${cost.toFixed(3)}`;
  return `$${cost.toFixed(2)}`;
}

export function formatTokens(n: number): string {
  if (n === 0) return '0';
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`;
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}k`;
  return n.toString();
}

export function formatTime(ts: number): string {
  const d = new Date(ts);
  return d.toLocaleTimeString('en-US', { hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit' });
}

export function providerColor(provider: string): string {
  switch (provider) {
    case 'openai': return '#10a37f';
    case 'anthropic': return '#d4a274';
    case 'ollama': return '#888888';
    default: return '#666666';
  }
}

export function statusColor(code: number): string {
  if (code === 0) return 'var(--text-2)';
  if (code >= 200 && code < 300) return 'var(--green)';
  if (code >= 400 && code < 500) return 'var(--orange)';
  return 'var(--red)';
}
