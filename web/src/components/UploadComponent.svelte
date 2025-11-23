<script>
  import { createEventDispatcher } from 'svelte';
  export let currentPath;

  const dispatch = createEventDispatcher();

  // Handle file input change
  function handleFileInput(event) {
    const files = event.target.files;
    if (files.length > 0) {
      dispatch('fileUpload', files[0]); // Only upload the first file for MVP
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