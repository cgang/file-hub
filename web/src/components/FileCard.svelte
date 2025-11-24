<script>
  // Component properties
  export let file = {};

  // Svelte utilities
  import { createEventDispatcher } from 'svelte';
  const dispatch = createEventDispatcher();

  // Utility functions
  import { getIcon, formatSize } from '../utils/fileUtils.js';

  /**
   * Format date for display
   * @param {string} dateString - Date string to format
   * @returns {string} Formatted date string
   */
  function formatDate(dateString) {
    if (!dateString) return '';
    const date = new Date(dateString);
    return date.toLocaleString();
  }

  /**
   * Handle click on file card
   * For directories: dispatch select event
   * For files: open in new tab
   */
  function handleCardClick() {
    // Validate file object
    if (!file) return;

    if (file.type === 'directory') {
      // Ensure directory path ends with '/'
      let dirPath = file.path || '/';
      if (dirPath !== '/' && !dirPath.endsWith('/')) {
        dirPath = dirPath + '/';
      }
      dispatch('select', dirPath);
    } else if (file.path) {
      // For files, open in a new tab
      window.open('/webdav' + file.path, '_blank');
    }
  }
</script>

<button
  type="button"
  class="file-card"
  on:click={handleCardClick}
  on:keydown={(e) => (e.key === 'Enter' || e.key === ' ') && handleCardClick()}
  aria-label={`Select ${file.type || ''}: ${file.name || ''}`}
>
  <div class="file-icon">{getIcon(file)}</div>
  <div class="file-name">{file.name || ''}</div>
  {#if file.type !== 'directory'}
    <div class="file-meta">{formatSize(file.size)}</div>
    <div class="file-meta">{formatDate(file.lastModified)}</div>
  {/if}
</button>
