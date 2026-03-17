<script lang="ts">
  import { session, connected } from '$lib/stores/events';
  import { formatCost, formatTokens } from '$lib/utils/format';
</script>

<header>
  <div class="left">
    <div class="logo">
      <span class="logo-icon">◉</span>
      <span class="logo-text">llmview</span>
    </div>
    <div class="status" class:online={$connected}>
      <span class="dot"></span>
      {$connected ? 'Live' : 'Disconnected'}
    </div>
  </div>

  <div class="stats">
    <div class="stat">
      <span class="stat-value cost">{formatCost($session.total_cost)}</span>
      <span class="stat-label">Cost</span>
    </div>
    <div class="stat">
      <span class="stat-value">{formatTokens($session.total_tokens)}</span>
      <span class="stat-label">Tokens</span>
    </div>
    <div class="stat">
      <span class="stat-value">{$session.request_count}</span>
      <span class="stat-label">Requests</span>
    </div>
  </div>
</header>

<style>
  header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 20px;
    border-bottom: 1px solid var(--border);
    background: var(--bg-1);
    position: sticky;
    top: 0;
    z-index: 100;
    backdrop-filter: blur(12px);
  }

  .left {
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .logo {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .logo-icon {
    font-size: 20px;
    color: var(--accent);
  }

  .logo-text {
    font-size: 16px;
    font-weight: 700;
    letter-spacing: -0.5px;
    color: var(--text-0);
  }

  .status {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 11px;
    color: var(--text-2);
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--red);
  }

  .status.online .dot {
    background: var(--green);
    box-shadow: 0 0 6px var(--green);
    animation: pulse 2s infinite;
  }

  @keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.5; }
  }

  .stats {
    display: flex;
    gap: 28px;
  }

  .stat {
    display: flex;
    flex-direction: column;
    align-items: flex-end;
  }

  .stat-value {
    font-size: 20px;
    font-weight: 700;
    color: var(--text-0);
    font-variant-numeric: tabular-nums;
    line-height: 1.2;
  }

  .stat-value.cost {
    color: var(--green);
  }

  .stat-label {
    font-size: 10px;
    color: var(--text-2);
    text-transform: uppercase;
    letter-spacing: 1px;
  }
</style>
