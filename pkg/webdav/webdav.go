package webdav

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/cgang/file-hub/pkg/db"
	"github.com/cgang/file-hub/pkg/model"
	"github.com/cgang/file-hub/pkg/stor"
	"github.com/gin-gonic/gin"
)

func setDavHeaders(c *gin.Context) {
	c.Header("DAV", "1")
	c.Header("Content-Type", "application/xml; charset=utf-8")
}

// getAuthenticatedUser retrieves the authenticated user from the context
func getAuthenticatedUser(c *gin.Context) (*model.User, error) {
	userValue, exists := c.Get("user")
	if !exists {
		return nil, fmt.Errorf("no authenticated user found")
	}

	user, ok := userValue.(*model.User)
	if !ok {
		return nil, fmt.Errorf("invalid user type in context")
	}

	return user, nil
}

// Register configures the WebDAV routes
func Register(v1 *gin.RouterGroup) {
	// Note: Authentication should be handled by a middleware in the calling code
	v1.Use(setDavHeaders)

	v1.PUT("/:repos/*path", handlePut)
	v1.DELETE("/:repos/*path", handleDelete)
	v1.GET("/:repos/*path", handleGet)

	v1.Handle("PROPFIND", "/:repos/*path", handlePropfind)
	v1.Handle("MKCOL", "/:repos/*path", handleMkcol)
	v1.Handle("COPY", "/:repos/*path", handleCopyMove)
	v1.Handle("MOVE", "/:repos/*path", handleCopyMove)
}

// XML structures for WebDAV
type Propfind struct {
	XMLName xml.Name `xml:"DAV: propfind"`
	Prop    Prop     `xml:"prop"`
}

type Prop struct {
	Name         string `xml:"name"`
	DisplayName  string `xml:"displayname"`
	IsCollection string `xml:"iscollection"`
	ContentType  string `xml:"getcontenttype"`
	Length       string `xml:"getcontentlength"`
	LastModified string `xml:"getlastmodified"`
}

type Multistatus struct {
	XMLName  xml.Name   `xml:"multistatus"`
	Response []Response `xml:"response"`
}

type Response struct {
	Href   string `xml:"href"`
	Prop   Prop   `xml:"propstat>prop"`
	Status string `xml:"status"`
}

// ErrorBody is used for WebDAV error responses
type ErrorBody struct {
	XMLName xml.Name `xml:"error"`
	Message string   `xml:",innerxml"`
}

// sendError sends a standardized WebDAV error response
func sendError(c *gin.Context, status int, format string, a ...any) {
	c.XML(status, &ErrorBody{
		XMLName: xml.Name{Space: "DAV", Local: "error"},
		Message: fmt.Sprintf(format, a...),
	})
}

func getResource(c *gin.Context) (*model.Resource, error) {
	name := c.Param("repos")
	repos, err := stor.GetRepository(c, name)
	if err != nil {
		sendError(c, http.StatusBadRequest, "Repository not found")
		return nil, fmt.Errorf("get repository %s failed: %w", name, err)
	}

	return &model.Resource{
		ReposID: repos.ID,
		OwnerID: repos.OwnerID,
		Path:    strings.TrimPrefix(c.Param("path"), "/"),
	}, nil
}

