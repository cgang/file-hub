package auth

import (
	"net/http"

	"github.com/cgang/file-hub/pkg/stor"
	"github.com/cgang/file-hub/pkg/users"
	"github.com/gin-gonic/gin"
)

func Roots(c *gin.Context) {
	c.JSON(http.StatusOK, availRoots)
}

type SetupRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	RootDir  string `json:"root_dir" binding:"required"`
}

// Setup handles the creation of the first user
func Setup(c *gin.Context) {
	// Check if database is empty, if not reject the request
	if ok, err := users.HasAnyUser(c); err != nil || ok {
		c.String(http.StatusBadRequest, "Setup already completed")
		return
	}

	var req SetupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.String(http.StatusBadRequest, "Invalid request format")
		return
	}

	if !stor.ValidRoot(req.RootDir) {
		c.String(http.StatusBadRequest, "Invalid root dir: %s", req.RootDir)
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

	if err := stor.CreateHomeRepo(c, user, req.RootDir); err != nil {
		c.String(http.StatusInternalServerError, "Failed to create home repository for %s: %s", req.Username, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Setup completed successfully. You can now login.",
		"user":    user,
	})
}
