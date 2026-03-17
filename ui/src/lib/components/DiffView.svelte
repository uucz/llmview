<script lang="ts">
  import type { APICallDetail } from '$lib/stores/events';
  import { fetchDetail } from '$lib/stores/events';
  import { formatDuration, formatCost, formatTokens, formatTime } from '$lib/utils/format';

  let { callIdA, callIdB, onclose }: {
    callIdA: string;
    callIdB: string;
    onclose: () => void;
  } = $props();

  let detailA = $state<APICallDetail | null>(null);
  let detailB = $state<APICallDetail | null>(null);
  let loading = $state(true);

  interface ChatMessage { role: string; content: string; }
  interface DiffMessage { status: 'shared' | 'added' | 'removed'; role: string; content: string; }

  $effect(() => {
    let cancelled = false;
    loading = true;
    Promise.all([fetchDetail(callIdA), fetchDetail(callIdB)]).then(([a, b]) => {
      if (!cancelled) { detailA = a; detailB = b; loading = false; }
    });
    return () => { cancelled = true; };
  });

  let diffMessages = $derived.by(() => {
    if (!detailA || !detailB) return [];
    return computeDiff(parseMessages(detailA), parseMessages(detailB));
  });

  let stats = $derived.by(() => {
    if (!detailA || !detailB) return null;
    const inDiff = detailB.input_tokens - detailA.input_tokens;
    const outDiff = detailB.output_tokens - detailA.output_tokens;
    return {
      tokenDiff: inDiff + outDiff,
      costDiff: detailB.cost - detailA.cost,
      durationDiff: detailB.duration_ms - detailA.duration_ms,
      added: diffMessages.filter(m => m.status === 'added').length,
      removed: diffMessages.filter(m => m.status === 'removed').length,
    };
  });

  function parseMessages(d: APICallDetail): ChatMessage[] {
    const msgs: ChatMessage[] = [];
    try {
      const req = JSON.parse(d.request_body);
      if (typeof req.system === 'string' && req.system) {
        msgs.push({ role: 'system', content: req.system });
      } else if (Array.isArray(req.system)) {
        const text = req.system.map((s: any) => typeof s === 'string' ? s : s.text || '').filter(Boolean).join('\n');
        if (text) msgs.push({ role: 'system', content: text });
      }
      if (Array.isArray(req.messages)) {
        for (const m of req.messages) {
          let content: string;
          if (typeof m.content === 'string') content = m.content;
          else if (Array.isArray(m.content)) {
            content = m.content.map((c: any) => {
              if (c.type === 'text') return c.text;
              if (c.type === 'image_url' || c.type === 'image') return '[Image]';
              if (c.type === 'tool_use') return `[Tool: ${c.name}]`;
              if (c.type === 'tool_result') return `[Tool Result]`;
              return JSON.stringify(c);
            }).join('\n');
          } else content = m.content ? JSON.stringify(m.content) : '';
          if (content) msgs.push({ role: m.role || 'unknown', content });
        }
      }
    } catch {}
    // Response
    if (d.response_body) {
      try {
        const resp = JSON.parse(d.response_body);
        if (resp.choices?.[0]?.message?.content) {
          msgs.push({ role: 'assistant', content: resp.choices[0].message.content });
        } else if (Array.isArray(resp.content)) {
          const text = resp.content.filter((c: any) => c.type === 'text').map((c: any) => c.text).join('\n');
          if (text) msgs.push({ role: 'assistant', content: text });
        }
      } catch {}
    }
    return msgs;
  }

  function fingerprint(m: ChatMessage): string {
    return `${m.role}:${m.content.substring(0, 300)}`;
  }

  function computeDiff(a: ChatMessage[], b: ChatMessage[]): DiffMessage[] {
    const fpA = a.map(fingerprint);
    const fpB = b.map(fingerprint);
    const n = fpA.length, m = fpB.length;

    // LCS dynamic programming
    const dp: number[][] = Array.from({ length: n + 1 }, () => Array(m + 1).fill(0));
    for (let i = 1; i <= n; i++) {
      for (let j = 1; j <= m; j++) {
        dp[i][j] = fpA[i - 1] === fpB[j - 1] ? dp[i - 1][j - 1] + 1 : Math.max(dp[i - 1][j], dp[i][j - 1]);
      }
    }

    // Backtrack to find shared pairs
    const shared: [number, number][] = [];
    let i = n, j = m;
    while (i > 0 && j > 0) {
      if (fpA[i - 1] === fpB[j - 1]) { shared.unshift([i - 1, j - 1]); i--; j--; }
      else if (dp[i - 1][j] > dp[i][j - 1]) i--;
      else j--;
    }

    // Build unified diff
    const result: DiffMessage[] = [];
    let ai = 0, bi = 0, si = 0;
    while (ai < n || bi < m) {
      if (si < shared.length) {
        const [sa, sb] = shared[si];
        while (ai < sa) { result.push({ status: 'removed', ...a[ai] }); ai++; }
        while (bi < sb) { result.push({ status: 'added', ...b[bi] }); bi++; }
        result.push({ status: 'shared', ...a[ai] });
        ai++; bi++; si++;
      } else {
        while (ai < n) { result.push({ status: 'removed', ...a[ai] }); ai++; }
        while (bi < m) { result.push({ status: 'added', ...b[bi] }); bi++; }
      }
    }
    return result;
  }

  function roleColor(role: string): string {
    switch (role) {
      case 'system': return 'var(--text-tertiary)';
      case 'user': return 'var(--brand-blue)';
      case 'assistant': return 'var(--brand-orange)';
      case 'tool': return 'var(--brand-green)';
      default: return 'var(--text-tertiary)';
    }
  }

  function signedNum(n: number): string {
    return n > 0 ? `+${n}` : `${n}`;
  }
