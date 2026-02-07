<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import { X, Loader2 } from 'lucide-svelte';
  import { api } from '../../lib/api';
  
  const dispatch = createEventDispatcher();

  let providers: string[] = [];
  let loadingProviders = true;
  let submitting = false;
  let error: string | null = null;

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
      providers = await api.getProviders();
      if (providers.length > 0) {
        formData.provider = providers[0];
      }
    } catch (e: any) {
      error = "Failed to load providers: " + e.message;
    } finally {
      loadingProviders = false;
    }
  });

  async function handleSubmit() {
    submitting = true;
    error = null;

    try {
      const expiresAt = formData.expires_in_days > 0 
        ? Math.floor(Date.now() / 1000) + (formData.expires_in_days * 86400)
        : undefined;

      await api.createKey({
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

      dispatch('created');
      dispatch('close');
    } catch (e: any) {
      error = e.message;
    } finally {
      submitting = false;
    }
  }
</script>

<div class="overlay" on:click|self={() => dispatch('close')}>
  <div class="modal">
    <div class="header">
      <h2>Create New Key</h2>
      <button class="close-btn" on:click={() => dispatch('close')}>
        <X size={20} />
      </button>
    </div>

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
          <select id="provider" bind:value={formData.provider} required>
            {#each providers as p}
              <option value={p}>{p}</option>
            {/each}
          </select>
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
          <select id="period" bind:value={formData.rate_period}>
            <option value="minute">Minute</option>
            <option value="hour">Hour</option>
            <option value="day">Day</option>
          </select>
        </div>
      </div>

      {#if error}
        <div class="error-msg">{error}</div>
      {/if}

      <div class="actions">
        <button type="button" class="btn-sec" on:click={() => dispatch('close')}>Cancel</button>
        <button type="submit" class="btn-pri" disabled={submitting}>
          {#if submitting}
            <Loader2 class="spin" size={16} /> Creating...
          {:else}
            Create Key
          {/if}
        </button>
      </div>
    </form>
  </div>
</div>

<style>
  .overlay {
    position: fixed;
    top: 0;
    left: 0;
    width: 100vw;
    height: 100vh;
    background: rgba(0, 0, 0, 0.6);
    backdrop-filter: blur(4px);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 100;
  }

  .modal {
    background: #171717;
    border: 1px solid #333;
    border-radius: 12px;
    width: 90%;
    max-width: 500px;
    box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
    animation: slideIn 0.2s ease-out;
  }

  @keyframes slideIn {
    from { opacity: 0; transform: translateY(10px); }
    to { opacity: 1; transform: translateY(0); }
  }

  .header {
    padding: 1.5rem;
    border-bottom: 1px solid #333;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  h2 {
    font-size: 1.25rem;
    font-weight: 600;
  }

  .close-btn {
    background: transparent;
    border: none;
    color: var(--color-text-muted);
    padding: 0.25rem;
  }

  .close-btn:hover {
    color: var(--color-text-main);
  }

  form {
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

  label {
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--color-text-muted);
  }

  .loading-input {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    color: var(--color-text-muted);
    font-size: 0.875rem;
  }

  .row {
    display: flex;
    gap: 1rem;
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
    padding: 0.5rem 1rem;
    border-radius: var(--radius);
    font-weight: 500;
    display: flex;
    align-items: center;
    gap: 0.5rem;
    transition: background 0.2s;
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
    color: var(--color-text-main);
    border: 1px solid var(--color-border);
    padding: 0.5rem 1rem;
    border-radius: var(--radius);
    font-weight: 500;
  }

  .btn-sec:hover {
    background: var(--color-surface-hover);
  }

  .error-msg {
    color: var(--color-danger);
    font-size: 0.875rem;
    background: rgba(239, 68, 68, 0.1);
    padding: 0.75rem;
    border-radius: var(--radius);
  }
  
  .spin {
    animation: spin 1s linear infinite;
  }
  
  @keyframes spin {
    from { transform: rotate(0deg); }
    to { transform: rotate(360deg); }
  }
</style>
