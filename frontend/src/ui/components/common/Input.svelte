<script lang="ts">
  export let value: string | number;
  export let type = 'text';
  export let placeholder = '';
  export let label = '';
  export let id = '';
  export let required = false;
  export let disabled = false;
  export let error: string | null = null;
  export let min: string | number | undefined = undefined;
  export let step: string | number | undefined = undefined;
</script>

<div class="input-group">
  {#if label}
    <label for={id}>
      {label}
      {#if required}<span class="required">*</span>{/if}
    </label>
  {/if}
  
  <div class="input-wrapper" class:has-error={!!error} class:disabled>
    <slot name="prefix" />
    <input
      {id}
      {type}
      bind:value
      {placeholder}
      {required}
      {disabled}
      {min}
      {step}
      on:input
      on:change
      on:blur
      on:focus
    />
    <slot name="suffix" />
  </div>

  {#if error}
    <span class="error-msg">{error}</span>
  {/if}
</div>

<style>
  .input-group {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    width: 100%;
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

  .input-wrapper {
    display: flex;
    align-items: center;
    background: var(--color-surface-glass);
    border: var(--glass-border);
    border-radius: var(--radius-sm);
    transition: all 0.2s;
    backdrop-filter: blur(12px);
  }

  .input-wrapper:focus-within {
    border-color: var(--color-primary);
    box-shadow: 0 0 0 2px var(--color-primary-glow);
    background: hsla(var(--surface-hue), var(--surface-sat), 15%, 0.6);
  }
  
  .input-wrapper.has-error {
    border-color: var(--color-danger);
  }
  
  .input-wrapper.has-error:focus-within {
     box-shadow: 0 0 0 2px rgba(239, 68, 68, 0.25);
  }

  .input-wrapper.disabled {
    opacity: 0.6;
    pointer-events: none;
  }

  input {
    flex: 1;
    background: transparent;
    border: none;
    padding: 0.6rem 0.8rem;
    color: var(--text-main);
    font-family: inherit;
    font-size: 0.95rem;
    width: 100%;
  }

  input:focus {
    outline: none;
  }

  input::placeholder {
    color: rgba(255, 255, 255, 0.3);
  }

  /* Remove number arrows */
  input[type=number]::-webkit-inner-spin-button, 
  input[type=number]::-webkit-outer-spin-button { 
    -webkit-appearance: none; 
    margin: 0; 
  }

  .error-msg {
    font-size: 0.8rem;
    color: var(--color-danger);
    margin-left: 2px;
  }
</style>
