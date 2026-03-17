<script lang="ts">
  import { session, connected } from '$lib/stores/events';
  import { formatCost, formatTokens } from '$lib/utils/format';
</script>

<header>
  <div class="left">
    <div class="logo">
      <div class="logo-mark"></div>
      <span class="logo-text">llmview</span>
    </div>
    <div class="status" class:online={$connected}>
      <span class="dot"></span>
      <span class="status-text">{$connected ? 'Live' : 'Disconnected'}</span>
    </div>
  </div>

  <div class="stats">
    <div class="stat">
      <div class="stat-indicator green"></div>
      <div class="stat-content">
        <span class="stat-value cost">{formatCost($session.total_cost)}</span>
        <span class="stat-label">COST</span>
      </div>
    </div>
    <div class="stat">
      <div class="stat-indicator accent"></div>
      <div class="stat-content">
        <span class="stat-value">{formatTokens($session.total_tokens)}</span>
        <span class="stat-label">TOKENS</span>
      </div>
    </div>
    <div class="stat">
      <div class="stat-indicator orange"></div>
      <div class="stat-content">
        <span class="stat-value">{$session.request_count}</span>
        <span class="stat-label">REQUESTS</span>
      </div>
    </div>
  </div>
</header>

<style>
  header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0 24px;
    height: 56px;
    border-bottom: 1px solid var(--border);
    background: rgba(14, 14, 13, 0.85);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    position: sticky;
    top: 0;
    z-index: 100;
  }

  .left {
    display: flex;
    align-items: center;
    gap: 18px;
  }

  .logo {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .logo-mark {
    width: 22px;
    height: 22px;
    border-radius: var(--radius-sm);
    background: var(--accent);
    opacity: 0.9;
  }

  .logo-text {
    font-family: var(--font-sans);
    font-size: 15px;
    font-weight: 700;
    letter-spacing: -0.5px;
    color: var(--text-0);
  }

  .status {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--red);
    transition: background 0.3s, box-shadow 0.3s;
  }

  .status.online .dot {
    background: var(--green);
    box-shadow: 0 0 8px rgba(102, 217, 142, 0.4);
    animation: pulse 2.5s ease-in-out infinite;
  }

  .status-text {
    font-family: var(--font-sans);
    font-size: 10px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: var(--text-2);
  }

  .stats {
    display: flex;
    gap: 8px;
  }

  .stat {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 6px 14px;
    border-radius: var(--radius-sm);
    background: var(--surface-1);
    border: 1px solid var(--border);
    transition: border-color 0.2s;
  }

  .stat:hover {
    border-color: var(--border-hover);
  }

  .stat-indicator {
    width: 3px;
    height: 24px;
    border-radius: 2px;
  }

  .stat-indicator.green { background: var(--green); }
  .stat-indicator.accent { background: var(--accent); }
  .stat-indicator.orange { background: var(--orange); }

  .stat-content {
    display: flex;
    flex-direction: column;
  }

  .stat-value {
    font-size: 17px;
    font-weight: 700;
    color: var(--text-0);
    font-variant-numeric: tabular-nums;
    line-height: 1.2;
    letter-spacing: -0.02em;
  }

  .stat-value.cost {
    color: var(--green);
  }

  .stat-label {
    font-family: var(--font-sans);
    font-size: 9px;
    font-weight: 600;
    color: var(--text-2);
    letter-spacing: 0.1em;
  }
</style>
