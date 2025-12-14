<script>
  import { onMount } from 'svelte';
  import { listDirectory, uploadFile } from '../utils/webdav.js';
  import FileCard from './FileCard.svelte';
  import NavigationBar from './NavigationBar.svelte';
  import UploadComponent from './UploadComponent.svelte';
  import { formatSize, getIcon } from '../utils/fileUtils.js';

  let currentPath = '/';
  let files = [];
  let loading = false;
  let error = null;
  let breadcrumbs = [{ name: 'Home', path: '/' }];
  let viewMode = 'table'; // 'table' or 'grid'
  let scanning = false;

  // Load the initial directory
  onMount(async () => {
    await loadDirectory(currentPath);
  });

  // Function to load directory contents
  async function loadDirectory(path) {
    loading = true;
    error = null;

    // Ensure path is a string, default to root if not provided
    if (typeof path !== 'string' || path === null || path === undefined) {
      path = '/';
    }

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
  async function navigateToDirectory(event) {
    const path = event.detail;
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

  // Function to toggle between table and grid views
  function toggleView() {
    viewMode = viewMode === 'table' ? 'grid' : 'table';
  }

  // Function to navigate to parent directory
  async function goToParent() {
    // Ensure currentPath is a string, default to root if not provided
    if (typeof currentPath !== 'string' || currentPath === null || currentPath === undefined) {
      await loadDirectory('/');
      return;
    }

    if (currentPath === '/') return;

    const pathParts = currentPath.split('/').filter(part => part !== '');
    pathParts.pop(); // Remove current directory

    let newPath = '/' + pathParts.join('/');
    if (newPath === '') newPath = '/';

    await loadDirectory(newPath);
  }

  // Function to trigger scan
  async function triggerScan() {
    scanning = true;
    error = null;
    try {
      const response = await fetch('/api/scan_files', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        }
      });

      if (!response.ok) {
        throw new Error(`Scan failed with status ${response.status}`);
      }

      // Refresh the directory after scan
      await loadDirectory(currentPath);
    } catch (err) {
      error = `Scan failed: ${err.message}`;
    } finally {
      scanning = false;
    }
  }
</script>

<div class="file-browser">
  <div class="header-section">
    <div class="navigation-controls">
      <NavigationBar {breadcrumbs} on:navigate={loadDirectory} />
      {#if currentPath !== '/'}
        <button class="move-up-button" on:click={goToParent} aria-label="Move up one level">
          Move Up
        </button>
      {/if}
    </div>
    <div class="header-actions">
      <button class="scan-button" on:click={triggerScan} disabled={scanning} aria-label="Scan files">
        {scanning ? 'Scanning...' : 'Scan Files'}
      </button>
      <UploadComponent on:fileUpload={handleFileUpload} {currentPath} />
    </div>
  </div>

  {#if error}
    <div class="error-message">{error}</div>
  {/if}

  {#if loading}
    <div class="loading-indicator">Loading files...</div>
  {:else}
    <div class="view-toggle">
      <button on:click={toggleView} aria-label="Toggle view mode">
        {viewMode === 'table' ? 'Grid View' : 'List View'}
      </button>
    </div>
    {#if viewMode === 'table'}
      <table class="file-table">
        <thead>
          <tr>
            <th>Name</th>
            <th>Size</th>
            <th>Last Modified</th>
          </tr>
        </thead>
        <tbody>
          {#each files as file}
            <tr>
              <td on:click={() => {
                if (file.type === 'directory') {
                  loadDirectory(file.path);
                } else {
                  window.open('/dav' + file.path, '_blank');
                }
              }} style="cursor: pointer;">
                <!-- Icon and name displayed directly in table cell -->
                <span class="file-icon-inline">{getIcon(file)}</span> {file.name}
              </td>
              <td>{formatSize(file.size)}</td>
              <td>{file.lastModified}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    {:else}
      <div class="file-grid">
        {#each files as file (file.path)}
          <FileCard {file} on:select={navigateToDirectory} />
        {/each}
      </div>
    {/if}
  {/if}
</div>
