<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { keyStore } from '../../../application/stores/keyStore';
  import type { Key } from '../../../domain/key/Key';
  import KeyCard from './KeyCard.svelte';
  import { Loader2 } from 'lucide-svelte';
  
  let keys: Key[] = [];
  let loading = false;
  let error: string | null = null;
  
  // Subscribe to store
  const unsubscribe = keyStore.subscribe(state => {
      keys = state.keys;
      loading = state.loading;
      error = state.error;
  });

  onMount(() => {
    keyStore.loadKeys();
  });
  
  onDestroy(() => {
      unsubscribe();
  });
</script>

<div class="container">
  {#if loading && keys.length === 0}
    <div class="center">
      <Loader2 class="spin" size={32} />
    </div>
  {:else if error}
    <div class="error">
      <p>Error loading keys: {error}</p>
      <button on:click={() => keyStore.loadKeys()}>Retry</button>
    </div>
  {:else if keys.length === 0}
    <div class="empty">
      <p>No keys found. Create one to get started.</p>
    </div>
  {:else}
    <div class="grid">
      {#each keys as key (key.id)}
        <KeyCard 
          keyData={key} 
          on:edit 
          on:delete 
        />
      {/each}
    </div>
  {/if}
</div>

<style>
  .container {
    width: 100%;
  }

  .center {
    display: flex;
    justify-content: center;
    padding: 4rem;
    color: #666;
  }
  
  :global(.spin) {
    animation: spin 1s linear infinite;
  }
  
  @keyframes spin {
    from { transform: rotate(0deg); }
    to { transform: rotate(360deg); }
  }

  .error {
    text-align: center;
    color: #ef4444;
    padding: 2rem;
  }

  .empty {
    text-align: center;
    padding: 4rem;
    color: #666;
    border: 2px dashed rgba(255, 255, 255, 0.1);
    border-radius: 12px;
  }

  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
    gap: 1.5rem;
  }
</style>
