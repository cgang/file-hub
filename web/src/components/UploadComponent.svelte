<script>
  import { createEventDispatcher } from 'svelte';
  export let currentPath;

  const dispatch = createEventDispatcher();

  // Handle file input change
  function handleFileInput(event) {
    const files = event.target.files;
    if (files.length > 0) {
      // Pass both the file and current path to the parent component
      dispatch('fileUpload', { file: files[0], path: currentPath });
    }
  }

  // Handle direct upload via input click
  function triggerFileInput() {
    document.getElementById('fileInput').click();
  }
</script>

<button class="upload-button" on:click={triggerFileInput}>
  Upload File
</button>
<input
  id="fileInput"
  type="file"
  style="display: none;"
  on:change={handleFileInput}
/>