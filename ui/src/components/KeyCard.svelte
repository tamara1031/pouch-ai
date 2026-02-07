<script lang="ts">
  import type { Key } from '../lib/types';
  import { Trash2, Edit2, Key as KeyIcon, Copy } from 'lucide-svelte';
  import { createEventDispatcher } from 'svelte';

  export let keyData: Key;
  
  const dispatch = createEventDispatcher();

  function copyKey() {
    // In a real app we might only have the prefix, but here we assume key_hash is for display or similar? 
    // Actually the raw key is only returned on creation.
    // So we can't copy it here usually. 
    // But let's verify what `keyData` has. `prefix`.
    // Maybe copy ID or Name?
  }

  function formatDate(str?: string) {
    if (!str) return 'Never';
    return new Date(str).toLocaleDateString();
  }

  // Calculate usage percentage
  $: usagePercent = (keyData.budget.usage / keyData.budget.limit) * 100;
  $: isOverBudget = keyData.budget.limit > 0 && keyData.budget.usage >= keyData.budget.limit;
</script>

<div class="card">
  <div class="header">
    <div class="title-group">
      <div class="icon-box">
        <KeyIcon size={20} />
      </div>
      <div>
        <h3 class="name">{keyData.name}</h3>
        <span class="prefix">Prefix: <code>{keyData.prefix}</code></span>
      </div>
    </div>
    <div class="badge {keyData.is_mock ? 'mock' : 'live'}">
      {keyData.provider}
    </div>
  </div>

  <div class="stats">
    <div class="stat">
      <span class="label">Usage</span>
      <div class="progress-bar">
        <div 
          class="fill" 
          style="width: {Math.min(usagePercent, 100)}%; background-color: {isOverBudget ? 'var(--color-danger)' : 'var(--color-primary)'}"
        ></div>
      </div>
      <span class="value">${keyData.budget.usage.toFixed(4)} / ${keyData.budget.limit.toFixed(2)}</span>
    </div>
    <div class="stat">
      <span class="label">Expires</span>
      <span class="value">{formatDate(keyData.expires_at)}</span>
    </div>
  </div>

  <div class="actions">
    <button class="btn-icon" on:click={() => dispatch('edit', keyData)}>
      <Edit2 size={16} />
    </button>
    <button class="btn-icon danger" on:click={() => dispatch('delete', keyData.id)}>
      <Trash2 size={16} />
    </button>
  </div>
</div>

<style>
  .card {
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 12px;
    padding: 1.25rem;
    display: flex;
    flex-direction: column;
    gap: 1rem;
    transition: all 0.2s ease;
    backdrop-filter: blur(10px);
  }
  
  .card:hover {
    transform: translateY(-2px);
    background: rgba(255, 255, 255, 0.05);
    border-color: rgba(255, 255, 255, 0.2);
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.2);
  }

  .header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
  }

  .title-group {
    display: flex;
    gap: 0.75rem;
    align-items: center;
  }

  .icon-box {
    width: 40px;
    height: 40px;
    border-radius: 10px;
    background: linear-gradient(135deg, rgba(59, 130, 246, 0.2), rgba(37, 99, 235, 0.1));
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--color-primary);
    border: 1px solid rgba(59, 130, 246, 0.2);
  }

  .name {
    font-size: 1rem;
    font-weight: 600;
    color: var(--color-text-main);
  }

  .prefix {
    font-size: 0.75rem;
    color: var(--color-text-muted);
    font-family: monospace;
  }
  
  .badge {
    padding: 0.25rem 0.75rem;
    border-radius: 20px;
    font-size: 0.75rem;
    font-weight: 500;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .badge.mock {
    background: rgba(239, 68, 68, 0.1);
    color: #ef4444;
    border: 1px solid rgba(239, 68, 68, 0.2);
  }

  .badge.live {
    background: rgba(16, 185, 129, 0.1);
    color: #10b981;
    border: 1px solid rgba(16, 185, 129, 0.2);
  }

  .stats {
    display: flex;
    gap: 1.5rem;
    padding-top: 0.5rem;
    border-top: 1px solid rgba(255, 255, 255, 0.05);
  }

  .stat {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    flex: 1;
  }

  .label {
    font-size: 0.75rem;
    color: var(--color-text-muted);
  }

  .value {
    font-size: 0.875rem;
    font-weight: 500;
  }

  .progress-bar {
    height: 4px;
    background: rgba(255, 255, 255, 0.1);
    border-radius: 2px;
    overflow: hidden;
    margin-bottom: 0.25rem;
  }

  .fill {
    height: 100%;
    transition: width 0.3s ease;
  }

  .actions {
    display: flex;
    justify-content: flex-end;
    gap: 0.5rem;
    margin-top: auto;
  }

  .btn-icon {
    background: transparent;
    border: none;
    color: var(--color-text-muted);
    padding: 0.5rem;
    border-radius: 6px;
    transition: all 0.2s;
  }

  .btn-icon:hover {
    background: rgba(255, 255, 255, 0.1);
    color: var(--color-text-main);
  }

  .btn-icon.danger:hover {
    background: rgba(239, 68, 68, 0.1);
    color: var(--color-danger);
  }
</style>
