<script lang="ts">
  export let hoverable = false;
  export let padding = '1.5rem';
</script>

<div class="card" class:hoverable style="--padding: {padding}" on:click>
  <slot />
</div>

<style>
  .card {
    background: var(--color-surface-glass);
    border: var(--glass-border);
    border-radius: var(--radius-md);
    padding: var(--padding);
    backdrop-filter: blur(var(--glass-blur));
    transition: all 0.3s cubic-bezier(0.16, 1, 0.3, 1);
    position: relative;
    overflow: hidden;
  }
  
  /* Subtle highlight on top edge */
  .card::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 1px;
    background: linear-gradient(90deg, 
      transparent 0%, 
      rgba(255, 255, 255, 0.1) 20%, 
      rgba(255, 255, 255, 0.1) 80%, 
      transparent 100%
    );
    opacity: 0.5;
  }

  .hoverable {
    cursor: pointer;
  }

  .hoverable:hover {
    transform: translateY(-2px);
    background: hsla(var(--surface-hue), var(--surface-sat), 16%, 0.5);
    border-color: var(--color-border-hover);
    box-shadow: 0 10px 30px -10px rgba(0, 0, 0, 0.3);
  }
  
  .hoverable:active {
    transform: translateY(0);
  }
</style>
