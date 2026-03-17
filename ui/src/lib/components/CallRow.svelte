<script lang="ts">
  import type { APICall } from '$lib/stores/events';
  import { formatDuration, formatCost, formatTokens, formatTime, statusColor } from '$lib/utils/format';

  let { call, expanded = false, ontoggle, index = 0 }: {
    call: APICall;
    expanded: boolean;
    ontoggle: () => void;
    index?: number;
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
      default: return 'var(--text-2)';
    }
  }
</script>

<div
  class="call"
  class:streaming={isStreaming}
  class:has-error={call.status_code >= 400}
  class:expanded
  style="animation-delay: {Math.min(index * 0.04, 0.3)}s"
  onclick={ontoggle}
  role="button"
  tabindex="0"
  onkeydown={(e) => e.key === 'Enter' && ontoggle()}
>
  {#if isStreaming}
    <div class="streaming-bar"></div>
  {/if}

  <div class="call-main">
    <div class="call-left">
      <span
        class="provider-badge"
        style="--pcolor: {providerVar(call.provider)}"
      >
        {providerLabel(call.provider)}
      </span>
      <span class="model">{call.model || call.endpoint}</span>
      <span class="time">{formatTime(call.started_at)}</span>
    </div>
    <div class="call-right">
      {#if call.completed}
        <span class="status-code" style="color: {statusColor(call.status_code)}">{call.status_code}</span>
        <span class="tokens">{formatTokens(call.input_tokens)}<span class="arrow">&#8594;</span>{formatTokens(call.output_tokens)}</span>
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

  {#if expanded && (call.streamText || call.error)}
    <div class="call-detail">
      {#if call.error}
        <div class="error-box">{call.error}</div>
      {/if}
      {#if call.streamText}
        <div class="stream-text">{call.streamText}</div>
      {/if}
    </div>
  {/if}
</div>

<style>
  .call {
    position: relative;
    padding: 11px 16px;
    border-radius: var(--radius);
    background: var(--surface-1);
    border: 1px solid var(--border);
    cursor: pointer;
    margin-bottom: 6px;
    transition: border-color 0.2s;
    animation: fadeUp 0.35s ease both;
    overflow: hidden;
  }

  .call:hover {
    border-color: var(--border-hover);
  }

  .call.expanded {
    border-color: var(--border-hover);
  }

  /* Streaming: left accent bar */
  .streaming-bar {
    position: absolute;
    left: 0;
    top: 0;
    bottom: 0;
    width: 3px;
    background: var(--accent);
    animation: streamGlow 2s ease-in-out infinite;
  }

  .call.has-error {
    border-left: 3px solid var(--red);
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
    font-family: var(--font-sans);
    font-size: 9.5px;
    padding: 2px 8px;
    border-radius: 4px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    background: color-mix(in srgb, var(--pcolor) 10%, transparent);
    color: var(--pcolor);
    border: 1px solid color-mix(in srgb, var(--pcolor) 15%, transparent);
    flex-shrink: 0;
    line-height: 1.5;
  }

  .model {
    font-weight: 500;
    color: var(--text-0);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    font-size: 13px;
  }

  .time {
    font-size: 11px;
    color: var(--text-2);
    flex-shrink: 0;
  }

  .status-code {
    font-family: var(--font-sans);
    font-weight: 700;
    font-size: 11px;
    min-width: 28px;
    text-align: center;
  }

  .tokens {
    color: var(--text-1);
    font-size: 12px;
    min-width: 100px;
    text-align: right;
    font-variant-numeric: tabular-nums;
  }

  .arrow {
    color: var(--text-2);
    margin: 0 3px;
    font-size: 10px;
  }

  .duration {
    color: var(--text-2);
    font-size: 12px;
    min-width: 52px;
    text-align: right;
    font-variant-numeric: tabular-nums;
  }

  .cost {
    color: var(--green);
    font-size: 12px;
    font-weight: 600;
    min-width: 60px;
    text-align: right;
    font-variant-numeric: tabular-nums;
  }

  .cost.free {
    color: var(--text-2);
    font-weight: 400;
  }

  .streaming-indicator {
    display: flex;
    align-items: center;
    gap: 6px;
    color: var(--accent);
    font-family: var(--font-sans);
    font-size: 10px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.08em;
  }

  .streaming-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--accent);
    animation: blink 1.2s ease-in-out infinite;
  }

  .call-detail {
    margin-top: 12px;
    padding-top: 12px;
    border-top: 1px solid var(--border);
    animation: fadeIn 0.2s ease;
  }

  .stream-text {
    padding: 12px 14px;
    background: var(--surface-0);
    border-radius: var(--radius-sm);
    border: 1px solid var(--border);
    font-size: 12px;
    color: var(--text-1);
    max-height: 320px;
    overflow-y: auto;
    white-space: pre-wrap;
    word-break: break-word;
    line-height: 1.7;
  }

  .error-box {
    padding: 10px 14px;
    background: var(--red-dim);
    border: 1px solid color-mix(in srgb, var(--red) 20%, transparent);
    border-radius: var(--radius-sm);
    color: var(--red);
    font-size: 12px;
    margin-bottom: 8px;
  }
</style>
