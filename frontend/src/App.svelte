<script lang="ts">
  import { Wallet, Plus } from 'lucide-svelte';
  import KeyList from './ui/features/keys/KeyList.svelte';
  import CreateKeyModal from './ui/features/keys/CreateKeyModal.svelte';
  import { keyStore } from './application/stores/keyStore';

  let showCreateModal = false;

  function handleKeyCreated() {
    // modal handles showing the key, we just refresh the list which is done via store
    // CreateKeyModal emits 'created', we can use it to close if we want, 
    // but the modal logic I wrote keeps it open to show the key.
    // So actually we might just want to reload the list in the background.
    keyStore.loadKeys();
  }
</script>

<div class="app-layout">
  <nav class="navbar">
    <div class="container nav-content">
      <div class="brand">
        <div class="logo">
          <Wallet size={20} />
        </div>
        <h1>Pouch AI</h1>
      </div>
      <div class="nav-actions">
        <!-- Add more nav items here if needed -->
      </div>
    </div>
  </nav>

  <main class="container main-content">
    <div class="page-header">
      <div>
        <h2>API Keys</h2>
        <p class="subtitle">Manage access to your AI providers</p>
      </div>
      <button class="btn-pri" on:click={() => showCreateModal = true}>
        <Plus size={18} /> New Key
      </button>
    </div>

    <KeyList />
  </main>

  {#if showCreateModal}
    <CreateKeyModal 
      on:close={() => showCreateModal = false} 
      on:created={handleKeyCreated}
    />
  {/if}
</div>

<style>
  :global(:root) {
    --color-primary: #3b82f6;
    --color-primary-hover: #2563eb;
    --color-bg: #0a0a0a;
    --color-surface: #171717;
    --color-text-main: #fff;
    --color-text-muted: #a3a3a3;
    --color-border: #333;
    
    font-family: 'Inter', system-ui, -apple-system, sans-serif;
  }

  :global(body) {
    background-color: var(--color-bg);
    color: var(--color-text-main);
    margin: 0;
  }
  
  :global(*) {
      box-sizing: border-box;
  }

  .app-layout {
    min-height: 100vh;
    background: linear-gradient(to bottom, #0a0a0a, #111);
  }

  .container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 1.5rem;
  }

  .navbar {
    height: 64px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.05);
    background: rgba(10, 10, 10, 0.8);
    backdrop-filter: blur(12px);
    position: sticky;
    top: 0;
    z-index: 50;
  }

  .nav-content {
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .brand {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .logo {
    width: 32px;
    height: 32px;
    background: var(--color-primary);
    border-radius: 8px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: white;
    box-shadow: 0 0 15px rgba(59, 130, 246, 0.3);
  }

  h1 {
    font-size: 1.1rem;
    font-weight: 700;
    letter-spacing: -0.01em;
  }

  .main-content {
    padding-top: 3rem;
    padding-bottom: 4rem;
  }

  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 2.5rem;
  }

  h2 {
    font-size: 1.875rem;
    font-weight: 700;
    letter-spacing: -0.03em;
    margin-bottom: 0.25rem;
  }

  .subtitle {
    color: var(--color-text-muted);
  }

  .btn-pri {
    background: var(--color-primary);
    color: white;
    border: none;
    padding: 0.6rem 1.2rem;
    border-radius: 8px;
    font-weight: 600;
    display: flex;
    align-items: center;
    gap: 0.5rem;
    cursor: pointer;
    transition: all 0.2s;
    box-shadow: 0 4px 12px rgba(59, 130, 246, 0.2);
  }

  .btn-pri:hover {
    background: var(--color-primary-hover);
    transform: translateY(-1px);
    box-shadow: 0 6px 16px rgba(59, 130, 246, 0.3);
  }

  .btn-pri:active {
    transform: translateY(0);
  }
</style>
