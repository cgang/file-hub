package api

import (
	"net/http"

	"github.com/cgang/file-hub/pkg/web/internal/auth"
	"github.com/gin-gonic/gin"
)

func Register(r *gin.RouterGroup) {
	// Public routes (no authentication required)
	public := r.Group("/")
	{
		public.POST("/login", auth.LoginHandler)
		public.POST("/logout", auth.LogoutHandler)
	}

	// Protected routes (authentication required)
	protected := r.Group("/")
	protected.Use(auth.Authenticate)
	{
		protected.GET("/hello", Hello)
	}
}

func Hello(c *gin.Context) {
	user, _ := auth.GetAuthenticatedUser(c)
	c.JSON(http.StatusOK, gin.H{
		"message": "Hello, " + user.Username,
		"user":    user,
	})
}
