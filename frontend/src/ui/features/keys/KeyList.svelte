<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { keyStore } from '../../../application/stores/keyStore';
  import type { Key } from '../../../domain/key/Key';
  import KeyCard from './KeyCard.svelte';
  import { Loader2, Key as KeyIcon } from 'lucide-svelte';
  import Button from '../../components/common/Button.svelte';
  
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

<div class="key-list-container">
  {#if loading && keys.length === 0}
    <div class="center-state">
      <Loader2 class="spin" size={32} />
    </div>
  {:else if error}
    <div class="center-state error">
      <p>Error loading keys: {error}</p>
      <Button variant="secondary" size="sm" on:click={() => keyStore.loadKeys()}>Retry</Button>
    </div>
  {:else if keys.length === 0}
    <div class="empty-state">
      <div class="icon-wrapper">
        <KeyIcon size={32} />
      </div>
      <h3>No API Keys Found</h3>
      <p>Create a new key to start using AI providers.</p>
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
  .key-list-container {
    width: 100%;
  }

  .center-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 1rem;
    padding: 6rem;
    color: var(--text-muted);
  }

  .error {
    color: var(--color-danger);
  }
  
  :global(.spin) {
    animation: spin 1s linear infinite;
  }
  
  @keyframes spin {
    from { transform: rotate(0deg); }
    to { transform: rotate(360deg); }
  }

  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 6rem 2rem;
    text-align: center;
    background: var(--color-surface-glass);
    border: 1px dashed var(--color-border);
    border-radius: var(--radius-lg);
  }

  .icon-wrapper {
    width: 64px;
    height: 64px;
    background: rgba(255, 255, 255, 0.05);
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    margin-bottom: 1.5rem;
    color: var(--text-muted);
  }

  .empty-state h3 {
    font-size: 1.25rem;
    font-weight: 600;
    margin-bottom: 0.5rem;
    color: var(--text-main);
  }

  .empty-state p {
    color: var(--text-muted);
    max-width: 400px;
  }

  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
    gap: 1.5rem;
  }
</style>
