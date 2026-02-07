<script lang="ts">
  import { Wallet, Plus } from 'lucide-svelte';
  import KeyList from './components/KeyList.svelte';
  import CreateKeyModal from './components/modals/CreateKeyModal.svelte';

  let showCreateModal = false;
  let keyList: any; // Reference to KeyList component

  function handleKeyCreated() {
    showCreateModal = false;
    // Reload keys
    if (keyList) {
      keyList.reload();
    }
  }
</script>

<div class="app-layout">
  <nav class="navbar">
    <div class="container nav-content">
      <div class="brand">
        <div class="logo">
          <Wallet size={24} />
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

    <KeyList bind:this={keyList} />
  </main>

  {#if showCreateModal}
    <CreateKeyModal 
      on:close={() => showCreateModal = false} 
      on:created={handleKeyCreated}
    />
  {/if}
</div>

<style>
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
    width: 36px;
    height: 36px;
    background: var(--color-primary);
    border-radius: 8px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: white;
    box-shadow: 0 0 20px rgba(59, 130, 246, 0.4);
  }

  h1 {
    font-size: 1.25rem;
    font-weight: 700;
    letter-spacing: -0.02em;
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
    border-radius: var(--radius);
    font-weight: 600;
    display: flex;
    align-items: center;
    gap: 0.5rem;
    transition: all 0.2s;
    box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
  }

  .btn-pri:hover {
    background: var(--color-primary-hover);
    transform: translateY(-1px);
    box-shadow: 0 6px 16px rgba(59, 130, 246, 0.4);
  }

  .btn-pri:active {
    transform: translateY(0);
  }
</style>
