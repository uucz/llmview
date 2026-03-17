<script lang="ts">
  import { session, connected, appConfig, budgetExceeded } from '$lib/stores/events';
  import { theme, toggleTheme } from '$lib/stores/theme';
  import { formatCost, formatTokens } from '$lib/utils/format';

  let budgetPct = $derived($appConfig.budget > 0 ? Math.min(($session.total_cost / $appConfig.budget) * 100, 100) : 0);
  let budgetColor = $derived(budgetPct >= 90 ? 'var(--risk-danger)' : budgetPct >= 70 ? 'var(--brand-orange)' : 'var(--brand-green)');
</script>

<header>
  <div class="left">
    <div class="logo">
      <!-- Eye icon — line-art, stroke only, represents "view" in llmview -->
      <div class="logo-icon">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" width="20" height="20">
          <path d="M2.06 12a10.47 10.47 0 0 1 19.88 0 10.47 10.47 0 0 1-19.88 0z" />
          <circle cx="12" cy="12" r="3" />
        </svg>
      </div>
      <span class="logo-text">llmview</span>
    </div>
    <div class="status" class:online={$connected}>
      <span class="dot"></span>
      <span class="status-label">{$connected ? 'Live' : 'Disconnected'}</span>
    </div>
  </div>

  <div class="right-group">
    <div class="stats">
      <div class="stat green">
        <div class="stat-bar"></div>
        <div class="stat-icon">
          <!-- Dollar/cost icon -->
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" width="16" height="16">
            <line x1="12" y1="2" x2="12" y2="22" />
            <path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6" />
          </svg>
        </div>
        <div class="stat-data">
          <span class="stat-value">{formatCost($session.total_cost)}</span>
          <span class="stat-label">COST</span>
        </div>
      </div>
      <div class="stat blue">
        <div class="stat-bar"></div>
        <div class="stat-icon">
          <!-- Zap/token icon -->
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" width="16" height="16">
            <polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2" />
          </svg>
        </div>
        <div class="stat-data">
          <span class="stat-value">{formatTokens($session.total_tokens)}</span>
          <span class="stat-label">TOKENS</span>
        </div>
      </div>
      <div class="stat orange">
        <div class="stat-bar"></div>
        <div class="stat-icon">
          <!-- Hash/request count icon -->
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" width="16" height="16">
            <line x1="4" y1="9" x2="20" y2="9" />
            <line x1="4" y1="15" x2="20" y2="15" />
            <line x1="10" y1="3" x2="8" y2="21" />
            <line x1="16" y1="3" x2="14" y2="21" />
          </svg>
        </div>
        <div class="stat-data">
          <span class="stat-value">{$session.request_count}</span>
          <span class="stat-label">REQUESTS</span>
        </div>
      </div>
    </div>

    {#if $appConfig.budget > 0}
      <div class="budget-pill" class:exceeded={$budgetExceeded} title="Budget: {formatCost($session.total_cost)} / {formatCost($appConfig.budget)}">
        <div class="budget-track">
          <div class="budget-fill" style="width: {budgetPct}%; background: {budgetColor}"></div>
        </div>
        <span class="budget-text" style="color: {budgetColor}">
          {formatCost($session.total_cost)}<span class="budget-sep">/</span>{formatCost($appConfig.budget)}
        </span>
      </div>
    {/if}

    <!-- Theme toggle -->
    <button class="theme-toggle" onclick={toggleTheme} title="Toggle theme">
      {#if $theme === 'light'}
        <!-- Moon icon -->
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" width="18" height="18">
          <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" />
        </svg>
      {:else}
        <!-- Sun icon -->
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" width="18" height="18">
          <circle cx="12" cy="12" r="5" />
          <line x1="12" y1="1" x2="12" y2="3" />
          <line x1="12" y1="21" x2="12" y2="23" />
          <line x1="4.22" y1="4.22" x2="5.64" y2="5.64" />
          <line x1="18.36" y1="18.36" x2="19.78" y2="19.78" />
          <line x1="1" y1="12" x2="3" y2="12" />
          <line x1="21" y1="12" x2="23" y2="12" />
          <line x1="4.22" y1="19.78" x2="5.64" y2="18.36" />
          <line x1="18.36" y1="5.64" x2="19.78" y2="4.22" />
        </svg>
      {/if}
    </button>
  </div>
</header>

<style>
  header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0 24px;
    height: var(--header-h);
    border-bottom: 1px solid var(--border-color);
    background: rgba(250, 249, 245, 0.88);
    backdrop-filter: blur(10px);
    -webkit-backdrop-filter: blur(10px);
    position: sticky;
    top: 0;
    z-index: 100;
    transition: background 0.25s, border-color 0.25s;
  }

  :global([data-theme="dark"]) header {
    background: rgba(20, 20, 19, 0.88);
  }

  .left {
    display: flex;
    align-items: center;
    gap: 18px;
  }

  .logo {
    display: flex;
    align-items: center;
    gap: 9px;
  }

  .logo-icon {
    width: 32px;
    height: 32px;
    border-radius: var(--radius-sm);
    background: var(--brand-orange);
    display: flex;
    align-items: center;
    justify-content: center;
    color: #141413;
  }

  .logo-text {
    font-family: var(--font-heading);
    font-size: 16px;
    font-weight: 700;
    letter-spacing: -0.5px;
    color: var(--text-primary);
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
    background: var(--risk-danger);
    transition: background 0.3s, box-shadow 0.3s;
  }

  .status.online .dot {
    background: var(--risk-safe);
    box-shadow: 0 0 8px rgba(120, 140, 93, 0.5);
    animation: pulse 2.5s ease-in-out infinite;
  }

  .status-label {
    font-family: var(--font-heading);
    font-size: 10px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: var(--text-tertiary);
  }

  .right-group {
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .stats {
    display: flex;
    gap: 8px;
  }

  .stat {
    position: relative;
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 7px 14px;
    border-radius: var(--radius-sm);
    background: var(--surface-1);
    border: 1px solid var(--border-color);
    overflow: hidden;
    transition: border-color 0.2s, transform 0.2s;
  }

  .stat:hover {
    border-color: rgba(217, 119, 87, 0.2);
    transform: translateY(-1px);
  }

  /* Left color accent bar */
  .stat-bar {
    position: absolute;
    left: 0;
    top: 0;
    bottom: 0;
    width: 3px;
  }

  .stat.green .stat-bar  { background: var(--brand-green); }
  .stat.blue .stat-bar   { background: var(--brand-blue); }
  .stat.orange .stat-bar { background: var(--brand-orange); }

  .stat-icon {
    width: 32px;
    height: 32px;
    border-radius: var(--radius-sm);
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  .stat.green .stat-icon  { background: rgba(120, 140, 93, 0.10); color: var(--brand-green); }
  .stat.blue .stat-icon   { background: rgba(106, 155, 204, 0.10); color: var(--brand-blue); }
  .stat.orange .stat-icon { background: rgba(217, 119, 87, 0.10); color: var(--brand-orange); }

  .stat-data {
    display: flex;
    flex-direction: column;
  }

  .stat-value {
    font-family: var(--font-heading);
    font-size: 17px;
    font-weight: 700;
    color: var(--text-primary);
    font-variant-numeric: tabular-nums;
    line-height: 1.2;
    letter-spacing: -0.02em;
  }

  .stat-label {
    font-family: var(--font-heading);
    font-size: 9px;
    font-weight: 600;
    color: var(--text-tertiary);
    letter-spacing: 0.1em;
  }

  /* Budget pill */
  .budget-pill {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 6px 12px;
    background: var(--surface-1);
    border: 1px solid var(--border-color);
    border-radius: var(--radius-sm);
    transition: border-color 0.2s;
  }

  .budget-pill.exceeded {
    border-color: rgba(192, 90, 60, 0.4);
    animation: budgetPulse 2s ease-in-out infinite;
  }

  @keyframes budgetPulse {
    0%, 100% { box-shadow: none; }
    50% { box-shadow: 0 0 8px rgba(192, 90, 60, 0.3); }
  }

  .budget-track {
    width: 48px;
    height: 4px;
    background: var(--surface-3);
    border-radius: 2px;
    overflow: hidden;
  }

  .budget-fill {
    height: 100%;
    border-radius: 2px;
    transition: width 0.4s ease, background 0.3s;
  }

  .budget-text {
    font-family: var(--font-mono);
    font-size: 10px;
    font-weight: 600;
    font-variant-numeric: tabular-nums;
  }

  .budget-sep {
    color: var(--text-tertiary);
    margin: 0 1px;
  }

  .theme-toggle {
    width: 34px;
    height: 34px;
    border-radius: var(--radius-sm);
    background: transparent;
    border: 1px solid var(--border-color);
    color: var(--text-secondary);
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: background 0.15s, border-color 0.2s, color 0.15s;
  }

  .theme-toggle:hover {
    background: var(--hover-overlay);
    border-color: var(--border-focus);
    color: var(--brand-orange);
  }
</style>
