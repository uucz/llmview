<script lang="ts">
  import '../app.css';
  import { onMount, onDestroy } from 'svelte';
  import { connectWS, disconnectWS } from '$lib/stores/events';

  let { children } = $props();

  onMount(() => connectWS());
  onDestroy(() => disconnectWS());
</script>

{@render children()}
