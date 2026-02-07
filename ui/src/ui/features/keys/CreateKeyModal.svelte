<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import { X, Loader2 } from 'lucide-svelte';
  import { keyStore } from '../../../application/stores/keyStore';
  
  const dispatch = createEventDispatcher();

  // Subscribe to store for providers
  let providers: string[] = [];
  keyStore.subscribe(state => {
      providers = state.providers;
  });

  let loadingProviders = true;
  let submitting = false;
  let error: string | null = null;
  let createdKey: string | null = null;

  let formData = {
    name: '',
    provider: '',
    budget_limit: 10.0,
    budget_period: 'monthly',
    rate_limit: 60,
    rate_period: 'minute',
    is_mock: false,
    mock_config: '{}',
    expires_in_days: 30, // Default to 30 days
  };

  onMount(async () => {
    try {
      await keyStore.loadProviders();
      // Auto-select first provider if available
      // Note: we need to wait for the store update or check current store value
      // Subscriptions run immediately, so providers variable might already be set if store was loaded
      if (providers.length > 0) {
        formData.provider = providers[0];
      }
    } catch (e: any) {
      error = "Failed to load providers";
    } finally {
      loadingProviders = false;
    }
  });
  
  // React to providers changes to set default
  $: if (providers.length > 0 && !formData.provider) {
      formData.provider = providers[0];
  }

  async function handleSubmit() {
    submitting = true;
    error = null;

    try {
      const expiresAt = formData.expires_in_days > 0 
        ? Math.floor(Date.now() / 1000) + (formData.expires_in_days * 86400)
        : undefined;

      const key = await keyStore.createKey({
        name: formData.name,
        provider: formData.provider,
        budget_limit: formData.budget_limit,
        budget_period: formData.budget_period,
        rate_limit: formData.rate_limit,
        rate_period: formData.rate_period,
        is_mock: formData.is_mock,
        mock_config: formData.mock_config,
        expires_at: expiresAt,
      });
      
      createdKey = key;
      dispatch('created');
      // Don't close immediately, show the key!
    } catch (e: any) {
      error = e.message;
    } finally {
      submitting = false;
    }
  }
  
  function handleClose() {
      dispatch('close');
  }
  
  function copyKey() {
      if (createdKey) {
          navigator.clipboard.writeText(createdKey);
          // Could show toast here
      }
  }
</script>

