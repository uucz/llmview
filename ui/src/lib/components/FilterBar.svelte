<script lang="ts">
  import { filters, filteredCalls, availableProviders, exportJSON, exportCSV } from '$lib/stores/events';

  let showExport = $state(false);

  function toggleProvider(p: string) {
    filters.update(f => {
      const next = new Set(f.providers);
      if (next.has(p)) next.delete(p);
      else next.add(p);
      return { ...f, providers: next };
    });
  }

  function setStatus(s: 'all' | 'success' | 'error') {
    filters.update(f => ({ ...f, status: s }));
  }

  function setQuery(e: Event) {
    filters.update(f => ({ ...f, query: (e.target as HTMLInputElement).value }));
  }

  function providerLabel(p: string): string {
    switch (p) {
      case 'openai': return 'OpenAI';
      case 'anthropic': return 'Anthropic';
      case 'ollama': return 'Ollama';
      default: return p;
    }
  }

  function providerColor(p: string): string {
    switch (p) {
      case 'openai': return 'var(--color-openai)';
      case 'anthropic': return 'var(--color-anthropic)';
      case 'ollama': return 'var(--color-ollama)';
      default: return 'var(--text-tertiary)';
    }
  }
</script>

<div class="filter-bar">
  <div class="search-group">
    <svg class="search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" width="14" height="14">
      <circle cx="11" cy="11" r="8" />
      <line x1="21" y1="21" x2="16.65" y2="16.65" />
    </svg>
    <input
      type="text"
      placeholder="Search model, endpoint..."
      value={$filters.query}
      oninput={setQuery}
    />
  </div>

  <div class="filter-chips">
    {#if $availableProviders.length > 1}
      <div class="chip-group">
        {#each $availableProviders as p}
          <button
            class="chip provider-chip"
            class:active={$filters.providers.has(p)}
            style="--chip-color: {providerColor(p)}"
            onclick={() => toggleProvider(p)}
          >{providerLabel(p)}</button>
        {/each}
      </div>
    {/if}

    <div class="chip-group">
      <button class="chip" class:active={$filters.status === 'all'} onclick={() => setStatus('all')}>All</button>
      <button class="chip" class:active={$filters.status === 'success'} onclick={() => setStatus('success')}>
        <span class="status-dot ok"></span>OK
      </button>
      <button class="chip" class:active={$filters.status === 'error'} onclick={() => setStatus('error')}>
        <span class="status-dot err"></span>Error
      </button>
    </div>
  </div>

  <div class="filter-right">
    <span class="count">{$filteredCalls.length}</span>

    <div class="export-wrapper">
      <button class="export-btn" onclick={() => showExport = !showExport} title="Export data">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" width="14" height="14">
          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
          <polyline points="7 10 12 15 17 10" />
          <line x1="12" y1="15" x2="12" y2="3" />
        </svg>
      </button>
      {#if showExport}
        <div class="export-dropdown">
          <button onclick={() => { exportJSON(); showExport = false; }}>Export JSON</button>
          <button onclick={() => { exportCSV(); showExport = false; }}>Export CSV</button>
        </div>
      {/if}
    </div>
  </div>
</div>

<style>
  .filter-bar {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 0 0 14px;
    animation: fadeIn 0.3s ease;
  }

  .search-group {
    position: relative;
    flex: 0 1 220px;
  }

  .search-icon {
    position: absolute;
    left: 10px;
    top: 50%;
    transform: translateY(-50%);
    color: var(--text-tertiary);
    pointer-events: none;
  }

  input {
    width: 100%;
    padding: 6px 12px 6px 30px;
    background: var(--surface-2);
    border: 1px solid var(--border-color);
    border-radius: var(--radius-sm);
    font-family: var(--font-body);
    font-size: 12px;
    color: var(--text-primary);
    outline: none;
    transition: border-color 0.2s;
  }

  input::placeholder { color: var(--text-tertiary); }
  input:focus { border-color: var(--border-focus); }

  .filter-chips {
    display: flex;
    gap: 6px;
    align-items: center;
  }

  .chip-group {
    display: flex;
    gap: 2px;
    background: var(--surface-2);
    border-radius: var(--radius-sm);
    padding: 2px;
  }

  .chip {
    font-family: var(--font-heading);
    font-size: 10px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    padding: 4px 10px;
    border: none;
    background: transparent;
    color: var(--text-tertiary);
    border-radius: 4px;
    cursor: pointer;
    transition: all 0.15s;
    display: flex;
    align-items: center;
    gap: 4px;
    white-space: nowrap;
  }

  .chip:hover { color: var(--text-secondary); }

  .chip.active {
    background: var(--surface-1);
    color: var(--text-primary);
    box-shadow: var(--shadow-sm);
  }

  .provider-chip.active { color: var(--chip-color); }

  .status-dot {
    width: 5px;
    height: 5px;
    border-radius: 50%;
  }

  .status-dot.ok { background: var(--risk-safe); }
  .status-dot.err { background: var(--risk-danger); }

  .filter-right {
    margin-left: auto;
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .count {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--text-tertiary);
    font-variant-numeric: tabular-nums;
  }

  .export-wrapper {
    position: relative;
  }

  .export-btn {
    width: 30px;
    height: 30px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: transparent;
    border: 1px solid var(--border-color);
    border-radius: var(--radius-sm);
    color: var(--text-tertiary);
    cursor: pointer;
    transition: all 0.15s;
  }

  .export-btn:hover {
    border-color: var(--border-focus);
    color: var(--brand-orange);
  }

  .export-dropdown {
    position: absolute;
    top: 100%;
    right: 0;
    margin-top: 4px;
    background: var(--surface-1);
    border: 1px solid var(--border-color);
    border-radius: var(--radius-sm);
    box-shadow: var(--shadow-md);
    overflow: hidden;
    z-index: 10;
    animation: fadeIn 0.15s ease;
  }

  .export-dropdown button {
    display: block;
    width: 100%;
    padding: 8px 20px;
    background: transparent;
    border: none;
    font-family: var(--font-heading);
    font-size: 11px;
    font-weight: 500;
    color: var(--text-secondary);
    cursor: pointer;
    text-align: left;
    white-space: nowrap;
    transition: background 0.1s;
  }

  .export-dropdown button:hover {
    background: var(--hover-overlay);
    color: var(--brand-orange);
  }

  .export-dropdown button + button {
    border-top: 1px solid var(--border-subtle);
  }
</style>
