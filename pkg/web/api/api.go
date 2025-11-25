package api

import (
	"net/http"

	"github.com/cgang/file-hub/pkg/web/internal/auth"
	"github.com/gin-gonic/gin"
)

func Register(r *gin.RouterGroup) {
	r.Use(auth.Authenticate)

	r.GET("/hello", Hello)
}

func Hello(c *gin.Context) {
	user, _ := auth.GetAuthenticatedUser(c)
	c.JSON(http.StatusOK, gin.H{
		"message": "Hello, " + user.Username,
		"user":    user,
	})
}