</script>

<div class="diff-view">
  <div class="diff-header">
    <div class="diff-title">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" width="16" height="16">
        <path d="M16 3h5v5" />
        <line x1="21" y1="3" x2="14" y2="10" />
        <path d="M8 21H3v-5" />
        <line x1="3" y1="21" x2="10" y2="14" />
      </svg>
      Prompt Diff
    </div>
    <button class="close-btn" onclick={onclose} title="Close diff">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" width="16" height="16">
        <line x1="18" y1="6" x2="6" y2="18" />
        <line x1="6" y1="6" x2="18" y2="18" />
      </svg>
    </button>
  </div>

  {#if loading}
    <div class="diff-body">
      <div class="diff-loading">
        <div class="skeleton"></div>
        <div class="skeleton short"></div>
        <div class="skeleton"></div>
      </div>
    </div>
  {:else if detailA && detailB && stats}
    <div class="diff-stats">
      <div class="diff-call">
        <span class="call-badge a">A</span>
        <span class="call-model">{detailA.model || detailA.endpoint}</span>
        <span class="call-meta">{formatTime(new Date(detailA.started_at).getTime())}</span>
        <span class="call-meta">{formatTokens(detailA.input_tokens + detailA.output_tokens)} tok</span>
        <span class="call-meta">{formatCost(detailA.cost)}</span>
      </div>
      <div class="diff-deltas">
        {#if stats.added > 0}
          <span class="delta added">+{stats.added} msg</span>
        {/if}
        {#if stats.removed > 0}
          <span class="delta removed">-{stats.removed} msg</span>
        {/if}
        <span class="delta" class:cost-up={stats.costDiff > 0} class:cost-down={stats.costDiff < 0}>
          {stats.costDiff > 0 ? '+' : ''}{formatCost(Math.abs(stats.costDiff))}
        </span>
        <span class="delta" class:cost-up={stats.tokenDiff > 0} class:cost-down={stats.tokenDiff < 0}>
          {signedNum(stats.tokenDiff)} tok
        </span>
      </div>
      <div class="diff-call">
        <span class="call-badge b">B</span>
        <span class="call-model">{detailB.model || detailB.endpoint}</span>
        <span class="call-meta">{formatTime(new Date(detailB.started_at).getTime())}</span>
        <span class="call-meta">{formatTokens(detailB.input_tokens + detailB.output_tokens)} tok</span>
        <span class="call-meta">{formatCost(detailB.cost)}</span>
      </div>
    </div>

    <div class="diff-body">
      {#if diffMessages.length > 0}
        {#each diffMessages as msg}
          <div class="diff-msg {msg.status}" style="--role-color: {roleColor(msg.role)}">
            <div class="msg-indicator">
              {#if msg.status === 'added'}
                <span class="ind-icon add">+</span>
              {:else if msg.status === 'removed'}
                <span class="ind-icon rem">-</span>
              {:else}
                <span class="ind-icon eq">=</span>
              {/if}
            </div>
            <div class="msg-body">
              <span class="msg-role">{msg.role}</span>
              <div class="msg-content" class:truncated={msg.status === 'shared'}>{msg.content}</div>
            </div>
          </div>
        {/each}
      {:else}
        <div class="no-data">No messages to compare</div>
      {/if}
    </div>
  {:else}
    <div class="diff-body">
      <div class="no-data">Failed to load call details</div>
    </div>
  {/if}
</div>

<style>
  .diff-view {
    background: var(--surface-1);
    border: 1px solid var(--border-focus);
    border-radius: var(--radius);
    overflow: hidden;
    animation: fadeUp 0.3s ease both;
    margin-bottom: 14px;
  }

  .diff-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 10px 16px;
    border-bottom: 1px solid var(--border-color);
  }

  .diff-title {
    display: flex;
    align-items: center;
    gap: 8px;
    font-family: var(--font-heading);
    font-size: 13px;
    font-weight: 600;
    color: var(--text-primary);
  }

  .diff-title svg { color: var(--brand-orange); }

  .close-btn {
    width: 28px;
    height: 28px;
    border: 1px solid var(--border-color);
    border-radius: var(--radius-sm);
    background: transparent;
    color: var(--text-tertiary);
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.15s;
  }

  .close-btn:hover {
    border-color: rgba(192, 90, 60, 0.25);
    color: var(--risk-danger);
  }

  .diff-stats {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 8px 16px;
    background: var(--surface-2);
    border-bottom: 1px solid var(--border-subtle);
    gap: 12px;
  }

  .diff-call {
    display: flex;
    align-items: center;
    gap: 8px;
    flex: 1;
    min-width: 0;
  }

  .diff-call:last-child { justify-content: flex-end; }

  .call-badge {
    font-family: var(--font-heading);
    font-size: 9px;
    font-weight: 700;
    padding: 1px 6px;
    border-radius: 4px;
    line-height: 16px;
    flex-shrink: 0;
  }

  .call-badge.a {
    background: rgba(192, 90, 60, 0.10);
    color: var(--risk-danger);
  }

  .call-badge.b {
    background: rgba(120, 140, 93, 0.10);
    color: var(--brand-green);
  }

  .call-model {
    font-family: var(--font-body);
    font-size: 12px;
    font-weight: 500;
    color: var(--text-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .call-meta {
    font-family: var(--font-mono);
    font-size: 10px;
    color: var(--text-tertiary);
    white-space: nowrap;
  }

  .diff-deltas {
    display: flex;
    gap: 10px;
    flex-shrink: 0;
  }

  .delta {
    font-family: var(--font-mono);
    font-size: 10px;
    font-weight: 600;
    color: var(--text-tertiary);
    white-space: nowrap;
  }

  .delta.added { color: var(--brand-green); }
  .delta.removed { color: var(--risk-danger); }
  .delta.cost-up { color: var(--risk-danger); }
  .delta.cost-down { color: var(--brand-green); }

  .diff-body {
    max-height: 480px;
    overflow-y: auto;
    padding: 10px 16px;
  }

  .diff-loading { padding: 16px 0; }

  .skeleton {
    height: 12px;
    background: var(--surface-3);
    border-radius: 4px;
    margin-bottom: 8px;
    animation: skeletonPulse 1.5s ease-in-out infinite;
    width: 80%;
  }

  .skeleton.short { width: 40%; }

  .diff-msg {
    display: flex;
    gap: 0;
    margin-bottom: 4px;
    border-radius: var(--radius-sm);
    overflow: hidden;
    border: 1px solid transparent;
  }

  .diff-msg.shared { opacity: 0.5; }

  .diff-msg.added {
    background: rgba(120, 140, 93, 0.05);
    border-color: rgba(120, 140, 93, 0.12);
  }

  .diff-msg.removed {
    background: rgba(192, 90, 60, 0.05);
    border-color: rgba(192, 90, 60, 0.12);
  }

  .msg-indicator {
    display: flex;
    align-items: flex-start;
    justify-content: center;
    width: 28px;
    padding-top: 10px;
    flex-shrink: 0;
  }

  .ind-icon {
    font-family: var(--font-mono);
    font-size: 13px;
    font-weight: 700;
    line-height: 1;
  }

  .ind-icon.add { color: var(--brand-green); }
  .ind-icon.rem { color: var(--risk-danger); }
  .ind-icon.eq { color: var(--text-tertiary); font-size: 10px; }

  .msg-body {
    flex: 1;
    min-width: 0;
    padding: 8px 12px 8px 0;
  }

  .msg-role {
    font-family: var(--font-heading);
    font-size: 9px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    color: var(--role-color);
    display: block;
    margin-bottom: 3px;
  }

  .msg-content {
    font-family: var(--font-body);
    font-size: 12px;
    color: var(--text-secondary);
    line-height: 1.6;
    white-space: pre-wrap;
    word-break: break-word;
  }

  .msg-content.truncated {
    max-height: 48px;
    overflow: hidden;
    -webkit-mask-image: linear-gradient(to bottom, black 60%, transparent 100%);
    mask-image: linear-gradient(to bottom, black 60%, transparent 100%);
  }

  .no-data {
    padding: 32px;
    text-align: center;
    color: var(--text-tertiary);
    font-family: var(--font-body);
    font-size: 13px;
  }
</style>
