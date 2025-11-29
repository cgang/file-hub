package webdav

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cgang/file-hub/pkg/model"
	"github.com/cgang/file-hub/pkg/stor"
	"github.com/cgang/file-hub/pkg/users"
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

	v1.PUT("/:user/*path", handlePut)
	v1.DELETE("/:user/*path", handleDelete)
	v1.GET("/:user/*path", handleGet)

	v1.Handle("PROPFIND", "/:user/*path", handlePropfind)
	v1.Handle("MKCOL", "/:user/*path", handleMkcol)
	v1.Handle("COPY", "/:user/*path", handleCopyMove)
	v1.Handle("MOVE", "/:user/*path", handleCopyMove)
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

func getUserStorage(c *gin.Context) (stor.Storage, error) {
	name := c.Param("user")
	user, err := users.GetByUsername(c, name)
	if err != nil {
		sendError(c, http.StatusBadRequest, "Path not found")
		return nil, fmt.Errorf("get user %s failed: %w", name, err)
	}

	userStorage, err := stor.ForUser(c, user)
	if err != nil {
		sendError(c, http.StatusBadRequest, "Path not found")
		return nil, fmt.Errorf("get storage for user %s failed: %w", name, err)
	}

	return userStorage, nil
}

// handlePropfind handles PROPFIND requests
func handlePropfind(c *gin.Context) {
	// Get authenticated user
	user, err := getAuthenticatedUser(c)
	if err != nil {
		sendError(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	storage, err := getUserStorage(c)
	if err != nil {
		return
	}

	name := c.Param("path")
	// Log request
	log.Printf("Handling PROPFIND request for %s", name)

	if err := storage.CheckPermission(c, name, user, stor.PermissionRead); err != nil {
		log.Printf("Permission denied for %s: %v", name, err)
		sendError(c, http.StatusForbidden, "Permission denied")
		return
	}

	// Get file info using storage abstraction
	file, err := storage.GetFileInfo(c, name)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("File not found: %s", name)
			sendError(c, http.StatusNotFound, "File not found")
			return
		}
		log.Printf("Error accessing file %s: %v", name, err)
		sendError(c, http.StatusInternalServerError, "Error accessing file: %v", err)
		return
	}

	// Build response
	var ms Multistatus

	// Add the file/directory itself
	ms.Response = append(ms.Response, createResponse(c.Request.URL.Path, file))

	// If it's a directory, list its contents
	if file.IsDir {
		files, err := storage.ListDir(c, name)
		if err != nil {
			log.Printf("Error reading directory %s: %v", name, err)
			sendError(c, http.StatusInternalServerError, "Failed to read directory: %v", err)
			return
		}

		// Add each entry to the response
		for _, entry := range files {
			entryUrlPath := strings.TrimSuffix(c.Request.URL.Path, "/") + "/" + entry.Name
			ms.Response = append(ms.Response, createResponse(entryUrlPath, entry))
		}
	}

	// Log response
	log.Printf("Sending PROPFIND response for %s with %d items", name, len(ms.Response))

	// Set XML content type and send response
	c.XML(http.StatusOK, ms)
}

// createResponse creates a WebDAV response with proper properties
func createResponse(href string, file *stor.FileObject) Response {
	prop := Prop{
		Name:         file.Name,
		DisplayName:  file.Name,
		LastModified: file.LastModified.UTC().Format(time.RFC1123),
	}

	if file.IsDir {
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

	storage, err := getUserStorage(c)
	if err != nil {
		return
	}

	name := c.Param("path")

	if err := storage.CheckPermission(c, name, user, stor.PermissionWrite); err != nil {
		log.Printf("Permission denied for %s: %v", name, err)
		sendError(c, http.StatusForbidden, "Permission denied")
		return
	}

	// Write file using storage abstraction
	if err := storage.WriteToFile(c, name, c.Request.Body); err != nil {
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

	storage, err := getUserStorage(c)
	if err != nil {
		return
	}

	name := c.Param("path")
	if err := storage.CheckPermission(c, name, user, stor.PermissionDelete); err != nil {
		log.Printf("Permission denied for %s: %v", name, err)
		sendError(c, http.StatusForbidden, "Permission denied")
		return
	}

	if err := storage.DeleteFile(c, name); err != nil {
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

	storage, err := getUserStorage(c)
	if err != nil {
		return
	}

	name := c.Param("path")
	if err := storage.CheckPermission(c, name, user, stor.PermissionWrite); err != nil {
		log.Printf("Permission denied for %s: %v", name, err)
		sendError(c, http.StatusForbidden, "Permission denied")
		return
	}

	if err := storage.CreateDir(c, name); err != nil {
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

	storage, err := getUserStorage(c)
	if err != nil {
		return
	}

	srcPath := c.Param("path")
	destination := c.Request.Header.Get("Destination")
	if destination == "" {
		sendError(c, http.StatusBadRequest, "Destination header required")
		return
	}

	// Parse destination path
	destPath := strings.TrimPrefix(destination, "/")

	if err := storage.CheckPermission(c, srcPath, user, stor.PermissionRead); err != nil {
		log.Printf("Permission denied for source %s: %v", srcPath, err)
		sendError(c, http.StatusForbidden, "Permission denied for source")
		return
	}

	// TODO check permission on destination parent directory

	// Create destination directory if needed
	if err := storage.CreateDir(c, filepath.Dir(destPath)); err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to create destination directory: %v", err)
		return
	}

	// Handle COPY or MOVE
	if c.Request.Method == "COPY" {
		// Copy file/directory using storage
		if err := storage.CopyFile(c, srcPath, destPath); err != nil {
			sendError(c, http.StatusInternalServerError, "Failed to copy file: %v", err)
			return
		}
	} else {
		// Move file/directory using storage
		if err := storage.MoveFile(c, srcPath, destPath); err != nil {
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

	storage, err := getUserStorage(c)
	if err != nil {
		return
	}

	name := c.Param("path")

	if err := storage.CheckPermission(c, name, user, stor.PermissionRead); err != nil {
		log.Printf("Permission denied for %s: %v", name, err)
		sendError(c, http.StatusForbidden, "Permission denied")
		return
	}

	info, err := storage.GetFileInfo(c, name)
	if err != nil {
		if os.IsNotExist(err) {
			sendError(c, http.StatusNotFound, "File not found")
			return
		}
		sendError(c, http.StatusInternalServerError, "Error accessing file: %v", err)
		return
	}

	if info.IsDir {
		sendError(c, http.StatusBadRequest, "Cannot GET a directory")
		return
	}

	c.Header("Content-Type", info.ContentType)
	c.Header("Content-Length", fmt.Sprintf("%d", info.Size))

	file, err := storage.OpenFile(c, name)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "Error opening file: %v", err)
		return
	}
	defer file.Close()

	if _, err := io.Copy(c.Writer, file); err != nil {
		log.Printf("Failed to copy file content: %s", err)
	}
}
