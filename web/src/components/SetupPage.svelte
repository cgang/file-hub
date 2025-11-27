<script>
  import { onMount } from 'svelte';

  let username = '';
  let email = '';
  let password = '';
  let confirmPassword = '';
  let errorMessage = '';
  let successMessage = '';
  let isLoading = false;

  // Check if setup is already completed
  onMount(async () => {
    try {
      const response = await fetch('/api/setup');
      const data = await response.json();

      if (!data.setupNeeded) {
        // If setup is not needed, redirect to main app
        window.location.href = '/ui/';
      }
    } catch (error) {
      // If there's a network error, we assume setup is not needed
      console.log('Error checking setup status:', error);
      window.location.href = '/ui/';
    }
  });

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
          password
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

<style>
  .setup-container {
    display: flex;
    justify-content: center;
    align-items: center;
    min-height: 100vh;
    padding: 20px;
    background-color: #f5f5f5;
  }
  
  .setup-form {
    background: white;
    padding: 2rem;
    border-radius: 8px;
    box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
    width: 100%;
    max-width: 400px;
  }
  
  h2 {
    margin-top: 0;
    color: #333;
    text-align: center;
  }
  
  p {
    text-align: center;
    color: #666;
    margin-bottom: 2rem;
  }
  
  .form-group {
    margin-bottom: 1rem;
  }
  
  label {
    display: block;
    margin-bottom: 0.5rem;
    font-weight: 500;
    color: #333;
  }
  
  input {
    width: 100%;
    padding: 0.75rem;
    border: 1px solid #ddd;
    border-radius: 4px;
    font-size: 1rem;
    box-sizing: border-box;
  }
  
  input:focus {
    outline: none;
    border-color: #007bff;
    box-shadow: 0 0 0 2px rgba(0, 123, 255, 0.25);
  }
  
  button {
    width: 100%;
    padding: 0.75rem;
    background-color: #007bff;
    color: white;
    border: none;
    border-radius: 4px;
    font-size: 1rem;
    cursor: pointer;
    transition: background-color 0.2s;
  }
  
  button:hover:not(:disabled) {
    background-color: #0056b3;
  }
  
  button:disabled {
    background-color: #6c757d;
    cursor: not-allowed;
  }
  
  .alert {
    padding: 0.75rem;
    border-radius: 4px;
    margin-bottom: 1rem;
  }
  
  .alert-error {
    background-color: #f8d7da;
    color: #721c24;
    border: 1px solid #f5c6cb;
  }
  
  .alert-success {
    background-color: #d4edda;
    color: #155724;
    border: 1px solid #c3e6cb;
  }
</style>