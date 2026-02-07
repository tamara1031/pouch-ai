<script lang="ts">
  import { Wallet } from 'lucide-svelte';
</script>

<nav class="navbar">
  <div class="container nav-content">
    <div class="brand">
      <div class="logo">
        <Wallet size={20} />
      </div>
      <h1>Pouch AI</h1>
    </div>
    <div class="nav-actions">
      <!-- Add more nav items here if needed -->
    </div>
  </div>
</nav>

<style>
  .navbar {
    height: 64px;
    border-bottom: var(--glass-border);
    background: rgba(10, 10, 10, 0.6);
    backdrop-filter: blur(var(--glass-blur));
    position: sticky;
    top: 0;
    z-index: 50;
  }

  .container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 1.5rem;
  }

  .nav-content {
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .brand {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    color: var(--text-main);
  }

  .logo {
    width: 32px;
    height: 32px;
    background: linear-gradient(135deg, var(--color-primary), #2563eb);
    border-radius: var(--radius-sm);
    display: flex;
    align-items: center;
    justify-content: center;
    color: white;
    box-shadow: 0 0 15px var(--color-primary-glow);
  }

  h1 {
    font-size: 1.1rem;
    font-weight: 700;
    letter-spacing: -0.01em;
  }
</style>
