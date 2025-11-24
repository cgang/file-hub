<script>
  // Component properties
  export let breadcrumbs = [];

  // Svelte utilities
  import { createEventDispatcher } from 'svelte';
  const dispatch = createEventDispatcher();
</script>

<nav class="navigation-bar" aria-label="Breadcrumb navigation">
  <ul class="breadcrumbs">
    {#each breadcrumbs as breadcrumb, index}
      {#if index === breadcrumbs.length - 1}
        <!-- Current location (last breadcrumb) -->
        <li class="breadcrumb-item">
          <span class="current-location">{breadcrumb.name}</span>
        </li>
      {:else}
        <!-- Navigable breadcrumb -->
        <li class="breadcrumb-item">
          <a
            href={`#${breadcrumb.path}`}
            class="breadcrumb-link"
            on:click|preventDefault={() => dispatch('navigate', breadcrumb.path)}
            on:keydown={(e) => (e.key === 'Enter' || e.key === ' ') && dispatch('navigate', breadcrumb.path)}
            aria-label={`Go to ${breadcrumb.name}`}
          >
            {breadcrumb.name}
          </a>
          <span class="breadcrumb-separator">/</span>
        </li>
      {/if}
    {/each}
  </ul>
</nav>