<div class="overlay" on:click|self={handleClose}>
  <div class="modal">
    <div class="header">
      <h2>{createdKey ? 'Key Created' : 'Create New Key'}</h2>
      <button class="close-btn" on:click={handleClose}>
        <X size={20} />
      </button>
    </div>

    {#if createdKey}
        <div class="success-content">
            <p>Your API Key has been generated securely.</p>
            <div class="key-display">
                <code>{createdKey}</code>
                <button class="btn-copy" on:click={copyKey}>Copy</button>
            </div>
            <div class="warning-box">
                Make sure to copy your key now. You won't be able to see it again!
            </div>
            <div class="actions">
                <button class="btn-pri" on:click={handleClose}>Done</button>
            </div>
        </div>
    {:else}
        <form on:submit|preventDefault={handleSubmit}>
          <div class="form-group">
            <label for="name">Name</label>
            <input type="text" id="name" bind:value={formData.name} placeholder="e.g. My App Key" required />
          </div>

          <div class="form-group">
            <label for="provider">Provider</label>
            {#if loadingProviders}
              <div class="loading-input"><Loader2 class="spin" size={16} /> Loading providers...</div>
            {:else}
              <div class="select-wrapper">
                  <select id="provider" bind:value={formData.provider} required>
                    {#each providers as p}
                      <option value={p}>{p}</option>
                    {/each}
                  </select>
              </div>
            {/if}
          </div>

          <div class="row">
            <div class="form-group">
              <label for="budget">Budget Limit ($)</label>
              <input type="number" id="budget" bind:value={formData.budget_limit} step="0.01" min="0" />
            </div>
            <div class="form-group">
              <label for="expires">Expires In (Days)</label>
              <input type="number" id="expires" bind:value={formData.expires_in_days} min="0" />
            </div>
          </div>

          <div class="row">
            <div class="form-group">
              <label for="rate">Rate Limit (req)</label>
              <input type="number" id="rate" bind:value={formData.rate_limit} min="0" />
            </div>
            <div class="form-group">
              <label for="period">Per Period</label>
              <div class="select-wrapper">
                  <select id="period" bind:value={formData.rate_period}>
                    <option value="minute">Minute</option>
                    <option value="hour">Hour</option>
                    <option value="day">Day</option>
                  </select>
              </div>
            </div>
          </div>
          
          <div class="form-group checkbox-group">
              <label>
                  <input type="checkbox" bind:checked={formData.is_mock} />
                  <span>Mock Provider (for testing)</span>
              </label>
          </div>

          {#if error}
            <div class="error-msg">{error}</div>
          {/if}

          <div class="actions">
            <button type="button" class="btn-sec" on:click={handleClose}>Cancel</button>
            <button type="submit" class="btn-pri" disabled={submitting}>
              {#if submitting}
                <Loader2 class="spin" size={16} /> Creating...
              {:else}
                Create Key
              {/if}
            </button>
          </div>
        </form>
    {/if}
  </div>
</div>

<style>
  .overlay {
    position: fixed;
    top: 0;
    left: 0;
    width: 100vw;
    height: 100vh;
    background: rgba(0, 0, 0, 0.7);
    backdrop-filter: blur(8px);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 100;
  }

  .modal {
    background: #111;
    border: 1px solid #333;
    border-radius: 16px;
    width: 90%;
    max-width: 500px;
    box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
    animation: slideIn 0.3s cubic-bezier(0.16, 1, 0.3, 1);
  }

  @keyframes slideIn {
    from { opacity: 0; transform: scale(0.95) translateY(10px); }
    to { opacity: 1; transform: scale(1) translateY(0); }
  }

  .header {
    padding: 1.5rem;
    border-bottom: 1px solid #222;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  h2 {
    font-size: 1.25rem;
    font-weight: 600;
    color: #fff;
  }

  .close-btn {
    background: transparent;
    border: none;
    color: #666;
    padding: 0.25rem;
    cursor: pointer;
    transition: color 0.2s;
  }

  .close-btn:hover {
    color: #fff;
  }

  form, .success-content {
    padding: 1.5rem;
    display: flex;
    flex-direction: column;
    gap: 1.25rem;
  }

  .form-group {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    flex: 1;
  }
  
  .checkbox-group label {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      color: #ccc;
      cursor: pointer;
  }
  
  input[type="checkbox"] {
      width: 1rem;
      height: 1rem;
      accent-color: var(--color-primary);
  }

  label {
    font-size: 0.875rem;
    font-weight: 500;
    color: #aaa;
  }
  
  input[type="text"], input[type="number"], select {
      background: #0a0a0a;
      border: 1px solid #333;
      padding: 0.6rem;
      border-radius: 8px;
      color: #fff;
      font-size: 0.95rem;
      transition: all 0.2s;
  }
  
  input:focus, select:focus {
      outline: none;
      border-color: var(--color-primary);
      box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.2);
  }
  
  .select-wrapper {
      position: relative;
  }
  
  /* Customizing select arrow could be done here */

  .loading-input {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    color: #666;
    font-size: 0.875rem;
  }

  .row {
    display: flex;
    gap: 1rem;
  }
  
  .key-display {
      background: #000;
      padding: 1rem;
      border-radius: 8px;
      border: 1px solid #333;
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 1rem;
  }
  
  code {
      font-family: 'JetBrains Mono', monospace;
      color: var(--color-primary);
      font-size: 1.1rem;
      word-break: break-all;
  }
  
  .btn-copy {
      background: #222;
      border: 1px solid #444;
      color: #fff;
      padding: 0.4rem 0.8rem;
      border-radius: 6px;
      cursor: pointer;
      font-size: 0.8rem;
  }
  
  .btn-copy:hover {
      background: #333;
  }
  
  .warning-box {
      background: rgba(245, 158, 11, 0.1);
      border: 1px solid rgba(245, 158, 11, 0.3);
      color: #fbbf24;
      padding: 0.75rem;
      border-radius: 8px;
      font-size: 0.875rem;
  }

  .actions {
    display: flex;
    justify-content: flex-end;
    gap: 0.75rem;
    margin-top: 1rem;
  }

  .btn-pri {
    background: var(--color-primary);
    color: white;
    border: none;
    padding: 0.6rem 1.2rem;
    border-radius: 8px;
    font-weight: 500;
    display: flex;
    align-items: center;
    gap: 0.5rem;
    transition: background 0.2s;
    cursor: pointer;
  }

  .btn-pri:hover:not(:disabled) {
    background: var(--color-primary-hover);
  }

  .btn-pri:disabled {
    opacity: 0.7;
    cursor: not-allowed;
  }

  .btn-sec {
    background: transparent;
    color: #fff;
    border: 1px solid #444;
    padding: 0.6rem 1.2rem;
    border-radius: 8px;
    font-weight: 500;
    cursor: pointer;
  }

  .btn-sec:hover {
    background: #222;
  }

  .error-msg {
    color: #ef4444;
    font-size: 0.875rem;
    background: rgba(239, 68, 68, 0.1);
    padding: 0.75rem;
    border-radius: 8px;
  }
  
  :global(.spin) {
    animation: spin 1s linear infinite;
  }
  
  @keyframes spin {
    from { transform: rotate(0deg); }
    to { transform: rotate(360deg); }
  }
</style>
