<script>
import '../styles/setup.css';

let username = '';
let email = '';
let password = '';
let confirmPassword = '';
let root_dir = '';
let errorMessage = '';
let successMessage = '';
let isLoading = false;

async function handleSubmit() {
  errorMessage = '';
  successMessage = '';

  // Basic validation
  if (!username || !email || !password || !confirmPassword) {
    errorMessage = 'All fields are required';
    return;
  }

  if (password !== confirmPassword) {
    errorMessage = 'Passwords do not match';
    return;
  }

  if (password.length < 6) {
    errorMessage = 'Password must be at least 6 characters';
    return;
  }

  isLoading = true;

  try {
    const response = await fetch('/api/setup', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        username,
        email,
        password,
        root_dir
      })
    });

    const data = await response.json();

    if (response.ok) {
      successMessage = data.message;
      // Redirect to login after a short delay
      setTimeout(() => {
        window.location.href = '/ui/';
      }, 2000);
    } else {
      errorMessage = data.error || 'Failed to create admin user';
    }
  } catch (error) {
    errorMessage = 'Network error. Please try again.';
  } finally {
    isLoading = false;
  }
}
</script>

<div class="setup-container">
  <div class="setup-form">
    <h2>Welcome to FileHub</h2>
    <p>Create your admin account to get started</p>
    
    {#if errorMessage}
      <div class="alert alert-error">{errorMessage}</div>
    {/if}
    
    {#if successMessage}
      <div class="alert alert-success">{successMessage}</div>
    {/if}
    
    <form on:submit|preventDefault={handleSubmit}>
      <div class="form-group">
        <label for="username">Username</label>
        <input 
          type="text" 
          id="username" 
          bind:value={username} 
          placeholder="Enter your username"
          required
        />
      </div>
      
      <div class="form-group">
        <label for="email">Email</label>
        <input 
          type="email" 
          id="email" 
          bind:value={email} 
          placeholder="Enter your email"
          required
        />
      </div>
      
      <div class="form-group">
        <label for="password">Password</label>
        <input 
          type="password" 
          id="password" 
          bind:value={password} 
          placeholder="Enter your password"
          required
        />
      </div>
      
      <div class="form-group">
        <label for="confirmPassword">Confirm Password</label>
        <input 
          type="password" 
          id="confirmPassword" 
          bind:value={confirmPassword} 
          placeholder="Confirm your password"
          required
        />
      </div>
      
      <div class="form-group">
        <label for="rootDir">Root Directory</label>
        <input 
          type="text" 
          id="rootDir" 
          bind:value={root_dir} 
          placeholder="/path/to/root"
          required
        />
      </div>
      
      <button 
        type="submit" 
        class="btn btn-primary"
        disabled={isLoading}
      >
        {#if isLoading}
          Creating Account...
        {:else}
          Create Admin Account
        {/if}
      </button>
    </form>
  </div>
</div>
