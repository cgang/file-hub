<script>
  import { onMount } from 'svelte';
  import FileBrowser from './components/FileBrowser.svelte';
  import SetupPage from './components/SetupPage.svelte';
  import Login from './components/Login.svelte';

  let status = null;
  let error = null;

  onMount(async () => {
    try {
      const res = await fetch('/api/status');
      status = await res.json();
    } catch (err) {
      error = 'Failed to fetch status';
      console.error(err);
    }
  });
</script>

<main>
  {#if status === null}
    <div>Loading...</div>
  {:else if status.setup}
    <SetupPage />
  {:else if status.username}
    <h1>File Hub - Personal File Storage</h1>
    <FileBrowser />
  {:else}
    <Login />
  {/if}
</main>
