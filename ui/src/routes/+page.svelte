<script lang="ts">
  import Header from '$lib/components/Header.svelte';
  import CallRow from '$lib/components/CallRow.svelte';
  import EmptyState from '$lib/components/EmptyState.svelte';
  import { sortedCalls } from '$lib/stores/events';

  let expandedId = $state<string | null>(null);

  function toggle(id: string) {
    expandedId = expandedId === id ? null : id;
  }
</script>

<div class="app">
  <Header />

  <main>
    {#if $sortedCalls.length === 0}
      <EmptyState />
    {:else}
      <div class="timeline">
        {#each $sortedCalls as call, i (call.id)}
          <CallRow
            {call}
            index={i}
            expanded={expandedId === call.id}
            ontoggle={() => toggle(call.id)}
          />
        {/each}
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
</style>
