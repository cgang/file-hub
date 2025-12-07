package api

import (
	"fmt"
	"net/http"

	"github.com/cgang/file-hub/pkg/web/auth"
	"github.com/gin-gonic/gin"
)

func Register(r *gin.RouterGroup) {
	r.GET("/status", auth.CheckStatus)
	r.GET("/roots", auth.Roots)

	r.POST("/setup", auth.Setup)
	r.POST("/login", auth.Login)
	r.POST("/logout", auth.Logout)

	r.Use(auth.Authenticate)
	r.GET("/hello", Hello)
}

func Hello(c *gin.Context) {
	user, ok := auth.GetSessionUser(c)
	if !ok {
		c.String(http.StatusInternalServerError, "Unable to get user from session")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Hello, " + user.Username,
		"dav":     fmt.Sprintf("/dav/%s", user.Username),
	})
}
