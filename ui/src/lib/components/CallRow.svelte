<script lang="ts">
  import type { APICall } from '$lib/stores/events';
  import { formatDuration, formatCost, formatTokens, formatTime, providerColor, statusColor } from '$lib/utils/format';

  let { call, expanded = false, ontoggle }: {
    call: APICall;
    expanded: boolean;
    ontoggle: () => void;
  } = $props();

  let isStreaming = $derived(!call.completed && call.streaming);
</script>

<div
  class="call"
  class:streaming={isStreaming}
  class:error={call.status_code >= 400}
  class:expanded
  onclick={ontoggle}
  role="button"
  tabindex="0"
  onkeydown={(e) => e.key === 'Enter' && ontoggle()}
>
  <div class="call-main">
    <div class="call-left">
      <span class="provider" style="--provider-color: {providerColor(call.provider)}">
        {call.provider}
      </span>
      <span class="model">{call.model || call.endpoint}</span>
      <span class="time">{formatTime(call.started_at)}</span>
    </div>
    <div class="call-right">
      {#if call.completed}
        <span class="status" style="color: {statusColor(call.status_code)}">{call.status_code}</span>
        <span class="tokens">{formatTokens(call.input_tokens)} → {formatTokens(call.output_tokens)}</span>
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
        <div class="error-text">{call.error}</div>
      {/if}
      {#if call.streamText}
        <div class="stream-text">{call.streamText}</div>
      {/if}
    </div>
  {/if}
</div>

<style>
  .call {
    padding: 10px 16px;
    border-radius: var(--radius);
    background: var(--bg-1);
    border: 1px solid var(--border);
    cursor: pointer;
    transition: all 0.15s ease;
    margin-bottom: 6px;
  }

  .call:hover {
    border-color: var(--border-hover);
    background: var(--bg-2);
  }

  .call.streaming {
    border-left: 3px solid var(--accent);
    animation: stream-glow 1.5s ease-in-out infinite;
  }

  @keyframes stream-glow {
    0%, 100% { box-shadow: inset 3px 0 8px -4px var(--accent-dim); }
    50% { box-shadow: inset 3px 0 16px -4px var(--accent); }
  }

  .call.error {
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

  .provider {
    font-size: 10px;
    padding: 2px 8px;
    border-radius: var(--radius-sm);
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    background: color-mix(in srgb, var(--provider-color) 15%, transparent);
    color: var(--provider-color);
    flex-shrink: 0;
  }

  .model {
    font-weight: 500;
    color: var(--text-0);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .time {
    font-size: 11px;
    color: var(--text-2);
    flex-shrink: 0;
  }

  .status {
    font-weight: 600;
    font-size: 12px;
    min-width: 28px;
    text-align: center;
  }

  .tokens {
    color: var(--text-1);
    font-size: 12px;
    min-width: 90px;
    text-align: right;
    font-variant-numeric: tabular-nums;
  }

  .duration {
    color: var(--text-2);
    font-size: 12px;
    min-width: 50px;
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
  }

  .streaming-indicator {
    display: flex;
    align-items: center;
    gap: 6px;
    color: var(--accent);
    font-size: 11px;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .streaming-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--accent);
    animation: blink 1s infinite;
  }

  @keyframes blink {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.3; }
  }

  .call-detail {
    margin-top: 10px;
    padding-top: 10px;
    border-top: 1px solid var(--border);
  }

  .stream-text {
    padding: 10px 12px;
    background: var(--bg-0);
    border-radius: var(--radius-sm);
    font-size: 12px;
    color: var(--text-1);
    max-height: 300px;
    overflow-y: auto;
    white-space: pre-wrap;
    word-break: break-word;
    line-height: 1.6;
  }

  .error-text {
    padding: 8px 12px;
    background: var(--red-dim);
    border-radius: var(--radius-sm);
    color: var(--red);
    font-size: 12px;
    margin-bottom: 8px;
  }
</style>
