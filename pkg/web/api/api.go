package api

import (
	"net/http"

	"github.com/cgang/file-hub/pkg/users"
	"github.com/cgang/file-hub/pkg/web/internal/auth"
	"github.com/gin-gonic/gin"
)

// SetupMiddleware allows access to setup routes only when database is empty
func SetupMiddleware(c *gin.Context) {
	if ok, err := users.HasAnyUser(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check user existence"})
		c.Abort()
		return
	} else if ok {
		// Database is not empty, redirect to main UI
		c.Redirect(http.StatusFound, "/ui/")
		c.Abort()
		return
	}
	c.Next()
}

func Register(r *gin.RouterGroup) {
	// Public routes (no authentication required)
	public := r.Group("/")
	{
		public.POST("/login", auth.LoginHandler)
		public.POST("/logout", auth.LogoutHandler)
		public.POST("/setup", SetupHandler)
	}

	// Setup route - accessible only when database is empty
	setup := r.Group("/setup")
	setup.Use(SetupMiddleware)
	{
		setup.GET("", SetupPageHandler)
	}

	// Protected routes (authentication required)
	protected := r.Group("/")
	protected.Use(auth.Authenticate)
	{
		protected.GET("/hello", Hello)
	}
}

func SetupPageHTMLHandler(c *gin.Context) {
	// Check if database is empty, if not redirect to login
	if ok, err := users.HasAnyUser(c.Request.Context()); err != nil || ok {
		c.Redirect(http.StatusFound, "/ui/")
		return
	}

	// Serve the setup page HTML
	htmlContent := `
<!DOCTYPE html>
<html>
<head>
    <title>FileHub Setup</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Open Sans', 'Helvetica Neue', sans-serif;
            background-color: #f5f5f5;
            margin: 0;
            padding: 0;
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
        }
        .container {
            background: white;
            padding: 2rem;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
            width: 100%;
            max-width: 400px;
        }
        h1 {
            margin-top: 0;
            color: #333;
            text-align: center;
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
        button:hover {
            background-color: #0056b3;
        }
        .message {
            padding: 0.75rem;
            border-radius: 4px;
            margin-bottom: 1rem;
        }
        .error {
            background-color: #f8d7da;
            color: #721c24;
            border: 1px solid #f5c6cb;
        }
        .success {
            background-color: #d4edda;
            color: #155724;
            border: 1px solid #c3e6cb;
        }
        .hidden {
            display: none;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>FileHub Setup</h1>
        <p>Create your admin account to get started</p>

        <div id="error-message" class="message error hidden"></div>
        <div id="success-message" class="message success hidden"></div>

        <form id="setup-form">
            <div class="form-group">
                <label for="username">Username</label>
                <input type="text" id="username" name="username" required>
            </div>

            <div class="form-group">
                <label for="email">Email</label>
                <input type="email" id="email" name="email" required>
            </div>

            <div class="form-group">
                <label for="password">Password</label>
                <input type="password" id="password" name="password" required>
            </div>

            <div class="form-group">
                <label for="confirm-password">Confirm Password</label>
                <input type="password" id="confirm-password" name="confirm-password" required>
            </div>

            <button type="submit">Create Admin Account</button>
        </form>
    </div>

    <script>
        document.getElementById('setup-form').addEventListener('submit', async function(e) {
            e.preventDefault();

            const username = document.getElementById('username').value;
            const email = document.getElementById('email').value;
            const password = document.getElementById('password').value;
            const confirmPassword = document.getElementById('confirm-password').value;

            const errorMessage = document.getElementById('error-message');
            const successMessage = document.getElementById('success-message');

            // Hide previous messages
            errorMessage.classList.add('hidden');
            successMessage.classList.add('hidden');

            // Basic validation
            if (!username || !email || !password || !confirmPassword) {
                errorMessage.textContent = 'All fields are required';
                errorMessage.classList.remove('hidden');
                return;
            }

            if (password !== confirmPassword) {
                errorMessage.textContent = 'Passwords do not match';
                errorMessage.classList.remove('hidden');
                return;
            }

            if (password.length < 6) {
                errorMessage.textContent = 'Password must be at least 6 characters';
                errorMessage.classList.remove('hidden');
                return;
            }

            try {
                const response = await fetch('/api/setup', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({
                        username: username,
                        email: email,
                        password: password
                    })
                });

                const data = await response.json();

                if (response.ok) {
                    successMessage.textContent = data.message;
                    successMessage.classList.remove('hidden');
                    // Redirect to login after a short delay
                    setTimeout(() => {
                        window.location.href = '/ui/';
                    }, 2000);
                } else {
                    errorMessage.textContent = data.error || 'Failed to create admin user';
                    errorMessage.classList.remove('hidden');
                }
            } catch (error) {
                errorMessage.textContent = 'Network error. Please try again.';
                errorMessage.classList.remove('hidden');
            }
        });
    </script>
</body>
</html>
`

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(htmlContent))
}