// handlePropfind handles PROPFIND requests
func handlePropfind(c *gin.Context) {
	// Get authenticated user
	user, err := getAuthenticatedUser(c)
	if err != nil {
		sendError(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	resource, err := getResource(c)
	if err != nil {
		return
	}

	// Log request
	log.Printf("Handling PROPFIND request for %s", resource)

	if err := stor.CheckPermission(c, user.ID, resource, stor.PermissionRead); err != nil {
		log.Printf("Permission denied for %s: %v", resource, err)
		sendError(c, http.StatusForbidden, "Permission denied")
		return
	}

	// Build response
	var ms Multistatus

	// Get file info using storage abstraction
	file, err := stor.GetFileInfo(c, resource)
	if err != nil {
		if errors.Is(err, db.ErrFileNotFound) {
			sendError(c, http.StatusNotFound, "File not found")
			return
		}

		log.Printf("Error accessing file %s: %v", resource, err)
		sendError(c, http.StatusInternalServerError, "Error accessing file: %v", err)
		return
	}

	// Add the file/directory itself
	ms.Response = append(ms.Response, createResponse(c.Request.URL.Path, file))
	// If it's a directory, list its contents
	files, err := stor.ListDir(c, file)
	if err != nil {
		log.Printf("Error reading directory %s: %v", resource, err)
		sendError(c, http.StatusInternalServerError, "Failed to read directory: %v", err)
		return
	}

	for _, entry := range files {
		entryUrlPath := strings.TrimSuffix(c.Request.URL.Path, "/") + "/" + entry.Path
		ms.Response = append(ms.Response, createResponse(entryUrlPath, entry))
	}

	c.XML(http.StatusOK, &ms)
}

// createResponse creates a WebDAV response with proper properties
func createResponse(href string, file *model.FileObject) Response {
	name := path.Base(file.Path)
	prop := Prop{
		Name:         name,
		DisplayName:  name,
		LastModified: file.UpdatedAt.UTC().Format(time.RFC1123),
	}

	if file.Directory {
		prop.IsCollection = "1"
		prop.ContentType = "httpd/unix-directory"
	} else {
		prop.IsCollection = "0"
		prop.ContentType = file.ContentType
		prop.Length = fmt.Sprintf("%d", file.Size)
	}

	return Response{
		Href:   href,
		Prop:   prop,
		Status: "HTTP/1.1 200 OK",
	}
}

// handlePut handles PUT requests
func handlePut(c *gin.Context) {
	// Get authenticated user
	user, err := getAuthenticatedUser(c)
	if err != nil {
		sendError(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	resource, err := getResource(c)
	if err != nil {
		return
	}

	if err := stor.CheckPermission(c, user.ID, resource, stor.PermissionWrite); err != nil {
		sendError(c, http.StatusForbidden, "Permission denied")
		return
	}

	// Write file using storage abstraction
	if err := stor.WriteToFile(c, resource, c.Request.Body); err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to write file: %v", err)
		return
	}

	c.Status(http.StatusCreated)
}

// handleDelete handles DELETE requests
func handleDelete(c *gin.Context) {
	// Get authenticated user
	user, err := getAuthenticatedUser(c)
	if err != nil {
		sendError(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	resource, err := getResource(c)
	if err != nil {
		return
	}

	if err := stor.CheckPermission(c, user.ID, resource, stor.PermissionDelete); err != nil {
		log.Printf("Permission denied for %s: %v", resource, err)
		sendError(c, http.StatusForbidden, "Permission denied")
		return
	}

	if err := stor.DeleteFile(c, resource); err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to delete file: %v", err)
		return
	}
	c.Status(http.StatusNoContent)
}

// handleMkcol handles MKCOL requests
func handleMkcol(c *gin.Context) {
	// Get authenticated user
	user, err := getAuthenticatedUser(c)
	if err != nil {
		sendError(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	resource, err := getResource(c)
	if err != nil {
		return
	}

	if err := stor.CheckPermission(c, user.ID, resource, stor.PermissionWrite); err != nil {
		log.Printf("Permission denied for %s: %v", resource, err)
		sendError(c, http.StatusForbidden, "Permission denied")
		return
	}

	if err := stor.CreateDir(c, resource); err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to create directory: %v", err)
		return
	}
	c.Status(http.StatusCreated)
}

// handleCopyMove handles COPY and MOVE requests
func handleCopyMove(c *gin.Context) {
	// Get authenticated user
	user, err := getAuthenticatedUser(c)
	if err != nil {
		sendError(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	resource, err := getResource(c)
	if err != nil {
		return
	}

	destination := c.Request.Header.Get("Destination")
	if destination == "" || !strings.HasPrefix(destination, "/dav") { // TODO fix hardcoded path
		sendError(c, http.StatusBadRequest, "Destination header required")
		return
	}

	// Parse destination path
	var destResource *model.Resource
	destPath := strings.TrimPrefix(destination, "/dav")
	if repos, name, ok := strings.Cut(destPath, "/"); ok {
		r, err := stor.GetRepository(c, repos)
		if err != nil {
			sendError(c, http.StatusBadRequest, "Destination repository not found")
			return
		}

		destResource = &model.Resource{
			ReposID: r.ID, // TODO lookup repository ID
			OwnerID: user.ID,
			Path:    name,
		}
	}

	if err := stor.CheckPermission(c, user.ID, resource, stor.PermissionRead); err != nil {
		sendError(c, http.StatusForbidden, "Permission denied for source")
		return
	}

	if err := stor.CheckPermission(c, user.ID, destResource, stor.PermissionWrite); err != nil {
		sendError(c, http.StatusForbidden, "Permission denied for destination")
		return
	}

	// Handle COPY or MOVE
	if c.Request.Method == "COPY" {
		// Copy file/directory using storage
		if err := stor.CopyFile(c, resource, destResource); err != nil {
			sendError(c, http.StatusInternalServerError, "Failed to copy file: %v", err)
			return
		}
	} else {
		// Move file/directory using storage
		if err := stor.MoveFile(c, resource, destResource); err != nil {
			sendError(c, http.StatusInternalServerError, "Failed to move file: %v", err)
			return
		}
	}

	c.Status(http.StatusCreated)
}

// handleGet handles GET requests
func handleGet(c *gin.Context) {
	// Get authenticated user
	user, err := getAuthenticatedUser(c)
	if err != nil {
		sendError(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	resource, err := getResource(c)
	if err != nil {
		return
	}

	if err := stor.CheckPermission(c, user.ID, resource, stor.PermissionRead); err != nil {
		log.Printf("Permission denied for %s: %v", resource, err)
		sendError(c, http.StatusForbidden, "Permission denied")
		return
	}

	info, err := stor.GetFileInfo(c, resource)
	if err != nil {
		if os.IsNotExist(err) {
			sendError(c, http.StatusNotFound, "File not found")
			return
		}
		sendError(c, http.StatusInternalServerError, "Error accessing file: %v", err)
		return
	}

	if info.Directory {
		sendError(c, http.StatusBadRequest, "Cannot GET a directory")
		return
	}

	c.Header("Content-Type", info.ContentType)
	c.Header("Content-Length", fmt.Sprintf("%d", info.Size))

	file, err := stor.OpenFile(c, resource)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "Error opening file: %v", err)
		return
	}
	defer file.Close()

	if _, err := io.Copy(c.Writer, file); err != nil {
		log.Printf("Failed to copy file content: %s", err)
	}
}
