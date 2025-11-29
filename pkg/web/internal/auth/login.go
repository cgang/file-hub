package auth

import (
	"net/http"

	"github.com/cgang/file-hub/pkg/users"
	"github.com/gin-gonic/gin"
)

// LoginHandler handles user login requests
func LoginHandler(c *gin.Context) {
	// Check if database is empty, redirect to setup page if it is
	if yes, err := users.HasAnyUser(c); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check user existence"})
		return
	} else if !yes {
		c.JSON(http.StatusFound, gin.H{"redirect": "/setup"})
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	user, err := users.Authenticate(c, req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// Create a session for the user
	if err := CreateSession(c, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user":    user,
	})
}

// LogoutHandler handles user logout requests
func LogoutHandler(c *gin.Context) {
	DestroySession(c)
	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}
