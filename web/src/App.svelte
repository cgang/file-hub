<script>
  import { onMount } from 'svelte';
  import FileBrowser from './components/FileBrowser.svelte';
  import SetupPage from './components/SetupPage.svelte';

  let currentPage = 'main'; // 'main' or 'setup'

  onMount(async () => {
    // Determine which page to show based on the URL
    if (window.location.pathname === '/setup') {
      try {
        // Check if setup is needed by checking if database is empty
        const response = await fetch('/api/setup');
        const data = await response.json();

        if (data.setupNeeded) {
          currentPage = 'setup';
        } else {
          // If setup is not needed, redirect to main app
          window.location.href = '/ui/';
        }
      } catch (error) {
        // If there's a network error, we assume setup is not needed
        console.log('Error checking setup status:', error);
        window.location.href = '/ui/';
      }
    } else {
      try {
        // On main page, check if setup is needed
        const response = await fetch('/api/setup');
        const data = await response.json();

        if (data.setupNeeded) {
          // If setup is needed, redirect to the setup page
          window.location.href = '/setup';
        }
      } catch (error) {
        // If there's a network error, we assume setup is not needed
        console.log('Error checking setup status:', error);
      }
    }
  });
</script>

<main>
  {#if currentPage === 'setup'}
    <SetupPage />
  {:else}
    <h1>File Hub - Personal File Storage</h1>
    <FileBrowser />
  {/if}
</main>