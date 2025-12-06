package api

import (
	"net/http"

	"github.com/cgang/file-hub/pkg/users"
	"github.com/cgang/file-hub/pkg/web/auth"
	"github.com/gin-gonic/gin"
)

func Register(r *gin.RouterGroup) {
	// Public routes (no authentication required)
	public := r.Group("/")
	{
		public.POST("/login", auth.LoginHandler)
		public.POST("/logout", auth.LogoutHandler)
		public.POST("/setup", SetupHandler)
	}

	// Protected routes (authentication required)
	protected := r.Group("/")
	protected.Use(auth.Authenticate)
	{
		protected.GET("/hello", Hello)
	}
}

// SetupHandler handles the creation of the first user
func SetupHandler(c *gin.Context) {
	// Check if database is empty, if not reject the request
	if ok, err := users.HasAnyUser(c.Request.Context()); err != nil || ok {
		c.String(http.StatusBadRequest, "Setup already completed")
		return
	}

	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.String(http.StatusBadRequest, "Invalid request format")
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
		c.String(http.StatusInternalServerError, "Failed to create user: "+err.Error())
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