func SetupPageHandler(c *gin.Context) {
	// Serve the setup page HTML
	htmlContent := `
<!DOCTYPE html>
<html>
<head>
    <title>FileHub Setup</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Open Sans', 'Helvetica Neue', sans-serif;
            background-color: #f5f5f5;
            margin: 0;
            padding: 0;
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
        }
        .container {
            background: white;
            padding: 2rem;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
            width: 100%;
            max-width: 400px;
        }
        h1 {
            margin-top: 0;
            color: #333;
            text-align: center;
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
        button:hover {
            background-color: #0056b3;
        }
        .message {
            padding: 0.75rem;
            border-radius: 4px;
            margin-bottom: 1rem;
        }
        .error {
            background-color: #f8d7da;
            color: #721c24;
            border: 1px solid #f5c6cb;
        }
        .success {
            background-color: #d4edda;
            color: #155724;
            border: 1px solid #c3e6cb;
        }
        .hidden {
            display: none;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>FileHub Setup</h1>
        <p>Create your admin account to get started</p>

        <div id="error-message" class="message error hidden"></div>
        <div id="success-message" class="message success hidden"></div>

        <form id="setup-form">
            <div class="form-group">
                <label for="username">Username</label>
                <input type="text" id="username" name="username" required>
            </div>

            <div class="form-group">
                <label for="email">Email</label>
                <input type="email" id="email" name="email" required>
            </div>

            <div class="form-group">
                <label for="password">Password</label>
                <input type="password" id="password" name="password" required>
            </div>

            <div class="form-group">
                <label for="confirm-password">Confirm Password</label>
                <input type="password" id="confirm-password" name="confirm-password" required>
            </div>

            <button type="submit">Create Admin Account</button>
        </form>
    </div>

    <script>
        document.getElementById('setup-form').addEventListener('submit', async function(e) {
            e.preventDefault();

            const username = document.getElementById('username').value;
            const email = document.getElementById('email').value;
            const password = document.getElementById('password').value;
            const confirmPassword = document.getElementById('confirm-password').value;

            const errorMessage = document.getElementById('error-message');
            const successMessage = document.getElementById('success-message');

            // Hide previous messages
            errorMessage.classList.add('hidden');
            successMessage.classList.add('hidden');

            // Basic validation
            if (!username || !email || !password || !confirmPassword) {
                errorMessage.textContent = 'All fields are required';
                errorMessage.classList.remove('hidden');
                return;
            }

            if (password !== confirmPassword) {
                errorMessage.textContent = 'Passwords do not match';
                errorMessage.classList.remove('hidden');
                return;
            }

            if (password.length < 6) {
                errorMessage.textContent = 'Password must be at least 6 characters';
                errorMessage.classList.remove('hidden');
                return;
            }

            try {
                const response = await fetch('/api/setup', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({
                        username: username,
                        email: email,
                        password: password
                    })
                });

                const data = await response.json();

                if (response.ok) {
                    successMessage.textContent = data.message;
                    successMessage.classList.remove('hidden');
                    // Redirect to login after a short delay
                    setTimeout(() => {
                        window.location.href = '/ui/';
                    }, 2000);
                } else {
                    errorMessage.textContent = data.error || 'Failed to create admin user';
                    errorMessage.classList.remove('hidden');
                }
            } catch (error) {
                errorMessage.textContent = 'Network error. Please try again.';
                errorMessage.classList.remove('hidden');
            }
        });
    </script>
</body>
</html>
`

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(htmlContent))
}

// SetupHandler handles the creation of the first user
func SetupHandler(c *gin.Context) {
	// Check if database is empty, if not reject the request
	if ok, err := users.HasAnyUser(c.Request.Context()); err != nil || ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Setup already completed"})
		return
	}

	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Create user request with admin privileges
	userReq := &users.CreateUserRequest{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		IsAdmin:  true, // First user gets admin privileges
	}

	// Save the user to the database
	user, err := users.CreateFirstUser(c, userReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Setup completed successfully. You can now login.",
		"user":    user,
	})
}

func Hello(c *gin.Context) {
	user, _ := auth.GetAuthenticatedUser(c)
	c.JSON(http.StatusOK, gin.H{
		"message": "Hello, " + user.Username,
		"user":    user,
	})
}
