<script>
  import { onMount } from 'svelte';
  import { listDirectory, uploadFile } from '../utils/webdav.js';
  import FileCard from './FileCard.svelte';
  import NavigationBar from './NavigationBar.svelte';
  import UploadComponent from './UploadComponent.svelte';

  let currentPath = '/';
  let files = [];
  let loading = false;
  let error = null;
  let breadcrumbs = [{ name: 'Home', path: '/' }];

  // Load the initial directory
  onMount(async () => {
    await loadDirectory(currentPath);
  });

  // Function to load directory contents
  async function loadDirectory(path) {
    loading = true;
    error = null;

    try {
      files = await listDirectory(path);
      currentPath = path;
      
      // Update breadcrumbs
      const pathParts = path.split('/').filter(part => part !== '');
      breadcrumbs = [{ name: 'Home', path: '/' }];
      
      let currentPathSoFar = '';
      for (let i = 0; i < pathParts.length; i++) {
        currentPathSoFar += '/' + pathParts[i];
        breadcrumbs.push({
          name: pathParts[i],
          path: currentPathSoFar
        });
      }
    } catch (err) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  // Function to handle navigation to a directory
  async function navigateToDirectory(path) {
    await loadDirectory(path);
  }

  // Function to handle file upload
  async function handleFileUpload(file) {
    try {
      await uploadFile(currentPath, file);
      // Refresh the directory after upload
      await loadDirectory(currentPath);
    } catch (err) {
      error = `Upload failed: ${err.message}`;
    }
  }

  // Function to navigate to parent directory
  async function goToParent() {
    if (currentPath === '/') return;
    
    const pathParts = currentPath.split('/').filter(part => part !== '');
    pathParts.pop(); // Remove current directory
    
    let newPath = '/' + pathParts.join('/');
    if (newPath === '') newPath = '/';
    
    await loadDirectory(newPath);
  }
</script>

<div class="file-browser">
  <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px;">
    <NavigationBar {breadcrumbs} on:navigate={loadDirectory} />
    <UploadComponent on:fileUpload={handleFileUpload} {currentPath} />
  </div>

  {#if error}
    <div class="error">{error}</div>
  {/if}

  {#if loading}
    <div class="loading">Loading files...</div>
  {:else}
    <div class="file-grid">
      {#if currentPath !== '/'}
        <div class="file-card" on:click={goToParent}>
          <div class="file-icon">📁</div>
          <div class="file-name">.. (Parent)</div>
        </div>
      {/if}

      {#each files as file (file.path)}
        <FileCard {file} on:select={navigateToDirectory} />
      {/each}
    </div>
  {/if}
</div>