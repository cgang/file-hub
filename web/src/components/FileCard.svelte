<script>
  export let file = {};
  import { createEventDispatcher } from 'svelte';

  const dispatch = createEventDispatcher();

  import { getIcon, formatSize } from '../utils/fileUtils.js';
  import { davBasePath } from '../utils/webdav.js';


  // Format date for display
  function formatDate(dateString) {
    if (!dateString) return '';
    const date = new Date(dateString);
    return date.toLocaleString();
  }

  // Handle click on file card
  function handleClick() {
    if (file && file.type === 'directory') {
      // Always ensure directory path ends with '/'
      let dirPath = file.path || '/';
      if (dirPath !== '/' && !dirPath.endsWith('/')) {
        dirPath = dirPath + '/';
      }
      dispatch('select', dirPath);
    } else if (file && file.path) {
      // For files, we might open in a new tab or handle differently
      window.open(davBasePath + file.path, '_blank');
    }
  }
</script>

<button type="button" class="file-card" on:click={handleClick} on:keydown={(e) => (e.key === 'Enter' || e.key === ' ') && handleClick()} aria-label={"Select " + (file && file.type ? file.type : '') + ": " + (file && file.name ? file.name : '')}>
  <div class="file-icon">{getIcon(file)}</div>
  <div class="file-name">{file && file.name ? file.name : ''}</div>
  {#if file && file.type !== 'directory'}
    <div class="file-meta">{formatSize(file.size)}</div>
    <div class="file-meta">{formatDate(file.lastModified)}</div>
  {/if}
</button>
