<script lang="ts">
  import Header from '$lib/components/Header.svelte';
  import CallRow from '$lib/components/CallRow.svelte';
  import EmptyState from '$lib/components/EmptyState.svelte';
  import FilterBar from '$lib/components/FilterBar.svelte';
  import DiffView from '$lib/components/DiffView.svelte';
  import { sortedCalls, filteredCalls } from '$lib/stores/events';
  import { toggleTheme } from '$lib/stores/theme';

  let expandedId = $state<string | null>(null);
  let compareMode = $state(false);
  let compareSelected = $state<string[]>([]);
  let focusIndex = $state(-1);

  function toggle(id: string) {
    expandedId = expandedId === id ? null : id;
  }

  function toggleCompare() {
    compareMode = !compareMode;
    if (!compareMode) compareSelected = [];
    expandedId = null;
  }

  function selectForCompare(id: string) {
    if (compareSelected.includes(id)) {
      compareSelected = compareSelected.filter(x => x !== id);
    } else if (compareSelected.length < 2) {
      compareSelected = [...compareSelected, id];
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    // Skip if user is typing in an input
    if ((e.target as HTMLElement).tagName === 'INPUT') return;

    const calls = $filteredCalls;
    switch (e.key) {
      case 'j': // Next call
        e.preventDefault();
        focusIndex = Math.min(focusIndex + 1, calls.length - 1);
        scrollToFocused();
        break;
      case 'k': // Previous call
        e.preventDefault();
        focusIndex = Math.max(focusIndex - 1, 0);
        scrollToFocused();
        break;
      case 'Enter':
      case 'o': // Open/close detail
        if (focusIndex >= 0 && focusIndex < calls.length) {
          e.preventDefault();
          const id = calls[focusIndex].id;
          if (compareMode) selectForCompare(id);
          else toggle(id);
        }
        break;
      case '/': // Focus search
        e.preventDefault();
        (document.querySelector('.filter-bar input') as HTMLInputElement)?.focus();
        break;
      case 'Escape': // Close / deselect
        if (expandedId) { expandedId = null; }
        else if (compareSelected.length > 0) { compareSelected = []; }
        else if (compareMode) { compareMode = false; }
        else { focusIndex = -1; }
        break;
      case 'c': // Toggle compare mode
        if (!e.metaKey && !e.ctrlKey) {
          e.preventDefault();
          toggleCompare();
        }
        break;
      case 't': // Toggle theme
        if (!e.metaKey && !e.ctrlKey) {
          e.preventDefault();
          toggleTheme();
        }
        break;
    }
  }

  function scrollToFocused() {
    requestAnimationFrame(() => {
      const el = document.querySelector(`.call:nth-child(${focusIndex + 1})`);
      el?.scrollIntoView({ block: 'nearest', behavior: 'smooth' });
    });
  }
</script>

<svelte:window onkeydown={handleKeydown} />

<div class="app">
  <Header />

  <main>
    {#if $sortedCalls.length === 0}
      <EmptyState />
    {:else}
      <div class="timeline">
        <FilterBar {compareMode} ontogglecompare={toggleCompare} />

        {#if compareSelected.length === 2}
          <DiffView
            callIdA={compareSelected[0]}
            callIdB={compareSelected[1]}
            onclose={() => { compareSelected = []; }}
          />
        {/if}

        {#each $filteredCalls as call, i (call.id)}
          <CallRow
            {call}
            index={i}
            expanded={!compareMode && expandedId === call.id}
            ontoggle={() => compareMode ? selectForCompare(call.id) : toggle(call.id)}
            {compareMode}
            selected={compareSelected.includes(call.id)}
            focused={focusIndex === i}
          />
        {/each}
        {#if $filteredCalls.length === 0}
          <div class="no-results">No calls match your filters</div>
        {/if}
      </div>
    {/if}
  </main>
</div>

<style>
  .app {
    min-height: 100vh;
    display: flex;
    flex-direction: column;
    background: var(--surface-0);
    transition: background 0.25s;
  }

  main {
    flex: 1;
    padding: 20px 28px;
    animation: fadeUp 0.3s ease both;
  }

  .timeline {
    max-width: 960px;
    margin: 0 auto;
  }

  .no-results {
    text-align: center;
    padding: 48px 24px;
    color: var(--text-tertiary);
    font-family: var(--font-body);
    font-size: 13px;
  }
</style>
