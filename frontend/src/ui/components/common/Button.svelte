<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  
  export let type: 'button' | 'submit' | 'reset' = 'button';
  export let variant: 'primary' | 'secondary' | 'ghost' | 'danger' = 'primary';
  export let size: 'sm' | 'md' | 'lg' = 'md';
  export let disabled = false;
  export let loading = false;
  export let block = false;
  
  const dispatch = createEventDispatcher();
  
  $: classes = [
    'btn',
    `btn-${variant}`,
    `btn-${size}`,
    block ? 'btn-block' : '',
    loading ? 'btn-loading' : ''
  ].filter(Boolean).join(' ');
</script>

<button {type} class={classes} {disabled} on:click>
  {#if loading}
    <span class="spinner"></span>
    <span class="loading-text"><slot /></span>
  {:else}
    <slot />
  {/if}
</button>

<style>
  .btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
    font-weight: 500;
    border-radius: var(--radius-sm);
    transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
    position: relative;
    overflow: hidden;
  }

  .btn:active:not(:disabled) {
    transform: scale(0.98);
  }

  .btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  /* Variants */
  .btn-primary {
    background: var(--color-primary);
    color: white;
    box-shadow: 0 4px 12px rgba(59, 130, 246, 0.25);
  }

  .btn-primary:hover:not(:disabled) {
    background: var(--color-primary-hover);
    box-shadow: 0 6px 16px rgba(59, 130, 246, 0.35);
  }

  .btn-secondary {
    background: var(--color-surface-glass);
    border: var(--glass-border);
    color: var(--text-main);
    backdrop-filter: blur(8px);
  }

  .btn-secondary:hover:not(:disabled) {
    background: var(--color-surface-hover);
    border-color: var(--color-border-hover);
  }

  .btn-ghost {
    background: transparent;
    color: var(--text-muted);
  }

  .btn-ghost:hover:not(:disabled) {
    background: var(--color-surface-hover);
    color: var(--text-main);
  }
  
  .btn-danger {
    background: rgba(239, 68, 68, 0.1);
    color: var(--color-danger);
    border: 1px solid rgba(239, 68, 68, 0.2);
  }
  
  .btn-danger:hover:not(:disabled) {
    background: rgba(239, 68, 68, 0.2);
    border-color: var(--color-danger);
  }

  /* Sizes */
  .btn-sm {
    padding: 0.4rem 0.8rem;
    font-size: 0.875rem;
    height: 32px;
  }

  .btn-md {
    padding: 0.6rem 1.2rem;
    font-size: 0.95rem;
    height: 40px;
  }

  .btn-lg {
    padding: 0.8rem 1.6rem;
    font-size: 1.1rem;
    height: 48px;
  }

  .btn-block {
    display: flex;
    width: 100%;
  }

  /* Spinner */
  .spinner {
    width: 1em;
    height: 1em;
    border: 2px solid currentColor;
    border-right-color: transparent;
    border-radius: 50%;
    animation: spin 0.75s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }
</style>
