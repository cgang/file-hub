<script>
  import { createEventDispatcher } from 'svelte';
  export let breadcrumbs = [];

  const dispatch = createEventDispatcher();
</script>

<div class="nav-bar">
  <nav>
    <ul class="breadcrumbs">
      {#each breadcrumbs as breadcrumb, index}
        {#if index === breadcrumbs.length - 1}
          <li class="breadcrumb-item">
            <span class="current-path">{breadcrumb.name}</span>
          </li>
        {:else}
          <li class="breadcrumb-item">
            <a href="#" class="breadcrumb-link" on:click|preventDefault={() => dispatch('navigate', breadcrumb.path)}>
              {breadcrumb.name}
            </a>
            <span class="breadcrumb-separator">/</span>
          </li>
        {/if}
      {/each}
    </ul>
  </nav>
</div>