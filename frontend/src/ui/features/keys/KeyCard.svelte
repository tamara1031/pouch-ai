<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { Shield, Key as KeyIcon, Clock, Activity, CalendarDays, MoreHorizontal } from 'lucide-svelte';
  import type { Key } from '../../../domain/key/Key';
  import Card from '../../components/common/Card.svelte';
  import Badge from '../../components/common/Badge.svelte';

  export let keyData: Key;

  const dispatch = createEventDispatcher();
  
  // Format currency
  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(amount);
  };

  function handleEdit(e: Event) {
    e.stopPropagation();
    dispatch('edit', keyData);
  }
</script>

<Card hoverable={true} on:click={() => dispatch('edit', keyData)}>
  <div class="card-header">
    <div class="key-identity">
      <div class="provider-icon">
        <Shield size={20} />
      </div>
      <div class="title-group">
        <h3>{keyData.name}</h3>
        <Badge variant={keyData.isMock ? 'warning' : 'default'}>{keyData.provider}</Badge>
      </div>
    </div>
    <!-- 
    <button class="icon-btn" on:click={handleEdit}>
        <MoreHorizontal size={18} />
    </button>
    -->
  </div>

  <div class="key-value-section">
    <div class="key-box">
      <KeyIcon size={14} class="key-icon" />
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
</Card>

<style>
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 1.25rem;
  }

  .key-identity {
    display: flex;
    gap: 1rem;
    width: 100%;
  }

  .provider-icon {
    width: 40px;
    height: 40px;
    background: linear-gradient(135deg, rgba(59, 130, 246, 0.1), rgba(37, 99, 235, 0.1));
    color: var(--color-primary);
    border-radius: var(--radius-md);
    display: flex;
    align-items: center;
    justify-content: center;
    border: 1px solid rgba(59, 130, 246, 0.1);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  }

  .title-group {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    align-items: flex-start;
  }

  h3 {
    font-size: 1rem;
    font-weight: 600;
    color: var(--text-main);
    line-height: 1.2;
  }

/*
  .icon-btn {
      background: transparent;
      border: none;
      color: var(--text-muted);
      cursor: pointer;
      padding: 4px;
      border-radius: var(--radius-sm);
      transition: all 0.2s;
  }
  
  .icon-btn:hover {
      color: var(--text-main);
      background: var(--color-surface-hover);
  }
*/

  .key-value-section {
    margin-bottom: 1.5rem;
    background: rgba(0, 0, 0, 0.2);
    padding: 0.75rem 1rem;
    border-radius: var(--radius-sm);
    border: 1px solid rgba(255, 255, 255, 0.03);
  }

  .key-box {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-family: var(--font-mono);
    font-size: 0.85rem;
    color: var(--text-muted);
  }
  
  :global(.key-icon) {
      color: var(--color-primary);
      opacity: 0.7;
  }

  .prefix {
    color: var(--color-primary);
    font-weight: 500;
  }
  
  .dots {
      letter-spacing: -1px;
      color: rgba(255, 255, 255, 0.1);
  }
  
  .hash {
      color: var(--text-main);
  }

  .stats-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1.25rem;
    padding-top: 1rem;
    border-top: 1px solid rgba(255, 255, 255, 0.05);
  }
  
  .stat-item {
      display: flex;
      flex-direction: column;
      gap: 0.4rem;
  }
  
  .stat-label {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      font-size: 0.75rem;
      color: var(--text-muted);
      text-transform: uppercase;
      font-weight: 600;
      letter-spacing: 0.05em;
  }
  
  .stat-value {
      font-size: 0.95rem;
      font-weight: 600;
      color: var(--text-main);
  }
  
  .text-sm {
      font-size: 0.85rem;
  }
  
  .stat-sub {
      font-size: 0.75rem;
      color: var(--text-muted);
      font-weight: 400;
  }
  
  .progress-bar {
      height: 4px;
      background: rgba(255, 255, 255, 0.1);
      border-radius: 2px;
      margin-top: 6px;
      overflow: hidden;
  }
  
  .progress-fill {
      height: 100%;
      background: var(--color-primary);
      border-radius: 2px;
      box-shadow: 0 0 10px var(--color-primary-glow);
  }
  
  .progress-fill.warning {
      background: #f59e0b;
      box-shadow: none;
  }
  
  .progress-fill.danger {
      background: var(--color-danger);
      box-shadow: none;
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
      opacity: 0.6;
      pointer-events: none;
  }
</style>
