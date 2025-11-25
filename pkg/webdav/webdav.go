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

	"github.com/cgang/file-hub/pkg/stor"
	"github.com/gin-gonic/gin"
)

// WebDAV provides WebDAV server functionality
type WebDAV struct {
	storage stor.Storage
}

// New creates a new WebDAV server
func New(storage stor.Storage) *WebDAV {
	return &WebDAV{
		storage: storage,
	}
}

func setDavHeaders(c *gin.Context) {
	c.Header("DAV", "1")
	c.Header("Content-Type", "application/xml; charset=utf-8")
}

// Register configures the WebDAV routes
func (h *WebDAV) Register(v1 *gin.RouterGroup) {
	v1.Use(setDavHeaders)

	v1.PUT("/*path", h.handlePut)
	v1.DELETE("/*path", h.handleDelete)
	v1.GET("/*path", h.handleGet)

	v1.Handle("PROPFIND", "/*path", h.handlePropfind)
	v1.Handle("MKCOL", "/*path", h.handleMkcol)
	v1.Handle("COPY", "/*path", h.handleCopyMove)
	v1.Handle("MOVE", "/*path", h.handleCopyMove)
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

// handlePropfind handles PROPFIND requests
func (h *WebDAV) handlePropfind(c *gin.Context) {
	name := c.Param("path")
	// Log request
	log.Printf("Handling PROPFIND request for %s", name)

	// Get file info using storage abstraction
	file, err := h.storage.GetFileInfo(name)
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
		files, err := h.storage.ListDir(name)
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
func createResponse(href string, file *stor.File) Response {
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
func (h *WebDAV) handlePut(c *gin.Context) {
	name := c.Param("path")
	// Write file using storage abstraction
	if err := h.storage.WriteToFile(name, c.Request.Body); err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to write file: %v", err)
		return
	}

	c.Status(http.StatusCreated)
}

// handleDelete handles DELETE requests
func (h *WebDAV) handleDelete(c *gin.Context) {
	name := c.Param("path")
	if err := h.storage.DeleteFile(name); err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to delete file: %v", err)
		return
	}
	c.Status(http.StatusNoContent)
}

// handleMkcol handles MKCOL requests
func (h *WebDAV) handleMkcol(c *gin.Context) {
	name := c.Param("path")
	if err := h.storage.CreateDir(name); err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to create directory: %v", err)
		return
	}
	c.Status(http.StatusCreated)
}

// handleCopyMove handles COPY and MOVE requests
func (h *WebDAV) handleCopyMove(c *gin.Context) {
	srcPath := c.Param("path")
	destination := c.Request.Header.Get("Destination")
	if destination == "" {
		sendError(c, http.StatusBadRequest, "Destination header required")
		return
	}

	// Parse destination path
	destPath := strings.TrimPrefix(destination, "/")

	// Create destination directory if needed
	if err := h.storage.CreateDir(filepath.Dir(destPath)); err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to create destination directory: %v", err)
		return
	}

	// Handle COPY or MOVE
	if c.Request.Method == "COPY" {
		// Copy file/directory using storage
		if err := h.storage.CopyFile(srcPath, destPath); err != nil {
			sendError(c, http.StatusInternalServerError, "Failed to copy file: %v", err)
			return
		}
	} else {
		// Move file/directory using storage
		if err := h.storage.MoveFile(srcPath, destPath); err != nil {
			sendError(c, http.StatusInternalServerError, "Failed to move file: %v", err)
			return
		}
	}

	c.Status(http.StatusCreated)
}

// handleGet handles GET requests
func (h *WebDAV) handleGet(c *gin.Context) {
	name := c.Param("path")

	info, err := h.storage.GetFileInfo(name)
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

	file, err := h.storage.OpenFile(name)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "Error opening file: %v", err)
		return
	}
	defer file.Close()

	if _, err := io.Copy(c.Writer, file); err != nil {
		log.Printf("Failed to copy file content: %s", err)
	}
}
