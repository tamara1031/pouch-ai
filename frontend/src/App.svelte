<script lang="ts">
  import { Plus } from 'lucide-svelte';
  import Navbar from './ui/components/Navbar.svelte';
  import Button from './ui/components/common/Button.svelte';
  import KeyList from './ui/features/keys/KeyList.svelte';
  import CreateKeyModal from './ui/features/keys/CreateKeyModal.svelte';
  import { keyStore } from './application/stores/keyStore';

  let showCreateModal = false;

  function handleKeyCreated() {
    keyStore.loadKeys();
    // Modal stays open to show result, user closes it manually
  }
</script>

<div class="app-layout">
  <Navbar />

  <main class="container main-content">
    <div class="page-header">
      <div>
        <h2>API Keys</h2>
        <p class="subtitle">Manage access to your AI providers</p>
      </div>
      <Button variant="primary" on:click={() => showCreateModal = true}>
        <Plus size={18} /> New Key
      </Button>
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
  .app-layout {
    min-height: 100vh;
    /* Background handled by global body style */
  }

  .container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 1.5rem;
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
    color: var(--text-main);
  }

  .subtitle {
    color: var(--text-muted);
  }
</style>

