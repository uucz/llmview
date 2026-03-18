<script lang="ts">
  import { insights, fetchInsights } from '$lib/stores/events';
  import { formatCost } from '$lib/utils/format';

  let open = $state(false);
  let loading = $state(false);

  let count = $derived($insights.length);
  let hasCritical = $derived($insights.some(i => i.severity === 'critical'));
  let hasWarning = $derived($insights.some(i => i.severity === 'warning'));

  $effect(() => {
    loading = true;
    fetchInsights().finally(() => { loading = false; });
  });

  function severityColor(s: string): string {
    switch (s) {
      case 'critical': return 'var(--risk-danger)';
      case 'warning': return 'var(--brand-orange)';
      default: return 'var(--brand-blue)';
    }
  }
</script>

{#if count > 0}
  <div class="insights-container" class:open>
    <button class="insights-header" onclick={() => open = !open}>
      <!-- Lightbulb icon -->
      <svg class="header-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" width="14" height="14">
        <path d="M9 18h6" />
        <path d="M10 22h4" />
        <path d="M12 2a7 7 0 0 0-4 12.7V17h8v-2.3A7 7 0 0 0 12 2z" />
      </svg>
      <span class="insights-title">Insights</span>
      <span class="insights-badge" class:critical={hasCritical} class:warning={hasWarning && !hasCritical}>
        {count}
      </span>
      <svg class="chevron" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" width="12" height="12">
        <polyline points="6 9 12 15 18 9" />
      </svg>
    </button>

    {#if open}
      <div class="insights-list">
        {#each $insights as insight, i}
          <div class="insight-card" style="--severity-color: {severityColor(insight.severity)}; animation-delay: {i * 0.05}s">
            <div class="insight-icon" style="color: {severityColor(insight.severity)}">
              {#if insight.type === 'loop_detected'}
                <!-- Refresh/loop icon -->
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" width="14" height="14">
                  <polyline points="23 4 23 10 17 10" />
                  <polyline points="1 20 1 14 7 14" />
                  <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10" />
                  <path d="M20.49 15a9 9 0 0 1-14.85 3.36L1 14" />
                </svg>
              {:else if insight.type === 'prompt_waste'}
                <!-- Scissors/cut icon -->
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" width="14" height="14">
                  <circle cx="6" cy="6" r="3" />
                  <circle cx="6" cy="18" r="3" />
                  <line x1="20" y1="4" x2="8.12" y2="15.88" />
                  <line x1="14.47" y1="14.48" x2="20" y2="20" />
                  <line x1="8.12" y1="8.12" x2="12" y2="12" />
                </svg>
              {:else if insight.type === 'model_downgrade'}
                <!-- Trending-down icon -->
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" width="14" height="14">
                  <polyline points="23 18 13.5 8.5 8.5 13.5 1 6" />
                  <polyline points="17 18 23 18 23 12" />
                </svg>
              {:else}
                <!-- Flame/burn rate icon -->
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" width="14" height="14">
                  <path d="M12 22c4.97 0 8-3.03 8-8 0-4-2.5-7-4-8.5-.5 2.5-2 4.5-4 5.5-2-1-3.5-3.5-3-6.5C5 7.5 4 12 4 14c0 4.97 3.03 8 8 8z" />
                </svg>
              {/if}
            </div>
            <div class="insight-body">
              <div class="insight-title">{insight.title}</div>
              <div class="insight-desc">{insight.description}</div>
            </div>
            {#if insight.savings && insight.savings > 0}
              <div class="insight-savings">
                <span class="savings-amount">-{formatCost(insight.savings)}</span>
                {#if insight.token_savings}
                  <span class="savings-tokens">{insight.token_savings.toLocaleString()} tokens</span>
                {/if}
              </div>
            {/if}
          </div>
        {/each}
      </div>
    {/if}
  </div>
{/if}

<style>
  .insights-container {
    margin-bottom: 10px;
    border: 1px solid var(--border-color);
    border-radius: var(--radius-sm);
    background: var(--surface-1);
    overflow: hidden;
    animation: fadeIn 0.3s ease;
    transition: border-color 0.2s;
  }

  .insights-container:hover {
    border-color: var(--border-focus);
  }

  .insights-header {
    display: flex;
    align-items: center;
    gap: 8px;
    width: 100%;
    padding: 8px 12px;
    background: transparent;
    border: none;
    cursor: pointer;
    color: var(--text-secondary);
    transition: background 0.15s;
  }

  .insights-header:hover {
    background: var(--hover-overlay);
  }

  .header-icon {
    color: var(--brand-orange);
    flex-shrink: 0;
  }

  .insights-title {
    font-family: var(--font-heading);
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    color: var(--text-secondary);
  }

  .insights-badge {
    font-family: var(--font-mono);
    font-size: 10px;
    font-weight: 600;
    min-width: 18px;
    height: 18px;
    padding: 0 5px;
    border-radius: 9px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(106, 155, 204, 0.12);
    color: var(--brand-blue);
    line-height: 1;
  }

  .insights-badge.critical {
    background: rgba(192, 90, 60, 0.12);
    color: var(--risk-danger);
  }

  .insights-badge.warning {
    background: rgba(217, 119, 87, 0.12);
    color: var(--brand-orange);
  }

  .chevron {
    margin-left: auto;
    color: var(--text-tertiary);
    transition: transform 0.2s ease;
    flex-shrink: 0;
  }

  .open .chevron {
    transform: rotate(180deg);
  }

  .insights-list {
    border-top: 1px solid var(--border-subtle);
    padding: 4px;
    display: flex;
    flex-direction: column;
    gap: 2px;
    max-height: 240px;
    overflow-y: auto;
  }

  .insight-card {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 7px 10px;
    border-radius: 4px;
    transition: background 0.12s;
    animation: fadeUp 0.25s ease both;
  }

  .insight-card:hover {
    background: var(--hover-overlay);
  }

  .insight-icon {
    width: 28px;
    height: 28px;
    border-radius: var(--radius-sm);
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    background: color-mix(in srgb, currentColor 8%, transparent);
  }

  .insight-body {
    flex: 1;
    min-width: 0;
  }

  .insight-title {
    font-family: var(--font-heading);
    font-size: 12px;
    font-weight: 600;
    color: var(--text-primary);
    line-height: 1.3;
  }

  .insight-desc {
    font-family: var(--font-body);
    font-size: 11px;
    color: var(--text-tertiary);
    line-height: 1.4;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .insight-savings {
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    flex-shrink: 0;
    gap: 1px;
  }

  .savings-amount {
    font-family: var(--font-mono);
    font-size: 12px;
    font-weight: 600;
    color: var(--brand-green);
    font-variant-numeric: tabular-nums;
  }

  .savings-tokens {
    font-family: var(--font-mono);
    font-size: 9px;
    color: var(--text-tertiary);
    font-variant-numeric: tabular-nums;
  }
</style>
