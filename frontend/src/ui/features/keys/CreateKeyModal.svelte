<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import { Loader2, Copy, Check } from 'lucide-svelte';
  import { keyStore } from '../../../application/stores/keyStore';
  import Modal from '../../components/common/Modal.svelte';
  import Input from '../../components/common/Input.svelte';
  import Button from '../../components/common/Button.svelte';
  
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
  let copied = false;

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
          copied = true;
          setTimeout(() => copied = false, 2000);
      }
  }
</script>

<Modal title={createdKey ? 'Key Created' : 'Create New Key'} on:close={handleClose}>
    {#if createdKey}
        <div class="success-content">
            <p class="success-msg">Your API Key has been generated securely.</p>
            
            <div class="key-display">
                <code>{createdKey}</code>
                <button class="btn-copy" on:click={copyKey} title="Copy to clipboard">
                    {#if copied}
                        <Check size={16} />
                    {:else}
                        <Copy size={16} />
                    {/if}
                </button>
            </div>
            
            <div class="warning-box">
                Make sure to copy your key now. You won't be able to see it again!
            </div>
        </div>
    {:else}
        <form id="create-key-form" on:submit|preventDefault={handleSubmit}>
          <div class="form-section">
              <Input 
                id="name" 
                label="Name" 
                placeholder="e.g. Production App Key" 
                bind:value={formData.name} 
                required 
              />

              <div class="form-group">
                <label for="provider">Provider <span class="required">*</span></label>
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
          </div>

          <div class="row">
            <Input 
                id="budget" 
                type="number" 
                label="Budget Limit ($)" 
                bind:value={formData.budget_limit} 
                step="0.01" 
                min="0" 
            />
            
            <Input 
                id="expires" 
                type="number" 
                label="Expires In (Days)" 
                bind:value={formData.expires_in_days} 
                min="0" 
            />
          </div>

          <div class="row">
            <Input 
                id="rate" 
                type="number" 
                label="Rate Limit (req)" 
                bind:value={formData.rate_limit} 
                min="0" 
            />
            
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
              <label class="checkbox-label">
                  <input type="checkbox" bind:checked={formData.is_mock} />
                  <span>Mock Provider (for testing)</span>
              </label>
          </div>

          {#if error}
            <div class="error-msg">{error}</div>
          {/if}
        </form>
    {/if}
    
    <div slot="footer">
        {#if createdKey}
            <Button variant="primary" on:click={handleClose}>Done</Button>
        {:else}
            <Button variant="ghost" on:click={handleClose}>Cancel</Button>
            <Button type="submit" variant="primary" loading={submitting} on:click={handleSubmit}>
                Create Key
            </Button>
        {/if}
    </div>
</Modal>

<style>
  .success-content {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
    padding: 0.5rem 0;
  }
  
  .success-msg {
      color: var(--text-muted);
  }

  form {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
    padding: 0.5rem 0;
  }
  
  .form-section {
      display: flex;
      flex-direction: column;
      gap: 1.5rem;
  }

  .form-group {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    flex: 1;
  }
  
  .checkbox-group {
      margin-top: 0.5rem;
  }
  
  .checkbox-label {
      display: flex;
      align-items: center;
      gap: 0.75rem;
      color: var(--text-main);
      cursor: pointer;
      font-weight: 500;
  }
  
  input[type="checkbox"] {
      width: 1.1rem;
      height: 1.1rem;
      accent-color: var(--color-primary);
      background: var(--color-surface);
      border: var(--glass-border);
      cursor: pointer;
  }

  label {
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--text-muted);
    margin-left: 2px;
  }
  
  .required {
      color: var(--color-primary);
  }
  
  select {
      background: var(--color-surface-glass);
      border: var(--glass-border);
      padding: 0.6rem 0.8rem;
      border-radius: var(--radius-sm);
      color: var(--text-main);
      font-family: inherit;
      font-size: 0.95rem;
      width: 100%;
      cursor: pointer;
      backdrop-filter: blur(12px);
      transition: all 0.2s;
  }
  
  select:focus {
      outline: none;
      border-color: var(--color-primary);
      box-shadow: 0 0 0 2px var(--color-primary-glow);
  }
  
  .select-wrapper {
      position: relative;
  }

  .loading-input {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    color: var(--text-muted);
    font-size: 0.875rem;
    padding: 0.6rem;
  }

  .row {
    display: flex;
    gap: 1rem;
  }
  
  .key-display {
      background: rgba(0, 0, 0, 0.3);
      padding: 1rem;
      border-radius: var(--radius-md);
      border: 1px solid rgba(255, 255, 255, 0.1);
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 1rem;
  }
  
  code {
      font-family: var(--font-mono);
      color: var(--color-primary);
      font-size: 1.1rem;
      word-break: break-all;
  }
  
  .btn-copy {
      background: rgba(255, 255, 255, 0.1);
      border: 1px solid rgba(255, 255, 255, 0.1);
      color: var(--text-main);
      width: 36px;
      height: 36px;
      border-radius: var(--radius-sm);
      cursor: pointer;
      display: flex;
      align-items: center;
      justify-content: center;
      transition: all 0.2s;
  }
  
  .btn-copy:hover {
      background: rgba(255, 255, 255, 0.2);
  }
  
  .warning-box {
      background: rgba(245, 158, 11, 0.1);
      border: 1px solid rgba(245, 158, 11, 0.2);
      color: #fbbf24;
      padding: 1rem;
      border-radius: var(--radius-md);
      font-size: 0.9rem;
  }

  .error-msg {
    color: var(--color-danger);
    font-size: 0.875rem;
    background: rgba(239, 68, 68, 0.1);
    padding: 0.75rem;
    border-radius: var(--radius-sm);
  }
  
  :global(.spin) {
    animation: spin 1s linear infinite;
  }
  
  @keyframes spin {
    from { transform: rotate(0deg); }
    to { transform: rotate(360deg); }
  }
</style>
