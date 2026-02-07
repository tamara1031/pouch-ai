<script lang="ts">
  import { onMount } from 'svelte';
  import { api } from '../lib/api';
  import type { Key } from '../lib/types';
  import KeyCard from './KeyCard.svelte';
  import { Loader2 } from 'lucide-svelte';

  let keys: Key[] = [];
  let loading = true;
  let error: string | null = null;

  async function loadKeys() {
    loading = true;
    try {
      keys = await api.getKeys();
    } catch (e: any) {
      error = e.message;
    } finally {
      loading = false;
    }
  }

  onMount(loadKeys);

  // Expose reload function to parent
  export const reload = loadKeys;
</script>

<div class="container">
  {#if loading}
    <div class="center">
      <Loader2 class="spin" size={32} />
    </div>
  {:else if error}
    <div class="error">
      <p>Error loading keys: {error}</p>
      <button on:click={loadKeys}>Retry</button>
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
    color: var(--color-text-muted);
  }
  
  .spin {
    animation: spin 1s linear infinite;
  }
  
  @keyframes spin {
    from { transform: rotate(0deg); }
    to { transform: rotate(360deg); }
  }

  .error {
    text-align: center;
    color: var(--color-danger);
    padding: 2rem;
  }

  .empty {
    text-align: center;
    padding: 4rem;
    color: var(--color-text-muted);
    border: 2px dashed rgba(255, 255, 255, 0.1);
    border-radius: 12px;
  }

  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
    gap: 1.5rem;
  }
</style>
