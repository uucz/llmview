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
        {#each $sortedCalls as call (call.id)}
          <CallRow
            {call}
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
  }

  main {
    flex: 1;
    padding: 16px 20px;
  }

  .timeline {
    max-width: 960px;
    margin: 0 auto;
  }
</style>
