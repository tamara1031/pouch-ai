<script lang="ts">
  import { createEventDispatcher, onMount, onDestroy } from 'svelte';
  import { X } from 'lucide-svelte';
  import { fade, fly } from 'svelte/transition';
  
  export let title = '';
  export let size: 'sm' | 'md' | 'lg' = 'md';
  
  const dispatch = createEventDispatcher();
  
  function close() {
    dispatch('close');
  }
  
  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      close();
    }
  }

  // Prevent body scroll when modal is open
    // Since we destroy the component when closed (via #if in parent), this works.
  onMount(() => {
    document.body.style.overflow = 'hidden';
  });
  
  onDestroy(() => {
    document.body.style.overflow = '';
  });
</script>

<svelte:window on:keydown={handleKeydown} />

<div class="backdrop" transition:fade={{ duration: 200 }} on:click|self={close}>
  <div 
    class="modal modal-{size}"
    role="dialog"
    aria-modal="true"
    transition:fly={{ y: 20, duration: 300, opacity: 0 }}
  >
    <div class="header">
      <h2>{title}</h2>
      <button class="close-btn" on:click={close} aria-label="Close modal">
        <X size={20} />
      </button>
    </div>
    
    <div class="content">
      <slot />
    </div>
    
    {#if $$slots.footer}
      <div class="footer">
        <slot name="footer" />
      </div>
    {/if}
  </div>
</div>

<style>
  .backdrop {
    position: fixed;
    top: 0;
    left: 0;
    width: 100vw;
    height: 100vh;
    background: rgba(0, 0, 0, 0.6);
    backdrop-filter: blur(8px);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 1000;
    padding: 1rem;
  }

  .modal {
    background: #111; /* Fallback */
    background: linear-gradient(180deg, 
      hsla(var(--surface-hue), var(--surface-sat), 10%, 0.95), 
      hsla(var(--surface-hue), var(--surface-sat), 8%, 0.98)
    );
    border: var(--glass-border);
    border-radius: var(--radius-lg);
    box-shadow: 
      0 0 0 1px rgba(255,255,255,0.05),
      0 25px 50px -12px rgba(0, 0, 0, 0.7);
    display: flex;
    flex-direction: column;
    max-height: 90vh;
    width: 100%;
    position: relative;
    overflow: hidden;
  }
  
  /* Sizes */
  .modal-sm { max-width: 400px; }
  .modal-md { max-width: 550px; }
  .modal-lg { max-width: 800px; }

  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1.5rem 1.5rem 1rem;
  }

  h2 {
    font-size: 1.25rem;
    font-weight: 600;
    color: var(--text-main);
    letter-spacing: -0.01em;
  }

  .close-btn {
    color: var(--text-muted);
    padding: 0.5rem;
    border-radius: var(--radius-sm);
    transition: all 0.2s;
    background: transparent;
    border: none;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .close-btn:hover {
    color: var(--text-main);
    background: rgba(255, 255, 255, 0.1);
  }

  .content {
    padding: 0 1.5rem 1.5rem;
    overflow-y: auto;
    flex: 1;
  }

  .footer {
    padding: 1rem 1.5rem;
    border-top: 1px solid rgba(255, 255, 255, 0.05);
    background: rgba(0, 0, 0, 0.2);
    display: flex;
    justify-content: flex-end;
    gap: 0.75rem;
  }
</style>
