package api

import (
	"fmt"
	"net/http"

	"github.com/cgang/file-hub/pkg/stor"
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
	r.POST("/scan_files", ScanFiles)
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

func ScanFiles(c *gin.Context) {
	user, ok := auth.GetSessionUser(c)
	if !ok {
		c.String(http.StatusInternalServerError, "Unable to get user from session")
		return
	}

	repo, err := stor.GetHomeRepo(c, user) // TODO check all repositories belong to current user
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to get home repository: %s", err)
		return
	}

	if err := stor.ScanFiles(c, repo); err != nil {
		c.String(http.StatusInternalServerError, "Failed to sync files: %s", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Files synced for %s successfully", repo.Name),
	})
}
