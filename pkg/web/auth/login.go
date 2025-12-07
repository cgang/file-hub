package auth

import (
	"net/http"

	"github.com/cgang/file-hub/pkg/users"
	"github.com/gin-gonic/gin"
)

type CheckStatusResponse struct {
	Username string `json:"username,omitempty"`
	Setup    bool   `json:"setup,omitempty"`
}

func CheckStatus(c *gin.Context) {
	resp := &CheckStatusResponse{}
	if user, ok := GetSessionUser(c); ok {
		resp.Username = user.Username
		c.JSON(http.StatusOK, resp)
		return
	}

	if ok, err := users.HasAnyUser(c); err == nil {
		if !ok {
			resp.Setup = true
		}
		c.JSON(http.StatusOK, resp)
	} else {
		c.String(http.StatusInternalServerError, "Failed to check configuration: %s", err)
	}
}

// Login handles user login requests
func Login(c *gin.Context) {
	// Check if database is empty, redirect to setup page if it is
	if yes, err := users.HasAnyUser(c); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check user existence"})
		return
	} else if !yes {
		c.JSON(http.StatusFound, gin.H{"redirect": "/setup"})
		return
	}

	var req struct {
		Username string `json:"username" form:"username"`
		Password string `json:"password" form:"password"`
	}

	if err := c.Bind(&req); err != nil {
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
	})
}

// Logout handles user logout requests
func Logout(c *gin.Context) {
	DestroySession(c)
	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}
