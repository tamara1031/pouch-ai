<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { Shield, Key as KeyIcon, Clock, Activity, CalendarDays, MoreHorizontal } from 'lucide-svelte';
  import type { Key } from '../../../domain/key/Key';

  export let keyData: Key;

  const dispatch = createEventDispatcher();
  
  // Format currency
  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(amount);
  };
</script>

<div class="key-card">
  <div class="card-header">
    <div class="key-identity">
      <div class="provider-icon">
        <Shield size={18} />
      </div>
      <div>
        <h3>{keyData.name}</h3>
        <span class="provider-badge">{keyData.provider}</span>
      </div>
    </div>
    <div class="card-actions">
        <button class="icon-btn" on:click={() => dispatch('edit', keyData)}>
            <MoreHorizontal size={18} />
        </button>
    </div>
  </div>

  <div class="key-value-section">
    <div class="key-box">
      <span class="key-icon-wrapper">
        <KeyIcon size={14} />
      </span>
      <span class="prefix">{keyData.prefix}</span>
      <span class="dots">••••••••••••••••</span>
      <span class="hash">{keyData.keyHash.substring(0, 4)}</span>
    </div>
  </div>

  <div class="stats-grid">
    <div class="stat-item">
      <div class="stat-label">
        <Activity size={14} />
        <span>Usage</span>
      </div>
      <div class="stat-value">
        {formatCurrency(keyData.budget.usage)}
        <span class="stat-sub">/ {formatCurrency(keyData.budget.limit)}</span>
      </div>
      <div class="progress-bar">
        <div 
          class="progress-fill" 
          style="width: {Math.min(keyData.usagePercentage, 100)}%"
          class:warning={keyData.usagePercentage > 80}
          class:danger={keyData.usagePercentage > 95}
        ></div>
      </div>
    </div>

    <div class="stat-item">
      <div class="stat-label">
        <Clock size={14} />
        <span>Rate Limit</span>
      </div>
      <div class="stat-value">
        {keyData.rateLimit.limit}
        <span class="stat-sub">req/{keyData.rateLimit.period}</span>
      </div>
    </div>
  
  {#if keyData.expiresAt}
    <div class="stat-item">
        <div class="stat-label">
            <CalendarDays size={14} />
            <span>Expires</span>
        </div>
        <div class="stat-value text-sm">
             {keyData.expiresAt.toLocaleDateString()}
        </div>
    </div>
  {/if}

  </div>
  
  {#if keyData.isMock}
      <div class="mock-badge">MOCK</div>
  {/if}
</div>

<style>
  .key-card {
    background: #171717;
    border: 1px solid #333;
    border-radius: 12px;
    padding: 1.5rem;
    transition: all 0.2s;
    position: relative;
    overflow: hidden;
  }

  .key-card:hover {
    border-color: #555;
    transform: translateY(-2px);
    box-shadow: 0 10px 20px rgba(0,0,0,0.2);
  }

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 1.25rem;
  }

  .key-identity {
    display: flex;
    gap: 0.75rem;
  }

  .provider-icon {
    width: 36px;
    height: 36px;
    background: rgba(59, 130, 246, 0.1);
    color: var(--color-primary);
    border-radius: 8px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  h3 {
    font-size: 1rem;
    font-weight: 600;
    margin-bottom: 0.1rem;
    color: #fff;
  }

  .provider-badge {
    font-size: 0.75rem;
    color: #888;
    background: #222;
    padding: 2px 6px;
    border-radius: 4px;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .icon-btn {
      background: transparent;
      border: none;
      color: #666;
      cursor: pointer;
      padding: 4px;
      border-radius: 4px;
      transition: color 0.2s;
  }
  
  .icon-btn:hover {
      color: #fff;
      background: #2a2a2a;
  }

  .key-value-section {
    margin-bottom: 1.25rem;
    background: #0a0a0a;
    padding: 0.75rem;
    border-radius: 6px;
    border: 1px solid #222;
  }

  .key-box {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.875rem;
    color: #aaa;
  }
  
  .key-icon-wrapper {
      color: #555;
      display: flex;
      align-items: center;
  }

  .prefix {
    color: var(--color-primary);
  }
  
  .dots {
      letter-spacing: -1px;
      color: #444;
  }
  
  .hash {
      color: #fff;
  }

  .stats-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1rem;
  }
  
  .stat-item {
      display: flex;
      flex-direction: column;
      gap: 0.25rem;
  }
  
  .stat-label {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      font-size: 0.75rem;
      color: #666;
      text-transform: uppercase;
      font-weight: 600;
      letter-spacing: 0.05em;
  }
  
  .stat-value {
      font-size: 0.9rem;
      font-weight: 500;
      color: #ddd;
  }
  
  .text-sm {
      font-size: 0.8rem;
  }
  
  .stat-sub {
      font-size: 0.75rem;
      color: #666;
  }
  
  .progress-bar {
      height: 4px;
      background: #333;
      border-radius: 2px;
      margin-top: 4px;
      overflow: hidden;
  }
  
  .progress-fill {
      height: 100%;
      background: var(--color-primary);
      border-radius: 2px;
  }
  
  .progress-fill.warning {
      background: #f59e0b;
  }
  
  .progress-fill.danger {
      background: #ef4444;
  }
  
  .mock-badge {
      position: absolute;
      top: 1rem;
      right: 1rem;
      font-size: 0.6rem;
      font-weight: 800;
      color: #f59e0b;
      border: 1px solid #f59e0b;
      padding: 2px 4px;
      border-radius: 4px;
      transform: rotate(15deg);
      opacity: 0.5;
  }
</style>
