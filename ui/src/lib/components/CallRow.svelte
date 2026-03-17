<script lang="ts">
  import type { APICall } from '$lib/stores/events';
  import { formatDuration, formatCost, formatTokens, formatTime, statusColor } from '$lib/utils/format';
  import DetailPanel from './DetailPanel.svelte';

  let { call, expanded = false, ontoggle, index = 0, compareMode = false, selected = false, focused = false }: {
    call: APICall;
    expanded: boolean;
    ontoggle: () => void;
    index?: number;
    compareMode?: boolean;
    selected?: boolean;
    focused?: boolean;
  } = $props();

  let isStreaming = $derived(!call.completed && call.streaming);

  function providerLabel(p: string): string {
    switch (p) {
      case 'openai': return 'OpenAI';
      case 'anthropic': return 'Anthropic';
      case 'ollama': return 'Ollama';
      default: return p;
    }
  }

  function providerVar(p: string): string {
    switch (p) {
      case 'openai': return 'var(--color-openai)';
      case 'anthropic': return 'var(--color-anthropic)';
      case 'ollama': return 'var(--color-ollama)';
      default: return 'var(--text-tertiary)';
    }
  }
</script>

<div
  class="call"
  class:streaming={isStreaming}
  class:has-error={call.status_code >= 400}
  class:expanded
  class:compare-selected={selected}
  class:focused
  style="animation-delay: {Math.min(index * 0.06, 0.3)}s"
  onclick={ontoggle}
  role="button"
  tabindex="0"
  onkeydown={(e) => e.key === 'Enter' && ontoggle()}
>
  {#if isStreaming}
    <div class="accent-bar streaming-bar"></div>
  {/if}
  {#if call.status_code >= 400}
    <div class="accent-bar error-bar"></div>
  {/if}

  <div class="call-main">
    {#if compareMode}
      <div class="compare-dot" class:active={selected}>
        {#if selected}
          <svg viewBox="0 0 24 24" fill="currentColor" width="14" height="14"><circle cx="12" cy="12" r="6" /></svg>
        {:else}
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="14" height="14"><circle cx="12" cy="12" r="6" /></svg>
        {/if}
      </div>
    {/if}
    <div class="call-left">
      <span class="provider-badge" style="--pcolor: {providerVar(call.provider)}">
        {providerLabel(call.provider)}
      </span>
      <span class="model">{call.model || call.endpoint}</span>
      <span class="time">{formatTime(call.started_at)}</span>
    </div>
    <div class="call-right">
      {#if call.completed}
        <span class="status-badge" style="color: {statusColor(call.status_code)}; background: {statusColor(call.status_code)}18; border-color: {statusColor(call.status_code)}30">
          {call.status_code}
        </span>
        <span class="tokens">
          {formatTokens(call.input_tokens)}
          <svg class="arrow-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" width="12" height="12">
            <line x1="5" y1="12" x2="19" y2="12" />
            <polyline points="12 5 19 12 12 19" />
          </svg>
          {formatTokens(call.output_tokens)}
        </span>
        <span class="duration">{formatDuration(call.duration_ms)}</span>
        <span class="cost" class:free={call.cost === 0}>{formatCost(call.cost)}</span>
      {:else}
        <span class="streaming-indicator">
          <span class="streaming-dot"></span>
          streaming
        </span>
      {/if}
    </div>
  </div>

  {#if expanded}
    <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
    <div class="call-detail" onclick={(e) => e.stopPropagation()}>
      <DetailPanel
        callId={call.id}
        streamText={call.streamText}
        error={call.error || ''}
        completed={call.completed}
      />
    </div>
  {/if}
</div>

<style>
  .call {
    position: relative;
    padding: 12px 16px;
    padding-left: 20px;
    border-radius: var(--radius);
    background: var(--surface-1);
    border: 1px solid var(--border-color);
    cursor: pointer;
    margin-bottom: 6px;
    transition: border-color 0.2s;
    animation: fadeUp 0.35s ease both;
    overflow: hidden;
  }

  .call:hover {
    border-color: rgba(217, 119, 87, 0.2);
  }

  .call.expanded {
    border-color: rgba(217, 119, 87, 0.25);
  }

  .call.focused {
    border-color: rgba(217, 119, 87, 0.3);
    box-shadow: 0 0 0 1px rgba(217, 119, 87, 0.15);
  }

  .call.compare-selected {
    border-color: var(--border-focus);
    background: var(--active-orange-bg);
  }

  .compare-dot {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 20px;
    flex-shrink: 0;
    color: var(--text-tertiary);
    transition: color 0.15s;
  }

  .compare-dot.active {
    color: var(--brand-orange);
  }

  .accent-bar {
    position: absolute;
    left: 0;
    top: 0;
    bottom: 0;
    width: 3px;
  }

  .streaming-bar {
    background: var(--brand-blue);
    animation: streamGlow 2s ease-in-out infinite;
  }

  .error-bar {
    background: var(--risk-danger);
  }

  .call-main {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 12px;
  }

  .call-left {
    display: flex;
    align-items: center;
    gap: 10px;
    min-width: 0;
    flex: 1;
  }

  .call-right {
    display: flex;
    align-items: center;
    gap: 16px;
    flex-shrink: 0;
  }

  .provider-badge {
    font-family: var(--font-heading);
    font-size: 9.5px;
    padding: 2px 8px;
    border-radius: 4px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.03em;
    background: color-mix(in srgb, var(--pcolor) 10%, transparent);
    color: var(--pcolor);
    border: 1px solid color-mix(in srgb, var(--pcolor) 18%, transparent);
    flex-shrink: 0;
    line-height: 1.5;
  }

  .model {
    font-family: var(--font-body);
    font-weight: 500;
    color: var(--text-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    font-size: 13.5px;
  }

  .time {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--text-tertiary);
    flex-shrink: 0;
  }

  .status-badge {
    font-family: var(--font-heading);
    font-weight: 700;
    font-size: 10px;
    padding: 1px 6px;
    border-radius: 4px;
    border: 1px solid;
    text-align: center;
    min-width: 36px;
  }

  .tokens {
    font-family: var(--font-mono);
    color: var(--text-secondary);
    font-size: 12px;
    min-width: 100px;
    text-align: right;
    font-variant-numeric: tabular-nums;
    display: inline-flex;
    align-items: center;
    gap: 2px;
  }

  .arrow-icon {
    color: var(--text-tertiary);
    flex-shrink: 0;
  }

  .duration {
    font-family: var(--font-mono);
    color: var(--text-tertiary);
    font-size: 12px;
    min-width: 52px;
    text-align: right;
    font-variant-numeric: tabular-nums;
  }

  .cost {
    font-family: var(--font-mono);
    color: var(--brand-green);
    font-size: 12px;
    font-weight: 600;
    min-width: 60px;
    text-align: right;
    font-variant-numeric: tabular-nums;
  }

  .cost.free {
    color: var(--text-tertiary);
    font-weight: 400;
  }

  .streaming-indicator {
    display: flex;
    align-items: center;
    gap: 6px;
    color: var(--brand-blue);
    font-family: var(--font-heading);
    font-size: 10px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.08em;
  }

  .streaming-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--brand-blue);
    animation: pulse 1.2s ease-in-out infinite;
  }

  .call-detail {
    margin-top: 12px;
    padding-top: 12px;
    border-top: 1px solid var(--border-subtle);
    cursor: default;
  }
</style>
