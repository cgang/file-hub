<script>
  let username = '';
  let password = '';
  let error = '';
  let loading = false;

  const handleSubmit = async (e) => {
    e.preventDefault();
    loading = true;
    error = '';

    try {
      const res = await fetch('/api/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password })
      });

      if (!res.ok) {
        error = 'Invalid credentials';
        return;
      }

      // Force page reload to trigger App.svelte's status check
      window.location.reload();
    } catch (err) {
      error = 'Login failed';
      console.error(err);
    } finally {
      loading = false;
    }
  };
</script>

<h2>Login</h2>

<div class="login-container">
  {#if error}
    <div class="error">{error}</div>
  {/if}
  <div class="form-wrapper">
    <form on:submit={handleSubmit}>
    <div class="form-group">
      <label for="username">Username</label>
      <input type="text" bind:value={username} required />
    </div>
    
    <div class="form-group">
      <label for="password">Password</label>
      <input type="password" bind:value={password} required />
    </div>
    
    <button type="submit" disabled={loading}>
      {loading ? 'Logging in...' : 'Login'}
    </button>
  </form>
  </div>
</div>
