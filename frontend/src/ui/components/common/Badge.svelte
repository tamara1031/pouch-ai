<script lang="ts">
  export let variant: 'default' | 'primary' | 'success' | 'warning' | 'danger' = 'default';
</script>

<span class="badge badge-{variant}">
  <slot />
</span>

<style>
  .badge {
    display: inline-flex;
    align-items: center;
    padding: 0.25rem 0.6rem;
    border-radius: 9999px;
    font-size: 0.75rem;
    font-weight: 600;
    letter-spacing: 0.02em;
    white-space: nowrap;
    line-height: 1;
  }

  .badge-default {
    background: rgba(255, 255, 255, 0.1);
    color: var(--text-muted);
    border: 1px solid rgba(255, 255, 255, 0.05);
  }

  .badge-primary {
    background: hsla(var(--primary-hue), var(--primary-sat), var(--primary-lit), 0.15);
    color: hsl(var(--primary-hue), var(--primary-sat), 75%);
    border: 1px solid hsla(var(--primary-hue), var(--primary-sat), var(--primary-lit), 0.2);
    box-shadow: 0 0 10px hsla(var(--primary-hue), var(--primary-sat), var(--primary-lit), 0.1);
  }

  .badge-success {
    background: hsla(var(--success-hue), var(--success-sat), var(--success-lit), 0.15);
    color: hsl(var(--success-hue), var(--success-sat), 75%);
    border: 1px solid hsla(var(--success-hue), var(--success-sat), var(--success-lit), 0.2);
  }

  .badge-danger {
    background: hsla(var(--danger-hue), var(--danger-sat), var(--danger-lit), 0.15);
    color: hsl(var(--danger-hue), var(--danger-sat), 75%);
    border: 1px solid hsla(var(--danger-hue), var(--danger-sat), var(--danger-lit), 0.2);
  }

  .badge-warning {
    background: hsla(35, 90%, 50%, 0.15);
    color: hsl(35, 90%, 75%);
    border: 1px solid hsla(35, 90%, 50%, 0.2);
  }
</style>
