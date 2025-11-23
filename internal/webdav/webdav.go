package webdav

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/cgang/file-hub/internal/stor"
	"github.com/gin-gonic/gin"
)

// WebDAVServer represents the WebDAV server
type WebDAVServer struct {
	storage stor.Storage
}

// New creates a new WebDAV server
func New(storage stor.Storage) *WebDAVServer {
	server := &WebDAVServer{
		storage: storage,
	}

	return server
}

func setDavHeaders(c *gin.Context) {
	c.Header("DAV", "1")
	c.Header("Content-Type", "application/xml; charset=utf-8")
}

// SetupRoutes configures the WebDAV routes
func (s *WebDAVServer) SetupRoutes(v1 *gin.RouterGroup) {
	v1.Use(setDavHeaders)

	v1.PUT("/*path", s.handlePut)
	v1.DELETE("/*path", s.handleDelete)
	v1.GET("/*path", s.handleGet)

	v1.Handle("PROPFIND", "/*path", s.handlePropfind)
	v1.Handle("MKCOL", "/*path", s.handleMkcol)
	v1.Handle("COPY", "/*path", s.handleCopyMove)
	v1.Handle("MOVE", "/*path", s.handleCopyMove)
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
func sendError(c *gin.Context, status int, message string) {
	c.XML(status, &ErrorBody{
		XMLName: xml.Name{Space: "DAV", Local: "error"},
		Message: message,
	})
}

// handlePropfind handles PROPFIND requests
func (s *WebDAVServer) handlePropfind(c *gin.Context) {
	name := c.Param("path")
	// Log request
	log.Printf("Handling PROPFIND request for %s", name)

	// Get file info using storage abstraction
	file, err := s.storage.GetFileInfo(name)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("File not found: %s", name)
			sendError(c, http.StatusNotFound, "File not found")
			return
		}
		log.Printf("Error accessing file %s: %v", name, err)
		sendError(c, http.StatusInternalServerError, fmt.Sprintf("Error accessing file: %v", err))
		return
	}

	// Build response
	var ms Multistatus

	// Add the file/directory itself
	ms.Response = append(ms.Response, createResponse(c.Request.URL.Path, file))

	// If it's a directory, list its contents
	if file.IsDir {
		files, err := s.storage.ListDir(name)
		if err != nil {
			log.Printf("Error reading directory %s: %v", name, err)
			sendError(c, http.StatusInternalServerError, fmt.Sprintf("Failed to read directory: %v", err))
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
func createResponse(href string, file *stor.File) Response {
	prop := Prop{
		Name:         file.Name,
		DisplayName:  file.Name,
		IsCollection: "1",
		ContentType:  "httpd/unix-directory",
		Length:       fmt.Sprintf("%d", file.Size),
		LastModified: file.LastModified.UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT"),
	}

	if !file.IsDir {
		prop.IsCollection = "0"
		prop.ContentType = file.ContentType
	}

	return Response{
		Href:   href,
		Prop:   prop,
		Status: "HTTP/1.1 200 OK",
	}
}

// handlePut handles PUT requests
func (s *WebDAVServer) handlePut(c *gin.Context) {
	name := c.Param("path")
	// Write file using storage abstraction
	if err := s.storage.WriteToFile(name, c.Request.Body); err != nil {
		sendError(c, http.StatusInternalServerError, fmt.Sprintf("Failed to write file: %v", err))
		return
	}

	c.Status(http.StatusCreated)
}

// handleDelete handles DELETE requests
func (s *WebDAVServer) handleDelete(c *gin.Context) {
	name := c.Param("path")
	if err := s.storage.DeleteFile(name); err != nil {
		sendError(c, http.StatusInternalServerError, fmt.Sprintf("Failed to delete file: %v", err))
		return
	}
	c.Status(http.StatusNoContent)
}

// handleMkcol handles MKCOL requests
func (s *WebDAVServer) handleMkcol(c *gin.Context) {
	name := c.Param("path")
	if err := s.storage.CreateDir(name); err != nil {
		sendError(c, http.StatusInternalServerError, fmt.Sprintf("Failed to create directory: %v", err))
		return
	}
	c.Status(http.StatusCreated)
}

// handleCopyMove handles COPY and MOVE requests
func (s *WebDAVServer) handleCopyMove(c *gin.Context) {
	srcPath := c.Param("path")
	destination := c.Request.Header.Get("Destination")
	if destination == "" {
		sendError(c, http.StatusBadRequest, "Destination header required")
		return
	}

	// Parse destination path
	destPath := strings.TrimPrefix(destination, "/")

	// Create destination directory if needed
	if err := s.storage.CreateDir(filepath.Dir(destPath)); err != nil {
		sendError(c, http.StatusInternalServerError, fmt.Sprintf("Failed to create destination directory: %v", err))
		return
	}

	// Handle COPY or MOVE
	if c.Request.Method == "COPY" {
		// Copy file/directory using storage
		if err := s.storage.CopyFile(srcPath, destPath); err != nil {
			sendError(c, http.StatusInternalServerError, fmt.Sprintf("Failed to copy file: %v", err))
			return
		}
	} else {
		// Move file/directory using storage
		if err := s.storage.MoveFile(srcPath, destPath); err != nil {
			sendError(c, http.StatusInternalServerError, fmt.Sprintf("Failed to move file: %v", err))
			return
		}
	}

	c.Status(http.StatusCreated)
}

// handleGet handles GET requests
func (s *WebDAVServer) handleGet(c *gin.Context) {
	name := c.Param("path")
	// Check if file exists
	exists, err := s.storage.Exists(name)
	if err != nil {
		sendError(c, http.StatusInternalServerError, fmt.Sprintf("Error accessing file: %v", err))
		return
	}
	if !exists {
		sendError(c, http.StatusNotFound, "File not found")
		return
	}

	http.ServeFile(c.Writer, c.Request, name)
}
