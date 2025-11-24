<script>
  // Component properties
  export let currentPath;

  // Svelte utilities
  import { createEventDispatcher } from 'svelte';
  const dispatch = createEventDispatcher();

  /**
   * Handle file input change event
   * @param {Event} event - Change event from file input
   */
  function handleFileInputChange(event) {
    const files = event.target.files;
    if (files.length > 0) {
      // Dispatch both the file and current path to the parent component
      dispatch('fileUpload', { file: files[0], path: currentPath });
    }
  }

  /**
   * Trigger file input click programmatically
   */
  function triggerFileInput() {
    document.getElementById('file-upload-input').click();
  }
</script>

<button
  class="upload-button"
  on:click={triggerFileInput}
  aria-label="Upload a file"
>
  Upload File
</button>

<input
  id="file-upload-input"
  type="file"
  class="file-input"
  on:change={handleFileInputChange}
/>