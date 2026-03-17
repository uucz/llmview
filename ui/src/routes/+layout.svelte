<script lang="ts">
  import '../app.css';
  import { onMount, onDestroy } from 'svelte';
  import { connectWS, disconnectWS } from '$lib/stores/events';
  import { theme } from '$lib/stores/theme';

  let { children } = $props();

  onMount(() => {
    // Apply saved theme on mount
    document.documentElement.setAttribute('data-theme', $theme);
    connectWS();
  });
  onDestroy(() => disconnectWS());
</script>

{@render children()}
