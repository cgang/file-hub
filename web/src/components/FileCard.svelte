<script>
  export let file;
  import { createEventDispatcher } from 'svelte';

  const dispatch = createEventDispatcher();

  // Determine icon based on file type
  function getIcon() {
    if (file.type === 'directory') {
      return '📁';
    }

    // Determine icon based on file extension
    const ext = file.name.split('.').pop().toLowerCase();
    switch (ext) {
      case 'pdf':
        return '📄';
      case 'jpg':
      case 'jpeg':
      case 'png':
      case 'gif':
      case 'webp':
        return '🖼️';
      case 'txt':
        return '📝';
      case 'doc':
      case 'docx':
        return '📝';
      case 'xls':
      case 'xlsx':
        return '📊';
      case 'ppt':
      case 'pptx':
        return '.slides';
      case 'zip':
      case 'rar':
      case '7z':
        return '📦';
      case 'mp3':
      case 'wav':
      case 'flac':
        return '🎵';
      case 'mp4':
      case 'avi':
      case 'mov':
        return '🎬';
      default:
        return '📎';
    }
  }

  // Format file size for display
  function formatSize(bytes) {
    if (bytes === undefined || bytes === null) return '';
    
    if (bytes < 1024) return bytes + ' B';
    else if (bytes < 1048576) return (bytes / 1024).toFixed(1) + ' KB';
    else if (bytes < 1073741824) return (bytes / 1048576).toFixed(1) + ' MB';
    else return (bytes / 1073741824).toFixed(1) + ' GB';
  }

  // Format date for display
  function formatDate(dateString) {
    if (!dateString) return '';
    const date = new Date(dateString);
    return date.toLocaleDateString();
  }

  // Handle click on file card
  function handleClick() {
    if (file.type === 'directory') {
      dispatch('select', file.path);
    } else {
      // For files, we might open in a new tab or handle differently
      window.open(file.path, '_blank');
    }
  }
</script>

<div class="file-card" on:click={handleClick}>
  <div class="file-icon">{getIcon()}</div>
  <div class="file-name">{file.name}</div>
  {#if file.type !== 'directory'}
    <div class="file-meta">{formatSize(file.size)}</div>
    <div class="file-meta">{formatDate(file.lastModified)}</div>
  {/if}
</div>