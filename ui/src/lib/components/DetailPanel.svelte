<script lang="ts">
  import type { APICallDetail } from '$lib/stores/events';
  import { fetchDetail } from '$lib/stores/events';

  let { callId, streamText = '', error = '', completed = true }: {
    callId: string;
    streamText: string;
    error: string;
    completed: boolean;
  } = $props();

  let detail = $state<APICallDetail | null>(null);
  let loading = $state(true);
  let activeTab = $state<'messages' | 'request' | 'response'>('messages');
  let copied = $state(false);

  interface ChatMessage {
    role: string;
    content: string;
  }

  let messages = $derived.by(() => {
    if (!detail) return [];
    return parseMessages(detail, streamText);
  });

  let hasMessages = $derived(messages.length > 0);

  $effect(() => {
    const id = callId;
    const done = completed;

    if (!done) {
      loading = false;
      detail = null;
      return;
    }

    let cancelled = false;
    loading = true;
    fetchDetail(id).then(d => {
      if (!cancelled) {
        detail = d;
        loading = false;
      }
    });
    return () => { cancelled = true; };
  });

  function parseMessages(d: APICallDetail, stream: string): ChatMessage[] {
    const msgs: ChatMessage[] = [];
    try {
      const req = JSON.parse(d.request_body);

      // Anthropic system prompt
      if (typeof req.system === 'string' && req.system) {
        msgs.push({ role: 'system', content: req.system });
      } else if (Array.isArray(req.system)) {
        const text = req.system
          .map((s: any) => typeof s === 'string' ? s : s.text || '')
          .filter(Boolean).join('\n');
        if (text) msgs.push({ role: 'system', content: text });
      }

      // Messages array (OpenAI + Anthropic)
      if (Array.isArray(req.messages)) {
        for (const m of req.messages) {
          let content: string;
          if (typeof m.content === 'string') {
            content = m.content;
          } else if (Array.isArray(m.content)) {
            content = m.content.map((c: any) => {
              if (c.type === 'text') return c.text;
              if (c.type === 'image_url' || c.type === 'image') return '[Image]';
              if (c.type === 'tool_use') return `[Tool: ${c.name}]\n${JSON.stringify(c.input, null, 2)}`;
              if (c.type === 'tool_result') return `[Tool Result]\n${typeof c.content === 'string' ? c.content : JSON.stringify(c.content)}`;
              return JSON.stringify(c);
            }).join('\n');
          } else {
            content = m.content ? JSON.stringify(m.content) : '';
          }
          if (content) msgs.push({ role: m.role || 'unknown', content });
        }
      }
    } catch { /* request body not JSON */ }

    // Extract assistant response
    let assistantContent = '';
    if (d.response_body) {
      try {
        const resp = JSON.parse(d.response_body);
        if (resp.choices?.[0]?.message?.content) {
          assistantContent = resp.choices[0].message.content;
        } else if (Array.isArray(resp.content)) {
          const parts: string[] = [];
          for (const c of resp.content) {
            if (c.type === 'text' && c.text) parts.push(c.text);
            else if (c.type === 'tool_use') parts.push(`[Tool: ${c.name}]\n${JSON.stringify(c.input, null, 2)}`);
          }
          assistantContent = parts.join('\n');
        }
      } catch { /* SSE data, not JSON */ }
    }
    if (!assistantContent && stream) assistantContent = stream;
    if (assistantContent) msgs.push({ role: 'assistant', content: assistantContent });

    return msgs;
  }

  function highlightJSON(raw: string): string {
    let json: string;
    try {
      json = JSON.stringify(JSON.parse(raw), null, 2);
    } catch {
      return raw.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
    }
    json = json.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
    return json.replace(
      /("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g,
      (match) => {
        let cls = 'jv-num';
        if (/^"/.test(match)) cls = /:$/.test(match) ? 'jv-key' : 'jv-str';
        else if (/true|false/.test(match)) cls = 'jv-bool';
        else if (/null/.test(match)) cls = 'jv-null';
        return `<span class="${cls}">${match}</span>`;
      }
    );
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

  async function copyToClipboard() {
    let text = '';
    if (activeTab === 'messages') {
      text = messages.map(m => `[${m.role}]\n${m.content}`).join('\n\n');
    } else if (activeTab === 'request' && detail?.request_body) {
      try { text = JSON.stringify(JSON.parse(detail.request_body), null, 2); } catch { text = detail.request_body; }
    } else if (activeTab === 'response' && detail?.response_body) {
      try { text = JSON.stringify(JSON.parse(detail.response_body), null, 2); } catch { text = detail.response_body; }
    }
    try {
      await navigator.clipboard.writeText(text);
      copied = true;
      setTimeout(() => { copied = false; }, 1500);
    } catch {}
  }
</script>

<div class="detail-panel">
  {#if error}
    <div class="error-box">{error}</div>
  {/if}

  {#if !completed}
    {#if streamText}
      <div class="stream-text">{streamText}</div>
    {:else}
      <div class="waiting">
        <span class="wait-dot"></span>
        Waiting for response...
      </div>
    {/if}
  {:else if loading}
    <div class="detail-loading">
      <div class="skeleton"></div>
      <div class="skeleton short"></div>
      <div class="skeleton"></div>
    </div>
  {:else if detail}
    <div class="detail-tabs">
      <div class="tabs">
        <button class="tab" class:active={activeTab === 'messages'} onclick={() => activeTab = 'messages'}>
          Messages
          {#if hasMessages}
            <span class="tab-count">{messages.length}</span>
          {/if}
        </button>
        <button class="tab" class:active={activeTab === 'request'} onclick={() => activeTab = 'request'}>
          Request
        </button>
        <button class="tab" class:active={activeTab === 'response'} onclick={() => activeTab = 'response'}>
          Response
        </button>
      </div>
      <button class="copy-btn" onclick={copyToClipboard} title="Copy to clipboard">
        {#if copied}
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" width="14" height="14">
            <polyline points="20 6 9 17 4 12" />
          </svg>
        {:else}
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" width="14" height="14">
            <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
            <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
          </svg>
        {/if}
      </button>
    </div>

    <div class="detail-body">
      {#if activeTab === 'messages'}
        {#if hasMessages}
          <div class="messages">
            {#each messages as msg}
              <div class="message" style="--role-color: {roleColor(msg.role)}">
                <span class="msg-role">{msg.role}</span>
                <div class="msg-content">{msg.content}</div>
              </div>
            {/each}
          </div>
        {:else}
          <div class="no-data">No chat messages found</div>
        {/if}
      {:else if activeTab === 'request'}
        <div class="json-view">
          {#if detail.request_body}
            <pre>{@html highlightJSON(detail.request_body)}</pre>
          {:else}
            <div class="no-data">No request body</div>
          {/if}
        </div>
      {:else}
        <div class="json-view">
          {#if detail.response_body}
            <pre>{@html highlightJSON(detail.response_body)}</pre>
          {:else if streamText}
            <pre>{streamText}</pre>
          {:else}
            <div class="no-data">No response body</div>
          {/if}
        </div>
      {/if}
    </div>
  {:else}
    {#if streamText}
      <div class="stream-text">{streamText}</div>
    {:else}
      <div class="no-data">Failed to load call details</div>
    {/if}
  {/if}
</div>

<style>
  .detail-panel {
    animation: fadeIn 0.2s ease;
  }

  .error-box {
    padding: 10px 14px;
    background: rgba(192, 90, 60, 0.08);
    border: 1px solid rgba(192, 90, 60, 0.18);
    border-radius: var(--radius-sm);
    color: var(--risk-danger);
    font-size: 12px;
    margin-bottom: 10px;
  }

  .detail-loading {
    padding: 16px 0;
  }

  .skeleton {
    height: 12px;
    background: var(--surface-3);
    border-radius: 4px;
    margin-bottom: 8px;
    animation: skeletonPulse 1.5s ease-in-out infinite;
    width: 80%;
  }

  .skeleton.short { width: 40%; }

  .detail-tabs {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 10px;
  }

  .tabs {
    display: flex;
    gap: 2px;
    background: var(--surface-2);
    border-radius: var(--radius-sm);
    padding: 2px;
  }

  .tab {
    font-family: var(--font-heading);
    font-size: 10px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    padding: 5px 12px;
    border: none;
    background: transparent;
    color: var(--text-tertiary);
    border-radius: 4px;
    cursor: pointer;
    transition: all 0.15s;
    display: flex;
    align-items: center;
    gap: 5px;
  }

  .tab:hover { color: var(--text-secondary); }

  .tab.active {
    background: var(--surface-1);
    color: var(--text-primary);
    box-shadow: var(--shadow-sm);
  }

  .tab-count {
    font-size: 9px;
    padding: 0 5px;
    border-radius: 8px;
    background: var(--active-orange-bg);
    color: var(--brand-orange);
    font-weight: 700;
    line-height: 16px;
  }

  .copy-btn {
    width: 30px;
    height: 30px;
    border: 1px solid var(--border-color);
    border-radius: var(--radius-sm);
    background: transparent;
    color: var(--text-tertiary);
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.15s;
    flex-shrink: 0;
  }

  .copy-btn:hover {
    border-color: var(--border-focus);
    color: var(--brand-orange);
    background: var(--active-orange-bg);
  }

  .detail-body {
    max-height: 480px;
    overflow-y: auto;
    border-radius: var(--radius-sm);
  }

  /* Messages */
  .messages {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .message {
    position: relative;
    padding: 10px 14px 10px 17px;
    background: var(--surface-2);
    border-radius: var(--radius-sm);
    border: 1px solid var(--border-subtle);
    overflow: hidden;
  }

  .message::before {
    content: '';
    position: absolute;
    left: 0;
    top: 0;
    bottom: 0;
    width: 3px;
    background: var(--role-color);
  }

  .msg-role {
    font-family: var(--font-heading);
    font-size: 9px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    color: var(--role-color);
    display: block;
    margin-bottom: 5px;
  }

  .msg-content {
    font-family: var(--font-body);
    font-size: 12.5px;
    color: var(--text-secondary);
    line-height: 1.7;
    white-space: pre-wrap;
    word-break: break-word;
  }

  /* JSON view */
  .json-view {
    background: var(--surface-2);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
    overflow: auto;
  }

  .json-view pre {
    font-family: var(--font-mono);
    font-size: 11.5px;
    line-height: 1.6;
    color: var(--text-secondary);
    padding: 14px;
    margin: 0;
    white-space: pre-wrap;
    word-break: break-all;
  }

  /* JSON syntax colors */
  :global(.jv-key) { color: var(--text-primary); font-weight: 500; }
  :global(.jv-str) { color: var(--brand-green); }
  :global(.jv-num) { color: var(--brand-blue); }
  :global(.jv-bool) { color: var(--brand-orange); }
  :global(.jv-null) { color: var(--text-tertiary); }

  /* Stream text */
  .stream-text {
    padding: 12px 14px;
    background: var(--surface-2);
    border-radius: var(--radius-sm);
    border: 1px solid var(--border-subtle);
    font-family: var(--font-mono);
    font-size: 12px;
    color: var(--text-secondary);
    max-height: 320px;
    overflow-y: auto;
    white-space: pre-wrap;
    word-break: break-word;
    line-height: 1.7;
  }

  .waiting {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 16px 0;
    color: var(--brand-blue);
    font-family: var(--font-heading);
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.06em;
  }

  .wait-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--brand-blue);
    animation: pulse 1.2s ease-in-out infinite;
  }

  .no-data {
    padding: 32px;
    text-align: center;
    color: var(--text-tertiary);
    font-family: var(--font-body);
    font-size: 13px;
  }
</style>
