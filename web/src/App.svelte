<script>
  import FileBrowser from './components/FileBrowser.svelte';
  import SetupPage from './components/SetupPage.svelte';
</script>

<main>
  {#if new URLSearchParams(window.location.search).get('mode') === 'setup'}
    <SetupPage />
  {:else}
    <h1>File Hub - Personal File Storage</h1>
    <FileBrowser />
  {/if}
</main>
